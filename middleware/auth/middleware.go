// middleware/auth/middleware.go
package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
)

// Custom claims structure
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// Middleware authenticates requests using JWT
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a logger with request info
		log := zerolog.New(os.Stdout).With().
			Timestamp().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_ip", r.RemoteAddr).
			Logger()

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn().Msg("Missing Authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check the format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn().Msg("Invalid Authorization header format")
			http.Error(w, "Authorization header must be in format: Bearer {token}", http.StatusUnauthorized)
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		log.Debug().Msg("JWT token extracted from header")

		// Parse and validate the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Make sure the signing method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Error().Str("signing_method", token.Method.Alg()).Msg("Unexpected signing method")
				return nil, errors.New("unexpected signing method")
			}

			// Return the secret key
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Handle validation errors
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				log.Warn().Err(err).Msg("Invalid token signature")
				http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				return
			}
			log.Warn().Err(err).Msg("Invalid token")
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Verify token is valid
		if !token.Valid {
			log.Warn().Msg("Token validation failed")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check token expiration
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			log.Warn().
				Time("expires_at", claims.ExpiresAt.Time).
				Msg("Token expired")
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		// Token is valid, add user ID to context
		log.Info().
			Int("user_id", claims.UserID).
			Msg("JWT authentication successful")
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)

		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int) (string, error) {
	// Get secret key from environment
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT_SECRET not set in environment")
	}

	// Get expiration duration from environment or use default
	expirationStr := os.Getenv("JWT_EXPIRATION")
	if expirationStr == "" {
		expirationStr = "24h" // Default to 24 hours
	}

	// Parse duration
	expiration, err := time.ParseDuration(expirationStr)
	if err != nil {
		return "", errors.New("invalid JWT_EXPIRATION format")
	}

	// Create the claims
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString([]byte(secretKey))
}
