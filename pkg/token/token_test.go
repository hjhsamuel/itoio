package token

import (
	"testing"
	"time"
)

func TestJWT(t *testing.T) {
	secret := "test_secret"
	userID := "123456"
	deviceID := "test_device"
	duration := time.Hour

	// Test Generate
	token, err := Encode(userID, deviceID, secret, duration)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Test Parse
	claims, err := Decode(token, secret)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	if claims.ID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.ID)
	}

	// Test Invalid Secret
	_, err = Decode(token, "wrong_secret")
	if err == nil {
		t.Error("Expected error for wrong secret, got nil")
	}

	// Test Expired Token
	expiredToken, err := Encode(userID, deviceID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}
	_, err = Decode(expiredToken, secret)
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}
