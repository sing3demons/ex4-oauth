package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents the user entity
type User struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Email         string         `json:"email" gorm:"unique;not null"`
	Username      string         `json:"username" gorm:"unique;not null"`
	Password      string         `json:"-" gorm:"not null"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Avatar        string         `json:"avatar"`
	Provider      string         `json:"provider" gorm:"default:'local'"` // 'local', 'google', 'github', etc.
	ProviderID    string         `json:"provider_id"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	Role          string         `json:"role" gorm:"default:'user'"` // 'user', 'admin', 'moderator'
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// OAuthState represents OAuth2 state for security
type OAuthState struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	State        string    `json:"state" gorm:"unique;not null"`
	CodeVerifier string    `json:"code_verifier" gorm:"not null"` // For PKCE
	Nonce        string    `json:"nonce"`                         // For OIDC
	UserID       *uint     `json:"user_id"`                       // Optional: for logged-in users
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
}

// RefreshToken represents refresh token storage
type RefreshToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Token     string         `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
}

// EmailVerification represents email verification tokens
type EmailVerification struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id" gorm:"not null"`
	Email      string     `json:"email" gorm:"not null"`
	Token      string     `json:"token" gorm:"unique;not null"`
	Type       string     `json:"type" gorm:"not null"` // 'registration', 'email_change', 'password_reset'
	ExpiresAt  time.Time  `json:"expires_at" gorm:"not null"`
	VerifiedAt *time.Time `json:"verified_at"`
	CreatedAt  time.Time  `json:"created_at"`
	User       User       `json:"user" gorm:"foreignKey:UserID"`
}

// EmailTemplate represents email template storage
type EmailTemplate struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"unique;not null"` // 'welcome', 'verification', 'password_reset'
	Subject   string    `json:"subject" gorm:"not null"`
	TextBody  string    `json:"text_body" gorm:"type:text"`
	HTMLBody  string    `json:"html_body" gorm:"type:text"`
	Variables string    `json:"variables" gorm:"type:text"` // JSON array of template variables
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate is a GORM hook to hash password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
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
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByProviderID(provider, providerID string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
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
	GetByUserID(userID uint) ([]*RefreshToken, error)
	Delete(token string) error
	DeleteByUserID(userID uint) error
	CleanupExpiredTokens() error
}

// EmailVerificationRepository defines the interface for email verification management
type EmailVerificationRepository interface {
	Create(verification *EmailVerification) error
	GetByToken(token string) (*EmailVerification, error)
	GetByUserID(userID uint, verificationType string) (*EmailVerification, error)
	GetByEmail(email string, verificationType string) (*EmailVerification, error)
	MarkAsVerified(token string) error
	Delete(token string) error
	DeleteByUserID(userID uint, verificationType string) error
	CleanupExpiredTokens() error
}

// EmailTemplateRepository defines the interface for email template management
type EmailTemplateRepository interface {
	Create(template *EmailTemplate) error
	GetByName(name string) (*EmailTemplate, error)
	GetAll() ([]*EmailTemplate, error)
	Update(template *EmailTemplate) error
	Delete(id uint) error
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
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	Action       string    `json:"action" gorm:"not null"` // 'login', 'logout', 'register', 'password_change', etc.
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Details      string    `json:"details" gorm:"type:text"` // JSON string with additional details
	Success      bool      `json:"success" gorm:"default:true"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
}

// SystemEvent represents system-level events
type SystemEvent struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	EventType string    `json:"event_type" gorm:"not null"` // 'security', 'system', 'admin', 'oauth'
	Severity  string    `json:"severity" gorm:"not null"`   // 'info', 'warning', 'error', 'critical'
	Message   string    `json:"message" gorm:"not null"`
	Details   string    `json:"details" gorm:"type:text"` // JSON string with event details
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	UserID    *uint     `json:"user_id"` // Optional: user who triggered the event
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AuditLog represents comprehensive audit trail for all system actions
type AuditLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          *uint     `json:"user_id"`                                 // Optional: user who performed the action
	SessionID       string    `json:"session_id"`                              // Session identifier
	Action          string    `json:"action" gorm:"not null"`                  // 'create', 'read', 'update', 'delete', 'login', 'logout', etc.
	ResourceType    string    `json:"resource_type" gorm:"not null"`           // 'user', 'client', 'token', 'role', 'permission', etc.
	ResourceID      string    `json:"resource_id"`                             // ID of the affected resource
	OldValues       string    `json:"old_values" gorm:"type:text"`             // JSON of previous values
	NewValues       string    `json:"new_values" gorm:"type:text"`             // JSON of new values
	IPAddress       string    `json:"ip_address"`                              // Client IP address
	UserAgent       string    `json:"user_agent"`                              // Client user agent
	RequestMethod   string    `json:"request_method"`                          // HTTP method
	RequestPath     string    `json:"request_path"`                            // API endpoint
	RequestHeaders  string    `json:"request_headers" gorm:"type:text"`        // Important request headers
	ResponseStatus  int       `json:"response_status"`                         // HTTP response status
	Success         bool      `json:"success" gorm:"default:true"`             // Whether action was successful
	ErrorMessage    string    `json:"error_message"`                           // Error details if failed
	Duration        int64     `json:"duration"`                                // Request duration in milliseconds
	Severity        string    `json:"severity" gorm:"default:'info'"`          // 'info', 'warning', 'error', 'critical'
	Category        string    `json:"category" gorm:"not null"`                // 'authentication', 'authorization', 'data', 'security', 'system'
	RiskLevel       string    `json:"risk_level" gorm:"default:'low'"`         // 'low', 'medium', 'high', 'critical'
	ComplianceFlags string    `json:"compliance_flags"`                        // Compliance-related flags (GDPR, HIPAA, etc.)
	GeoLocation     string    `json:"geo_location"`                            // Geographic location
	DeviceInfo      string    `json:"device_info"`                             // Device information
	CreatedAt       time.Time `json:"created_at"`                              // Timestamp
	User            *User     `json:"user,omitempty" gorm:"foreignKey:UserID"` // User relationship
}

