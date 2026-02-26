package models

import "time"

type Tool struct {
	Id          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug" gorm:"unique;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Category    string    `json:"category"`
	Icon        string    `json:"icon"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Order       int       `json:"order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
