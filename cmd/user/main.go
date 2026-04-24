package main

import (
	"context"
	"fmt"

	"github.com/billykore/project-one/internal/app/user/adapters/handler"
	"github.com/billykore/project-one/internal/app/user/adapters/hasher"
	"github.com/billykore/project-one/internal/app/user/adapters/logger"
	"github.com/billykore/project-one/internal/app/user/adapters/repository"
	"github.com/billykore/project-one/internal/app/user/adapters/token"
	"github.com/billykore/project-one/internal/app/user/config" // Import the new config package
	"github.com/billykore/project-one/internal/app/user/core/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()
	lgr := logger.NewZerologLogger()

	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		lgr.Fatal(ctx, "failed to load config", "error", err)
	}

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
	repo := repository.NewPostgresUserRepository(db)
	tks := token.NewJWTTokenService(cfg.JWT.SecretKey, cfg.JWT.ExpirationTime) // Pass JWT secret and expiration from config
	hsh := hasher.NewBcryptHasher()

	// 4. Initialize Service
	svc := service.NewLoginService(repo, tks, hsh, lgr)
	userSvc := service.NewUserService(repo)

	// 5. Initialize Handler
	userHdl := handler.NewUserHandler(userSvc, svc, val)

	// 6. Set up Echo
	e := echo.New()
	e.POST("/user/login", userHdl.HandleLogin)
	e.POST("/user/logout", userHdl.HandleLogout, handler.AuthMiddleware(tks))
	e.GET("/user/me", userHdl.Me, handler.AuthMiddleware(tks))

	// Start server
	lgr.Info(ctx, "starting server", "port", cfg.App.Port)
	if err := e.Start(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
		lgr.Fatal(ctx, "failed to start server", "error", err)
	}
}
