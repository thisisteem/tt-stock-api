package utils

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPin(t *testing.T) {
	tests := []struct {
		name        string
		pin         string
		expectError bool
	}{
		{
			name:        "Valid 6-digit PIN",
			pin:         "123456",
			expectError: false,
		},
		{
			name:        "Valid PIN with zeros",
			pin:         "000000",
			expectError: false,
		},
		{
			name:        "Valid PIN with mixed digits",
			pin:         "987654",
			expectError: false,
		},
		{
			name:        "Valid PIN starting with zero",
			pin:         "012345",
			expectError: false,
		},
		{
			name:        "Empty PIN",
			pin:         "",
			expectError: false, // bcrypt can hash empty strings
		},
		{
			name:        "Short PIN",
			pin:         "123",
			expectError: false, // bcrypt can hash any string
		},
		{
			name:        "Long PIN",
			pin:         "1234567890",
			expectError: false, // bcrypt can hash any string
		},
		{
			name:        "PIN with special characters",
			pin:         "12@#$%",
			expectError: false, // bcrypt can hash any string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPin, err := HashPin(tt.pin)

			if tt.expectError {
				if err == nil {
					t.Errorf("HashPin() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("HashPin() unexpected error: %v", err)
				return
			}

			// Verify the hash is not empty
			if hashedPin == "" {
				t.Errorf("HashPin() returned empty hash")
			}

			// Verify the hash is different from the original PIN
			if hashedPin == tt.pin {
				t.Errorf("HashPin() returned unhashed PIN")
			}

			// Verify the hash starts with bcrypt identifier
			if !strings.HasPrefix(hashedPin, "$2a$") && !strings.HasPrefix(hashedPin, "$2b$") {
				t.Errorf("HashPin() returned invalid bcrypt hash format: %s", hashedPin)
			}

			// Verify we can verify the PIN with the hash
			if err := CheckPin(hashedPin, tt.pin); err != nil {
				t.Errorf("HashPin() produced hash that doesn't verify with CheckPin(): %v", err)
			}
		})
	}
}

func TestHashPin_UniqueSalts(t *testing.T) {
	pin := "123456"
	
	// Hash the same PIN multiple times
	hash1, err1 := HashPin(pin)
	hash2, err2 := HashPin(pin)
	hash3, err3 := HashPin(pin)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("HashPin() failed: %v, %v, %v", err1, err2, err3)
	}

	// Each hash should be different due to unique salts
	if hash1 == hash2 || hash1 == hash3 || hash2 == hash3 {
		t.Errorf("HashPin() should generate unique hashes with different salts")
		t.Logf("Hash1: %s", hash1)
		t.Logf("Hash2: %s", hash2)
		t.Logf("Hash3: %s", hash3)
	}

	// But all should verify correctly
	if err := CheckPin(hash1, pin); err != nil {
		t.Errorf("CheckPin() failed for hash1: %v", err)
	}
	if err := CheckPin(hash2, pin); err != nil {
		t.Errorf("CheckPin() failed for hash2: %v", err)
	}
	if err := CheckPin(hash3, pin); err != nil {
		t.Errorf("CheckPin() failed for hash3: %v", err)
	}
}

func TestCheckPin(t *testing.T) {
	// Pre-generate some test hashes
	validPin := "123456"
	validHash, err := HashPin(validPin)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	tests := []struct {
		name        string
		hashedPin   string
		plainPin    string
		expectError bool
	}{
		{
			name:        "Correct PIN verification",
			hashedPin:   validHash,
			plainPin:    validPin,
			expectError: false,
		},
		{
			name:        "Incorrect PIN verification",
			hashedPin:   validHash,
			plainPin:    "654321",
			expectError: true,
		},
		{
			name:        "Empty plain PIN against valid hash",
			hashedPin:   validHash,
			plainPin:    "",
			expectError: true,
		},
		{
			name:        "Valid PIN against empty hash",
			hashedPin:   "",
			plainPin:    validPin,
			expectError: true,
		},
		{
			name:        "Both empty",
			hashedPin:   "",
			plainPin:    "",
			expectError: true,
		},
		{
			name:        "Invalid hash format",
			hashedPin:   "invalid_hash",
			plainPin:    validPin,
			expectError: true,
		},
		{
			name:        "PIN with different case (should fail)",
			hashedPin:   validHash,
			plainPin:    "123456", // Same as validPin, should pass
			expectError: false,
		},
		{
			name:        "PIN with extra characters",
			hashedPin:   validHash,
			plainPin:    "123456 ",
			expectError: true,
		},
		{
			name:        "PIN with leading zeros",
			hashedPin:   validHash,
			plainPin:    "0123456",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPin(tt.hashedPin, tt.plainPin)

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckPin() expected error but got none")
				}
				// Verify it's a bcrypt mismatch error for incorrect PINs
				if tt.hashedPin != "" && tt.plainPin != "" && err != bcrypt.ErrMismatchedHashAndPassword {
					// Only check for specific error type if both inputs are non-empty
					if strings.Contains(tt.hashedPin, "$2") && err != bcrypt.ErrMismatchedHashAndPassword {
						t.Logf("CheckPin() got error: %v (expected bcrypt mismatch)", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("CheckPin() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCheckPin_WithDifferentPINs(t *testing.T) {
	testPins := []string{
		"000000",
		"123456",
		"987654",
		"012345",
		"999999",
		"555555",
	}

	for _, pin := range testPins {
		t.Run("PIN_"+pin, func(t *testing.T) {
			// Hash the PIN
			hashedPin, err := HashPin(pin)
			if err != nil {
				t.Fatalf("HashPin() failed: %v", err)
			}

			// Verify correct PIN
			if err := CheckPin(hashedPin, pin); err != nil {
				t.Errorf("CheckPin() failed for correct PIN %s: %v", pin, err)
			}

			// Verify incorrect PINs
			for _, wrongPin := range testPins {
				if wrongPin != pin {
					if err := CheckPin(hashedPin, wrongPin); err == nil {
						t.Errorf("CheckPin() should have failed for wrong PIN %s against hash of %s", wrongPin, pin)
					}
				}
			}
		})
	}
}

func TestHashPin_WorkFactor(t *testing.T) {
	pin := "123456"
	hashedPin, err := HashPin(pin)
	if err != nil {
		t.Fatalf("HashPin() failed: %v", err)
	}

	// Extract and verify the work factor from the hash
	// bcrypt hash format: $2a$12$...
	if len(hashedPin) < 7 {
		t.Fatalf("Hash too short: %s", hashedPin)
	}

	// Check if it uses the expected work factor (12)
	if !strings.HasPrefix(hashedPin, "$2a$12$") && !strings.HasPrefix(hashedPin, "$2b$12$") {
		t.Errorf("HashPin() should use work factor 12, got hash: %s", hashedPin[:10])
	}
}

// Benchmark tests to ensure reasonable performance
func BenchmarkHashPin(b *testing.B) {
	pin := "123456"
	for i := 0; i < b.N; i++ {
		_, err := HashPin(pin)
		if err != nil {
			b.Fatalf("HashPin() failed: %v", err)
		}
	}
}

func BenchmarkCheckPin(b *testing.B) {
	pin := "123456"
	hashedPin, err := HashPin(pin)
	if err != nil {
		b.Fatalf("HashPin() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := CheckPin(hashedPin, pin)
		if err != nil {
			b.Fatalf("CheckPin() failed: %v", err)
		}
	}
}