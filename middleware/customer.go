package middleware

import (
	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func CustomerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		token, err := c.Cookie("token")
		if err != nil || token == "" {
			c.JSON(401, gin.H{
				"error":       "Please login to add items to cart",
				"redirect_to": "/login",
			})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error":       "Please login to add items to cart",
				"redirect_to": "/login",
			})
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

		// Check if user is NOT an admin
		if models.IsAdmin(db.DB, uid) {
			c.JSON(403, gin.H{"error": "Cart functionality is for customers only"})
			c.Abort()
			return
		}

		// Check if user exists in database
		if _, err := models.GetUserByID(db.DB, int(uid)); err != nil {
			c.JSON(401, gin.H{
				"error":       "Please login to add items to cart",
				"redirect_to": "/login",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
