package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name          string
		secretKey     string
		tokenDuration time.Duration
		issuer        string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid configuration",
			secretKey:     "test-secret-key-12345",
			tokenDuration: 24 * time.Hour,
			issuer:        "ipquorum-agent",
			expectError:   false,
		},
		{
			name:          "empty secret key",
			secretKey:     "",
			tokenDuration: 24 * time.Hour,
			issuer:        "ipquorum-agent",
			expectError:   true,
			errorContains: "secret key cannot be empty",
		},
		{
			name:          "zero duration",
			secretKey:     "test-secret-key",
			tokenDuration: 0,
			issuer:        "ipquorum-agent",
			expectError:   true,
			errorContains: "token duration must be positive",
		},
		{
			name:          "negative duration",
			secretKey:     "test-secret-key",
			tokenDuration: -1 * time.Hour,
			issuer:        "ipquorum-agent",
			expectError:   true,
			errorContains: "token duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewJWTManager(tt.secretKey, tt.tokenDuration, tt.issuer)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				assert.Equal(t, []byte(tt.secretKey), manager.secretKey)
				assert.Equal(t, tt.tokenDuration, manager.tokenDuration)
				assert.Equal(t, tt.issuer, manager.issuer)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-12345", 24*time.Hour, "ipquorum-agent")
	require.NoError(t, err)

	tests := []struct {
		name     string
		userID   string
		username string
		roles    []string
	}{
		{
			name:     "admin user",
			userID:   "user-123",
			username: "admin",
			roles:    []string{"admin", "user"},
		},
		{
			name:     "regular user",
			userID:   "user-456",
			username: "john",
			roles:    []string{"user"},
		},
		{
			name:     "user with no roles",
			userID:   "user-789",
			username: "guest",
			roles:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tt.userID, tt.username, tt.roles)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token can be parsed
			claims, err := manager.ValidateToken(token)
			assert.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, tt.roles, claims.Roles)
			assert.Equal(t, "ipquorum-agent", claims.Issuer)
			assert.Equal(t, tt.userID, claims.Subject)
		})
	}
}

