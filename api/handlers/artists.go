// api/handlers/artists.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dukerupert/doxie-discs/db/models"
)

type ArtistHandler struct {
	DB           *sql.DB
	ArtistService *models.ArtistService
}

// NewArtistHandler creates a new ArtistHandler with the given db connection
func NewArtistHandler(db *sql.DB) *ArtistHandler {
	return &ArtistHandler{
		DB:           db,
		ArtistService: models.NewArtistService(db),
	}
}

// Basic CRUD methods for ArtistHandler
func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *ArtistHandler) ListArtists(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}

func (h *ArtistHandler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusNoContent)
}

func (h *ArtistHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}