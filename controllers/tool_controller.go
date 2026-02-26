package controllers

import (
	"arlchoose/backend-api/database"
	"arlchoose/backend-api/helpers"
	"arlchoose/backend-api/models"
	"arlchoose/backend-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/tools — list tools aktif (publik)
func FindTools(c *gin.Context) {

	var tools []models.Tool

	category := c.Query("category")
	query := database.DB.Where("is_active = ?", true)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Order("`order` asc, created_at desc").Find(&tools)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Tools",
		Data:    tools,
	})
}

// GET /api/tools/all — list semua tools (auth, termasuk nonaktif)
func FindAllTools(c *gin.Context) {

	var tools []models.Tool
	database.DB.Order("`order` asc, created_at desc").Find(&tools)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "List Data Tools",
		Data:    tools,
	})
}

// GET /api/tools/registry — daftar slug + input_schema dari registry Go
// Endpoint ini yang dipakai frontend untuk tahu tool apa saja yang tersedia
// dan apa saja field input-nya — sehingga TIDAK PERLU hardcode di frontend.
func FindRegistry(c *gin.Context) {

	type RegistryItem struct {
		Slug        string                `json:"slug"`
		Name        string                `json:"name"`
		InputSchema []helpers.FieldSchema `json:"input_schema"`
		Docs        *helpers.ToolDocs     `json:"docs,omitempty"`
	}

	items := make([]RegistryItem, 0, len(helpers.ToolRegistry))
	for slug, meta := range helpers.ToolRegistry {
		items = append(items, RegistryItem{
			Slug:        slug,
			Name:        meta.Name,
			InputSchema: meta.InputSchema,
			Docs:        meta.Docs,
		})
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool Registry",
		Data:    items,
	})
}

// GET /api/tools/:slug — detail tool (publik)
// Sekarang juga menyertakan input_schema dari registry
func FindToolBySlug(c *gin.Context) {

	slug := c.Param("slug")
	var tool models.Tool

	if err := database.DB.Where("slug = ? AND is_active = ?", slug, true).First(&tool).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tool not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Sertakan input_schema + docs dari registry
	schema, _ := helpers.GetInputSchema(slug)
	docs := helpers.GetDocs(slug)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool Found",
		Data: map[string]any{
			"tool":         tool,
			"input_schema": schema,
			"docs":         docs,
		},
	})
}

// POST/GET /api/tools/:slug/run — eksekusi tool (publik, rate limited)
func RunTool(c *gin.Context) {

	slug := c.Param("slug")
	var tool models.Tool

	if err := database.DB.Where("slug = ? AND is_active = ?", slug, true).First(&tool).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tool not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Ambil input dari query (GET) atau body (POST)
	input := make(map[string]any)
	if c.Request.Method == "GET" {
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				input[key] = values[0]
			}
		}
	} else {
		c.ShouldBindJSON(&input)
	}

	// Eksekusi Go function — murni Go, super cepat
	result, err := helpers.ExecuteTool(slug, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Success: false,
			Message: err.Error(),
			Errors:  map[string]string{"tool": err.Error()},
		})
		return
	}

	// Catat usage secara async — tidak blocking response
	go func() {
		usage := models.ToolUsage{
			ToolId:   tool.Id,
			ToolSlug: tool.Slug,
			IP:       c.ClientIP(),
		}
		database.DB.Create(&usage)
	}()

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool executed successfully",
		Data:    result,
	})
}

// GET /api/tools/stats — statistik penggunaan tools (auth)
func ToolStats(c *gin.Context) {

	// Total usage per tool (join dengan nama tool)
	type ToolUsageStat struct {
		ToolSlug  string `json:"tool_slug"`
		ToolName  string `json:"tool_name"`
		TotalRuns int    `json:"total_runs"`
	}

	var perTool []ToolUsageStat
	database.DB.Raw(`
		SELECT tu.tool_slug, t.name as tool_name, COUNT(*) as total_runs
		FROM tool_usages tu
		LEFT JOIN tools t ON t.slug = tu.tool_slug
		GROUP BY tu.tool_slug, t.name
		ORDER BY total_runs DESC
	`).Scan(&perTool)

	// Usage 7 hari terakhir (per hari)
	type DailyUsage struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}
	var daily []DailyUsage
	database.DB.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM tool_usages
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`).Scan(&daily)

	// Total keseluruhan
	var totalAll int64
	database.DB.Model(&models.ToolUsage{}).Count(&totalAll)

	// Total hari ini
	var todayCount int64
	database.DB.Model(&models.ToolUsage{}).
		Where("DATE(created_at) = CURDATE()").
		Count(&todayCount)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool Stats",
		Data: map[string]any{
			"per_tool":    perTool,
			"daily":       daily,
			"total_all":   totalAll,
			"today_count": todayCount,
		},
	})
}

func CreateTool(c *gin.Context) {

	var req structs.ToolRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Validasi slug ada di registry
	if _, exists := helpers.ToolRegistry[req.Slug]; !exists {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Success: false,
			Message: "Handler not found in registry",
			Errors:  map[string]string{"slug": "slug '" + req.Slug + "' tidak ada di registry"},
		})
		return
	}

	// Upload icon jika ada
	iconUrl := ""
	if _, err := c.FormFile("icon"); err == nil {
		path, err := helpers.UploadFile(c, "icon", "tools")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload icon",
				Errors:  map[string]string{"icon": err.Error()},
			})
			return
		}
		iconUrl = helpers.GetFileUrl(path)
	}

	tool := models.Tool{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Category:    req.Category,
		Icon:        iconUrl,
		IsActive:    req.IsActive,
		Order:       req.Order,
	}

	if err := database.DB.Create(&tool).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to create tool",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Tool created successfully",
		Data:    tool,
	})
}

func UpdateTool(c *gin.Context) {

	id := c.Param("id")
	var tool models.Tool

	if err := database.DB.First(&tool, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tool not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	var req structs.ToolRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, structs.ErrorResponse{
			Success: false,
			Message: "Validation Errors",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Upload icon baru jika ada, hapus yang lama
	if _, err := c.FormFile("icon"); err == nil {
		helpers.DeleteFile(tool.Icon)
		path, err := helpers.UploadFile(c, "icon", "tools")
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Success: false,
				Message: "Failed to upload icon",
				Errors:  map[string]string{"icon": err.Error()},
			})
			return
		}
		tool.Icon = helpers.GetFileUrl(path)
	}

	tool.Name = req.Name
	tool.Description = req.Description
	tool.Category = req.Category
	tool.IsActive = req.IsActive
	tool.Order = req.Order

	if err := database.DB.Save(&tool).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to update tool",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool updated successfully",
		Data:    tool,
	})
}

// PUT /api/tools/:id/toggle — on/off tool (auth)
func ToggleTool(c *gin.Context) {

	id := c.Param("id")
	var tool models.Tool

	if err := database.DB.First(&tool, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tool not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	// Toggle is_active
	tool.IsActive = !tool.IsActive

	if err := database.DB.Save(&tool).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to toggle tool",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	status := "activated"
	if !tool.IsActive {
		status = "deactivated"
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool " + status + " successfully",
		Data:    tool,
	})
}

// DELETE /api/tools/:id — hapus tool (auth)
func DeleteTool(c *gin.Context) {

	id := c.Param("id")
	var tool models.Tool

	if err := database.DB.First(&tool, id).Error; err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Success: false,
			Message: "Tool not found",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	if err := database.DB.Delete(&tool).Error; err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Success: false,
			Message: "Failed to delete tool",
			Errors:  helpers.TranslateErrorMessage(err),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Tool deleted successfully",
		Data:    nil,
	})
}
