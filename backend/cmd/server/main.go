package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	firebase "firebase.google.com/go/v4"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/option"

	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/handler"
	appmw "github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/repository/postgres"
	"github.com/nailx/backend/internal/service/impl"
)

type customValidator struct {
	v *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.v.Struct(i)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// ---- 設定 ----
	dbURL := mustEnv("DATABASE_URL")
	redisURL := mustEnv("REDIS_URL")
	firebaseCredPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	cdnBase := os.Getenv("CDN_BASE_URL")
	port := envOr("PORT", "8080")
	s3Bucket := envOr("S3_BUCKET", "nailx-media-dev")
	sqsQueueURL := os.Getenv("SQS_SIMILARITY_QUEUE_URL") // optional in dev

	// ---- DB接続 ----
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("database ping failed", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// ---- Redis接続 ----
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("invalid redis URL", slog.String("err", err.Error()))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("redis ping failed", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// ---- Firebase初期化 ----
	if firebaseCredPath != "" {
		if _, statErr := os.Stat(firebaseCredPath); statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				absPath, _ := filepath.Abs(firebaseCredPath)
				logger.Error("firebase credentials file not found",
					slog.String("path", firebaseCredPath),
					slog.String("abs_path", absPath),
					slog.String("hint", "set FIREBASE_CREDENTIALS_PATH to a valid service account JSON file"),
				)
			} else {
				logger.Error("cannot access firebase credentials file",
					slog.String("path", firebaseCredPath),
					slog.String("err", statErr.Error()),
				)
			}
			os.Exit(1)
		}
	}

	var fbApp *firebase.App
	if firebaseCredPath != "" {
		fbApp, err = firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(firebaseCredPath))
	} else {
		fbApp, err = firebase.NewApp(context.Background(), nil)
	}
	if err != nil {
		logger.Error("failed to init firebase", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// ---- AWS クライアント（S3 / SQS）----
	awsCfg, awsErr := config.LoadDefaultConfig(context.Background())

	var sqsClient *sqs.Client
	if awsErr == nil && sqsQueueURL != "" {
		sqsClient = sqs.NewFromConfig(awsCfg)
	} else {
		logger.Warn("AWS SQS not configured; similarity worker will be skipped in dev mode")
	}

	// ---- リポジトリ初期化 ----
	userRepo := postgres.NewUserRepository(pool)
	arRepo := postgres.NewARRepository(pool)
	designRepo := postgres.NewDesignRepository(pool)
	salonRepo := postgres.NewSalonRepository(pool)
	auctionRepo := postgres.NewAuctionRepository(pool)
	bookingRepo := postgres.NewBookingRepository(pool)
	paymentRepo := postgres.NewPaymentRepository(pool)
	reviewRepo := postgres.NewReviewRepository(pool)
	notifRepo := postgres.NewNotificationRepository(pool)

	// ---- サービス初期化 ----
	authSvc := impl.NewAuthService(fbApp, userRepo)
	userSvc := impl.NewUserService(userRepo)
	arSvc := impl.NewARService(arRepo)
	designSvc := impl.NewDesignService(designRepo, sqsClient, sqsQueueURL)
	salonSvc := impl.NewSalonService(salonRepo, userRepo)
	auctionSvc := impl.NewAuctionService(auctionRepo, salonRepo, sqsClient, sqsQueueURL)
	bookingSvc := impl.NewBookingService(bookingRepo, auctionRepo, paymentRepo, salonRepo)
	reviewSvc := impl.NewReviewService(reviewRepo, bookingRepo, auctionRepo)
	notifSvc := impl.NewNotificationService(notifRepo)

	mediaSvc, err := impl.NewMediaService(s3Bucket, cdnBase)
	if err != nil {
		logger.Warn("MediaService init failed; presigned URL generation disabled", slog.String("err", err.Error()))
	}

	// ---- Auth Middleware ----
	authMw, err := appmw.NewAuthMiddleware(fbApp, userRepo)
	if err != nil {
		logger.Error("failed to init auth middleware",
			slog.String("err", err.Error()),
			slog.String("hint", "verify FIREBASE_CREDENTIALS_PATH in backend/.env and ensure the JSON file exists inside the container"),
		)
		os.Exit(1)
	}

	// ---- SSE Hub ----
	sseHub := handler.NewSSEHub(rdb, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sseHub.Start(ctx)

	// ---- ハンドラー初期化 ----
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	arH := handler.NewARHandler(arSvc, cdnBase)
	designH := handler.NewDesignHandler(designSvc)
	salonH := handler.NewSalonHandler(salonSvc)
	auctionH := handler.NewAuctionHandler(auctionSvc, sseHub)
	bookingH := handler.NewBookingHandler(bookingSvc)
	reviewH := handler.NewReviewHandler(reviewSvc, cdnBase)
	notifH := handler.NewNotificationHandler(notifSvc)
	adminH := handler.NewAdminHandler(salonSvc, designSvc)
	mediaH := handler.NewMediaHandler(mediaSvc)

	// ---- Echo セットアップ ----
	e := echo.New()
	e.HideBanner = true
	e.Validator = &customValidator{v: validator.New()}

	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.CORS())
	e.Use(echomw.RateLimiter(echomw.NewRateLimiterMemoryStore(100)))

	// ---- ルーティング ----
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")

	// Auth（認証不要）
	v1.POST("/auth/register", authH.Register)

	// Users（要認証）
	users := v1.Group("/users", authMw.RequireAuth)
	users.GET("/me", userH.GetMe)
	users.PATCH("/me", userH.UpdateMe)
	users.PUT("/me/nail-profile", userH.UpsertNailProfile)
	users.GET("/:id", userH.GetPublicProfile)
	users.GET("/:id/designs", designH.ListByCreator)

	// AR
	ar := v1.Group("/ar", authMw.RequireAuth)
	ar.POST("/sessions", arH.CreateSession)
	ar.GET("/sessions/:id", arH.GetSession)

	// Designs
	v1.GET("/designs", designH.ListFeed)
	v1.GET("/designs/:id", designH.GetDesign)
	designs := v1.Group("/designs", authMw.RequireAuth)
	designs.POST("", designH.Register)
	designs.GET("/similarity-check", designH.SimilarityCheck)
	designs.PATCH("/:id", designH.Update)

	// Salons
	v1.GET("/salons", salonH.List)
	v1.GET("/salons/:id", salonH.Get)
	v1.GET("/salons/:id/reviews", reviewH.ListBySalon)
	salons := v1.Group("/salons", authMw.RequireAuth)
	salons.POST("", salonH.Register)
	salons.PATCH("/:id", salonH.Update)
	salons.POST("/:id/portfolio", salonH.AddPortfolio)

	// Auctions
	auctions := v1.Group("/auctions", authMw.RequireAuth)
	auctions.POST("/requests", auctionH.CreateRequest)
	auctions.GET("/requests/:id", auctionH.GetRequest)
	auctions.DELETE("/requests/:id", auctionH.CancelRequest)
	auctions.GET("/requests", auctionH.ListMatchingRequests) // [Salon]
	auctions.POST("/requests/:request_id/bids", auctionH.PlaceBid)
	auctions.PATCH("/requests/:request_id/bids/:bid_id", auctionH.UpdateBid)
	auctions.GET("/requests/:request_id/bids", auctionH.ListBids)
	auctions.GET("/requests/:request_id/bids/stream", sseHub.StreamBids) // SSE

	// Bookings
	bookings := v1.Group("/bookings", authMw.RequireAuth)
	bookings.POST("", bookingH.Confirm)
	bookings.GET("", bookingH.List)
	bookings.GET("/:id", bookingH.Get)
	bookings.POST("/:id/complete", bookingH.Complete)
	bookings.POST("/:id/cancel", bookingH.Cancel)

	// Reviews
	reviews := v1.Group("/reviews", authMw.RequireAuth)
	reviews.POST("", reviewH.Post)

	// Notifications
	notifs := v1.Group("/notifications", authMw.RequireAuth)
	notifs.GET("", notifH.List)
	notifs.PATCH("/:id/read", notifH.MarkRead)
	notifs.POST("/fcm-token", notifH.RegisterFCMToken)

	// Media
	media := v1.Group("/media", authMw.RequireAuth)
	media.POST("/presigned-url", mediaH.GeneratePresignedURL)

	// Admin
	admin := v1.Group("/admin", authMw.RequireAuth, appmw.RequireRole(domain.RoleAdmin))
	admin.GET("/salons", adminH.ListPendingSalons)
	admin.PATCH("/salons/:id/verify", adminH.VerifySalon)
	admin.GET("/designs/flagged", adminH.ListFlaggedDesigns)
	admin.PATCH("/designs/:id/moderate", adminH.ModerateDesign)

	// ---- Graceful Shutdown ----
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("err", err.Error()))
		}
	}()
	logger.Info("server started", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("err", err.Error()))
	}
	logger.Info("server stopped")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable not set", slog.String("key", key))
		os.Exit(1)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
