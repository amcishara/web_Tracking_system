package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/amcishara/web_Tracking_system/utils"
	"gorm.io/gorm"
)

type User struct {
	UserID    uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:user" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// Add this constant for valid roles
var validRoles = map[string]bool{
	"admin": true,
	"user":  true,
}

// Add role validation function
func validateRole(role string) error {
	if !validRoles[role] {
		return fmt.Errorf("invalid role")
	}
	return nil
}

// User-related functions
func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Simple email regex pattern
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailPattern, email)
	if !match {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// Add password validation function
func validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func CreateUser(db *gorm.DB, u *User) error {
	// Validate email
	if err := validateEmail(u.Email); err != nil {
		return err
	}

	// Validate password
	if err := validatePassword(u.Password); err != nil {
		return err
	}

	// Validate role
	if err := validateRole(u.Role); err != nil {
		return err
	}

	// Check for duplicate email
	var count int64
	db.Model(&User{}).Where("email = ?", u.Email).Count(&count)
	if count > 0 {
		return fmt.Errorf("user with email '%s' already exists", u.Email)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	// Create user
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

	return existingUser.UserID, nil
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
	// First check if user exists
	var existingUser User
	if err := db.First(&existingUser, u.UserID).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if email is being changed and if it already exists
	var count int64
	db.Model(&User{}).Where("email = ? AND user_id != ?", u.Email, u.UserID).Count(&count)
	if count > 0 {
		return fmt.Errorf("email '%s' already taken", u.Email)
	}

	// Validate email if it's being updated
	if err := validateEmail(u.Email); err != nil {
		return err
	}

	// Validate role if it's being updated
	if err := validateRole(u.Role); err != nil {
		return err
	}

	// If password is being updated, validate and hash it
	if u.Password != "" {
		if err := validatePassword(u.Password); err != nil {
			return err
		}
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword
	} else {
		// Keep existing password if not updating
		u.Password = existingUser.Password
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
