package models

import "time"

type Education struct {
	Id          uint      `json:"id" gorm:"primaryKey"`
	School      string    `json:"school" gorm:"not null"`
	Degree      string    `json:"degree"`
	Field       string    `json:"field"`
	StartYear   int       `json:"start_year"`
	EndYear     int       `json:"end_year"`
	Description string    `json:"description" gorm:"type:text"`
	LogoUrl     string    `json:"logo_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
