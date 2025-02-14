package routes

import (
	"net/http"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/middleware"
	"github.com/amcishara/web_Tracking_system/models"
	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	// Public routes
	router.POST("/signup", signup)
	router.POST("/login", login)
	router.POST("/logout", logout)

	// Guest product routes
	router.GET("/products", getProducts)
	router.GET("/products/search", searchProducts)
	router.GET("/guest/products/:id", getProductAsGuest) // Guest product view
	router.GET("/guest/view-history", getGuestViewHistory)
	router.GET("/trending", getTrendingProducts)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		protected.PUT("/user/:id", updateUser)
		protected.DELETE("/user/:id", deleteUser)
		protected.GET("/my/view-history", getUserViewHistory)
		protected.GET("/products/:id", getProductAsUser) // Authenticated user product view
	}

	// Customer routes (with cart functionality)
	customer := router.Group("/")
	customer.Use(middleware.AuthRequired())
	customer.Use(middleware.CustomerMiddleware())
	{
		customer.POST("/cart", addToCart)
		customer.DELETE("/cart/:id", removeFromCart)
		customer.GET("/cart", getCart)
		customer.PATCH("/cart/:id/quantity", updateQuantity)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", getUsers)
		admin.PUT("/users/:id", manageUser)
		admin.GET("/analytics", getAnalytics)
		admin.POST("/products", createProduct)
		admin.POST("/products/bulk", createBulkProducts)
		admin.PUT("/products/:id", updateProduct)
		admin.PUT("/update-products/:id", adminUpdateProduct)
		admin.DELETE("/products/:id", deleteProduct)
		admin.DELETE("/delete-products/:id", adminDeleteProduct)
		admin.DELETE("/users/:id", deleteUserAdmin)
	}
}

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
