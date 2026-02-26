package structs

type BookmarkRequest struct {
	Url         string   `json:"url" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}
