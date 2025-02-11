package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/utils"
	"github.com/gin-gonic/gin"
)

func signup(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.CreateUser(db.DB, &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from ValidateUser
	userID, err := models.ValidateUser(db.DB, &user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	fmt.Printf("Login - User authenticated with ID: %d\n", userID)

	// Generate token
	token, err := utils.GenerateToken(userID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Create session
	if err := models.CreateSession(db.DB, userID, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.SetCookie(
		"token",     // name
		token,       // value
		3600*24,     // max age in seconds (24 hours)
		"/",         // path
		"localhost", // domain
		false,       // secure
		true,        // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.UserID = uint(id)
	if err := models.UpdateUser(db.DB, &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := models.DeleteUser(db.DB, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func logout(c *gin.Context) {
	// Get token from either cookie or header
	token, _ := c.Cookie("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if len(strings.Split(authHeader, " ")) == 2 {
			token = strings.Split(authHeader, " ")[1]
		}
	}

	// Try to delete session if token exists
	if token != "" {
		if err := models.DeleteSession(db.DB, token); err != nil {
			fmt.Printf("Failed to delete session: %v\n", err) // Add logging
		}
	}

	// Clear the cookie
	c.SetCookie(
		"token",     // name
		"",          // value
		-1,          // max age
		"/",         // path
		"localhost", // domain
		false,       // secure
		true,        // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
