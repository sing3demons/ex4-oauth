package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
)

// OAuth2Client represents an OAuth2 client application
type OAuth2Client struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	ClientID     string    `json:"client_id" bson:"client_id"`
	ClientSecret string    `json:"-" bson:"client_secret"` // Never expose in JSON
	Name         string    `json:"name" bson:"name"`
	Description  string    `json:"description" bson:"description"`
	RedirectURIs []string  `json:"redirect_uris" bson:"redirect_uris"`
	Scopes       []string  `json:"scopes" bson:"scopes"`
	GrantTypes   []string  `json:"grant_types" bson:"grant_types"` // authorization_code, refresh_token
	IsPublic     bool      `json:"is_public" bson:"is_public"`     // Public clients (no client secret required)
	IsActive     bool      `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}

// OAuth2AuthorizationCode represents an authorization code
type OAuth2AuthorizationCode struct {
	ID              string    `json:"id" bson:"_id,omitempty"`
	Code            string    `json:"code" bson:"code"`
	ClientID        string    `json:"client_id" bson:"client_id"`
	UserID          string    `json:"user_id" bson:"user_id"`
	RedirectURI     string    `json:"redirect_uri" bson:"redirect_uri"`
	Scope           string    `json:"scope" bson:"scope"`
	CodeChallenge   string    `json:"code_challenge" bson:"code_challenge"`     // For PKCE
	ChallengeMethod string    `json:"challenge_method" bson:"challenge_method"` // S256 or plain
	Nonce           string    `json:"nonce" bson:"nonce"`                       // For OIDC
	ExpiresAt       time.Time `json:"expires_at" bson:"expires_at"`
	Used            bool      `json:"used" bson:"used"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	User            *User     `json:"user,omitempty" bson:"-"` // Populated during queries if needed
}

// OAuth2AccessToken represents an access token
type OAuth2AccessToken struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Token     string    `json:"token" bson:"token"`
	ClientID  string    `json:"client_id" bson:"client_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Scope     string    `json:"scope" bson:"scope"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	User      *User     `json:"user,omitempty" bson:"-"` // Populated during queries if needed
}

// SetDefaults sets default values for OAuth2Client
func (c *OAuth2Client) SetDefaults() {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.ClientID == "" {
		c.ClientID = generateRandomString(32)
	}
	// Only generate client secret for confidential clients
	if !c.IsPublic && c.ClientSecret == "" {
		c.ClientSecret = generateRandomString(64)
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()
}

// OAuth2ClientRepository defines the interface for OAuth2 client data access
type OAuth2ClientRepository interface {
	Create(client *OAuth2Client) error
	GetByID(id string) (*OAuth2Client, error)
	GetByClientID(clientID string) (*OAuth2Client, error)
	Update(client *OAuth2Client) error
	Delete(id string) error
	List(limit, offset int) ([]*OAuth2Client, error)
	ValidateClientCredentials(clientID, clientSecret string) (*OAuth2Client, error)
	ValidatePublicClient(clientID string) (*OAuth2Client, error)
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
	GetByUserAndClient(userID string, clientID string) ([]*OAuth2AccessToken, error)
	Delete(token string) error
	DeleteByUserAndClient(userID string, clientID string) error
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
