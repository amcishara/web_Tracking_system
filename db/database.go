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

	// Check if tables exist
	hasUsers := DB.Migrator().HasTable(&models.User{})
	hasProducts := DB.Migrator().HasTable(&models.Product{})
	hasInteractions := DB.Migrator().HasTable("user_interactions")
	hasGuestInteractions := DB.Migrator().HasTable("guest_interactions")
	hasTrendingItems := DB.Migrator().HasTable("trending_items")

	// Only create tables if they don't exist
	if !hasUsers {
		err = DB.Exec(`
			CREATE TABLE users (
				user_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
				email VARCHAR(255) NOT NULL UNIQUE,
				password VARCHAR(255) NOT NULL,
				role VARCHAR(50) DEFAULT 'user',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`).Error
		if err != nil {
			log.Fatal("Failed to create users table:", err)
		}
	}

	if !hasProducts {
		err = DB.AutoMigrate(&models.Product{})
		if err != nil {
			log.Fatal("Failed to create products table:", err)
		}
	}

	if !hasInteractions {
		err = DB.Exec(`
			CREATE TABLE user_interactions (
				user_id BIGINT UNSIGNED NOT NULL,
				product_id BIGINT UNSIGNED NOT NULL,
				viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (user_id, product_id, viewed_at),
				FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
			)
		`).Error
		if err != nil {
			log.Fatal("Failed to create user_interactions table:", err)
		}
	}

	if !hasGuestInteractions {
		err = DB.Exec(`
			CREATE TABLE guest_interactions (
				guest_id VARCHAR(36) NOT NULL,
				product_id BIGINT UNSIGNED NOT NULL,
				viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (guest_id, product_id, viewed_at),
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
			)
		`).Error
		if err != nil {
			log.Fatal("Failed to create guest_interactions table:", err)
		}
	}

	if !hasTrendingItems {
		err = DB.Exec(`
			CREATE TABLE trending_items (
				product_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
				product_title VARCHAR(255) NOT NULL,
				view_count BIGINT UNSIGNED DEFAULT 0,
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
			)
		`).Error
		if err != nil {
			log.Fatal("Failed to create trending_items table:", err)
		}
	}

	// Create other tables only if they don't exist
	err = DB.AutoMigrate(
		&models.Session{},
		&models.CartItem{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("Database migration completed successfully")
}
