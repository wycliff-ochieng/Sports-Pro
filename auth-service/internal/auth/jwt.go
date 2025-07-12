package handlers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID    int
	Email string
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

var JWTSecret = []byte("mydogsnameisrufus")
var RefreshSecret = []byte("myotherdogsnameistommy")

func GenerateToken(id int, email string, jwtSecret string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := &Claims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println(token)
	return token.SignedString([]byte(jwtSecret))
}

//validate token

func ValidateToken(tokenString string, jwtSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token : %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func GenerateTokenPair(userID int, email string, jwtSecret string, refreshSecret string, jwtExpiry time.Duration, refreshExpiry time.Duration) (*TokenPair, error) {
	accessToken, err := GenerateToken(userID, email, jwtSecret, jwtExpiry)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateToken(userID, email, refreshSecret, refreshExpiry)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
