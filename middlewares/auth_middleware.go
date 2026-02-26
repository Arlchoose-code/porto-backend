package middlewares

import (
	"arlchoose/backend-api/helpers"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

		claims, err := helpers.ValidateToken(tokenString)
		if err != nil {
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
