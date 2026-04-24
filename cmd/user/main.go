package main

import (
	"context"
	"os"

	"github.com/billykore/project-one/internal/app/user/adapters/handler"
	"github.com/billykore/project-one/internal/app/user/adapters/hasher"
	"github.com/billykore/project-one/internal/app/user/adapters/logger"
	"github.com/billykore/project-one/internal/app/user/adapters/repository"
	"github.com/billykore/project-one/internal/app/user/adapters/token"
	"github.com/billykore/project-one/internal/app/user/core/ports"
	"github.com/billykore/project-one/internal/app/user/core/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()
	lgr := logger.NewZerologLogger()

	// Configuration (using default values or env)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default local postgres for development
		dsn = "host=localhost user=postgres password=password dbname=project-one port=5432 sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "my-super-secret-key"
	}

	// 1. Initialize DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		lgr.Fatal(ctx, "failed to connect to database", "error", err)
	}

	// 2. Initialize Validator
	val := validator.New()

	// 3. Initialize Adapters
	repo := repository.NewPostgresUserRepository(db)
	tks := token.NewJWTTokenService(jwtSecret)
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

	// Seed a test user if needed
	seedTestUser(ctx, db, hsh, lgr)

	// Start server
	lgr.Info(ctx, "starting server", "port", 8080)
	if err := e.Start(":8080"); err != nil {
		lgr.Fatal(ctx, "failed to start server", "error", err)
	}
}

func seedTestUser(ctx context.Context, db *gorm.DB, hsh interface {
	Hash(ctx context.Context, password string) (string, error)
}, log ports.Logger) {
	var count int64
	db.Table("users").Count(&count)
	if count == 0 {
		hashed, _ := hsh.Hash(ctx, "password123")
		db.Table("users").Create(map[string]interface{}{
			"email":    "user@example.com",
			"password": hashed,
		})
		log.Info(ctx, "seeded test user", "email", "user@example.com", "password", "password123")
	}
}
