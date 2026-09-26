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
	featureflagrepository "github.com/billykore/project-one/internal/featureflags/adapters"
	featureflagadapter "github.com/billykore/project-one/internal/featureflags/adapters/featureflag"
	featureflaghandler "github.com/billykore/project-one/internal/featureflags/api/handler"
	featureflagmiddleware "github.com/billykore/project-one/internal/featureflags/api/middleware"
	featureflagdomain "github.com/billykore/project-one/internal/featureflags/domain"
	featureflagports "github.com/billykore/project-one/internal/featureflags/ports"
	featureflagusecase "github.com/billykore/project-one/internal/featureflags/usecase"
	identityrepository "github.com/billykore/project-one/internal/identity/adapters"
	"github.com/billykore/project-one/internal/identity/adapters/hasher"
	"github.com/billykore/project-one/internal/identity/adapters/token"
	identityhandler "github.com/billykore/project-one/internal/identity/api/handler"
	identitymiddleware "github.com/billykore/project-one/internal/identity/api/middleware"
	identityports "github.com/billykore/project-one/internal/identity/ports"
	identityusecase "github.com/billykore/project-one/internal/identity/usecase"
	notificationrepository "github.com/billykore/project-one/internal/notifications/adapters"
	sseadapter "github.com/billykore/project-one/internal/notifications/adapters/sse"
	notificationhandler "github.com/billykore/project-one/internal/notifications/api/handler"
	notificationusecase "github.com/billykore/project-one/internal/notifications/usecase"
	healthadapter "github.com/billykore/project-one/internal/operations/adapters/health"
	metricsadapter "github.com/billykore/project-one/internal/operations/adapters/metrics"
	operationshandler "github.com/billykore/project-one/internal/operations/api/handler"
	operationsmiddleware "github.com/billykore/project-one/internal/operations/api/middleware"
	operationsports "github.com/billykore/project-one/internal/operations/ports"
	operationsusecase "github.com/billykore/project-one/internal/operations/usecase"
	"github.com/billykore/project-one/internal/platform/adapters/logger"
	"github.com/billykore/project-one/internal/platform/adapters/pubsub"
	"github.com/billykore/project-one/internal/platform/adapters/validator"
	platformmiddleware "github.com/billykore/project-one/internal/platform/api/middleware"
	"github.com/billykore/project-one/internal/platform/config"
	platformports "github.com/billykore/project-one/internal/platform/ports"
	publishingrepository "github.com/billykore/project-one/internal/publishing/adapters"
	publishinghandler "github.com/billykore/project-one/internal/publishing/api/handler"
	publishingusecase "github.com/billykore/project-one/internal/publishing/usecase"
	socialrepository "github.com/billykore/project-one/internal/social/adapters"
	socialhandler "github.com/billykore/project-one/internal/social/api/handler"
	socialusecase "github.com/billykore/project-one/internal/social/usecase"
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
	publisher           platformports.Publisher
	subscriber          platformports.Subscriber
	sseManager          *sseadapter.Manager
	notificationHandler *notificationhandler.NotificationHandler
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
	publisher, err := pubsub.NewRabbitMQPublisher(cfg.MessageBroker.RabbitMQ, lgr)
	if err != nil {
		return nil, err
	}
	subscriber, err := pubsub.NewRabbitMQSubscriber(cfg.MessageBroker.RabbitMQ, lgr)
	if err != nil {
		return nil, err
	}

	sseManager := sseadapter.NewManager()

	val := validator.New()

	userRepo := identityrepository.NewUserRepository(db)
	userSearchRepo := identityrepository.NewUserSearchRepository(db)
	userTokenRepo := identityrepository.NewUserTokenRepository(db)
	postCommandRepo := publishingrepository.NewPostCommandRepository(db)
	postQueryRepo := publishingrepository.NewPostQueryRepository(db)
	featureFlagRepo := featureflagrepository.NewFeatureFlagRepository(db)
	followRepo := socialrepository.NewFollowRepository(db)
	commentRepo := publishingrepository.NewCommentRepository(db)
	likeRepo := publishingrepository.NewLikeRepository(db)
	notificationRepo := notificationrepository.NewNotificationRepository(db)

	tokenSvc := token.NewJWTTokenService(privateKey, publicKey, cfg.JWT.ExpirationTime)
	hasherSvc := hasher.NewBcryptHasher()
	authenticator := identityusecase.NewAuthenticationUseCase(tokenSvc, userTokenRepo, userRepo)

	loginUc := identityusecase.NewLoginUseCase(userRepo, tokenSvc, userTokenRepo, hasherSvc, lgr)
	userUc := identityusecase.NewUserUseCase(userRepo, hasherSvc, userSearchRepo)
	featureFlagEvaluator, err := featureflagadapter.NewEvaluator(
		featureFlagRepo,
		lgr,
		featureflagdomain.Environment(cfg.FeatureFlags.Environment),
		cfg.FeatureFlags.RefreshInterval,
	)
	if err != nil {
		return nil, err
	}
	featureFlagEvaluator.StartRefreshLoop(context.Background())
	featureFlagUc := featureflagusecase.NewFeatureFlagUseCase(featureFlagRepo, featureFlagEvaluator, lgr)
	postCommandUc := publishingusecase.NewPostCommandUseCase(postCommandRepo, likeRepo, userRepo, publisher, lgr, featureFlagEvaluator)
	postQueryUc := publishingusecase.NewPostQueryUseCase(postQueryRepo, likeRepo, lgr)
	followUc := socialusecase.NewFollowUseCase(followRepo, userRepo, publisher, lgr)
	commentUc := publishingusecase.NewCommentUseCase(commentRepo, postCommandRepo, userRepo, publisher)
	notificationUc := notificationusecase.NewNotificationUseCase(notificationRepo, userRepo, lgr)
	feedUc := socialusecase.NewFeedUseCase(postQueryRepo, followRepo, lgr)

	userHdl := identityhandler.NewUserHandler(userUc, loginUc, followUc, postQueryUc, val, lgr)
	postCommandHdl := publishinghandler.NewPostCommandHandler(postCommandUc, commentUc, val, lgr)
	postQueryHdl := publishinghandler.NewPostQueryHandler(postQueryUc, commentUc, lgr)
	commentHdl := publishinghandler.NewCommentHandler(commentUc, val, lgr)
	notificationHdl := notificationhandler.NewNotificationHandler(lgr, subscriber, notificationUc, userUc, val, sseManager)
	feedHdl := socialhandler.NewFeedHandler(feedUc, lgr)
	featureFlagHdl := featureflaghandler.NewFeatureFlagHandler(featureFlagUc, val, cfg.FeatureFlags.Environment)

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to access database connection for health checks: %w", err)
	}
	// ponytail: the notification component reports the subscriber connection
	// state, not the lazily-connecting publisher, so readiness is meaningful
	// before the first publish.
	healthCheckers := []operationsports.DependencyChecker{
		healthadapter.NewDatabaseChecker(sqlDB),
	}
	if reporter, ok := subscriber.(platformports.HealthReporter); ok {
		healthCheckers = append(healthCheckers, healthadapter.NewNotificationChecker(cfg.MessageBroker.Type, reporter))
	} else {
		lgr.Warn(context.Background(), "message broker does not report connection health", "type", cfg.MessageBroker.Type)
	}

	// The metrics recorder is the health observer, so every readiness
	// assessment keeps the readiness and dependency gauges current.
	metricsRecorder := metricsadapter.NewPrometheus()

	monitoringCredential, err := metricsadapter.LoadCredential(cfg.Monitoring.Username, cfg.Monitoring.PasswordFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load monitoring credentials: %w", err)
	}
	if monitoringCredential == nil {
		lgr.Warn(context.Background(), "monitoring credentials are not configured; /metrics rejects every scrape")
	}

	healthUc := operationsusecase.NewHealthUseCase(healthCheckers, metricsRecorder, 0, lgr)
	healthHdl := operationshandler.NewHealthHandler(healthUc, lgr)

	e := echo.New()
	// Registered first so every completed request, including recovered panics,
	// is observed with the status the error handler will write.
	e.Use(operationsmiddleware.Metrics(metricsRecorder))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(platformmiddleware.RequestLogging(lgr))
	e.HTTPErrorHandler = platformmiddleware.ErrorHandler(lgr, cfg.App.ErrorTypeBaseURL, cfg.App.Env == "debug")

	if cfg.App.Env != "production" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	registerRoutes(e, authenticator, userHdl, postCommandHdl, postQueryHdl, commentHdl, notificationHdl, feedHdl, featureFlagHdl, healthHdl, metricsRecorder.Handler(monitoringCredential), cfg.FeatureFlags.Operators, featureFlagEvaluator)

	return &application{
		echo:                e,
		db:                  db,
		publisher:           publisher,
		subscriber:          subscriber,
		sseManager:          sseManager,
		notificationHandler: notificationHdl,
	}, nil
}

