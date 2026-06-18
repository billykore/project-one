package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/billykore/project-one/api/swagger"
	"github.com/billykore/project-one/internal/adapters/hasher"
	"github.com/billykore/project-one/internal/adapters/logger"
	"github.com/billykore/project-one/internal/adapters/pubsub"
	"github.com/billykore/project-one/internal/adapters/repository"
	"github.com/billykore/project-one/internal/adapters/token"
	"github.com/billykore/project-one/internal/adapters/validator"
	wsadapter "github.com/billykore/project-one/internal/adapters/websocket"
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
	configPath := flag.String("config", "./configs", "path to config directory")
	flag.Parse()

	ctx := context.Background()
	lgr := logger.New()

	// Load configuration.
	cfg, err := config.Load(*configPath)
	if err != nil {
		lgr.Fatal(ctx, "failed to load config", "error", err)
	}

	// Set dynamic Swagger host.
	swagger.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.App.Port)

	// 1. Initialize DB.
	db, err := setupDB(cfg.Database)
	if err != nil {
		lgr.Fatal(ctx, "failed to connect to database", "error", err)
	}

	// 2. Initialize Validator.
	val := validator.New()

	// 3. Initialize Adapters.
	userRepo := repository.NewUserRepository(db)
	userTokenRepo := repository.NewUserTokenRepository(db)
	postRepo := repository.NewPostRepository(db)
	followRepo := repository.NewFollowRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	tokenSvc := token.NewJWTTokenService(cfg.JWT.SecretKey, cfg.JWT.ExpirationTime)
	hasher := hasher.NewBcryptHasher()
	inMemoryPubSub := pubsub.NewInMemoryPubSub()
	publisher := pubsub.NewInMemoryPublisher(inMemoryPubSub)
	subcriber := pubsub.NewInMemorySubscriber(inMemoryPubSub)
	notificationRepo := repository.NewNotificationRepository(db)
	wsManager := wsadapter.NewManager()

	// 4. Initialize UseCase.
	loginUc := usecase.NewLoginUseCase(userRepo, tokenSvc, userTokenRepo, hasher, lgr)
	userUc := usecase.NewUserUseCase(userRepo, userTokenRepo, hasher)
	postUc := usecase.NewPostUseCase(postRepo, likeRepo, userRepo, publisher, lgr)
	followUc := usecase.NewFollowUseCase(followRepo, userRepo, publisher, lgr)
	commentUc := usecase.NewCommentUseCase(commentRepo, postRepo, userRepo, publisher, lgr)
	notificationUc := usecase.NewNotificationUseCase(notificationRepo, userRepo, lgr)

	// 5. Initialize Handler.
	userHdl := handler.NewUserHandler(userUc, loginUc, followUc, postUc, val, lgr)
	postHdl := handler.NewPostHandler(postUc, commentUc, val)
	commentHdl := handler.NewCommentHandler(commentUc, val, lgr)
	notificationHdl := handler.NewNotificationHandler(lgr, subcriber, notificationUc, val, wsManager)
	wsHdl := handler.NewWebSocketHandler(lgr, tokenSvc, userUc, wsManager)

	// 6. Set up Echo.
	e := echo.New()
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.RequestLogger())

	// WebSocket endpoint.
	e.GET("/websocket", wsHdl.HandleUpgrade, middleware.Authorize(tokenSvc))

	// Only expose Swagger UI in non-production environments.
	if cfg.App.Env != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// Authentication Group
	auth := e.Group("/auth")
	{
		auth.POST("/register", userHdl.HandleRegister)
		auth.POST("/login", userHdl.HandleLogin)
		auth.POST("/logout", userHdl.HandleLogout, middleware.Authorize(tokenSvc))
	}

	// Users Group
	users := e.Group("/users")
	{
		// User profile endpoints.
		users.GET("/:username", userHdl.GetUser)

		// User's posts.
		users.GET("/:username/posts", userHdl.GetUserPosts)

		// Social sub-resources (authorized).
		usersAuth := users.Group("", middleware.Authorize(tokenSvc))
		{
			usersAuth.PUT("/password", userHdl.HandleChangePassword)
			usersAuth.GET("/:username/following", userHdl.GetFollowing)
			usersAuth.GET("/:username/followers", userHdl.GetFollowers)
			usersAuth.POST("/:username/followers", userHdl.HandleFollow)
			usersAuth.DELETE("/:username/followers", userHdl.HandleUnfollow)
		}
	}

	// Public Post Routes
	e.GET("/posts/:id", postHdl.GetPostByID)

	// Posts Group
	posts := e.Group("/posts", middleware.Authorize(tokenSvc))
	{
		posts.POST("", postHdl.CreatePost)
		posts.GET("", postHdl.GetPosts)
		posts.PUT("/:id", postHdl.UpdatePost)
		posts.DELETE("/:id", postHdl.DeletePost)

		// Comments on posts.
		posts.POST("/:id/comments", postHdl.CreateComment)

		// Likes on posts.
		posts.POST("/:id/likes", postHdl.LikePost)
		posts.DELETE("/:id/likes", postHdl.UnlikePost)
		posts.GET("/:id/likes", postHdl.GetLikeStatus)
	}

	// Comments Group
	comments := e.Group("/comments", middleware.Authorize(tokenSvc))
	{
		comments.PUT("/:id", commentHdl.EditComment)
		comments.DELETE("/:id", commentHdl.DeleteComment)
	}

	// Notifications Group
	notifications := e.Group("/notifications", middleware.Authorize(tokenSvc))
	{
		notifications.GET("", notificationHdl.GetNotifications)
		notifications.PUT("/:id/read", notificationHdl.MarkAsRead)
		notifications.PUT("/read-all", notificationHdl.MarkAllAsRead)
	}

	// Start server.
	lgr.Info(ctx, "starting server", "port", cfg.App.Port)

	go func(ctx context.Context) {
		if err := notificationHdl.Listen(ctx); err != nil {
			lgr.Fatal(ctx, "failed to start notification consumer", "error", err)
		}
	}(ctx)

	go func() {
		err := e.Start(fmt.Sprintf(":%d", cfg.App.Port))
		if err != nil && err != http.ErrServerClosed {
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

	if err := closeDB(db); err != nil {
		lgr.Error(ctx, "failed to close database connection", "error", err)
	}

	if err := subcriber.Close(); err != nil {
		lgr.Error(ctx, "failed to close subscriber", "error", err)
	}

	if err := publisher.Close(); err != nil {
		lgr.Error(ctx, "failed to close publisher", "error", err)
	}

	if err := wsManager.Close(); err != nil {
		lgr.Error(ctx, "failed to close websocket manager", "error", err)
	}

	if err := e.Shutdown(ctxShutdown); err != nil {
		lgr.Fatal(ctx, "server forced to shutdown", "error", err)
	}

	lgr.Info(ctx, "server exited gracefully")
}

// setupDB initializes the database connection using GORM and configures connection pooling.
func setupDB(dbConfig config.DatabaseConfig) (*gorm.DB, error) {
	// Construct DSN from config.
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User,
		dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)

	// Open the database connection using GORM.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure DB connection pool tuning.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	// Set reasonable defaults for connection pooling.
	// These can be further tuned based on load testing.
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

	// Verify the database connection is alive.
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// closeDB gracefully closes the database connection.
func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
