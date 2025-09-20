package database

import (
	"errors"
	"fmt"
	"time"

	"ex4-oauth2/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database wraps the GORM database instance
type Database struct {
	db *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databaseURL string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schemas
	err = db.AutoMigrate(
		&models.User{},
		&models.OAuthState{},
		&models.RefreshToken{},
		&models.OAuth2Client{},
		&models.OAuth2AuthorizationCode{},
		&models.OAuth2AccessToken{},
		&models.EmailVerification{},
		&models.EmailTemplate{},
		&models.UserActivity{},
		&models.SystemEvent{},
		&models.AuditLog{},
		&models.SecurityAlert{},
		&models.ComplianceReport{},
	)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// GetDB returns the underlying GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// UserRepository implements the UserRepository interface
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) models.UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByProviderID retrieves a user by provider and provider ID
func (r *UserRepository) GetByProviderID(provider, providerID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// List retrieves users with pagination
func (r *UserRepository) List(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// OAuthRepository implements the OAuthRepository interface
type OAuthRepository struct {
	db *gorm.DB
}

// NewOAuthRepository creates a new OAuth repository
func NewOAuthRepository(db *gorm.DB) models.OAuthRepository {
	return &OAuthRepository{db: db}
}

// CreateState creates a new OAuth state
func (r *OAuthRepository) CreateState(state *models.OAuthState) error {
	return r.db.Create(state).Error
}

// GetStateByState retrieves OAuth state by state string
func (r *OAuthRepository) GetStateByState(state string) (*models.OAuthState, error) {
	var oauthState models.OAuthState
	err := r.db.Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("state not found")
		}
		return nil, err
	}
	return &oauthState, nil
}

// DeleteState deletes OAuth state by state string
func (r *OAuthRepository) DeleteState(state string) error {
	return r.db.Where("state = ?", state).Delete(&models.OAuthState{}).Error
}

// CleanupExpiredStates removes expired OAuth states
func (r *OAuthRepository) CleanupExpiredStates() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.OAuthState{}).Error
}

// RefreshTokenRepository implements the RefreshTokenRepository interface
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) models.RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// GetByToken retrieves refresh token by token string
func (r *RefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Preload("User").Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &refreshToken, nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *RefreshTokenRepository) GetByUserID(userID uint) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// Delete deletes refresh token by token string
func (r *RefreshTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *RefreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

// CleanupExpiredTokens removes expired refresh tokens
func (r *RefreshTokenRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}

// EmailVerificationRepository implements the EmailVerificationRepository interface
type EmailVerificationRepository struct {
	db *gorm.DB
}

// NewEmailVerificationRepository creates a new email verification repository
func NewEmailVerificationRepository(db *gorm.DB) models.EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

// Create creates a new email verification record
func (r *EmailVerificationRepository) Create(verification *models.EmailVerification) error {
	return r.db.Create(verification).Error
}

// GetByToken retrieves email verification by token
func (r *EmailVerificationRepository) GetByToken(token string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.Preload("User").Where("token = ?", token).First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("verification token not found")
		}
		return nil, err
	}
	return &verification, nil
}

// GetByUserID retrieves email verification by user ID and type
func (r *EmailVerificationRepository) GetByUserID(userID uint, verificationType string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.Where("user_id = ? AND type = ? AND verified_at IS NULL", userID, verificationType).First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("verification not found")
		}
		return nil, err
	}
	return &verification, nil
}

// GetByEmail retrieves email verification by email and type
func (r *EmailVerificationRepository) GetByEmail(email string, verificationType string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.Where("email = ? AND type = ? AND verified_at IS NULL", email, verificationType).First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("verification not found")
		}
		return nil, err
	}
	return &verification, nil
}

// MarkAsVerified marks email verification as verified
func (r *EmailVerificationRepository) MarkAsVerified(token string) error {
	now := time.Now()
	return r.db.Model(&models.EmailVerification{}).Where("token = ?", token).Update("verified_at", &now).Error
}

// Delete deletes email verification by token
func (r *EmailVerificationRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.EmailVerification{}).Error
}

