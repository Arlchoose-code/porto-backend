package models

import "time"

type Project struct {
	Id          uint               `json:"id" gorm:"primaryKey"`
	Title       string             `json:"title" gorm:"not null"`
	Slug        string             `json:"slug" gorm:"unique;not null"`
	Description string             `json:"description" gorm:"type:text"`
	Platform    string             `json:"platform"`
	Url         string             `json:"url"`
	TechStacks  []ProjectTechStack `json:"tech_stacks" gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE"`
	Images      []ProjectImage     `json:"images" gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type ProjectTechStack struct {
	Id        uint   `json:"id" gorm:"primaryKey"`
	ProjectId uint   `json:"project_id"`
	Name      string `json:"name" gorm:"not null"`
}

type ProjectImage struct {
	Id        uint   `json:"id" gorm:"primaryKey"`
	ProjectId uint   `json:"project_id"`
	ImageUrl  string `json:"image_url" gorm:"not null"`
	Order     int    `json:"order" gorm:"default:0"`
}
