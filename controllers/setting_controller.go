package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/settings — ambil semua setting (publik)
func GetSettings(c *gin.Context) {

	var settings []models.Setting

	database.DB.Find(&settings)

	// Konversi ke map supaya lebih mudah dipakai di frontend
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Settings Found",
		Data:    result,
	})
}

// GET /api/settings/:key — ambil satu setting by key (publik)
func GetSettingByKey(c *gin.Context) {

	key := c.Param("key")
	var setting models.Setting

	if err := database.DB.Where("`key` = ?", key).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Setting not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Setting Found",
		Data:    setting,
	})
}

// PUT /api/settings — upsert banyak setting sekaligus (auth)
func UpsertSettings(c *gin.Context) {

	var req structs.SettingUpsertRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Loop tiap key-value, upsert ke DB
	for key, value := range req.Settings {
		var setting models.Setting

		if err := database.DB.Where("`key` = ?", key).First(&setting).Error; err != nil {
			// Belum ada → buat baru
			setting = models.Setting{Key: key, Value: value}
			database.DB.Create(&setting)
		} else {
			// Sudah ada → update value
			setting.Value = value
			database.DB.Save(&setting)
		}
	}

	// Return semua setting terbaru
	var allSettings []models.Setting
	database.DB.Find(&allSettings)

	result := make(map[string]string)
	for _, s := range allSettings {
		result[s.Key] = s.Value
	}

	go helpers.RevalidateFrontend("settings", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Settings saved successfully",
		Data:    result,
	})
}
