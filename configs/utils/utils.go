package utils

import (
	"github.com/google/uuid"
)

// generateUUID creates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}
