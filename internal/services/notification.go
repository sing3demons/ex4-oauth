package services

import (
	"fmt"
	"time"

	"ex4-oauth2/internal/models"

	"github.com/google/uuid"
)

// NotificationService handles notification creation and management
type NotificationService struct {
	notificationRepo models.NotificationRepository
	websocketService *WebSocketService
	auditService     *AuditService
}

// NewNotificationService creates a new notification service
func NewNotificationService(notificationRepo models.NotificationRepository, websocketService *WebSocketService, auditService *AuditService) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		websocketService: websocketService,
		auditService:     auditService,
	}
}

// CreateSecurityAlert creates and sends a security alert notification
func (ns *NotificationService) CreateSecurityAlert(title, message string, userID *uint, severity string, data map[string]interface{}) {
	priority := models.PriorityHigh
	if severity == "critical" {
		priority = models.PriorityCritical
	}

	notification := &models.Notification{
		ID:            uuid.New().String(),
		Type:          models.NotificationSecurityAlert,
		Priority:      priority,
		Title:         title,
		Message:       message,
		Data:          data,
		UserID:        userID,
		RecipientRole: "admin", // Security alerts go to admins
		CreatedAt:     time.Now(),
	}

	if data != nil {
		if sourceIP, ok := data["ip_address"].(string); ok {
			notification.SourceIP = sourceIP
		}
		if userAgent, ok := data["user_agent"].(string); ok {
			notification.UserAgent = userAgent
		}
		if sessionID, ok := data["session_id"].(string); ok {
			notification.SessionID = sessionID
		}
	}

	ns.sendNotification(notification)
}

// CreateUserNotification creates a notification for a specific user
func (ns *NotificationService) CreateUserNotification(userID uint, notificationType models.NotificationType, title, message string, data map[string]interface{}) {
	notification := &models.Notification{
		ID:        uuid.New().String(),
		Type:      notificationType,
		Priority:  models.PriorityNormal,
		Title:     title,
		Message:   message,
		Data:      data,
		UserID:    &userID,
		CreatedAt: time.Now(),
	}

	// Set expiry for user notifications (7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	notification.ExpiresAt = &expiresAt

	ns.sendNotification(notification)
}

// CreateAdminNotification creates a notification for administrators
func (ns *NotificationService) CreateAdminNotification(notificationType models.NotificationType, title, message string, data map[string]interface{}) {
	notification := &models.Notification{
		ID:            uuid.New().String(),
		Type:          notificationType,
		Priority:      models.PriorityNormal,
		Title:         title,
		Message:       message,
		Data:          data,
		RecipientRole: "admin",
		CreatedAt:     time.Now(),
	}

	ns.sendNotification(notification)
}

// CreateSystemNotification creates a system-wide notification
func (ns *NotificationService) CreateSystemNotification(notificationType models.NotificationType, title, message string, priority models.NotificationPriority, data map[string]interface{}) {
	notification := &models.Notification{
		ID:        uuid.New().String(),
		Type:      notificationType,
		Priority:  priority,
		Title:     title,
		Message:   message,
		Data:      data,
		CreatedAt: time.Now(),
	}

	ns.sendNotification(notification)
}

// CreateComplianceAlert creates a compliance-related alert
func (ns *NotificationService) CreateComplianceAlert(title, message string, findings map[string]interface{}) {
	notification := &models.Notification{
		ID:            uuid.New().String(),
		Type:          models.NotificationComplianceAlert,
		Priority:      models.PriorityHigh,
		Title:         title,
		Message:       message,
		Data:          findings,
		RecipientRole: "admin",
		CreatedAt:     time.Now(),
	}

	ns.sendNotification(notification)
}

// NotifyUserRegistration notifies admins of new user registration
func (ns *NotificationService) NotifyUserRegistration(user *models.User, requestIP string) {
	data := map[string]interface{}{
		"user_id":    user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"provider":   user.Provider,
		"ip_address": requestIP,
		"timestamp":  time.Now(),
	}

	title := "New User Registration"
	message := fmt.Sprintf("New user registered: %s (%s)", user.Username, user.Email)

	ns.CreateAdminNotification(models.NotificationUserRegistered, title, message, data)
}

