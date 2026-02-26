package structs

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}
