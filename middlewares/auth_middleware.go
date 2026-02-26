package middlewares

import (
	"arlchoose/backend-api/config"
	"arlchoose/backend-api/helpers"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(config.GetEnv("JWT_SECRET", "secret_key"))

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		tokenString := c.GetHeader("Authorization")

		// Fallback untuk SSE — baca dari query param
		if tokenString == "" {
			queryToken := c.Query("token")
			if queryToken != "" {
				tokenString = "Bearer " + queryToken
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims := &helpers.CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserId)
		c.Set("username", claims.Username)

		c.Next()
	}
}
