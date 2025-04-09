// api/handlers/genres.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dukerupert/doxie-discs/db/models"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid record ID")
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	log.Debug().Int("id", id).Int("userID", userID).Msg("Getting genre")

	genre, err := h.GenreService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Genre not found")
			http.Error(w, "Genre not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when fetching genre")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify record belongs to user
	if genre.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("genreUserID", genre.UserID).Msg("Unauthorized attempt to access genre")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Implementation
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genre)
	log.Debug().Int("id", id).Int("userID", userID).Msg("Successfully retrieved genre")
}

func (h *GenreHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	log.Debug().Int("userID", userID).Msg("Listing genres for user")

	genres, err := h.GenreService.ListByUserID(userID)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Error fetching genres")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(genres)
	log.Debug().Int("userID", userID).Int("count", len(genres)).Msg("Successfully listed genres")
}

func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	var genre models.Genre
	if err := json.NewDecoder(r.Body).Decode(&genre); err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Invalid request body for genre creation")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	genre.UserID = userID

	log.Debug().Int("userID", userID).Str("name", genre.Name).Msg("Creating new genre")

	createdGenre, err := h.GenreService.Create(&genre)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Str("name", genre.Name).Msg("Error creating genre")
		http.Error(w, "Error creating genre", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdGenre)
	log.Info().Int("id", createdGenre.ID).Int("userID", userID).Str("name", createdGenre.Name).Msg("Genre created successfully")
}

func (h *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid record ID")
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	log.Debug().Int("id", id).Int("userID", userID).Msg("Updating genre")

	// Check if record exists and belongs to user
	existingGenre, err := h.GenreService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Genre not found for update")
			http.Error(w, "Genre not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking genre existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingGenre.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("genreUserID", existingGenre.UserID).Msg("Unauthorized attempt to update genre")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var genre models.Genre
	if err := json.NewDecoder(r.Body).Decode(&genre); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Invalid request body for genre update")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set ID and user ID
	genre.ID = id
	genre.UserID = userID

	updatedGenre, err := h.GenreService.Update(&genre)
	if err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error updating genre")
		http.Error(w, "Error updating genre", http.StatusInternalServerError)
		return
	}

	// Implementation
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedGenre)
	log.Info().Int("id", id).Int("userID", userID).Str("name", updatedGenre.Name).Msg("Genre updated successfully")
}

func (h *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid genre ID")
		http.Error(w, "Invalid genre ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	log.Debug().Int("id", id).Int("userID", userID).Msg("Deleting genre")

	// Check if record exists and belongs to user
	genre, err := h.GenreService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Genre not found for deletion")
			http.Error(w, "Genre not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking genre existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if genre.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("genreUserID", genre.UserID).Msg("Unauthorized attempt to delete genre")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.GenreService.Delete(id, userID); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error deleting genre")
		http.Error(w, "Error deleting record", http.StatusInternalServerError)
		return
	}

	// Implementation
	w.WriteHeader(http.StatusNoContent)
	log.Info().Int("id", id).Int("userID", userID).Msg("Genre deleted successfully")
}
