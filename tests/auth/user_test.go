package auth_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/tests/utils"
)

func TestMain(m *testing.M) {
	// Setup
	utils.SetupTestDB()

	// Run tests
	code := m.Run()

	// Cleanup
	utils.PrintReport()
	utils.CleanupTestDB()

	os.Exit(code)
}

func TestUserSignup(t *testing.T) {
	utils.TruncateTable("users")

	t.Run("Valid Signup", func(t *testing.T) {
		user := &models.User{
			Email:    "test@example.com",
			Password: "SecureP@ss123",
		}

		err := models.CreateUser(utils.TestDB, user)
		passed := err == nil && user.UserID != 0
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("Failed to create user: %v", err)
		} else if user.UserID == 0 {
			errMsg = "Expected user ID to be set after creation"
		}
		utils.RecordTest(t, "User Signup - Valid", passed, errMsg)
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		user := &models.User{
			Email:    "test@example.com",
			Password: "SecureP@ss123",
		}

		err := models.CreateUser(utils.TestDB, user)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for duplicate email, got nil"
		}
		utils.RecordTest(t, "User Signup - Duplicate Email", passed, errMsg)
	})

	t.Run("Invalid Email", func(t *testing.T) {
		user := &models.User{
			Email:    "invalid-email",
			Password: "SecureP@ss123",
		}

		err := models.CreateUser(utils.TestDB, user)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for invalid email format, got nil"
		}
		utils.RecordTest(t, "User Signup - Invalid Email", passed, errMsg)
	})
}

func TestUserLogin(t *testing.T) {
	utils.TruncateTable("users")

	// Setup: Create a test user with valid password
	testUser := &models.User{
		Email:    "login@example.com",
		Password: "SecureP@ss123",
	}
	models.CreateUser(utils.TestDB, testUser)

	t.Run("Valid Login", func(t *testing.T) {
		user := &models.User{
			Email:    "login@example.com",
			Password: "SecureP@ss123",
		}

		userID, err := models.ValidateUser(utils.TestDB, user)
		passed := err == nil && userID != 0
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("Failed to validate user: %v", err)
		} else if userID == 0 {
			errMsg = "Expected valid user ID, got 0"
		}
		utils.RecordTest(t, "User Login - Valid", passed, errMsg)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		user := &models.User{
			Email:    "login@example.com",
			Password: "wrongpassword",
		}

		_, err := models.ValidateUser(utils.TestDB, user)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for wrong password, got nil"
		}
		utils.RecordTest(t, "User Login - Wrong Password", passed, errMsg)
	})

	t.Run("Non-existent User", func(t *testing.T) {
		user := &models.User{
			Email:    "nonexistent@example.com",
			Password: "SecureP@ss123",
		}

		_, err := models.ValidateUser(utils.TestDB, user)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for non-existent user, got nil"
		}
		utils.RecordTest(t, "User Login - Non-existent User", passed, errMsg)
	})
}

func TestUserUpdate(t *testing.T) {
	utils.TruncateTable("users")

	testUser := &models.User{
		Email:    "update@example.com",
		Password: "SecureP@ss123",
	}
	models.CreateUser(utils.TestDB, testUser)

	t.Run("Valid Update", func(t *testing.T) {
		testUser.Email = "updated@example.com"
		err := models.UpdateUser(utils.TestDB, testUser)
		passed := err == nil
		errMsg := ""
		if !passed {
			errMsg = fmt.Sprintf("Failed to update user: %v", err)
		}
		utils.RecordTest(t, "User Update - Valid", passed, errMsg)
	})

	t.Run("Duplicate Email Update", func(t *testing.T) {
		// Create another user first
		otherUser := &models.User{
			Email:    "other@example.com",
			Password: "SecureP@ss123",
		}
		models.CreateUser(utils.TestDB, otherUser)

		// Try to update to existing email
		otherUser.Email = "updated@example.com"
		err := models.UpdateUser(utils.TestDB, otherUser)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for duplicate email"
		}
		utils.RecordTest(t, "User Update - Duplicate Email", passed, errMsg)
	})

	t.Run("Password Update", func(t *testing.T) {
		testUser.Password = "NewP@ssword123"
		err := models.UpdateUser(utils.TestDB, testUser)

		// Verify new password works
		_, validateErr := models.ValidateUser(utils.TestDB, &models.User{
			Email:    testUser.Email,
			Password: "NewP@ssword123",
		})

		passed := err == nil && validateErr == nil
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("Failed to update password: %v", err)
		} else if validateErr != nil {
			errMsg = fmt.Sprintf("Failed to validate with new password: %v", validateErr)
		}

		utils.RecordTest(t, "User Update - Password", passed, errMsg)
	})
}

