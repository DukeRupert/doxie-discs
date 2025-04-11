// api/handlers/labels.go
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

type LabelHandler struct {
	DB          *sql.DB
	LabelService *models.LabelService
}

// NewLabelHandler creates a new LabelHandler with the given db connection
func NewLabelHandler(db *sql.DB) *LabelHandler {
	return &LabelHandler{
		DB:          db,
		LabelService: models.NewLabelService(db),
	}
}

// Basic CRUD methods for LabelHandler
func (h *LabelHandler) GetLabel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid label ID")
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

		// Get user ID from context (set by auth middleware)
		userID, err := GetUserIDFromContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to retrieve userID from context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	
	log.Debug().Int("id", id).Int("userID", userID).Msg("Getting label")

	label, err := h.LabelService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Label not found")
			http.Error(w, "Label not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when fetching label")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify label belongs to user
	if label.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("labelUserID", label.UserID).Msg("Unauthorized attempt to access label")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(label)
	log.Debug().Int("id", id).Int("userID", userID).Msg("Successfully retrieved label")
}

func (h *LabelHandler) ListLabels(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by auth middleware)
		userID, err := GetUserIDFromContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to retrieve userID from context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	
	log.Debug().Int("userID", userID).Msg("Listing labels for user")

	labels, err := h.LabelService.ListByUserID(userID)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Error fetching labels")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(labels)
	log.Debug().Int("userID", userID).Int("count", len(labels)).Msg("Successfully listed labels")
}

func (h *LabelHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by auth middleware)
		userID, err := GetUserIDFromContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to retrieve userID from context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

	var label models.Label
	if err := json.NewDecoder(r.Body).Decode(&label); err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Invalid request body for label creation")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	label.UserID = userID
	
	log.Debug().Int("userID", userID).Str("name", label.Name).Msg("Creating new label")

	createdLabel, err := h.LabelService.Create(&label)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Str("name", label.Name).Msg("Error creating label")
		http.Error(w, "Error creating label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdLabel)
	log.Info().Int("id", createdLabel.ID).Int("userID", userID).Str("name", createdLabel.Name).Msg("Label created successfully")
}

func (h *LabelHandler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid label ID")
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

		// Get user ID from context (set by auth middleware)
		userID, err := GetUserIDFromContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to retrieve userID from context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	
	log.Debug().Int("id", id).Int("userID", userID).Msg("Updating label")

	// Check if label exists and belongs to user
	existingLabel, err := h.LabelService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Label not found for update")
			http.Error(w, "Label not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking label existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingLabel.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("labelUserID", existingLabel.UserID).Msg("Unauthorized attempt to update label")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var label models.Label
	if err := json.NewDecoder(r.Body).Decode(&label); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Invalid request body for label update")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set ID and user ID
	label.ID = id
	label.UserID = userID

	updatedLabel, err := h.LabelService.Update(&label)
	if err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error updating label")
		http.Error(w, "Error updating label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLabel)
	log.Info().Int("id", id).Int("userID", userID).Str("name", updatedLabel.Name).Msg("Label updated successfully")
}

func (h *LabelHandler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Invalid label ID")
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

		// Get user ID from context (set by auth middleware)
		userID, err := GetUserIDFromContext(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to retrieve userID from context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	
	log.Debug().Int("id", id).Int("userID", userID).Msg("Deleting label")

	// Check if label exists and belongs to user
	label, err := h.LabelService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Label not found for deletion")
			http.Error(w, "Label not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking label existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if label.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("labelUserID", label.UserID).Msg("Unauthorized attempt to delete label")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.LabelService.Delete(id, userID); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error deleting label")
		http.Error(w, "Error deleting label", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info().Int("id", id).Int("userID", userID).Msg("Label deleted successfully")
}