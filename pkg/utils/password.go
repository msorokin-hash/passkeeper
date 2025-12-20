package utils

import (
	"errors"
	"unicode"
)

// ValidatePassword checks whether the given password meets the minimum
// security requirements.
//
// The password must:
//   - be at least 8 characters long
//   - contain at least one letter
//   - contain at least one digit
//   - not contain whitespace characters
//
// It returns an error describing the first violated rule, or nil if the
// password is considered valid.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasLetter, hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsSpace(r):
			return errors.New("password must not contain spaces")
		}
	}

	if !hasLetter || !hasDigit {
		return errors.New("password must contain at least one letter and one digit")
	}

	return nil
}
