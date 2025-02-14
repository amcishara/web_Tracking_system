package middleware

import (
	"net/http"
	"strings"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie or Authorization header
		token, _ := c.Cookie("token")
		if token == "" {
			// Check Authorization header
			authHeader := c.GetHeader("Authorization")
			if len(strings.Split(authHeader, " ")) == 2 {
				token = strings.Split(authHeader, " ")[1]
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Validate session
		session, err := models.GetSession(db.DB, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", session.UserID)
		c.Next()
	}
}
