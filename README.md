# 🛍️ Web Tracking System

A smart e-commerce tracking system that provides personalized product recommendations based on user behavior.

## 📊 System Design

[ER Diagram Preview](https://dbdiagram.io/d/67ac55c9263d6cf9a0de6d79)

### Core Entities
- Users (Admin, Customer, Guest)
- Products
- Cart Items
- User/Guest Interactions
- Sessions
- Trending Products

## 🛠️ Technologies & Tools

### Backend
- Go (1.19+)
- GORM (ORM for Go)
- MySQL 8.0
- JWT for Authentication
- Gin Web Framework


## 🌟 Features

### 👥 User Management
- Role-based access (Admin/Customer/Guest)
- Secure authentication with JWT
- Session management


### 📊 Product Analytics
- View tracking for both users and guests
- Trending products calculation
- User interaction history


### 🎯 Smart Recommendations
- Collaborative filtering
- Category-based recommendations
- Price-range matching
- Hybrid recommendation system

### 🛒 Shopping Features
- Cart management
- Product search and filtering
- Stock validation
- Order tracking (planned)

## 🔒 Security Features
1. Password hashing with bcrypt
2. JWT token-based authentication
3. Role-based access control
4. Input validation


## 🚀 Getting Started

### Prerequisites
- Go 1.19 or higher
- MySQL 8.0
- Git

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/web_tracking_system.git
cd web_tracking_system
```

2. Install dependencies
```bash
go mod tidy
```

3. Set up MySQL database
```bash
mysql -u root -p
CREATE DATABASE web_db;
```

4. Configure environment variables (create .env file)
```env
DB_HOST=localhost
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=web_db
JWT_SECRET=your_secret_key
```

5. Run migrations
```bash
go run migrations/migrate.go
```

6. Start the server
```bash
go run .
```

## 📝 API Documentation

### Public Endpoints
- `POST /signup` - Create new user account
- `POST /login` - Authenticate user
- `GET /products` - List all products
- `GET /trending` - Get trending products

### Customer Endpoints (Authenticated)
- `GET /cart` - View shopping cart
- `POST /cart` - Add item to cart
- `DELETE /cart/:id` - Remove item from cart


### Admin Endpoints
- `POST /admin/products` - Create product
- `POST /admin/products/bulk` - Bulk create products
- `GET /admin/analytics` - View system analytics
- `GET /admin/users` - Manage users

## 🧪 Testing

Run all tests:
```bash
go test -v ./tests/...
```

### Test Coverage
- Authentication & Authorization
- Product Management
- Cart Operations
- User Interactions
- Recommendation Engine





