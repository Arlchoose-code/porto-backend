package models

import "time"

type Setting struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"column:key;unique;not null"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
