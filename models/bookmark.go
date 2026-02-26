package models

import "time"

type Bookmark struct {
	Id          uint            `json:"id" gorm:"primaryKey"`
	Url         string          `json:"url" gorm:"unique;not null"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Topics      []BookmarkTopic `json:"topics" gorm:"foreignKey:BookmarkId;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type BookmarkTopic struct {
	Id         uint   `json:"id" gorm:"primaryKey"`
	BookmarkId uint   `json:"bookmark_id"`
	Name       string `json:"name"`
}
