package structs

type ProfileRequest struct {
	Name      string `form:"name" binding:"required"`
	Tagline   string `form:"tagline"`
	Bio       string `form:"bio"`
	Github    string `form:"github"`
	Linkedin  string `form:"linkedin"`
	Twitter   string `form:"twitter"`
	Instagram string `form:"instagram"`
	Email     string `form:"email"`
	Phone     string `form:"phone"`
	Location  string `form:"location"`
}
