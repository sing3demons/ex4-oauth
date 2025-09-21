package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ex4-oauth2/internal/auth"
	"ex4-oauth2/internal/models"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth2/OIDC server requests
type OAuthHandler struct {
	userRepo        models.UserRepository
	clientRepo      models.OAuth2ClientRepository
	authCodeRepo    models.OAuth2AuthorizationCodeRepository
	accessTokenRepo models.OAuth2AccessTokenRepository
	jwtService      *auth.JWTService
	baseURL         string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	userRepo models.UserRepository,
	clientRepo models.OAuth2ClientRepository,
	authCodeRepo models.OAuth2AuthorizationCodeRepository,
	accessTokenRepo models.OAuth2AccessTokenRepository,
	jwtService *auth.JWTService,
	baseURL string,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:        userRepo,
		clientRepo:      clientRepo,
		authCodeRepo:    authCodeRepo,
		accessTokenRepo: accessTokenRepo,
		jwtService:      jwtService,
		baseURL:         baseURL,
	}
}

// AuthorizeRequest represents OAuth2 authorization request
type AuthorizeRequest struct {
	ResponseType        string `form:"response_type" binding:"required"`
	ClientID            string `form:"client_id" binding:"required"`
	RedirectURI         string `form:"redirect_uri" binding:"required"`
	Scope               string `form:"scope"`
	State               string `form:"state"`
	CodeChallenge       string `form:"code_challenge"`
	CodeChallengeMethod string `form:"code_challenge_method"`
	Nonce               string `form:"nonce"`
}

// TokenRequest represents OAuth2 token request
type TokenRequest struct {
	GrantType    string `form:"grant_type" binding:"required"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret"`
	CodeVerifier string `form:"code_verifier"`
	RefreshToken string `form:"refresh_token"`
}

// TokenResponse represents OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfoResponse represents OIDC UserInfo response
type UserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// Authorize handles OAuth2 authorization endpoint
func (h *OAuthHandler) Authorize(c *gin.Context) {
	var req AuthorizeRequest
	if err := c.ShouldBind(&req); err != nil {
		h.redirectError(c, req.RedirectURI, "invalid_request", "Invalid request parameters", req.State)
		return
	}

	// Validate response_type
	if req.ResponseType != "code" {
		h.redirectError(c, req.RedirectURI, "unsupported_response_type", "Only authorization code flow is supported", req.State)
		return
	}

	// Validate client
	client, err := h.clientRepo.GetByClientID(req.ClientID)
	if err != nil {
		h.redirectError(c, req.RedirectURI, "invalid_client", "Invalid client", req.State)
		return
	}

	// Validate redirect URI
	if !h.isValidRedirectURI(client, req.RedirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
		return
	}

	// Check if user is authenticated by looking for Authorization header
	authHeader := c.GetHeader("Authorization")
	var userID string
	var authenticated bool

	if authHeader != "" {
		// Try to extract user from JWT token
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			// Validate JWT token
			claims, err := h.jwtService.ValidateAccessToken(tokenParts[1])
			if err == nil {
				userID = claims.UserID
				authenticated = true
			}
		}
	}

	if !authenticated {
		// User is not authenticated, show login form
		h.showLoginForm(c, &req)
		return
	}

	// User is authenticated, auto-approve and generate authorization code
	// For demo purposes, we'll auto-approve all requests
	h.handleAPIAuthorization(c, userID, &req, client)
}

