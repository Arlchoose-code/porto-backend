package structs

// Struct ini digunakan saat membuat experience baru
type ExperienceCreateRequest struct {
	Company     string `form:"company" binding:"required"`
	Role        string `form:"role" binding:"required"`
	Location    string `form:"location"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	IsCurrent   bool   `form:"is_current"`
	Description string `form:"description"`
}

// Struct ini digunakan saat mengupdate experience
type ExperienceUpdateRequest struct {
	Company     string `form:"company" binding:"required"`
	Role        string `form:"role" binding:"required"`
	Location    string `form:"location"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	IsCurrent   bool   `form:"is_current"`
	Description string `form:"description"`
}
