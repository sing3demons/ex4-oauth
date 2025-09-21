package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents the user entity
type User struct {
	ID            string    `json:"id" bson:"_id,omitempty"`
	Email         string    `json:"email" bson:"email"`
	Username      string    `json:"username" bson:"username"`
	Password      string    `json:"-" bson:"password"`
	FirstName     string    `json:"first_name" bson:"first_name"`
	LastName      string    `json:"last_name" bson:"last_name"`
	Avatar        string    `json:"avatar" bson:"avatar"`
	Provider      string    `json:"provider" bson:"provider"` // 'local', 'google', 'github', etc.
	ProviderID    string    `json:"provider_id" bson:"provider_id"`
	EmailVerified bool      `json:"email_verified" bson:"email_verified"`
	IsActive      bool      `json:"is_active" bson:"is_active"`
	Role          string    `json:"role" bson:"role"` // 'user', 'admin', 'moderator'
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

// OAuthState represents OAuth2 state for security
type OAuthState struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	State        string    `json:"state" bson:"state"`
	CodeVerifier string    `json:"code_verifier" bson:"code_verifier"`         // For PKCE
	Nonce        string    `json:"nonce" bson:"nonce"`                         // For OIDC
	UserID       *string   `json:"user_id,omitempty" bson:"user_id,omitempty"` // Optional: for logged-in users
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

// RefreshToken represents refresh token storage
type RefreshToken struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Token     string    `json:"token" bson:"token"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	User      *User     `json:"user,omitempty" bson:"-"` // Populated during queries if needed
}

// EmailVerification represents email verification tokens
type EmailVerification struct {
	ID         string     `json:"id" bson:"_id,omitempty"`
	UserID     string     `json:"user_id" bson:"user_id"`
	Email      string     `json:"email" bson:"email"`
	Token      string     `json:"token" bson:"token"`
	Type       string     `json:"type" bson:"type"` // 'registration', 'email_change', 'password_reset'
	ExpiresAt  time.Time  `json:"expires_at" bson:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	User       *User      `json:"user,omitempty" bson:"-"` // Populated during queries if needed
}

// EmailTemplate represents email template storage
type EmailTemplate struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name"` // 'welcome', 'verification', 'password_reset'
	Subject   string    `json:"subject" bson:"subject"`
	TextBody  string    `json:"text_body" bson:"text_body"`
	HTMLBody  string    `json:"html_body" bson:"html_body"`
	Variables string    `json:"variables" bson:"variables"` // JSON array of template variables
	IsActive  bool      `json:"is_active" bson:"is_active"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// SetDefaults sets default values for User
func (u *User) SetDefaults() {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.Provider == "" {
		u.Provider = "local"
	}
	if u.Role == "" {
		u.Role = "user"
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	u.UpdatedAt = time.Now()
	u.IsActive = true
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	if u.Password != "" && u.Provider == "local" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByProviderID(provider, providerID string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(limit, offset int) ([]*User, error)
}

// OAuthRepository defines the interface for OAuth state management
type OAuthRepository interface {
	CreateState(state *OAuthState) error
	GetStateByState(state string) (*OAuthState, error)
	DeleteState(state string) error
	CleanupExpiredStates() error
}

// RefreshTokenRepository defines the interface for refresh token management
type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	GetByToken(token string) (*RefreshToken, error)
	GetByUserID(userID string) ([]*RefreshToken, error)
	Delete(token string) error
	DeleteByUserID(userID string) error
	CleanupExpiredTokens() error
}

// EmailVerificationRepository defines the interface for email verification management
type EmailVerificationRepository interface {
	Create(verification *EmailVerification) error
	GetByToken(token string) (*EmailVerification, error)
	GetByUserID(userID string, verificationType string) (*EmailVerification, error)
	GetByEmail(email string, verificationType string) (*EmailVerification, error)
	MarkAsVerified(token string) error
	Delete(token string) error
	DeleteByUserID(userID string, verificationType string) error
	CleanupExpiredTokens() error
}

// EmailTemplateRepository defines the interface for email template management
type EmailTemplateRepository interface {
	Create(template *EmailTemplate) error
	GetByName(name string) (*EmailTemplate, error)
	GetAll() ([]*EmailTemplate, error)
	Update(template *EmailTemplate) error
	Delete(id string) error
}

// AdminStats represents admin dashboard statistics
type AdminStats struct {
	TotalUsers           int64 `json:"total_users"`
	ActiveUsers          int64 `json:"active_users"`
	VerifiedUsers        int64 `json:"verified_users"`
	TotalClients         int64 `json:"total_clients"`
	ActiveClients        int64 `json:"active_clients"`
	TotalTokens          int64 `json:"total_tokens"`
	ExpiredTokens        int64 `json:"expired_tokens"`
	PendingVerifications int64 `json:"pending_verifications"`
	RecentLogins         int64 `json:"recent_logins_24h"`
}

// UserActivity represents user activity log
type UserActivity struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	UserID       string    `json:"user_id" bson:"user_id"`
	Action       string    `json:"action" bson:"action"` // 'login', 'logout', 'register', 'password_change', etc.
	IPAddress    string    `json:"ip_address" bson:"ip_address"`
	UserAgent    string    `json:"user_agent" bson:"user_agent"`
	Details      string    `json:"details" bson:"details"` // JSON string with additional details
	Success      bool      `json:"success" bson:"success"`
	ErrorMessage string    `json:"error_message" bson:"error_message"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	User         *User     `json:"user,omitempty" bson:"-"` // Populated during queries if needed
}

