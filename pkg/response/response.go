package response

import (
	"github.com/gofiber/fiber/v2"
)

// LoginResponse represents the response structure for successful authentication
type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"` // Access token expiration in seconds
		User         struct {
			ID          string `json:"id"`
			PhoneNumber string `json:"phone_number"`
		} `json:"user"`
	} `json:"data"`
}

// ErrorResponse represents the response structure for errors
type ErrorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// SuccessResponse represents the response structure for general success responses
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SendLoginSuccess sends a successful login response with tokens and user info
func SendLoginSuccess(c *fiber.Ctx, accessToken, refreshToken string, expiresIn int64, userID, phoneNumber string) error {
	response := LoginResponse{
		Success: true,
	}
	response.Data.AccessToken = accessToken
	response.Data.RefreshToken = refreshToken
	response.Data.ExpiresIn = expiresIn
	response.Data.User.ID = userID
	response.Data.User.PhoneNumber = phoneNumber

	return c.Status(fiber.StatusOK).JSON(response)
}

// SendError sends an error response with the specified status code, error code, and message
func SendError(c *fiber.Ctx, statusCode int, errorCode, message string) error {
	response := ErrorResponse{
		Success: false,
	}
	response.Error.Code = errorCode
	response.Error.Message = message

	return c.Status(statusCode).JSON(response)
}

// SendSuccess sends a general success response with optional data and message
func SendSuccess(c *fiber.Ctx, data interface{}, message string) error {
	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// SendValidationError sends a 400 Bad Request error for validation failures
func SendValidationError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", message)
}

// SendAuthenticationError sends a 401 Unauthorized error for authentication failures
func SendAuthenticationError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusUnauthorized, "AUTHENTICATION_ERROR", message)
}

// SendNotFoundError sends a 404 Not Found error
func SendNotFoundError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusNotFound, "NOT_FOUND", message)
}

// SendInternalServerError sends a 500 Internal Server Error
func SendInternalServerError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// SendTokenExpiredError sends a 401 Unauthorized error specifically for expired tokens
func SendTokenExpiredError(c *fiber.Ctx, message string) error {
	return SendError(c, fiber.StatusUnauthorized, "TOKEN_EXPIRED", message)
}