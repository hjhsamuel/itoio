package password

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	numbers  = "0123456789"
	lower    = "abcdefghijklmnopqrstuvwxyz"
	upper    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	special  = "!@#$%^&*()-_=+[]{}<>?/|"
	allChars = numbers + lower + upper + special
)

func Random(length int) (string, error) {
	if length < 8 {
		return "", errors.New("password should larger than 8")
	}

	password := make([]byte, length)

	required := []byte{
		randomChar(numbers),
		randomChar(lower),
		randomChar(upper),
		randomChar(special),
	}

	copy(password, required)

	for i := len(required); i < length; i++ {
		password[i] = randomChar(allChars)
	}

	// Fisher-Yates
	for i := len(password) - 1; i > 0; i-- {
		j, err := cryptoRandInt(i + 1)
		if err != nil {
			return "", err
		}
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

func randomChar(charset string) byte {
	idx, _ := cryptoRandInt(len(charset))
	return charset[idx]
}

func cryptoRandInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}
