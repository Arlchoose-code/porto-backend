package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/educations — ambil semua education (publik)
func FindEducations(c *gin.Context) {

	// Inisialisasi slice untuk menampung data education
	var educations []models.Education

	// Ambil data education diurutkan dari yang terbaru
	database.DB.Order("start_year desc").Find(&educations)

	// Kirimkan response sukses dengan data education
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Educations",
		Data:    educations,
	})
}

// GET /api/educations/:id — ambil detail education (publik)
func FindEducationById(c *gin.Context) {

	// Ambil ID education dari parameter URL
	id := c.Param("id")

	// Inisialisasi education
	var education models.Education

	// Cari education berdasarkan ID
	if err := database.DB.First(&education, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Education not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data education
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Education Found",
		Data:    education,
	})
}

// POST /api/educations — buat education baru (auth)
func CreateEducation(c *gin.Context) {

	// Struct education request
	var req structs.EducationCreateRequest

	// Bind form data ke struct EducationCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload logo jika ada
	logoUrl := ""
	if _, err := c.FormFile("logo"); err == nil {
		path, err := helpers.UploadFile(c, "logo", "educations")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload logo",
				Errors:  map[string]string{"logo": err.Error()},
			})
			return
		}
		logoUrl = helpers.GetFileUrl(path)
	}

	// Inisialisasi education baru
	education := models.Education{
		School:      req.School,
		Degree:      req.Degree,
		Field:       req.Field,
		StartYear:   req.StartYear,
		EndYear:     req.EndYear,
		Description: req.Description,
		LogoUrl:     logoUrl,
	}

	// Simpan education ke database
	if err := database.DB.Create(&education).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create education",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	go helpers.RevalidateFrontend("education", "")

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Education created successfully",
		Data:    education,
	})
}

// PUT /api/educations/:id — update education (auth)
func UpdateEducation(c *gin.Context) {

	// Ambil ID education dari parameter URL
	id := c.Param("id")

	// Inisialisasi education
	var education models.Education

	// Cari education berdasarkan ID
	if err := database.DB.First(&education, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Education not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Struct education request
	var req structs.EducationUpdateRequest

	// Bind form data ke struct EducationUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload logo baru jika ada, hapus yang lama
	if _, err := c.FormFile("logo"); err == nil {
		helpers.DeleteFile(education.LogoUrl)
		path, err := helpers.UploadFile(c, "logo", "educations")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload logo",
				Errors:  map[string]string{"logo": err.Error()},
			})
			return
		}
		education.LogoUrl = helpers.GetFileUrl(path)
	}

	// Update data education
	education.School = req.School
	education.Degree = req.Degree
	education.Field = req.Field
	education.StartYear = req.StartYear
	education.EndYear = req.EndYear
	education.Description = req.Description

	// Simpan perubahan ke database
	if err := database.DB.Save(&education).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update education",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	go helpers.RevalidateFrontend("education", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Education updated successfully",
		Data:    education,
	})
}

// DELETE /api/educations/:id — hapus education (auth)
func DeleteEducation(c *gin.Context) {

	// Ambil ID education dari parameter URL
	id := c.Param("id")

	// Inisialisasi education
	var education models.Education

	// Cari education berdasarkan ID
	if err := database.DB.First(&education, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Education not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus logo dari lokal jika ada
	helpers.DeleteFile(education.LogoUrl)

	// Hapus education dari database
	if err := database.DB.Delete(&education).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete education",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	go helpers.RevalidateFrontend("education", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Education deleted successfully",
		Data:    nil,
	})
}