// SystemEvent represents system-level events
type SystemEvent struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	EventType string    `json:"event_type" bson:"event_type"` // 'security', 'system', 'admin', 'oauth'
	Severity  string    `json:"severity" bson:"severity"`     // 'info', 'warning', 'error', 'critical'
	Message   string    `json:"message" bson:"message"`
	Details   string    `json:"details" bson:"details"` // JSON string with event details
	IPAddress string    `json:"ip_address" bson:"ip_address"`
	UserAgent string    `json:"user_agent" bson:"user_agent"`
	UserID    *string   `json:"user_id,omitempty" bson:"user_id,omitempty"` // Optional: user who triggered the event
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	User      *User     `json:"user,omitempty" bson:"-"` // Populated during queries if needed
}

// AuditLog represents comprehensive audit trail for all system actions
type AuditLog struct {
	ID              string    `json:"id" bson:"_id,omitempty"`
	UserID          *string   `json:"user_id,omitempty" bson:"user_id,omitempty"` // Optional: user who performed the action
	SessionID       string    `json:"session_id" bson:"session_id"`               // Session identifier
	Action          string    `json:"action" bson:"action"`                       // 'create', 'read', 'update', 'delete', 'login', 'logout', etc.
	ResourceType    string    `json:"resource_type" bson:"resource_type"`         // 'user', 'client', 'token', 'role', 'permission', etc.
	ResourceID      string    `json:"resource_id" bson:"resource_id"`             // ID of the affected resource
	OldValues       string    `json:"old_values" bson:"old_values"`               // JSON of previous values
	NewValues       string    `json:"new_values" bson:"new_values"`               // JSON of new values
	IPAddress       string    `json:"ip_address" bson:"ip_address"`               // Client IP address
	UserAgent       string    `json:"user_agent" bson:"user_agent"`               // Client user agent
	RequestMethod   string    `json:"request_method" bson:"request_method"`       // HTTP method
	RequestPath     string    `json:"request_path" bson:"request_path"`           // API endpoint
	RequestHeaders  string    `json:"request_headers" bson:"request_headers"`     // Important request headers
	ResponseStatus  int       `json:"response_status" bson:"response_status"`     // HTTP response status
	Success         bool      `json:"success" bson:"success"`                     // Whether action was successful
	ErrorMessage    string    `json:"error_message" bson:"error_message"`         // Error details if failed
	Duration        int64     `json:"duration" bson:"duration"`                   // Request duration in milliseconds
	Severity        string    `json:"severity" bson:"severity"`                   // 'info', 'warning', 'error', 'critical'
	Category        string    `json:"category" bson:"category"`                   // 'authentication', 'authorization', 'data', 'security', 'system'
	RiskLevel       string    `json:"risk_level" bson:"risk_level"`               // 'low', 'medium', 'high', 'critical'
	ComplianceFlags string    `json:"compliance_flags" bson:"compliance_flags"`   // Compliance-related flags (GDPR, HIPAA, etc.)
	GeoLocation     string    `json:"geo_location" bson:"geo_location"`           // Geographic location
	DeviceInfo      string    `json:"device_info" bson:"device_info"`             // Device information
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`               // Timestamp
	User            *User     `json:"user,omitempty" bson:"-"`                    // User relationship
}

// SecurityAlert represents security-related alerts and threats
type SecurityAlert struct {
	ID             string     `json:"id" bson:"_id,omitempty"`
	AlertType      string     `json:"alert_type" bson:"alert_type"`                       // 'failed_login', 'suspicious_activity', 'rate_limit_exceeded', etc.
	Severity       string     `json:"severity" bson:"severity"`                           // 'low', 'medium', 'high', 'critical'
	Status         string     `json:"status" bson:"status"`                               // 'open', 'investigating', 'resolved', 'false_positive'
	UserID         *string    `json:"user_id,omitempty" bson:"user_id,omitempty"`         // Optional: affected user
	IPAddress      string     `json:"ip_address" bson:"ip_address"`                       // Source IP
	UserAgent      string     `json:"user_agent" bson:"user_agent"`                       // User agent
	ThreatLevel    int        `json:"threat_level" bson:"threat_level"`                   // 1-10 threat scale
	Description    string     `json:"description" bson:"description"`                     // Alert description
	Evidence       string     `json:"evidence" bson:"evidence"`                           // JSON evidence data
	Metadata       string     `json:"metadata" bson:"metadata"`                           // Additional metadata
	TriggeredBy    string     `json:"triggered_by" bson:"triggered_by"`                   // What triggered this alert
	ResolvedBy     *string    `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"` // Admin who resolved
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"` // Resolution timestamp
	Resolution     string     `json:"resolution" bson:"resolution"`                       // Resolution notes
	CreatedAt      time.Time  `json:"created_at" bson:"created_at"`                       // Creation timestamp
	UpdatedAt      time.Time  `json:"updated_at" bson:"updated_at"`                       // Last update
	User           *User      `json:"user,omitempty" bson:"-"`                            // User relationship
	ResolvedByUser *User      `json:"resolved_by_user,omitempty" bson:"-"`                // Resolver relationship
}

