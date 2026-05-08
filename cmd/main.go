package main

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/api/swagger"
	"github.com/billykore/project-one/internal/api/handler"
	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/usecase"
	"github.com/billykore/project-one/internal/infrastructure/hasher"
	"github.com/billykore/project-one/internal/infrastructure/logger"
	"github.com/billykore/project-one/internal/infrastructure/repository"
	"github.com/billykore/project-one/internal/infrastructure/token"
	"github.com/billykore/project-one/internal/infrastructure/validator"
	"github.com/labstack/echo/v4"

	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           User Service API
// @version         1.0
// @description     This is the API server for the User Service.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	ctx := context.Background()
	lgr := logger.New()

	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		lgr.Fatal(ctx, "failed to load config", "error", err)
	}

	// Set dynamic Swagger host
	swagger.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.App.Port)

	// Construct DSN from config
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	// 1. Initialize DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		lgr.Fatal(ctx, "failed to connect to database", "error", err)
	}

	// 2. Initialize Validator
	val := validator.New()

	// 3. Initialize Adapters
	userRepo := repository.NewUserRepository(db)
	userTokenRepo := repository.NewUserTokenRepository(db)
	postRepo := repository.NewPostRepository(db)
	tokenSvc := token.NewJWTTokenService(cfg.JWT.SecretKey, cfg.JWT.ExpirationTime)
	hasher := hasher.NewBcryptHasher()

	// 4. Initialize UseCase
	loginUc := usecase.NewLoginUseCase(userRepo, tokenSvc, userTokenRepo, hasher, lgr)
	userUc := usecase.NewUserUseCase(userRepo, userTokenRepo, hasher)
	postUc := usecase.NewPostUseCase(postRepo, lgr)

	// 5. Initialize Handler
	userHdl := handler.NewUserHandler(userUc, loginUc, val)
	postHdl := handler.NewPostHandler(postUc, val)

	// 6. Set up Echo
	e := echo.New()

	// Only expose Swagger UI in non-production environments
	if cfg.App.Env != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	e.POST("/users/register", userHdl.HandleRegister)
	e.POST("/users/login", userHdl.HandleLogin)
	e.POST("/users/logout", userHdl.HandleLogout, handler.AuthMiddleware(tokenSvc))
	e.GET("/users/me", userHdl.Me, handler.AuthMiddleware(tokenSvc))
	e.POST("/posts", postHdl.CreatePost, handler.AuthMiddleware(tokenSvc))
	e.GET("/posts/:id", postHdl.GetPostByID, handler.AuthMiddleware(tokenSvc))
	e.PUT("/posts/:id", postHdl.UpdatePost, handler.AuthMiddleware(tokenSvc))
	e.DELETE("/posts/:id", postHdl.DeletePost, handler.AuthMiddleware(tokenSvc))

	// Start server
	lgr.Info(ctx, "starting server", "port", cfg.App.Port)
	if err := e.Start(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
		lgr.Fatal(ctx, "failed to start server", "error", err)
	}
}
