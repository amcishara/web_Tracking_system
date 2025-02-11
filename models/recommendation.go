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

	// First get recommendations from user interactions
	err := db.Table("products").
		Select("products.id, products.name, products.description, products.price, products.category, products.stock, COUNT(*) as view_count").
		Joins("JOIN user_interactions ui1 ON products.id = ui1.product_id").
		Joins("JOIN user_interactions ui2 ON ui1.user_id = ui2.user_id AND ui2.product_id = ?", productID).
		Where("products.id != ?", productID).
		Group("products.id").
		Order("view_count DESC").
		Limit(limit).
		Find(&recommendations).Error

	if err != nil {
		return nil, err
	}

	// If we don't have enough recommendations, add from guest interactions
	if len(recommendations) < limit {
		var guestRecommendations []ProductRecommendation
		err := db.Table("products").
			Select("products.id, products.name, products.description, products.price, products.category, products.stock, COUNT(*) as view_count").
			Joins("JOIN guest_interactions gi1 ON products.id = gi1.product_id").
			Joins("JOIN guest_interactions gi2 ON gi1.guest_id = gi2.guest_id AND gi2.product_id = ?", productID).
			Where("products.id != ?", productID).
			Group("products.id").
			Order("view_count DESC").
			Limit(limit - len(recommendations)).
			Find(&guestRecommendations).Error

		if err == nil {
			recommendations = append(recommendations, guestRecommendations...)
		}
	}

	return recommendations, nil
}

// GetCategoryRecommendations returns other popular products in the same category
func GetCategoryRecommendations(db *gorm.DB, productID uint, limit int) ([]ProductRecommendation, error) {
	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return nil, err
	}

	var recommendations []ProductRecommendation

	// Combine view counts from both user and guest interactions
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
		Order("view_count DESC").
		Limit(limit)

	err := query.Find(&recommendations).Error
	return recommendations, err
}
