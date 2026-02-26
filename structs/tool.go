package structs

type ToolRequest struct {
	Name        string `form:"name" binding:"required"`
	Slug        string `form:"slug" binding:"required"`
	Description string `form:"description"`
	Category    string `form:"category"`
	IsActive    bool   `form:"is_active"`
	Order       int    `form:"order"`
}
