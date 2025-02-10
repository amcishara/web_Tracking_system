package models

import (
	"fmt"
	"time"

	"github.com/amcishara/web_Tracking_system/utils"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// User-related functions
func CreateUser(db *gorm.DB, u *User) error {
	var count int64
	db.Model(&User{}).Where("email = ?", u.Email).Count(&count)
	if count > 0 {
		return fmt.Errorf("user with email '%s' already exists", u.Email)
	}

	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	result := db.Create(u)
	return result.Error
}

func ValidateUser(db *gorm.DB, u *User) (uint, error) {
	var existingUser User
	result := db.Where("email = ?", u.Email).First(&existingUser)
	if result.Error != nil {
		return 0, fmt.Errorf("invalid credentials")
	}

	if err := utils.ComparePasswords(existingUser.Password, u.Password); err != nil {
		return 0, fmt.Errorf("invalid credentials")
	}

	return existingUser.ID, nil
}

func GetUserByID(db *gorm.DB, id int) (*User, error) {
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func UpdateUser(db *gorm.DB, u *User) error {
	// Check if email is being changed and if it already exists
	var count int64
	db.Model(&User{}).Where("email = ? AND id != ?", u.Email, u.ID).Count(&count)
	if count > 0 {
		return fmt.Errorf("email '%s' already taken", u.Email)
	}

	// If password is being updated, hash it
	if u.Password != "" {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword
	}

	result := db.Save(u)
	return result.Error
}

func DeleteUser(db *gorm.DB, id int) error {
	result := db.Delete(&User{}, id)
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return result.Error
}

func IsAdmin(db *gorm.DB, userID uint) bool {
	var user User
	result := db.First(&user, userID)
	if result.Error != nil {
		return false
	}
	return user.Role == "admin"
}
