package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/amcishara/web_Tracking_system/models"
)

func Migrate(db *gorm.DB) error {
	// Drop existing tables in correct order
	db.Migrator().DropTable(&models.CartItem{})
	db.Migrator().DropTable(&models.GuestInteraction{})
	db.Migrator().DropTable(&models.Product{})
	db.Migrator().DropTable(&models.Session{})
	db.Migrator().DropTable(&models.User{})

	// Create tables in correct order
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf("failed to migrate users table: %v", err)
	}

	if err := db.AutoMigrate(&models.Product{}); err != nil {
		return fmt.Errorf("failed to migrate products table: %v", err)
	}

	if err := db.AutoMigrate(&models.Session{}); err != nil {
		return fmt.Errorf("failed to migrate sessions table: %v", err)
	}

	if err := db.AutoMigrate(&models.CartItem{}); err != nil {
		return fmt.Errorf("failed to migrate cart items table: %v", err)
	}

	if err := db.AutoMigrate(&models.GuestInteraction{}); err != nil {
		return fmt.Errorf("failed to migrate guest interactions table: %v", err)
	}

	return nil
}
