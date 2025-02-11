package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateGuestID() string {
	return fmt.Sprintf("guest_%s", uuid.New().String())
}
