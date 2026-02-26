package models

import "time"

type Profile struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Tagline   string    `json:"tagline"`
	Bio       string    `json:"bio" gorm:"type:text"`
	Avatar    string    `json:"avatar"`
	ResumeUrl string    `json:"resume_url"`
	Github    string    `json:"github"`
	Linkedin  string    `json:"linkedin"`
	Twitter   string    `json:"twitter"`
	Instagram string    `json:"instagram"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
