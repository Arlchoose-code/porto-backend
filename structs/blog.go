package structs

type BlogCreateRequest struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	Content     string `form:"content"`
	TagIds      []uint `form:"tag_ids"`
}

type BlogUpdateRequest struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	Content     string `form:"content"`
	TagIds      []uint `form:"tag_ids"`
	UpdateTags  bool   `form:"update_tags"`
}

type AiBlogGenerateRequest struct {
	Keyword string `json:"keyword"`
	Total   int    `json:"total" binding:"required,min=1,max=10"`
}

type BlogRejectRequest struct {
	Comment string `json:"comment" binding:"required"`
}

// Bulk action request
type BlogBulkRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Action string `json:"action" binding:"required,oneof=publish reject archive delete"`
	// Hanya dipakai kalau action = reject
	Comment string `json:"comment"`
}
