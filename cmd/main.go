package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
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
)
var tokenAuth *jwtauth.JWTAuth

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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
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

	// Initialize JWT auth with secret from environment variable
	jwtSecret := getEnvOrDefault("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal().Msg("JWT_SECRET environment variable must be set")
	}
	tokenAuth = jwtauth.New("HS256", []byte(jwtSecret), nil)
	log.Debug().Msg("JWT authentication initialized")

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
	userHandler := handlers.NewUserHandler(db, tokenAuth)
	artistHandler := handlers.NewArtistHandler(db)
	genreHandler := handlers.NewGenreHandler(db)
	labelHandler := handlers.NewLabelHandler(db)

	// Initialize router with Chi
	r := chi.NewRouter()

	// Custom zerolog middleware
	r.Use(zerologMiddleware)

	// Standard Chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/health"))
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

	workDir, _ := os.Getwd()
	staticDir := http.Dir(filepath.Join(workDir, "static"))
	fileServer := http.FileServer(staticDir)

	// Serve static files
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, fileServer)
		fs.ServeHTTP(w, r)
	})

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
			log.Info().Str("path", "/api/health").Msg("Health check endpoint accessed")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Authentication routes
		r.Post("/api/auth/login", userHandler.Login)
		r.Post("/api/auth/register", userHandler.Register)

		// Serve the login page
		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			log.Debug().Str("path", "/login").Msg("Login page requested")
			http.ServeFile(w, r, filepath.Join(workDir, "static", "login.html"))
		})

		// Create a registration page handler (you'll need to create this HTML file)
		r.Get("/register", func(w http.ResponseWriter, r *http.Request) {
			log.Debug().Str("path", "/register").Msg("Registration page requested")
			http.ServeFile(w, r, filepath.Join(workDir, "static", "register.html"))
		})

		// Serve index page as default
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			log.Debug().Str("path", "/").Msg("Index page requested")
			http.ServeFile(w, r, filepath.Join(workDir, "static", "index.html"))
		})
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		// Apply JWT authentication middleware
		r.Use(jwtauth.Verify(tokenAuth, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie))
		r.Use(jwtAuthenticator)

		// Serve the dashboard page (protected)
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(workDir, "static", "dashboard.html"))
		})

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

	// Expose Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// Start server
	port := getEnvOrDefault("PORT", "8080")
	log.Info().Str("port", port).Msg("Starting server")
	http.ListenAndServe(":"+port, r)
}

// Custom zerolog middleware
func zerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a request ID for tracking
		requestID := middleware.GetReqID(r.Context())

		// Add this request to the context
		logger := log.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_ip", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Logger()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log request completion
		logger.Info().
			Int("status_code", 200). // You might need to use a custom ResponseWriter to capture the real status code
			Dur("duration_ms", time.Since(start)).
			Msg("Request completed")
	})
}

// Custom JWT authenticator middleware with zerolog
func jwtAuthenticator(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token, claims, err := jwtauth.FromContext(r.Context())

        requestID := middleware.GetReqID(r.Context())
        logger := log.With().
            Str("request_id", requestID).
            Str("path", r.URL.Path).
            Logger()

        if err != nil {
            logger.Warn().
                Err(err).
                Msg("JWT authentication error")

            // For API requests, return 401
            if strings.Contains(r.Header.Get("Accept"), "application/json") || 
               strings.HasPrefix(r.URL.Path, "/api/") {
                http.Error(w, "JWT authentication error: "+err.Error(), http.StatusUnauthorized)
                return
            }
            
            // For browser requests to protected pages, redirect to login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        if token == nil {
            logger.Warn().
                Msg("Invalid or missing JWT token")

            // For API requests, return 401
            if strings.Contains(r.Header.Get("Accept"), "application/json") || 
               strings.HasPrefix(r.URL.Path, "/api/") {
                http.Error(w, "Invalid or missing JWT token", http.StatusUnauthorized)
                return
            }
            
            // For browser requests to protected pages, redirect to login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Check if the token contains the user ID
        userIDClaim, ok := claims["user_id"]
        if !ok {
            logger.Warn().
                Msg("JWT token missing required user_id claim")

            http.Error(w, "JWT token missing required user_id claim", http.StatusUnauthorized)
            return
        }

        // Convert userID to string to avoid type conversion issues
        userID := fmt.Sprintf("%v", userIDClaim)

        // Log successful authentication
        logger.Debug().
            Str("user_id", userID).
            Msg("JWT authentication successful")

        // Add user ID to the request context for handlers to use
        ctx := context.WithValue(r.Context(), "userID", userID)

        // Token is authenticated, pass it through
        next.ServeHTTP(w, r.WithContext(ctx))
    })
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
