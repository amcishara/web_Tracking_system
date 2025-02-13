package models_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	// Get database configuration from environment variables
	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "root" // default
	}

	dbPass := os.Getenv("TEST_DB_PASSWORD")
	if dbPass == "" {
		dbPass = "" // empty password
	}

	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "test_db"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up existing tables
	db.Exec("DROP TABLE IF EXISTS users")

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Test cases
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid user",
			user: &models.User{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			wantErr: false,
		},
		{
			name: "Duplicate email",
			user: &models.User{
				Email:    "test@example.com",
				Password: "password456",
				Role:     "user",
			},
			wantErr: true,
			errMsg:  "user with email 'test@example.com' already exists",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.CreateUser(db, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.UserID)

				// Verify user was created
				var savedUser models.User
				err = db.First(&savedUser, tt.user.UserID).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.user.Email, savedUser.Email)
				assert.Equal(t, tt.user.Role, savedUser.Role)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Create a test user first
	testUser := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, testUser)
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name    string
		input   *models.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid credentials",
			input: &models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			input: &models.User{
				Email:    "wrong@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "invalid credentials",
		},
		{
			name: "Invalid password",
			input: &models.User{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantErr: true,
			errMsg:  "invalid credentials",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := models.ValidateUser(db, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Zero(t, userID)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, userID)
				assert.Equal(t, testUser.UserID, userID)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Create initial user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		update  *models.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid update",
			update: &models.User{
				UserID:   user.UserID,
				Email:    "updated@example.com",
				Password: "newpassword123",
				Role:     "user",
			},
			wantErr: false,
		},
		{
			name: "Non-existent user",
			update: &models.User{
				UserID:   999,
				Email:    "fake@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "Duplicate email",
			update: &models.User{
				UserID:   user.UserID,
				Email:    "duplicate@example.com", // Will create this email first
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email 'duplicate@example.com' already taken",
		},
	}

	// Create a user with duplicate email for testing
	duplicateUser := &models.User{
		Email:    "duplicate@example.com",
		Password: "password123",
		Role:     "user",
	}
	err = models.CreateUser(db, duplicateUser)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.UpdateUser(db, tt.update)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify update
				updated, err := models.GetUserByID(db, int(tt.update.UserID))
				assert.NoError(t, err)
				assert.Equal(t, tt.update.Email, updated.Email)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Create test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		userID  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid deletion",
			userID:  int(user.UserID),
			wantErr: false,
		},
		{
			name:    "Non-existent user",
			userID:  999,
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:    "Already deleted user",
			userID:  int(user.UserID),
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.DeleteUser(db, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify deletion
				_, err := models.GetUserByID(db, tt.userID)
				assert.Error(t, err)
				assert.Equal(t, "user not found", err.Error())
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Create test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		userID  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Existing user",
			userID:  int(user.UserID),
			wantErr: false,
		},
		{
			name:    "Non-existent user",
			userID:  999,
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := models.GetUserByID(db, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
				assert.Nil(t, foundUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, foundUser)
				assert.Equal(t, user.Email, foundUser.Email)
				assert.Equal(t, user.Role, foundUser.Role)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Create test users
	adminUser := &models.User{
		Email:    "admin@example.com",
		Password: "password123",
		Role:     "admin",
	}
	err := models.CreateUser(db, adminUser)
	assert.NoError(t, err)

	regularUser := &models.User{
		Email:    "user@example.com",
		Password: "password123",
		Role:     "user",
	}
	err = models.CreateUser(db, regularUser)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		userID  uint
		isAdmin bool
		wantErr bool
	}{
		{
			name:    "Admin user",
			userID:  adminUser.UserID,
			isAdmin: true,
		},
		{
			name:    "Regular user",
			userID:  regularUser.UserID,
			isAdmin: false,
		},
		{
			name:    "Non-existent user",
			userID:  999,
			isAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAdmin := models.IsAdmin(db, tt.userID)
			assert.Equal(t, tt.isAdmin, isAdmin)
		})
	}
}

func TestUserTableName(t *testing.T) {
	user := models.User{}
	assert.Equal(t, "users", user.TableName(), "Table name should be 'users'")
}

func TestCreateUserWithInvalidEmail(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "Invalid email format",
			email:   "notanemail",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "Valid email",
			email:   "test@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				Email:    tt.email,
				Password: "password123",
				Role:     "user",
			}

			err := models.CreateUser(db, user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, user.UserID)
			}
		})
	}
}

