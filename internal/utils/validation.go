package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents validation error details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// EmailValidator validates email format using RFC 5322
func EmailValidator(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

// PasswordValidator validates password (minimum 6 characters)
func PasswordValidator(password string) bool {
	return len(password) >= 6
}

// PhoneValidator validates Indian phone number (10 digits, starting with 6-9)
func PhoneValidator(phone string) bool {
	// Remove spaces and hyphens
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if len(phone) != 10 {
		return false
	}

	// Check if all characters are digits
	if !regexp.MustCompile(`^\d+$`).MatchString(phone) {
		return false
	}

	// Check if starts with 6-9
	firstDigit := phone[0]
	return firstDigit >= '6' && firstDigit <= '9'
}

// NameValidator validates name length (2-50 characters)
func NameValidator(name string) bool {
	trimmed := strings.TrimSpace(name)
	return len(trimmed) >= 2 && len(trimmed) <= 50
}

// AddressValidator validates address length (10-200 characters)
func AddressValidator(address string) bool {
	trimmed := strings.TrimSpace(address)
	return len(trimmed) >= 10 && len(trimmed) <= 200
}

// PincodeValidator validates Indian pincode (exactly 6 digits)
func PincodeValidator(pincode string) bool {
	// Remove spaces and hyphens
	pincode = strings.ReplaceAll(pincode, " ", "")
	pincode = strings.ReplaceAll(pincode, "-", "")

	if len(pincode) != 6 {
		return false
	}

	return regexp.MustCompile(`^\d{6}$`).MatchString(pincode)
}

// DescriptionValidator validates description max length (1000 characters)
func DescriptionValidator(description string) bool {
	return len(description) <= 1000
}

// CommentValidator validates comment length (10-500 characters)
func CommentValidator(comment string) bool {
	trimmed := strings.TrimSpace(comment)
	return len(trimmed) >= 10 && len(trimmed) <= 500
}

// ObjectIDValidator validates MongoDB ObjectID format
func ObjectIDValidator(id string) bool {
	if len(id) != 24 {
		return false
	}
	_, err := strconv.ParseInt(id, 16, 64)
	return err == nil
}

// RatingValidator validates rating (1-5)
func RatingValidator(rating int) bool {
	return rating >= 1 && rating <= 5
}

// CategoryValidator validates category enum
func CategoryValidator(category string) bool {
	validCategories := []string{"upper", "lower", "shoes"}
	for _, v := range validCategories {
		if v == category {
			return true
		}
	}
	return false
}

// PaymentMethodValidator validates payment method enum
func PaymentMethodValidator(method string) bool {
	validMethods := []string{"credit_card", "debit_card", "upi", "net_banking", "wallet", "cod"}
	for _, v := range validMethods {
		if v == method {
			return true
		}
	}
	return false
}

// OrderStatusValidator validates order status enum
func OrderStatusValidator(status string) bool {
	validStatuses := []string{"pending", "confirmed", "shipped", "delivered", "cancelled", "refunded"}
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}
