package middleware

import (
	"context"
	"net/http"
	handlers "sports/authservice/internal/auth"
	"sports/authservice/internal/config"
)

type ContextKey string

const (
	UserIDKey ContextKey = "userId"
	EmailKey  ContextKey = "email"
	RolesKey  ContextKey = "roles"
)

// Authorization, access control depends on this (granting permissions to coaches ,admins and players), protecting routes and stuff

func UserValidationMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func AuthMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization Header required", http.StatusUnauthorized)
				return
			}

			tokenString := handlers.ExtractBearerToken(authHeader)
			if tokenString == "" {
				http.Error(w, "Invalid Token Format ", http.StatusUnauthorized)
				return
			}

			claims, err := handlers.ValidateToken(tokenString, *&cfg.JWTSecret)
			if err != nil {
				http.Error(w, "Invalid or Expired Token", http.StatusUnauthorized)
				return
			}

			//add information to request context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.ID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)

			//call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// get userId from request context
func GetUserID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(UserIDKey).(string)
	return id, ok
}
func GetUserEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(EmailKey).(string)
	return email, ok
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
