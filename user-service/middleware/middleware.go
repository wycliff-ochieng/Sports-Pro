package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const UserUUIDKey ContextKey = "userUUID"
const RolesdKey ContextKey = "roles"

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
				ctx = context.WithValue(r.Context(), RolesdKey, claims.Roles)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "could not parse token claims", http.StatusFailedDependency)
			}
		})
	}
}

func GetUserUUIDFromContext(ctx context.Context) (string, error) {
	userUUID, ok := ctx.Value(UserUUIDKey).(string)
	if !ok || userUUID == "" {
		return "", errors.New("UUID not found for this user")
	}
	return userUUID, nil
}
