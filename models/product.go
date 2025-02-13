package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Category    string    `gorm:"not null" json:"category"`
	Stock       int       `gorm:"not null" json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Add this struct for API responses
type ProductResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
}

type ProductWithRecommendations struct {
	Product struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		Stock       int     `json:"stock"`
	} `json:"product"`
	CustomersAlsoViewed  []ProductRecommendation `json:"customers_also_viewed"`
	OtherRecommendations []ProductRecommendation `json:"other_recommendations"`
	TrendingProducts     []TrendingProduct       `json:"trending_products"`
}

func CreateProduct(db *gorm.DB, product *Product) error {
	return db.Create(product).Error
}

func GetAllProducts(db *gorm.DB) []Product {
	var products []Product
	db.Find(&products)
	return products
}

// Modify GetProductByID to use the new response type
func GetProductByID(db *gorm.DB, id int) (*ProductResponse, error) {
	var product Product
	result := db.First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// Convert to response type
	response := &ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		Stock:       product.Stock,
	}
	return response, nil
}

func UpdateProduct(db *gorm.DB, p *Product) error {
	var count int64
	db.Model(&Product{}).Where("name = ? AND id != ?", p.Name, p.ID).Count(&count)
	if count > 0 {
		return fmt.Errorf("product with name '%s' already exists", p.Name)
	}

	result := db.Save(p)
	return result.Error
}

func DeleteProduct(db *gorm.DB, id int) error {
	result := db.Delete(&Product{}, id)
	return result.Error
}

// Add this function to check if product name exists
func IsProductNameExists(db *gorm.DB, name string) bool {
	var count int64
	db.Model(&Product{}).Where("name = ?", name).Count(&count)
	return count > 0
}

// Add new function for search and sort
func SearchProducts(db *gorm.DB, query string, category string, sortBy string, order string) ([]Product, error) {
	var products []Product
	tx := db.Model(&Product{})

	// Apply category filter if provided
	if category != "" {
		tx = tx.Where("category = ?", category)
	}

	// Apply search if query exists (now only searches name and description)
	if query != "" {
		searchQuery := "%" + query + "%"
		tx = tx.Where("(name LIKE ? OR description LIKE ?)",
			searchQuery, searchQuery)
	}

	// Apply sorting
	switch sortBy {
	case "price":
		if order == "desc" {
			tx = tx.Order("price DESC")
		} else {
			tx = tx.Order("price ASC")
		}
	case "name":
		if order == "desc" {
			tx = tx.Order("name DESC")
		} else {
			tx = tx.Order("name ASC")
		}
	case "date": // Add sorting by date
		if order == "desc" {
			tx = tx.Order("created_at DESC")
		} else {
			tx = tx.Order("created_at ASC")
		}
	default:
		tx = tx.Order("id ASC") // Default sorting
	}

	err := tx.Find(&products).Error
	return products, err
}
