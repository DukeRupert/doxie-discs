package models

import (
	"database/sql"
	"time"
)

type Artist struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Only used in artist_record context
	Role        string    `json:"role,omitempty"`
}

type ArtistService struct {
	DB *sql.DB
}

func NewArtistService(db *sql.DB) *ArtistService {
	return &ArtistService{DB: db}
}

func (s *ArtistService) GetByID(id int) (*Artist, error) {
	var artist Artist
	var description sql.NullString
	
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM artists
		WHERE id = $1
	`
	
	err := s.DB.QueryRow(query, id).Scan(
		&artist.ID,
		&artist.Name,
		&description,
		&artist.UserID,
		&artist.CreatedAt,
		&artist.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if description.Valid {
		artist.Description = description.String
	}
	
	return &artist, nil
}

func (s *ArtistService) ListByUserID(userID int) ([]Artist, error) {
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM artists
		WHERE user_id = $1
		ORDER BY name ASC
	`
	
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var artists []Artist
	
	for rows.Next() {
		var artist Artist
		var description sql.NullString
		
		err := rows.Scan(
			&artist.ID,
			&artist.Name,
			&description,
			&artist.UserID,
			&artist.CreatedAt,
			&artist.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if description.Valid {
			artist.Description = description.String
		}
		
		artists = append(artists, artist)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return artists, nil
}

func (s *ArtistService) Create(artist *Artist) (*Artist, error) {
	query := `
		INSERT INTO artists (name, description, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	var description sql.NullString
	if artist.Description != "" {
		description = sql.NullString{String: artist.Description, Valid: true}
	}
	
	err := s.DB.QueryRow(
		query,
		artist.Name,
		description,
		artist.UserID,
	).Scan(
		&artist.ID,
		&artist.CreatedAt,
		&artist.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return artist, nil
}

func (s *ArtistService) Update(artist *Artist) (*Artist, error) {
	query := `
		UPDATE artists
		SET name = $1,
			description = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4
		RETURNING updated_at
	`
	
	var description sql.NullString
	if artist.Description != "" {
		description = sql.NullString{String: artist.Description, Valid: true}
	}
	
	err := s.DB.QueryRow(
		query,
		artist.Name,
		description,
		artist.ID,
		artist.UserID,
	).Scan(&artist.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return artist, nil
}

func (s *ArtistService) Delete(id int, userID int) error {
	_, err := s.DB.Exec("DELETE FROM artists WHERE id = $1 AND user_id = $2", id, userID)
	return err
}

func (s *ArtistService) Search(userID int, query string) ([]Artist, error) {
	sqlQuery := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM artists
		WHERE user_id = $1 AND name ILIKE $2
		ORDER BY name ASC
	`
	
	rows, err := s.DB.Query(sqlQuery, userID, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var artists []Artist
	
	for rows.Next() {
		var artist Artist
		var description sql.NullString
		
		err := rows.Scan(
			&artist.ID,
			&artist.Name,
			&description,
			&artist.UserID,
			&artist.CreatedAt,
			&artist.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if description.Valid {
			artist.Description = description.String
		}
		
		artists = append(artists, artist)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return artists, nil
}