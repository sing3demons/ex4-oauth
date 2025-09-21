package middleware

import (
	"fmt"
	"time"

	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware provides comprehensive audit logging (simplified)
type AuditMiddleware struct {
	auditService *services.AuditService
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditService *services.AuditService) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
	}
}

// AuditMiddleware logs all requests
func (m *AuditMiddleware) AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log the request
		duration := time.Since(start)
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		fmt.Printf("Audit: Method=%s, Path=%s, UserID=%s, StatusCode=%d, Duration=%v, IP=%s\n",
			c.Request.Method,
			c.Request.URL.Path,
			userID,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)
	}
}

// SecurityAuditMiddleware logs security-related events
func (m *AuditMiddleware) SecurityAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log security events for authentication endpoints
		if c.Request.URL.Path == "/api/auth/login" || 
		   c.Request.URL.Path == "/api/auth/register" ||
		   c.Request.URL.Path == "/api/auth/oauth/authorize" {
			
			userID := c.GetString("user_id")
			if userID == "" {
				userID = "anonymous"
			}

			fmt.Printf("Security Audit: Path=%s, UserID=%s, StatusCode=%d, IP=%s\n",
				c.Request.URL.Path,
				userID,
				c.Writer.Status(),
				c.ClientIP(),
			)
		}
	}
}

// AdminAuditMiddleware logs admin actions
func (m *AuditMiddleware) AdminAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log admin actions
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		fmt.Printf("Admin Audit: Method=%s, Path=%s, UserID=%s, StatusCode=%d, IP=%s\n",
			c.Request.Method,
			c.Request.URL.Path,
			userID,
			c.Writer.Status(),
			c.ClientIP(),
		)
	}
}

// DataAccessAuditMiddleware logs data access events
func (m *AuditMiddleware) DataAccessAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log data access
		duration := time.Since(start)
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		fmt.Printf("Data Access Audit: Method=%s, Path=%s, UserID=%s, Duration=%v, IP=%s\n",
			c.Request.Method,
			c.Request.URL.Path,
			userID,
			duration,
			c.ClientIP(),
		)
	}
}