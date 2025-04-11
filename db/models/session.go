package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// Session represents a user session in the system
type Session struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Token     string          `json:"token"`
	Data      json.RawMessage `json:"-"` // Data stored as JSON but not exported
	IPAddress string          `json:"ip_address"`
	UserAgent string          `json:"user_agent"`
	ExpiresAt time.Time       `json:"expires_at"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// SessionData represents the data stored in the session
type SessionData struct {
	UserEmail string `json:"user_email"`
	UserName  string `json:"user_name"`
	// Add more fields as needed
}

// SessionService handles session-related operations
type SessionService struct {
	DB *sql.DB
}

// NewSessionService creates a new SessionService
func NewSessionService(db *sql.DB) *SessionService {
	return &SessionService{DB: db}
}

// Create creates a new session for a user
func (s *SessionService) Create(userID int, userData map[string]interface{}, ipAddress, userAgent string, duration time.Duration) (*Session, error) {
	// Generate a random token
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return nil, err
	}
	sessionToken := base64.URLEncoding.EncodeToString(token)

	// Set expiration time
	expiresAt := time.Now().Add(duration)

	// Marshal session data
	dataJSON, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}

	// Create session in database
	var sessionID int
	query := `
		INSERT INTO sessions (user_id, token, data, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err = s.DB.QueryRow(
		query,
		userID,
		sessionToken,
		dataJSON,
		ipAddress,
		userAgent,
		expiresAt,
	).Scan(&sessionID, &expiresAt, &expiresAt) // Reusing expiresAt for created_at and updated_at
	if err != nil {
		return nil, err
	}

	// Return the created session
	return &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     sessionToken,
		Data:      dataJSON,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
		CreatedAt: expiresAt,
		UpdatedAt: expiresAt,
	}, nil
}

// GetByToken retrieves a session by its token
func (s *SessionService) GetByToken(token string) (*Session, error) {
	var session Session
	
	query := `
		SELECT id, user_id, token, data, ip_address, user_agent, expires_at, created_at, updated_at
		FROM sessions
		WHERE token = $1
	`
	
	err := s.DB.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.Data,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete the expired session
		s.Delete(session.ID)
		return nil, errors.New("session expired")
	}
	
	return &session, nil
}

// GetDataFromSession extracts and parses the session data
func (s *SessionService) GetDataFromSession(session *Session) (*SessionData, error) {
	var data SessionData
	err := json.Unmarshal(session.Data, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// Refresh extends a session's expiration time
func (s *SessionService) Refresh(token string, duration time.Duration) error {
	newExpiresAt := time.Now().Add(duration)
	
	query := `
		UPDATE sessions
		SET expires_at = $1, updated_at = NOW()
		WHERE token = $2
	`
	
	_, err := s.DB.Exec(query, newExpiresAt, token)
	return err
}

// Delete removes a session
func (s *SessionService) Delete(sessionID int) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.DB.Exec(query, sessionID)
	return err
}

// DeleteByToken removes a session by token
func (s *SessionService) DeleteByToken(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := s.DB.Exec(query, token)
	return err
}

// DeleteByUserID removes all sessions for a user
func (s *SessionService) DeleteByUserID(userID int) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := s.DB.Exec(query, userID)
	return err
}

// CleanExpiredSessions removes all expired sessions
func (s *SessionService) CleanExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := s.DB.Exec(query)
	return err
}