// DeleteByUserID deletes email verifications by user ID and type
func (r *EmailVerificationRepository) DeleteByUserID(userID uint, verificationType string) error {
	return r.db.Where("user_id = ? AND type = ?", userID, verificationType).Delete(&models.EmailVerification{}).Error
}

// CleanupExpiredTokens removes expired email verification tokens
func (r *EmailVerificationRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.EmailVerification{}).Error
}

// EmailTemplateRepository implements the EmailTemplateRepository interface
type EmailTemplateRepository struct {
	db *gorm.DB
}

// NewEmailTemplateRepository creates a new email template repository
func NewEmailTemplateRepository(db *gorm.DB) models.EmailTemplateRepository {
	return &EmailTemplateRepository{db: db}
}

// Create creates a new email template
func (r *EmailTemplateRepository) Create(template *models.EmailTemplate) error {
	return r.db.Create(template).Error
}

// GetByName retrieves email template by name
func (r *EmailTemplateRepository) GetByName(name string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return &template, nil
}

// GetAll retrieves all email templates
func (r *EmailTemplateRepository) GetAll() ([]*models.EmailTemplate, error) {
	var templates []*models.EmailTemplate
	err := r.db.Where("is_active = ?", true).Find(&templates).Error
	return templates, err
}

// Update updates an email template
func (r *EmailTemplateRepository) Update(template *models.EmailTemplate) error {
	return r.db.Save(template).Error
}

// Delete deletes an email template
func (r *EmailTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.EmailTemplate{}, id).Error
}

// AdminRepository implements the AdminRepository interface
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *gorm.DB) models.AdminRepository {
	return &AdminRepository{db: db}
}

// GetStats returns dashboard statistics
func (r *AdminRepository) GetStats() (*models.AdminStats, error) {
	stats := &models.AdminStats{}

	// Count total users
	r.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// Count active users
	r.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)

	// Count verified users
	r.db.Model(&models.User{}).Where("email_verified = ?", true).Count(&stats.VerifiedUsers)

	// Count total OAuth2 clients
	r.db.Model(&models.OAuth2Client{}).Count(&stats.TotalClients)

	// Count active OAuth2 clients
	r.db.Model(&models.OAuth2Client{}).Where("is_active = ?", true).Count(&stats.ActiveClients)

	// Count total refresh tokens
	r.db.Model(&models.RefreshToken{}).Count(&stats.TotalTokens)

	// Count expired tokens
	r.db.Model(&models.RefreshToken{}).Where("expires_at < ?", time.Now()).Count(&stats.ExpiredTokens)

	// Count pending email verifications
	r.db.Model(&models.EmailVerification{}).Where("verified_at IS NULL AND expires_at > ?", time.Now()).Count(&stats.PendingVerifications)

	// Count recent logins (last 24 hours)
	yesterday := time.Now().Add(-24 * time.Hour)
	r.db.Model(&models.UserActivity{}).Where("action = ? AND created_at > ? AND success = ?", "login", yesterday, true).Count(&stats.RecentLogins)

	return stats, nil
}

// GetUsers returns paginated user list with filters
func (r *AdminRepository) GetUsers(limit, offset int, filters map[string]interface{}) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if role, ok := filters["role"]; ok {
		query = query.Where("role = ?", role)
	}
	if isActive, ok := filters["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}
	if emailVerified, ok := filters["email_verified"]; ok {
		query = query.Where("email_verified = ?", emailVerified)
	}
	if provider, ok := filters["provider"]; ok {
		query = query.Where("provider = ?", provider)
	}
	if search, ok := filters["search"]; ok {
		searchStr := fmt.Sprintf("%%%v%%", search)
		query = query.Where("email LIKE ? OR username LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
			searchStr, searchStr, searchStr, searchStr)
	}

	// Count total
	query.Count(&total)

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error

	return users, total, err
}

// GetUserActivity returns user activity logs
func (r *AdminRepository) GetUserActivity(userID uint, limit, offset int) ([]*models.UserActivity, error) {
	var activities []*models.UserActivity

	query := r.db.Model(&models.UserActivity{}).Preload("User")

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&activities).Error
	return activities, err
}

