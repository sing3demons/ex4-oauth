package database

import (
	"time"

	"ex4-oauth2/internal/models"

	"gorm.io/gorm"
)

// NotificationRepository implements notification data access
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// CreateNotification creates a new notification
func (r *NotificationRepository) CreateNotification(notification *models.PersistentNotification) error {
	return r.db.Create(notification).Error
}

// GetNotifications retrieves notifications with filtering
func (r *NotificationRepository) GetNotifications(filter *models.NotificationFilter) ([]*models.PersistentNotification, int64, error) {
	var notifications []*models.PersistentNotification
	var total int64

	query := r.db.Model(&models.PersistentNotification{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ? OR (user_id IS NULL AND recipient_role IS NOT NULL)", *filter.UserID)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}

	if filter.Unread != nil {
		query = query.Where("read = ?", !*filter.Unread)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// Count total
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Order by creation time (newest first)
	query = query.Order("created_at DESC")

	// Include user information
	query = query.Preload("User")

	err := query.Find(&notifications).Error
	return notifications, total, err
}

// GetNotificationByID retrieves a notification by ID
func (r *NotificationRepository) GetNotificationByID(id uint) (*models.PersistentNotification, error) {
	var notification models.PersistentNotification
	err := r.db.Preload("User").First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(notificationID uint, userID uint) error {
	now := time.Now()
	return r.db.Model(&models.PersistentNotification{}).
		Where("id = ? AND (user_id = ? OR user_id IS NULL)", notificationID, userID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": &now,
		}).Error
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	now := time.Now()
	return r.db.Model(&models.PersistentNotification{}).
		Where("user_id = ? AND read = false", userID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": &now,
		}).Error
}

// DeleteNotification deletes a notification
func (r *NotificationRepository) DeleteNotification(id uint) error {
	return r.db.Delete(&models.PersistentNotification{}, id).Error
}

// DeleteExpiredNotifications deletes expired notifications
func (r *NotificationRepository) DeleteExpiredNotifications() error {
	return r.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.PersistentNotification{}).Error
}

// GetNotificationStats returns notification statistics
func (r *NotificationRepository) GetNotificationStats(userID *uint) (*models.NotificationStats, error) {
	stats := &models.NotificationStats{
		NotificationsByType:     make(map[models.NotificationType]int64),
		NotificationsByPriority: make(map[models.NotificationPriority]int64),
		ChannelSubscriptions:    make(map[models.NotificationChannel]int),
	}

	query := r.db.Model(&models.PersistentNotification{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// Total notifications
	if err := query.Count(&stats.TotalNotifications).Error; err != nil {
		return nil, err
	}

	// Unread notifications
	unreadQuery := query.Where("read = false")
	if err := unreadQuery.Count(&stats.UnreadNotifications).Error; err != nil {
		return nil, err
	}

	// Notifications by type
	var typeStats []struct {
		Type  models.NotificationType
		Count int64
	}

	if err := query.Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range typeStats {
		stats.NotificationsByType[stat.Type] = stat.Count
	}

	// Notifications by priority
	var priorityStats []struct {
		Priority models.NotificationPriority
		Count    int64
	}

	if err := query.Select("priority, COUNT(*) as count").
		Group("priority").
		Scan(&priorityStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range priorityStats {
		stats.NotificationsByPriority[stat.Priority] = stat.Count
	}

	return stats, nil
}
