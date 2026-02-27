package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/profile — ambil profile (publik)
func GetProfile(c *gin.Context) {

	var profile models.Profile

	// Ambil data pertama — profile hanya ada 1
	if err := database.DB.First(&profile).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Profile not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Profile Found",
		Data:    profile,
	})
}

// PUT /api/profile — update atau buat profile (auth)
func UpsertProfile(c *gin.Context) {

	var req structs.ProfileRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	var profile models.Profile
	database.DB.First(&profile)

	// Upload avatar jika ada
	if _, err := c.FormFile("avatar"); err == nil {
		if profile.Avatar != "" {
			helpers.DeleteFile(profile.Avatar)
		}
		path, err := helpers.UploadFile(c, "avatar", "profile")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload avatar",
				Errors:  map[string]string{"avatar": err.Error()},
			})
			return
		}
		profile.Avatar = helpers.GetFileUrl(path)
	}

	// Upload resume/CV jika ada
	if _, err := c.FormFile("resume"); err == nil {
		if profile.ResumeUrl != "" {
			helpers.DeleteFile(profile.ResumeUrl)
		}
		path, err := helpers.UploadFile(c, "resume", "resume")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload resume",
				Errors:  map[string]string{"resume": err.Error()},
			})
			return
		}
		profile.ResumeUrl = helpers.GetFileUrl(path)
	}

	profile.Name = req.Name
	profile.Tagline = req.Tagline
	profile.Bio = req.Bio
	profile.Github = req.Github
	profile.Linkedin = req.Linkedin
	profile.Twitter = req.Twitter
	profile.Instagram = req.Instagram
	profile.Email = req.Email
	profile.Phone = req.Phone
	profile.Location = req.Location

	if err := database.DB.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to save profile",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	go helpers.RevalidateFrontend("profile", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Profile saved successfully",
		Data:    profile,
	})
}
