package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/storage"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret"
	tokenExpiry := 3600
	refreshExpiry := 7200

	manager := NewJWTManager(secret, tokenExpiry, refreshExpiry)

	if manager == nil {
		t.Fatal("NewJWTManager returned nil")
	}

	if string(manager.secret) != secret {
		t.Errorf("Expected secret %s, got %s", secret, string(manager.secret))
	}

	if manager.tokenExpiry != time.Duration(tokenExpiry)*time.Second {
		t.Errorf("Expected tokenExpiry %v, got %v", time.Duration(tokenExpiry)*time.Second, manager.tokenExpiry)
	}

	if manager.refreshExpiry != time.Duration(refreshExpiry)*time.Second {
		t.Errorf("Expected refreshExpiry %v, got %v", time.Duration(refreshExpiry)*time.Second, manager.refreshExpiry)
	}
}

func TestGenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateToken(user)

	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected 3 token parts, got %d", len(parts))
	}
}

func TestValidateToken_Success(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, claims.Username)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected Role %s, got %s", user.Role, claims.Role)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	_, err := manager.ValidateToken("invalid-token")

	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret1", 3600, 7200)
	manager2 := NewJWTManager("secret2", 3600, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager1.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = manager2.ValidateToken(token)

	if err == nil {
		t.Error("Expected error for token with wrong secret, got nil")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create manager with very short expiry
	manager := NewJWTManager("test-secret", 1, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	_, err = manager.ValidateToken(token)

	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateRefreshToken(user)

	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty refresh token")
	}

	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected 3 token parts, got %d", len(parts))
	}
}

func TestValidateRefreshToken_Success(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Refresh tokens use the same validation method
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, claims.UserID)
	}
}

func TestValidateRefreshToken_ExpiredToken(t *testing.T) {
	// Create manager with very short refresh expiry
	manager := NewJWTManager("test-secret", 3600, 1)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	_, err = manager.ValidateToken(token)

	if err == nil {
		t.Error("Expected error for expired refresh token, got nil")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	tests := []struct {
		name string
		user *storage.User
	}{
		{
			name: "admin user",
			user: &storage.User{
				ID:       "admin-123",
				Username: "admin",
				Role:     "admin",
			},
		},
		{
			name: "operator user",
			user: &storage.User{
				ID:       "op-456",
				Username: "operator",
				Role:     "operator",
			},
		},
		{
			name: "viewer user",
			user: &storage.User{
				ID:       "viewer-789",
				Username: "viewer",
				Role:     "viewer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tt.user)
			if err != nil {
				t.Fatalf("GenerateToken failed: %v", err)
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken failed: %v", err)
			}

			if claims.UserID != tt.user.ID {
				t.Errorf("Expected UserID %s, got %s", tt.user.ID, claims.UserID)
			}

			if claims.Username != tt.user.Username {
				t.Errorf("Expected Username %s, got %s", tt.user.Username, claims.Username)
			}

			if claims.Role != tt.user.Role {
				t.Errorf("Expected Role %s, got %s", tt.user.Role, claims.Role)
			}
		})
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	users := []*storage.User{
		{ID: "user-1", Username: "user1", Role: "admin"},
		{ID: "user-2", Username: "user2", Role: "operator"},
		{ID: "user-3", Username: "user3", Role: "viewer"},
	}

	for _, user := range users {
		t.Run(user.ID, func(t *testing.T) {
			token, err := manager.GenerateRefreshToken(user)
			if err != nil {
				t.Fatalf("GenerateRefreshToken failed: %v", err)
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken failed: %v", err)
			}

			if claims.UserID != user.ID {
				t.Errorf("Expected UserID %s, got %s", user.ID, claims.UserID)
			}
		})
	}
}

func TestClaims_ExpiresAt(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	user := &storage.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	token, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Check that ExpiresAt is in the future
	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token ExpiresAt should be in the future")
	}

	// Check that ExpiresAt is approximately tokenExpiry from now
	expectedExpiry := time.Now().Add(manager.tokenExpiry)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Token expiry time is off by %v", diff)
	}
}

func TestGetTokenExpiry(t *testing.T) {
	tests := []struct {
		name          string
		tokenExpiry   int
		refreshExpiry int
	}{
		{
			name:          "1 hour",
			tokenExpiry:   3600,
			refreshExpiry: 7200,
		},
		{
			name:          "24 hours",
			tokenExpiry:   86400,
			refreshExpiry: 172800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJWTManager("test-secret", tt.tokenExpiry, tt.refreshExpiry)

			expiry := manager.GetTokenExpiry()
			if expiry != tt.tokenExpiry {
				t.Errorf("Expected token expiry %d, got %d", tt.tokenExpiry, expiry)
			}
		})
	}
}

func TestNilUser(t *testing.T) {
	manager := NewJWTManager("test-secret", 3600, 7200)

	// This should panic or return an error - testing defensive programming
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when passing nil user, got none")
		}
	}()

	_, _ = manager.GenerateToken(nil)
}
