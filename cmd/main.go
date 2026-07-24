package main

import (
	"context"
	"crypto/rsa"
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
	"github.com/billykore/project-one/internal/core/ports"
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
type application struct {
	echo                *echo.Echo
	db                  *gorm.DB
	publisher           ports.Publisher
	subscriber          ports.Subscriber
	wsManager           *wsadapter.Manager
	notificationHandler *handler.NotificationHandler
}

func main() {
	configPath := flag.String("config", "./configs", "path to config directory")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	lgr := logger.New()

	cfg, err := config.Load(*configPath)
	if err != nil {
		lgr.Fatal(ctx, "failed to load config", "error", err)
	}

	privateKey, publicKey, err := loadRSAKeyPair(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
	if err != nil {
		lgr.Fatal(ctx, "failed to load jwt keys", "error", err)
	}

	swagger.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.App.Port)

	app, err := newApplication(cfg, privateKey, publicKey, lgr)
	if err != nil {
		lgr.Fatal(ctx, "failed to initialize application", "error", err)
	}

	lgr.Info(ctx, "starting server", "port", cfg.App.Port)

	go func() {
		if err := app.notificationHandler.Listen(ctx); err != nil {
			lgr.Fatal(ctx, "failed to start notification consumer", "error", err)
		}
	}()

	go func() {
		err := app.echo.Start(fmt.Sprintf(":%d", cfg.App.Port))
		if err != nil && err != http.ErrServerClosed {
			lgr.Fatal(ctx, "failed to start server", "error", err)
		}
	}()

	<-ctx.Done()

	lgr.Info(ctx, "shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.shutdown(shutdownCtx, lgr); err != nil {
		lgr.Fatal(ctx, "server forced to shutdown", "error", err)
	}
}

func newApplication(cfg *config.Config, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, lgr *logger.Logger) (*application, error) {
	db, err := setupDB(cfg.Database)
	if err != nil {
		return nil, err
	}

	val := validator.New()

	userRepo := repository.NewUserRepository(db)
	userTokenRepo := repository.NewUserTokenRepository(db)
	postRepo := repository.NewPostRepository(db)
	followRepo := repository.NewFollowRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	tokenSvc := token.NewJWTTokenService(privateKey, publicKey, cfg.JWT.ExpirationTime)
	hasherSvc := hasher.NewBcryptHasher()
	inMemoryPubSub := pubsub.NewInMemoryPubSub()
	publisher := pubsub.NewInMemoryPublisher(inMemoryPubSub)
	subscriber := pubsub.NewInMemorySubscriber(inMemoryPubSub)
	wsManager := wsadapter.NewManager()

	loginUc := usecase.NewLoginUseCase(userRepo, tokenSvc, userTokenRepo, hasherSvc, lgr)
	userUc := usecase.NewUserUseCase(userRepo, hasherSvc)
	postUc := usecase.NewPostUseCase(postRepo, likeRepo, userRepo, publisher, lgr)
	followUc := usecase.NewFollowUseCase(followRepo, userRepo, publisher, lgr)
	commentUc := usecase.NewCommentUseCase(commentRepo, postRepo, userRepo, publisher)
	notificationUc := usecase.NewNotificationUseCase(notificationRepo, userRepo, lgr)
	feedUc := usecase.NewFeedUseCase(postRepo, followRepo, userRepo, lgr)

	userHdl := handler.NewUserHandler(userUc, loginUc, followUc, postUc, val, lgr)
	postHdl := handler.NewPostHandler(postUc, commentUc, val, lgr)
	commentHdl := handler.NewCommentHandler(commentUc, val, lgr)
	notificationHdl := handler.NewNotificationHandler(lgr, subscriber, notificationUc, val, wsManager)
	wsHdl := handler.NewWebSocketHandler(lgr, tokenSvc, userUc, wsManager)
	feedHdl := handler.NewFeedHandler(feedUc, lgr)

	e := echo.New()
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.RequestLogger())
	e.HTTPErrorHandler = middleware.ErrorHandler(lgr, cfg.App.ErrorTypeBaseURL, cfg.App.Env == "debug")

	e.GET("/websocket", wsHdl.HandleUpgrade, middleware.Authorize(tokenSvc))
	if cfg.App.Env != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	registerRoutes(e, tokenSvc, userHdl, postHdl, commentHdl, notificationHdl, feedHdl)

	return &application{
		echo:                e,
		db:                  db,
		publisher:           publisher,
		subscriber:          subscriber,
		wsManager:           wsManager,
		notificationHandler: notificationHdl,
	}, nil
}

