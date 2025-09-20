package models

import (
	"time"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	// Security notifications
	NotificationSecurityAlert   NotificationType = "security_alert"
	NotificationLoginAttempt    NotificationType = "login_attempt"
	NotificationSuspiciousLogin NotificationType = "suspicious_login"
	NotificationAccountLocked   NotificationType = "account_locked"

	// Admin notifications
	NotificationUserRegistered   NotificationType = "user_registered"
	NotificationUserStatusChange NotificationType = "user_status_change"
	NotificationSystemAlert      NotificationType = "system_alert"
	NotificationComplianceAlert  NotificationType = "compliance_alert"

	// User notifications
	NotificationProfileUpdate  NotificationType = "profile_update"
	NotificationPasswordChange NotificationType = "password_change"
	NotificationEmailVerified  NotificationType = "email_verified"
	NotificationTokenExpiry    NotificationType = "token_expiry"

	// System notifications
	NotificationSystemMaintenance NotificationType = "system_maintenance"
	NotificationServiceUpdate     NotificationType = "service_update"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// Notification represents a real-time notification
type Notification struct {
	ID            string                 `json:"id"`
	Type          NotificationType       `json:"type"`
	Priority      NotificationPriority   `json:"priority"`
	Title         string                 `json:"title"`
	Message       string                 `json:"message"`
	Data          map[string]interface{} `json:"data,omitempty"`
	UserID        *uint                  `json:"user_id,omitempty"`        // nil for broadcast
	RecipientRole string                 `json:"recipient_role,omitempty"` // "admin", "user", etc.
	CreatedAt     time.Time              `json:"created_at"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty"`
	Read          bool                   `json:"read"`

	// Optional metadata
	SourceIP  string `json:"source_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// WebSocketMessage represents a WebSocket message structure
type WebSocketMessage struct {
	Type      string      `json:"type"`              // "notification", "ping", "pong", "subscribe", "unsubscribe"
	Channel   string      `json:"channel,omitempty"` // "security", "admin", "user", "system"
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	MessageID string      `json:"message_id"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID          string    `json:"id"`
	UserID      *uint     `json:"user_id,omitempty"`
	Role        string    `json:"role"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Channels    []string  `json:"channels"` // subscribed channels
	ConnectedAt time.Time `json:"connected_at"`
	LastPingAt  time.Time `json:"last_ping_at"`
	IsActive    bool      `json:"is_active"`
}

// NotificationChannel represents different notification channels
type NotificationChannel string

const (
	ChannelSecurity   NotificationChannel = "security"
	ChannelAdmin      NotificationChannel = "admin"
	ChannelUser       NotificationChannel = "user"
	ChannelSystem     NotificationChannel = "system"
	ChannelCompliance NotificationChannel = "compliance"
	ChannelAudit      NotificationChannel = "audit"
)

// NotificationFilter represents filtering options for notifications
type NotificationFilter struct {
	UserID    *uint                `json:"user_id,omitempty"`
	Type      NotificationType     `json:"type,omitempty"`
	Priority  NotificationPriority `json:"priority,omitempty"`
	Channel   NotificationChannel  `json:"channel,omitempty"`
	Unread    *bool                `json:"unread,omitempty"`
	StartDate *time.Time           `json:"start_date,omitempty"`
	EndDate   *time.Time           `json:"end_date,omitempty"`
	Limit     int                  `json:"limit,omitempty"`
	Offset    int                  `json:"offset,omitempty"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalNotifications      int64                          `json:"total_notifications"`
	UnreadNotifications     int64                          `json:"unread_notifications"`
	NotificationsByType     map[NotificationType]int64     `json:"notifications_by_type"`
	NotificationsByPriority map[NotificationPriority]int64 `json:"notifications_by_priority"`
	ActiveConnections       int                            `json:"active_connections"`
	ConnectedUsers          int                            `json:"connected_users"`
	ChannelSubscriptions    map[NotificationChannel]int    `json:"channel_subscriptions"`
}

// PersistentNotification represents notifications stored in database
type PersistentNotification struct {
	ID            uint                 `json:"id" gorm:"primaryKey"`
	Type          NotificationType     `json:"type" gorm:"not null"`
	Priority      NotificationPriority `json:"priority" gorm:"not null"`
	Title         string               `json:"title" gorm:"not null"`
	Message       string               `json:"message" gorm:"type:text;not null"`
	DataJSON      string               `json:"data_json" gorm:"type:text"`
	UserID        *uint                `json:"user_id"`
	RecipientRole string               `json:"recipient_role"`
	SourceIP      string               `json:"source_ip"`
	UserAgent     string               `json:"user_agent"`
	SessionID     string               `json:"session_id"`
	Read          bool                 `json:"read" gorm:"default:false"`
	ReadAt        *time.Time           `json:"read_at"`
	CreatedAt     time.Time            `json:"created_at"`
	ExpiresAt     *time.Time           `json:"expires_at"`
	User          *User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	CreateNotification(notification *PersistentNotification) error
	GetNotifications(filter *NotificationFilter) ([]*PersistentNotification, int64, error)
	GetNotificationByID(id uint) (*PersistentNotification, error)
	MarkAsRead(notificationID uint, userID uint) error
	MarkAllAsRead(userID uint) error
	DeleteNotification(id uint) error
	DeleteExpiredNotifications() error
	GetNotificationStats(userID *uint) (*NotificationStats, error)
}
