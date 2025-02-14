package models

import (
	"gorm.io/gorm"
)

type ProductRecommendation struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
	ViewCount   int     `json:"-"` // Hide from JSON output but keep in struct
	Relevance   float64 `json:"-"` // Hide from JSON output but keep in struct
}

// GetCollaborativeRecommendations returns exactly 5 most relevant products
func GetCollaborativeRecommendations(db *gorm.DB, productID uint, limit int) ([]ProductRecommendation, error) {
	var recommendations []ProductRecommendation

	// Try collaborative filtering first (users who viewed this also viewed)
	result := db.Raw(`
		WITH ProductViews AS (
			SELECT 
				p.id, p.name, p.description, p.price, p.category, p.stock,
				COUNT(*) as view_count,
				COUNT(*) * 1.0 / (
					SELECT COUNT(*) FROM user_interactions 
					WHERE product_id = ui2.product_id
				) as relevance_score
			FROM products p
			JOIN user_interactions ui1 ON p.id = ui1.product_id
			JOIN user_interactions ui2 ON ui1.user_id = ui2.user_id AND ui2.product_id = ?
			WHERE p.id != ? AND p.stock > 0
			GROUP BY p.id, p.name, p.description, p.price, p.category, p.stock
			HAVING COUNT(*) >= 2
		)
		SELECT * FROM ProductViews
		ORDER BY relevance_score DESC, view_count DESC, id ASC
		LIMIT ?
	`, productID, productID, limit).Scan(&recommendations)

	if result.Error != nil {
		return nil, result.Error
	}

	// If we don't have enough recommendations, supplement with category-based
	if len(recommendations) < limit {
		var product Product
		if err := db.First(&product, productID).Error; err != nil {
			return nil, err
		}

		remainingCount := limit - len(recommendations)
		var categoryRecs []ProductRecommendation

		// Get popular products from same category
		result = db.Raw(`
			SELECT 
				p.id, p.name, p.description, p.price, p.category, p.stock,
				COALESCE(t.total_views, 0) as view_count,
				CASE 
					WHEN ABS(p.price - ?) <= 200 THEN 3
					WHEN ABS(p.price - ?) <= 400 THEN 2
					ELSE 1
				END as relevance_score
			FROM products p
			LEFT JOIN trending_products t ON p.id = t.product_id
			WHERE p.category = ? 
			AND p.id != ? 
			AND p.id NOT IN (?)
			AND p.stock > 0
			ORDER BY relevance_score DESC, view_count DESC, id ASC
			LIMIT ?
		`, product.Price, product.Price,
			product.Category, productID, getProductIDs(recommendations),
			remainingCount).
			Scan(&categoryRecs)

		if result.Error == nil && len(categoryRecs) > 0 {
			recommendations = append(recommendations, categoryRecs...)
		}

		return recommendations, result.Error
	}

	return recommendations, nil
}

// Helper function to extract product IDs from recommendations
func getProductIDs(recommendations []ProductRecommendation) []uint {
	ids := make([]uint, len(recommendations))
	for i, rec := range recommendations {
		ids[i] = rec.ID
	}
	return ids
}

// GetCategoryRecommendations returns exactly 5 products from the same category
func GetCategoryRecommendations(db *gorm.DB, productID uint, limit int) ([]ProductRecommendation, error) {
	var recommendations []ProductRecommendation

	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return nil, err
	}

	// First try: Get products from the same category
	result := db.Raw(`
		WITH CategoryScores AS (
			SELECT 
				p.id, 
				p.name, 
				p.description, 
				p.price, 
				p.category, 
				p.stock,
				COALESCE(t.total_views, 0) as view_count,
				(
					COALESCE(t.total_views, 0) + 
					CASE 
						WHEN ABS(p.price - ?) <= 200 THEN 50
						WHEN ABS(p.price - ?) <= 400 THEN 30
						ELSE 10
					END +
					CASE WHEN p.stock > 0 THEN 20 ELSE 0 END
				) as relevance_score
			FROM products p
			LEFT JOIN trending_products t ON p.id = t.product_id
			WHERE p.category = ? 
			AND p.id != ?
			AND p.stock > 0
			LIMIT ?
		)
		SELECT * FROM CategoryScores
		ORDER BY relevance_score DESC, view_count DESC, id ASC
	`,
		product.Price, product.Price,
		product.Category, productID, limit).
		Scan(&recommendations)

	// If we don't have enough recommendations, get products from similar price range
	if len(recommendations) < limit {
		remainingCount := limit - len(recommendations)
		var priceRangeRecs []ProductRecommendation

		result = db.Raw(`
			SELECT 
				p.id, 
				p.name, 
				p.description, 
				p.price, 
				p.category, 
				p.stock,
				COALESCE(t.total_views, 0) as view_count,
				ABS(p.price - ?) as price_diff
			FROM products p
			LEFT JOIN trending_products t ON p.id = t.product_id
			WHERE p.id != ? 
			AND p.id NOT IN (?)
			AND p.category != ?
			AND p.stock > 0
			AND ABS(p.price - ?) <= 300
			ORDER BY price_diff ASC, view_count DESC, id ASC
			LIMIT ?
		`,
			product.Price, productID, getProductIDs(recommendations), product.Category,
			product.Price, remainingCount).
			Scan(&priceRangeRecs)

		if result.Error == nil && len(priceRangeRecs) > 0 {
			recommendations = append(recommendations, priceRangeRecs...)
		}
	}

	return recommendations, result.Error
}
