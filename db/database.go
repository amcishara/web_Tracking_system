package db

import (
	"fmt"
	"log"

	"github.com/amcishara/web_Tracking_system/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/web_db?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database connection successful")

	err = DB.AutoMigrate(&models.Product{}, &models.User{}, &models.UserInteraction{}, &models.Session{}, &models.CartItem{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
