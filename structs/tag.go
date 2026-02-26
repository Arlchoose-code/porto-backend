package structs

// Struct ini digunakan saat membuat tag baru
type TagCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// Struct ini digunakan saat mengupdate tag
type TagUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}
