package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /api/courses — ambil semua course (publik)
func FindCourses(c *gin.Context) {

	// Inisialisasi slice untuk menampung data course
	var courses []models.Course

	// Ambil data course diurutkan dari yang terbaru
	database.DB.Order("issued_at desc").Find(&courses)

	// Kirimkan response sukses dengan data course
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Courses",
		Data:    courses,
	})
}

// GET /api/courses/:id — ambil detail course (publik)
func FindCourseById(c *gin.Context) {

	// Ambil ID course dari parameter URL
	id := c.Param("id")

	// Inisialisasi course
	var course models.Course

	// Cari course berdasarkan ID
	if err := database.DB.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Course not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data course
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Course Found",
		Data:    course,
	})
}

// POST /api/courses — buat course baru (auth)
func CreateCourse(c *gin.Context) {

	// Struct course request
	var req structs.CourseCreateRequest

	// Bind form data ke struct CourseCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// CREATE — Upload certificate image jika ada
	certificateImage := ""
	if _, err := c.FormFile("certificate_image"); err == nil {
		path, err := helpers.UploadFile(c, "certificate_image", "certificates")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload certificate image",
				Errors:  map[string]string{"certificate_image": err.Error()},
			})
			return
		}
		certificateImage = helpers.GetFileUrl(path)
	}

	// Inisialisasi course baru
	course := models.Course{
		Title:            req.Title,
		Issuer:           req.Issuer,
		CredentialUrl:    req.CredentialUrl,
		CertificateImage: certificateImage,
		Description:      req.Description,
	}

	// Parse issued_at jika ada
	if req.IssuedAt != "" {
		t, err := time.Parse("2006-01-02", req.IssuedAt)
		if err == nil {
			course.IssuedAt = &t
		}
	}

	// Parse expired_at jika ada
	if req.ExpiredAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiredAt)
		if err == nil {
			course.ExpiredAt = &t
		}
	}

	// Simpan course ke database
	if err := database.DB.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create course",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Course created successfully",
		Data:    course,
	})
}

// PUT /api/courses/:id — update course (auth)
func UpdateCourse(c *gin.Context) {

	// Ambil ID course dari parameter URL
	id := c.Param("id")

	// Inisialisasi course
	var course models.Course

	// Cari course berdasarkan ID
	if err := database.DB.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Course not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Struct course request
	var req structs.CourseUpdateRequest

	// Bind form data ke struct CourseUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// UPDATE — Upload certificate baru jika ada, hapus yang lama
	if _, err := c.FormFile("certificate_image"); err == nil {
		helpers.DeleteFile(course.CertificateImage)
		path, err := helpers.UploadFile(c, "certificate_image", "certificates")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload certificate image",
				Errors:  map[string]string{"certificate_image": err.Error()},
			})
			return
		}
		course.CertificateImage = helpers.GetFileUrl(path)
	}

	// Update data course
	course.Title = req.Title
	course.Issuer = req.Issuer
	course.CredentialUrl = req.CredentialUrl
	course.Description = req.Description

	// Reset dulu supaya bisa di-update ke null kalau dikosongkan
	course.IssuedAt = nil
	course.ExpiredAt = nil

	// Parse issued_at jika ada
	if req.IssuedAt != "" {
		t, err := time.Parse("2006-01-02", req.IssuedAt)
		if err == nil {
			course.IssuedAt = &t
		}
	}

	// Parse expired_at jika ada
	if req.ExpiredAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiredAt)
		if err == nil {
			course.ExpiredAt = &t
		}
	}

	// Simpan perubahan ke database
	if err := database.DB.Save(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update course",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Course updated successfully",
		Data:    course,
	})
}

// DELETE /api/courses/:id — hapus course (auth)
func DeleteCourse(c *gin.Context) {

	// Ambil ID course dari parameter URL
	id := c.Param("id")

	// Inisialisasi course
	var course models.Course

	// Cari course berdasarkan ID
	if err := database.DB.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Course not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus certificate image dari lokal jika ada
	helpers.DeleteFile(course.CertificateImage)

	// Hapus course dari database
	if err := database.DB.Delete(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete course",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Course deleted successfully",
		Data:    nil,
	})
}
