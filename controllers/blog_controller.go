package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/blogs — publik, hanya published
func FindBlogs(c *gin.Context) {

	var blogs []models.Blog
	var total int64

	tag := c.Query("tag")
	search := c.Query("search")
	pg := helpers.GetPagination(c)

	query := database.DB.Model(&models.Blog{}).Preload("Tags").Preload("User").
		Where("status = ?", "published")

	if tag != "" {
		query = query.
			Joins("JOIN blog_tags ON blog_tags.blog_id = blogs.id").
			Joins("JOIN tags ON tags.id = blog_tags.tag_id").
			Where("tags.slug = ?", tag)
	}

	if search != "" {
		query = query.Where("blogs.title LIKE ? OR blogs.description LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	query.Order("blogs.created_at desc").Limit(pg.Limit).Offset(pg.Offset).Find(&blogs)

	totalPages := int(total) / pg.Limit
	if int(total)%pg.Limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, structs.PaginatedResponse{
		Success: true,
		Message: "List Data Blogs",
		Data:    blogs,
		Meta: structs.PaginationMeta{
			Page:       pg.Page,
			Limit:      pg.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GET /api/blogs/stats — hitung total per status (auth)
func BlogStats(c *gin.Context) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var results []StatusCount
	database.DB.Model(&models.Blog{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results)

	stats := map[string]int64{
		"total":     0,
		"published": 0,
		"pending":   0,
		"rejected":  0,
		"archived":  0,
	}

	for _, r := range results {
		stats[r.Status] = r.Count
		stats["total"] += r.Count
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog Stats",
		Data:    stats,
	})
}

// GET /api/blogs/all — semua status dengan pagination (auth)
func FindAllBlogs(c *gin.Context) {

	var blogs []models.Blog
	var total int64

	status := c.Query("status")
	tag := c.Query("tag")
	search := c.Query("search")
	pg := helpers.GetPagination(c)

	query := database.DB.Model(&models.Blog{}).Preload("Tags").Preload("User")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if tag != "" {
		query = query.
			Joins("JOIN blog_tags ON blog_tags.blog_id = blogs.id").
			Joins("JOIN tags ON tags.id = blog_tags.tag_id").
			Where("tags.slug = ?", tag)
	}

	if search != "" {
		query = query.Where("blogs.title LIKE ? OR blogs.description LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	query.Order("blogs.created_at desc").Limit(pg.Limit).Offset(pg.Offset).Find(&blogs)

	totalPages := int(total) / pg.Limit
	if int(total)%pg.Limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, structs.PaginatedResponse{
		Success: true,
		Message: "List Data Blogs",
		Data:    blogs,
		Meta: structs.PaginationMeta{
			Page:       pg.Page,
			Limit:      pg.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GET /api/blogs/:slug — detail by slug (publik)
func FindBlogBySlug(c *gin.Context) {

	slug := c.Param("slug")
	var blog models.Blog

	if err := database.DB.Preload("Tags").Preload("User").Where("slug = ?", slug).First(&blog).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Blog not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog Found",
		Data:    blog,
	})
}

// POST /api/blogs — buat blog baru (auth)
func CreateBlog(c *gin.Context) {

	var req structs.BlogCreateRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	userId := c.MustGet("userId").(uint)

	coverImage := ""
	if _, err := c.FormFile("cover_image"); err == nil {
		path, err := helpers.UploadFile(c, "cover_image", "blogs")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload cover image",
				Errors:  map[string]string{"cover_image": err.Error()},
			})
			return
		}
		coverImage = helpers.GetFileUrl(path)
	}

	blog := models.Blog{
		Title:       req.Title,
		Slug:        helpers.GenerateSlug(req.Title),
		Description: req.Description,
		Content:     req.Content,
		CoverImage:  coverImage,
		Author:      "user",
		Status:      "published",
		UserId:      &userId,
	}

	if err := database.DB.Create(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create blog",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if len(req.TagIds) > 0 {
		var tags []models.Tag
		if err := database.DB.Where("id IN ?", req.TagIds).Find(&tags).Error; err == nil {
			database.DB.Model(&blog).Association("Tags").Replace(tags)
		}
	}

	database.DB.Preload("Tags").Preload("User").First(&blog, blog.Id)

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Blog created successfully",
		Data:    blog,
	})
}

// PUT /api/blogs/:id — update blog (auth)
func UpdateBlog(c *gin.Context) {

	id := c.Param("id")
	var blog models.Blog

	if err := database.DB.Preload("Tags").First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Blog not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	var req structs.BlogUpdateRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if _, err := c.FormFile("cover_image"); err == nil {
		helpers.DeleteFile(blog.CoverImage)
		path, err := helpers.UploadFile(c, "cover_image", "blogs")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload cover image",
				Errors:  map[string]string{"cover_image": err.Error()},
			})
			return
		}
		blog.CoverImage = helpers.GetFileUrl(path)
	}

	blog.Title = req.Title
	blog.Slug = helpers.GenerateSlug(req.Title)
	blog.Description = req.Description
	blog.Content = req.Content

	if err := database.DB.Save(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update blog",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if req.UpdateTags {
		var tags []models.Tag
		if len(req.TagIds) > 0 {
			database.DB.Where("id IN ?", req.TagIds).Find(&tags)
		}
		database.DB.Model(&blog).Association("Tags").Replace(tags)
	}

	database.DB.Preload("Tags").Preload("User").First(&blog, blog.Id)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog updated successfully",
		Data:    blog,
	})
}

