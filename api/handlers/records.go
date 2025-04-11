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

type RecordHandler struct {
	DB            *sql.DB
	RecordService *models.RecordService
}

func NewRecordHandler(db *sql.DB) *RecordHandler {
	return &RecordHandler{
		DB:            db,
		RecordService: models.NewRecordService(db),
	}
}

func (h *RecordHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
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

	log.Debug().Int("id", id).Int("userID", userID).Msg("Getting record")

	record, err := h.RecordService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Record not found")
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when fetching record")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify record belongs to user
	if record.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("recordUserID", record.UserID).Msg("Unauthorized attempt to access record")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
	log.Debug().Int("id", id).Int("userID", userID).Msg("Successfully retrieved record")
}

// ListRecords gets all records for the user
func (h *RecordHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Debug().Int("userID", userID).Msg("Listing records for user")

	records, err := h.RecordService.ListByUserID(userID)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Error fetching records")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(records)).Msg("Records fetched")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
	log.Debug().Int("userID", userID).Int("count", len(records)).Msg("Successfully listed records")
}

// CreateRecord adds a new record to the database
func (h *RecordHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var record models.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		log.Error().Err(err).Int("userID", userID).Msg("Invalid request body for record creation")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	record.UserID = userID

	log.Debug().Int("userID", userID).Str("title", record.Title).Msg("Creating new record")

	createdRecord, err := h.RecordService.Create(&record)
	if err != nil {
		log.Error().Err(err).Int("userID", userID).Str("title", record.Title).Msg("Error creating record")
		http.Error(w, "Error creating record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRecord)
	log.Info().Int("id", createdRecord.ID).Int("userID", userID).Str("title", createdRecord.Title).Msg("Record created successfully")
}

// UpdateRecord modifies an existing record
func (h *RecordHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
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

	log.Debug().Int("id", id).Int("userID", userID).Msg("Updating record")

	// Check if record exists and belongs to user
	existingRecord, err := h.RecordService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Record not found for update")
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking record existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingRecord.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("recordUserID", existingRecord.UserID).Msg("Unauthorized attempt to update record")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var record models.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Invalid request body for record update")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set ID and user ID
	record.ID = id
	record.UserID = userID

	updatedRecord, err := h.RecordService.Update(&record)
	if err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error updating record")
		http.Error(w, "Error updating record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedRecord)
	log.Info().Int("id", id).Int("userID", userID).Str("title", updatedRecord.Title).Msg("Record updated successfully")
}

// DeleteRecord removes a record from the database
func (h *RecordHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
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

	log.Debug().Int("id", id).Int("userID", userID).Msg("Deleting record")

	// Check if record exists and belongs to user
	record, err := h.RecordService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Int("id", id).Int("userID", userID).Msg("Record not found for deletion")
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Database error when checking record existence")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if record.UserID != userID {
		log.Warn().Int("id", id).Int("userID", userID).Int("recordUserID", record.UserID).Msg("Unauthorized attempt to delete record")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.RecordService.Delete(id); err != nil {
		log.Error().Err(err).Int("id", id).Int("userID", userID).Msg("Error deleting record")
		http.Error(w, "Error deleting record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info().Int("id", id).Int("userID", userID).Msg("Record deleted successfully")
}

// SearchRecords searches for records based on query parameters
func (h *RecordHandler) SearchRecords(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get search parameters from query string
	query := r.URL.Query().Get("q")
	artist := r.URL.Query().Get("artist")
	genre := r.URL.Query().Get("genre")
	label := r.URL.Query().Get("label")
	location := r.URL.Query().Get("location")

	log.Debug().
		Int("userID", userID).
		Str("query", query).
		Str("artist", artist).
		Str("genre", genre).
		Str("label", label).
		Str("location", location).
		Msg("Searching records")

	// Call service to search with these params
	records, err := h.RecordService.Search(userID, query, artist, genre, label, location)
	if err != nil {
		log.Error().Err(err).
			Int("userID", userID).
			Str("query", query).
			Str("artist", artist).
			Str("genre", genre).
			Str("label", label).
			Str("location", location).
			Msg("Error searching records")
		http.Error(w, "Error searching records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
	log.Debug().
		Int("userID", userID).
		Int("count", len(records)).
		Str("query", query).
		Msg("Successfully completed record search")
}
