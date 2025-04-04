// cmd/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	
	"github.com/dukerupert/doxie-discs/api/handlers"
	"github.com/dukerupert/doxie-discs/middleware/auth"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database connection string
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_DATABASE"),
	)

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Successfully connected to database")

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
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
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