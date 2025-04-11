package handlers

import (
	"errors"
	"net/http"
)

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(r *http.Request) (int, error) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		return 0, errors.New("unauthorized: user ID not found in context")
	}
	return userID, nil
}
