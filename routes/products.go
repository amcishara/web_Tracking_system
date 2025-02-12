package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func getProductAsGuest(c *gin.Context) {
	// First check if user is logged in by looking for a valid token
	token := c.GetHeader("Authorization")
	if token == "" {
		token, _ = c.Cookie("token")
	}

	// If token exists, check if it's valid
	if token != "" {
		token = strings.TrimPrefix(token, "Bearer ")
		var session models.Session
		if err := db.DB.Where("token = ?", token).First(&session).Error; err == nil {
			// Valid session found - user is logged in
			c.JSON(http.StatusForbidden, gin.H{
				"error": "This route is only for guest users. Please use /products/:id for authenticated users",
			})
			return
		}
	}

	// Continue with guest view logic
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

	// Handle guest view tracking
	guestID, err := c.Cookie("guest_id")
	if err != nil || guestID == "" {
		guestID = generateGuestID()
		c.SetCookie("guest_id", guestID, 86400*30, "/", "", false, true)
	}

	// Track in guest_interactions table
	if err := models.TrackGuestView(db.DB, guestID, product.ID); err != nil {
		fmt.Printf("Failed to track guest view: %v\n", err)
	}

	c.JSON(http.StatusOK, product)
}

func getProductAsUser(c *gin.Context) {
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

	// Track in user_interactions table
	userID, _ := c.Get("user_id")
	if err := models.TrackUserView(db.DB, userID.(uint), product.ID); err != nil {
		fmt.Printf("Failed to track user view: %v\n", err)
	}

	// Get recommendations
	collaborative, err := models.GetCollaborativeRecommendations(db.DB, product.ID, 5)
	if err != nil {
		fmt.Printf("Failed to get collaborative recommendations: %v\n", err)
	}

	categoryBased, err := models.GetCategoryRecommendations(db.DB, product.ID, 5)
	if err != nil {
		fmt.Printf("Failed to get category recommendations: %v\n", err)
	}

	response := models.ProductWithRecommendations{
		ProductResponse:              product,
		CollaborativeRecommendations: collaborative,
		CategoryRecommendations:      categoryBased,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to generate guest ID
func generateGuestID() string {
	return fmt.Sprintf("guest_%s", uuid.New().String())
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

func getTrendingProducts(c *gin.Context) {
	trending, err := models.GetTrendingProducts(db.DB, 10) // Get top 10 trending products
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending products"})
		return
	}

	// Get full product details for each trending item
	var trendingWithDetails []gin.H
	for _, item := range trending {
		product, err := models.GetProductByID(db.DB, int(item.ProductID))
		if err != nil {
			continue // Skip if product not found
		}

		trendingWithDetails = append(trendingWithDetails, gin.H{
			"product":     product,
			"total_views": item.TotalViews,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"trending": trendingWithDetails,
	})
}

func adminDeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	// Start a transaction
	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Delete related records first
	// 1. Delete from trending_products
	if err := tx.Exec("DELETE FROM trending_products WHERE product_id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trending data"})
		return
	}

	// 2. Delete from user_interactions
	if err := tx.Exec("DELETE FROM user_interactions WHERE product_id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user interactions"})
		return
	}

	// 3. Delete from guest_interactions
	if err := tx.Exec("DELETE FROM guest_interactions WHERE product_id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete guest interactions"})
		return
	}

	// 4. Delete from cart_items if exists
	if err := tx.Exec("DELETE FROM cart_items WHERE product_id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart items"})
		return
	}

	// Finally delete the product
	if err := tx.Delete(&models.Product{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Product %d and all related data deleted successfully", id),
	})
}

// AdminProductUpdate represents the request body for updating a product
type AdminProductUpdate struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

func adminUpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	// Get existing product
	var existingProduct models.Product
	if err := db.DB.First(&existingProduct, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Bind update request
	var update AdminProductUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Check if new name conflicts with existing products (excluding current product)
	if update.Name != existingProduct.Name {
		var count int64
		if err := tx.Model(&models.Product{}).Where("name = ? AND id != ?", update.Name, id).Count(&count).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check product name"})
			return
		}
		if count > 0 {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{"error": "Product name already exists"})
			return
		}
	}

	// Update product
	updates := map[string]interface{}{
		"name":        update.Name,
		"description": update.Description,
		"price":       update.Price,
		"category":    update.Category,
		"stock":       update.Stock,
	}

	if err := tx.Model(&existingProduct).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Update trending products table if name changed
	if update.Name != existingProduct.Name {
		if err := tx.Exec("UPDATE trending_products SET title = ? WHERE product_id = ?", update.Name, id).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trending data"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
		return
	}

	// Get updated product for response
	var updatedProduct models.Product
	if err := db.DB.First(&updatedProduct, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": updatedProduct,
	})
}
