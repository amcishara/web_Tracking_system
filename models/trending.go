package models

import (
	"gorm.io/gorm"
)

type TrendingItem struct {
	ProductID    uint   `gorm:"primaryKey;column:product_id" json:"product_id"`
	ProductTitle string `gorm:"column:product_title" json:"product_title"`
	ViewCount    uint64 `gorm:"column:view_count" json:"view_count"`
}

// Update view count for a product
func UpdateTrendingItem(db *gorm.DB, productID uint, productTitle string) error {
	// Using raw SQL for upsert operation
	return db.Exec(`
        INSERT INTO trending_items (product_id, product_title, view_count)
        VALUES (?, ?, 1)
        ON DUPLICATE KEY UPDATE 
        view_count = view_count + 1,
        product_title = VALUES(product_title)
    `, productID, productTitle).Error
}

// Get trending items
func GetTrendingItems(db *gorm.DB, limit int) ([]TrendingItem, error) {
	var items []TrendingItem
	err := db.Table("trending_items").
		Order("view_count DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}