// GetSystemEvents returns system event logs
func (r *AdminRepository) GetSystemEvents(limit, offset int, filters map[string]interface{}) ([]*models.SystemEvent, error) {
	var events []*models.SystemEvent

	query := r.db.Model(&models.SystemEvent{}).Preload("User")

	// Apply filters
	if eventType, ok := filters["event_type"]; ok {
		query = query.Where("event_type = ?", eventType)
	}
	if severity, ok := filters["severity"]; ok {
		query = query.Where("severity = ?", severity)
	}
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if from, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", from)
	}
	if to, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", to)
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&events).Error
	return events, err
}

// CreateUserActivity creates a new user activity log
func (r *AdminRepository) CreateUserActivity(activity *models.UserActivity) error {
	return r.db.Create(activity).Error
}

// CreateSystemEvent creates a new system event log
func (r *AdminRepository) CreateSystemEvent(event *models.SystemEvent) error {
	return r.db.Create(event).Error
}

// UpdateUserStatus updates user active status
func (r *AdminRepository) UpdateUserStatus(userID uint, isActive bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", isActive).Error
}

// UpdateUserRole updates user role
func (r *AdminRepository) UpdateUserRole(userID uint, role string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

// DeleteUser soft deletes a user
func (r *AdminRepository) DeleteUser(userID uint) error {
	return r.db.Delete(&models.User{}, userID).Error
}

// CleanupOldActivities removes old activity logs
func (r *AdminRepository) CleanupOldActivities(olderThan time.Time) error {
	return r.db.Where("created_at < ?", olderThan).Delete(&models.UserActivity{}).Error
}

// CleanupOldEvents removes old system events
func (r *AdminRepository) CleanupOldEvents(olderThan time.Time) error {
	return r.db.Where("created_at < ?", olderThan).Delete(&models.SystemEvent{}).Error
}

// === Audit Repository Methods ===

// CreateAuditLog creates a new audit log entry
func (d *Database) CreateAuditLog(log *models.AuditLog) error {
	return d.db.Create(log).Error
}

// GetAuditLogs retrieves audit logs with filtering
func (d *Database) GetAuditLogs(limit, offset int, filters map[string]interface{}) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64

	query := d.db.Model(&models.AuditLog{}).Preload("User")

	// Apply filters
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if resourceType, ok := filters["resource_type"]; ok {
		query = query.Where("resource_type = ?", resourceType)
	}
	if category, ok := filters["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if riskLevel, ok := filters["risk_level"]; ok {
		query = query.Where("risk_level = ?", riskLevel)
	}
	if success, ok := filters["success"]; ok {
		query = query.Where("success = ?", success)
	}
	if from, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", from)
	}
	if to, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", to)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (d *Database) GetAuditLogsByUser(userID uint, limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := d.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetAuditLogsByResource retrieves audit logs for a specific resource
