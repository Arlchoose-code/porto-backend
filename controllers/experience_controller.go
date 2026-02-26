package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /api/experiences — ambil semua experience (publik)
func FindExperiences(c *gin.Context) {

	// Inisialisasi slice untuk menampung data experience
	var experiences []models.Experience

	// Ambil data experience beserta relasi images diurutkan dari yang terbaru
	database.DB.Preload("Images").Order("start_date desc").Find(&experiences)

	// Kirimkan response sukses dengan data experience
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Experiences",
		Data:    experiences,
	})
}

// GET /api/experiences/:id — ambil detail experience (publik)
func FindExperienceById(c *gin.Context) {

	// Ambil ID experience dari parameter URL
	id := c.Param("id")

	// Inisialisasi experience
	var experience models.Experience

	// Cari experience berdasarkan ID beserta relasi images
	if err := database.DB.Preload("Images").First(&experience, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Experience not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data experience
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Experience Found",
		Data:    experience,
	})
}

// POST /api/experiences — buat experience baru (auth)
func CreateExperience(c *gin.Context) {

	// Struct experience request
	var req structs.ExperienceCreateRequest

	// Bind form data ke struct ExperienceCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Inisialisasi experience baru
	experience := models.Experience{
		Company:     req.Company,
		Role:        req.Role,
		Location:    req.Location,
		IsCurrent:   req.IsCurrent,
		Description: req.Description,
	}

	// Parse start_date jika ada
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			experience.StartDate = &t
		}
	}

	// Parse end_date jika ada dan tidak is_current
	if req.EndDate != "" && !req.IsCurrent {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			experience.EndDate = &t
		}
	}

	// Simpan experience ke database
	if err := database.DB.Create(&experience).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create experience",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload images jika ada (bisa multiple)
	form, _ := c.MultipartForm()
	if form != nil {
		files := form.File["images"]
		for i, file := range files {
			// Simpan file sementara ke context
			c.Request.MultipartForm.File["images"] = []*multipart.FileHeader{file}
			path, err := helpers.UploadFile(c, "images", "experiences")
			if err != nil {
				continue
			}

			// Simpan image ke database
			image := models.ExperienceImage{
				ExperienceId: experience.Id,
				ImageUrl:     helpers.GetFileUrl(path),
				Order:        i,
			}
			database.DB.Create(&image)
		}
	}

	// Ambil ulang experience beserta relasinya untuk response
	database.DB.Preload("Images").First(&experience, experience.Id)

	// Kirimkan response sukses
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Experience created successfully",
		Data:    experience,
	})
}

// PUT /api/experiences/:id — update experience (auth)
func UpdateExperience(c *gin.Context) {

	// Ambil ID experience dari parameter URL
	id := c.Param("id")

	// Inisialisasi experience
	var experience models.Experience

	// Cari experience berdasarkan ID
	if err := database.DB.First(&experience, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Experience not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Struct experience request
	var req structs.ExperienceUpdateRequest

	// Bind form data ke struct ExperienceUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Update data experience
	experience.Company = req.Company
	experience.Role = req.Role
	experience.Location = req.Location
	experience.IsCurrent = req.IsCurrent
	experience.Description = req.Description

	// Reset date dulu
	experience.StartDate = nil
	experience.EndDate = nil

	// Parse start_date jika ada
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			experience.StartDate = &t
		}
	}

	// Parse end_date jika ada dan tidak is_current
	if req.EndDate != "" && !req.IsCurrent {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			experience.EndDate = &t
		}
	}

	// Simpan perubahan ke database
	if err := database.DB.Save(&experience).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update experience",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Ambil ulang experience beserta relasinya untuk response
	database.DB.Preload("Images").First(&experience, experience.Id)

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Experience updated successfully",
		Data:    experience,
	})
}

// DELETE /api/experiences/:id — hapus experience (auth)
func DeleteExperience(c *gin.Context) {

	// Ambil ID experience dari parameter URL
	id := c.Param("id")

	// Inisialisasi experience
	var experience models.Experience

	// Cari experience beserta images
	if err := database.DB.Preload("Images").First(&experience, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Experience not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus semua image dari lokal
	for _, image := range experience.Images {
		helpers.DeleteFile(image.ImageUrl)
	}

	// Hapus experience dari database (images ikut terhapus karena OnDelete:CASCADE)
	if err := database.DB.Delete(&experience).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete experience",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Experience deleted successfully",
		Data:    nil,
	})
}

// POST /api/experiences/:id/images — tambah image ke experience (auth)
func AddExperienceImage(c *gin.Context) {

	// Ambil ID experience dari parameter URL
	id := c.Param("id")

	// Inisialisasi experience
	var experience models.Experience

	// Cari experience berdasarkan ID
	if err := database.DB.First(&experience, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Experience not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload image
	path, err := helpers.UploadFile(c, "image", "experiences")
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Success: false,
			Message: "Failed to upload image",
			Errors:  map[string]string{"image": err.Error()},
		})
		return
	}

	// Simpan image ke database
	image := models.ExperienceImage{
		ExperienceId: experience.Id,
		ImageUrl:     helpers.GetFileUrl(path),
	}

	if err := database.DB.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to save image",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Image added successfully",
		Data:    image,
	})
}

// DELETE /api/experiences/:id/images/:imageId — hapus image dari experience (auth)
func DeleteExperienceImage(c *gin.Context) {

	// Ambil imageId dari parameter URL
	imageId := c.Param("imageId")

	// Inisialisasi image
	var image models.ExperienceImage

	// Cari image berdasarkan ID
	if err := database.DB.First(&image, imageId).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Image not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus file dari lokal
	helpers.DeleteFile(image.ImageUrl)

	// Hapus image dari database
	if err := database.DB.Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete image",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Image deleted successfully",
		Data:    nil,
	})
}
