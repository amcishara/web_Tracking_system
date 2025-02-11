package models

import (
	"time"

	"gorm.io/gorm"
)

type GuestInteraction struct {
	GuestID   string    `gorm:"primaryKey;column:guest_id" json:"guest_id"`
	ProductID uint      `gorm:"primaryKey;not null" json:"product_id"`
	ViewedAt  time.Time `gorm:"primaryKey;default:CURRENT_TIMESTAMP" json:"viewed_at"`
	Product   Product   `gorm:"foreignKey:ProductID"`
}

// Track guest view
func TrackGuestView(db *gorm.DB, guestID string, productID uint) error {
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

	// Track guest interaction
	interaction := GuestInteraction{
		GuestID:   guestID,
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

// Get guest's view history
func GetGuestViewHistory(db *gorm.DB, guestID string) ([]ProductView, error) {
	var products []ProductView
	err := db.Table("guest_interactions").
		Select("products.id, products.name, products.description, products.price, products.category, products.stock").
		Joins("JOIN products ON guest_interactions.product_id = products.id").
		Where("guest_interactions.guest_id = ?", guestID).
		Order("guest_interactions.viewed_at DESC").
		Find(&products).Error
	return products, err
}
