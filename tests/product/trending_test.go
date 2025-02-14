package product_test

import (
	"testing"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/tests/utils"
)

func TestTrendingProducts(t *testing.T) {
	utils.TruncateTable("trending_products")
	utils.TruncateTable("products")

	// Create test products
	products := []models.Product{
		{Name: "Popular Phone", Description: "Best Seller", Price: 999.99, Category: "Smartphones", Stock: 50},
		{Name: "Regular TV", Description: "Normal TV", Price: 499.99, Category: "TVs", Stock: 30},
		{Name: "Gaming Laptop", Description: "Fast Laptop", Price: 1499.99, Category: "Laptops", Stock: 20},
	}

	// Create products and store them back in the array to get their IDs
	for i := range products {
		if err := utils.TestDB.Create(&products[i]).Error; err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	t.Run("Update Trending Views", func(t *testing.T) {
		// Use the first product with its assigned ID
		err := models.UpdateTrendingViews(utils.TestDB, products[0].ID, products[0].Name)
		passed := err == nil
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Trending - Update Views", passed, errMsg)
	})

	t.Run("Get Trending Products", func(t *testing.T) {
		product := products[0] // Use the first product

		// View products multiple times
		for i := 0; i < 5; i++ {
			models.UpdateTrendingViews(utils.TestDB, product.ID, product.Name)
		}

		product2 := products[1] // Use the second product
		for i := 0; i < 3; i++ {
			models.UpdateTrendingViews(utils.TestDB, product2.ID, product2.Name)
		}

		trending, err := models.GetTrendingProducts(utils.TestDB, 5)
		passed := err == nil &&
			len(trending) > 0 &&
			trending[0].Name == product.Name && // Most viewed should be first
			trending[0].ViewCount > trending[1].ViewCount // Verify order

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Trending - Get Products", passed, errMsg)
	})

	t.Run("Zero Stock Products", func(t *testing.T) {
		// Set stock to 0 for most viewed product
		utils.TestDB.Model(&models.Product{}).Where("id = ?", products[0].ID).Update("stock", 0)

		trending, err := models.GetTrendingProducts(utils.TestDB, 5)
		passed := err == nil
		if len(trending) > 0 {
			// First product should now be the one with stock (Regular TV)
			passed = passed && trending[0].Name == "Regular TV"
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else if !passed {
			errMsg = "Zero stock product appeared in trending list"
		}
		utils.RecordTest(t, "Trending - Zero Stock Handling", passed, errMsg)
	})

	t.Run("Limit Results", func(t *testing.T) {
		trending, err := models.GetTrendingProducts(utils.TestDB, 2)
		passed := err == nil && len(trending) == 2

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else if !passed {
			errMsg = "Expected exactly 2 trending products"
		}
		utils.RecordTest(t, "Trending - Limit Results", passed, errMsg)
	})
}
