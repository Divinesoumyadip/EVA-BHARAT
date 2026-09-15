package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("supersecret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "supersecret123" {
		t.Fatal("password was not hashed")
	}

	valid, err := VerifyPassword("supersecret123", hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Fatal("expected password to verify successfully")
	}
}

func TestVerifyPasswordWrongPassword(t *testing.T) {
	hash, err := HashPassword("supersecret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valid, err := VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatal("expected verification to fail for wrong password")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	hash1, _ := HashPassword("samepassword")
	hash2, _ := HashPassword("samepassword")
	if hash1 == hash2 {
		t.Fatal("expected different hashes due to random salt")
	}
}