func TestUserDelete(t *testing.T) {
	utils.TruncateTable("users")

	testUser := &models.User{
		Email:    "delete@example.com",
		Password: "SecureP@ss123",
	}
	models.CreateUser(utils.TestDB, testUser)

	t.Run("Valid Delete", func(t *testing.T) {
		err := models.DeleteUser(utils.TestDB, int(testUser.UserID))
		passed := err == nil
		errMsg := ""
		if !passed {
			errMsg = fmt.Sprintf("Failed to delete user: %v", err)
		}
		utils.RecordTest(t, "User Delete - Valid", passed, errMsg)

		// Verify user is deleted
		_, err = models.GetUserByID(utils.TestDB, int(testUser.UserID))
		if err == nil {
			t.Error("Expected error when getting deleted user")
		}
	})

	t.Run("Delete Non-existent User", func(t *testing.T) {
		err := models.DeleteUser(utils.TestDB, 99999)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error when deleting non-existent user"
		}
		utils.RecordTest(t, "User Delete - Non-existent", passed, errMsg)
	})
}

func TestUserRole(t *testing.T) {
	utils.TruncateTable("users")

	adminUser := &models.User{
		Email:    "admin@example.com",
		Password: "AdminP@ss123",
		Role:     "admin",
	}
	models.CreateUser(utils.TestDB, adminUser)

	regularUser := &models.User{
		Email:    "user@example.com",
		Password: "UserP@ss123",
		Role:     "user",
	}
	models.CreateUser(utils.TestDB, regularUser)

	t.Run("Admin Check", func(t *testing.T) {
		isAdmin := models.IsAdmin(utils.TestDB, adminUser.UserID)
		passed := isAdmin
		errMsg := ""
		if !passed {
			errMsg = "Expected user to be admin"
		}
		utils.RecordTest(t, "User Role - Admin Check", passed, errMsg)
	})

	t.Run("Regular User Check", func(t *testing.T) {
		isAdmin := models.IsAdmin(utils.TestDB, regularUser.UserID)
		passed := !isAdmin
		errMsg := ""
		if !passed {
			errMsg = "Expected user to not be admin"
		}
		utils.RecordTest(t, "User Role - Regular User Check", passed, errMsg)
	})

	t.Run("Non-existent User Check", func(t *testing.T) {
		isAdmin := models.IsAdmin(utils.TestDB, 99999)
		passed := !isAdmin
		errMsg := ""
		if !passed {
			errMsg = "Expected non-existent user to not be admin"
		}
		utils.RecordTest(t, "User Role - Non-existent User", passed, errMsg)
	})
}

func TestPasswordValidation(t *testing.T) {
	utils.TruncateTable("users")

	testCases := []struct {
		name     string
		password string
		wantErr  string
	}{
		{
			name:     "Valid Password",
			password: "SecureP@ss123",
			wantErr:  "",
		},
		{
			name:     "Empty Password",
			password: "",
			wantErr:  "password cannot be empty",
		},
		{
			name:     "Too Short",
			password: "Sh@rt1",
			wantErr:  "password must be at least 8 characters",
		},
		{
			name:     "Too Long",
			password: strings.Repeat("A", 73) + "@1a",
			wantErr:  "password must not exceed 72 characters",
		},
		{
			name:     "No Numbers",
			password: "SecurePass@",
			wantErr:  "password must contain at least one number",
		},
		{
			name:     "No Uppercase",
			password: "securepass@123",
			wantErr:  "password must contain at least one uppercase letter",
		},
		{
			name:     "No Lowercase",
			password: "SECUREPASS@123",
			wantErr:  "password must contain at least one lowercase letter",
		},
		{
			name:     "No Special Characters",
			password: "SecurePass123",
			wantErr:  "password must contain at least one special character",
		},
		{
			name:     "Common Password",
			password: "password123",
			wantErr:  "password is too common",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &models.User{
				Email:    "test@example.com",
				Password: tc.password,
			}

			err := models.CreateUser(utils.TestDB, user)
			passed := (err == nil && tc.wantErr == "") ||
				(err != nil && tc.wantErr != "" && strings.Contains(err.Error(), tc.wantErr))

			errMsg := ""
			if !passed {
				if err == nil {
					errMsg = fmt.Sprintf("expected error containing '%s', got no error", tc.wantErr)
				} else {
					errMsg = fmt.Sprintf("expected error containing '%s', got '%s'", tc.wantErr, err.Error())
				}
			}

			utils.RecordTest(t, "Password Validation - "+tc.name, passed, errMsg)
		})
	}
}

