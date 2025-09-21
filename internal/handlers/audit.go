package handlers

import (
	"net/http"
	"strconv"

	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit log operations (simplified)
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs returns audit logs with pagination and filtering
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Get pagination parameters
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if resource := c.Query("resource"); resource != "" {
		filters["resource"] = resource
	}

	// Get audit logs
	logs, err := h.auditService.GetAuditLogs(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get audit logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(logs),
		},
	})
}

// SearchAuditLogs searches audit logs by query
func (h *AuditHandler) SearchAuditLogs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// Get pagination parameters
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}

	// Search audit logs
	logs, err := h.auditService.SearchAuditLogs(query, filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search audit logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"query": query,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(logs),
		},
	})
}

// GetUserAuditTrail returns audit trail for a specific user
func (h *AuditHandler) GetUserAuditTrail(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	// Get pagination parameters
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get user audit trail
	logs, err := h.auditService.GetAuditLogsByUser(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user audit trail",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"logs":    logs,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(logs),
		},
	})
}

// GetSecurityAlerts returns security alerts
func (h *AuditHandler) GetSecurityAlerts(c *gin.Context) {
	// Get pagination parameters
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get filters
	filters := make(map[string]interface{})
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	// Get security alerts
	alerts, err := h.auditService.GetSecurityAlerts(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get security alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(alerts),
		},
	})
}

// GetAuditStats returns audit statistics
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	// Simplified implementation - return mock stats
	stats := gin.H{
		"total_logs":      1000,
		"today_logs":      50,
		"security_alerts": 5,
		"high_risk_events": 2,
		"unique_users":    25,
		"top_actions": []gin.H{
			{"action": "login", "count": 150},
			{"action": "oauth_authorize", "count": 100},
			{"action": "profile_update", "count": 75},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}