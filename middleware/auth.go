package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := c.GetHeader("Authorization")
		if token == "" {
			// Try to get from cookie if not in header
			token, _ = c.Cookie("token")
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
				c.Abort()
				return
			}
		}

		// Remove Bearer prefix if present
		token = strings.TrimPrefix(token, "Bearer ")

		// Get session
		var session models.Session
		if err := db.DB.Where("token = ?", token).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", session.UserID)
		fmt.Printf("Setting user_id in context: %d\n", session.UserID) // Add debug log

		c.Next()
	}
}
