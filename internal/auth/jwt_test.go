package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := []byte("test-secret")
	token, err := GenerateToken(secret, 42, "user@example.com", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", claims.Email)
	}
}

func TestParseTokenExpired(t *testing.T) {
	secret := []byte("test-secret")
	token, err := GenerateToken(secret, 1, "user@example.com", -time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ParseToken(secret, token)
	if err != ErrExpiredToken {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
}

func TestParseTokenInvalidSignature(t *testing.T) {
	secret := []byte("test-secret")
	token, err := GenerateToken(secret, 1, "user@example.com", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ParseToken([]byte("wrong-secret"), token)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseTokenMalformed(t *testing.T) {
	secret := []byte("test-secret")
	_, err := ParseToken(secret, "not-a-valid-token")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
