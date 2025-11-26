package middleware

import (
	"context"
	"crypto/rand"
	
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var jwtSecret = []byte("super-secret-key-change-this") // TODO: move to env var

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT token for a user (admin or normal)
func GenerateToken(userID, role string, expiryHours int) (string, error) {
	exp := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-2fa-email-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken parses and validates JWT
func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

//////////////////////////////////////////////
// 🧩  EMAIL-BASED 2FA (Admins only)
//////////////////////////////////////////////

// Generate2FACode creates a random 6-digit code and stores it in Redis
func Generate2FACode(ctx context.Context, rdb *redis.Client, adminEmail string) (string, error) {
	// random 30-byte base32 token then cut to 6 digits
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", int(b[0])%1000000)

	key := fmt.Sprintf("2fa:%s", adminEmail)
	if err := rdb.Set(ctx, key, code, 5*time.Minute).Err(); err != nil {
		return "", err
	}
	return code, nil
}

// Verify2FACode checks the submitted code against Redis
func Verify2FACode(ctx context.Context, rdb *redis.Client, adminEmail, submitted string) (bool, error) {
	key := fmt.Sprintf("2fa:%s", adminEmail)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, errors.New("code expired or not found")
	} else if err != nil {
		return false, err
	}
	if val != submitted {
		return false, nil
	}
	// One-time use
	rdb.Del(ctx, key)
	return true, nil
}
