package models

import "time"

type Contact struct {
	Id        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name" gorm:"not null"`
	Email     string     `json:"email" gorm:"not null"`
	Subject   string     `json:"subject" gorm:"not null"`
	Message   string     `json:"message" gorm:"type:text;not null"`
	Status    string     `json:"status" gorm:"type:enum('pending','read','done');default:'pending'"`
	ReadAt    *time.Time `json:"read_at"`
	DoneAt    *time.Time `json:"done_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
