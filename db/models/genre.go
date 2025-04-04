package models

import (
	"database/sql"
	"time"
)

type Genre struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GenreService struct {
	DB *sql.DB
}

func NewGenreService(db *sql.DB) *GenreService {
	return &GenreService{DB: db}
}

func (s *GenreService) GetByID(id int) (*Genre, error) {
	var genre Genre
	var description sql.NullString
	
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM genres
		WHERE id = $1
	`
	
	err := s.DB.QueryRow(query, id).Scan(
		&genre.ID,
		&genre.Name,
		&description,
		&genre.UserID,
		&genre.CreatedAt,
		&genre.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if description.Valid {
		genre.Description = description.String
	}
	
	return &genre, nil
}

func (s *GenreService) ListByUserID(userID int) ([]Genre, error) {
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM genres
		WHERE user_id = $1
		ORDER BY name ASC
	`
	
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var genres []Genre
	
	for rows.Next() {
		var genre Genre
		var description sql.NullString
		
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&description,
			&genre.UserID,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if description.Valid {
			genre.Description = description.String
		}
		
		genres = append(genres, genre)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return genres, nil
}

func (s *GenreService) Create(genre *Genre) (*Genre, error) {
	query := `
		INSERT INTO genres (name, description, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	var description sql.NullString
	if genre.Description != "" {
		description = sql.NullString{String: genre.Description, Valid: true}
	}
	
	err := s.DB.QueryRow(
		query,
		genre.Name,
		description,
		genre.UserID,
	).Scan(
		&genre.ID,
		&genre.CreatedAt,
		&genre.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return genre, nil
}

func (s *GenreService) Update(genre *Genre) (*Genre, error) {
	query := `
		UPDATE genres
		SET name = $1,
			description = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4
		RETURNING updated_at
	`
	
	var description sql.NullString
	if genre.Description != "" {
		description = sql.NullString{String: genre.Description, Valid: true}
	}
	
	err := s.DB.QueryRow(
		query,
		genre.Name,
		description,
		genre.ID,
		genre.UserID,
	).Scan(&genre.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return genre, nil
}

func (s *GenreService) Delete(id int, userID int) error {
	_, err := s.DB.Exec("DELETE FROM genres WHERE id = $1 AND user_id = $2", id, userID)
	return err
}