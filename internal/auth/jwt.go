package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"ex4-oauth2/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// IDTokenClaims represents OIDC ID Token claims
type IDTokenClaims struct {
	// Standard OIDC claims
	Subject   string `json:"sub"`                 // Subject identifier
	Audience  string `json:"aud"`                 // Client ID
	Issuer    string `json:"iss"`                 // Issuer identifier
	IssuedAt  int64  `json:"iat"`                 // Issued at time
	ExpiresAt int64  `json:"exp"`                 // Expiration time
	AuthTime  int64  `json:"auth_time,omitempty"` // Authentication time
	Nonce     string `json:"nonce,omitempty"`     // Nonce from authorization request

	// Profile claims
	Name              string `json:"name,omitempty"`               // Full name
	GivenName         string `json:"given_name,omitempty"`         // First name
	FamilyName        string `json:"family_name,omitempty"`        // Last name
	MiddleName        string `json:"middle_name,omitempty"`        // Middle name
	Nickname          string `json:"nickname,omitempty"`           // Nickname
	PreferredUsername string `json:"preferred_username,omitempty"` // Username
	Picture           string `json:"picture,omitempty"`            // Profile picture URL
	Website           string `json:"website,omitempty"`            // Website URL
	Gender            string `json:"gender,omitempty"`             // Gender
	Birthdate         string `json:"birthdate,omitempty"`          // Birthdate
	Zoneinfo          string `json:"zoneinfo,omitempty"`           // Timezone
	Locale            string `json:"locale,omitempty"`             // Locale
	UpdatedAt         int64  `json:"updated_at,omitempty"`         // Last update time

	// Email claims
	Email         string `json:"email,omitempty"`          // Email address
	EmailVerified bool   `json:"email_verified,omitempty"` // Email verification status

	// Phone claims
	PhoneNumber   string `json:"phone_number,omitempty"`          // Phone number
	PhoneVerified bool   `json:"phone_number_verified,omitempty"` // Phone verification status

	// Address claim
	Address *AddressClaim `json:"address,omitempty"` // Address information
}

// AddressClaim represents OIDC address claim
type AddressClaim struct {
	Formatted     string `json:"formatted,omitempty"`      // Full formatted address
	StreetAddress string `json:"street_address,omitempty"` // Street address
	Locality      string `json:"locality,omitempty"`       // City/locality
	Region        string `json:"region,omitempty"`         // State/region
	PostalCode    string `json:"postal_code,omitempty"`    // Postal/ZIP code
	Country       string `json:"country,omitempty"`        // Country
}

