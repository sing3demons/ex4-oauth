package services

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"ex4-oauth2/internal/database"
	"ex4-oauth2/internal/models"

	"github.com/gin-gonic/gin"
)

// AuditService handles comprehensive audit logging
type AuditService struct {
	repo *database.Database
}

// NewAuditService creates a new audit service
func NewAuditService(db *database.Database) *AuditService {
	return &AuditService{
		repo: db,
	}
}

// AuditContext represents the context for an audit event
type AuditContext struct {
	UserID         *uint
	SessionID      string
	IPAddress      string
	UserAgent      string
	RequestMethod  string
	RequestPath    string
	RequestHeaders map[string]string
	GeoLocation    string
	DeviceInfo     string
}

// LogAction logs a comprehensive audit event
func (s *AuditService) LogAction(ctx *AuditContext, action, resourceType, resourceID string, oldValues, newValues interface{}) error {
	oldJSON := ""
	newJSON := ""

	if oldValues != nil {
		if data, err := json.Marshal(oldValues); err == nil {
			oldJSON = string(data)
		}
	}

	if newValues != nil {
		if data, err := json.Marshal(newValues); err == nil {
			newJSON = string(data)
		}
	}

	headers := ""
	if ctx.RequestHeaders != nil {
		if data, err := json.Marshal(ctx.RequestHeaders); err == nil {
			headers = string(data)
		}
	}

	auditLog := &models.AuditLog{
		UserID:         ctx.UserID,
		SessionID:      ctx.SessionID,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		OldValues:      oldJSON,
		NewValues:      newJSON,
		IPAddress:      ctx.IPAddress,
		UserAgent:      ctx.UserAgent,
		RequestMethod:  ctx.RequestMethod,
		RequestPath:    ctx.RequestPath,
		RequestHeaders: headers,
		GeoLocation:    ctx.GeoLocation,
		DeviceInfo:     ctx.DeviceInfo,
		Success:        true,
		Category:       s.categorizeAction(action, resourceType),
		RiskLevel:      s.assessRiskLevel(action, resourceType),
		CreatedAt:      time.Now(),
	}

	return s.repo.CreateAuditLog(auditLog)
}

// LogFailedAction logs a failed action with error details
func (s *AuditService) LogFailedAction(ctx *AuditContext, action, resourceType, resourceID, errorMsg string) error {
	auditLog := &models.AuditLog{
		UserID:        ctx.UserID,
		SessionID:     ctx.SessionID,
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		IPAddress:     ctx.IPAddress,
		UserAgent:     ctx.UserAgent,
		RequestMethod: ctx.RequestMethod,
		RequestPath:   ctx.RequestPath,
		Success:       false,
		ErrorMessage:  errorMsg,
		Category:      s.categorizeAction(action, resourceType),
		RiskLevel:     s.assessRiskLevel(action, resourceType),
		Severity:      "warning",
		CreatedAt:     time.Now(),
	}

	return s.repo.CreateAuditLog(auditLog)
}

