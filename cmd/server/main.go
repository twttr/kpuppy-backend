package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/twttr/kpuppy-backend/admin"
	"github.com/twttr/kpuppy-backend/config"
	"github.com/twttr/kpuppy-backend/internal/handler"
	custommw "github.com/twttr/kpuppy-backend/internal/middleware"
	"github.com/twttr/kpuppy-backend/internal/repository/sqlite"
	"github.com/twttr/kpuppy-backend/internal/usecase"
	"github.com/twttr/kpuppy-backend/internal/websocket"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	if cfg.Sentry.DSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Sentry.DSN,
			Environment:      cfg.Sentry.Environment,
			TracesSampleRate: 0.1,
		})
		if err != nil {
			logger.Error("failed to initialize sentry", "error", err)
		} else {
			logger.Info("sentry initialized", "environment", cfg.Sentry.Environment)
			defer sentry.Flush(2 * time.Second)
		}
	}

	db, err := sqlite.NewDB(cfg.Database.Path)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	userRepo := sqlite.NewUserRepository(db)
	contentRepo := sqlite.NewContentRepository(db)
	commentRepo := sqlite.NewCommentRepository(db)

	userService := usecase.NewUserService(userRepo)
	commentService := usecase.NewCommentService(commentRepo, contentRepo, userRepo)

	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Stop() // Fix #1: stop hub goroutine on shutdown

	e := echo.New()
	e.HideBanner = true

	e.Use(custommw.RequestID())
	e.Use(custommw.SlogLogger(logger))
	if cfg.Sentry.DSN != "" {
		e.Use(custommw.SentryCapture())
		e.Use(custommw.SentryRecover())
	} else {
		e.Use(middleware.Recover())
	}
	e.Use(middleware.CORS())

	if cfg.RateLimit.Enabled {
		rateLimiters := custommw.CreateRateLimiters()
		e.Use(rateLimiters.Middleware())
	}

	userHandler := handler.NewUserHandler(userService)
	commentHandler := handler.NewCommentHandler(commentService, hub)
	wsHandler := handler.NewWSHandler(hub, cfg.Server.AllowedOrigins) // Fix #4: pass allowed origins

	e.POST("/users/provision", userHandler.Provision)

	e.GET("/content/:kinopubItemId/comments", commentHandler.GetComments)
	e.POST("/content/:kinopubItemId/comments", commentHandler.CreateComment)
	e.POST("/comments/:commentId/reply", commentHandler.ReplyToComment)
	e.PATCH("/comments/:commentId", commentHandler.UpdateComment)
	e.DELETE("/comments/:commentId", commentHandler.DeleteComment)

	e.GET("/ws/content/:kinopubItemId", wsHandler.HandleWebSocket)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/admin")
	})

	if cfg.Admin.Username != "" && cfg.Admin.PasswordHash != "" {
		adminAuth := custommw.NewAdminAuth(cfg.Admin.Username, cfg.Admin.PasswordHash)
		adminWebHandler := handler.NewAdminWebHandler(userService, commentService, admin.TemplateFS, cfg.Server.BasePath)

		adminGroup := e.Group("/admin", adminAuth.Middleware())
		adminGroup.GET("", adminWebHandler.Dashboard)
		adminGroup.GET("/", adminWebHandler.Dashboard)
		adminGroup.GET("/comments", adminWebHandler.Comments)
		adminGroup.GET("/users", adminWebHandler.Users)
		adminGroup.POST("/api/comments/:id/delete", adminWebHandler.DeleteComment)
		adminGroup.POST("/api/comments/:id/toggle-spoiler", adminWebHandler.ToggleSpoiler)
		adminGroup.POST("/api/users/:id/ban", adminWebHandler.BanUser)
		adminGroup.POST("/api/users/:id/unban", adminWebHandler.UnbanUser)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", addr)
		if err := e.StartServer(server); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("server stopped")
}
