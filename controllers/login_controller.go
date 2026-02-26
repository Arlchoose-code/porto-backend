package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {

	var req = structs.UserLoginRequest{}
	var user = models.User{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Success: false,
			Message: "User Not Found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Success: false,
			Message: "Invalid Password",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Generate access token (60 menit) dan refresh token (7 hari)
	token := helpers.GenerateToken(user.Id, user.Username)
	refreshToken := helpers.GenerateRefreshToken(user.Id, user.Username)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Login Success",
		Data: map[string]any{
			"id":            user.Id,
			"name":          user.Name,
			"username":      user.Username,
			"email":         user.Email,
			"created_at":    user.CreatedAt.String(),
			"updated_at":    user.UpdatedAt.String(),
			"token":         token,
			"refresh_token": refreshToken,
		},
	})
}

// POST /api/refresh — generate access token baru dari refresh token
func RefreshToken(c *gin.Context) {

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Validasi refresh token
	claims, err := helpers.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Success: false,
			Message: "Invalid or expired refresh token",
			Errors:  map[string]string{"refresh_token": "invalid or expired"},
		})
		return
	}

	// Generate access token baru
	newToken := helpers.GenerateToken(claims.UserId, claims.Username)
	newRefreshToken := helpers.GenerateRefreshToken(claims.UserId, claims.Username)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Token refreshed",
		Data: map[string]any{
			"token":         newToken,
			"refresh_token": newRefreshToken,
		},
	})
}
