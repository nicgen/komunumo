package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const Cost = 12

type BcryptHasher struct{}

func New() *BcryptHasher { return &BcryptHasher{} }

func (BcryptHasher) Hash(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("password.Hash: empty plaintext")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), Cost)
	if err != nil {
		return "", fmt.Errorf("password.Hash: %w", err)
	}
	return string(b), nil
}

func (BcryptHasher) Verify(hash, plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, fmt.Errorf("password.Verify: %w", err)
}
