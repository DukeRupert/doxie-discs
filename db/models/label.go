package models

import (
	"database/sql"
	"time"
)

type Label struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LabelService struct {
	DB *sql.DB
}

func NewLabelService(db *sql.DB) *LabelService {
	return &LabelService{DB: db}
}

func (s *LabelService) GetByID(id int) (*Label, error) {
	var label Label
	var description sql.NullString
	
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM labels
		WHERE id = $1
	`
	
	err := s.DB.QueryRow(query, id).Scan(
		&label.ID,
		&label.Name,
		&description,
		&label.UserID,
		&label.CreatedAt,
		&label.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if description.Valid {
		label.Description = description.String
	}
	
	return &label, nil
}

func (s *LabelService) ListByUserID(userID int) ([]Label, error) {
	query := `
		SELECT id, name, description, user_id, created_at, updated_at
		FROM labels
		WHERE user_id = $1
		ORDER BY name ASC
	`
	
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var labels []Label
	
	for rows.Next() {
		var label Label
		var description sql.NullString
		
		err := rows.Scan(
			&label.ID,
			&label.Name,
			&description,
			&label.UserID,
			&label.CreatedAt,
			&label.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if description.Valid {
			label.Description = description.String
		}
		
		labels = append(labels, label)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return labels, nil
}

func (s *LabelService) Create(label *Label) (*Label, error) {
	query := `
		INSERT INTO labels (name, description, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	var description sql.NullString
	if label.Description != "" {
		description = sql.NullString{String: label.Description, Valid: true}
	}
	
	err := s.DB.QueryRow(
		query,
		label.Name,
		description,
		label.UserID,
	).Scan(
		&label.ID,
		&label.CreatedAt,
		&label.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return label, nil
}

func (s *LabelService) Update(label *Label) (*Label, error) {
	query := `
		UPDATE labels
		SET name = $1,
			description = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4
		RETURNING updated_at
	`
	
	var description sql.NullString
	if label.Description != "" {
		description = sql.NullString{String: label.Description, Valid: true}
	}
	
	err := s.DB.QueryRow(
		query,
		label.Name,
		description,
		label.ID,
		label.UserID,
	).Scan(&label.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return label, nil
}

func (s *LabelService) Delete(id int, userID int) error {
	_, err := s.DB.Exec("DELETE FROM labels WHERE id = $1 AND user_id = $2", id, userID)
	return err
}