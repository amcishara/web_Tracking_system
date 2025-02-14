package db

import (
	"fmt"
	"log"
	"os"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using default values")
	}

	// Get environment variables with fallback values
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "test_db")
	dbPort := getEnv("DB_PORT", "3306")

	// Create DSN string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	DB = db // Set the global DB variable

	// Check if tables exist
	hasUsers := DB.Migrator().HasTable(&models.User{})
	hasProducts := DB.Migrator().HasTable(&models.Product{})
	hasInteractions := DB.Migrator().HasTable("user_interactions")
	hasGuestInteractions := DB.Migrator().HasTable("guest_interactions")
	hasTrending := DB.Migrator().HasTable("trending_products")

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
				guest_id VARCHAR(255) NOT NULL,
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

	if !hasTrending {
		err = DB.Exec(`
			CREATE TABLE trending_products (
				product_id BIGINT UNSIGNED PRIMARY KEY,
				title VARCHAR(255) NOT NULL,
				total_views INT NOT NULL DEFAULT 0,
				FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
			)
		`).Error
		if err != nil {
			log.Fatal("Failed to create trending_products table:", err)
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

	fmt.Println("Database connection and migration completed successfully")
	return DB, nil
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
