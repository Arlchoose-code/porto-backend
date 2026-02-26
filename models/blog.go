package models

import "time"

type Blog struct {
	Id            uint      `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title" gorm:"not null"`
	Slug          string    `json:"slug" gorm:"unique;not null"`
	Description   string    `json:"description" gorm:"type:text"`
	Content       string    `json:"content" gorm:"type:longtext"`
	CoverImage    string    `json:"cover_image"`
	Author        string    `json:"author" gorm:"type:enum('user','aibys');default:'user'"`
	Status        string    `json:"status" gorm:"type:enum('pending','published','rejected','archived');default:'published'"`
	RejectComment string    `json:"reject_comment" gorm:"type:text"`
	UserId        *uint     `json:"user_id"`
	User          *User     `json:"user,omitempty" gorm:"foreignKey:UserId;constraint:OnDelete:SET NULL"`
	Tags          []Tag     `json:"tags" gorm:"many2many:blog_tags;"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
