package models

import "time"

type ToolUsage struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	ToolId    uint      `json:"tool_id" gorm:"not null;index"`
	ToolSlug  string    `json:"tool_slug" gorm:"not null;index"`
	IP        string    `json:"ip" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
