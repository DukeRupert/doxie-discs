// api/handlers/artists.go
package handlers

import (
	"fmt"
	"log"
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

// CreateArtist adds a new artist to the database
func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	var artist models.Artist
	if err := json.NewDecoder(r.Body).Decode(&artist); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	artist.UserID = userID

	createdArtist, err := h.ArtistService.Create(&artist)
	if err != nil {
		errorMsg := fmt.Sprintf("Error creating artist: %v", err)
		log.Println(errorMsg)
		http.Error(w, "Error creating artist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdArtist)
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