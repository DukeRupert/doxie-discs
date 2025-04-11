// api/handlers/artists.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/dukerupert/doxie-discs/db/models"
)

type ArtistHandler struct {
	DB            *sql.DB
	ArtistService *models.ArtistService
}

// NewArtistHandler creates a new ArtistHandler with the given db connection
func NewArtistHandler(db *sql.DB) *ArtistHandler {
	return &ArtistHandler{
		DB:            db,
		ArtistService: models.NewArtistService(db),
	}
}

// Basic CRUD methods for ArtistHandler
func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid record ID")
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Debug().Int("id", id).Int("userID", userID).Msg("Getting artist")

	artist, err := h.ArtistService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Artist not found")
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when fetching artist")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify record belongs to user
	if artist.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("genreUserID", artist.UserID).Msg("Unauthorized attempt to access artist")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Implementation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artist)
	log.Debug().Int("id", id).Int("userID", userID).Msg("Successfully retrieved artist")
}

func (h *ArtistHandler) ListArtists(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Debug().Int("userID", userID).Msg("Listing artists for user")

	artists, err := h.ArtistService.ListByUserID(userID)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Error fetching artists")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Implementation
	json.NewEncoder(w).Encode(artists)
	log.Debug().Int("userID", userID).Int("count", len(artists)).Msg("Successfully listed genres")
}

// CreateArtist adds a new artist to the database
func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var artist models.Artist
	if err := json.NewDecoder(r.Body).Decode(&artist); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	artist.UserID = userID

	createdArtist, err := h.ArtistService.Create(&artist)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Str("name", artist.Name).Msg("Error creating genre")
		http.Error(w, "Error creating artist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdArtist)
	log.Info().Int("id", createdArtist.ID).Int("userID", userID).Str("name", createdArtist.Name).Msg("Genre created successfully")
}

func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid artist ID")
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Debug().Int("id", id).Int("userID", userID).Msg("Updating artist")

	// Check if record exists and belongs to user
	existingArtist, err := h.ArtistService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Artist not found for update")
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking artist existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingArtist.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("genreUserID", existingArtist.UserID).Msg("Unauthorized attempt to update genre")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var artist models.Artist
	if err := json.NewDecoder(r.Body).Decode(&artist); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Invalid request body for artist update")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set ID and user ID
	artist.ID = id
	artist.UserID = userID

	updatedArtist, err := h.ArtistService.Update(&artist)
	if err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error updating artist")
		http.Error(w, "Error updating artist", http.StatusInternalServerError)
		return
	}

	// Implementation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedArtist)
	log.Info().Int("id", id).Int("userID", userID).Str("name", updatedArtist.Name).Msg("Genre updated successfully")
}

func (h *ArtistHandler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid artist ID")
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Debug().Int("id", id).Int("userID", userID).Msg("Deleting artist")

	// Check if record exists and belongs to user
	artist, err := h.ArtistService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Artist not found for deletion")
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking artist existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if artist.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("genreUserID", artist.UserID).Msg("Unauthorized attempt to delete artist")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.ArtistService.Delete(id, userID); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error deleting artist")
		http.Error(w, "Error deleting artist", http.StatusInternalServerError)
		return
	}

	// Implementation
	w.WriteHeader(http.StatusNoContent)
	log.Info().Int("id", id).Int("userID", userID).Msg("Artist deleted successfully")
}

func (h *ArtistHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	// Implementation
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "not implemented"})
}