func registerRoutes(
	e *echo.Echo,
	tokenSvc ports.TokenService,
	userHdl *handler.UserHandler,
	postHdl *handler.PostHandler,
	commentHdl *handler.CommentHandler,
	notificationHdl *handler.NotificationHandler,
	feedHdl *handler.FeedHandler,
) {
	auth := e.Group("/auth")
	auth.POST("/register", userHdl.HandleRegister)
	auth.POST("/login", userHdl.HandleLogin)
	auth.POST("/logout", userHdl.HandleLogout, middleware.Authorize(tokenSvc))

	users := e.Group("/users")
	users.GET("/:username", userHdl.GetUser)
	users.GET("/:username/posts", userHdl.GetUserPosts)

	usersAuth := users.Group("", middleware.Authorize(tokenSvc))
	usersAuth.PUT("/password", userHdl.HandleChangePassword)
	usersAuth.PUT("/profile", userHdl.HandleUpdateProfile)
	usersAuth.GET("/:username/following", userHdl.GetFollowing)
	usersAuth.GET("/:username/followers", userHdl.GetFollowers)
	usersAuth.POST("/:username/followers", userHdl.HandleFollow)
	usersAuth.DELETE("/:username/followers", userHdl.HandleUnfollow)

	e.GET("/posts/:id", postHdl.GetPostByID)
	posts := e.Group("/posts", middleware.Authorize(tokenSvc))
	posts.POST("", postHdl.CreatePost)
	posts.GET("", postHdl.GetPosts)
	posts.PUT("/:id", postHdl.UpdatePost)
	posts.DELETE("/:id", postHdl.DeletePost)
	posts.POST("/:id/comments", postHdl.CreateComment)
	posts.POST("/:id/likes", postHdl.LikePost)
	posts.DELETE("/:id/likes", postHdl.UnlikePost)
	posts.GET("/:id/likes", postHdl.GetLikeStatus)

	comments := e.Group("/comments", middleware.Authorize(tokenSvc))
	comments.PUT("/:id", commentHdl.EditComment)
	comments.DELETE("/:id", commentHdl.DeleteComment)

	notifications := e.Group("/notifications", middleware.Authorize(tokenSvc))
	notifications.GET("", notificationHdl.GetNotifications)
	notifications.PUT("/:id/read", notificationHdl.MarkAsRead)
	notifications.PUT("/read-all", notificationHdl.MarkAllAsRead)

	feeds := e.Group("/feeds", middleware.Authorize(tokenSvc))
	feeds.GET("", feedHdl.HandleGetFeed)
}

func (a *application) shutdown(ctx context.Context, lgr *logger.Logger) error {
	var shutdownErr error

	if err := a.echo.Shutdown(ctx); err != nil {
		lgr.Error(ctx, "failed to shutdown server", "error", err)
		shutdownErr = err
	}

	if err := a.subscriber.Close(); err != nil {
		lgr.Error(ctx, "failed to close subscriber", "error", err)
		if shutdownErr == nil {
			shutdownErr = err
		}
	}

	if err := a.publisher.Close(); err != nil {
		lgr.Error(ctx, "failed to close publisher", "error", err)
		if shutdownErr == nil {
			shutdownErr = err
		}
	}

	if err := a.wsManager.Close(); err != nil {
		lgr.Error(ctx, "failed to close websocket manager", "error", err)
		if shutdownErr == nil {
			shutdownErr = err
		}
	}

	if sqlDB, err := a.db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			lgr.Error(ctx, "failed to close database connection", "error", err)
			if shutdownErr == nil {
				shutdownErr = err
			}
		}
	} else {
		lgr.Error(ctx, "failed to get sql.DB for closing", "error", err)
		if shutdownErr == nil {
			shutdownErr = err
		}
	}

	return shutdownErr
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

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
