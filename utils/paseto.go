package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/o1egl/paseto"
)

var pasetoV2 = paseto.NewV2()

// GeneratePasetoToken membuat token baru
func GeneratePasetoToken(userID uint, role string) (string, error) {
	// Ambil secret dari ENV
	secret := os.Getenv("PASETO_SECRET")
	if len(secret) < 32 {
		return "", fmt.Errorf("PASETO_SECRET must be at least 32 characters")
	}

	// Expired 24 jam
	exp := time.Now().Add(24 * time.Hour)

	// Custom payload
	jsonToken := paseto.JSONToken{
		Expiration: exp,
		NotBefore:  time.Now(),
		IssuedAt:   time.Now(),
		Subject:    fmt.Sprintf("%d", userID),
	}

	// include explicit custom claims to make extraction easier
	jsonToken.Set("user_id", fmt.Sprintf("%d", userID))
	jsonToken.Set("role", role)

	// Encrypt pakai v2.local (symmetric)
	token, err := pasetoV2.Encrypt([]byte(secret), jsonToken, nil)
	if err != nil {
		return "", err
	}

	return token, nil
}
