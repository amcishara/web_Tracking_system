package models

import (
	"gorm.io/gorm"
)

type ProductRecommendation struct {
	ProductView         // Embed the ProductView struct for basic product info
	ViewCount   int     `json:"view_count,omitempty"`
	Relevance   float64 `json:"relevance_score,omitempty"`
}

// GetCollaborativeRecommendations returns exactly 5 most relevant products
func GetCollaborativeRecommendations(db *gorm.DB, productID uint, limit int) ([]ProductRecommendation, error) {
	var recommendations []ProductRecommendation

	// Try collaborative filtering first (users who viewed this also viewed)
	err := db.Raw(`
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
			WHERE p.id != ?
			GROUP BY p.id, p.name, p.description, p.price, p.category, p.stock
			HAVING COUNT(*) >= 2
		)
		SELECT * FROM ProductViews
		ORDER BY relevance_score DESC, view_count DESC
		LIMIT ?
	`, productID, productID, limit).Scan(&recommendations).Error

	// If we don't have enough recommendations, supplement with category-based
	if len(recommendations) < limit {
		var product Product
		if err := db.First(&product, productID).Error; err != nil {
			return nil, err
		}

		remainingCount := limit - len(recommendations)
		var categoryRecs []ProductRecommendation

		// Get popular products from same category
		err = db.Raw(`
			WITH CategoryPopular AS (
				SELECT 
					p.id, p.name, p.description, p.price, p.category, p.stock,
					COALESCE(ui.view_count, 0) + COALESCE(gi.view_count, 0) as view_count,
					COALESCE(ui.view_count, 0) + COALESCE(gi.view_count, 0) * 1.0 / 
						(SELECT MAX(COALESCE(ui_max.view_count, 0) + COALESCE(gi_max.view_count, 0)) 
						 FROM products p_max
						 LEFT JOIN (SELECT product_id, COUNT(*) as view_count FROM user_interactions GROUP BY product_id) ui_max ON p_max.id = ui_max.product_id
						 LEFT JOIN (SELECT product_id, COUNT(*) as view_count FROM guest_interactions GROUP BY product_id) gi_max ON p_max.id = gi_max.product_id
						 WHERE p_max.category = ?) as relevance_score
				FROM products p
				LEFT JOIN (
					SELECT product_id, COUNT(*) as view_count 
					FROM user_interactions 
					GROUP BY product_id
				) ui ON p.id = ui.product_id
				LEFT JOIN (
					SELECT product_id, COUNT(*) as view_count 
					FROM guest_interactions 
					GROUP BY product_id
				) gi ON p.id = gi.product_id
				WHERE p.category = ? AND p.id != ? AND p.id NOT IN (?)
			)
			SELECT * FROM CategoryPopular
			ORDER BY relevance_score DESC, view_count DESC, created_at DESC
			LIMIT ?
		`, product.Category, product.Category, productID,
			getProductIDs(recommendations), remainingCount).
			Scan(&categoryRecs).Error

		if err == nil {
			recommendations = append(recommendations, categoryRecs...)
		}
	}

	// If still not enough, add newest products
	if len(recommendations) < limit {
		remainingCount := limit - len(recommendations)
		var newProducts []ProductRecommendation

		err = db.Raw(`
			SELECT 
				p.id, p.name, p.description, p.price, p.category, p.stock,
				0 as view_count,
				DATEDIFF(CURRENT_TIMESTAMP, p.created_at) * -1 as relevance_score
			FROM products p
			WHERE p.id NOT IN (?)
			ORDER BY p.created_at DESC
			LIMIT ?
		`, getProductIDs(recommendations), remainingCount).
			Scan(&newProducts).Error

		if err == nil {
			recommendations = append(recommendations, newProducts...)
		}
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
						WHEN p.price BETWEEN ? * 0.8 AND ? * 1.2 THEN 50
						WHEN p.price BETWEEN ? * 0.6 AND ? * 1.4 THEN 30
						ELSE 10
					END +
					CASE WHEN p.stock > 0 THEN 20 ELSE 0 END
				) as relevance_score
			FROM products p
			LEFT JOIN (
				SELECT 
					product_id,
					COUNT(*) as total_views
				FROM (
					SELECT product_id FROM user_interactions
					UNION ALL
					SELECT product_id FROM guest_interactions
				) all_views
				GROUP BY product_id
			) t ON p.id = t.product_id
			WHERE p.category = ? 
			AND p.id != ?
		)
		SELECT * FROM CategoryScores
		ORDER BY relevance_score DESC, view_count DESC, id ASC
	`,
		product.Price, product.Price, product.Price, product.Price,
		product.Category, productID).
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
				(
					COALESCE(t.total_views, 0) + 
					CASE 
						WHEN p.price BETWEEN ? * 0.7 AND ? * 1.3 THEN 40
						WHEN p.price BETWEEN ? * 0.5 AND ? * 1.5 THEN 20
						ELSE 5
					END +
					CASE WHEN p.stock > 0 THEN 10 ELSE 0 END
				) as relevance_score
			FROM products p
			LEFT JOIN (
				SELECT 
					product_id,
					COUNT(*) as total_views
				FROM (
					SELECT product_id FROM user_interactions
					UNION ALL
					SELECT product_id FROM guest_interactions
				) all_views
				GROUP BY product_id
			) t ON p.id = t.product_id
			WHERE p.id != ? 
			AND p.id NOT IN (?)
			AND p.category != ?
			ORDER BY ABS(p.price - ?) ASC, relevance_score DESC
			LIMIT ?
		`,
			product.Price, product.Price, product.Price, product.Price,
			productID, getProductIDs(recommendations), product.Category,
			product.Price, remainingCount).
			Scan(&priceRangeRecs)

		if result.Error == nil && len(priceRangeRecs) > 0 {
			recommendations = append(recommendations, priceRangeRecs...)
		}
	}

	// If still not enough, get popular products from other categories
	if len(recommendations) < limit {
		remainingCount := limit - len(recommendations)
		var popularRecs []ProductRecommendation

		result = db.Raw(`
			SELECT 
				p.id, 
				p.name, 
				p.description, 
				p.price, 
				p.category, 
				p.stock,
				COALESCE(t.total_views, 0) as view_count,
				COALESCE(t.total_views, 0) as relevance_score
			FROM products p
			LEFT JOIN (
				SELECT 
					product_id,
					COUNT(*) as total_views
				FROM (
					SELECT product_id FROM user_interactions
					UNION ALL
					SELECT product_id FROM guest_interactions
				) all_views
				GROUP BY product_id
			) t ON p.id = t.product_id
			WHERE p.id != ? 
			AND p.id NOT IN (?)
			ORDER BY view_count DESC, created_at DESC
			LIMIT ?
		`,
			productID, getProductIDs(recommendations), remainingCount).
			Scan(&popularRecs)

		if result.Error == nil && len(popularRecs) > 0 {
			recommendations = append(recommendations, popularRecs...)
		}
	}

	return recommendations, result.Error
}
