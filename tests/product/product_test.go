package product_test

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

func TestCreateProduct(t *testing.T) {
	utils.TruncateTable("products")

	t.Run("Valid Product", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			Category:    "Test Category",
			Stock:       100,
		}

		err := models.CreateProduct(utils.TestDB, product)
		passed := err == nil && product.ID != 0
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Product - Create Valid", passed, errMsg)
	})

	t.Run("Duplicate Name", func(t *testing.T) {
		product := &models.Product{
			Name:        "Test Product",
			Description: "Another Description",
			Price:       79.99,
			Category:    "Test Category",
			Stock:       50,
		}

		err := models.CreateProduct(utils.TestDB, product)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for duplicate product name"
		}
		utils.RecordTest(t, "Product - Create Duplicate", passed, errMsg)
	})

	t.Run("Invalid Price", func(t *testing.T) {
		product := &models.Product{
			Name:        "Negative Price Product",
			Description: "Test Description",
			Price:       -10.00,
			Category:    "Test Category",
			Stock:       100,
		}

		err := models.CreateProduct(utils.TestDB, product)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for negative price"
		}
		utils.RecordTest(t, "Product - Create Invalid Price", passed, errMsg)
	})
}

func TestGetProduct(t *testing.T) {
	utils.TruncateTable("products")

	// Create test product
	testProduct := &models.Product{
		Name:        "Get Test Product",
		Description: "Test Description",
		Price:       99.99,
		Category:    "Test Category",
		Stock:       100,
	}
	utils.TestDB.Create(testProduct)

	t.Run("Get Existing Product", func(t *testing.T) {
		product, err := models.GetProductByID(utils.TestDB, int(testProduct.ID))
		passed := err == nil && product.Name == testProduct.Name
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Product - Get Existing", passed, errMsg)
	})

	t.Run("Get Non-existent Product", func(t *testing.T) {
		_, err := models.GetProductByID(utils.TestDB, 9999)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for non-existent product"
		}
		utils.RecordTest(t, "Product - Get Non-existent", passed, errMsg)
	})
}

func TestSearchProducts(t *testing.T) {
	utils.TruncateTable("products")

	// Create test products
	products := []models.Product{
		{Name: "iPhone 13", Description: "Apple Smartphone", Price: 999.99, Category: "Smartphones", Stock: 50},
		{Name: "Samsung TV", Description: "4K Smart TV", Price: 799.99, Category: "TVs", Stock: 30},
		{Name: "Gaming Mouse", Description: "RGB Gaming Mouse", Price: 59.99, Category: "Accessories", Stock: 100},
	}
	for _, p := range products {
		utils.TestDB.Create(&p)
	}

	t.Run("Search by Name", func(t *testing.T) {
		results, err := models.SearchProducts(utils.TestDB, "iPhone", "", "", "")
		passed := err == nil && len(results) == 1 && results[0].Name == "iPhone 13"
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Product - Search by Name", passed, errMsg)
	})

	t.Run("Search by Category", func(t *testing.T) {
		results, err := models.SearchProducts(utils.TestDB, "", "Smartphones", "", "")
		passed := err == nil && len(results) == 1 && results[0].Category == "Smartphones"
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Product - Search by Category", passed, errMsg)
	})

	t.Run("Sort by Price", func(t *testing.T) {
		results, err := models.SearchProducts(utils.TestDB, "", "", "price", "desc")
		passed := err == nil && len(results) > 0 && results[0].Price == 999.99
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Product - Sort by Price", passed, errMsg)
	})
}

func TestUpdateProduct(t *testing.T) {
	utils.TruncateTable("products")

	// Create test product
	product := &models.Product{
		Name:        "Original Product",
		Description: "Original Description",
		Price:       99.99,
		Category:    "Test Category",
		Stock:       100,
	}
	utils.TestDB.Create(product)

	t.Run("Valid Update", func(t *testing.T) {
		product.Name = "Updated Product"
		product.Price = 149.99
		err := models.UpdateProduct(utils.TestDB, product)
		passed := err == nil
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		utils.RecordTest(t, "Product - Update Valid", passed, errMsg)
	})

	t.Run("Update to Existing Name", func(t *testing.T) {
		// Create another product
		otherProduct := &models.Product{
			Name:        "Other Product",
			Description: "Other Description",
			Price:       79.99,
			Category:    "Test Category",
			Stock:       50,
		}
		utils.TestDB.Create(otherProduct)

		// Try to update to existing name
		otherProduct.Name = "Updated Product"
		err := models.UpdateProduct(utils.TestDB, otherProduct)
		passed := err != nil
		errMsg := ""
		if !passed {
			errMsg = "Expected error for duplicate name"
		}
		utils.RecordTest(t, "Product - Update Duplicate Name", passed, errMsg)
	})
}