// Consent handles consent form submission
func (h *OAuthHandler) Consent(c *gin.Context) {
	// Check if user is authenticated by looking for Authorization header (similar to Authorize method)
	authHeader := c.GetHeader("Authorization")
	var userID string
	var authenticated bool

	if authHeader != "" {
		// Try to extract user from JWT token
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			// Validate JWT token
			claims, err := h.jwtService.ValidateAccessToken(tokenParts[1])
			if err == nil {
				userID = claims.UserID
				authenticated = true
			}
		}
	}

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication_required"})
		return
	}

	// Get form data
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	state := c.PostForm("state")
	scope := c.PostForm("scope")
	codeChallenge := c.PostForm("code_challenge")
	challengeMethod := c.PostForm("code_challenge_method")
	nonce := c.PostForm("nonce")
	approved := c.PostForm("approved")

	if approved != "true" {
		h.redirectError(c, redirectURI, "access_denied", "User denied the request", state)
		return
	}

	// Generate authorization code
	code := models.GenerateAuthorizationCode()
	authCode := &models.OAuth2AuthorizationCode{
		Code:            code,
		ClientID:        clientID,
		UserID:          userID,
		RedirectURI:     redirectURI,
		Scope:           scope,
		CodeChallenge:   codeChallenge,
		ChallengeMethod: challengeMethod,
		Nonce:           nonce,
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	}

	if err := h.authCodeRepo.Create(authCode); err != nil {
		h.redirectError(c, redirectURI, "server_error", "Failed to create authorization code", state)
		return
	}

	// Redirect back to client with code
	redirectURL, _ := url.Parse(redirectURI)
	params := redirectURL.Query()
	params.Add("code", code)
	if state != "" {
		params.Add("state", state)
	}
	redirectURL.RawQuery = params.Encode()

	c.Redirect(http.StatusFound, redirectURL.String())
}

// Token handles OAuth2 token endpoint
func (h *OAuthHandler) Token(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid request parameters"})
		return
	}

	switch req.GrantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(c, &req)
	case "refresh_token":
		h.handleRefreshTokenGrant(c, &req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
	}
}

// UserInfo handles OIDC UserInfo endpoint
func (h *OAuthHandler) UserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
		return
	}

	tokenParts := strings.SplitN(authHeader, " ", 2)
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	token := tokenParts[1]
	accessToken, err := h.accessTokenRepo.GetByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return
	}

	user := accessToken.User
	userInfo := UserInfoResponse{
		Sub:           fmt.Sprintf("%d", user.ID),
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Name:          fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		GivenName:     user.FirstName,
		FamilyName:    user.LastName,
		Picture:       user.Avatar,
	}

	c.JSON(http.StatusOK, userInfo)
}

// WellKnown handles OIDC discovery endpoint
func (h *OAuthHandler) WellKnown(c *gin.Context) {
	discovery := gin.H{
		"issuer":                                h.baseURL,
		"authorization_endpoint":                h.baseURL + "/api/auth/oauth/authorize",
		"token_endpoint":                        h.baseURL + "/api/auth/oauth/token",
		"userinfo_endpoint":                     h.baseURL + "/api/auth/oauth/userinfo",
		"jwks_uri":                              h.baseURL + "/api/auth/oauth/.well-known/jwks.json",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256", "HS256"},
		"scopes_supported":                      []string{"openid", "email", "profile"},
		"claims_supported":                      []string{"sub", "email", "email_verified", "name", "given_name", "family_name", "picture"},
		"code_challenge_methods_supported":      []string{"S256"},
	}

	c.JSON(http.StatusOK, discovery)
}

// JWKs handles OIDC JSON Web Key Set endpoint
func (h *OAuthHandler) JWKs(c *gin.Context) {
	jwkSet, err := h.jwtService.GetJWKSet()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to generate JWK set",
		})
		return
	}

	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	c.JSON(http.StatusOK, jwkSet)
}

// Helper methods

