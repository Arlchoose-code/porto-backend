package controllers

import (
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /api/upload — upload file (auth)
func UploadFile(c *gin.Context) {

	// Ambil folder tujuan dari query param, contoh: /api/upload?folder=projects
	// Default ke "general" kalau ga diisi
	folder := c.DefaultQuery("folder", "general")

	// Upload file menggunakan helper
	filePath, err := helpers.UploadFile(c, "file", folder)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Success: false,
			Message: "Failed to upload file",
			Errors:  map[string]string{"file": err.Error()},
		})
		return
	}

	// Kirimkan response sukses dengan path file
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "File uploaded successfully",
		Data: map[string]string{
			"path": filePath,
			"url":  "/" + filePath,
		},
	})
}

// DELETE /api/upload — hapus file (auth)
func DeleteFile(c *gin.Context) {

	// Ambil path file dari body request
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  map[string]string{"path": "path is required"},
		})
		return
	}

	// Pastikan path tidak keluar dari folder uploads/ (keamanan)
	if len(req.Path) < 8 || req.Path[:8] != "uploads/" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Success: false,
			Message: "Invalid file path",
			Errors:  map[string]string{"path": "path must start with uploads/"},
		})
		return
	}

	// Hapus file menggunakan helper
	if err := helpers.DeleteFile(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete file",
			Errors:  map[string]string{"file": err.Error()},
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "File deleted successfully",
		Data:    nil,
	})
}
