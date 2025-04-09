// db/models/record.go
package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	Label   *Label   `json:"label,omitempty"`
}

type RecordService struct {
	DB *sql.DB
}

func NewRecordService(db *sql.DB) *RecordService {
	return &RecordService{DB: db}
}

// GetByID retrieves a record by its ID
func (s *RecordService) GetByID(id int) (*Record, error) {
	query := `
		SELECT id, title, release_year, catalog_number, condition, notes, 
		       cover_image_url, storage_location, user_id, label_id, 
		       created_at, updated_at
		FROM records
		WHERE id = $1
	`
	
	var record Record
	var releaseYear, labelID sql.NullInt32
	var catalogNumber, condition, notes, coverImageURL, storageLocation sql.NullString
	
	err := s.DB.QueryRow(query, id).Scan(
		&record.ID,
		&record.Title,
		&releaseYear,
		&catalogNumber,
		&condition,
		&notes,
		&coverImageURL,
		&storageLocation,
		&record.UserID,
		&labelID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Handle nullable fields
	if releaseYear.Valid {
		record.ReleaseYear = int(releaseYear.Int32)
	}
	if catalogNumber.Valid {
		record.CatalogNumber = catalogNumber.String
	}
	if condition.Valid {
		record.Condition = condition.String
	}
	if notes.Valid {
		record.Notes = notes.String
	}
	if coverImageURL.Valid {
		record.CoverImageURL = coverImageURL.String
	}
	if storageLocation.Valid {
		record.StorageLocation = storageLocation.String
	}
	if labelID.Valid {
		record.LabelID = int(labelID.Int32)
	}
	
	// Load related entities
	if err := s.loadArtists(&record); err != nil {
		return nil, err
	}
	if err := s.loadGenres(&record); err != nil {
		return nil, err
	}
	if err := s.loadTracks(&record); err != nil {
		return nil, err
	}
	if labelID.Valid {
		if err := s.loadLabel(&record); err != nil {
			return nil, err
		}
	}
	
	return &record, nil
}

// ListByUserID retrieves all records for a specific user
func (s *RecordService) ListByUserID(userID int) ([]Record, error) {
	query := `
		SELECT id, title, release_year, catalog_number, condition, notes, 
		       cover_image_url, storage_location, user_id, label_id, 
		       created_at, updated_at
		FROM records
		WHERE user_id = $1
		ORDER BY title ASC
	`
	
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []Record
	
	for rows.Next() {
		var record Record
		var releaseYear, labelID sql.NullInt32
		var catalogNumber, condition, notes, coverImageURL, storageLocation sql.NullString
		
		err := rows.Scan(
			&record.ID,
			&record.Title,
			&releaseYear,
			&catalogNumber,
			&condition,
			&notes,
			&coverImageURL,
			&storageLocation,
			&record.UserID,
			&labelID,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		// Handle nullable fields
		if releaseYear.Valid {
			record.ReleaseYear = int(releaseYear.Int32)
		}
		if catalogNumber.Valid {
			record.CatalogNumber = catalogNumber.String
		}
		if condition.Valid {
			record.Condition = condition.String
		}
		if notes.Valid {
			record.Notes = notes.String
		}
		if coverImageURL.Valid {
			record.CoverImageURL = coverImageURL.String
		}
		if storageLocation.Valid {
			record.StorageLocation = storageLocation.String
		}
		if labelID.Valid {
			record.LabelID = int(labelID.Int32)
		}
		
		records = append(records, record)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// Load related entities for each record
	for i := range records {
		if err := s.loadArtists(&records[i]); err != nil {
			return nil, err
		}
		if err := s.loadGenres(&records[i]); err != nil {
			return nil, err
		}
		
		// Only load tracks for a list query if specifically requested
		// For performance reasons we don't load them by default
		
		if records[i].LabelID != 0 {
			if err := s.loadLabel(&records[i]); err != nil {
				return nil, err
			}
		}
	}
	
	return records, nil
}

// Helper methods to load related entities

// loadArtists loads artists for a record
func (s *RecordService) loadArtists(record *Record) error {
	query := `
		SELECT a.id, a.name, a.description, a.user_id, ar.role
		FROM artists a
		JOIN artist_record ar ON a.id = ar.artist_id
		WHERE ar.record_id = $1
		ORDER BY a.name ASC
	`
	
	rows, err := s.DB.Query(query, record.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var artist Artist
		var description sql.NullString
		var role sql.NullString
		
		err := rows.Scan(
			&artist.ID,
			&artist.Name,
			&description,
			&artist.UserID,
			&role,
		)
		
		if err != nil {
			return err
		}
		
		if description.Valid {
			artist.Description = description.String
		}
		
		if role.Valid {
			artist.Role = role.String
		}
		
		record.Artists = append(record.Artists, artist)
	}
	
	return rows.Err()
}

// loadGenres loads genres for a record
func (s *RecordService) loadGenres(record *Record) error {
	query := `
		SELECT g.id, g.name, g.description, g.user_id
		FROM genres g
		JOIN genre_record gr ON g.id = gr.genre_id
		WHERE gr.record_id = $1
		ORDER BY g.name ASC
	`
	
	rows, err := s.DB.Query(query, record.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var genre Genre
		var description sql.NullString
		
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&description,
			&genre.UserID,
		)
		
		if err != nil {
			return err
		}
		
		if description.Valid {
			genre.Description = description.String
		}
		
		record.Genres = append(record.Genres, genre)
	}
	
	return rows.Err()
}

// loadTracks loads tracks for a record
func (s *RecordService) loadTracks(record *Record) error {
	query := `
		SELECT id, title, duration, position, created_at, updated_at
		FROM tracks
		WHERE record_id = $1
		ORDER BY position ASC
	`
	
	rows, err := s.DB.Query(query, record.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var track Track
		var duration, position sql.NullString
		
		err := rows.Scan(
			&track.ID,
			&track.Title,
			&duration,
			&position,
			&track.CreatedAt,
			&track.UpdatedAt,
		)
		
		if err != nil {
			return err
		}
		
		if duration.Valid {
			track.Duration = duration.String
		}
		
		if position.Valid {
			track.Position = position.String
		}
		
		track.RecordID = record.ID
		record.Tracks = append(record.Tracks, track)
	}
	
	return rows.Err()
}

// loadLabel loads the label for a record
func (s *RecordService) loadLabel(record *Record) error {
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM labels
		WHERE id = $1
	`
	
	var label Label
	var description sql.NullString
	
	err := s.DB.QueryRow(query, record.LabelID).Scan(
		&label.ID,
		&label.Name,
		&description,
		&label.UserID,
		&label.CreatedAt,
		&label.UpdatedAt,
	)
	
	if err != nil {
		return err
	}
	
	if description.Valid {
		label.Description = description.String
	}
	
	record.Label = &label
	return nil
}

// Create adds a new record to the database
func (s *RecordService) Create(record *Record) (*Record, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Check if the artists exist
	if len(record.Artists) > 0 {
		for _, artist := range record.Artists {
			var exists bool
			err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM artists WHERE id = $1)", artist.ID).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("error checking if artist exists: %w", err)
			}
			if !exists {
				return nil, fmt.Errorf("artist with ID %d does not exist", artist.ID)
			}
		}
	}
	
	// Check if the genres exist
	if len(record.Genres) > 0 {
		for _, genre := range record.Genres {
			var exists bool
			err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM genres WHERE id = $1)", genre.ID).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("error checking if genre exists: %w", err)
			}
			if !exists {
				return nil, fmt.Errorf("genre with ID %d does not exist", genre.ID)
			}
		}
	}
	
	// Check if label exists if label_id is provided
	if record.LabelID != 0 {
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM labels WHERE id = $1)", record.LabelID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("error checking if label exists: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("label with ID %d does not exist", record.LabelID)
		}
	}
	
	query := `
		INSERT INTO records (
			title, release_year, catalog_number, condition, notes, 
			cover_image_url, storage_location, user_id, label_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	
	var labelID interface{}
	if record.LabelID != 0 {
		labelID = record.LabelID
	} else {
		labelID = nil
	}
	
	err = tx.QueryRow(
		query,
		record.Title,
		sql.NullInt32{Int32: int32(record.ReleaseYear), Valid: record.ReleaseYear != 0},
		sql.NullString{String: record.CatalogNumber, Valid: record.CatalogNumber != ""},
		sql.NullString{String: record.Condition, Valid: record.Condition != ""},
		sql.NullString{String: record.Notes, Valid: record.Notes != ""},
		sql.NullString{String: record.CoverImageURL, Valid: record.CoverImageURL != ""},
		sql.NullString{String: record.StorageLocation, Valid: record.StorageLocation != ""},
		record.UserID,
		labelID,
	).Scan(
		&record.ID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to insert record: %w", err)
	}
	
	// Add artists if provided
	if len(record.Artists) > 0 {
		artistStmt, err := tx.Prepare(
			"INSERT INTO artist_record (artist_id, record_id, role) VALUES ($1, $2, $3)")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare artist statement: %w", err)
		}
		defer artistStmt.Close()
		
		for _, artist := range record.Artists {
			_, err = artistStmt.Exec(artist.ID, record.ID, artist.Role)
			if err != nil {
				return nil, fmt.Errorf("failed to insert artist association: %w", err)
			}
		}
	}
	
	// Add genres if provided
	if len(record.Genres) > 0 {
		genreStmt, err := tx.Prepare(
			"INSERT INTO genre_record (genre_id, record_id) VALUES ($1, $2)")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare genre statement: %w", err)
		}
		defer genreStmt.Close()
		
		for _, genre := range record.Genres {
			_, err = genreStmt.Exec(genre.ID, record.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to insert genre association: %w", err)
			}
		}
	}
	
	// Add tracks if provided
	if len(record.Tracks) > 0 {
		trackStmt, err := tx.Prepare(
			`INSERT INTO tracks (title, duration, position, record_id) 
			 VALUES ($1, $2, $3, $4) RETURNING id`)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare track statement: %w", err)
		}
		defer trackStmt.Close()
		
		for i := range record.Tracks {
			var trackID int
			err = trackStmt.QueryRow(
				record.Tracks[i].Title,
				record.Tracks[i].Duration,
				record.Tracks[i].Position,
				record.ID,
			).Scan(&trackID)
			
			if err != nil {
				return nil, fmt.Errorf("failed to insert track: %w", err)
			}
			record.Tracks[i].ID = trackID
		}
	}
	
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return record, nil
}

