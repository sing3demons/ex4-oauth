package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs retrieves audit logs with filtering
// @Summary Get audit logs
// @Description Retrieve audit logs with optional filtering
// @Tags audit
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param user_id query int false "User ID"
// @Param action query string false "Action"
// @Param resource_type query string false "Resource type"
// @Param category query string false "Category"
// @Param risk_level query string false "Risk level"
// @Param success query bool false "Success status"
// @Param from_date query string false "From date (RFC3339)"
// @Param to_date query string false "To date (RFC3339)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/logs [get]
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := make(map[string]interface{})

	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			filters["user_id"] = uint(id)
		}
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	if riskLevel := c.Query("risk_level"); riskLevel != "" {
		filters["risk_level"] = riskLevel
	}

	if success := c.Query("success"); success != "" {
		if s, err := strconv.ParseBool(success); err == nil {
			filters["success"] = s
		}
	}

	if fromDate := c.Query("from_date"); fromDate != "" {
		if date, err := time.Parse(time.RFC3339, fromDate); err == nil {
			filters["from_date"] = date
		}
	}

	if toDate := c.Query("to_date"); toDate != "" {
		if date, err := time.Parse(time.RFC3339, toDate); err == nil {
			filters["to_date"] = date
		}
	}

	logs, total, err := h.auditService.GetAuditLogs(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// SearchAuditLogs performs text search on audit logs
// @Summary Search audit logs
// @Description Search audit logs using text query
// @Tags audit
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param category query string false "Category"
// @Param risk_level query string false "Risk level"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/search [get]
func (h *AuditHandler) SearchAuditLogs(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := make(map[string]interface{})
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if riskLevel := c.Query("risk_level"); riskLevel != "" {
		filters["risk_level"] = riskLevel
	}

	logs, total, err := h.auditService.SearchAuditLogs(query, filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"query": query,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUserAuditTrail retrieves audit trail for a specific user
// @Summary Get user audit trail
// @Description Retrieve complete audit trail for a specific user
// @Tags audit
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/users/{id}/trail [get]
func (h *AuditHandler) GetUserAuditTrail(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.auditService.GetUserAuditTrail(uint(userID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user audit trail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"user_id": userID,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetResourceAuditTrail retrieves audit trail for a specific resource
// @Summary Get resource audit trail
// @Description Retrieve audit trail for a specific resource
// @Tags audit
// @Accept json
// @Produce json
// @Param resource_type path string true "Resource type"
// @Param resource_id path string true "Resource ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/resources/{resource_type}/{resource_id}/trail [get]
func (h *AuditHandler) GetResourceAuditTrail(c *gin.Context) {
	resourceType := c.Param("resource_type")
	resourceID := c.Param("resource_id")

	if resourceType == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource type and ID are required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.auditService.GetResourceAuditTrail(resourceType, resourceID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve resource audit trail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"resource": gin.H{
			"type": resourceType,
			"id":   resourceID,
		},
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetHighRiskActivity retrieves high-risk audit logs
// @Summary Get high-risk activity
// @Description Retrieve high-risk audit logs
// @Tags audit
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/high-risk [get]
func (h *AuditHandler) GetHighRiskActivity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.auditService.GetHighRiskActivity(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve high-risk activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetSecurityAlerts retrieves security alerts
// @Summary Get security alerts
// @Description Retrieve security alerts with filtering
// @Tags audit
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param alert_type query string false "Alert type"
// @Param severity query string false "Severity"
// @Param status query string false "Status"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/security-alerts [get]
func (h *AuditHandler) GetSecurityAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := make(map[string]interface{})
	if alertType := c.Query("alert_type"); alertType != "" {
		filters["alert_type"] = alertType
	}
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	alerts, total, err := h.auditService.GetSecurityAlerts(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// ResolveSecurityAlert marks a security alert as resolved
// @Summary Resolve security alert
// @Description Mark a security alert as resolved
// @Tags audit
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param body body map[string]string true "Resolution details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/security-alerts/{id}/resolve [post]
func (h *AuditHandler) ResolveSecurityAlert(c *gin.Context) {
	alertID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var requestBody struct {
		Resolution string `json:"resolution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin authentication required"})
		return
	}

	resolvedBy, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin user ID"})
		return
	}

	err = h.auditService.ResolveSecurityAlert(uint(alertID), resolvedBy, requestBody.Resolution)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve security alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Security alert resolved successfully",
		"alert_id":  alertID,
		"resolved_by": resolvedBy,
	})
}

// GenerateComplianceReport generates a compliance report
// @Summary Generate compliance report
// @Description Generate a compliance report for a specific type and period
// @Tags audit
// @Accept json
// @Produce json
// @Param body body map[string]string true "Report parameters"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/compliance/reports [post]
func (h *AuditHandler) GenerateComplianceReport(c *gin.Context) {
	var requestBody struct {
		ReportType string `json:"report_type" binding:"required"`
		Period     string `json:"period" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin authentication required"})
		return
	}

	generatedBy, ok := adminID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin user ID"})
		return
	}

	report, err := h.auditService.GenerateComplianceReport(requestBody.ReportType, requestBody.Period, generatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate compliance report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Compliance report generated successfully",
		"report":  report,
	})
}

// CleanupOldLogs removes old audit logs
// @Summary Cleanup old audit logs
// @Description Remove audit logs older than specified retention period
// @Tags audit
// @Accept json
// @Produce json
// @Param body body map[string]int true "Cleanup parameters"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/audit/cleanup [post]
func (h *AuditHandler) CleanupOldLogs(c *gin.Context) {
	var requestBody struct {
		RetentionDays int `json:"retention_days" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.auditService.CleanupOldLogs(requestBody.RetentionDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup old logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Old audit logs cleaned up successfully",
		"retention_days": requestBody.RetentionDays,
	})
}