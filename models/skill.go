package models

import "time"

type Skill struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	Category  string    `json:"category" gorm:"type:enum('language','framework','database','tool','other');not null"`
	Name      string    `json:"name" gorm:"not null"`
	Level     string    `json:"level" gorm:"type:enum('beginner','intermediate','advanced','expert');not null"`
	IconUrl   string    `json:"icon_url"`
	Order     int       `json:"order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
