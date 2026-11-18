package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidClaims is returned when token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Claims represents JWT claims for the agent
type Claims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	issuer        string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration, issuer string) (*JWTManager, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}
	if tokenDuration <= 0 {
		return nil, errors.New("token duration must be positive")
	}

	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
		issuer:        issuer,
	}, nil
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(userID, username string, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secretKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (m *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh of expired tokens within a grace period (e.g., 7 days)
		if !errors.Is(err, ErrExpiredToken) {
			return "", err
		}

		// Try to parse expired token
		token, parseErr := jwt.ParseWithClaims(
			tokenString,
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return m.secretKey, nil
			},
			jwt.WithoutClaimsValidation(),
		)

		if parseErr != nil {
			return "", fmt.Errorf("failed to parse expired token: %w", parseErr)
		}

		var ok bool
		claims, ok = token.Claims.(*Claims)
		if !ok {
			return "", ErrInvalidClaims
		}

		// Check if token is within grace period (7 days after expiration)
		gracePeriod := 7 * 24 * time.Hour
		if time.Since(claims.ExpiresAt.Time) > gracePeriod {
			return "", errors.New("token expired beyond grace period")
		}
	}

	// Generate new token with same user info
	return m.GenerateToken(claims.UserID, claims.Username, claims.Roles)
}

// HasRole checks if the claims contain a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims contain any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin role
func (c *Claims) IsAdmin() bool {
	return c.HasRole("admin")
}

// GenerateSecretKey generates a random secret key for JWT signing
func GenerateSecretKey(length int) (string, error) {
	if length < 32 {
		return "", errors.New("secret key length must be at least 32 bytes")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// TokenInfo represents information about a token
type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Issuer    string    `json:"issuer"`
}

// GetTokenInfo extracts information from a token without full validation
func (m *JWTManager) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Roles:     claims.Roles,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
		Issuer:    claims.Issuer,
	}, nil
}

// Made with Bob
