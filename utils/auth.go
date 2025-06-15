package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain text password and returns a bcrypt hash
// Cost of 14 provides a good balance between security and performance
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a plain text password against a bcrypt hash
// Returns true if they match, false otherwise
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
