package services

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"ex4-oauth2/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// IntrospectionService handles token introspection operations
type IntrospectionService struct {
	userRepo         models.UserRepository
	refreshTokenRepo models.RefreshTokenRepository
	auditService     *AuditService
}

// NewIntrospectionService creates a new introspection service
func NewIntrospectionService(userRepo models.UserRepository, refreshTokenRepo models.RefreshTokenRepository, auditService *AuditService) *IntrospectionService {
	return &IntrospectionService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		auditService:     auditService,
	}
}

// IntrospectToken validates and returns metadata for a given token
func (s *IntrospectionService) IntrospectToken(tokenString string, tokenTypeHint string, clientIP string) (*models.TokenIntrospectionResponse, error) {
	// Log introspection attempt
	if s.auditService != nil {
		ctx := &AuditContext{
			IPAddress: clientIP,
		}
		s.auditService.LogAction(ctx, "token_introspection", "token", "attempted", map[string]interface{}{
			"token_type_hint": tokenTypeHint,
			"token_length":    len(tokenString),
		}, nil)
	}

	// Try to parse as JWT first
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	// If JWT parsing fails, check if it's a refresh token
	if err != nil || !token.Valid {
		return s.introspectRefreshToken(tokenString, clientIP)
	}

	// Handle JWT access token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &models.TokenIntrospectionResponse{Active: false}, nil
	}

	// Check if token is expired
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return &models.TokenIntrospectionResponse{Active: false}, nil
		}
	}

	// Extract user information
	userID, _ := claims["user_id"].(float64)
	user, err := s.userRepo.GetByID(uint(userID))
	if err != nil || user == nil || !user.IsActive {
		return &models.TokenIntrospectionResponse{Active: false}, nil
	}

	// Build successful response
	response := &models.TokenIntrospectionResponse{
		Active:    true,
		Scope:     s.extractScope(claims),
		ClientID:  s.extractClientID(claims),
		Username:  user.Username,
		TokenType: "Bearer",
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		Provider:  user.Provider,
	}

	// Add standard JWT claims
	if exp, ok := claims["exp"].(float64); ok {
		response.Exp = int64(exp)
	}
	if iat, ok := claims["iat"].(float64); ok {
		response.Iat = int64(iat)
	}
	if nbf, ok := claims["nbf"].(float64); ok {
		response.Nbf = int64(nbf)
	}
	if sub, ok := claims["sub"].(string); ok {
		response.Sub = sub
	}
	if iss, ok := claims["iss"].(string); ok {
		response.Iss = iss
	}
	if jti, ok := claims["jti"].(string); ok {
		response.Jti = jti
	}
	if aud, ok := claims["aud"].([]interface{}); ok {
		audSlice := make([]string, len(aud))
		for i, v := range aud {
			if str, ok := v.(string); ok {
				audSlice[i] = str
			}
		}
		response.Aud = audSlice
	}

	// Log successful introspection
	if s.auditService != nil {
		ctx := &AuditContext{
			UserID:    &user.ID,
			IPAddress: clientIP,
		}
		s.auditService.LogAction(ctx, "token_introspection", "token", "success", map[string]interface{}{
			"token_type": "access_token",
			"expires_at": response.Exp,
		}, nil)
	}

	return response, nil
}

// introspectRefreshToken handles refresh token introspection
func (s *IntrospectionService) introspectRefreshToken(tokenString string, clientIP string) (*models.TokenIntrospectionResponse, error) {
	// Look up refresh token in database
	refreshToken, err := s.refreshTokenRepo.GetByToken(tokenString)
	if err != nil {
		// Log failed introspection
		if s.auditService != nil {
			ctx := &AuditContext{
				IPAddress: clientIP,
			}
			s.auditService.LogFailedAction(ctx, "token_introspection", "token", "", "refresh_token_not_found")
		}
		return &models.TokenIntrospectionResponse{Active: false}, nil
	}

	// Check if refresh token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		if s.auditService != nil {
			ctx := &AuditContext{
				UserID:    &refreshToken.UserID,
				IPAddress: clientIP,
			}
			s.auditService.LogFailedAction(ctx, "token_introspection", "token", "", "refresh_token_expired")
		}
		return &models.TokenIntrospectionResponse{Active: false}, nil
	}

	// Check if user is still active
	user := &refreshToken.User
	if !user.IsActive {
		if s.auditService != nil {
			ctx := &AuditContext{
				UserID:    &user.ID,
				IPAddress: clientIP,
			}
			s.auditService.LogFailedAction(ctx, "token_introspection", "token", "", "user_inactive")
		}
		return &models.TokenIntrospectionResponse{Active: false}, nil
	}

	// Build successful response for refresh token
	response := &models.TokenIntrospectionResponse{
		Active:    true,
		Username:  user.Username,
		TokenType: "refresh_token",
		Exp:       refreshToken.ExpiresAt.Unix(),
		Iat:       refreshToken.CreatedAt.Unix(),
		Sub:       strconv.Itoa(int(user.ID)),
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		Provider:  user.Provider,
	}

	// Log successful refresh token introspection
	if s.auditService != nil {
		ctx := &AuditContext{
			UserID:    &user.ID,
			IPAddress: clientIP,
		}
		s.auditService.LogAction(ctx, "token_introspection", "token", "success", map[string]interface{}{
			"token_type": "refresh_token",
			"expires_at": response.Exp,
		}, nil)
	}

	return response, nil
}