// NotifyLoginAttempt notifies of login attempts (especially failed ones)
func (ns *NotificationService) NotifyLoginAttempt(email string, success bool, requestIP, userAgent string) {
	var notificationType models.NotificationType
	var title, message string
	var priority models.NotificationPriority

	if success {
		return // Don't notify for successful logins unless suspicious
	}

	notificationType = models.NotificationLoginAttempt
	title = "Failed Login Attempt"
	message = fmt.Sprintf("Failed login attempt for email: %s from IP: %s", email, requestIP)
	priority = models.PriorityNormal

	data := map[string]interface{}{
		"email":      email,
		"ip_address": requestIP,
		"user_agent": userAgent,
		"success":    success,
		"timestamp":  time.Now(),
	}

	notification := &models.Notification{
		ID:            uuid.New().String(),
		Type:          notificationType,
		Priority:      priority,
		Title:         title,
		Message:       message,
		Data:          data,
		SourceIP:      requestIP,
		UserAgent:     userAgent,
		RecipientRole: "admin",
		CreatedAt:     time.Now(),
	}

	ns.sendNotification(notification)
}

// NotifySuspiciousActivity notifies of suspicious activities
func (ns *NotificationService) NotifySuspiciousActivity(userID *uint, activityType, description string, metadata map[string]interface{}) {
	title := "Suspicious Activity Detected"
	message := fmt.Sprintf("Suspicious activity: %s - %s", activityType, description)

	data := map[string]interface{}{
		"activity_type": activityType,
		"description":   description,
		"metadata":      metadata,
		"timestamp":     time.Now(),
	}

	if userID != nil {
		data["user_id"] = *userID
	}

	notification := &models.Notification{
		ID:            uuid.New().String(),
		Type:          models.NotificationSuspiciousLogin,
		Priority:      models.PriorityHigh,
		Title:         title,
		Message:       message,
		Data:          data,
		UserID:        userID,
		RecipientRole: "admin",
		CreatedAt:     time.Now(),
	}

	ns.sendNotification(notification)
}

// NotifyPasswordChange notifies user of password change
func (ns *NotificationService) NotifyPasswordChange(userID uint, requestIP string) {
	title := "Password Changed"
	message := "Your password has been successfully changed."

	data := map[string]interface{}{
		"ip_address": requestIP,
		"timestamp":  time.Now(),
	}

	ns.CreateUserNotification(userID, models.NotificationPasswordChange, title, message, data)
}

// NotifyEmailVerified notifies user of email verification
func (ns *NotificationService) NotifyEmailVerified(userID uint) {
	title := "Email Verified"
	message := "Your email address has been successfully verified."

	data := map[string]interface{}{
		"timestamp": time.Now(),
	}

	ns.CreateUserNotification(userID, models.NotificationEmailVerified, title, message, data)
}

// NotifyTokenExpiry notifies user of upcoming token expiry
func (ns *NotificationService) NotifyTokenExpiry(userID uint, tokenType string, expiresAt time.Time) {
	title := "Token Expiring Soon"
	message := fmt.Sprintf("Your %s will expire on %s", tokenType, expiresAt.Format("2006-01-02 15:04:05"))

	data := map[string]interface{}{
		"token_type": tokenType,
		"expires_at": expiresAt,
		"timestamp":  time.Now(),
	}

	ns.CreateUserNotification(userID, models.NotificationTokenExpiry, title, message, data)
}

// NotifySystemMaintenance notifies all users of system maintenance
func (ns *NotificationService) NotifySystemMaintenance(title, message string, scheduledTime time.Time) {
	data := map[string]interface{}{
		"scheduled_time": scheduledTime,
		"timestamp":      time.Now(),
	}

	ns.CreateSystemNotification(models.NotificationSystemMaintenance, title, message, models.PriorityHigh, data)
}

// sendNotification sends notification through WebSocket and persists if needed
func (ns *NotificationService) sendNotification(notification *models.Notification) {
	// Send real-time notification
	if ns.websocketService != nil {
		ns.websocketService.SendNotification(notification)
	}

	// Log notification creation
	if ns.auditService != nil {
		ctx := &AuditContext{
			UserID: notification.UserID,
		}
		ns.auditService.LogAction(ctx, "notification_created", "notification", string(notification.Type), map[string]interface{}{
			"notification_id": notification.ID,
			"priority":        notification.Priority,
			"recipient_role":  notification.RecipientRole,
		}, nil)
	}
}

