package structs

// Struct ini digunakan saat pengunjung mengirim pesan contact (publik)
type ContactCreateRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=100"`
	Email   string `json:"email" binding:"required,email,max=100"`
	Subject string `json:"subject" binding:"required,min=3,max=150"`
	Message string `json:"message" binding:"required,min=10,max=5000"`
}

// Struct ini digunakan saat admin mengupdate status contact
type ContactUpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending read done"`
}

// Struct ini digunakan untuk menampilkan data contact sebagai response API
type ContactResponse struct {
	Id        uint    `json:"id"`
	Email     string  `json:"email"`
	Subject   string  `json:"subject"`
	Message   string  `json:"message"`
	Status    string  `json:"status"`
	ReadAt    *string `json:"read_at"`
	DoneAt    *string `json:"done_at"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