// JWT Claims interface implementation for IDTokenClaims
func (c IDTokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

func (c IDTokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

func (c IDTokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

func (c IDTokenClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

func (c IDTokenClaims) GetSubject() (string, error) {
	return c.Subject, nil
}

func (c IDTokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{c.Audience}, nil
}

// OIDCTokenPair represents OIDC token response with ID token
type OIDCTokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// PKCEChallenge represents PKCE challenge and verifier
type PKCEChallenge struct {
	CodeVerifier  string `json:"code_verifier"`
	CodeChallenge string `json:"code_challenge"`
	Method        string `json:"method"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType   string `json:"kty"`                // Key Type
	Use       string `json:"use,omitempty"`      // Public Key Use
	KeyOps    string `json:"key_ops,omitempty"`  // Key Operations
	Algorithm string `json:"alg,omitempty"`      // Algorithm
	KeyID     string `json:"kid,omitempty"`      // Key ID
	X5U       string `json:"x5u,omitempty"`      // X.509 URL
	X5C       string `json:"x5c,omitempty"`      // X.509 Certificate Chain
	X5T       string `json:"x5t,omitempty"`      // X.509 Certificate SHA-1 Thumbprint
	X5TS256   string `json:"x5t#S256,omitempty"` // X.509 Certificate SHA-256 Thumbprint
	// RSA specific fields
	Modulus  string `json:"n,omitempty"` // RSA modulus
	Exponent string `json:"e,omitempty"` // RSA exponent
}

// JWKSet represents a set of JSON Web Keys
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// RSAKeyPair holds RSA private and public keys
type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
}

// JWTService handles JWT operations
type JWTService struct {
	secretKey        []byte
	rsaKeyPair       *RSAKeyPair
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	refreshTokenRepo models.RefreshTokenRepository
	useRSA           bool
}

// generateRSAKeyPair generates a new RSA key pair for JWT signing
func generateRSAKeyPair() (*RSAKeyPair, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Generate key ID
	keyIDBytes := make([]byte, 16)
	_, err = rand.Read(keyIDBytes)
	if err != nil {
		return nil, err
	}
	keyID := base64.URLEncoding.EncodeToString(keyIDBytes)

	return &RSAKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      keyID,
	}, nil
}

// GetJWKSet returns the JWK Set for this service
func (j *JWTService) GetJWKSet() (*JWKSet, error) {
	if !j.useRSA || j.rsaKeyPair == nil {
		return nil, errors.New("RSA keys not available")
	}

	// Convert RSA public key to JWK
	jwk, err := j.rsaPublicKeyToJWK(j.rsaKeyPair.PublicKey, j.rsaKeyPair.KeyID)
	if err != nil {
		return nil, err
	}

	return &JWKSet{
		Keys: []JWK{*jwk},
	}, nil
}

// rsaPublicKeyToJWK converts RSA public key to JWK format
func (j *JWTService) rsaPublicKeyToJWK(publicKey *rsa.PublicKey, keyID string) (*JWK, error) {
	// Convert modulus to base64url
	nBytes := publicKey.N.Bytes()
	n := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(nBytes)

	// Convert exponent to base64url
	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	e := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(eBytes)

	return &JWK{
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: "RS256",
		KeyID:     keyID,
		Modulus:   n,
		Exponent:  e,
	}, nil
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, refreshTokenRepo models.RefreshTokenRepository) *JWTService {
	service := &JWTService{
		secretKey:        []byte(secretKey),
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		refreshTokenRepo: refreshTokenRepo,
		useRSA:           false, // Default to HMAC for backward compatibility
	}

	// Try to generate RSA key pair for JWK support
	keyPair, err := generateRSAKeyPair()
	if err == nil {
		service.rsaKeyPair = keyPair
		service.useRSA = true
	}

	return service
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Generate access token
	accessToken, err := j.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := j.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessTokenTTL.Seconds()),
	}, nil
}

// GenerateAccessToken generates a JWT access token
func (j *JWTService) GenerateAccessToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ex4-oauth2",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	var token *jwt.Token
	if j.useRSA && j.rsaKeyPair != nil {
		// Use RSA256 for OIDC compliance
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = j.rsaKeyPair.KeyID
		return token.SignedString(j.rsaKeyPair.PrivateKey)
	}

	// Fallback to HMAC256
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateIDToken generates an OIDC ID Token
func (j *JWTService) GenerateIDToken(user *models.User, clientID, nonce string, authTime time.Time) (string, error) {
	now := time.Now()

	claims := IDTokenClaims{
		Subject:   fmt.Sprintf("%d", user.ID),
		Audience:  clientID,
		Issuer:    "ex4-oauth2",
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(j.accessTokenTTL).Unix(),
		AuthTime:  authTime.Unix(),
		Nonce:     nonce,

		// Profile information
		Name:              user.FirstName + " " + user.LastName,
		GivenName:         user.FirstName,
		FamilyName:        user.LastName,
		PreferredUsername: user.Username,
		Picture:           user.Avatar,
		UpdatedAt:         user.UpdatedAt.Unix(),

		// Email information
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}

	var token *jwt.Token
	if j.useRSA && j.rsaKeyPair != nil {
		// Use RSA256 for OIDC compliance
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = j.rsaKeyPair.KeyID
		return token.SignedString(j.rsaKeyPair.PrivateKey)
	}

	// Fallback to HMAC256
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateOIDCTokenPair generates access token, refresh token, and ID token
func (j *JWTService) GenerateOIDCTokenPair(user *models.User, clientID, nonce string, authTime time.Time, scope string) (*OIDCTokenPair, error) {
	// Generate access token
	accessToken, err := j.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := j.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Generate ID token if openid scope is present
	var idToken string
	if scope != "" && contains(scope, "openid") {
		idToken, err = j.GenerateIDToken(user, clientID, nonce, authTime)
		if err != nil {
			return nil, err
		}
	}

	return &OIDCTokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessTokenTTL.Seconds()),
		Scope:        scope,
	}, nil
}

// contains checks if a string contains a substring (helper function)
func contains(s, substr string) bool {
	return s == substr ||
		strings.HasPrefix(s, substr+" ") ||
		strings.HasSuffix(s, " "+substr) ||
		strings.Contains(s, " "+substr+" ")
}

// GenerateRefreshToken generates a refresh token and stores it in database
func (j *JWTService) GenerateRefreshToken(userID string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store in database
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(j.refreshTokenTTL),
	}

	err = j.refreshTokenRepo.Create(refreshToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateAccessToken validates and parses JWT access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			if j.useRSA && j.rsaKeyPair != nil {
				return j.rsaKeyPair.PublicKey, nil
			}
			return nil, fmt.Errorf("RSA key not available")
		case *jwt.SigningMethodHMAC:
			return j.secretKey, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshAccessToken refreshes access token using refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	// Validate refresh token
	refreshToken, err := j.refreshTokenRepo.GetByToken(refreshTokenString)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	// Generate new token pair
	tokenPair, err := j.GenerateTokenPair(refreshToken.User)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token
	err = j.refreshTokenRepo.Delete(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// RevokeRefreshToken revokes a refresh token
func (j *JWTService) RevokeRefreshToken(token string) error {
	return j.refreshTokenRepo.Delete(token)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (j *JWTService) RevokeAllUserTokens(userID string) error {
	return j.refreshTokenRepo.DeleteByUserID(userID)
}

// GeneratePKCEChallenge generates PKCE code verifier and challenge
func GeneratePKCEChallenge() (*PKCEChallenge, error) {
	// Generate code verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	_, err := rand.Read(verifierBytes)
	if err != nil {
		return nil, err
	}

	codeVerifier := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(verifierBytes)

	// Generate code challenge using S256 method
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return &PKCEChallenge{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		Method:        "S256",
	}, nil
}

// VerifyPKCEChallenge verifies PKCE code verifier against challenge
func VerifyPKCEChallenge(codeVerifier, codeChallenge string) bool {
	hash := sha256.Sum256([]byte(codeVerifier))
	expectedChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	return expectedChallenge == codeChallenge
}

// GenerateState generates a random state for OAuth2
func GenerateState() (string, error) {
	stateBytes := make([]byte, 32)
	_, err := rand.Read(stateBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(stateBytes), nil
}

// GenerateNonce generates a random nonce for OIDC
func GenerateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}
