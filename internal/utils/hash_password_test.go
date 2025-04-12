package utils_test

import (
	"testing"

	"AvitoPVZ/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateHashPassword_Success(t *testing.T) {
	password := "mysecretpassword"

	hash, err := utils.CreateHashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Errorf("hash does not match password: %v", err)
	}
}

func TestCreateHashPassword_EmptyPassword(t *testing.T) {
	hash, err := utils.CreateHashPassword("")
	if err != nil {
		t.Fatalf("unexpected error for empty password: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash for empty password")
	}

	valid, err := utils.CompareHashAndPassword(hash, "")
	if err != nil {
		t.Errorf("unexpected error comparing empty password: %v", err)
	}
	if !valid {
		t.Error("expected valid comparison for empty password")
	}
}

func TestCompareHashAndPassword_Success(t *testing.T) {
	password := "anothersecret"

	hash, err := utils.CreateHashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valid, err := utils.CompareHashAndPassword(hash, password)
	if err != nil {
		t.Fatalf("unexpected error comparing hash: %v", err)
	}
	if !valid {
		t.Error("expected passwords to match")
	}
}

func TestCompareHashAndPassword_Failure(t *testing.T) {
	password := "testpassword"
	wrongPassword := "wrongpassword"

	hash, err := utils.CreateHashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valid, err := utils.CompareHashAndPassword(hash, wrongPassword)
	if err == nil {
		t.Error("expected error comparing hash with wrong password")
	}
	if valid {
		t.Error("expected passwords not to match")
	}
}

func TestCompareHashAndPassword_InvalidHash(t *testing.T) {
	invalidHash := "notavalidhash"
	valid, err := utils.CompareHashAndPassword(invalidHash, "any")
	if err == nil {
		t.Error("expected error for invalid hash")
	}
	if valid {
		t.Error("expected false for invalid hash")
	}
}
