// pkg/jwt/jwt.go
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims defines the payload stored inside every JWT token.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	// jti (JWT ID) is a unique identifier for this specific token instance.
	// It is what we store in the Redis blacklist on logout.
	// Without it, we'd have to store the entire token string — wasteful and
	// unnecessary. The jti is a compact, opaque handle for the token.
	JTI string `json:"jti"`
	jwt.RegisteredClaims
}

// RemainingTime returns how much time is left before this token expires.
// We use this as the Redis TTL so blacklist entries clean themselves up
// exactly when the token would have expired anyway.
func (c *Claims) RemainingTime() time.Duration {
	if c.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(c.ExpiresAt.Time)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Manager holds the signing secret and expiry configuration.
type Manager struct {
	secret []byte
	expiry time.Duration
}

func NewManager(secret string, expiryHours int) *Manager {
	return &Manager{
		secret: []byte(secret),
		expiry: time.Duration(expiryHours) * time.Hour,
	}
}

// Generate creates and signs a new JWT.
// Every call produces a token with a globally unique jti.
func (m *Manager) Generate(userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		// uuid.NewString() uses crypto/rand — cryptographically random,
		// globally unique, zero chance of collision across all tokens
		// ever issued by your application.
		JTI: uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID, // optional, but can be useful for some libraries and conventions
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("jwt: failed to sign token: %w", err)
	}

	return signed, nil
}

// Validate parses and verifies a JWT string, returning the claims.
func (m *Manager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("jwt: unexpected signing method: %v", token.Header["alg"])
			}
			return m.secret, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)
