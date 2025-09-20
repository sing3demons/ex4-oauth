package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"gorm.io/gorm"
)

// OAuth2Client represents an OAuth2 client application
type OAuth2Client struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ClientID        string         `json:"client_id" gorm:"unique;not null"`
	ClientSecret    string         `json:"-" gorm:"not null"` // Never expose in JSON
	Name            string         `json:"name" gorm:"not null"`
	Description     string         `json:"description"`
	RedirectURIs    []string       `json:"redirect_uris" gorm:"-"`        // Will be stored as string and converted
	RedirectURIsStr string         `json:"-" gorm:"column:redirect_uris"` // JSON string in DB
	Scopes          []string       `json:"scopes" gorm:"-"`               // Will be stored as string and converted
	ScopesStr       string         `json:"-" gorm:"column:scopes"`        // JSON string in DB
	GrantTypes      []string       `json:"grant_types" gorm:"-"`          // authorization_code, refresh_token
	GrantTypesStr   string         `json:"-" gorm:"column:grant_types"`   // JSON string in DB
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// OAuth2AuthorizationCode represents an authorization code
type OAuth2AuthorizationCode struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Code            string    `json:"code" gorm:"unique;not null"`
	ClientID        string    `json:"client_id" gorm:"not null"`
	UserID          uint      `json:"user_id" gorm:"not null"`
	RedirectURI     string    `json:"redirect_uri" gorm:"not null"`
	Scope           string    `json:"scope"`
	CodeChallenge   string    `json:"code_challenge"`   // For PKCE
	ChallengeMethod string    `json:"challenge_method"` // S256 or plain
	Nonce           string    `json:"nonce"`            // For OIDC
	ExpiresAt       time.Time `json:"expires_at" gorm:"not null"`
	Used            bool      `json:"used" gorm:"default:false"`
	CreatedAt       time.Time `json:"created_at"`
	User            User      `json:"user" gorm:"foreignKey:UserID"`
}

// OAuth2AccessToken represents an access token
type OAuth2AccessToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Token     string         `json:"token" gorm:"unique;not null"`
	ClientID  string         `json:"client_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Scope     string         `json:"scope"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
}

// BeforeCreate generates client ID and secret
func (c *OAuth2Client) BeforeCreate(tx *gorm.DB) error {
	if c.ClientID == "" {
		c.ClientID = generateRandomString(32)
	}
	if c.ClientSecret == "" {
		c.ClientSecret = generateRandomString(64)
	}
	return nil
}

// OAuth2ClientRepository defines the interface for OAuth2 client data access
type OAuth2ClientRepository interface {
	Create(client *OAuth2Client) error
	GetByID(id uint) (*OAuth2Client, error)
	GetByClientID(clientID string) (*OAuth2Client, error)
	Update(client *OAuth2Client) error
	Delete(id uint) error
	List(limit, offset int) ([]*OAuth2Client, error)
	ValidateClientCredentials(clientID, clientSecret string) (*OAuth2Client, error)
}

// OAuth2AuthorizationCodeRepository defines the interface for authorization code management
type OAuth2AuthorizationCodeRepository interface {
	Create(code *OAuth2AuthorizationCode) error
	GetByCode(code string) (*OAuth2AuthorizationCode, error)
	MarkAsUsed(code string) error
	CleanupExpiredCodes() error
}

// OAuth2AccessTokenRepository defines the interface for access token management
type OAuth2AccessTokenRepository interface {
	Create(token *OAuth2AccessToken) error
	GetByToken(token string) (*OAuth2AccessToken, error)
	GetByUserAndClient(userID uint, clientID string) ([]*OAuth2AccessToken, error)
	Delete(token string) error
	DeleteByUserAndClient(userID uint, clientID string) error
	CleanupExpiredTokens() error
}

// GenerateAuthorizationCode generates a new authorization code
func GenerateAuthorizationCode() string {
	return generateRandomString(32)
}

// GenerateAccessToken generates a new access token
func GenerateAccessToken() string {
	return generateRandomString(64)
}

// Helper function to generate random string
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
