package password

import (
	"fmt"
	"testing"
)

func TestArgon2Password(t *testing.T) {
	password := "super_secret$password"

	// Test Hashing
	hash, err := HashPassword(password, DefaultParams)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("Expected non-empty hash")
	}
	fmt.Println(hash)

	// Test Verification (Correct Password)
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !match {
		t.Fatal("Expected password to match")
	}

	// Test Verification (Wrong Password)
	match, err = VerifyPassword("wrong_password", hash)
	if err != nil {
		t.Fatalf("Failed to verify password with wrong input: %v", err)
	}
	if match {
		t.Fatal("Expected password not to match")
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	_, err := VerifyPassword("password", "invalid-hash-format")
	if err == nil {
		t.Fatal("Expected error for invalid hash format")
	}
}
