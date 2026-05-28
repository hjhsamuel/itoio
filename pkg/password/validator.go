package password

import (
	"errors"
)

const (
	hasNumber = 1 << iota
	hasLower
	hasUpper
	hasSpecial
)

func Validator(password string) error {
	if len(password) < 8 {
		return errors.New("password should larger than 8")
	}

	var flags uint8
	for _, c := range password {
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			return errors.New("password should not contain space")
		}
		switch {
		case c >= '0' && c <= '9':
			flags |= hasNumber
		case c >= 'a' && c <= 'z':
			flags |= hasLower
		case c >= 'A' && c <= 'Z':
			flags |= hasUpper
		default:
			flags |= hasSpecial
		}
	}

	var require uint8 = hasNumber | hasLower | hasUpper | hasSpecial
	if flags != require {
		return errors.New("password should contain at least one number, one lowercase letter, one uppercase letter, and one special character")
	}
	return nil
}
