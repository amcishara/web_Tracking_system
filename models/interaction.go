package models

import (
	"time"

	"gorm.io/gorm"
)

// For registered users
type UserInteraction struct {
	UserID    uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	ProductID uint      `gorm:"primaryKey;not null" json:"product_id"`
	ViewedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"viewed_at"`
	User      User      `gorm:"foreignKey:UserID;references:UserID"`
	Product   Product   `gorm:"foreignKey:ProductID"`
}

// Custom struct for view history response
type ProductView struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

// Track product view for authenticated user
func TrackUserView(db *gorm.DB, userID uint, productID uint) error {
	// First get product title
	var product Product
	if err := db.Select("name").First(&product, productID).Error; err != nil {
		return err
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Track user interaction
	interaction := UserInteraction{
		UserID:    userID,
		ProductID: productID,
	}
	if err := tx.Create(&interaction).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update trending count
	if err := UpdateTrendingViews(tx, productID, product.Name); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Get user's view history
func GetUserViewHistory(db *gorm.DB, userID uint) ([]ProductView, error) {
	var products []ProductView
	err := db.Table("user_interactions").
		Select("products.id, products.name, products.description, products.price, products.category, products.stock").
		Joins("JOIN products ON user_interactions.product_id = products.id").
		Where("user_interactions.user_id = ?", userID).
		Order("user_interactions.viewed_at DESC").
		Find(&products).Error
	return products, err
}
