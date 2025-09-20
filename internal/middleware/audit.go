package middleware

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware provides automatic audit logging for API requests
type AuditMiddleware struct {
	auditService *services.AuditService
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditService *services.AuditService) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
	}
}

// LogRequest logs all incoming requests automatically
func (m *AuditMiddleware) LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Create audit context
		auditCtx := m.auditService.CreateAuditContextFromGin(c)

		// Read request body for logging (if present)
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Use a custom ResponseWriter to capture response
		writer := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Determine action based on HTTP method and path
		action := m.determineAction(c.Request.Method, c.FullPath())
		resourceType := m.determineResourceType(c.FullPath())
		resourceID := m.extractResourceID(c)

		// Log the request
		success := writer.status < 400
		
		if success {
			err := m.auditService.LogAction(auditCtx, action, resourceType, resourceID, 
				m.createRequestData(c, string(requestBody)), 
				m.createResponseData(writer))
			if err != nil {
				// Log error but don't fail the request
				c.Header("X-Audit-Error", "Failed to log audit")
			}
		} else {
			errorMsg := m.extractErrorMessage(writer)
			err := m.auditService.LogFailedAction(auditCtx, action, resourceType, resourceID, errorMsg)
			if err != nil {
				c.Header("X-Audit-Error", "Failed to log failed audit")
			}
		}

		// Update audit log with response details
		m.updateAuditLogWithResponse(auditCtx, writer.status, duration)
	}
}

// LogAuthenticationEvents logs authentication-specific events
func (m *AuditMiddleware) LogAuthenticationEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		auditCtx := m.auditService.CreateAuditContextFromGin(c)

		// Process request
		c.Next()

		// Check if this is an authentication endpoint
		if m.isAuthenticationEndpoint(c.FullPath()) {
			action := m.determineAuthAction(c.Request.Method, c.FullPath())
			success := c.Writer.Status() < 400
			userID := m.getUserIDFromResponse(c)

			details := map[string]interface{}{
				"endpoint":    c.FullPath(),
				"method":      c.Request.Method,
				"status_code": c.Writer.Status(),
			}

			err := m.auditService.LogAuthenticationEvent(auditCtx, action, success, userID, details)
			if err != nil {
				c.Header("X-Auth-Audit-Error", "Failed to log authentication event")
			}
		}
	}
}

// LogSecurityEvents logs security-related events
func (m *AuditMiddleware) LogSecurityEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		auditCtx := m.auditService.CreateAuditContextFromGin(c)

		// Check for security-related conditions before processing
		m.checkForSecurityThreats(c, auditCtx)

		// Process request
		c.Next()

		// Check for security events after processing
		if m.shouldLogSecurityEvent(c) {
			m.logSecurityEvent(c, auditCtx)
		}
	}
}

// Helper types and methods

type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.body == nil {
		w.body = &bytes.Buffer{}
	}
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Helper methods

func (m *AuditMiddleware) determineAction(method, path string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		if m.isCreateEndpoint(path) {
			return "create"
		}
		return "action"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

func (m *AuditMiddleware) determineResourceType(path string) string {
	switch {
	case contains(path, "/users"):
		return "user"
	case contains(path, "/clients"):
		return "oauth_client"
	case contains(path, "/tokens"):
		return "token"
	case contains(path, "/auth"):
		return "authentication"
	case contains(path, "/admin"):
		return "admin"
	case contains(path, "/audit"):
		return "audit"
	default:
		return "system"
	}
}

func (m *AuditMiddleware) extractResourceID(c *gin.Context) string {
	// Try to get ID from URL parameters
	if id := c.Param("id"); id != "" {
		return id
	}
	if userID := c.Param("user_id"); userID != "" {
		return userID
	}
	if clientID := c.Param("client_id"); clientID != "" {
		return clientID
	}
	return ""
}

func (m *AuditMiddleware) createRequestData(c *gin.Context, body string) map[string]interface{} {
	data := map[string]interface{}{
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"query":       c.Request.URL.RawQuery,
		"content_type": c.GetHeader("Content-Type"),
	}

	// Only include body for POST/PUT requests and if it's not too large
	if (c.Request.Method == "POST" || c.Request.Method == "PUT") && len(body) < 1000 {
		data["body_preview"] = body
	}

	return data
}

func (m *AuditMiddleware) createResponseData(writer *responseWriter) map[string]interface{} {
	data := map[string]interface{}{
		"status_code": writer.status,
	}

	if writer.body != nil && writer.body.Len() < 500 {
		data["body_preview"] = writer.body.String()
	}

	return data
}

func (m *AuditMiddleware) extractErrorMessage(writer *responseWriter) string {
	if writer.body != nil {
		return writer.body.String()
	}
	return "Unknown error"
}

func (m *AuditMiddleware) updateAuditLogWithResponse(auditCtx *services.AuditContext, status int, duration int64) {
	// This could be implemented to update the last audit log entry with response details
	// For now, we include this in the original log entry
}

func (m *AuditMiddleware) isAuthenticationEndpoint(path string) bool {
	authPaths := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/logout",
		"/api/auth/refresh",
		"/api/auth/oauth/token",
		"/api/auth/oauth/authorize",
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}
	return false
}

