package main

import (
	"context"
	"fmt"

	user "github.com/billykore/project-one/api/swagger"
	"github.com/billykore/project-one/internal/adapters/handler"
	"github.com/billykore/project-one/internal/adapters/hasher"
	"github.com/billykore/project-one/internal/adapters/logger"
	"github.com/billykore/project-one/internal/adapters/repository"
	"github.com/billykore/project-one/internal/adapters/token"
	"github.com/billykore/project-one/internal/adapters/validator"
	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/service"
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
	user.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.App.Port)

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

	// 4. Initialize Service
	loginSvc := service.NewLoginService(userRepo, tokenSvc, userTokenRepo, hasher, lgr)
	userSvc := service.NewUserService(userRepo, userTokenRepo, hasher)
	postSvc := service.NewPostService(postRepo, lgr)

	// 5. Initialize Handler
	userHdl := handler.NewUserHandler(userSvc, loginSvc, val)
	postHdl := handler.NewPostHandler(postSvc, val)

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

	// Start server
	lgr.Info(ctx, "starting server", "port", cfg.App.Port)
	if err := e.Start(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
		lgr.Fatal(ctx, "failed to start server", "error", err)
	}
}