func registerRoutes(
	e *echo.Echo,
	authenticator identityports.Authenticator,
	userHdl *identityhandler.UserHandler,
	postCommandHdl *publishinghandler.PostCommandHandler,
	postQueryHdl *publishinghandler.PostQueryHandler,
	commentHdl *publishinghandler.CommentHandler,
	notificationHdl *notificationhandler.NotificationHandler,
	feedHdl *socialhandler.FeedHandler,
	featureFlagHdl *featureflaghandler.FeatureFlagHandler,
	healthHdl *operationshandler.HealthHandler,
	metricsHandler http.Handler,
	featureFlagOperators []string,
	featureFlagEvaluator featureflagports.FeatureFlagEvaluator,
) {
	// Liveness answers for the process only; readiness reports dependency state.
	e.GET("/healthz", healthHdl.HandleLiveness)
	e.GET("/status", healthHdl.HandleReadiness)

	registerMetricsRoute(e, metricsHandler)

	auth := e.Group("/auth")
	auth.POST("/register", userHdl.HandleRegister)
	auth.POST("/login", userHdl.HandleLogin)
	auth.POST("/logout", userHdl.HandleLogout, identitymiddleware.Authorize(authenticator))

	users := e.Group("/users")
	users.GET("/search", userHdl.SearchUsers)
	users.GET("/:username", userHdl.GetUser)
	users.GET("/:username/posts", userHdl.GetUserPosts)

	usersAuth := users.Group("", identitymiddleware.Authorize(authenticator))
	usersAuth.PUT("/password", userHdl.HandleChangePassword)
	usersAuth.PUT("/profile", userHdl.HandleUpdateProfile)
	usersAuth.GET("/:username/following", userHdl.GetFollowing)
	usersAuth.GET("/:username/followers", userHdl.GetFollowers)
	usersAuth.POST("/:username/followers", userHdl.HandleFollow)
	usersAuth.DELETE("/:username/followers", userHdl.HandleUnfollow)

	featureFlags := e.Group("/admin/feature-flags", identitymiddleware.Authorize(authenticator), featureflagmiddleware.OperatorOnly(featureFlagOperators))
	featureFlags.GET("", featureFlagHdl.ListFlags)
	featureFlags.POST("", featureFlagHdl.CreateFlag)
	featureFlags.GET("/:key", featureFlagHdl.GetFlag)
	featureFlags.PUT("/:key", featureFlagHdl.UpdateFlag)
	featureFlags.PATCH("/:key/environment/:environment", featureFlagHdl.SetEnvironment)
	featureFlags.PUT("/:key/overrides", featureFlagHdl.SetOverrides)
	featureFlags.POST("/:key/archive", featureFlagHdl.Archive)
	featureFlags.GET("/:key/audit", featureFlagHdl.ListAudit)

	e.GET("/feature-flags/evaluate", featureFlagHdl.Evaluate, identitymiddleware.OptionalAuthorize(authenticator))

	e.GET("/posts/:id", postQueryHdl.GetPostByID)
	posts := e.Group("/posts", identitymiddleware.Authorize(authenticator))
	posts.POST("", postCommandHdl.CreatePost, featureflagmiddleware.FeatureFlagGate(featureFlagEvaluator, "post-creation"))
	posts.GET("", postQueryHdl.GetPosts)
	posts.PUT("/:id", postCommandHdl.UpdatePost)
	posts.DELETE("/:id", postCommandHdl.DeletePost)
	posts.POST("/:id/comments", postCommandHdl.CreateComment)
	posts.POST("/:id/likes", postCommandHdl.LikePost)
	posts.DELETE("/:id/likes", postCommandHdl.UnlikePost)
	posts.GET("/:id/likes", postQueryHdl.GetLikeStatus)

	comments := e.Group("/comments", identitymiddleware.Authorize(authenticator))
	comments.PUT("/:id", commentHdl.EditComment)
	comments.DELETE("/:id", commentHdl.DeleteComment)

	notifications := e.Group("/notifications", identitymiddleware.Authorize(authenticator))
	notifications.GET("", notificationHdl.GetNotifications)
	notifications.GET("/stream", notificationHdl.StreamNotifications)
	notifications.PUT("/:id/read", notificationHdl.MarkAsRead)
	notifications.PUT("/read-all", notificationHdl.MarkAllAsRead)

	feeds := e.Group("/feeds", identitymiddleware.Authorize(authenticator))
	feeds.GET("", feedHdl.HandleGetFeed)
}

// registerMetricsRoute mounts the authenticated Prometheus scrape handler on
// the private Compose network path. It is intentionally absent from the public
// API surface and from the Swagger contract.
func registerMetricsRoute(e *echo.Echo, handler http.Handler) {
	e.GET(operationsmiddleware.SelfScrapeRoute, echo.WrapHandler(handler))
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

	if err := a.sseManager.Close(); err != nil {
		lgr.Error(ctx, "failed to close sse manager", "error", err)
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
