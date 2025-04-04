// api/handlers/genres.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dukerupert/doxie-discs/db/models"
)

type GenreHandler struct {
	DB          *sql.DB
	GenreService *models.GenreService
}

// NewGenreHandler creates a new GenreHandler with the given db connection
func NewGenreHandler(db *sql.DB) *GenreHandler {
	return &GenreHandler{
		DB:          db,
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
	// Implementation
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
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