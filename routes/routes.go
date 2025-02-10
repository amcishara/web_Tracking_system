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
		protected.POST("/products", createProduct)
		protected.PUT("/products/:id", updateProduct)
		protected.DELETE("/products/:id", deleteProduct)
	}

	// Public product routes
	router.GET("/products", getProducts)
	router.GET("/products/:id", getProduct)

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", getUsers)
		admin.PUT("/users/:id", manageUser)
		admin.GET("/analytics", getAnalytics)
		admin.POST("/products", createProduct)
		admin.PUT("/products/:id", updateProduct)
		admin.DELETE("/products/:id", deleteProduct)
		admin.DELETE("/users/:id", deleteUserAdmin)
	}
}