func TestValidateToken(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-12345", 24*time.Hour, "ipquorum-agent")
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		token, err := manager.GenerateToken("user-123", "admin", []string{"admin"})
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "admin", claims.Username)
	})

	t.Run("invalid token format", func(t *testing.T) {
		claims, err := manager.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create manager with very short duration
		shortManager, err := NewJWTManager("test-secret-key-12345", 1*time.Millisecond, "ipquorum-agent")
		require.NoError(t, err)

		token, err := shortManager.GenerateToken("user-123", "admin", []string{"admin"})
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		claims, err := shortManager.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})

	t.Run("token with wrong secret", func(t *testing.T) {
		token, err := manager.GenerateToken("user-123", "admin", []string{"admin"})
		require.NoError(t, err)

		// Create manager with different secret
		wrongManager, err := NewJWTManager("wrong-secret-key-12345", 24*time.Hour, "ipquorum-agent")
		require.NoError(t, err)

		claims, err := wrongManager.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("tampered token", func(t *testing.T) {
		token, err := manager.GenerateToken("user-123", "admin", []string{"admin"})
		require.NoError(t, err)

		// Tamper with a character in the middle of the signature. The final base64
		// character only carries partial bits, so changing it is not always detected.
		idx := len(token) - 10
		replacement := byte('A')
		if token[idx] == replacement {
			replacement = 'B'
		}
		tamperedToken := token[:idx] + string(replacement) + token[idx+1:]

		claims, err := manager.ValidateToken(tamperedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestRefreshToken(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-12345", 24*time.Hour, "ipquorum-agent")
	require.NoError(t, err)

	t.Run("refresh valid token", func(t *testing.T) {
		originalToken, err := manager.GenerateToken("user-123", "admin", []string{"admin"})
		require.NoError(t, err)

		// Wait a full second to ensure new token has different timestamp
		time.Sleep(1 * time.Second)

		newToken, err := manager.RefreshToken(originalToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		// Note: tokens might be identical if generated at same second, so we just verify it works

		// Verify new token has same user info
		claims, err := manager.ValidateToken(newToken)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "admin", claims.Username)
		assert.Equal(t, []string{"admin"}, claims.Roles)
	})

	t.Run("refresh recently expired token", func(t *testing.T) {
		// Create manager with very short duration
		shortManager, err := NewJWTManager("test-secret-key-12345", 1*time.Millisecond, "ipquorum-agent")
		require.NoError(t, err)

		token, err := shortManager.GenerateToken("user-123", "admin", []string{"admin"})
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		// Should still be able to refresh within grace period
		newToken, err := shortManager.RefreshToken(token)
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
	})

	t.Run("refresh invalid token", func(t *testing.T) {
		newToken, err := manager.RefreshToken("invalid-token")
		assert.Error(t, err)
		assert.Empty(t, newToken)
	})
}

func TestClaimsHelpers(t *testing.T) {
	claims := &Claims{
		UserID:   "user-123",
		Username: "admin",
		Roles:    []string{"admin", "user", "operator"},
	}

	t.Run("HasRole", func(t *testing.T) {
		assert.True(t, claims.HasRole("admin"))
		assert.True(t, claims.HasRole("user"))
		assert.True(t, claims.HasRole("operator"))
		assert.False(t, claims.HasRole("superadmin"))
		assert.False(t, claims.HasRole(""))
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		assert.True(t, claims.HasAnyRole("admin"))
		assert.True(t, claims.HasAnyRole("admin", "superadmin"))
		assert.True(t, claims.HasAnyRole("superadmin", "admin"))
		assert.False(t, claims.HasAnyRole("superadmin", "guest"))
		assert.False(t, claims.HasAnyRole())
	})

	t.Run("IsAdmin", func(t *testing.T) {
		assert.True(t, claims.IsAdmin())

		nonAdminClaims := &Claims{
			UserID:   "user-456",
			Username: "john",
			Roles:    []string{"user"},
		}
		assert.False(t, nonAdminClaims.IsAdmin())
	})
}

func TestGenerateSecretKey(t *testing.T) {
	t.Run("valid length", func(t *testing.T) {
		key, err := GenerateSecretKey(32)
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
		assert.Greater(t, len(key), 32) // Base64 encoded will be longer
	})

	t.Run("longer key", func(t *testing.T) {
		key, err := GenerateSecretKey(64)
		assert.NoError(t, err)
		assert.NotEmpty(t, key)
	})

	t.Run("too short", func(t *testing.T) {
		key, err := GenerateSecretKey(16)
		assert.Error(t, err)
		assert.Empty(t, key)
		assert.Contains(t, err.Error(), "at least 32 bytes")
	})

	t.Run("uniqueness", func(t *testing.T) {
		key1, err := GenerateSecretKey(32)
		require.NoError(t, err)

		key2, err := GenerateSecretKey(32)
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2)
	})
}

func TestGetTokenInfo(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-12345", 24*time.Hour, "ipquorum-agent")
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		token, err := manager.GenerateToken("user-123", "admin", []string{"admin", "user"})
		require.NoError(t, err)

		info, err := manager.GetTokenInfo(token)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "user-123", info.UserID)
		assert.Equal(t, "admin", info.Username)
		assert.Equal(t, []string{"admin", "user"}, info.Roles)
		assert.Equal(t, "ipquorum-agent", info.Issuer)
		assert.False(t, info.IssuedAt.IsZero())
		assert.False(t, info.ExpiresAt.IsZero())
		assert.True(t, info.ExpiresAt.After(info.IssuedAt))
	})

	t.Run("invalid token", func(t *testing.T) {
		info, err := manager.GetTokenInfo("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestTokenExpiration(t *testing.T) {
	// JWT timestamps have 1s precision, so use 1s expiry and sleep well past it
	manager, err := NewJWTManager("test-secret-key-12345", 1*time.Second, "ipquorum-agent")
	require.NoError(t, err)

	token, err := manager.GenerateToken("user-123", "admin", []string{"admin"})
	require.NoError(t, err)

	// Token should be valid immediately
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err, "Token should be valid immediately after generation")
	require.NotNil(t, claims)

	// Wait for token to expire with buffer
	time.Sleep(2100 * time.Millisecond)

	// Token should now be expired
	claims, err = manager.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestTokenWithDifferentSigningMethods(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-12345", 24*time.Hour, "ipquorum-agent")
	require.NoError(t, err)

	// Create a token with RSA signing method (not HMAC)
	// This should fail because we only accept HMAC methods
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ipquorum-agent",
			Subject:   "user-123",
		},
		UserID:   "user-123",
		Username: "admin",
		Roles:    []string{"admin"},
	}

	// Use None signing method which should fail
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// Should fail validation due to unsupported signing method
	validatedClaims, err := manager.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

// Made with Bob
