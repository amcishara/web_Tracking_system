package auth_test

import (
	"fmt"
	"testing"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/tests/utils"
)

func TestSession(t *testing.T) {
	// Clean tables before test
	utils.TruncateTable("users")
	utils.TruncateTable("sessions")

	// Create test user
	user := &models.User{
		Email:    "session@example.com",
		Password: "password123",
	}
	models.CreateUser(utils.TestDB, user)

	t.Run("Create Session", func(t *testing.T) {
		token := "test-token"
		err := models.CreateSession(utils.TestDB, user.UserID, token)

		session, getErr := models.GetSession(utils.TestDB, token)
		passed := err == nil && getErr == nil && session.UserID == user.UserID
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("Failed to create session: %v", err)
		} else if getErr != nil {
			errMsg = fmt.Sprintf("Failed to get session: %v", getErr)
		} else if session.UserID != user.UserID {
			errMsg = fmt.Sprintf("Expected user ID %d, got %d", user.UserID, session.UserID)
		}
		utils.RecordTest(t, "Session Creation", passed, errMsg)
	})

	t.Run("Delete Session", func(t *testing.T) {
		token := "test-token-delete"
		models.CreateSession(utils.TestDB, user.UserID, token)

		err := models.DeleteSession(utils.TestDB, token)
		_, getErr := models.GetSession(utils.TestDB, token)
		passed := err == nil && getErr != nil
		errMsg := ""
		if err != nil {
			errMsg = fmt.Sprintf("Failed to delete session: %v", err)
		} else if getErr == nil {
			errMsg = "Expected error for deleted session, got nil"
		}
		utils.RecordTest(t, "Session Deletion", passed, errMsg)
	})
}

func TestSessionExpiration(t *testing.T) {
	utils.TruncateTable("users")
	utils.TruncateTable("sessions")

	// Create test user
	user := &models.User{
		Email:    "session@example.com",
		Password: "password123",
	}
	models.CreateUser(utils.TestDB, user)

	t.Run("Session Expiry", func(t *testing.T) {
		token := "test-token-expiry"
		if err := models.CreateSession(utils.TestDB, user.UserID, token); err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Simulate session expiration by directly updating the created_at timestamp
		utils.TestDB.Exec("UPDATE sessions SET created_at = DATE_SUB(NOW(), INTERVAL 25 HOUR) WHERE token = ?", token)

		session, getErr := models.GetSession(utils.TestDB, token)
		passed := getErr != nil || session == nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for expired session"
		}
		utils.RecordTest(t, "Session - Expiration", passed, errMsg)
	})
}

func TestConcurrentSessions(t *testing.T) {
	utils.TruncateTable("users")
	utils.TruncateTable("sessions")

	user := &models.User{
		Email:    "multi@example.com",
		Password: "password123",
	}
	models.CreateUser(utils.TestDB, user)

	t.Run("Multiple Active Sessions", func(t *testing.T) {
		// Create multiple sessions
		tokens := []string{"token1", "token2", "token3"}
		for _, token := range tokens {
			err := models.CreateSession(utils.TestDB, user.UserID, token)
			if err != nil {
				t.Errorf("Failed to create session with token %s: %v", token, err)
			}
		}

		// Verify all sessions are active
		var count int64
		utils.TestDB.Model(&models.Session{}).Where("user_id = ?", user.UserID).Count(&count)

		passed := count == int64(len(tokens))
		errMsg := ""
		if !passed {
			errMsg = fmt.Sprintf("Expected %d sessions, got %d", len(tokens), count)
		}
		utils.RecordTest(t, "Session - Multiple Active", passed, errMsg)
	})
}
