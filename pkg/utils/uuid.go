package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID v4
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// ParseUUID parses a string into a UUID, returning an error if invalid
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// UUIDToString converts a UUID to string, handling nil UUID
func UUIDToString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}