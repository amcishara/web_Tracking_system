package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CartItem represents an item in the cart
type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"-"` // Hide internal ID
	UserID    uint      `gorm:"not null" json:"-"`   // Hide UserID
	ProductID uint      `gorm:"not null" json:"-"`   // Hide ProductID
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"-"` // Hide timestamps
	UpdatedAt time.Time `json:"-"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"-"` // Hide full product
}

// CartItemResponse is the JSON response structure for cart items
type CartItemResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// CartSummary represents the cart summary with organized items
type CartSummary struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}

// Add TotalPrice as a computed field
func (ci *CartItem) TotalPrice() float64 {
	return float64(ci.Quantity) * ci.Product.Price
}

// Custom JSON marshaling to include total_price
func (ci CartItem) MarshalJSON() ([]byte, error) {
	type Alias CartItem
	return json.Marshal(&struct {
		Alias
		TotalPrice float64 `json:"total_price"`
	}{
		Alias:      Alias(ci),
		TotalPrice: ci.TotalPrice(),
	})
}

// AddToCart adds or updates an item in the cart
func AddToCart(db *gorm.DB, userID, productID uint, quantity int) error {
	// Verify product exists and has enough stock
	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return fmt.Errorf("product not found")
	}

	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock")
	}

	// Check if item already exists in cart
	var existingItem CartItem
	result := db.Where("user_id = ? AND product_id = ?", userID, productID).First(&existingItem)

	if result.Error == nil {
		// Update existing item quantity
		existingItem.Quantity = quantity // Replace old quantity with new
		return db.Save(&existingItem).Error
	}

	// Create new item if it doesn't exist
	cartItem := CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}

	return db.Create(&cartItem).Error
}

// RemoveFromCart removes an item from the cart
func RemoveFromCart(db *gorm.DB, userID, itemID uint) error {
	result := db.Where("id = ? AND user_id = ?", itemID, userID).Delete(&CartItem{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("item not found in cart")
	}
	return result.Error
}

// GetCart retrieves the cart items for a user with organized response
func GetCart(db *gorm.DB, userID uint) (*CartSummary, error) {
	var items []CartItem

	err := db.Preload("Product").
		Where("user_id = ?", userID).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	// Create organized response
	summary := &CartSummary{
		Items: make([]CartItemResponse, 0, len(items)),
	}

	for _, item := range items {
		// Create response item
		responseItem := CartItemResponse{
			ID:          item.Product.ID,
			Name:        item.Product.Name,
			Description: item.Product.Description,
			Price:       item.Product.Price,
			Category:    item.Product.Category,
			Quantity:    item.Quantity,
			Subtotal:    float64(item.Quantity) * item.Product.Price,
		}

		summary.Items = append(summary.Items, responseItem)
		summary.TotalItems += item.Quantity
		summary.TotalPrice += responseItem.Subtotal
	}

	return summary, nil
}

// TableName overrides the table name
func (CartItem) TableName() string {
	return "cart_items"
}