func (m *AuditMiddleware) determineAuthAction(method, path string) string {
	switch {
	case contains(path, "/login"):
		return "login"
	case contains(path, "/register"):
		return "register"
	case contains(path, "/logout"):
		return "logout"
	case contains(path, "/refresh"):
		return "token_refresh"
	case contains(path, "/oauth/token"):
		return "oauth_token"
	case contains(path, "/oauth/authorize"):
		return "oauth_authorize"
	default:
		return "auth_action"
	}
}

func (m *AuditMiddleware) getUserIDFromResponse(c *gin.Context) *uint {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(uint); ok {
			return &id
		}
		if id, ok := userID.(string); ok {
			if parsed, err := strconv.ParseUint(id, 10, 32); err == nil {
				uid := uint(parsed)
				return &uid
			}
		}
	}
	return nil
}

func (m *AuditMiddleware) checkForSecurityThreats(c *gin.Context, auditCtx *services.AuditContext) {
	// Check for suspicious patterns
	userAgent := c.GetHeader("User-Agent")
	
	// Check for bot/scanner patterns
	if m.isSuspiciousUserAgent(userAgent) {
		evidence := map[string]interface{}{
			"user_agent": userAgent,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
		}
		
		m.auditService.LogSecurityEvent(auditCtx, "suspicious_user_agent", 
			"Suspicious user agent detected", 3, evidence)
	}

	// Check for SQL injection patterns in query parameters
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			if m.hasSQLInjectionPattern(value) {
				evidence := map[string]interface{}{
					"parameter": key,
					"value":     value,
					"path":      c.Request.URL.Path,
				}
				
				m.auditService.LogSecurityEvent(auditCtx, "sql_injection_attempt",
					"Potential SQL injection detected", 7, evidence)
			}
		}
	}
}

func (m *AuditMiddleware) shouldLogSecurityEvent(c *gin.Context) bool {
	// Log security events for failed authentication, admin access, etc.
	status := c.Writer.Status()
	path := c.FullPath()

	return status == 401 || status == 403 || contains(path, "/admin")
}

func (m *AuditMiddleware) logSecurityEvent(c *gin.Context, auditCtx *services.AuditContext) {
	status := c.Writer.Status()
	path := c.FullPath()

	var eventType, description string
	threatLevel := 1

	switch status {
	case 401:
		eventType = "unauthorized_access"
		description = "Unauthorized access attempt"
		threatLevel = 4
	case 403:
		eventType = "forbidden_access"
		description = "Forbidden access attempt"
		threatLevel = 5
	default:
		if contains(path, "/admin") {
			eventType = "admin_access"
			description = "Admin endpoint access"
			threatLevel = 2
		}
	}

	if eventType != "" {
		evidence := map[string]interface{}{
			"status_code": status,
			"path":        path,
			"method":      c.Request.Method,
		}

		m.auditService.LogSecurityEvent(auditCtx, eventType, description, threatLevel, evidence)
	}
}

func (m *AuditMiddleware) isCreateEndpoint(path string) bool {
	createPaths := []string{
		"/register",
		"/clients",
		"/users",
	}

	for _, createPath := range createPaths {
		if contains(path, createPath) {
			return true
		}
	}
	return false
}

func (m *AuditMiddleware) isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"curl",
		"wget",
		"python-requests",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if contains(userAgentLower, pattern) {
			return true
		}
	}
	return false
}

func (m *AuditMiddleware) hasSQLInjectionPattern(value string) bool {
	sqlPatterns := []string{
		"'",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"union",
		"select",
		"drop",
		"delete",
		"insert",
		"update",
	}

	valueLower := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if contains(valueLower, pattern) {
			return true
		}
	}
	return false
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}