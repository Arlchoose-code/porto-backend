package structs

// Struct ini digunakan saat membuat course baru
type CourseCreateRequest struct {
	Title         string `form:"title" binding:"required"`
	Issuer        string `form:"issuer"`
	IssuedAt      string `form:"issued_at"`
	ExpiredAt     string `form:"expired_at"`
	CredentialUrl string `form:"credential_url"`
	Description   string `form:"description"`
}

// Struct ini digunakan saat mengupdate course
type CourseUpdateRequest struct {
	Title         string `form:"title" binding:"required"`
	Issuer        string `form:"issuer"`
	IssuedAt      string `form:"issued_at"`
	ExpiredAt     string `form:"expired_at"`
	CredentialUrl string `form:"credential_url"`
	Description   string `form:"description"`
}
