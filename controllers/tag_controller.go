package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/tags — ambil semua tag dengan pagination & search (publik)
func FindTags(c *gin.Context) {

	var tags []models.Tag
	var total int64

	search := c.Query("search")
	pg := helpers.GetPagination(c)

	query := database.DB.Model(&models.Tag{})

	if search != "" {
		query = query.Where("name LIKE ? OR slug LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	query.Order("name asc").
		Limit(pg.Limit).
		Offset(pg.Offset).
		Find(&tags)

	totalPages := int(total) / pg.Limit
	if int(total)%pg.Limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, structs.PaginatedResponse{
		Success: true,
		Message: "List Data Tags",
		Data:    tags,
		Meta: structs.PaginationMeta{
			Page:       pg.Page,
			Limit:      pg.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GET /api/tags/:id — ambil detail tag by id (publik)
func FindTagById(c *gin.Context) {

	// Ambil ID tag dari parameter URL
	id := c.Param("id")

	// Inisialisasi tag
	var tag models.Tag

	// Cari tag berdasarkan ID
	if err := database.DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tag not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data tag
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tag Found",
		Data:    tag,
	})
}

// GET /api/tags/slug/:slug — ambil detail tag by slug (publik)
func FindTagBySlug(c *gin.Context) {

	// Ambil slug tag dari parameter URL
	slug := c.Param("slug")

	// Inisialisasi tag
	var tag models.Tag

	// Cari tag berdasarkan slug
	if err := database.DB.Where("slug = ?", slug).First(&tag).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tag not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses dengan data tag
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tag Found",
		Data:    tag,
	})
}

// POST /api/tags — buat tag baru (auth)
func CreateTag(c *gin.Context) {

	// Struct tag request
	var req structs.TagCreateRequest

	// Bind JSON request ke struct TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Cek apakah tag dengan nama yang sama sudah ada
	var existing models.Tag
	if err := database.DB.Where("slug = ?", helpers.GenerateSlug(req.Name)).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, structs.ErrorResponse{
			Success: false,
			Message: "Tag already exists",
			Errors:  map[string]string{"name": "tag with this name already exists"},
		})
		return
	}

	// Inisialisasi tag baru
	tag := models.Tag{
		Name: req.Name,
		Slug: helpers.GenerateSlug(req.Name),
	}

	// Simpan tag ke database
	if err := database.DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create tag",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Tag created successfully",
		Data:    tag,
	})
}

// PUT /api/tags/:id — update tag (auth)
func UpdateTag(c *gin.Context) {

	// Ambil ID tag dari parameter URL
	id := c.Param("id")

	// Inisialisasi tag
	var tag models.Tag

	// Cari tag berdasarkan ID
	if err := database.DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tag not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Struct tag request
	var req structs.TagUpdateRequest

	// Bind JSON request ke struct TagUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Cek apakah slug baru sudah dipakai tag lain
	newSlug := helpers.GenerateSlug(req.Name)
	var existing models.Tag
	if err := database.DB.Where("slug = ? AND id != ?", newSlug, tag.Id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, structs.ErrorResponse{
			Success: false,
			Message: "Tag already exists",
			Errors:  map[string]string{"name": "tag with this name already exists"},
		})
		return
	}

	// Update data tag
	tag.Name = req.Name
	tag.Slug = newSlug

	// Simpan perubahan ke database
	if err := database.DB.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update tag",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Kirimkan response sukses
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tag updated successfully",
		Data:    tag,
	})
}

// DELETE /api/tags/:id — hapus tag (auth)
func DeleteTag(c *gin.Context) {

	id := c.Param("id")
	var tag models.Tag

	if err := database.DB.First(&tag, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tag not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus relasi blog_tags dulu sebelum hapus tag
	database.DB.Model(&tag).Association("Blogs").Clear()

	if err := database.DB.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete tag",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tag deleted successfully",
		Data:    nil,
	})
}
