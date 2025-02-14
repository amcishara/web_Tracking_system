package cart_test

import (
	"os"
	"testing"

	"github.com/amcishara/web_Tracking_system/models"
	"github.com/amcishara/web_Tracking_system/tests/utils"
)

func TestMain(m *testing.M) {
	utils.SetupTestDB()
	code := m.Run()
	utils.PrintReport()
	utils.CleanupTestDB()
	os.Exit(code)
}

func setupTestData() (*models.User, *models.Product) {
	// Create test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "SecureP@ss123",
	}
	utils.TestDB.Create(user)

	// Create test product
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Category:    "Test Category",
		Stock:       100,
	}
	utils.TestDB.Create(product)

	return user, product
}

func TestAddToCart(t *testing.T) {
	utils.TruncateTable("cart_items")
	utils.TruncateTable("products")
	utils.TruncateTable("users")
	user, product := setupTestData()

	t.Run("Valid Add", func(t *testing.T) {
		err := models.AddToCart(utils.TestDB, user.UserID, product.ID, 2)
		passed := err == nil
		errMsg := ""
		if !passed {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Cart - Add Item", passed, errMsg)
	})

	t.Run("Invalid Product", func(t *testing.T) {
		err := models.AddToCart(utils.TestDB, user.UserID, 9999, 1)
		passed := err != nil && err.Error() == "product not found"
		errMsg := ""
		if !passed {
			errMsg = "Expected 'product not found' error"
		}
		utils.RecordTest(t, "Cart - Invalid Product", passed, errMsg)
	})

	t.Run("Insufficient Stock", func(t *testing.T) {
		err := models.AddToCart(utils.TestDB, user.UserID, product.ID, 101)
		passed := err != nil && err.Error() == "insufficient stock"
		errMsg := ""
		if !passed {
			errMsg = "Expected 'insufficient stock' error"
		}
		utils.RecordTest(t, "Cart - Insufficient Stock", passed, errMsg)
	})

	t.Run("Update Existing Item", func(t *testing.T) {
		// First add
		models.AddToCart(utils.TestDB, user.UserID, product.ID, 2)
		// Update quantity
		err := models.AddToCart(utils.TestDB, user.UserID, product.ID, 3)

		var cartItem models.CartItem
		utils.TestDB.Where("user_id = ? AND product_id = ?", user.UserID, product.ID).First(&cartItem)

		passed := err == nil && cartItem.Quantity == 3
		errMsg := ""
		if !passed {
			errMsg = "Failed to update cart item quantity"
		}
		utils.RecordTest(t, "Cart - Update Quantity", passed, errMsg)
	})
}

func TestRemoveFromCart(t *testing.T) {
	utils.TruncateTable("cart_items")
	utils.TruncateTable("products")
	utils.TruncateTable("users")
	user, product := setupTestData()

	t.Run("Valid Remove", func(t *testing.T) {
		// Add item first
		models.AddToCart(utils.TestDB, user.UserID, product.ID, 1)
		var cartItem models.CartItem
		utils.TestDB.Where("user_id = ? AND product_id = ?", user.UserID, product.ID).First(&cartItem)

		// Remove item
		err := models.RemoveFromCart(utils.TestDB, user.UserID, cartItem.ID)
		passed := err == nil
		errMsg := ""
		if !passed {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Cart - Remove Item", passed, errMsg)
	})

	t.Run("Remove Non-existent Item", func(t *testing.T) {
		err := models.RemoveFromCart(utils.TestDB, user.UserID, 9999)
		passed := err != nil && err.Error() == "item not found in cart"
		errMsg := ""
		if !passed {
			errMsg = "Expected 'item not found in cart' error"
		}
		utils.RecordTest(t, "Cart - Remove Non-existent Item", passed, errMsg)
	})
}

func TestGetCart(t *testing.T) {
	utils.TruncateTable("cart_items")
	utils.TruncateTable("products")
	utils.TruncateTable("users")
	user, product := setupTestData()

	t.Run("Empty Cart", func(t *testing.T) {
		summary, err := models.GetCart(utils.TestDB, user.UserID)
		passed := err == nil && len(summary.Items) == 0 && summary.TotalItems == 0 && summary.TotalPrice == 0
		errMsg := ""
		if !passed {
			errMsg = "Expected empty cart summary"
		}
		utils.RecordTest(t, "Cart - Empty Cart", passed, errMsg)
	})

	t.Run("Cart With Items", func(t *testing.T) {
		// Add items
		models.AddToCart(utils.TestDB, user.UserID, product.ID, 2)

		summary, err := models.GetCart(utils.TestDB, user.UserID)
		passed := err == nil &&
			len(summary.Items) == 1 &&
			summary.Items[0].ID == product.ID &&
			summary.Items[0].Name == product.Name &&
			summary.Items[0].Quantity == 2 &&
			summary.TotalItems == 2 &&
			summary.TotalPrice == 2*product.Price

		errMsg := ""
		if !passed {
			errMsg = "Cart summary does not match expected values"
		}
		utils.RecordTest(t, "Cart - With Items", passed, errMsg)
	})
}
