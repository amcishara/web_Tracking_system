package middleware

import (
	"strings"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If no token in header, try cookie
		if tokenString == "" {
			var err error
			tokenString, err = c.Cookie("token")
			if err != nil {
				c.JSON(401, gin.H{"error": "Authorization required"})
				c.Abort()
				return
			}
		}

		// Check if token exists in sessions table
		session, err := models.GetSession(db.DB, tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("user_id", session.UserID)
		c.Set("email", (*claims)["email"])

		c.Next()
	}
}