// extractScope extracts scope from JWT claims
func (s *IntrospectionService) extractScope(claims jwt.MapClaims) string {
	if scope, ok := claims["scope"].(string); ok {
		return scope
	}
	if scopes, ok := claims["scopes"].([]interface{}); ok {
		scopeSlice := make([]string, len(scopes))
		for i, v := range scopes {
			if str, ok := v.(string); ok {
				scopeSlice[i] = str
			}
		}
		return strings.Join(scopeSlice, " ")
	}
	return ""
}

// extractClientID extracts client ID from JWT claims
func (s *IntrospectionService) extractClientID(claims jwt.MapClaims) string {
	if clientID, ok := claims["client_id"].(string); ok {
		return clientID
	}
	if aud, ok := claims["aud"].(string); ok {
		return aud
	}
	return ""
}

// ValidateIntrospectionClient validates if the requesting client is authorized for introspection
func (s *IntrospectionService) ValidateIntrospectionClient(clientID, clientSecret string) bool {
	// For now, we'll use a simple validation
	// In production, you should have a proper client registry
	expectedClientID := os.Getenv("INTROSPECTION_CLIENT_ID")
	expectedClientSecret := os.Getenv("INTROSPECTION_CLIENT_SECRET")

	if expectedClientID == "" || expectedClientSecret == "" {
		// If not configured, allow any client for development
		return true
	}

	return clientID == expectedClientID && clientSecret == expectedClientSecret
}

// GetTokenUsageStats returns statistics about token usage
func (s *IntrospectionService) GetTokenUsageStats(userID uint, days int) (map[string]interface{}, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	stats := map[string]interface{}{
		"user_id":    userID,
		"period":     fmt.Sprintf("%d days", days),
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   time.Now().Format("2006-01-02"),
	}

	// You can extend this to query actual usage from audit logs
	if s.auditService != nil {
		// This would require additional audit queries
		stats["introspection_count"] = 0
		stats["last_activity"] = time.Now().Format("2006-01-02 15:04:05")
	}

	return stats, nil
}

// RevokeToken revokes a token (marks it as inactive)
func (s *IntrospectionService) RevokeToken(tokenString string, clientIP string) error {
	// For JWT tokens, we would need a blacklist/revocation list
	// For refresh tokens, we can delete them from the database

	// Try to find and delete refresh token
	refreshToken, err := s.refreshTokenRepo.GetByToken(tokenString)
	if err == nil && refreshToken != nil {
		err = s.refreshTokenRepo.Delete(tokenString)
		if err != nil {
			return err
		}

		// Log token revocation
		if s.auditService != nil {
			ctx := &AuditContext{
				UserID:    &refreshToken.UserID,
				IPAddress: clientIP,
			}
			s.auditService.LogAction(ctx, "token_revocation", "token", "success", map[string]interface{}{
				"token_type": "refresh_token",
			}, nil)
		}
		return nil
	}

	// For JWT access tokens, you would implement a blacklist here
	// This is a simplified implementation
	if s.auditService != nil {
		ctx := &AuditContext{
			IPAddress: clientIP,
		}
		s.auditService.LogAction(ctx, "token_revocation", "token", "attempted", map[string]interface{}{
			"token_type": "access_token",
			"note":       "JWT revocation requires blacklist implementation",
		}, nil)
	}

	return fmt.Errorf("token revocation not fully implemented for JWT tokens")
}
