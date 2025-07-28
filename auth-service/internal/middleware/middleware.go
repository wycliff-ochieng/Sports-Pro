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

			//call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

//get userId from request context

//get user email from request context
