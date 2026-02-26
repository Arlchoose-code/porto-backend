package models

import "time"

type Tag struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"unique;not null"`
	Slug      string    `json:"slug" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
