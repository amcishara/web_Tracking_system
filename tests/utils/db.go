package utils

import (
	"fmt"
	"log"

	"github.com/amcishara/web_Tracking_system/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var TestDB *gorm.DB

// SetupTestDB initializes test database connection
func SetupTestDB() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	TestDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}
	fmt.Println("Test database connection successful")

	// Migrate test database
	err = TestDB.AutoMigrate(
		&models.User{},
		&models.Session{},
	)
	if err != nil {
		log.Fatal("Failed to migrate test database:", err)
	}
}

// CleanupTestDB drops all test tables
func CleanupTestDB() {
	TestDB.Migrator().DropTable(&models.User{})
	TestDB.Migrator().DropTable(&models.Session{})
}

// TruncateTable cleans a specific table between tests
func TruncateTable(tableName string) {
	TestDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
}
