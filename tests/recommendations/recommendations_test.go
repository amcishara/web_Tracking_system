package recommendations_test

import (
	"os"
	"testing"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/tests/utils"
)

func TestMain(m *testing.M) {
	// Setup
	utils.SetupTestDB()

	// Run tests
	code := m.Run()

	// Cleanup and print report
	utils.PrintReport()
	utils.CleanupTestDB()

	os.Exit(code)
}

func setupTestProducts(t *testing.T) []models.Product {
	utils.TruncateTable("user_interactions")
	utils.TruncateTable("guest_interactions")
	utils.TruncateTable("products")

	// Create diverse test products
	products := []models.Product{
		{Name: "iPhone 13", Description: "Latest iPhone", Price: 999.99, Category: "Smartphones", Stock: 50},
		{Name: "Samsung S21", Description: "Android Flagship", Price: 899.99, Category: "Smartphones", Stock: 45},
		{Name: "Google Pixel", Description: "Google Phone", Price: 799.99, Category: "Smartphones", Stock: 30},
		{Name: "AirPods Pro", Description: "Wireless Earbuds", Price: 249.99, Category: "Accessories", Stock: 100},
		{Name: "Galaxy Watch", Description: "Smartwatch", Price: 299.99, Category: "Accessories", Stock: 40},
		{Name: "iPad Pro", Description: "Pro Tablet", Price: 1099.99, Category: "Tablets", Stock: 35},
		{Name: "MacBook Air", Description: "Lightweight Laptop", Price: 1299.99, Category: "Laptops", Stock: 25},
	}

	// Create products and store IDs
	for i := range products {
		if err := utils.TestDB.Create(&products[i]).Error; err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	return products
}

func TestCollaborativeRecommendations(t *testing.T) {
	products := setupTestProducts(t)

	t.Run("Basic Collaborative Filtering", func(t *testing.T) {
		// Simulate user interactions
		// User 1: iPhone -> AirPods pattern
		models.TrackUserView(utils.TestDB, 1, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 1, products[3].ID) // AirPods

		// User 2: iPhone -> Galaxy Watch pattern
		models.TrackUserView(utils.TestDB, 2, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 2, products[4].ID) // Galaxy Watch

		// User 3: iPhone -> AirPods pattern (reinforcing)
		models.TrackUserView(utils.TestDB, 3, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 3, products[3].ID) // AirPods

		// Get recommendations for iPhone
		recs, err := models.GetCollaborativeRecommendations(utils.TestDB, products[0].ID, 5)

		passed := err == nil && len(recs) > 0
		if passed {
			// AirPods should be recommended first (2 users)
			passed = recs[0].ID == products[3].ID
		}

		utils.RecordTest(t, "Collaborative - Basic Pattern", passed, "Expected AirPods as top recommendation")
	})

	t.Run("Cross-Category Recommendations", func(t *testing.T) {
		// User views multiple categories
		models.TrackUserView(utils.TestDB, 4, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 4, products[5].ID) // iPad
		models.TrackUserView(utils.TestDB, 4, products[6].ID) // MacBook

		recs, err := models.GetCollaborativeRecommendations(utils.TestDB, products[0].ID, 5)

		passed := err == nil && len(recs) > 0
		if passed {
			// Check for cross-category recommendations
			categories := make(map[string]bool)
			for _, rec := range recs {
				categories[rec.Category] = true
			}
			passed = len(categories) > 1
		}

		utils.RecordTest(t, "Collaborative - Cross Category", passed, "Expected recommendations from multiple categories")
	})
}

func TestCategoryRecommendations(t *testing.T) {
	products := setupTestProducts(t)

	t.Run("Same Category Priority", func(t *testing.T) {
		recs, err := models.GetCategoryRecommendations(utils.TestDB, products[0].ID, 5)

		passed := err == nil && len(recs) > 0
		if passed {
			// First recommendations should be from same category
			passed = recs[0].Category == "Smartphones"
		}

		utils.RecordTest(t, "Category - Same Category Priority", passed, "Expected same category recommendations first")
	})

	t.Run("Price Range Within Category", func(t *testing.T) {
		recs, err := models.GetCategoryRecommendations(utils.TestDB, products[2].ID, 5) // Google Pixel

		passed := err == nil && len(recs) > 0
		if passed {
			// Check if recommendations are within reasonable price range
			for _, rec := range recs {
				priceDiff := rec.Price - products[2].Price
				if priceDiff < -300 || priceDiff > 300 {
					passed = false
					break
				}
			}
		}

		utils.RecordTest(t, "Category - Price Range", passed, "Expected recommendations within similar price range")
	})

	t.Run("Stock Availability", func(t *testing.T) {
		// Set some products out of stock
		utils.TestDB.Model(&models.Product{}).Where("id = ?", products[1].ID).Update("stock", 0)

		recs, err := models.GetCategoryRecommendations(utils.TestDB, products[0].ID, 5)

		passed := err == nil && len(recs) > 0
		if passed {
			// Verify no out-of-stock products are recommended
			for _, rec := range recs {
				if rec.Stock == 0 {
					passed = false
					break
				}
			}
		}

		utils.RecordTest(t, "Category - Stock Check", passed, "Expected only in-stock recommendations")
	})
}

func TestHybridRecommendations(t *testing.T) {
	products := setupTestProducts(t)

	t.Run("Combined Recommendations", func(t *testing.T) {
		// Create user interactions to ensure collaborative recommendations
		models.TrackUserView(utils.TestDB, 1, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 1, products[3].ID) // AirPods
		models.TrackUserView(utils.TestDB, 2, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 2, products[3].ID) // AirPods
		models.TrackUserView(utils.TestDB, 3, products[0].ID) // iPhone
		models.TrackUserView(utils.TestDB, 3, products[4].ID) // Galaxy Watch

		// Get both types of recommendations
		collab, err1 := models.GetCollaborativeRecommendations(utils.TestDB, products[0].ID, 3)
		cat, err2 := models.GetCategoryRecommendations(utils.TestDB, products[0].ID, 3)

		passed := err1 == nil && err2 == nil &&
			len(collab) > 0 && len(cat) > 0

		if passed {
			// Verify collaborative recs include accessories (cross-category)
			hasAccessory := false
			for _, rec := range collab {
				if rec.Category == "Accessories" {
					hasAccessory = true
					break
				}
			}

			// Verify category recs are smartphones
			hasSmartphone := false
			for _, rec := range cat {
				if rec.Category == "Smartphones" {
					hasSmartphone = true
					break
				}
			}

			passed = hasAccessory && hasSmartphone
		}

		errMsg := ""
		if err1 != nil {
			errMsg = "Collaborative error: " + err1.Error()
		} else if err2 != nil {
			errMsg = "Category error: " + err2.Error()
		} else if !passed {
			errMsg = "Expected different types of recommendations (cross-category and same-category)"
		}

		utils.RecordTest(t, "Hybrid - Diverse Recommendations", passed, errMsg)
	})
}
