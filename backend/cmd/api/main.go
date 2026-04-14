package main

import (
	"log"
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
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	dsn := mustEnv("DATABASE_URL")
	redisURL := mustEnv("REDIS_URL")
	tmdbKey := mustEnv("TMDB_API_KEY")

	db, err := database.Open(dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err = database.Migrate(db, "file://migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations applied")

	redisClient, err := pkgredis.New(redisURL)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	mediaHandler := handler.NewMediaHandler(
		usecase.NewMediaUseCase(
			repository.NewMediaRepository(db),
			adaptortmdb.New(tmdbKey),
			adaptercache.NewRedisCache(redisClient),
		),
	)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: strings.Split(envOr("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mediaHandler.Routes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server on :%s", port)
	if err = http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
