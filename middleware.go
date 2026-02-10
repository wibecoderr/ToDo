package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	database "github.com/wibecoderr/ToDo/database/dbhelper"
	models "github.com/wibecoderr/ToDo/model"
)

type contextKey struct {
}

var usercontextKey = contextKey{}

//const (
//	UserIDKey       contextKey = "user_id"
//	SessionTokenKey contextKey = "session_token"
//)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondError(w, http.StatusUnauthorized, nil, "missing authorization header")
			return
		}

		// Verify Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			RespondError(w, http.StatusUnauthorized, nil, "invalid authorization format, expected 'Bearer <token>'")
			return
		}

		// Extract token
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			RespondError(w, http.StatusUnauthorized, nil, "empty token")
			return
		}

		// Verify JWT and extract claims
		userID, sessionID, err := VerifyJWT(tokenStr)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, err, "invalid or expired token")
			return
		}

		// Verify session exists in database and is not expired
		dbUserID, err := database.GetUserIDBySession(sessionID)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, err, "session not found or expired")
			return
		}

		// Verify userID from JWT matches userID from database
		if dbUserID != userID {
			RespondError(w, http.StatusUnauthorized, nil, "token user mismatch")
			return
		}

		// Set user context
		user := &models.UserCxt{
			UserId:    userID,
			SessionId: sessionID,
		}

		ctx := context.WithValue(r.Context(), usercontextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractSessionToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization format, expected 'Bearer <token>'")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

//	func GetUserID(r *http.Request) (string, error) {
//		userID, ok := r.Context().Value(UserIDKey).(string)
//		if !ok || userID == "" {
//			return "", errors.New("user ID not found in context")
//		}
//		return userID, nil
//	}
//
//	func GetSessionToken(r *http.Request) (string, error) {
//		token, ok := r.Context().Value(SessionTokenKey).(string)
//		if !ok || token == "" {
//			return "", errors.New("session token not found in context")
//		}
//		return token, nil
//	}
func UserContext(r *http.Request) *models.UserCxt {
	user, _ := r.Context().Value(usercontextKey).(*models.UserCxt)
	return user
}
