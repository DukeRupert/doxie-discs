// main.go
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
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database connection string
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
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

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v\n", err)
	}
	log.Println("Migrations completed successfully")

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

	// Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		
		// TODO: Add more routes for your CRUD operations
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

// db/models/user.go
package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Don't expose in JSON
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Simplified user service example
type UserService struct {
	DB *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{DB: db}
}

func (s *UserService) GetByID(id int) (*User, error) {
	user := &User{}
	err := s.DB.QueryRow(
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// db/models/record.go
package models

import (
	"database/sql"
	"time"
)

type Record struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	ReleaseYear     int       `json:"release_year,omitempty"`
	CatalogNumber   string    `json:"catalog_number,omitempty"`
	Condition       string    `json:"condition,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CoverImageURL   string    `json:"cover_image_url,omitempty"`
	StorageLocation string    `json:"storage_location,omitempty"`
	UserID          int       `json:"user_id"`
	LabelID         int       `json:"label_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	
	// Related entities
	Artists []Artist `json:"artists,omitempty"`
	Genres  []Genre  `json:"genres,omitempty"`
	Tracks  []Track  `json:"tracks,omitempty"`
}

type RecordService struct {
	DB *sql.DB
}

func NewRecordService(db *sql.DB) *RecordService {
	return &RecordService{DB: db}
}

func (s *RecordService) GetByID(id int) (*Record, error) {
	// Implementation details...
	return nil, nil
}

func (s *RecordService) ListByUserID(userID int) ([]Record, error) {
	// Implementation details...
	return nil, nil
}

// Add more models (Artist, Label, Genre, Track) with similar structure