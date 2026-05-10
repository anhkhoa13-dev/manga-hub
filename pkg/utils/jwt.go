package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("default_fallback_secret") 
	}
	return []byte(secret)
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Tạo token JWT
func GenerateToken(userID, username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTSecret())
}