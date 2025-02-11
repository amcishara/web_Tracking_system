package routes

import (
	"github.com/amcishara/web_Tracking_system/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	// Public routes
	router.POST("/signup", signup)
	router.POST("/login", login)
	router.POST("/logout", logout)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/user/:id", updateUser)
		protected.DELETE("/user/:id", deleteUser)
		protected.GET("/my/view-history", getUserViewHistory)
	}

	// Customer routes (with cart functionality)
	customer := router.Group("/")
	customer.Use(middleware.AuthMiddleware())
	customer.Use(middleware.CustomerMiddleware())
	{
		customer.POST("/cart", addToCart)
		customer.DELETE("/cart/:id", removeFromCart)
		customer.GET("/cart", getCart)
	}

	// Public product routes
	router.GET("/products", getProducts)
	router.GET("/products/search", searchProducts)
	router.GET("/products/:id", getProductPublic)

	// Protected product routes - change path to make it clearer
	auth := router.Group("/api/auth")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/products/:id", getProductAuth)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", getUsers)
		admin.PUT("/users/:id", manageUser)
		admin.GET("/analytics", getAnalytics)
		admin.POST("/products", createProduct)
		admin.POST("/products/bulk", createBulkProducts)
		admin.PUT("/products/:id", updateProduct)
		admin.DELETE("/products/:id", deleteProduct)
		admin.DELETE("/users/:id", deleteUserAdmin)
	}

	// Guest routes
	router.GET("/guest/view-history", getGuestViewHistory)

	// Inside SetupRouter function, add with public routes
	router.GET("/trending", getTrendingItems)
}
