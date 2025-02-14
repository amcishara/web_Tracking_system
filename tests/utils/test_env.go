package utils

import (
	"os"
)

func SetupTestEnv() {
	// Set test database configuration
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWORD", "your_test_password")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("DB_PORT", "3306")
}
