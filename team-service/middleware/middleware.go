package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextKey string

const UserUUIDKey ContextKey = "userUUID"
const RolesKey ContextKey = "roles"
const UserIDKey ContextKey = "userID"

type Claims struct {
	UserUUID string   `json:"sub"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func UserMiddlware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "empty authorization header", http.StatusExpectationFailed)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid token format", http.StatusBadRequest)
			}

			//validate token
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("signing method not okay")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || token.Valid {
				http.Error(w, "wrong/expired token", http.StatusUnauthorized)
				return
			}
			//Extract claims (the populate context)
			if claims, ok := token.Claims.(*Claims); ok {
				ctx := context.WithValue(r.Context(), UserUUIDKey, claims.UserUUID)
				ctx = context.WithValue(r.Context(), RolesKey, claims.Roles)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "could not parse token claims", http.StatusFailedDependency)
			}
		})
	}
}

func GetUserUUIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userUUID, ok := ctx.Value(UserUUIDKey).(uuid.UUID)
	if !ok || userUUID == uuid.Nil {
		return uuid.Nil, errors.New("UUID not found for this user")
	}
	return userUUID, nil
}

func GetUserRoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value(RolesKey).(string)
	if !ok || role == "" {
		return "", errors.New("error: No roles for this user")
	}
	return role, nil
}

func GetUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok || userID == 0 {
		return 0, errors.New("UserId not foumd for this user")
	}
	return userID, nil
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the user's roles from the context set by the AuthMiddleware.
			userRoles, ok := r.Context().Value(RolesKey).([]interface{})
			if !ok {
				// This should not happen if AuthMiddleware is used correctly
				http.Error(w, "Could not retrieve user roles from context", http.StatusInternalServerError)
				return
			}

			// Create a set for efficient lookup
			allowedRolesSet := make(map[string]struct{})
			for _, role := range allowedRoles {
				allowedRolesSet[role] = struct{}{}
			}

			// Check if the user has at least one of the allowed roles
			for _, role := range userRoles {
				roleStr, ok := role.(string)
				if !ok {
					continue
				} // Skip if role is not a string

				if _, found := allowedRolesSet[roleStr]; found {
					// User has a required role, proceed to the next handler
					next.ServeHTTP(w, r)
					return
				}
			}
			// If we get here, the user has none of the required roles
			http.Error(w, "Forbidden: You don't have the required permissions.", http.StatusForbidden)
		})
	}
}
