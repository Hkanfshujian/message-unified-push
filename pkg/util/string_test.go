package util

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	const length = 64
	value, err := GenerateRandomString(length)
	if err != nil {
		t.Fatalf("GenerateRandomString returned error: %v", err)
	}
	if len(value) != length {
		t.Fatalf("length = %d, want %d", len(value), length)
	}
	alphabet := lowercaseLetters + uppercaseLetters + digits
	for _, char := range value {
		if !strings.ContainsRune(alphabet, char) {
			t.Fatalf("unexpected character %q", char)
		}
	}
}

func TestGenerateRandomStringRejectsNegativeLength(t *testing.T) {
	if _, err := GenerateRandomString(-1); err == nil {
		t.Fatal("expected an error for a negative length")
	}
}
