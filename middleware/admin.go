package middleware

import (
	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Convert userID to uint correctly
		var uid uint
		switch v := userID.(type) {
		case uint:
			uid = v
		case float64:
			uid = uint(v)
		default:
			c.JSON(500, gin.H{"error": "Invalid user ID type"})
			c.Abort()
			return
		}

		// Check if user is admin
		if !models.IsAdmin(db.DB, uid) {
			c.JSON(403, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
