package main

import (
	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to Release mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize database
	db.InitDB()

	// Create router with default middleware
	router := gin.Default()

	// Set trusted proxies
	router.SetTrustedProxies([]string{"127.0.0.1"})

	// Setup routes
	routes.SetupRouter(router)

	// Start server
	router.Run(":8000")
}
