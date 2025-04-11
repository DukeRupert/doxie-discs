// middleware/auth/middleware.go
package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"database/sql"
	"time"

	"github.com/dukerupert/doxie-discs/db/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
)

// Custom claims structure
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// SessionAuthMiddleware authenticates requests using session cookies
func SessionAuthMiddleware(sessionService *models.SessionService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a logger with request info
			log := zerolog.New(os.Stdout).With().
				Timestamp().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_ip", r.RemoteAddr).
				Logger()

			// Get the session cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				if err == http.ErrNoCookie {
					log.Warn().Msg("No session cookie present")
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				log.Error().Err(err).Msg("Error retrieving session cookie")
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}

			// Extract the token
			sessionToken := cookie.Value
			log.Debug().Msg("Session token extracted from cookie")

			// Verify session exists and is valid
			session, err := sessionService.GetByToken(sessionToken)
			if err != nil {
				if err.Error() == "session expired" {
					log.Warn().Msg("Session has expired")
					// Clear the cookie
					http.SetCookie(w, &http.Cookie{
						Name:     "session_token",
						Value:    "",
						Path:     "/",
						MaxAge:   -1,
						HttpOnly: true,
					})
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				if errors.Is(err, sql.ErrNoRows) {
					log.Warn().Msg("Invalid session token")
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				log.Error().Err(err).Msg("Database error while retrieving session")
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}

			// Check if session is expired (additional safety check)
			if time.Now().After(session.ExpiresAt) {
				log.Warn().
					Time("expires_at", session.ExpiresAt).
					Msg("Session expired")
				
				// Clean up expired session
				_ = sessionService.DeleteByToken(sessionToken)
				
				// Clear the cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Optionally refresh the session to extend expiry time
			// (useful for keeping users logged in during active usage)
			if time.Until(session.ExpiresAt) < 12*time.Hour {
				err = sessionService.Refresh(sessionToken, 24*time.Hour)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to refresh session")
					// Continue anyway as this is not critical
				} else {
					// Update the cookie with new expiration
					http.SetCookie(w, &http.Cookie{
						Name:     "session_token",
						Value:    sessionToken,
						Path:     "/",
						HttpOnly: true,
						Secure:   r.TLS != nil,
						MaxAge:   int(24 * time.Hour.Seconds()),
						SameSite: http.SameSiteStrictMode,
					})
				}
			}

			// Session is valid, add user ID to context
			log.Info().
				Int("user_id", session.UserID).
				Msg("Session authentication successful")
			
			ctx := context.WithValue(r.Context(), "userID", session.UserID)
			// Optionally add the entire session to context if needed
			ctx = context.WithValue(ctx, "session", session)

			// Call the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}