package structs

// Struct ini digunakan saat membuat project baru
type ProjectCreateRequest struct {
	Title       string   `form:"title" binding:"required"`
	Description string   `form:"description"`
	Platform    string   `form:"platform"`
	Url         string   `form:"url"`
	TechStacks  []string `form:"tech_stacks"`
}

// Struct ini digunakan saat mengupdate project
type ProjectUpdateRequest struct {
	Title       string   `form:"title" binding:"required"`
	Description string   `form:"description"`
	Platform    string   `form:"platform"`
	Url         string   `form:"url"`
	TechStacks  []string `form:"tech_stacks"`
}
