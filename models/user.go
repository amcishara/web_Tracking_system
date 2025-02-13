package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/amcishara/web_Tracking_system/utils"
	"gorm.io/gorm"
)

// Add email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Add constants for validation
const (
	MinPasswordLength = 8
	MaxPasswordLength = 72 // bcrypt max length
)

// Password validation rules
var (
	hasNumber = regexp.MustCompile(`\d`)
	hasUpper  = regexp.MustCompile(`[A-Z]`)
	hasLower  = regexp.MustCompile(`[a-z]`)
	hasSymbol = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
)

// Add valid roles constant
var ValidRoles = map[string]bool{
	"user":  true,
	"admin": true,
}

// Add validation function
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// Enhanced password validation function
func isValidPassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Check for common passwords first
	commonPasswords := map[string]bool{
		"password123": true,
		"12345678":    true,
		"qwerty123":   true,
		// Add more common passwords as needed
	}
	if commonPasswords[strings.ToLower(password)] { // Convert to lowercase for comparison
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)
	}

	if !hasNumber.MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasUpper.MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower.MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasSymbol.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// Add role validation function
func isValidRole(role string) bool {
	return ValidRoles[role]
}

type User struct {
	UserID    uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	Email     string    `gorm:"unique;not null" json:"email" binding:"required"`
	Password  string    `gorm:"not null" json:"password" binding:"required"`
	Role      string    `gorm:"default:user" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// User-related functions
func CreateUser(db *gorm.DB, u *User) error {
	// Validate email format
	if !isValidEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if err := isValidPassword(u.Password); err != nil {
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
	// Validate role if it's being updated
	if u.Role != "" && !isValidRole(u.Role) {
		return fmt.Errorf("invalid role: %s", u.Role)
	}

	// Check if email is being changed and if it already exists
	var count int64
	db.Model(&User{}).Where("email = ? AND user_id != ?", u.Email, u.UserID).Count(&count)
	if count > 0 {
		return fmt.Errorf("email '%s' already taken", u.Email)
	}

	// If password is being updated, hash it
	if u.Password != "" {
		if err := isValidPassword(u.Password); err != nil {
			return err
		}
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
