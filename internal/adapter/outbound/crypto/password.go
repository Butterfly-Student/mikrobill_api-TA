package crypto_adapter

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	outbound_port "MikrOps/internal/port/outbound"
)

type passwordHasherAdapter struct{}

func NewPasswordHasherAdapter() outbound_port.PasswordHasherPort {
	return &passwordHasherAdapter{}
}

func (a *passwordHasherAdapter) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

func (a *passwordHasherAdapter) VerifyPassword(hashedPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}
	return true, nil
}

func (a *passwordHasherAdapter) GenerateRandomToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(b)%len(charset)]
		if i < len(charset)-1 {
			b[i] = charset[i%len(charset)]
		} else {
			b[i] = charset[len(b)%len(charset)]
		}
	}
	return string(b)
}
