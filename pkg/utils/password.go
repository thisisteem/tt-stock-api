package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPin hashes a PIN using bcrypt with a secure work factor
// Returns the hashed PIN or an error if hashing fails
func HashPin(pin string) (string, error) {
	// Use work factor of 12 for good security/performance balance
	// bcrypt automatically generates a unique salt for each hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pin), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPin verifies a plain text PIN against a hashed PIN
// Returns nil if the PIN matches, or an error if it doesn't match or verification fails
func CheckPin(hashedPin, plainPin string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPin), []byte(plainPin))
}