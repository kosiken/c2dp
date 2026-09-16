package models

import "time"

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	User      User      `json:"user" gorm:"constraint:OnDelete:CASCADE"`
	Caption   string    `json:"caption" gorm:"size:2200"`
	ImagePath string    `json:"-" gorm:"not null"`
	ImageURL  string    `json:"image_url" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