// GetNotifications retrieves notifications for a user
func (ns *NotificationService) GetNotifications(userID uint, filter *models.NotificationFilter) ([]*models.PersistentNotification, int64, error) {
	if filter == nil {
		filter = &models.NotificationFilter{}
	}

	filter.UserID = &userID
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	return ns.notificationRepo.GetNotifications(filter)
}

// GetAdminNotifications retrieves notifications for administrators
func (ns *NotificationService) GetAdminNotifications(filter *models.NotificationFilter) ([]*models.PersistentNotification, int64, error) {
	if filter == nil {
		filter = &models.NotificationFilter{}
	}

	if filter.Limit == 0 {
		filter.Limit = 100
	}

	return ns.notificationRepo.GetNotifications(filter)
}

// MarkAsRead marks a notification as read
func (ns *NotificationService) MarkAsRead(notificationID uint, userID uint) error {
	err := ns.notificationRepo.MarkAsRead(notificationID, userID)
	if err != nil {
		return err
	}

	// Log read action
	if ns.auditService != nil {
		ctx := &AuditContext{
			UserID: &userID,
		}
		ns.auditService.LogAction(ctx, "notification_read", "notification", fmt.Sprintf("%d", notificationID), nil, nil)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (ns *NotificationService) MarkAllAsRead(userID uint) error {
	err := ns.notificationRepo.MarkAllAsRead(userID)
	if err != nil {
		return err
	}

	// Log bulk read action
	if ns.auditService != nil {
		ctx := &AuditContext{
			UserID: &userID,
		}
		ns.auditService.LogAction(ctx, "notifications_mark_all_read", "notification", "bulk", nil, nil)
	}

	return nil
}

// GetNotificationStats returns notification statistics
func (ns *NotificationService) GetNotificationStats(userID *uint) (*models.NotificationStats, error) {
	return ns.notificationRepo.GetNotificationStats(userID)
}

// DeleteNotification deletes a notification
func (ns *NotificationService) DeleteNotification(notificationID uint, userID uint) error {
	// Check if notification belongs to user
	notification, err := ns.notificationRepo.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	if notification.UserID == nil || *notification.UserID != userID {
		return fmt.Errorf("notification not found or access denied")
	}

	err = ns.notificationRepo.DeleteNotification(notificationID)
	if err != nil {
		return err
	}

	// Log deletion
	if ns.auditService != nil {
		ctx := &AuditContext{
			UserID: &userID,
		}
		ns.auditService.LogAction(ctx, "notification_deleted", "notification", fmt.Sprintf("%d", notificationID), nil, nil)
	}

	return nil
}

// ProcessSecurityEvents processes security events and creates appropriate notifications
func (ns *NotificationService) ProcessSecurityEvents() {
	// This could be called periodically to process security events
	// and create notifications based on patterns, thresholds, etc.

	// Example: Check for multiple failed login attempts
	// Example: Check for unusual access patterns
	// Example: Check for privilege escalation attempts

	// Implementation would depend on specific security requirements
}

// CreateNotificationFromAuditLog creates notifications based on audit log events
func (ns *NotificationService) CreateNotificationFromAuditLog(auditLog *models.AuditLog) {
	// Create notifications based on audit log events
	switch auditLog.Action {
	case "login_failed":
		if auditLog.Category == "high_risk" {
			ns.NotifySuspiciousActivity(auditLog.UserID, "Multiple Failed Logins",
				"Multiple failed login attempts detected", map[string]interface{}{
					"audit_log_id": auditLog.ID,
					"ip_address":   auditLog.IPAddress,
					"user_agent":   auditLog.UserAgent,
				})
		}

	case "user_created":
		if user := auditLog.User; user != nil {
			ns.NotifyUserRegistration(user, auditLog.IPAddress)
		}

	case "privilege_escalation":
		ns.CreateSecurityAlert("Privilege Escalation Detected",
			"Unauthorized privilege escalation attempt", auditLog.UserID, "critical", map[string]interface{}{
				"audit_log_id":  auditLog.ID,
				"resource_type": auditLog.ResourceType,
				"resource_id":   auditLog.ResourceID,
			})
	}
}
