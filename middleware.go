package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	models "github.com/wibecoderr/ToDo/model"
	database "github.com/wibecoderr/ToDo/database/dbhelper"
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
		sessionToken, err := extractSessionToken(r)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := database.GetUserIDBySession(sessionToken)
		if err != nil {
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}
		user := &models.UserCxt{
			UserId:    userID,
			SessionId: sessionToken,
		}

		ctx := context.WithValue(r.Context(), usercontextKey, user)
		//ctx = context.WithValue(ctx, SessionTokenKey, sessionToken)

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