// SecurityAlert represents security-related alerts and threats
type SecurityAlert struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	AlertType      string     `json:"alert_type" gorm:"not null"`                              // 'failed_login', 'suspicious_activity', 'rate_limit_exceeded', etc.
	Severity       string     `json:"severity" gorm:"not null"`                                // 'low', 'medium', 'high', 'critical'
	Status         string     `json:"status" gorm:"default:'open'"`                            // 'open', 'investigating', 'resolved', 'false_positive'
	UserID         *uint      `json:"user_id"`                                                 // Optional: affected user
	IPAddress      string     `json:"ip_address"`                                              // Source IP
	UserAgent      string     `json:"user_agent"`                                              // User agent
	ThreatLevel    int        `json:"threat_level" gorm:"default:1"`                           // 1-10 threat scale
	Description    string     `json:"description" gorm:"not null"`                             // Alert description
	Evidence       string     `json:"evidence" gorm:"type:text"`                               // JSON evidence data
	Metadata       string     `json:"metadata" gorm:"type:text"`                               // Additional metadata
	TriggeredBy    string     `json:"triggered_by"`                                            // What triggered this alert
	ResolvedBy     *uint      `json:"resolved_by"`                                             // Admin who resolved
	ResolvedAt     *time.Time `json:"resolved_at"`                                             // Resolution timestamp
	Resolution     string     `json:"resolution" gorm:"type:text"`                             // Resolution notes
	CreatedAt      time.Time  `json:"created_at"`                                              // Creation timestamp
	UpdatedAt      time.Time  `json:"updated_at"`                                              // Last update
	User           *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`                 // User relationship
	ResolvedByUser *User      `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"` // Resolver relationship
}

// ComplianceReport represents compliance audit reports
type ComplianceReport struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	ReportType      string    `json:"report_type" gorm:"not null"`                     // 'gdpr', 'hipaa', 'sox', 'pci_dss', etc.
	Period          string    `json:"period" gorm:"not null"`                          // Reporting period
	GeneratedBy     uint      `json:"generated_by" gorm:"not null"`                    // Admin who generated
	Status          string    `json:"status" gorm:"default:'draft'"`                   // 'draft', 'final', 'archived'
	Summary         string    `json:"summary" gorm:"type:text"`                        // Executive summary
	Findings        string    `json:"findings" gorm:"type:text"`                       // JSON findings data
	Recommendations string    `json:"recommendations" gorm:"type:text"`                // Recommendations
	RiskAssessment  string    `json:"risk_assessment" gorm:"type:text"`                // Risk assessment
	AttachmentPath  string    `json:"attachment_path"`                                 // Path to detailed report file
	CreatedAt       time.Time `json:"created_at"`                                      // Creation timestamp
	UpdatedAt       time.Time `json:"updated_at"`                                      // Last update
	GeneratedByUser User      `json:"generated_by_user" gorm:"foreignKey:GeneratedBy"` // Generator relationship
}

// AdminRepository defines the interface for admin operations
type AdminRepository interface {
	GetStats() (*AdminStats, error)
	GetUsers(limit, offset int, filters map[string]interface{}) ([]*User, int64, error)
	GetUserActivity(userID uint, limit, offset int) ([]*UserActivity, error)
	GetSystemEvents(limit, offset int, filters map[string]interface{}) ([]*SystemEvent, error)
	CreateUserActivity(activity *UserActivity) error
	CreateSystemEvent(event *SystemEvent) error
	UpdateUserStatus(userID uint, isActive bool) error
	UpdateUserRole(userID uint, role string) error
	DeleteUser(userID uint) error
	CleanupOldActivities(olderThan time.Time) error
	CleanupOldEvents(olderThan time.Time) error
}

// AuditRepository defines the interface for comprehensive audit operations
type AuditRepository interface {
	CreateAuditLog(log *AuditLog) error
	GetAuditLogs(limit, offset int, filters map[string]interface{}) ([]*AuditLog, int64, error)
	GetAuditLogsByUser(userID uint, limit, offset int) ([]*AuditLog, error)
	GetAuditLogsByResource(resourceType, resourceID string, limit, offset int) ([]*AuditLog, error)
	GetAuditLogsByCategory(category string, limit, offset int) ([]*AuditLog, error)
	GetHighRiskAuditLogs(limit, offset int) ([]*AuditLog, error)
	CreateSecurityAlert(alert *SecurityAlert) error
	GetSecurityAlerts(limit, offset int, filters map[string]interface{}) ([]*SecurityAlert, int64, error)
	UpdateSecurityAlert(alertID uint, updates map[string]interface{}) error
	ResolveSecurityAlert(alertID uint, resolvedBy uint, resolution string) error
	CreateComplianceReport(report *ComplianceReport) error
	GetComplianceReports(limit, offset int) ([]*ComplianceReport, error)
	UpdateComplianceReport(reportID uint, updates map[string]interface{}) error
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
	UserID   uint   `json:"user_id,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	Provider string `json:"provider,omitempty"`
}
