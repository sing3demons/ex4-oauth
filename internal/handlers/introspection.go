package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"ex4-oauth2/internal/models"
	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// IntrospectionHandler handles token introspection endpoints
type IntrospectionHandler struct {
	introspectionService *services.IntrospectionService
}

// NewIntrospectionHandler creates a new introspection handler
func NewIntrospectionHandler(introspectionService *services.IntrospectionService) *IntrospectionHandler {
	return &IntrospectionHandler{
		introspectionService: introspectionService,
	}
}

// IntrospectToken handles POST /oauth2/introspect - RFC 7662 Token Introspection
func (h *IntrospectionHandler) IntrospectToken(c *gin.Context) {
	// Validate client authentication (Basic Auth or client credentials in body)
	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if !hasAuth {
		// Try to get from form data
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
	}

	// Validate client credentials
	if !h.introspectionService.ValidateIntrospectionClient(clientID, clientSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_client",
			"error_description": "Client authentication failed",
		})
		return
	}

	// Parse request
	var req models.TokenIntrospectionRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Invalid token introspection request",
		})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Perform token introspection
	response, err := h.introspectionService.IntrospectToken(req.Token, req.TokenTypeHint, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to introspect token",
		})
		return
	}

	// Return introspection response
	c.JSON(http.StatusOK, response)
}

// RevokeToken handles POST /oauth2/revoke - RFC 7009 Token Revocation
func (h *IntrospectionHandler) RevokeToken(c *gin.Context) {
	// Validate client authentication
	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if !hasAuth {
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
	}

	if !h.introspectionService.ValidateIntrospectionClient(clientID, clientSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_client",
			"error_description": "Client authentication failed",
		})
		return
	}

	// Get token from request
	token := c.PostForm("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing token parameter",
		})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Revoke token
	err := h.introspectionService.RevokeToken(token, clientIP)
	if err != nil {
		// For RFC 7009 compliance, we should return 200 even if token doesn't exist
		// But log the actual error
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{})
}

// GetTokenInfo handles GET /oauth2/tokeninfo - Non-standard endpoint for debugging
func (h *IntrospectionHandler) GetTokenInfo(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing Authorization header",
		})
		return
	}

	// Parse Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_request",
			"error_description": "Invalid Authorization header format",
		})
		return
	}

	token := parts[1]
	clientIP := c.ClientIP()

	// Introspect token
	response, err := h.introspectionService.IntrospectToken(token, "access_token", clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to introspect token",
		})
		return
	}

	// If token is not active, return unauthorized
	if !response.Active {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "Token is not active",
		})
		return
	}

	// Return token information
	c.JSON(http.StatusOK, response)
}

// GetTokenUsageStats handles GET /admin/tokens/stats/:userID - Admin endpoint for token usage statistics
func (h *IntrospectionHandler) GetTokenUsageStats(c *gin.Context) {
	// This would require admin middleware for authorization
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing user ID",
		})
		return
	}

	// Convert userID to uint
	var uid uint
	if _, err := fmt.Sscanf(userID, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get days parameter (default 30)
	daysParam := c.DefaultQuery("days", "30")
	var days int
	if _, err := fmt.Sscanf(daysParam, "%d", &days); err != nil {
		days = 30
	}

	// Get token usage statistics
	stats, err := h.introspectionService.GetTokenUsageStats(uid, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get token usage statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
