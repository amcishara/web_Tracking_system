package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Category    string    `gorm:"not null" json:"category"`
	Stock       int       `gorm:"not null" json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserInteraction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Type      string    `gorm:"not null" json:"type"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID"`
	Product   Product   `gorm:"foreignKey:ProductID"`
}

// ProductRepository interface for database operations
type ProductRepository interface {
	Create(product *Product) error
	GetAll() ([]Product, error)
	GetByID(id int) (*Product, error)
	Update(product *Product) error
	Delete(id int) error
}

// For now, we'll use an in-memory store
var products = make(map[int]Product)
var lastID = 0

// Repository methods
func CreateProduct(db *gorm.DB, p *Product) error {
	result := db.Create(p)
	return result.Error
}

func GetAllProducts(db *gorm.DB) []Product {
	var products []Product
	db.Find(&products)
	return products
}

func GetProductByID(db *gorm.DB, id int) (*Product, error) {
	var product Product
	result := db.First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

func UpdateProduct(db *gorm.DB, p *Product) error {
	result := db.Save(p)
	return result.Error
}

func DeleteProduct(db *gorm.DB, id int) error {
	result := db.Delete(&Product{}, id)
	return result.Error
}
