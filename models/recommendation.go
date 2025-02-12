package models

import (
	"gorm.io/gorm"
)

type ProductRecommendation struct {
	ProductView     // Embed the ProductView struct for basic product info
	ViewCount   int `json:"view_count,omitempty"`
}

// GetCollaborativeRecommendations returns "Users who viewed this also viewed"
func GetCollaborativeRecommendations(db *gorm.DB, productID uint, limit int) ([]ProductRecommendation, error) {
	var recommendations []ProductRecommendation

	// First try to get collaborative recommendations
	err := db.Table("products").
		Select("products.id, products.name, products.description, products.price, products.category, products.stock, COUNT(*) as view_count").
		Joins("JOIN user_interactions ui1 ON products.id = ui1.product_id").
		Joins("JOIN user_interactions ui2 ON ui1.user_id = ui2.user_id AND ui2.product_id = ?", productID).
		Where("products.id != ?", productID).
		Group("products.id").
		Order("view_count DESC").
		Limit(limit).
		Find(&recommendations).Error

	// If no collaborative recommendations found (new product), fall back to category-based
	if err != nil || len(recommendations) == 0 {
		var product Product
		if err := db.First(&product, productID).Error; err != nil {
			return nil, err
		}

		// Get popular products from same category
		err = db.Table("products").
			Select(`
				products.id, 
				products.name, 
				products.description, 
				products.price, 
				products.category, 
				products.stock,
				COALESCE(view_counts.total_views, 0) as view_count
			`).
			Joins(`LEFT JOIN (
				SELECT product_id, COUNT(*) as total_views 
				FROM user_interactions 
				GROUP BY product_id
			) view_counts ON products.id = view_counts.product_id`).
			Where("category = ? AND products.id != ?", product.Category, productID).
			Order("view_count DESC, created_at DESC"). // Consider newer products if no views
			Limit(limit).
			Find(&recommendations).Error

		// If still no recommendations, get newest products across all categories
		if err != nil || len(recommendations) == 0 {
			err = db.Table("products").
				Select("products.id, products.name, products.description, products.price, products.category, products.stock, 0 as view_count").
				Where("products.id != ?", productID).
				Order("created_at DESC").
				Limit(limit).
				Find(&recommendations).Error
		}
	}

	return recommendations, err
}

// GetCategoryRecommendations returns other popular products in the same category
func GetCategoryRecommendations(db *gorm.DB, productID uint, limit int) ([]ProductRecommendation, error) {
	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return nil, err
	}

	var recommendations []ProductRecommendation

	// Try to get popular products in the same category
	query := db.Table("products").
		Select(`
			products.id, 
			products.name, 
			products.description, 
			products.price, 
			products.category, 
			products.stock,
			(COALESCE(ui_counts.user_views, 0) + COALESCE(gi_counts.guest_views, 0)) as view_count
		`).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) as user_views 
			FROM user_interactions 
			GROUP BY product_id
		) ui_counts ON products.id = ui_counts.product_id`).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) as guest_views 
			FROM guest_interactions 
			GROUP BY product_id
		) gi_counts ON products.id = gi_counts.product_id`).
		Where("category = ? AND products.id != ?", product.Category, productID).
		Order("view_count DESC, created_at DESC"). // Consider newer products if no views
		Limit(limit)

	err := query.Find(&recommendations).Error

	// If no recommendations found, get newest products in the same category
	if err != nil || len(recommendations) == 0 {
		err = db.Table("products").
			Select("products.id, products.name, products.description, products.price, products.category, products.stock, 0 as view_count").
			Where("category = ? AND products.id != ?", product.Category, productID).
			Order("created_at DESC").
			Limit(limit).
			Find(&recommendations).Error
	}

	return recommendations, err
}