func TestPasswordUpdateValidation(t *testing.T) {
	utils.TruncateTable("users")

	// Create initial user with valid password
	user := &models.User{
		Email:    "update@example.com",
		Password: "InitialP@ss123",
	}
	if err := models.CreateUser(utils.TestDB, user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	testCases := []struct {
		name     string
		password string
		wantErr  string
	}{
		{
			name:     "Valid Update",
			password: "NewSecureP@ss123",
			wantErr:  "",
		},
		{
			name:     "Too Short Update",
			password: "Sh@rt1",
			wantErr:  "password must be at least 8 characters",
		},
		{
			name:     "No Special Char Update",
			password: "NewPassword123",
			wantErr:  "password must contain at least one special character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateUser := &models.User{
				UserID:   user.UserID,
				Email:    user.Email,
				Password: tc.password,
			}

			err := models.UpdateUser(utils.TestDB, updateUser)
			passed := (err == nil && tc.wantErr == "") ||
				(err != nil && tc.wantErr != "" && strings.Contains(err.Error(), tc.wantErr))

			errMsg := ""
			if !passed {
				if err == nil {
					errMsg = fmt.Sprintf("expected error containing '%s', got no error", tc.wantErr)
				} else {
					errMsg = fmt.Sprintf("expected error containing '%s', got '%s'", tc.wantErr, err.Error())
				}
			}

			utils.RecordTest(t, "Password Update - "+tc.name, passed, errMsg)

			// If update was successful, verify we can login with new password
			if err == nil {
				loginUser := &models.User{
					Email:    user.Email,
					Password: tc.password,
				}
				_, loginErr := models.ValidateUser(utils.TestDB, loginUser)
				passed = loginErr == nil
				errMsg = ""
				if !passed {
					errMsg = fmt.Sprintf("Failed to login with updated password: %v", loginErr)
				}
				utils.RecordTest(t, "Password Update Login - "+tc.name, passed, errMsg)
			}
		})
	}
}

func TestEmailValidationEdgeCases(t *testing.T) {
	utils.TruncateTable("users")

	testCases := []struct {
		name  string
		email string
	}{
		{"Empty Email", ""},
		{"No Domain", "test@"},
		{"No Username", "@domain.com"},
		{"Special Chars", "test!@#$@domain.com"},
		{"Multiple @", "test@test@domain.com"},
		{"No TLD", "test@domain"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &models.User{
				Email:    tc.email,
				Password: "SecureP@ss123",
			}
			err := models.CreateUser(utils.TestDB, user)
			passed := err != nil
			errMsg := ""
			if !passed {
				errMsg = fmt.Sprintf("Expected error for invalid email: %s", tc.email)
			}
			utils.RecordTest(t, "Email Validation - "+tc.name, passed, errMsg)
		})
	}
}

func TestRoleChange(t *testing.T) {
	utils.TruncateTable("users")

	user := &models.User{
		Email:    "role@example.com",
		Password: "RoleP@ss123",
		Role:     "user",
	}
	models.CreateUser(utils.TestDB, user)

	t.Run("User to Admin", func(t *testing.T) {
		user.Role = "admin"
		err := models.UpdateUser(utils.TestDB, user)
		passed := err == nil && models.IsAdmin(utils.TestDB, user.UserID)
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("Failed to update role: %v", err)
		} else if !models.IsAdmin(utils.TestDB, user.UserID) {
			errMsg = "User not marked as admin after role update"
		}
		utils.RecordTest(t, "Role Change - User to Admin", passed, errMsg)
	})

	t.Run("Invalid Role", func(t *testing.T) {
		user.Role = "invalid_role"
		err := models.UpdateUser(utils.TestDB, user)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for invalid role"
		}
		utils.RecordTest(t, "Role Change - Invalid Role", passed, errMsg)
	})
}