func (h *OAuthHandler) handleAuthorizationCodeGrant(c *gin.Context, req *TokenRequest) {
	// First, try to get the client to check if it's public
	client, err := h.clientRepo.GetByClientID(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}

	// For public clients, validate without secret
	if client.IsPublic {
		// Public clients must use PKCE
		if req.CodeVerifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Public clients must use PKCE"})
			return
		}
		// Validate public client
		_, err = h.clientRepo.ValidatePublicClient(req.ClientID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}
	} else {
		// For confidential clients, validate credentials
		_, err = h.clientRepo.ValidateClientCredentials(req.ClientID, req.ClientSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}
	}

	// Get authorization code
	authCode, err := h.authCodeRepo.GetByCode(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Validate client ID
	if authCode.ClientID != req.ClientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Validate redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Validate PKCE if used
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "code_verifier required"})
			return
		}
		if !auth.VerifyPKCEChallenge(req.CodeVerifier, authCode.CodeChallenge) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
			return
		}
	}

	// Mark code as used
	h.authCodeRepo.MarkAsUsed(req.Code)

	// Get user for token generation
	user, err := h.userRepo.GetByID(authCode.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Generate OIDC token pair (includes ID token if openid scope)
	authTime := authCode.CreatedAt // Use authorization code creation time as auth_time
	tokenPair, err := h.jwtService.GenerateOIDCTokenPair(user, authCode.ClientID, authCode.Nonce, authTime, authCode.Scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Store access token record (use the same JWT token that will be returned to client)
	tokenRecord := &models.OAuth2AccessToken{
		Token:     tokenPair.AccessToken,
		ClientID:  authCode.ClientID,
		UserID:    authCode.UserID,
		Scope:     authCode.Scope,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := h.accessTokenRepo.Create(tokenRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Prepare response
	response := gin.H{
		"access_token": tokenPair.AccessToken,
		"token_type":   tokenPair.TokenType,
		"expires_in":   tokenPair.ExpiresIn,
	}

	// Add refresh token if available
	if tokenPair.RefreshToken != "" {
		response["refresh_token"] = tokenPair.RefreshToken
	}

	// Add ID token if available (OIDC)
	if tokenPair.IDToken != "" {
		response["id_token"] = tokenPair.IDToken
	}

	// Add scope if available
	if tokenPair.Scope != "" {
		response["scope"] = tokenPair.Scope
	}

	c.JSON(http.StatusOK, response)
}

func (h *OAuthHandler) handleRefreshTokenGrant(c *gin.Context, req *TokenRequest) {
	// First, try to get the client to check if it's public
	client, err := h.clientRepo.GetByClientID(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}

	// For public clients, validate without secret
	if client.IsPublic {
		// Validate public client
		_, err = h.clientRepo.ValidatePublicClient(req.ClientID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}
	} else {
		// For confidential clients, validate credentials
		_, err = h.clientRepo.ValidateClientCredentials(req.ClientID, req.ClientSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}
	}

	// Refresh the token
	tokens, err := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	// Get user info from the refreshed token to store access token record
	userClaims, err := h.jwtService.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Store the new access token record
	tokenRecord := &models.OAuth2AccessToken{
		Token:     tokens.AccessToken,
		ClientID:  req.ClientID,
		UserID:    userClaims.UserID,
		Scope:     "openid email profile", // Default scope for refresh
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := h.accessTokenRepo.Create(tokenRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *OAuthHandler) isValidRedirectURI(client *models.OAuth2Client, redirectURI string) bool {
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

func (h *OAuthHandler) redirectError(c *gin.Context, redirectURI, errorCode, errorDescription, state string) {
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": errorCode, "error_description": errorDescription})
		return
	}

	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
		return
	}

	params := redirectURL.Query()
	params.Add("error", errorCode)
	params.Add("error_description", errorDescription)
	if state != "" {
		params.Add("state", state)
	}
	redirectURL.RawQuery = params.Encode()

	c.Redirect(http.StatusFound, redirectURL.String())
}

func (h *OAuthHandler) showConsentScreen(c *gin.Context, userID uint, req *AuthorizeRequest, client *models.OAuth2Client) {
	// In a real implementation, you would render an HTML template
	// For now, we'll return a simple HTML response
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Required</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 50px; }
        .container { max-width: 500px; margin: 0 auto; }
        .btn { padding: 10px 20px; margin: 5px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; border: none; }
        .btn-secondary { background: #6c757d; color: white; border: none; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Authorization Required</h2>
        <p><strong>%s</strong> would like to:</p>
        <ul>
            <li>Access your profile information</li>
            <li>Access your email address</li>
        </ul>
        <form method="post" action="/oauth2/consent">
            <input type="hidden" name="client_id" value="%s">
            <input type="hidden" name="redirect_uri" value="%s">
            <input type="hidden" name="state" value="%s">
            <input type="hidden" name="scope" value="%s">
            <input type="hidden" name="code_challenge" value="%s">
            <input type="hidden" name="code_challenge_method" value="%s">
            <input type="hidden" name="nonce" value="%s">
            <button type="submit" name="approved" value="true" class="btn btn-primary">Allow</button>
            <button type="submit" name="approved" value="false" class="btn btn-secondary">Deny</button>
        </form>
    </div>
</body>
</html>`

	html = fmt.Sprintf(html,
		client.Name,
		req.ClientID,
		req.RedirectURI,
		req.State,
		req.Scope,
		req.CodeChallenge,
		req.CodeChallengeMethod,
		req.Nonce,
	)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// handleAPIAuthorization handles OAuth authorization for API clients (auto-approve)
func (h *OAuthHandler) handleAPIAuthorization(c *gin.Context, userID string, req *AuthorizeRequest, client *models.OAuth2Client) {
	// Generate authorization code
	code := models.GenerateAuthorizationCode()
	authCode := &models.OAuth2AuthorizationCode{
		Code:            code,
		ClientID:        req.ClientID,
		UserID:          userID,
		RedirectURI:     req.RedirectURI,
		Scope:           req.Scope,
		CodeChallenge:   req.CodeChallenge,
		ChallengeMethod: req.CodeChallengeMethod,
		Nonce:           req.Nonce,
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	}

	if err := h.authCodeRepo.Create(authCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to create authorization code"})
		return
	}

	// Build redirect URL with authorization code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", req.RedirectURI, code, req.State)

	// Check if this is an API request (looking for JSON Accept header)
	acceptHeader := c.GetHeader("Accept")
	isAPIRequest := strings.Contains(acceptHeader, "application/json")

	if isAPIRequest {
		// Return redirect URL for API client to handle
		c.JSON(http.StatusOK, gin.H{
			"redirect_url": redirectURL,
			"code":         code,
			"state":        req.State})
	} else {
		// For browser requests, redirect directly
		c.Redirect(http.StatusFound, redirectURL)
	}
}

// showLoginForm shows a login form for unauthenticated users
func (h *OAuthHandler) showLoginForm(c *gin.Context, req *AuthorizeRequest) {
	// Build the login form with OAuth parameters embedded
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Login Required</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 50px; background-color: #f5f5f5; }
        .container { max-width: 400px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="email"], input[type="password"] { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        .btn { padding: 10px 20px; margin: 5px 0; cursor: pointer; border: none; border-radius: 4px; width: 100%%; }
        .btn-primary { background: #007bff; color: white; }
        .btn-primary:hover { background: #0056b3; }
        .error { color: red; margin-top: 10px; }
        .oauth-info { background: #f8f9fa; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="oauth-info">
            <h3>Authorization Required</h3>
            <p>You need to login to authorize the OAuth2 application.</p>
        </div>
        
        <form method="post" action="/api/auth/oauth/login">
            <input type="hidden" name="response_type" value="%s">
            <input type="hidden" name="client_id" value="%s">
            <input type="hidden" name="redirect_uri" value="%s">
            <input type="hidden" name="scope" value="%s">
            <input type="hidden" name="state" value="%s">
            <input type="hidden" name="code_challenge" value="%s">
            <input type="hidden" name="code_challenge_method" value="%s">
            <input type="hidden" name="nonce" value="%s">
            
            <div class="form-group">
                <label for="email">Email:</label>
                <input type="email" id="email" name="email" required>
            </div>
            
            <div class="form-group">
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
            </div>
            
            <button type="submit" class="btn btn-primary">Login & Authorize</button>
        </form>
        
        <p style="text-align: center; margin-top: 20px; font-size: 14px; color: #666;">
            Demo credentials: admin@example.com / admin123
        </p>
    </div>
</body>
</html>`

	html = fmt.Sprintf(html,
		req.ResponseType,
		req.ClientID,
		req.RedirectURI,
		req.Scope,
		req.State,
		req.CodeChallenge,
		req.CodeChallengeMethod,
		req.Nonce,
	)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// LoginAndAuthorize handles login form submission and OAuth authorization
func (h *OAuthHandler) LoginAndAuthorize(c *gin.Context) {
	// Get form data
	email := c.PostForm("email")
	password := c.PostForm("password")

	// OAuth parameters
	responseType := c.PostForm("response_type")
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	scope := c.PostForm("scope")
	state := c.PostForm("state")
	codeChallenge := c.PostForm("code_challenge")
	challengeMethod := c.PostForm("code_challenge_method")
	nonce := c.PostForm("nonce")

	// Validate login credentials
	user, err := h.userRepo.GetByEmail(email)
	if err != nil {
		h.showLoginError(c, "Invalid email or password", responseType, clientID, redirectURI, scope, state, codeChallenge, challengeMethod, nonce)
		return
	}

	// Verify password
	if !h.verifyPassword(user, password) {
		h.showLoginError(c, "Invalid email or password", responseType, clientID, redirectURI, scope, state, codeChallenge, challengeMethod, nonce)
		return
	}

	// Validate client
	client, err := h.clientRepo.GetByClientID(clientID)
	if err != nil {
		h.redirectError(c, redirectURI, "invalid_client", "Invalid client", state)
		return
	}

	// Validate redirect URI
	if !h.isValidRedirectURI(client, redirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri"})
		return
	}

	// Auto-approve OAuth request and generate authorization code
	code := models.GenerateAuthorizationCode()
	authCode := &models.OAuth2AuthorizationCode{
		Code:            code,
		ClientID:        clientID,
		UserID:          user.ID,
		RedirectURI:     redirectURI,
		Scope:           scope,
		CodeChallenge:   codeChallenge,
		ChallengeMethod: challengeMethod,
		Nonce:           nonce,
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	}

	if err := h.authCodeRepo.Create(authCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to create authorization code"})
		return
	}

	// Redirect back to client with authorization code
	redirectURL, _ := url.Parse(redirectURI)
	params := redirectURL.Query()
	params.Add("code", code)
	if state != "" {
		params.Add("state", state)
	}
	redirectURL.RawQuery = params.Encode()

	c.Redirect(http.StatusFound, redirectURL.String())
}

// verifyPassword verifies user password (simplified version)
func (h *OAuthHandler) verifyPassword(user *models.User, password string) bool {
	// For demo purposes, we'll accept "admin123" for admin@example.com
	if user.Email == "admin@example.com" && password == "admin123" {
		return true
	}
	return false
}

// showLoginError shows login form with error message
func (h *OAuthHandler) showLoginError(c *gin.Context, errorMsg, responseType, clientID, redirectURI, scope, state, codeChallenge, challengeMethod, nonce string) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Login Required</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 50px; background-color: #f5f5f5; }
        .container { max-width: 400px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input[type="email"], input[type="password"] { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        .btn { padding: 10px 20px; margin: 5px 0; cursor: pointer; border: none; border-radius: 4px; width: 100%%; }
        .btn-primary { background: #007bff; color: white; }
        .error { color: red; margin-bottom: 15px; padding: 10px; background: #ffe6e6; border-radius: 4px; }
        .oauth-info { background: #f8f9fa; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="oauth-info">
            <h3>Authorization Required</h3>
            <p>You need to login to authorize the OAuth2 application.</p>
        </div>
        
        <div class="error">%s</div>
        
        <form method="post" action="/api/auth/oauth/login">
            <input type="hidden" name="response_type" value="%s">
            <input type="hidden" name="client_id" value="%s">
            <input type="hidden" name="redirect_uri" value="%s">
            <input type="hidden" name="scope" value="%s">
            <input type="hidden" name="state" value="%s">
            <input type="hidden" name="code_challenge" value="%s">
            <input type="hidden" name="code_challenge_method" value="%s">
            <input type="hidden" name="nonce" value="%s">
            
            <div class="form-group">
                <label for="email">Email:</label>
                <input type="email" id="email" name="email" required>
            </div>
            
            <div class="form-group">
                <label for="password">Password:</label>
                <input type="password" id="password" name="password" required>
            </div>
            
            <button type="submit" class="btn btn-primary">Login & Authorize</button>
        </form>
        
        <p style="text-align: center; margin-top: 20px; font-size: 14px; color: #666;">
            Demo credentials: admin@example.com / admin123
        </p>
    </div>
</body>
</html>`

	html = fmt.Sprintf(html,
		errorMsg, responseType, clientID, redirectURI, scope, state, codeChallenge, challengeMethod, nonce,
	)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}