// LogSecurityEvent logs security-related events
func (s *AuditService) LogSecurityEvent(ctx *AuditContext, eventType, description string, threatLevel int, evidence map[string]interface{}) error {
	evidenceJSON := ""
	if evidence != nil {
		if data, err := json.Marshal(evidence); err == nil {
			evidenceJSON = string(data)
		}
	}

	severity := s.determineSeverity(threatLevel)

	alert := &models.SecurityAlert{
		AlertType:   eventType,
		Severity:    severity,
		UserID:      ctx.UserID,
		IPAddress:   ctx.IPAddress,
		UserAgent:   ctx.UserAgent,
		ThreatLevel: threatLevel,
		Description: description,
		Evidence:    evidenceJSON,
		TriggeredBy: fmt.Sprintf("%s from %s", ctx.RequestPath, ctx.IPAddress),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.repo.CreateSecurityAlert(alert)
}

// LogAuthenticationEvent logs authentication-related events
func (s *AuditService) LogAuthenticationEvent(ctx *AuditContext, action string, success bool, userID *uint, details map[string]interface{}) error {
	detailsJSON := ""
	if details != nil {
		if data, err := json.Marshal(details); err == nil {
			detailsJSON = string(data)
		}
	}

	auditLog := &models.AuditLog{
		UserID:        userID,
		SessionID:     ctx.SessionID,
		Action:        action,
		ResourceType:  "authentication",
		IPAddress:     ctx.IPAddress,
		UserAgent:     ctx.UserAgent,
		RequestMethod: ctx.RequestMethod,
		RequestPath:   ctx.RequestPath,
		Success:       success,
		Category:      "authentication",
		RiskLevel:     s.assessAuthRiskLevel(action, success, ctx.IPAddress),
		NewValues:     detailsJSON,
		CreatedAt:     time.Now(),
	}

	if !success {
		auditLog.Severity = "warning"
		auditLog.ErrorMessage = "Authentication failed"
	}

	return s.repo.CreateAuditLog(auditLog)
}

// GetAuditLogs retrieves audit logs with filtering
func (s *AuditService) GetAuditLogs(limit, offset int, filters map[string]interface{}) ([]*models.AuditLog, int64, error) {
	return s.repo.GetAuditLogs(limit, offset, filters)
}

// GetSecurityAlerts retrieves security alerts
func (s *AuditService) GetSecurityAlerts(limit, offset int, filters map[string]interface{}) ([]*models.SecurityAlert, int64, error) {
	return s.repo.GetSecurityAlerts(limit, offset, filters)
}

// ResolveSecurityAlert marks a security alert as resolved
func (s *AuditService) ResolveSecurityAlert(alertID uint, resolvedBy uint, resolution string) error {
	return s.repo.ResolveSecurityAlert(alertID, resolvedBy, resolution)
}

// GenerateComplianceReport generates a compliance report
func (s *AuditService) GenerateComplianceReport(reportType, period string, generatedBy uint) (*models.ComplianceReport, error) {
	startDate, endDate := s.parsePeriod(period)

	complianceData, err := s.repo.GetComplianceData(reportType, startDate, endDate)
	if err != nil {
		return nil, err
	}

	findings, err := json.Marshal(complianceData)
	if err != nil {
		return nil, err
	}

	report := &models.ComplianceReport{
		ReportType:      reportType,
		Period:          period,
		GeneratedBy:     generatedBy,
		Status:          "draft",
		Summary:         s.generateSummary(reportType, complianceData),
		Findings:        string(findings),
		Recommendations: s.generateRecommendations(reportType, complianceData),
		RiskAssessment:  s.generateRiskAssessment(complianceData),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return report, s.repo.CreateComplianceReport(report)
}

// SearchAuditLogs performs advanced search on audit logs
func (s *AuditService) SearchAuditLogs(query string, filters map[string]interface{}, limit, offset int) ([]*models.AuditLog, int64, error) {
	return s.repo.SearchAuditLogs(query, filters, limit, offset)
}

// GetUserAuditTrail gets complete audit trail for a user
func (s *AuditService) GetUserAuditTrail(userID uint, limit, offset int) ([]*models.AuditLog, error) {
	return s.repo.GetAuditLogsByUser(userID, limit, offset)
}

// GetResourceAuditTrail gets audit trail for a specific resource
func (s *AuditService) GetResourceAuditTrail(resourceType, resourceID string, limit, offset int) ([]*models.AuditLog, error) {
	return s.repo.GetAuditLogsByResource(resourceType, resourceID, limit, offset)
}

// CleanupOldLogs removes old audit logs based on retention policy
func (s *AuditService) CleanupOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.CleanupOldAuditLogs(cutoffDate)
}

// GetHighRiskActivity retrieves high-risk audit logs
func (s *AuditService) GetHighRiskActivity(limit, offset int) ([]*models.AuditLog, error) {
	return s.repo.GetHighRiskAuditLogs(limit, offset)
}

// CreateAuditContextFromGin creates audit context from Gin context
func (s *AuditService) CreateAuditContextFromGin(c *gin.Context) *AuditContext {
	userID := s.getUserIDFromContext(c)
	sessionID := s.getSessionIDFromContext(c)

	// Extract important headers
	headers := make(map[string]string)
	headers["Content-Type"] = c.GetHeader("Content-Type")
	headers["Authorization"] = s.maskAuthHeader(c.GetHeader("Authorization"))
	headers["X-Forwarded-For"] = c.GetHeader("X-Forwarded-For")
	headers["X-Real-IP"] = c.GetHeader("X-Real-IP")

	return &AuditContext{
		UserID:         userID,
		SessionID:      sessionID,
		IPAddress:      s.getClientIP(c),
		UserAgent:      c.GetHeader("User-Agent"),
		RequestMethod:  c.Request.Method,
		RequestPath:    c.Request.URL.Path,
		RequestHeaders: headers,
		GeoLocation:    s.getGeoLocation(s.getClientIP(c)),
		DeviceInfo:     s.parseDeviceInfo(c.GetHeader("User-Agent")),
	}
}

// Helper methods

func (s *AuditService) categorizeAction(action, resourceType string) string {
	switch {
	case strings.Contains(action, "login") || strings.Contains(action, "auth"):
		return "authentication"
	case strings.Contains(action, "permission") || strings.Contains(action, "role"):
		return "authorization"
	case strings.Contains(action, "create") || strings.Contains(action, "update") || strings.Contains(action, "delete"):
		return "data"
	case strings.Contains(action, "security") || strings.Contains(action, "alert"):
		return "security"
	default:
		return "system"
	}
}

func (s *AuditService) assessRiskLevel(action, resourceType string) string {
	switch {
	case strings.Contains(action, "delete") || strings.Contains(action, "admin"):
		return "high"
	case strings.Contains(action, "update") || strings.Contains(action, "role"):
		return "medium"
	case strings.Contains(action, "login") || strings.Contains(action, "failed"):
		return "medium"
	default:
		return "low"
	}
}

func (s *AuditService) assessAuthRiskLevel(action string, success bool, ipAddress string) string {
	if !success {
		return "medium"
	}

	// Check for suspicious IP patterns
	if s.isSuspiciousIP(ipAddress) {
		return "high"
	}

	return "low"
}

func (s *AuditService) determineSeverity(threatLevel int) string {
	switch {
	case threatLevel >= 8:
		return "critical"
	case threatLevel >= 6:
		return "high"
	case threatLevel >= 4:
		return "medium"
	default:
		return "low"
	}
}

func (s *AuditService) getUserIDFromContext(c *gin.Context) *uint {
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

func (s *AuditService) getSessionIDFromContext(c *gin.Context) string {
	if sessionID, exists := c.Get("sessionID"); exists {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	// Generate a temporary session ID based on request
	return fmt.Sprintf("temp_%d", time.Now().UnixNano())
}

func (s *AuditService) getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Use remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

func (s *AuditService) maskAuthHeader(auth string) string {
	if auth == "" {
		return ""
	}
	if len(auth) > 20 {
		return auth[:10] + "***masked***"
	}
	return "***masked***"
}

func (s *AuditService) getGeoLocation(ip string) string {
	// Simple geo-location detection (in production, use a proper service)
	if ip == "127.0.0.1" || ip == "::1" {
		return "localhost"
	}

	// Check for private IP ranges
	if s.isPrivateIP(ip) {
		return "private_network"
	}

	return "unknown"
}

func (s *AuditService) parseDeviceInfo(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	// Simple device detection
	switch {
	case strings.Contains(userAgent, "Mobile"):
		return "mobile"
	case strings.Contains(userAgent, "Tablet"):
		return "tablet"
	default:
		return "desktop"
	}
}

func (s *AuditService) isPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func (s *AuditService) isSuspiciousIP(ip string) bool {
	// Simple suspicious IP detection (in production, use threat intelligence)
	// For now, just check for localhost access to admin endpoints
	return ip == "127.0.0.1" || ip == "::1"
}

func (s *AuditService) parsePeriod(period string) (time.Time, time.Time) {
	now := time.Now()

	switch period {
	case "last_24h":
		return now.AddDate(0, 0, -1), now
	case "last_week":
		return now.AddDate(0, 0, -7), now
	case "last_month":
		return now.AddDate(0, -1, 0), now
	case "last_quarter":
		return now.AddDate(0, -3, 0), now
	case "last_year":
		return now.AddDate(-1, 0, 0), now
	default:
		return now.AddDate(0, -1, 0), now
	}
}

func (s *AuditService) generateSummary(reportType string, data map[string]interface{}) string {
	return fmt.Sprintf("Compliance report for %s generated with %d findings", reportType, len(data))
}

func (s *AuditService) generateRecommendations(reportType string, data map[string]interface{}) string {
	recommendations := []string{
		"Regular security training for all users",
		"Implement multi-factor authentication",
		"Regular security audits and penetration testing",
		"Update password policies",
		"Monitor and log all administrative actions",
	}

	result, _ := json.Marshal(recommendations)
	return string(result)
}

func (s *AuditService) generateRiskAssessment(data map[string]interface{}) string {
	assessment := map[string]string{
		"overall_risk": "medium",
		"data_risk":    "low",
		"access_risk":  "medium",
		"compliance":   "satisfactory",
	}

	result, _ := json.Marshal(assessment)
	return string(result)
}
