// db/models/user.go
package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Don't expose in JSON
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserService provides methods for user data operations
type UserService struct {
	DB *sql.DB
}

// NewUserService creates a new UserService with the given DB connection
func NewUserService(db *sql.DB) *UserService {
	return &UserService{DB: db}
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(id int) (*User, error) {
	user := &User{}
	err := s.DB.QueryRow(
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetByEmail retrieves a user by their email address
func (s *UserService) GetByEmail(email string) (*User, error) {
	user := &User{}
	err := s.DB.QueryRow(
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// Create adds a new user to the database
func (s *UserService) Create(user *User) (*User, error) {
	err := s.DB.QueryRow(
		`INSERT INTO users (email, password_hash, name) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, created_at, updated_at`,
		user.Email,
		user.PasswordHash,
		user.Name,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// Update modifies an existing user
func (s *UserService) Update(user *User) (*User, error) {
	err := s.DB.QueryRow(
		`UPDATE users 
		 SET email = $1, name = $2, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $3 
		 RETURNING updated_at`,
		user.Email,
		user.Name,
		user.ID,
	).Scan(&user.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// UpdatePassword changes a user's password
func (s *UserService) UpdatePassword(userID int, passwordHash string) error {
	_, err := s.DB.Exec(
		`UPDATE users 
		 SET password_hash = $1, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $2`,
		passwordHash,
		userID,
	)
	
	return err
}

// Delete removes a user from the database
func (s *UserService) Delete(id int) error {
	_, err := s.DB.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// EmailExists checks if an email is already registered
func (s *UserService) EmailExists(email string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		email,
	).Scan(&exists)
	
	if err != nil {
		return false, err
	}
	
	return exists, nil
}