package models

import (
	"gorm.io/gorm"
)

// TrendingProduct represents a trending product with its view count
type TrendingProduct struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
	ViewCount   int     `json:"view_count,omitempty"`
}

// TrendingProductDB is the database model for trending products
type TrendingProductDB struct {
	ProductID  uint   `gorm:"primaryKey;column:product_id"`
	Title      string `gorm:"column:title"`
	TotalViews int    `gorm:"column:total_views"`
}

// TableName sets the table name for TrendingProductDB
func (TrendingProductDB) TableName() string {
	return "trending_products"
}

// GetTrendingProducts returns top N trending products that are in stock
func GetTrendingProducts(db *gorm.DB, limit int) ([]TrendingProduct, error) {
	var trending []TrendingProduct

	err := db.Raw(`
        SELECT 
            p.id,
            p.name,
            p.description,
            p.price,
            p.category,
            p.stock,
            COALESCE(t.total_views, 0) as view_count
        FROM products p
        LEFT JOIN trending_products t ON p.id = t.product_id
        WHERE p.stock > 0  -- Only include products with stock
        ORDER BY COALESCE(t.total_views, 0) DESC, p.created_at DESC
        LIMIT ?
    `, limit).Scan(&trending).Error

	return trending, err
}

// UpdateTrendingViews updates the view count for a product in trending_products
func UpdateTrendingViews(tx *gorm.DB, productID uint, title string) error {
	result := tx.Exec(`
        INSERT INTO trending_products (product_id, title, total_views)
        VALUES (?, ?, 1)
        ON DUPLICATE KEY UPDATE 
        total_views = total_views + 1,
        title = ?
    `, productID, title, title)

	return result.Error
}