// ComplianceReport represents compliance audit reports
type ComplianceReport struct {
	ID              string    `json:"id" bson:"_id,omitempty"`
	ReportType      string    `json:"report_type" bson:"report_type"`         // 'gdpr', 'hipaa', 'sox', 'pci_dss', etc.
	Period          string    `json:"period" bson:"period"`                   // Reporting period
	GeneratedBy     string    `json:"generated_by" bson:"generated_by"`       // Admin who generated
	Status          string    `json:"status" bson:"status"`                   // 'draft', 'final', 'archived'
	Summary         string    `json:"summary" bson:"summary"`                 // Executive summary
	Findings        string    `json:"findings" bson:"findings"`               // JSON findings data
	Recommendations string    `json:"recommendations" bson:"recommendations"` // Recommendations
	RiskAssessment  string    `json:"risk_assessment" bson:"risk_assessment"` // Risk assessment
	AttachmentPath  string    `json:"attachment_path" bson:"attachment_path"` // Path to detailed report file
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`           // Creation timestamp
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`           // Last update
	GeneratedByUser *User     `json:"generated_by_user,omitempty" bson:"-"`   // Generator relationship
}

// AdminRepository defines the interface for admin operations
type AdminRepository interface {
	GetStats() (*AdminStats, error)
	GetUsers(limit, offset int, filters map[string]interface{}) ([]*User, int64, error)
	GetUserActivity(userID string, limit, offset int) ([]*UserActivity, error)
	GetSystemEvents(limit, offset int, filters map[string]interface{}) ([]*SystemEvent, error)
	CreateUserActivity(activity *UserActivity) error
	CreateSystemEvent(event *SystemEvent) error
	UpdateUserStatus(userID string, isActive bool) error
	UpdateUserRole(userID string, role string) error
	DeleteUser(userID string) error
	CleanupOldActivities(olderThan time.Time) error
	CleanupOldEvents(olderThan time.Time) error
}

