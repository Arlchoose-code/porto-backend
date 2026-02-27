package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/skills — ambil semua skill (publik)
func FindSkills(c *gin.Context) {

	// Inisialisasi slice untuk menampung data skill
	var skills []models.Skill

	// Bisa filter by category via query param, contoh: /api/skills?category=framework
	category := c.Query("category")
	if category != "" {
		database.DB.Where("category = ?", category).Order("category, `order` asc").Find(&skills)
	} else {
		database.DB.Order("category, `order` asc").Find(&skills)
	}

	// Kirimkan response sukses dengan data skill
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Skills",
		Data:    skills,
	})
}

// GET /api/skills/:id — ambil detail skill (publik)
func FindSkillById(c *gin.Context) {

	// Ambil ID skill dari parameter URL
	id := c.Param("id")

	// Inisialisasi skill
	var skill models.Skill

	// Cari skill berdasarkan ID
	if err := database.DB.First(&skill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Skill not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data skill
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Skill Found",
		Data:    skill,
	})
}

// POST /api/skills — buat skill baru (auth)
func CreateSkill(c *gin.Context) {

	// Struct skill request
	var req structs.SkillCreateRequest

	// Bind form data ke struct SkillCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload icon jika ada
	iconUrl := ""
	if _, err := c.FormFile("icon"); err == nil {
		path, err := helpers.UploadFile(c, "icon", "skills")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload icon",
				Errors:  map[string]string{"icon": err.Error()},
			})
			return
		}
		iconUrl = helpers.GetFileUrl(path)
	}

	// Inisialisasi skill baru
	skill := models.Skill{
		Category: req.Category,
		Name:     req.Name,
		Level:    req.Level,
		IconUrl:  iconUrl,
		Order:    req.Order,
	}

	// Simpan skill ke database
	if err := database.DB.Create(&skill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create skill",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	go helpers.RevalidateFrontend("skill", "")

	// Kirimkan response sukses
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Skill created successfully",
		Data:    skill,
	})
}

// PUT /api/skills/:id — update skill (auth)
func UpdateSkill(c *gin.Context) {

	// Ambil ID skill dari parameter URL
	id := c.Param("id")

	// Inisialisasi skill
	var skill models.Skill

	// Cari skill berdasarkan ID
	if err := database.DB.First(&skill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Skill not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Struct skill request
	var req structs.SkillUpdateRequest

	// Bind form data ke struct SkillUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload icon baru jika ada, hapus yang lama
	if _, err := c.FormFile("icon"); err == nil {
		helpers.DeleteFile(skill.IconUrl) // hapus file lama
		path, err := helpers.UploadFile(c, "icon", "skills")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload icon",
				Errors:  map[string]string{"icon": err.Error()},
			})
			return
		}
		skill.IconUrl = helpers.GetFileUrl(path)
	}

	// Update data skill
	skill.Category = req.Category
	skill.Name = req.Name
	skill.Level = req.Level
	skill.Order = req.Order

	// Simpan perubahan ke database
	if err := database.DB.Save(&skill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update skill",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	go helpers.RevalidateFrontend("skill", "")

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Skill updated successfully",
		Data:    skill,
	})
}

// DELETE /api/skills/:id — hapus skill (auth)
func DeleteSkill(c *gin.Context) {

	// Ambil ID skill dari parameter URL
	id := c.Param("id")

	// Inisialisasi skill
	var skill models.Skill

	// Cari skill berdasarkan ID
	if err := database.DB.First(&skill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Skill not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus icon dari lokal jika ada
	helpers.DeleteFile(skill.IconUrl)

	// Hapus skill dari database
	if err := database.DB.Delete(&skill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete skill",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	go helpers.RevalidateFrontend("skill", "")

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Skill deleted successfully",
		Data:    nil,
	})
}
