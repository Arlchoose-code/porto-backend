package structs

// Struct ini digunakan saat membuat skill baru
type SkillCreateRequest struct {
	Name     string `form:"name" binding:"required"`
	Category string `form:"category" binding:"required,oneof=language framework database tool other"`
	Level    string `form:"level" binding:"required,oneof=beginner intermediate advanced expert"`
	Order    int    `form:"order"`
}

// Struct ini digunakan saat mengupdate skill
type SkillUpdateRequest struct {
	Name     string `form:"name" binding:"required"`
	Category string `form:"category" binding:"required,oneof=language framework database tool other"`
	Level    string `form:"level" binding:"required,oneof=beginner intermediate advanced expert"`
	Order    int    `form:"order"`
}
