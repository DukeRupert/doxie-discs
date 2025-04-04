package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/dukerupert/doxie-discs/db/models"
)

type RecordHandler struct {
	DB *sql.DB
	RecordService *models.RecordService
}

func NewRecordHandler(db *sql.DB) *RecordHandler {
	return &RecordHandler{
		DB:            db,
		RecordService: models.NewRecordService(db),
	}
}

// GetRecord retrieves a record by ID
func (h *RecordHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	record, err := h.RecordService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify record belongs to user
	if record.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// ListRecords gets all records for the user
func (h *RecordHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	records, err := h.RecordService.ListByUserID(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// CreateRecord adds a new record to the database
func (h *RecordHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	var record models.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set user ID from authenticated user
	record.UserID = userID

	createdRecord, err := h.RecordService.Create(&record)
	if err != nil {
		http.Error(w, "Error creating record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRecord)
}

// UpdateRecord modifies an existing record
func (h *RecordHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	// Check if record exists and belongs to user
	existingRecord, err := h.RecordService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingRecord.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var record models.Record
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set ID and user ID
	record.ID = id
	record.UserID = userID

	updatedRecord, err := h.RecordService.Update(&record)
	if err != nil {
		http.Error(w, "Error updating record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedRecord)
}

// DeleteRecord removes a record from the database
func (h *RecordHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	// Check if record exists and belongs to user
	record, err := h.RecordService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Record not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if record.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.RecordService.Delete(id); err != nil {
		http.Error(w, "Error deleting record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchRecords searches for records based on query parameters
func (h *RecordHandler) SearchRecords(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("userID").(int)

	// Get search parameters from query string
	query := r.URL.Query().Get("q")
	artist := r.URL.Query().Get("artist")
	genre := r.URL.Query().Get("genre")
	label := r.URL.Query().Get("label")
	location := r.URL.Query().Get("location")

	// Call service to search with these params
	records, err := h.RecordService.Search(userID, query, artist, genre, label, location)
	if err != nil {
		http.Error(w, "Error searching records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}