package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	auth "github.com/wycliff-ochieng/sports-common-package/middleware"
)

/*
	type Claims struct {
		ID     int       `json:"id"`
		UserID uuid.UUID `json:"userid"`
		Roles  []string  `json:"roles"`
		Email  string    `json:"email"`
		jwt.RegisteredClaims
	}
*/
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

var JWTSecret = []byte("mydogsnameisrufus")
var RefreshSecret = []byte("myotherdogsnameistommy")

func GenerateToken(id int, userID uuid.UUID, roles []string, email string, jwtSecret string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := &auth.Claims{
		ID:     id,
		UserID: userID.String(),
		Roles:  roles,
		Email:  email,
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

func ValidateToken(tokenString string, jwtSecret string) (*auth.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token : %v", err)
	}

	if claims, ok := token.Claims.(*auth.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func GenerateTokenPair(userID int, userUUID uuid.UUID, roles []string, email string, jwtSecret string, refreshSecret string, jwtExpiry time.Duration, refreshExpiry time.Duration) (*TokenPair, error) {
	accessToken, err := GenerateToken(userID, userUUID, roles, email, jwtSecret, jwtExpiry)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateToken(userID, userUUID, roles, email, refreshSecret, refreshExpiry)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func ExtractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
