package structs

// Struct ini digunakan saat membuat education baru
type EducationCreateRequest struct {
	School      string `form:"school" binding:"required"`
	Degree      string `form:"degree"`
	Field       string `form:"field"`
	StartYear   int    `form:"start_year"`
	EndYear     int    `form:"end_year"`
	Description string `form:"description"`
}

// Struct ini digunakan saat mengupdate education
type EducationUpdateRequest struct {
	School      string `form:"school" binding:"required"`
	Degree      string `form:"degree"`
	Field       string `form:"field"`
	StartYear   int    `form:"start_year"`
	EndYear     int    `form:"end_year"`
	Description string `form:"description"`
}
