package utils

import (
	"fmt"
	"log"

	"github.com/amcishara/web_Tracking_system/db"
	"github.com/amcishara/web_Tracking_system/models"
	"gorm.io/gorm"
)

var TestDB *gorm.DB

// SetupTestDB initializes test database connection
func SetupTestDB() {
	// Setup test environment variables
	SetupTestEnv()

	// Initialize test database connection
	db, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	TestDB = db
	fmt.Println("Test database connection successful")

	// Drop existing tables in correct order
	TestDB.Migrator().DropTable(&models.CartItem{})
	TestDB.Migrator().DropTable(&models.GuestInteraction{})
	TestDB.Migrator().DropTable("trending_products")
	TestDB.Migrator().DropTable(&models.Product{})
	TestDB.Migrator().DropTable(&models.Session{})
	TestDB.Migrator().DropTable(&models.User{})

	// First migrate the base tables
	err = TestDB.AutoMigrate(
		&models.User{},
		&models.Product{}, // Products table must exist before trending_products
		&models.Session{},
		&models.CartItem{},
		&models.GuestInteraction{},
	)
	if err != nil {
		log.Fatal("Failed to migrate test database:", err)
	}

	// Now create trending_products table after products table exists
	err = TestDB.Exec(`
		CREATE TABLE IF NOT EXISTS trending_products (
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

// CleanupTestDB drops all test tables
func CleanupTestDB() {
	TestDB.Migrator().DropTable(&models.CartItem{})
	TestDB.Migrator().DropTable(&models.GuestInteraction{})
	TestDB.Migrator().DropTable(&models.Product{})
	TestDB.Migrator().DropTable(&models.Session{})
	TestDB.Migrator().DropTable(&models.User{})
}

// TruncateTable cleans a specific table between tests
func TruncateTable(tableName string) {
	// Disable foreign key checks
	TestDB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// Truncate table
	TestDB.Exec("TRUNCATE TABLE " + tableName)

	// Re-enable foreign key checks
	TestDB.Exec("SET FOREIGN_KEY_CHECKS = 1")
}
