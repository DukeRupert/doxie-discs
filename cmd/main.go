// cmd/main.go
package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dukerupert/doxie-discs/api/handlers"
	"github.com/dukerupert/doxie-discs/middleware/auth"
)

// Define Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)

	// Configure zerolog
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(zerolog.InfoLevel)
}

// LoggerMiddleware creates a zerolog middleware for Chi
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use Chi's middleware to capture response data
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(ww, r)

		// Calculate duration
		duration := time.Since(start)

		// Record metrics
		statusCode := ww.Status()
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())

		// Log the request
		log.Info().
			Str("method", r.Method).
			Str("url", r.URL.Path).
			Int("status", statusCode).
			Dur("duration", duration).
			Int("size", ww.BytesWritten()).
			Str("remote", r.RemoteAddr).
			Msg("HTTP request")
	})
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg("No .env file found, using environment variables")
	}

	// Get database connection details from environment or use defaults
	dbUser := getEnvOrDefault("POSTGRES_USER", "postgres")
	dbPassword := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	dbHost := getEnvOrDefault("POSTGRES_HOST", "localhost") // Uses 'localhost' since we're using network_mode: service:db
	dbPort := getEnvOrDefault("POSTGRES_PORT", "5432")
	dbName := getEnvOrDefault("POSTGRES_DB", "dev")

	// Database connection string
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to connect to database: %v\n", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to ping database: %v\n", err)
	}

	log.Info().Msg("Successfully connected to the database")

	// Run database migrations
	log.Info().Msg("Running database migrations...")
	err = runMigrations(db)
	if err != nil {
		log.Fatal().Err(err).Msgf("Migration failed: %v\n", err)
	}
	log.Info().Msg("Migrations completed successfully")
	defer db.Close()

	// Initialize handlers
	recordHandler := handlers.NewRecordHandler(db)
	userHandler := handlers.NewUserHandler(db)
	artistHandler := handlers.NewArtistHandler(db)
	genreHandler := handlers.NewGenreHandler(db)
	labelHandler := handlers.NewLabelHandler(db)

	// Initialize router with Chi
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Adjust this for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Authentication routes
		r.Post("/api/auth/login", userHandler.Login)
		r.Post("/api/auth/register", userHandler.Register)
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		// Apply authentication middleware
		r.Use(auth.Middleware)

		// Record routes
		r.Route("/api/records", func(r chi.Router) {
			r.Get("/", recordHandler.ListRecords)
			r.Post("/", recordHandler.CreateRecord)
			r.Get("/search", recordHandler.SearchRecords)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", recordHandler.GetRecord)
				r.Put("/", recordHandler.UpdateRecord)
				r.Delete("/", recordHandler.DeleteRecord)
			})
		})

		// Artist routes
		r.Route("/api/artists", func(r chi.Router) {
			r.Get("/", artistHandler.ListArtists)
			r.Post("/", artistHandler.CreateArtist)
			r.Get("/search", artistHandler.SearchArtists)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", artistHandler.GetArtist)
				r.Put("/", artistHandler.UpdateArtist)
				r.Delete("/", artistHandler.DeleteArtist)
			})
		})

		// Genre routes
		r.Route("/api/genres", func(r chi.Router) {
			r.Get("/", genreHandler.ListGenres)
			r.Post("/", genreHandler.CreateGenre)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", genreHandler.GetGenre)
				r.Put("/", genreHandler.UpdateGenre)
				r.Delete("/", genreHandler.DeleteGenre)
			})
		})

		// Label routes
		r.Route("/api/labels", func(r chi.Router) {
			r.Get("/", labelHandler.ListLabels)
			r.Post("/", labelHandler.CreateLabel)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", labelHandler.GetLabel)
				r.Put("/", labelHandler.UpdateLabel)
				r.Delete("/", labelHandler.DeleteLabel)
			})
		})

		// User routes (for profile, etc.)
		r.Route("/api/users", func(r chi.Router) {
			r.Get("/me", userHandler.GetProfile)
			r.Put("/me", userHandler.UpdateProfile)
			r.Put("/password", userHandler.UpdatePassword)
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Expose Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// Start server
	log.Info().Msg("Starting server on :"+port)
	http.ListenAndServe(":"+port, r)
}

// Helper function to get environment variable or return default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	// Path to migration files
	migrationsPath := "file://db/migrations"

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}

	// Run migrations up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	return nil
}