// Update modifies an existing record
func (s *RecordService) Update(record *Record) (*Record, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	query := `
		UPDATE records
		SET title = $1,
			release_year = $2,
			catalog_number = $3,
			condition = $4,
			notes = $5,
			cover_image_url = $6,
			storage_location = $7,
			label_id = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND user_id = $10
		RETURNING updated_at
	`
	
	var labelID interface{}
	if record.LabelID != 0 {
		labelID = record.LabelID
	} else {
		labelID = nil
	}
	
	err = tx.QueryRow(
		query,
		record.Title,
		sql.NullInt32{Int32: int32(record.ReleaseYear), Valid: record.ReleaseYear != 0},
		sql.NullString{String: record.CatalogNumber, Valid: record.CatalogNumber != ""},
		sql.NullString{String: record.Condition, Valid: record.Condition != ""},
		sql.NullString{String: record.Notes, Valid: record.Notes != ""},
		sql.NullString{String: record.CoverImageURL, Valid: record.CoverImageURL != ""},
		sql.NullString{String: record.StorageLocation, Valid: record.StorageLocation != ""},
		labelID,
		record.ID,
		record.UserID,
	).Scan(&record.UpdatedAt)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("record not found or does not belong to user")
		}
		return nil, err
	}
	
	// Update artists (delete and re-add)
	_, err = tx.Exec("DELETE FROM artist_record WHERE record_id = $1", record.ID)
	if err != nil {
		return nil, err
	}
	
	if len(record.Artists) > 0 {
		for _, artist := range record.Artists {
			_, err = tx.Exec(
				"INSERT INTO artist_record (artist_id, record_id, role) VALUES ($1, $2, $3)",
				artist.ID,
				record.ID,
				artist.Role,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	
	// Update genres (delete and re-add)
	_, err = tx.Exec("DELETE FROM genre_record WHERE record_id = $1", record.ID)
	if err != nil {
		return nil, err
	}
	
	if len(record.Genres) > 0 {
		for _, genre := range record.Genres {
			_, err = tx.Exec(
				"INSERT INTO genre_record (genre_id, record_id) VALUES ($1, $2)",
				genre.ID,
				record.ID,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	
	// Update tracks (delete and re-add)
	if len(record.Tracks) > 0 {
		_, err = tx.Exec("DELETE FROM tracks WHERE record_id = $1", record.ID)
		if err != nil {
			return nil, err
		}
		
		for i := range record.Tracks {
			var trackID int
			err = tx.QueryRow(
				`INSERT INTO tracks (title, duration, position, record_id) 
				 VALUES ($1, $2, $3, $4) RETURNING id`,
				record.Tracks[i].Title,
				record.Tracks[i].Duration,
				record.Tracks[i].Position,
				record.ID,
			).Scan(&trackID)
			
			if err != nil {
				return nil, err
			}
			record.Tracks[i].ID = trackID
		}
	}
	
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	
	return record, nil
}

// Delete removes a record from the database
func (s *RecordService) Delete(id int) error {
	_, err := s.DB.Exec("DELETE FROM records WHERE id = $1", id)
	return err
}

// Search finds records based on various criteria
func (s *RecordService) Search(userID int, query, artist, genre, label, location string) ([]Record, error) {
	// Base query
	sqlQuery := `
		SELECT DISTINCT r.id, r.title, r.release_year, r.catalog_number, r.condition, r.notes, 
		       r.cover_image_url, r.storage_location, r.user_id, r.label_id, 
		       r.created_at, r.updated_at
		FROM records r
		LEFT JOIN artist_record ar ON r.id = ar.record_id
		LEFT JOIN artists a ON ar.artist_id = a.id
		LEFT JOIN genre_record gr ON r.id = gr.record_id
		LEFT JOIN genres g ON gr.genre_id = g.id
		LEFT JOIN labels l ON r.label_id = l.id
		WHERE r.user_id = $1
	`
	
	// Add search conditions
	var conditions []string
	var params []interface{}
	params = append(params, userID)
	paramCount := 2
	
	if query != "" {
		condition := fmt.Sprintf("(r.title ILIKE $%d OR r.notes ILIKE $%d OR r.catalog_number ILIKE $%d)", 
			paramCount, paramCount, paramCount)
		conditions = append(conditions, condition)
		params = append(params, "%"+query+"%")
		paramCount++
	}
	
	if artist != "" {
		condition := fmt.Sprintf("a.name ILIKE $%d", paramCount)
		conditions = append(conditions, condition)
		params = append(params, "%"+artist+"%")
		paramCount++
	}
	
	if genre != "" {
		condition := fmt.Sprintf("g.name ILIKE $%d", paramCount)
		conditions = append(conditions, condition)
		params = append(params, "%"+genre+"%")
		paramCount++
	}
	
	if label != "" {
		condition := fmt.Sprintf("l.name ILIKE $%d", paramCount)
		conditions = append(conditions, condition)
		params = append(params, "%"+label+"%")
		paramCount++
	}
	
	if location != "" {
		condition := fmt.Sprintf("r.storage_location ILIKE $%d", paramCount)
		conditions = append(conditions, condition)
		params = append(params, "%"+location+"%")
		paramCount++
	}
	
	// Add conditions to query
	if len(conditions) > 0 {
		sqlQuery += " AND " + strings.Join(conditions, " AND ")
	}
	
	sqlQuery += " ORDER BY r.title ASC"
	
	// Execute query
	rows, err := s.DB.Query(sqlQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []Record
	
	for rows.Next() {
		var record Record
		var releaseYear, labelID sql.NullInt32
		var catalogNumber, condition, notes, coverImageURL, storageLocation sql.NullString
		
		err := rows.Scan(
			&record.ID,
			&record.Title,
			&releaseYear,
			&catalogNumber,
			&condition,
			&notes,
			&coverImageURL,
			&storageLocation,
			&record.UserID,
			&labelID,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		// Handle nullable fields
		if releaseYear.Valid {
			record.ReleaseYear = int(releaseYear.Int32)
		}
		if catalogNumber.Valid {
			record.CatalogNumber = catalogNumber.String
		}
		if condition.Valid {
			record.Condition = condition.String
		}
		if notes.Valid {
			record.Notes = notes.String
		}
		if coverImageURL.Valid {
			record.CoverImageURL = coverImageURL.String
		}
		if storageLocation.Valid {
			record.StorageLocation = storageLocation.String
		}
		if labelID.Valid {
			record.LabelID = int(labelID.Int32)
		}
		
		records = append(records, record)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// Load related entities for each record
	for i := range records {
		if err := s.loadArtists(&records[i]); err != nil {
			return nil, err
		}
		if err := s.loadGenres(&records[i]); err != nil {
			return nil, err
		}
		
		if records[i].LabelID != 0 {
			if err := s.loadLabel(&records[i]); err != nil {
				return nil, err
			}
		}
	}
	
	return records, nil
}