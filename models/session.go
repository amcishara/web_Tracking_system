package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Session struct {
	UserID    uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	Token     string    `gorm:"primaryKey;unique" json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName overrides the table name
func (Session) TableName() string {
	return "sessions"
}

func CreateSession(db *gorm.DB, userID uint, token string) error {
	session := Session{
		UserID: userID,
		Token:  token,
	}
	return db.Create(&session).Error
}

func DeleteSession(db *gorm.DB, token string) error {
	result := db.Where("token = ?", token).Delete(&Session{})
	if result.Error != nil {
		return result.Error
	}

	// Verify deletion
	var count int64
	db.Model(&Session{}).Where("token = ?", token).Count(&count)
	if count > 0 {
		return fmt.Errorf("failed to delete session")
	}

	return nil
}

func GetSession(db *gorm.DB, token string) (*Session, error) {
	var session Session
	result := db.Preload("User").Where("token = ?", token).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}
