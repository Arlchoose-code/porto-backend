package models

import "time"

type Experience struct {
	Id          uint              `json:"id" gorm:"primaryKey"`
	Company     string            `json:"company" gorm:"not null"`
	Role        string            `json:"role" gorm:"not null"`
	Location    string            `json:"location"`
	StartDate   *time.Time        `json:"start_date"`
	EndDate     *time.Time        `json:"end_date"`
	IsCurrent   bool              `json:"is_current" gorm:"default:false"`
	Description string            `json:"description" gorm:"type:text"`
	Images      []ExperienceImage `json:"images" gorm:"foreignKey:ExperienceId;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ExperienceImage struct {
	Id           uint   `json:"id" gorm:"primaryKey"`
	ExperienceId uint   `json:"experience_id"`
	ImageUrl     string `json:"image_url" gorm:"not null"`
	Order        int    `json:"order" gorm:"default:0"`
}
