package routes

import (
	"net/http"
	"strconv"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func getUsers(c *gin.Context) {
	var users []models.User
	db.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func getAnalytics(c *gin.Context) {
	var interactions []models.UserInteraction
	db.DB.Preload("User").Preload("Product").Find(&interactions)

	c.JSON(http.StatusOK, gin.H{
		"total_interactions": len(interactions),
		"interactions":       interactions,
	})
}

func manageUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func deleteUserAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Begin transaction
	tx := db.DB.Begin()

	// Delete user's sessions
	if err := tx.Where("user_id = ?", id).Delete(&models.Session{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user sessions"})
		return
	}

	// Delete user's interactions
	if err := tx.Where("user_id = ?", id).Delete(&models.UserInteraction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user interactions"})
		return
	}

	// Delete the user
	if err := tx.Delete(&models.User{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User and all related data deleted successfully",
	})
}
