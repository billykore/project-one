package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/billykore/project-one/api/swagger"
	"github.com/billykore/project-one/internal/adapters/hasher"
	"github.com/billykore/project-one/internal/adapters/logger"
	"github.com/billykore/project-one/internal/adapters/repository"
	"github.com/billykore/project-one/internal/adapters/token"
	"github.com/billykore/project-one/internal/adapters/validator"
	"github.com/billykore/project-one/internal/api/handler"
	"github.com/billykore/project-one/internal/api/middleware"
	"github.com/billykore/project-one/internal/config"
	"github.com/billykore/project-one/internal/core/usecase"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title						User Service API
// @version					1.0
// @description				This is the API server for the User Service.
// @host						localhost:8080
// @BasePath					/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
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
	followRepo := repository.NewFollowRepository(db)
	tokenSvc := token.NewJWTTokenService(cfg.JWT.SecretKey, cfg.JWT.ExpirationTime)
	hasher := hasher.NewBcryptHasher()

	// 4. Initialize UseCase
	loginUc := usecase.NewLoginUseCase(userRepo, tokenSvc, userTokenRepo, hasher, lgr)
	userUc := usecase.NewUserUseCase(userRepo, userTokenRepo, hasher)
	postUc := usecase.NewPostUseCase(postRepo, lgr)
	followUc := usecase.NewFollowUseCase(followRepo, userRepo)

	// 5. Initialize Handler
	userHdl := handler.NewUserHandler(userUc, loginUc, followUc, val, lgr)
	postHdl := handler.NewPostHandler(postUc, val)

	// 6. Set up Echo
	e := echo.New()
	e.Use(echomiddleware.RequestLogger())
	e.Use(echomiddleware.Recover())

	// Only expose Swagger UI in non-production environments
	if cfg.App.Env != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	e.POST("/users/register", userHdl.HandleRegister)
	e.POST("/users/login", userHdl.HandleLogin)
	e.POST("/users/logout", userHdl.HandleLogout, middleware.Authorize(tokenSvc))
	e.GET("/users/:username/following", userHdl.GetFollowing, middleware.Authorize(tokenSvc))
	e.GET("/users/:username/followers", userHdl.GetFollowers, middleware.Authorize(tokenSvc))
	e.GET("/users/:username", userHdl.GetUser)
	e.POST("/users/:username/follow", userHdl.HandleFollow, middleware.Authorize(tokenSvc))
	e.DELETE("/users/:username/follow", userHdl.HandleUnfollow, middleware.Authorize(tokenSvc))
	e.GET("/users/:username/posts", postHdl.GetUserPosts)
	e.POST("/posts", postHdl.CreatePost, middleware.Authorize(tokenSvc))
	e.GET("/posts", postHdl.GetPosts, middleware.Authorize(tokenSvc))
	e.GET("/posts/:id", postHdl.GetPostByID, middleware.Authorize(tokenSvc))
	e.PUT("/posts/:id", postHdl.UpdatePost, middleware.Authorize(tokenSvc))
	e.DELETE("/posts/:id", postHdl.DeletePost, middleware.Authorize(tokenSvc))

	// Start server
	lgr.Info(ctx, "starting server", "port", cfg.App.Port)

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", cfg.App.Port)); err != nil && err != http.ErrServerClosed {
			lgr.Fatal(ctx, "failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	lgr.Info(ctx, "shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctxShutdown); err != nil {
		lgr.Fatal(ctx, "server forced to shutdown", "error", err)
	}

	if err := closeDB(ctx, db, lgr); err != nil {
		lgr.Error(ctx, "failed to close database connection", "error", err)
	}

	lgr.Info(ctx, "server exited gracefully")
}

func closeDB(ctx context.Context, db *gorm.DB, lgr *logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		lgr.Error(ctx, "failed to get database", "error", err)
		return err
	}
	return sqlDB.Close()
}
