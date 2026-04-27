package account

import "unicode"

// ValidatePassword checks that the password meets the minimum requirements:
// at least 12 characters, containing uppercase, lowercase, digit, and special characters.
func ValidatePassword(password string) error {
	if len(password) < 12 {
		return ErrPasswordTooShort
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrPasswordTooWeak
	}
	return nil
}
