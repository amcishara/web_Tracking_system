package models

import (
	"time"
)

type UserInteraction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Type      string    `gorm:"not null" json:"type"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID"`
	Product   Product   `gorm:"foreignKey:ProductID"`
}