func TestCreateUserWithInvalidRole(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name    string
		role    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid admin role",
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "Valid user role",
			role:    "user",
			wantErr: false,
		},
		{
			name:    "Invalid role",
			role:    "superuser",
			wantErr: true,
			errMsg:  "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				Email:    fmt.Sprintf("test_%s@example.com", tt.role),
				Password: "password123",
				Role:     tt.role,
			}

			err := models.CreateUser(db, user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.role, user.Role)
			}
		})
	}
}

func TestPasswordHashing(t *testing.T) {
	db := setupTestDB(t)

	// Test that passwords are properly hashed
	plainPassword := "password123"
	user := &models.User{
		Email:    "test@example.com",
		Password: plainPassword,
		Role:     "user",
	}

	// Create user
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	// Retrieve user from database
	var savedUser models.User
	err = db.First(&savedUser, user.UserID).Error
	assert.NoError(t, err)

	// Verify password was hashed
	assert.NotEqual(t, plainPassword, savedUser.Password, "Password should be hashed")
	assert.True(t, len(savedUser.Password) > len(plainPassword), "Hashed password should be longer than plain password")
}

func TestUpdateUserPassword(t *testing.T) {
	db := setupTestDB(t)

	// Create initial user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	// Store original hashed password
	originalHash := user.Password

	// Update password
	user.Password = "newpassword123"
	err = models.UpdateUser(db, user)
	assert.NoError(t, err)

	// Verify password was updated and hashed
	var updatedUser models.User
	err = db.First(&updatedUser, user.UserID).Error
	assert.NoError(t, err)
	assert.NotEqual(t, originalHash, updatedUser.Password, "Password should be newly hashed")
}

func TestUpdateUserEmail(t *testing.T) {
	db := setupTestDB(t)

	// Create initial user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid email update",
			email:   "newemail@example.com",
			wantErr: false,
		},
		{
			name:    "Empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "Invalid email format",
			email:   "notanemail",
			wantErr: true,
			errMsg:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateUser := &models.User{
				UserID:   user.UserID,
				Email:    tt.email,
				Password: user.Password,
				Role:     user.Role,
			}

			err := models.UpdateUser(db, updateUser)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				// Verify update
				updated, err := models.GetUserByID(db, int(user.UserID))
				assert.NoError(t, err)
				assert.Equal(t, tt.email, updated.Email)
			}
		})
	}
}

func TestUpdateUserRole(t *testing.T) {
	db := setupTestDB(t)

	// Create initial user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := models.CreateUser(db, user)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		role    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Update to admin",
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "Invalid role",
			role:    "superuser",
			wantErr: true,
			errMsg:  "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateUser := &models.User{
				UserID:   user.UserID,
				Email:    user.Email,
				Password: user.Password,
				Role:     tt.role,
			}

			err := models.UpdateUser(db, updateUser)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				// Verify update
				updated, err := models.GetUserByID(db, int(user.UserID))
				assert.NoError(t, err)
				assert.Equal(t, tt.role, updated.Role)
			}
		})
	}
}

func TestCreateUserWithEmptyPassword(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email: "test@example.com",
		Role:  "user",
		// Password intentionally left empty
	}

	err := models.CreateUser(db, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestUserCreatedAtTimestamp(t *testing.T) {
	db := setupTestDB(t)

	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}

	before := time.Now()
	err := models.CreateUser(db, user)
	after := time.Now()

	assert.NoError(t, err)
	assert.True(t, user.CreatedAt.After(before) || user.CreatedAt.Equal(before))
	assert.True(t, user.CreatedAt.Before(after) || user.CreatedAt.Equal(after))
}
