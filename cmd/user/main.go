package main

import (
	"context"
	"log"
	"os"

	"github.com/billykore/project-one/internal/app/user/adapters/handler"
	"github.com/billykore/project-one/internal/app/user/adapters/hasher"
	"github.com/billykore/project-one/internal/app/user/adapters/logger"
	"github.com/billykore/project-one/internal/app/user/adapters/repository"
	"github.com/billykore/project-one/internal/app/user/adapters/token"
	"github.com/billykore/project-one/internal/app/user/core/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 2. Initialize Validator
	val := validator.New()

	// 3. Initialize Adapters
	repo := repository.NewPostgresUserRepository(db)
	tks := token.NewJWTTokenService(jwtSecret)
	hsh := hasher.NewBcryptHasher()
	lgr := logger.NewZerologLogger()

	// 4. Initialize Service
	svc := service.NewLoginService(repo, tks, hsh, lgr)
	userSvc := service.NewUserService(repo)

	// 5. Initialize Handler
	hdl := handler.NewLoginHandler(svc, val)
	userHdl := handler.NewUserHandler(userSvc)

	// 6. Set up Echo
	e := echo.New()
	e.POST("/login", hdl.HandleLogin)
	e.POST("/logout", hdl.HandleLogout, handler.AuthMiddleware(tks))
	e.GET("/user/me", userHdl.Me, handler.AuthMiddleware(tks))

	// Seed a test user if needed
	seedTestUser(db, hsh)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func seedTestUser(db *gorm.DB, hsh interface {
	Hash(ctx context.Context, password string) (string, error)
}) {
	var count int64
	db.Table("users").Count(&count)
	if count == 0 {
		hashed, _ := hsh.Hash(context.Background(), "password123")
		db.Table("users").Create(map[string]interface{}{
			"email":    "user@example.com",
			"password": hashed,
		})
		log.Println("Seeded test user: user@example.com / password123")
	}
}
