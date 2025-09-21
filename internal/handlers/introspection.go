package handlers

import (
	"net/http"
	"strings"

	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// IntrospectionHandler handles token introspection
type IntrospectionHandler struct {
	introspectionService *services.IntrospectionService
}

// NewIntrospectionHandler creates a new introspection handler
func NewIntrospectionHandler(introspectionService *services.IntrospectionService) *IntrospectionHandler {
	return &IntrospectionHandler{
		introspectionService: introspectionService,
	}
}

// IntrospectToken handles token introspection requests (RFC 7662)
func (h *IntrospectionHandler) IntrospectToken(c *gin.Context) {
	// Parse request
	token := c.PostForm("token")
	tokenTypeHint := c.PostForm("token_type_hint")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing token parameter",
		})
		return
	}

	// Basic client authentication
	clientID, clientSecret, hasAuth := c.Request.BasicAuth()
	if !hasAuth {
		c.Header("WWW-Authenticate", `Basic realm="Token Introspection"`)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_client",
			"error_description": "Client authentication required",
		})
		return
	}

	// Introspect token
	response, err := h.introspectionService.IntrospectToken(token, tokenTypeHint, clientID, clientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to introspect token",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateToken handles token validation for internal services
func (h *IntrospectionHandler) ValidateToken(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing Authorization header",
		})
		return
	}

	// Extract token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_request",
			"error_description": "Invalid Authorization header format",
		})
		return
	}

	token := parts[1]

	// Validate token
	isValid, tokenInfo, err := h.introspectionService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to validate token",
		})
		return
	}

	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "Token is invalid or expired",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"token_info": tokenInfo,
	})
}
