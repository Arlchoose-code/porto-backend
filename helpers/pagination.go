package helpers

import "github.com/gin-gonic/gin"

type PaginationParams struct {
	Page  int
	Limit int
	// Offset untuk GORM
	Offset int
}

// GetPagination ambil page & limit dari query string
func GetPagination(c *gin.Context) PaginationParams {

	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if val := atoiSafe(p); val > 0 {
			page = val
		}
	}

	if l := c.Query("limit"); l != "" {
		if val := atoiSafe(l); val > 0 && val <= 100 {
			limit = val
		}
	}

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

// atoiSafe konversi string ke int tanpa panic
func atoiSafe(s string) int {
	result := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		result = result*10 + int(ch-'0')
	}
	return result
}