// AuditRepository defines the interface for comprehensive audit operations
type AuditRepository interface {
	CreateAuditLog(log *AuditLog) error
	GetAuditLogs(limit, offset int, filters map[string]interface{}) ([]*AuditLog, int64, error)
	GetAuditLogsByUser(userID string, limit, offset int) ([]*AuditLog, error)
	GetAuditLogsByResource(resourceType, resourceID string, limit, offset int) ([]*AuditLog, error)
	GetAuditLogsByCategory(category string, limit, offset int) ([]*AuditLog, error)
	GetHighRiskAuditLogs(limit, offset int) ([]*AuditLog, error)
	CreateSecurityAlert(alert *SecurityAlert) error
	GetSecurityAlerts(limit, offset int, filters map[string]interface{}) ([]*SecurityAlert, int64, error)
	UpdateSecurityAlert(alertID string, updates map[string]interface{}) error
	ResolveSecurityAlert(alertID string, resolvedBy string, resolution string) error
	CreateComplianceReport(report *ComplianceReport) error
	GetComplianceReports(limit, offset int) ([]*ComplianceReport, error)
	UpdateComplianceReport(reportID string, updates map[string]interface{}) error
	CleanupOldAuditLogs(olderThan time.Time) error
	GetComplianceData(reportType string, startDate, endDate time.Time) (map[string]interface{}, error)
	SearchAuditLogs(query string, filters map[string]interface{}, limit, offset int) ([]*AuditLog, int64, error)
}

// TokenIntrospectionRequest represents OAuth2 token introspection request
type TokenIntrospectionRequest struct {
	Token         string `json:"token" form:"token" binding:"required"`
	TokenTypeHint string `json:"token_type_hint" form:"token_type_hint"` // "access_token" or "refresh_token"
}

// TokenIntrospectionResponse represents OAuth2 token introspection response (RFC 7662)
type TokenIntrospectionResponse struct {
	Active    bool     `json:"active"`               // Required: whether the token is active
	Scope     string   `json:"scope,omitempty"`      // Space-separated list of scopes
	ClientID  string   `json:"client_id,omitempty"`  // Client identifier
	Username  string   `json:"username,omitempty"`   // Human-readable identifier for the resource owner
	TokenType string   `json:"token_type,omitempty"` // Type of token (e.g., "Bearer")
	Exp       int64    `json:"exp,omitempty"`        // Expiration time (Unix timestamp)
	Iat       int64    `json:"iat,omitempty"`        // Issued at time (Unix timestamp)
	Nbf       int64    `json:"nbf,omitempty"`        // Not before time (Unix timestamp)
	Sub       string   `json:"sub,omitempty"`        // Subject identifier
	Aud       []string `json:"aud,omitempty"`        // Audience
	Iss       string   `json:"iss,omitempty"`        // Issuer
	Jti       string   `json:"jti,omitempty"`        // JWT ID

	// Extended fields for our implementation
	UserID   string `json:"user_id,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	Provider string `json:"provider,omitempty"`
}
