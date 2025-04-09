// api/handlers/genres.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dukerupert/doxie-discs/db/models"
)

type GenreHandler struct {
	DB           *sql.DB
	GenreService *models.GenreService
}

// NewGenreHandler creates a new GenreHandler with the given db connection
func NewGenreHandler(db *sql.DB) *GenreHandler {
	return &GenreHandler{
		DB:           db,
		GenreService: models.NewGenreService(db),
	}
}

// Basic CRUD methods for GenreHandler
func (h *GenreHandler) GetGenre(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *GenreHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	var genre models.Genre
	if err := json.NewDecoder(r.Body).Decode(&genre); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	genre.UserID = userID

	createdGenre, err := h.GenreService.Create(&genre)
	if err != nil {
		errorMsg := fmt.Sprintf("Error creating genre: %v", err)
		log.Println(errorMsg)
		http.Error(w, "Error creating genre", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdGenre)
}

func (h *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusNoContent)
}
