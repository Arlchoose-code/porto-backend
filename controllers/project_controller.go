package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/projects — ambil semua project (publik)
func FindProjects(c *gin.Context) {

	var projects []models.Project

	database.DB.Preload("TechStacks").Preload("Images").Order("created_at desc").Find(&projects)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Projects",
		Data:    projects,
	})
}

// GET /api/projects/:slug — ambil detail project by slug (publik)
func FindProjectBySlug(c *gin.Context) {

	slug := c.Param("slug")
	var project models.Project

	if err := database.DB.Preload("TechStacks").Preload("Images").Where("slug = ?", slug).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Project not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Project Found",
		Data:    project,
	})
}

// POST /api/projects — buat project baru (auth)
func CreateProject(c *gin.Context) {

	var req structs.ProjectCreateRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	project := models.Project{
		Title:       req.Title,
		Slug:        helpers.GenerateSlug(req.Title),
		Description: req.Description,
		Platform:    req.Platform,
		Url:         req.Url,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create project",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Simpan tech stacks jika ada
	for _, tech := range req.TechStacks {
		database.DB.Create(&models.ProjectTechStack{
			ProjectId: project.Id,
			Name:      tech,
		})
	}

	// Upload images jika ada (multiple)
	form, _ := c.MultipartForm()
	if form != nil {
		files := form.File["images"]
		for i, file := range files {
			c.Request.MultipartForm.File["images"] = []*multipart.FileHeader{file}
			path, err := helpers.UploadFile(c, "images", "projects")
			if err != nil {
				continue
			}
			database.DB.Create(&models.ProjectImage{
				ProjectId: project.Id,
				ImageUrl:  helpers.GetFileUrl(path),
				Order:     i,
			})
		}
	}

	database.DB.Preload("TechStacks").Preload("Images").First(&project, project.Id)

	go helpers.RevalidateFrontend("project", project.Slug)

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Project created successfully",
		Data:    project,
	})
}

// PUT /api/projects/:id — update project (auth)
func UpdateProject(c *gin.Context) {

	id := c.Param("id")
	var project models.Project

	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Project not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	var req structs.ProjectUpdateRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	project.Title = req.Title
	project.Slug = helpers.GenerateSlug(req.Title)
	project.Description = req.Description
	project.Platform = req.Platform
	project.Url = req.Url

	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update project",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Reset tech stacks lama lalu simpan yang baru
	database.DB.Where("project_id = ?", project.Id).Delete(&models.ProjectTechStack{})
	for _, tech := range req.TechStacks {
		database.DB.Create(&models.ProjectTechStack{
			ProjectId: project.Id,
			Name:      tech,
		})
	}

	database.DB.Preload("TechStacks").Preload("Images").First(&project, project.Id)

	go helpers.RevalidateFrontend("project", project.Slug)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Project updated successfully",
		Data:    project,
	})
}

// DELETE /api/projects/:id — hapus project (auth)
func DeleteProject(c *gin.Context) {

	id := c.Param("id")
	var project models.Project

	if err := database.DB.Preload("Images").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Project not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Hapus semua image dari lokal
	for _, image := range project.Images {
		helpers.DeleteFile(image.ImageUrl)
	}

	if err := database.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete project",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	go helpers.RevalidateFrontend("project", "")

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Project deleted successfully",
		Data:    nil,
	})
}

// POST /api/projects/:id/images — tambah image ke project (auth)
func AddProjectImage(c *gin.Context) {

	id := c.Param("id")
	var project models.Project

	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Project not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	path, err := helpers.UploadFile(c, "image", "projects")
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Success: false,
			Message: "Failed to upload image",
			Errors:  map[string]string{"image": err.Error()},
		})
		return
	}

	image := models.ProjectImage{
		ProjectId: project.Id,
		ImageUrl:  helpers.GetFileUrl(path), // ✅ fix — pakai GetFileUrl
	}

	if err := database.DB.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to save image",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Image added successfully",
		Data:    image,
	})
}

// DELETE /api/projects/:id/images/:imageId — hapus image dari project (auth)
func DeleteProjectImage(c *gin.Context) {

	imageId := c.Param("imageId")
	var image models.ProjectImage

	if err := database.DB.First(&image, imageId).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Image not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	helpers.DeleteFile(image.ImageUrl)

	if err := database.DB.Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete image",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Image deleted successfully",
		Data:    nil,
	})
}
