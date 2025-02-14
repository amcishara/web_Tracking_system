package routes

import (
	"net/http"
	"strconv"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

// addToCart handles POST /cart request
func addToCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := models.AddToCart(db.DB, userID, input.ProductID, input.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart"})
}

// removeFromCart handles DELETE /cart/:id request
func removeFromCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := models.RemoveFromCart(db.DB, userID, uint(itemID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// getCart handles GET /cart request
func getCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	summary, err := models.GetCart(db.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// updateQuantity handles PATCH /cart/:id/quantity
func updateQuantity(c *gin.Context) {
	userID := c.GetUint("user_id")
	itemID := c.Param("id")

	var input struct {
		Increment int `json:"increment" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update quantity - add user check
	var cartItem models.CartItem
	if err := db.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&cartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	newQuantity := cartItem.Quantity + input.Increment
	if newQuantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be at least 1"})
		return
	}

	cartItem.Quantity = newQuantity
	if err := db.DB.Save(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quantity updated"})
}
