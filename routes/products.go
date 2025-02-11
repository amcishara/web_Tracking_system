package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func createProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product name already exists
	if models.IsProductNameExists(db.DB, product.Name) {
		c.JSON(http.StatusConflict, gin.H{"error": "Product with this name already exists"})
		return
	}

	if err := models.CreateProduct(db.DB, &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// Update the request type to be an array
type BulkProductRequest []struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

func createBulkProducts(c *gin.Context) {
	var request BulkProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(request) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No products provided"})
		return
	}

	// Begin transaction
	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	createdProducts := make([]models.Product, 0)
	var failedProducts []string

	for _, reqProduct := range request {
		// Convert request product to model
		product := models.Product{
			Name:        reqProduct.Name,
			Description: reqProduct.Description,
			Price:       reqProduct.Price,
			Category:    reqProduct.Category,
			Stock:       reqProduct.Stock,
		}

		// Check if product name already exists
		if models.IsProductNameExists(tx, product.Name) {
			failedProducts = append(failedProducts, fmt.Sprintf("Product '%s' already exists", product.Name))
			continue
		}

		// Create product
		if err := models.CreateProduct(tx, &product); err != nil {
			failedProducts = append(failedProducts, fmt.Sprintf("Failed to create '%s': %v", product.Name, err))
			continue
		}

		createdProducts = append(createdProducts, product)
	}

	// If there were any failures, rollback
	if len(failedProducts) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Some products failed to create",
			"details": failedProducts,
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  fmt.Sprintf("Successfully created %d products", len(createdProducts)),
		"products": createdProducts,
	})
}

func getProducts(c *gin.Context) {
	products := models.GetAllProducts(db.DB)
	c.JSON(http.StatusOK, products)
}

func getProductPublic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	product, err := models.GetProductByID(db.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check if user is authenticated
	if userID, exists := c.Get("user_id"); exists {
		fmt.Printf("Found user_id in context: %v\n", userID)
		// Track view for authenticated user
		if err := models.TrackUserView(db.DB, userID.(uint), product.ID); err != nil {
			fmt.Printf("Failed to track user view: %v\n", err)
		}
	} else {
		// Handle guest user
		guestID, err := c.Cookie("guest_id")
		if err != nil || guestID == "" {
			// Generate new guest ID if not exists
			guestID = models.GenerateGuestID()
			c.SetCookie("guest_id", guestID, 86400*30, "/", "", false, true) // 30 days expiry
		}

		// Track guest view
		if err := models.TrackGuestView(db.DB, guestID, product.ID); err != nil {
			fmt.Printf("Failed to track guest view: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, product)
}

func getProductAuth(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	product, err := models.GetProductByID(db.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func updateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.ID = uint(id)
	if err := models.UpdateProduct(db.DB, &product); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := models.DeleteProduct(db.DB, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func searchProducts(c *gin.Context) {
	// Get query parameters
	query := c.Query("q")           // Search query
	category := c.Query("category") // Category filter
	sortBy := c.Query("sort")       // Sort field (price/name/date)
	order := c.Query("order")       // Sort order (asc/desc)

	// Validate sort parameters
	if sortBy != "" && sortBy != "price" && sortBy != "name" && sortBy != "date" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sort field. Use 'price', 'name', or 'date'",
		})
		return
	}

	if order != "" && order != "asc" && order != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sort order. Use 'asc' or 'desc'",
		})
		return
	}

	products, err := models.SearchProducts(db.DB, query, category, sortBy, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"filters": gin.H{
			"query":    query,
			"category": category,
			"sort":     sortBy,
			"order":    order,
		},
	})
}

func getUserViewHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	products, err := models.GetUserViewHistory(db.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get view history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": products,
	})
}

// Add new handler for guest view history
func getGuestViewHistory(c *gin.Context) {
	guestID, err := c.Cookie("guest_id")
	if err != nil || guestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No guest ID found"})
		return
	}

	products, err := models.GetGuestViewHistory(db.DB, guestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get view history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": products,
	})
}

func getTrendingItems(c *gin.Context) {
	// Get limit from query parameter, default to 10
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	items, err := models.GetTrendingItems(db.DB, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trending": items,
	})
}
