package database

import (
	"encoding/json"
	"errors"
	"time"

	"ex4-oauth2/internal/models"

	"gorm.io/gorm"
)

// OAuth2ClientRepository implements the OAuth2ClientRepository interface
type OAuth2ClientRepository struct {
	db *gorm.DB
}

// NewOAuth2ClientRepository creates a new OAuth2 client repository
func NewOAuth2ClientRepository(db *gorm.DB) models.OAuth2ClientRepository {
	return &OAuth2ClientRepository{db: db}
}

// Create creates a new OAuth2 client
func (r *OAuth2ClientRepository) Create(client *models.OAuth2Client) error {
	// Convert slices to JSON strings
	if len(client.RedirectURIs) > 0 {
		data, _ := json.Marshal(client.RedirectURIs)
		client.RedirectURIsStr = string(data)
	}
	if len(client.Scopes) > 0 {
		data, _ := json.Marshal(client.Scopes)
		client.ScopesStr = string(data)
	}
	if len(client.GrantTypes) > 0 {
		data, _ := json.Marshal(client.GrantTypes)
		client.GrantTypesStr = string(data)
	}

	return r.db.Create(client).Error
}

// GetByID retrieves an OAuth2 client by ID
func (r *OAuth2ClientRepository) GetByID(id uint) (*models.OAuth2Client, error) {
	var client models.OAuth2Client
	err := r.db.First(&client, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	r.populateSlices(&client)
	return &client, nil
}

// GetByClientID retrieves an OAuth2 client by client ID
func (r *OAuth2ClientRepository) GetByClientID(clientID string) (*models.OAuth2Client, error) {
	var client models.OAuth2Client
	err := r.db.Where("client_id = ?", clientID).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	r.populateSlices(&client)
	return &client, nil
}

// Update updates an OAuth2 client
func (r *OAuth2ClientRepository) Update(client *models.OAuth2Client) error {
	// Convert slices to JSON strings
	if len(client.RedirectURIs) > 0 {
		data, _ := json.Marshal(client.RedirectURIs)
		client.RedirectURIsStr = string(data)
	}
	if len(client.Scopes) > 0 {
		data, _ := json.Marshal(client.Scopes)
		client.ScopesStr = string(data)
	}
	if len(client.GrantTypes) > 0 {
		data, _ := json.Marshal(client.GrantTypes)
		client.GrantTypesStr = string(data)
	}

	return r.db.Save(client).Error
}

// Delete soft deletes an OAuth2 client
func (r *OAuth2ClientRepository) Delete(id uint) error {
	return r.db.Delete(&models.OAuth2Client{}, id).Error
}

// List retrieves OAuth2 clients with pagination
func (r *OAuth2ClientRepository) List(limit, offset int) ([]*models.OAuth2Client, error) {
	var clients []*models.OAuth2Client
	err := r.db.Limit(limit).Offset(offset).Find(&clients).Error
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		r.populateSlices(client)
	}
	return clients, nil
}

// ValidateClientCredentials validates client ID and secret
func (r *OAuth2ClientRepository) ValidateClientCredentials(clientID, clientSecret string) (*models.OAuth2Client, error) {
	var client models.OAuth2Client
	err := r.db.Where("client_id = ? AND client_secret = ? AND is_active = ?", clientID, clientSecret, true).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid client credentials")
		}
		return nil, err
	}
	r.populateSlices(&client)
	return &client, nil
}

// populateSlices converts JSON strings back to slices
func (r *OAuth2ClientRepository) populateSlices(client *models.OAuth2Client) {
	if client.RedirectURIsStr != "" {
		json.Unmarshal([]byte(client.RedirectURIsStr), &client.RedirectURIs)
	}
	if client.ScopesStr != "" {
		json.Unmarshal([]byte(client.ScopesStr), &client.Scopes)
	}
	if client.GrantTypesStr != "" {
		json.Unmarshal([]byte(client.GrantTypesStr), &client.GrantTypes)
	}
}

// OAuth2AuthorizationCodeRepository implements OAuth2AuthorizationCodeRepository
type OAuth2AuthorizationCodeRepository struct {
	db *gorm.DB
}

// NewOAuth2AuthorizationCodeRepository creates a new authorization code repository
func NewOAuth2AuthorizationCodeRepository(db *gorm.DB) models.OAuth2AuthorizationCodeRepository {
	return &OAuth2AuthorizationCodeRepository{db: db}
}

// Create creates a new authorization code
func (r *OAuth2AuthorizationCodeRepository) Create(code *models.OAuth2AuthorizationCode) error {
	return r.db.Create(code).Error
}

// GetByCode retrieves authorization code by code string
func (r *OAuth2AuthorizationCodeRepository) GetByCode(code string) (*models.OAuth2AuthorizationCode, error) {
	var authCode models.OAuth2AuthorizationCode
	err := r.db.Preload("User").Where("code = ? AND used = ?", code, false).First(&authCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("authorization code not found")
		}
		return nil, err
	}

	// Check if expired
	if authCode.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("authorization code expired")
	}

	return &authCode, nil
}

// MarkAsUsed marks authorization code as used
func (r *OAuth2AuthorizationCodeRepository) MarkAsUsed(code string) error {
	return r.db.Model(&models.OAuth2AuthorizationCode{}).Where("code = ?", code).Update("used", true).Error
}

// CleanupExpiredCodes removes expired authorization codes
func (r *OAuth2AuthorizationCodeRepository) CleanupExpiredCodes() error {
	return r.db.Where("expires_at < ? OR used = ?", time.Now(), true).Delete(&models.OAuth2AuthorizationCode{}).Error
}

// OAuth2AccessTokenRepository implements OAuth2AccessTokenRepository
type OAuth2AccessTokenRepository struct {
	db *gorm.DB
}

// NewOAuth2AccessTokenRepository creates a new access token repository
func NewOAuth2AccessTokenRepository(db *gorm.DB) models.OAuth2AccessTokenRepository {
	return &OAuth2AccessTokenRepository{db: db}
}

// Create creates a new access token
func (r *OAuth2AccessTokenRepository) Create(token *models.OAuth2AccessToken) error {
	return r.db.Create(token).Error
}

// GetByToken retrieves access token by token string
func (r *OAuth2AccessTokenRepository) GetByToken(token string) (*models.OAuth2AccessToken, error) {
	var accessToken models.OAuth2AccessToken
	err := r.db.Preload("User").Where("token = ?", token).First(&accessToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("access token not found")
		}
		return nil, err
	}

	// Check if expired
	if accessToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("access token expired")
	}

	return &accessToken, nil
}

// GetByUserAndClient retrieves access tokens by user and client
func (r *OAuth2AccessTokenRepository) GetByUserAndClient(userID uint, clientID string) ([]*models.OAuth2AccessToken, error) {
	var tokens []*models.OAuth2AccessToken
	err := r.db.Where("user_id = ? AND client_id = ?", userID, clientID).Find(&tokens).Error
	return tokens, err
}

// Delete deletes access token by token string
func (r *OAuth2AccessTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.OAuth2AccessToken{}).Error
}

// DeleteByUserAndClient deletes all access tokens for a user and client
func (r *OAuth2AccessTokenRepository) DeleteByUserAndClient(userID uint, clientID string) error {
	return r.db.Where("user_id = ? AND client_id = ?", userID, clientID).Delete(&models.OAuth2AccessToken{}).Error
}

// CleanupExpiredTokens removes expired access tokens
func (r *OAuth2AccessTokenRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.OAuth2AccessToken{}).Error
}
