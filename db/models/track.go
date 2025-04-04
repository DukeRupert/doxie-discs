package models

import (
	"database/sql"
	"time"
)

type Track struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Duration  string    `json:"duration,omitempty"` // Stored as INTERVAL in DB, represented as string
	Position  string    `json:"position,omitempty"` // e.g., "A1", "B3"
	RecordID  int       `json:"record_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TrackService struct {
	DB *sql.DB
}

func NewTrackService(db *sql.DB) *TrackService {
	return &TrackService{DB: db}
}

func (s *TrackService) GetByID(id int) (*Track, error) {
	var track Track
	var duration, position sql.NullString
	
	query := `
		SELECT id, title, duration, position, record_id, created_at, updated_at
		FROM tracks
		WHERE id = $1
	`
	
	err := s.DB.QueryRow(query, id).Scan(
		&track.ID,
		&track.Title,
		&duration,
		&position,
		&track.RecordID,
		&track.CreatedAt,
		&track.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if duration.Valid {
		track.Duration = duration.String
	}
	
	if position.Valid {
		track.Position = position.String
	}
	
	return &track, nil
}

func (s *TrackService) ListByRecordID(recordID int) ([]Track, error) {
	query := `
		SELECT id, title, duration, position, record_id, created_at, updated_at
		FROM tracks
		WHERE record_id = $1
		ORDER BY position ASC
	`
	
	rows, err := s.DB.Query(query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tracks []Track
	
	for rows.Next() {
		var track Track
		var duration, position sql.NullString
		
		err := rows.Scan(
			&track.ID,
			&track.Title,
			&duration,
			&position,
			&track.RecordID,
			&track.CreatedAt,
			&track.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if duration.Valid {
			track.Duration = duration.String
		}
		
		if position.Valid {
			track.Position = position.String
		}
		
		tracks = append(tracks, track)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return tracks, nil
}