// DELETE /api/blogs/:id — hapus blog (auth)
func DeleteBlog(c *gin.Context) {

	id := c.Param("id")
	var blog models.Blog

	if err := database.DB.First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Blog not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	helpers.DeleteFile(blog.CoverImage)
	database.DB.Model(&blog).Association("Tags").Clear()

	if err := database.DB.Delete(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete blog",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog deleted successfully",
		Data:    nil,
	})
}

// PUT /api/blogs/:id/archive — archive blog (auth)
func ArchiveBlog(c *gin.Context) {
	id := c.Param("id")
	var blog models.Blog

	if err := database.DB.First(&blog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Blog not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	blog.Status = "archived"

	if err := database.DB.Save(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to archive blog",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Blog archived successfully",
		Data:    blog,
	})
}

// POST /api/blogs/bulk — bulk action (publish/reject/archive/delete)
func BulkActionBlog(c *gin.Context) {
	var req structs.BlogBulkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if req.Action == "reject" && req.Comment == "" {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Comment is required for reject action",
			Errors:  map[string]string{"comment": "required"},
		})
		return
	}

	var affected int64
	// Untuk menyimpan blog AI yang perlu di-regenerate
	var aiBlogs []models.Blog

	switch req.Action {
	case "publish":
		result := database.DB.Model(&models.Blog{}).
			Where("id IN ?", req.IDs).
			Updates(map[string]any{"status": "published", "reject_comment": ""})
		affected = result.RowsAffected

	case "archive":
		result := database.DB.Model(&models.Blog{}).
			Where("id IN ?", req.IDs).
			Update("status", "archived")
		affected = result.RowsAffected

	case "reject":
		// Update semua ke rejected dulu
		result := database.DB.Model(&models.Blog{}).
			Where("id IN ?", req.IDs).
			Updates(map[string]any{"status": "rejected", "reject_comment": req.Comment})
		affected = result.RowsAffected

		// Ambil semua blog AI dari IDs yang di-reject → perlu regenerate
		database.DB.Where("id IN ? AND author = ?", req.IDs, "aibys").Find(&aiBlogs)

		// Spawn goroutine regenerate untuk setiap blog AI
		for _, blog := range aiBlogs {
			blogCopy := blog
			go regenerateBlog(&blogCopy, req.Comment)
		}

	case "delete":
		var blogs []models.Blog
		database.DB.Where("id IN ?", req.IDs).Find(&blogs)
		for _, b := range blogs {
			helpers.DeleteFile(b.CoverImage)
			database.DB.Model(&b).Association("Tags").Clear()
		}
		result := database.DB.Where("id IN ?", req.IDs).Delete(&models.Blog{})
		affected = result.RowsAffected
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Bulk action completed",
		Data: map[string]any{
			"action":        req.Action,
			"affected":      affected,
			"ai_regenerate": len(aiBlogs), // Jumlah blog AI yang di-regenerate
		},
	})
}
