package util

import (
	"crypto/rand"
	"fmt"
)

const (
	lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits           = "0123456789"
)

func GenerateRandomString(length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("random string length must not be negative")
	}
	alphabet := lowercaseLetters + uppercaseLetters + digits
	result := make([]byte, length)
	limit := byte(256 - (256 % len(alphabet)))
	for i := 0; i < length; {
		var sample [1]byte
		if _, err := rand.Read(sample[:]); err != nil {
			return "", fmt.Errorf("generate secure random string: %w", err)
		}
		if sample[0] >= limit {
			continue
		}
		result[i] = alphabet[int(sample[0])%len(alphabet)]
		i++
	}
	return string(result), nil
}

func GenerateUniqueID() (string, error) {
	return GenerateRandomString(10)
}
