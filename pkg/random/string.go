package random

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func encodeBase62(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}
	res := ""
	for num > 0 {
		rem := num % 62
		res = string(base62Chars[rem]) + res
		num = num / 62
	}
	return res
}

func Generate62Str(key string) (string, error) {
	hash := sha256.Sum256([]byte(key))
	num := binary.BigEndian.Uint64(hash[:8])
	code := encodeBase62(num)
	if len(code) > 8 {
		return code[:8], nil
	}
	return fmt.Sprintf("%08s", code), nil
}