func (d *Database) GetAuditLogsByResource(resourceType, resourceID string, limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := d.db.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetAuditLogsByCategory retrieves audit logs by category
func (d *Database) GetAuditLogsByCategory(category string, limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := d.db.Where("category = ?", category).Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetHighRiskAuditLogs retrieves high-risk audit logs
func (d *Database) GetHighRiskAuditLogs(limit, offset int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := d.db.Where("risk_level IN ('high', 'critical') OR success = false").
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// CreateSecurityAlert creates a new security alert
func (d *Database) CreateSecurityAlert(alert *models.SecurityAlert) error {
	return d.db.Create(alert).Error
}

// GetSecurityAlerts retrieves security alerts with filtering
func (d *Database) GetSecurityAlerts(limit, offset int, filters map[string]interface{}) ([]*models.SecurityAlert, int64, error) {
	var alerts []*models.SecurityAlert
	var total int64

	query := d.db.Model(&models.SecurityAlert{}).Preload("User").Preload("ResolvedByUser")

	// Apply filters
	if alertType, ok := filters["alert_type"]; ok {
		query = query.Where("alert_type = ?", alertType)
	}
	if severity, ok := filters["severity"]; ok {
		query = query.Where("severity = ?", severity)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if from, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", from)
	}
	if to, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", to)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error
	return alerts, total, err
}

// UpdateSecurityAlert updates a security alert
func (d *Database) UpdateSecurityAlert(alertID uint, updates map[string]interface{}) error {
	return d.db.Model(&models.SecurityAlert{}).Where("id = ?", alertID).Updates(updates).Error
}

// ResolveSecurityAlert marks a security alert as resolved
func (d *Database) ResolveSecurityAlert(alertID uint, resolvedBy uint, resolution string) error {
	now := time.Now()
	return d.db.Model(&models.SecurityAlert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_by": resolvedBy,
		"resolved_at": &now,
		"resolution":  resolution,
		"updated_at":  now,
	}).Error
}

// CreateComplianceReport creates a new compliance report
func (d *Database) CreateComplianceReport(report *models.ComplianceReport) error {
	return d.db.Create(report).Error
}

// GetComplianceReports retrieves compliance reports
func (d *Database) GetComplianceReports(limit, offset int) ([]*models.ComplianceReport, error) {
	var reports []*models.ComplianceReport
	err := d.db.Preload("GeneratedByUser").Limit(limit).Offset(offset).Order("created_at DESC").Find(&reports).Error
	return reports, err
}

// UpdateComplianceReport updates a compliance report
func (d *Database) UpdateComplianceReport(reportID uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return d.db.Model(&models.ComplianceReport{}).Where("id = ?", reportID).Updates(updates).Error
}

// CleanupOldAuditLogs removes old audit logs
func (d *Database) CleanupOldAuditLogs(olderThan time.Time) error {
	return d.db.Where("created_at < ?", olderThan).Delete(&models.AuditLog{}).Error
}

// GetComplianceData retrieves compliance-related data for reporting
func (d *Database) GetComplianceData(reportType string, startDate, endDate time.Time) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get basic statistics
	var userCount, activeUserCount int64
	d.db.Model(&models.User{}).Count(&userCount)
	d.db.Model(&models.User{}).Where("is_active = true").Count(&activeUserCount)

	// Get audit log statistics
	var auditLogCount, failedActionCount int64
	auditQuery := d.db.Model(&models.AuditLog{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)
	auditQuery.Count(&auditLogCount)
	auditQuery.Where("success = false").Count(&failedActionCount)

	// Get security alert statistics
	var securityAlertCount int64
	d.db.Model(&models.SecurityAlert{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&securityAlertCount)

	data["report_type"] = reportType
	data["period"] = map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	}
	data["user_statistics"] = map[string]interface{}{
		"total_users":  userCount,
		"active_users": activeUserCount,
	}
	data["audit_statistics"] = map[string]interface{}{
		"total_audit_logs": auditLogCount,
		"failed_actions":   failedActionCount,
	}
	data["security_statistics"] = map[string]interface{}{
		"security_alerts": securityAlertCount,
	}

	return data, nil
}

// SearchAuditLogs performs text search on audit logs
func (d *Database) SearchAuditLogs(query string, filters map[string]interface{}, limit, offset int) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64

	dbQuery := d.db.Model(&models.AuditLog{}).Preload("User")

	// Text search across multiple fields
	if query != "" {
		searchTerm := "%" + query + "%"
		dbQuery = dbQuery.Where("action LIKE ? OR resource_type LIKE ? OR error_message LIKE ? OR request_path LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Apply additional filters
	if userID, ok := filters["user_id"]; ok {
		dbQuery = dbQuery.Where("user_id = ?", userID)
	}
	if category, ok := filters["category"]; ok {
		dbQuery = dbQuery.Where("category = ?", category)
	}
	if riskLevel, ok := filters["risk_level"]; ok {
		dbQuery = dbQuery.Where("risk_level = ?", riskLevel)
	}

	// Get total count
	dbQuery.Count(&total)

	// Get paginated results
	err := dbQuery.Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
