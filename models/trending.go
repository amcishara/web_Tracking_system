package models

import (
	"gorm.io/gorm"
)

type TrendingProduct struct {
	ProductID  uint    `gorm:"primaryKey" json:"product_id"`
	Title      string  `json:"title"`
	TotalViews int     `gorm:"not null;default:0" json:"total_views"`
	Product    Product `gorm:"foreignKey:ProductID"`
}

// UpdateTrendingViews increments view count for a product
func UpdateTrendingViews(db *gorm.DB, productID uint, title string) error {
	return db.Exec(`
        INSERT INTO trending_products (product_id, title, total_views)
        VALUES (?, ?, 1)
        ON DUPLICATE KEY UPDATE 
        total_views = total_views + 1,
        title = VALUES(title)
    `, productID, title).Error
}

// GetTrendingProducts returns top N trending products
func GetTrendingProducts(db *gorm.DB, limit int) ([]TrendingProduct, error) {
	var trending []TrendingProduct
	err := db.Order("total_views DESC").Limit(limit).Find(&trending).Error
	return trending, err
}
