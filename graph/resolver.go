package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Resolver struct {
	RedisClient *redis.Client
}

// Helper functions for pointer conversion
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i uint) *int32 {
	v := int32(i)
	return &v
}

func float64Ptr(f float64) *float64 {
	return &f
}

// createToken generates a JWT token for the given user ID
func createToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}
