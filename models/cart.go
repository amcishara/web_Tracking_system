package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product"`
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

// Add item to cart
func AddToCart(db *gorm.DB, userID uint, productID uint, quantity int) error {
	// Check if product exists and has enough stock
	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return err
	}
	if product.Stock < quantity {
		return fmt.Errorf("insufficient stock")
	}

	// Check if item already exists in cart
	var cartItem CartItem
	result := db.Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem)

	if result.Error == nil {
		// Update quantity if item exists
		cartItem.Quantity += quantity
		return db.Save(&cartItem).Error
	}

	// Create new cart item if it doesn't exist
	cartItem = CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}
	return db.Create(&cartItem).Error
}

// Remove item from cart
func RemoveFromCart(db *gorm.DB, userID uint, cartItemID uint) error {
	result := db.Where("id = ? AND user_id = ?", cartItemID, userID).Delete(&CartItem{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}
	return result.Error
}

// Add this new struct for cart summary
type CartSummary struct {
	CartItems  []CartItem `json:"cart_items"`
	CartTotal  float64    `json:"cart_total"`
	TotalItems int        `json:"total_items"`
}

// Update GetCart to return CartSummary
func GetCart(db *gorm.DB, userID uint) (*CartSummary, error) {
	var cartItems []CartItem
	err := db.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error
	if err != nil {
		return nil, err
	}

	// Calculate totals
	var cartTotal float64
	var totalItems int
	for i := range cartItems {
		cartItems[i].Product = cartItems[i].Product // Ensure Product is loaded
		cartTotal += cartItems[i].TotalPrice()
		totalItems += cartItems[i].Quantity
	}

	summary := &CartSummary{
		CartItems:  cartItems,
		CartTotal:  cartTotal,
		TotalItems: totalItems,
	}

	return summary, nil
}
