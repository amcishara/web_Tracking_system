package product_test

import (
	"testing"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/tests/utils"
)

func TestProductRecommendations(t *testing.T) {
	utils.TruncateTable("user_interactions")
	utils.TruncateTable("products")

	// Create test products
	products := []models.Product{
		{Name: "iPhone 13", Description: "Latest iPhone", Price: 999.99, Category: "Smartphones", Stock: 50},
		{Name: "Samsung S21", Description: "Android Flagship", Price: 899.99, Category: "Smartphones", Stock: 45},
		{Name: "Google Pixel", Description: "Google Phone", Price: 799.99, Category: "Smartphones", Stock: 30},
		{Name: "AirPods Pro", Description: "Wireless Earbuds", Price: 249.99, Category: "Accessories", Stock: 100},
		{Name: "Galaxy Watch", Description: "Smartwatch", Price: 299.99, Category: "Accessories", Stock: 40},
	}

	// Create products and store IDs
	for i := range products {
		if err := utils.TestDB.Create(&products[i]).Error; err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	t.Run("Collaborative Recommendations", func(t *testing.T) {
		// Simulate user interactions
		// User 1 views iPhone and AirPods
		models.TrackUserView(utils.TestDB, 1, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 1, products[3].ID) // AirPods

		// User 2 views iPhone and Galaxy Watch
		models.TrackUserView(utils.TestDB, 2, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 2, products[4].ID) // Galaxy Watch

		// Get recommendations for iPhone
		recs, err := models.GetCollaborativeRecommendations(utils.TestDB, products[0].ID, 5)

		passed := err == nil && len(recs) > 0
		if passed {
			// Check if accessories viewed by other users are recommended
			hasAccessory := false
			for _, rec := range recs {
				if rec.Category == "Accessories" {
					hasAccessory = true
					break
				}
			}
			passed = hasAccessory
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else if !passed {
			errMsg = "Expected accessory recommendations for users who viewed similar products"
		}
		utils.RecordTest(t, "Recommendations - Collaborative", passed, errMsg)
	})

	t.Run("Category Recommendations", func(t *testing.T) {
		// Get recommendations for iPhone
		recs, err := models.GetCategoryRecommendations(utils.TestDB, products[0].ID, 5)

		passed := err == nil && len(recs) > 0
		if passed {
			// Check if other smartphones are recommended
			hasSmartphone := false
			for _, rec := range recs {
				if rec.Category == "Smartphones" && rec.ID != products[0].ID {
					hasSmartphone = true
					break
				}
			}
			passed = hasSmartphone
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else if !passed {
			errMsg = "Expected same category recommendations"
		}
		utils.RecordTest(t, "Recommendations - Category", passed, errMsg)
	})

	t.Run("Price Range Recommendations", func(t *testing.T) {
		// Get recommendations for mid-range phone
		recs, err := models.GetCategoryRecommendations(utils.TestDB, products[2].ID, 5) // Google Pixel

		passed := err == nil && len(recs) > 0
		if passed {
			// Check if products in similar price range are recommended
			hasSimilarPrice := false
			for _, rec := range recs {
				priceDiff := rec.Price - products[2].Price
				if priceDiff >= -200 && priceDiff <= 200 {
					hasSimilarPrice = true
					break
				}
			}
			passed = hasSimilarPrice
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else if !passed {
			errMsg = "Expected recommendations in similar price range"
		}
		utils.RecordTest(t, "Recommendations - Price Range", passed, errMsg)
	})
}
