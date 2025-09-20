package handlers

import (
	"net/http"
	"strconv"

	"ex4-oauth2/internal/middleware"
	"ex4-oauth2/internal/models"

	"github.com/gin-gonic/gin"
)

// OAuth2ClientHandler handles OAuth2 client management
type OAuth2ClientHandler struct {
	clientRepo models.OAuth2ClientRepository
}

// NewOAuth2ClientHandler creates a new OAuth2 client handler
func NewOAuth2ClientHandler(clientRepo models.OAuth2ClientRepository) *OAuth2ClientHandler {
	return &OAuth2ClientHandler{
		clientRepo: clientRepo,
	}
}

// CreateClientRequest represents OAuth2 client creation request
type CreateClientRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirect_uris" binding:"required"`
	Scopes       []string `json:"scopes"`
	GrantTypes   []string `json:"grant_types"`
}

// ClientResponse represents OAuth2 client response
type ClientResponse struct {
	ID           uint     `json:"id"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret,omitempty"` // Only shown on creation
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	GrantTypes   []string `json:"grant_types"`
	IsActive     bool     `json:"is_active"`
	CreatedAt    string   `json:"created_at"`
}

// CreateClient creates a new OAuth2 client
func (h *OAuth2ClientHandler) CreateClient(c *gin.Context) {
	// For now, allow any authenticated user to create clients
	// In production, implement proper admin authorization
	_, _, _, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	if len(req.Scopes) == 0 {
		req.Scopes = []string{"openid", "email", "profile"}
	}
	if len(req.GrantTypes) == 0 {
		req.GrantTypes = []string{"authorization_code", "refresh_token"}
	}

	// Create client
	client := &models.OAuth2Client{
		Name:         req.Name,
		Description:  req.Description,
		RedirectURIs: req.RedirectURIs,
		Scopes:       req.Scopes,
		GrantTypes:   req.GrantTypes,
		IsActive:     true,
	}

	if err := h.clientRepo.Create(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

	// Return response with client secret (only shown once)
	response := ClientResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret, // Show secret only on creation
		Name:         client.Name,
		Description:  client.Description,
		RedirectURIs: client.RedirectURIs,
		Scopes:       client.Scopes,
		GrantTypes:   client.GrantTypes,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

// GetClients retrieves OAuth2 clients
func (h *OAuth2ClientHandler) GetClients(c *gin.Context) {
	// For now, allow any authenticated user
	// In production, implement proper admin authorization
	_, _, _, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	clients, err := h.clientRepo.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve clients"})
		return
	}

	var responses []ClientResponse
	for _, client := range clients {
		response := ClientResponse{
			ID:       client.ID,
			ClientID: client.ClientID,
			// Don't include client secret in list
			Name:         client.Name,
			Description:  client.Description,
			RedirectURIs: client.RedirectURIs,
			Scopes:       client.Scopes,
			GrantTypes:   client.GrantTypes,
			IsActive:     client.IsActive,
			CreatedAt:    client.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": responses,
		"limit":   limit,
		"offset":  offset,
		"total":   len(responses),
	})
}

// GetClient retrieves a specific OAuth2 client
func (h *OAuth2ClientHandler) GetClient(c *gin.Context) {
	// For now, allow any authenticated user
	// In production, implement proper admin authorization
	_, _, _, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	client, err := h.clientRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	response := ClientResponse{
		ID:       client.ID,
		ClientID: client.ClientID,
		// Don't include client secret
		Name:         client.Name,
		Description:  client.Description,
		RedirectURIs: client.RedirectURIs,
		Scopes:       client.Scopes,
		GrantTypes:   client.GrantTypes,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateClient updates an OAuth2 client
func (h *OAuth2ClientHandler) UpdateClient(c *gin.Context) {
	// For now, allow any authenticated user
	// In production, implement proper admin authorization
	_, _, _, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	client, err := h.clientRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		RedirectURIs []string `json:"redirect_uris"`
		Scopes       []string `json:"scopes"`
		GrantTypes   []string `json:"grant_types"`
		IsActive     *bool    `json:"is_active"` // Use pointer to distinguish between false and not provided
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.Name != "" {
		client.Name = req.Name
	}
	if req.Description != "" {
		client.Description = req.Description
	}
	if len(req.RedirectURIs) > 0 {
		client.RedirectURIs = req.RedirectURIs
	}
	if len(req.Scopes) > 0 {
		client.Scopes = req.Scopes
	}
	if len(req.GrantTypes) > 0 {
		client.GrantTypes = req.GrantTypes
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}

	if err := h.clientRepo.Update(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update client"})
		return
	}

	response := ClientResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		Name:         client.Name,
		Description:  client.Description,
		RedirectURIs: client.RedirectURIs,
		Scopes:       client.Scopes,
		GrantTypes:   client.GrantTypes,
		IsActive:     client.IsActive,
		CreatedAt:    client.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteClient deletes an OAuth2 client
func (h *OAuth2ClientHandler) DeleteClient(c *gin.Context) {
	// For now, allow any authenticated user
	// In production, implement proper admin authorization
	_, _, _, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	if err := h.clientRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}
