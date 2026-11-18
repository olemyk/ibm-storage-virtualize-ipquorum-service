package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// RefreshResponse represents a token refresh response
type RefreshResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserValidator is an interface for validating user credentials
type UserValidator interface {
	ValidateUser(username, password string) (*UserInfo, error)
}

// UserInfo represents user information
type UserInfo struct {
	UserID   string
	Username string
	Roles    []string
}

// AuthHandlers provides HTTP handlers for authentication
type AuthHandlers struct {
	jwtManager    *JWTManager
	userValidator UserValidator
	logger        *zap.Logger
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(jwtManager *JWTManager, userValidator UserValidator, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{
		jwtManager:    jwtManager,
		userValidator: userValidator,
		logger:        logger,
	}
}

// Login handles user login
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Validate credentials
	userInfo, err := h.userValidator.ValidateUser(req.Username, req.Password)
	if err != nil {
		h.logger.Warn("Login failed",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(userInfo.UserID, userInfo.Username, userInfo.Roles)
	if err != nil {
		h.logger.Error("Failed to generate token",
			zap.String("user_id", userInfo.UserID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	// Get token info
	tokenInfo, err := h.jwtManager.GetTokenInfo(token)
	if err != nil {
		h.logger.Error("Failed to get token info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get token info",
		})
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("user_id", userInfo.UserID),
		zap.String("username", userInfo.Username),
	)

	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: tokenInfo.ExpiresAt,
		UserID:    userInfo.UserID,
		Username:  userInfo.Username,
		Roles:     userInfo.Roles,
	})
}

// Refresh handles token refresh
func (h *AuthHandlers) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Refresh token
	newToken, err := h.jwtManager.RefreshToken(req.Token)
	if err != nil {
		h.logger.Warn("Token refresh failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to refresh token",
		})
		return
	}

	// Get new token info
	tokenInfo, err := h.jwtManager.GetTokenInfo(newToken)
	if err != nil {
		h.logger.Error("Failed to get token info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get token info",
		})
		return
	}

	h.logger.Info("Token refreshed successfully")

	c.JSON(http.StatusOK, RefreshResponse{
		Token:     newToken,
		ExpiresAt: tokenInfo.ExpiresAt,
	})
}

// Logout handles user logout (optional - mainly for logging)
func (h *AuthHandlers) Logout(c *gin.Context) {
	userID, _ := GetUserID(c)

	h.logger.Info("User logged out",
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// Me returns current user information
func (h *AuthHandlers) Me(c *gin.Context) {
	claims, ok := GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    claims.UserID,
		"username":   claims.Username,
		"roles":      claims.Roles,
		"issuer":     claims.Issuer,
		"issued_at":  claims.IssuedAt.Time,
		"expires_at": claims.ExpiresAt.Time,
	})
}

// SimpleUserValidator is a simple in-memory user validator for testing
type SimpleUserValidator struct {
	users map[string]*UserInfo
}

// NewSimpleUserValidator creates a new simple user validator
func NewSimpleUserValidator() *SimpleUserValidator {
	return &SimpleUserValidator{
		users: make(map[string]*UserInfo),
	}
}

// AddUser adds a user to the validator
func (v *SimpleUserValidator) AddUser(username, password string, userInfo *UserInfo) {
	v.users[username+":"+password] = userInfo
}

// ValidateUser validates user credentials
func (v *SimpleUserValidator) ValidateUser(username, password string) (*UserInfo, error) {
	key := username + ":" + password
	userInfo, ok := v.users[key]
	if !ok {
		return nil, ErrInvalidToken // Reuse error for invalid credentials
	}
	return userInfo, nil
}

// Made with Bob
