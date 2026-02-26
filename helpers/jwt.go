package helpers

import (
	"arlchoose/backend-api/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(config.GetEnv("JWT_SECRET", "secret_key"))

// CustomClaims tambah userId ke JWT
type CustomClaims struct {
	UserId   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken generate access token (60 menit)
func GenerateToken(userId uint, username string) string {
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &CustomClaims{
		UserId:   userId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
	return token
}

// GenerateRefreshToken generate refresh token (7 hari)
func GenerateRefreshToken(userId uint, username string) string {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &CustomClaims{
		UserId:   userId,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
	return token
}

// ValidateToken validasi token dan return claims
func ValidateToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
