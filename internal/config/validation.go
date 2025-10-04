package config

import (
	"fmt"
	"os"
	"strings"
)

// ValidationError represents an environment validation error
type ValidationError struct {
	Variable string
	Message  string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Environment validation error for %s: %s", e.Variable, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

// ValidateEnvironment validates all required environment variables
// Returns an error if any required variables are missing or invalid
func ValidateEnvironment() error {
	var errors ValidationErrors

	// Validate JWT_SECRET - required, no default for security
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		errors = append(errors, ValidationError{
			Variable: "JWT_SECRET",
			Message:  "is required and must be set (no default provided for security)",
		})
	} else if len(jwtSecret) < 32 {
		errors = append(errors, ValidationError{
			Variable: "JWT_SECRET",
			Message:  "must be at least 32 characters long for security",
		})
	}

	// Validate DB_PASSWORD - required, no default for security
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		errors = append(errors, ValidationError{
			Variable: "DB_PASSWORD",
			Message:  "is required and must be set (no default provided for security)",
		})
	} else if len(dbPassword) < 8 {
		errors = append(errors, ValidationError{
			Variable: "DB_PASSWORD",
			Message:  "must be at least 8 characters long for security",
		})
	}

	// Validate other required variables with reasonable defaults
	requiredVars := map[string]string{
		"DB_HOST": "Database host",
		"DB_NAME": "Database name",
		"DB_USER": "Database user",
	}

	for envVar, description := range requiredVars {
		if os.Getenv(envVar) == "" {
			errors = append(errors, ValidationError{
				Variable: envVar,
				Message:  fmt.Sprintf("is required (%s)", description),
			})
		}
	}

	// Validate PORT if provided
	if port := os.Getenv("PORT"); port != "" {
		if !isValidPort(port) {
			errors = append(errors, ValidationError{
				Variable: "PORT",
				Message:  "must be a valid port number (1-65535)",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// isValidPort checks if the port string is a valid port number
func isValidPort(port string) bool {
	// Simple validation - check if it's a number between 1 and 65535
	if len(port) == 0 {
		return false
	}
	
	portNum := 0
	for _, char := range port {
		if char < '0' || char > '9' {
			return false
		}
		portNum = portNum*10 + int(char-'0')
		if portNum > 65535 {
			return false
		}
	}
	
	return portNum >= 1 && portNum <= 65535
}

// PrintValidationError prints a user-friendly validation error message
func PrintValidationError(err error) {
	fmt.Fprintf(os.Stderr, "\n❌ Environment Configuration Error\n")
	fmt.Fprintf(os.Stderr, "=====================================\n\n")
	
	if validationErrors, ok := err.(ValidationErrors); ok {
		for _, validationErr := range validationErrors {
			fmt.Fprintf(os.Stderr, "• %s\n", validationErr.Error())
		}
	} else {
		fmt.Fprintf(os.Stderr, "• %s\n", err.Error())
	}
	
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Please check your .env file and ensure all required variables are configured.\n")
	fmt.Fprintf(os.Stderr, "The application will not start without explicit security configuration.\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Required variables:\n")
	fmt.Fprintf(os.Stderr, "  - JWT_SECRET: Must be at least 32 characters (no default for security)\n")
	fmt.Fprintf(os.Stderr, "  - DB_PASSWORD: Must be at least 8 characters (no default for security)\n")
	fmt.Fprintf(os.Stderr, "  - DB_HOST: Database host address\n")
	fmt.Fprintf(os.Stderr, "  - DB_NAME: Database name\n")
	fmt.Fprintf(os.Stderr, "  - DB_USER: Database username\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "See .env.example for a template.\n")
}