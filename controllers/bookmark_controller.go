package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/bookmarks — ambil semua bookmark dengan pagination (publik)
func FindBookmarks(c *gin.Context) {

	var bookmarks []models.Bookmark
	var total int64

	topic := c.Query("topic")
	search := c.Query("search")
	pg := helpers.GetPagination(c)

	query := database.DB.Model(&models.Bookmark{}).Preload("Topics")

	if topic != "" {
		query = query.
			Joins("JOIN bookmark_topics ON bookmark_topics.bookmark_id = bookmarks.id").
			Where("bookmark_topics.name = ?", topic)
	}

	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	query.Order("bookmarks.id desc").
		Limit(pg.Limit).
		Offset(pg.Offset).
		Find(&bookmarks)

	totalPages := int(total) / pg.Limit
	if int(total)%pg.Limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, structs.PaginatedResponse{
		Success: true,
		Message: "List Data Bookmarks",
		Data:    bookmarks,
		Meta: structs.PaginationMeta{
			Page:       pg.Page,
			Limit:      pg.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GET /api/bookmarks/:id — ambil detail bookmark (publik)
func FindBookmarkById(c *gin.Context) {

	id := c.Param("id")
	var bookmark models.Bookmark

	if err := database.DB.Preload("Topics").First(&bookmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Bookmark not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Bookmark Found",
		Data:    bookmark,
	})
}

// POST /api/bookmarks — buat bookmark baru manual (auth)
func CreateBookmark(c *gin.Context) {

	var req structs.BookmarkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	bookmark := models.Bookmark{
		Url:         req.Url,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := database.DB.Create(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create bookmark",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Simpan topics
	for _, topic := range req.Topics {
		if topic != "" {
			database.DB.Create(&models.BookmarkTopic{
				BookmarkId: bookmark.Id,
				Name:       topic,
			})
		}
	}

	// Reload dengan topics
	database.DB.Preload("Topics").First(&bookmark, bookmark.Id)

	go helpers.RevalidateFrontend("bookmark", "")

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Bookmark created successfully",
		Data:    bookmark,
	})
}

// PUT /api/bookmarks/:id — update bookmark (auth)
func UpdateBookmark(c *gin.Context) {

	id := c.Param("id")
	var bookmark models.Bookmark

	if err := database.DB.First(&bookmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Bookmark not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	var req structs.BookmarkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	bookmark.Url = req.Url
	bookmark.Title = req.Title
	bookmark.Description = req.Description

	if err := database.DB.Save(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update bookmark",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Reset topics lama lalu simpan yang baru
	database.DB.Where("bookmark_id = ?", bookmark.Id).Delete(&models.BookmarkTopic{})
	for _, topic := range req.Topics {
		if topic != "" {
			database.DB.Create(&models.BookmarkTopic{
				BookmarkId: bookmark.Id,
				Name:       topic,
			})
		}
	}

	// Reload dengan topics
	database.DB.Preload("Topics").First(&bookmark, bookmark.Id)

	go helpers.RevalidateFrontend("bookmark", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Bookmark updated successfully",
		Data:    bookmark,
	})
}

// DELETE /api/bookmarks/:id — hapus bookmark (auth)
func DeleteBookmark(c *gin.Context) {

	id := c.Param("id")
	var bookmark models.Bookmark

	if err := database.DB.First(&bookmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Bookmark not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if err := database.DB.Delete(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete bookmark",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	go helpers.RevalidateFrontend("bookmark", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Bookmark deleted successfully",
		Data:    nil,
	})
}
