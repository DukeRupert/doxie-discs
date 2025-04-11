// api/handlers/users.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"strings"
	
	"github.com/go-chi/chi/v5/middleware"
	"github.com/dukerupert/doxie-discs/db/models"
	"golang.org/x/crypto/bcrypt"
	"github.com/rs/zerolog/log"

)

type UserHandler struct {
	DB         *sql.DB
	UserService *models.UserService
	SessionService *models.SessionService
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type UpdateProfileRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// NewUserHandler creates a new UserHandler with the given DB connection
func NewUserHandler(db *sql.DB, sessionService *models.SessionService) *UserHandler {
	return &UserHandler{
		DB:         db,
		UserService: models.NewUserService(db),
		SessionService: sessionService,
	}
}

// Register creates a new user account and starts a session
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "Email, password, and name are required", http.StatusBadRequest)
		return
	}

	// Check if email already exists
	exists, err := h.UserService.EmailExists(req.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
	}

	createdUser, err := h.UserService.Create(user)
	if err != nil {
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get IP address and user agent for the session
	ipAddress := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ipAddress = forwardedFor
	}
	userAgent := r.Header.Get("User-Agent")
	
	// Create session data
	sessionData := map[string]interface{}{
		"user_email": createdUser.Email,
		"user_name":  createdUser.Name,
	}
	
	// Create a new session
	session, err := h.SessionService.Create(
		createdUser.ID,
		sessionData,
		ipAddress,
		userAgent,
		24*time.Hour, // Session duration
	)
	
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set to true in production with HTTPS
		MaxAge:   int(24 * time.Hour.Seconds()),
		SameSite: http.SameSiteStrictMode,
	})

	// Return user info (without token in response body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"id":    createdUser.ID,
			"email": createdUser.Email,
			"name":  createdUser.Name,
		},
	})
}

// Login authenticates a user and creates a session
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Create a logger with request context
    logger := log.With().
        Str("request_id", middleware.GetReqID(r.Context())).
        Str("handler", "UserHandler.Login").
        Logger()

    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Warn().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
	// Validate request
	if req.Email == "" || req.Password == "" {
		logger.Warn().
			Str("email", req.Email).
			Bool("password_empty", req.Password == "").
			Msg("Missing required fields")
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	logger.Debug().Str("email", req.Email).Msg("Attempting login")

	// Get user by email
	user, err := h.UserService.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn().
				Str("email", req.Email).
				Msg("User not found")
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		logger.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Database error when retrieving user")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.Warn().
			Str("email", req.Email).
			Int("user_id", user.ID).
			Msg("Invalid password")
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

    // Get IP address and user agent
    ipAddress := r.RemoteAddr
    if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
        ipAddress = forwardedFor
    }
    userAgent := r.Header.Get("User-Agent")
    
    // Create session data
    sessionData := map[string]interface{}{
        "user_email": user.Email,
        "user_name": user.Name,
    }
    
    // Create session using the service
    session, err := h.SessionService.Create(
        user.ID,
        sessionData,
        ipAddress,
        userAgent,
        24*time.Hour, // Session duration
    )
    
    if err != nil {
        logger.Error().
            Err(err).
            Int("user_id", user.ID).
            Msg("Failed to create session")
        http.Error(w, "Error creating session", http.StatusInternalServerError)
        return
    }

    // Set the session token as a cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    session.Token,
        Path:     "/",
        HttpOnly: true,
        Secure:   r.TLS != nil, // Set to true if using HTTPS
        MaxAge:   int(24 * time.Hour.Seconds()),
        SameSite: http.SameSiteStrictMode,
    })

    logger.Info().
        Int("user_id", user.ID).
        Str("email", user.Email).
        Msg("User authenticated successfully")

    // Check if this is an API request or a form submission
    if strings.Contains(r.Header.Get("Accept"), "application/json") {
        // For API clients, return JSON response with user info
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": true,
            "user": map[string]interface{}{
                "id": user.ID,
                "email": user.Email,
                "name": user.Name,
            },
        })
    } else {
        // For browser clients, redirect to dashboard
        http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
    }
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve userID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return user profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile updates the current user's profile information
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(int)

	// Decode request
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Email == "" || req.Name == "" {
		http.Error(w, "Email and name are required", http.StatusBadRequest)
		return
	}

	// Get current user
	currentUser, err := h.UserService.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if email is being changed and if it already exists
	if req.Email != currentUser.Email {
		exists, err := h.UserService.EmailExists(req.Email)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "Email already in use", http.StatusConflict)
			return
		}
	}

	// Update user
	user := &models.User{
		ID:           userID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: currentUser.PasswordHash,
	}

	updatedUser, err := h.UserService.Update(user)
	if err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	// Return updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

// UpdatePassword changes the current user's password
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(int)

	// Decode request
	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Current password and new password are required", http.StatusBadRequest)
		return
	}

	// Get current user
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// Update password
	err = h.UserService.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully",
	})
}