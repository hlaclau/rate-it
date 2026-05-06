package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	adaptercache "github.com/hlaclau/rate-it-api/internal/adapter/cache"
	adaptortmdb "github.com/hlaclau/rate-it-api/internal/adapter/tmdb"
	"github.com/hlaclau/rate-it-api/internal/handler"
	"github.com/hlaclau/rate-it-api/internal/repository"
	"github.com/hlaclau/rate-it-api/internal/usecase"
	"github.com/hlaclau/rate-it-api/pkg/database"
	pkgredis "github.com/hlaclau/rate-it-api/pkg/redis"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using environment variables")
	}

	dsn := mustEnv("DATABASE_URL")
	redisURL := mustEnv("REDIS_URL")
	tmdbKey := mustEnv("TMDB_API_KEY")

	db, err := database.Open(dsn)
	if err != nil {
		slog.Error("connect db", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected")

	if err = database.Migrate(db, "file://migrations"); err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	redisClient, err := pkgredis.New(redisURL)
	if err != nil {
		slog.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("redis connected")

	userRepo := repository.NewUserRepository(db)
	sessionCache := adaptercache.NewSessionCache(redisClient)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionCache)
	authHandler := handler.NewAuthHandler(authUseCase)

	tmdbFetcher := adaptortmdb.New(tmdbKey)
	mediaRepo := repository.NewMediaRepository(db)
	mediaHandler := handler.NewMediaHandler(
		usecase.NewMediaUseCase(
			mediaRepo,
			tmdbFetcher,
			adaptercache.NewRedisCache(redisClient),
		),
	)

	listHandler := handler.NewListHandler(
		usecase.NewListUseCase(
			repository.NewUserMediaRepository(db),
			mediaRepo,
			tmdbFetcher,
		),
	)

	allowedOrigins := envOr("ALLOWED_ORIGINS", "http://localhost:3000")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(allowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Cookie"},
		AllowCredentials: true,
	}))

	r.Use(handler.AuthMiddleware(authUseCase))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		authHandler.Routes(r)
		mediaHandler.Routes(r)
		listHandler.Routes(r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("starting server", "port", port, "allowed_origins", allowedOrigins)
	if err = http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required env var missing", "key", key)
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
