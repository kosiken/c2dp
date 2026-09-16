package models

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:40;uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Posts        []Post    `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
