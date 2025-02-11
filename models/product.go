package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Category    string    `json:"category"`
	Stock       int       `gorm:"not null" json:"stock"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func CreateProduct(db *gorm.DB, p *Product) error {
	var count int64
	db.Model(&Product{}).Where("name = ?", p.Name).Count(&count)
	if count > 0 {
		return fmt.Errorf("product with name '%s' already exists", p.Name)
	}

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
