package main

import (
	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db.InitDB()

	server := gin.Default()
	routes.SetupRouter(server)

	server.Run(":8000")
}

