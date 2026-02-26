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

	query.Order("bookmarks.id asc").
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

// POST /api/bookmarks/sync — sync semua repo dari GitHub (auth)
func SyncAllBookmarks(c *gin.Context) {

	// Fetch semua repo dari GitHub
	repos, err := helpers.FetchAllGithubRepos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to fetch GitHub repositories",
			Errors:  map[string]string{"github": err.Error()},
		})
		return
	}

	var created, updated int

	for _, repo := range repos {
		var bookmark models.Bookmark

		// Cek apakah repo sudah ada di DB by URL
		if err := database.DB.Where("url = ?", repo.HtmlUrl).First(&bookmark).Error; err != nil {
			// Belum ada, buat baru
			bookmark = models.Bookmark{
				Url:         repo.HtmlUrl,
				Title:       repo.Name,
				Description: repo.Description,
			}
			database.DB.Create(&bookmark)

			// Simpan topics
			for _, topic := range repo.Topics {
				database.DB.Create(&models.BookmarkTopic{
					BookmarkId: bookmark.Id,
					Name:       topic,
				})
			}
			created++
		} else {
			// Sudah ada, update data terbaru
			bookmark.Title = repo.Name
			bookmark.Description = repo.Description
			database.DB.Save(&bookmark)

			// Reset topics lama lalu simpan yang baru
			database.DB.Where("bookmark_id = ?", bookmark.Id).Delete(&models.BookmarkTopic{})
			for _, topic := range repo.Topics {
				database.DB.Create(&models.BookmarkTopic{
					BookmarkId: bookmark.Id,
					Name:       topic,
				})
			}
			updated++
		}
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Bookmarks synced successfully",
		Data: map[string]any{
			"total":   len(repos),
			"created": created,
			"updated": updated,
		},
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

	// Hapus topics dulu (sudah handle otomatis karena OnDelete:CASCADE)
	if err := database.DB.Delete(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete bookmark",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Bookmark deleted successfully",
		Data:    nil,
	})
}
