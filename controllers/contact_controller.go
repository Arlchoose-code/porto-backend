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

// GET /api/contacts — ambil semua pesan (admin only)
func FindContacts(c *gin.Context) {

	// Inisialisasi slice untuk menampung data contact
	var contacts []models.Contact

	// Bisa filter by status via query param, contoh: /api/contacts?status=pending
	status := c.Query("status")
	if status != "" {
		database.DB.Where("status = ?", status).Order("created_at desc").Find(&contacts)
	} else {
		database.DB.Order("created_at desc").Find(&contacts)
	}

	// Kirimkan response sukses dengan data contact
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Contacts",
		Data:    contacts,
	})
}

// GET /api/contacts/:id — ambil detail satu pesan (admin only)
func FindContactById(c *gin.Context) {

	// Ambil ID contact dari parameter URL
	id := c.Param("id")

	// Inisialisasi contact
	var contact models.Contact

	// Cari contact berdasarkan ID
	if err := database.DB.First(&contact, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Contact not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data contact
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Contact Found",
		Data:    contact,
	})
}

// POST /api/contacts — kirim pesan baru (publik)
func CreateContact(c *gin.Context) {

	// Struct contact request
	var req structs.ContactCreateRequest

	// Bind JSON request ke struct ContactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Inisialisasi contact baru
	contact := models.Contact{
		Name:    req.Name,
		Email:   req.Email,
		Subject: req.Subject,
		Message: req.Message,
		Status:  "pending",
	}

	// Simpan contact ke database
	if err := database.DB.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to send message",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Message sent successfully",
		Data:    contact,
	})
}

// PUT /api/contacts/:id/status — update status pesan (admin only)
func UpdateContactStatus(c *gin.Context) {

	// Ambil ID contact dari parameter URL
	id := c.Param("id")

	// Inisialisasi contact
	var contact models.Contact

	// Cari contact berdasarkan ID
	if err := database.DB.First(&contact, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Contact not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Struct contact request
	var req structs.ContactUpdateStatusRequest

	// Bind JSON request ke struct ContactUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Update status dan catat timestamp sesuai status baru
	now := time.Now()
	contact.Status = req.Status

	if req.Status == "read" && contact.ReadAt == nil {
		contact.ReadAt = &now
	}

	if req.Status == "done" {
		if contact.ReadAt == nil {
			contact.ReadAt = &now
		}
		contact.DoneAt = &now
	}

	// Simpan perubahan ke database
	if err := database.DB.Save(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update contact status",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Contact status updated successfully",
		Data:    contact,
	})
}

// DELETE /api/contacts/:id — hapus pesan (admin only)
func DeleteContact(c *gin.Context) {

	// Ambil ID contact dari parameter URL
	id := c.Param("id")

	// Inisialisasi contact
	var contact models.Contact

	// Cari contact berdasarkan ID
	if err := database.DB.First(&contact, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Contact not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus contact dari database
	if err := database.DB.Delete(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete contact",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Contact deleted successfully",
		Data:    nil,
	})
}
