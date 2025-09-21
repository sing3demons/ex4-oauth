package services

import (
	"time"

	"ex4-oauth2/internal/auth"
	"ex4-oauth2/internal/models"
)

// IntrospectionService handles token introspection
type IntrospectionService struct {
	accessTokenRepo  models.OAuth2AccessTokenRepository
	refreshTokenRepo models.RefreshTokenRepository
	clientRepo       models.OAuth2ClientRepository
	userRepo         models.UserRepository
	jwtService       *auth.JWTService
}

// NewIntrospectionService creates a new introspection service
func NewIntrospectionService(
	accessTokenRepo models.OAuth2AccessTokenRepository,
	refreshTokenRepo models.RefreshTokenRepository,
	clientRepo models.OAuth2ClientRepository,
	userRepo models.UserRepository,
	jwtService *auth.JWTService,
) *IntrospectionService {
	return &IntrospectionService{
		accessTokenRepo:  accessTokenRepo,
		refreshTokenRepo: refreshTokenRepo,
		clientRepo:       clientRepo,
		userRepo:         userRepo,
		jwtService:       jwtService,
	}
}

// TokenIntrospectionResponse represents the response for token introspection
type TokenIntrospectionResponse struct {
	Active    bool   `json:"active"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Scope     string `json:"scope,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	Issuer    string `json:"iss,omitempty"`
}

// IntrospectToken introspects a token and returns its information
func (s *IntrospectionService) IntrospectToken(token, tokenTypeHint, clientID, clientSecret string) (*TokenIntrospectionResponse, error) {
	// Verify client credentials
	client, err := s.clientRepo.GetByClientID(clientID)
	if err != nil || client.ClientSecret != clientSecret {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	// Try to introspect as access token first
	if tokenTypeHint == "" || tokenTypeHint == "access_token" {
		if response, err := s.introspectAccessToken(token); err == nil && response.Active {
			return response, nil
		}
	}

	// Try to introspect as refresh token
	if tokenTypeHint == "" || tokenTypeHint == "refresh_token" {
		if response, err := s.introspectRefreshToken(token); err == nil && response.Active {
			return response, nil
		}
	}

	// Token not found or invalid
	return &TokenIntrospectionResponse{Active: false}, nil
}

// ValidateToken validates a token and returns its validity and information
func (s *IntrospectionService) ValidateToken(token string) (bool, *TokenIntrospectionResponse, error) {
	// Try JWT token validation first
	claims, err := s.jwtService.ValidateAccessToken(token)
	if err == nil {
		// Get user information
		user, err := s.userRepo.GetByID(claims.UserID)
		if err != nil {
			return false, nil, err
		}

		response := &TokenIntrospectionResponse{
			Active:    true,
			Username:  user.Username,
			UserID:    claims.UserID,
			TokenType: "Bearer",
			ExpiresAt: claims.ExpiresAt.Time.Unix(),
			IssuedAt:  claims.IssuedAt.Time.Unix(),
			Subject:   claims.UserID,
		}

		return true, response, nil
	}

	// Try database lookup for access tokens
	accessToken, err := s.accessTokenRepo.GetByToken(token)
	if err == nil && accessToken.ExpiresAt.After(time.Now()) {
		user, err := s.userRepo.GetByID(accessToken.UserID)
		if err != nil {
			return false, nil, err
		}

		response := &TokenIntrospectionResponse{
			Active:    true,
			ClientID:  accessToken.ClientID,
			Username:  user.Username,
			UserID:    accessToken.UserID,
			Scope:     accessToken.Scope,
			TokenType: "Bearer",
			ExpiresAt: accessToken.ExpiresAt.Unix(),
			IssuedAt:  accessToken.CreatedAt.Unix(),
			Subject:   accessToken.UserID,
		}

		return true, response, nil
	}

	return false, nil, nil
}

// introspectAccessToken introspects an access token
func (s *IntrospectionService) introspectAccessToken(token string) (*TokenIntrospectionResponse, error) {
	// Try JWT validation first
	claims, err := s.jwtService.ValidateAccessToken(token)
	if err == nil {
		user, err := s.userRepo.GetByID(claims.UserID)
		if err != nil {
			return &TokenIntrospectionResponse{Active: false}, nil
		}

		return &TokenIntrospectionResponse{
			Active:    true,
			Username:  user.Username,
			UserID:    claims.UserID,
			TokenType: "Bearer",
			ExpiresAt: claims.ExpiresAt.Time.Unix(),
			IssuedAt:  claims.IssuedAt.Time.Unix(),
			Subject:   claims.UserID,
		}, nil
	}

	// Try database lookup
	accessToken, err := s.accessTokenRepo.GetByToken(token)
	if err != nil {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	// Check if token is expired
	if accessToken.ExpiresAt.Before(time.Now()) {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	// Get user information
	user, err := s.userRepo.GetByID(accessToken.UserID)
	if err != nil {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	return &TokenIntrospectionResponse{
		Active:    true,
		ClientID:  accessToken.ClientID,
		Username:  user.Username,
		UserID:    accessToken.UserID,
		Scope:     accessToken.Scope,
		TokenType: "Bearer",
		ExpiresAt: accessToken.ExpiresAt.Unix(),
		IssuedAt:  accessToken.CreatedAt.Unix(),
		Subject:   accessToken.UserID,
	}, nil
}

// introspectRefreshToken introspects a refresh token
func (s *IntrospectionService) introspectRefreshToken(token string) (*TokenIntrospectionResponse, error) {
	// Check if refresh token exists and is valid
	refreshToken, err := s.refreshTokenRepo.GetByToken(token)
	if err != nil {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	// Check if token is expired
	if refreshToken.ExpiresAt.Before(time.Now()) {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	// Get user information
	user, err := s.userRepo.GetByID(refreshToken.UserID)
	if err != nil {
		return &TokenIntrospectionResponse{Active: false}, nil
	}

	return &TokenIntrospectionResponse{
		Active:    true,
		Username:  user.Username,
		UserID:    refreshToken.UserID,
		TokenType: "refresh_token",
		ExpiresAt: refreshToken.ExpiresAt.Unix(),
		IssuedAt:  refreshToken.CreatedAt.Unix(),
		Subject:   refreshToken.UserID,
	}, nil
}
