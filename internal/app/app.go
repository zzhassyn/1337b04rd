package app

import (
	"1337b04rd/internal/adapters/api"
	dbadapter "1337b04rd/internal/adapters/db"
	httphandler "1337b04rd/internal/adapters/http"
	"1337b04rd/internal/adapters/s3"
	"1337b04rd/internal/service"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultPort     = "8080"
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 15 * time.Second
)

func Run() error {
	var (
		port     string
		showHelp bool
	)

	flag.StringVar(&port, "port", defaultPort, "Port to listen on")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		return nil
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	dsn := envOrDefault("DB_DSN", "postgres://postgres:postgres@localhost:5432/1337b04rd?sslmode=disable")
	s3Endpoint := envOrDefault("S3_ENDPOINT", "http://localhost:9000")
	s3PublicEndpoint := envOrDefault("S3_PUBLIC_ENDPOINT", s3Endpoint)
	s3AccessKey := envOrDefault("S3_ACCESS_KEY", "minioadmin")
	s3SecretKey := envOrDefault("S3_SECRET_KEY", "minioadmin")
	s3Region := envOrDefault("S3_REGION", "us-east-1")
	templateDir := envOrDefault("TEMPLATE_DIR", "./template")

	log.Info("connecting to database", "dsn", maskDSN(dsn))

	db, err := dbadapter.Open(dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	migrationSQL, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}

	if err := dbadapter.RunMigrations(db, string(migrationSQL)); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Info("migration applied")

	postRepo := dbadapter.NewPostRepo(db)
	commentRepo := dbadapter.NewCommentRepo(db)
	sessionRepo := dbadapter.NewSessionRepo(db)

	imageStore, err := s3.New(s3.Config{
		Endpoint:        s3Endpoint,
		AccessKeyID:     s3AccessKey,
		SecretAccessKey: s3SecretKey,
		Region:          s3Region,
		PostsBucket:     "posts-images",
		CommentsBucket:  "comments-images",
		PublicEndpoint:  s3PublicEndpoint,
	}, log)
	if err != nil {
		return fmt.Errorf("initialize s3: %w", err)
	}

	avatarSvc := api.NewRickMortyClient(log)

	svc := service.New(postRepo, commentRepo, sessionRepo, imageStore, avatarSvc, log)

	renderer, err := httphandler.NewRenderer(templateDir, log)
	if err != nil {
		return fmt.Errorf("initialize renderer: %w", err)
	}

	router := httphandler.NewRouter(svc, renderer, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go service.RunArchiveWorker(ctx, postRepo, log)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutting down...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Info("server stopped gracefully")

	return nil
}

func printUsage() {
	fmt.Println(`hacker board

Usage:
  1337b04rd [--port <N>]
  1337b04rd --help

Options:
  --help       Show this screen.
  --port N     Port number.`)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func maskDSN(dsn string) string {
	if len(dsn) > 20 {
		return dsn[:15] + "***"
	}

	return "***"
}
