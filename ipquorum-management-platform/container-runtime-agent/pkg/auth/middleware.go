package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// AuthorizationHeader is the HTTP header for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// ClaimsContextKey is the context key for storing claims
	ClaimsContextKey = "jwt_claims"
	// UserIDContextKey is the context key for storing user ID
	UserIDContextKey = "user_id"
)

// JWTMiddleware creates a Gin middleware for JWT authentication
func JWTMiddleware(jwtManager *JWTManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			logger.Warn("Missing authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			logger.Warn("Invalid authorization header format",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format, expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, BearerPrefix)
		if token == "" {
			logger.Warn("Empty token in authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "empty token",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.Warn("Token validation failed",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
			)

			// Provide specific error messages
			var message string
			switch err {
			case ErrExpiredToken:
				message = "token has expired"
			case ErrInvalidToken:
				message = "invalid token"
			case ErrInvalidClaims:
				message = "invalid token claims"
			default:
				message = "token validation failed"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": message,
			})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set(ClaimsContextKey, claims)
		c.Set(UserIDContextKey, claims.UserID)

		logger.Debug("JWT authentication successful",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.Strings("roles", claims.Roles),
		)

		c.Next()
	}
}

// OptionalJWTMiddleware creates a middleware that validates JWT if present but doesn't require it
func OptionalJWTMiddleware(jwtManager *JWTManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Token provided, validate it
		if strings.HasPrefix(authHeader, BearerPrefix) {
			token := strings.TrimPrefix(authHeader, BearerPrefix)
			if token != "" {
				claims, err := jwtManager.ValidateToken(token)
				if err == nil {
					// Valid token, store claims
					c.Set(ClaimsContextKey, claims)
					c.Set(UserIDContextKey, claims.UserID)
					logger.Debug("Optional JWT authentication successful",
						zap.String("user_id", claims.UserID),
					)
				} else {
					logger.Debug("Optional JWT validation failed, continuing without auth",
						zap.Error(err),
					)
				}
			}
		}

		c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsValue, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		claims, ok := claimsValue.(*Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims type",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		if !claims.HasAnyRole(roles...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "insufficient permissions",
				"required_roles": roles,
				"user_roles":     claims.Roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates a middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// GetClaims extracts JWT claims from the Gin context
func GetClaims(c *gin.Context) (*Claims, bool) {
	claimsValue, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil, false
	}

	claims, ok := claimsValue.(*Claims)
	return claims, ok
}

// GetUserID extracts user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userIDValue, exists := c.Get(UserIDContextKey)
	if !exists {
		return "", false
	}

	userID, ok := userIDValue.(string)
	return userID, ok
}

// MustGetClaims extracts JWT claims from context or panics
func MustGetClaims(c *gin.Context) *Claims {
	claims, ok := GetClaims(c)
	if !ok {
		panic("claims not found in context")
	}
	return claims
}

// MustGetUserID extracts user ID from context or panics
func MustGetUserID(c *gin.Context) string {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user ID not found in context")
	}
	return userID
}

// Made with Bob
