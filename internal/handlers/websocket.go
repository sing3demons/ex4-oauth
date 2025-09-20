package handlers

import (
	"net/http"
	"strconv"

	"ex4-oauth2/internal/models"
	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler handles WebSocket connections and notifications
type WebSocketHandler struct {
	websocketService    *services.WebSocketService
	notificationService *services.NotificationService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(websocketService *services.WebSocketService, notificationService *services.NotificationService) *WebSocketHandler {
	return &WebSocketHandler{
		websocketService:    websocketService,
		notificationService: notificationService,
	}
}

// HandleWebSocketConnection handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocketConnection(c *gin.Context) {
	// Get user info from context (should be set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	roleInterface, exists := c.Get("user_role")
	role := "user"
	if exists {
		if userRole, ok := roleInterface.(string); ok {
			role = userRole
		}
	}

	// Handle WebSocket upgrade
	err := h.websocketService.HandleConnection(c, &userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to establish WebSocket connection",
			"details": err.Error(),
		})
		return
	}
}

// HandleAnonymousWebSocket handles WebSocket connections for anonymous users (public notifications)
func (h *WebSocketHandler) HandleAnonymousWebSocket(c *gin.Context) {
	// Handle WebSocket upgrade for anonymous users (limited channels)
	err := h.websocketService.HandleConnection(c, nil, "anonymous")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to establish WebSocket connection",
			"details": err.Error(),
		})
		return
	}
}

// GetNotifications handles GET /api/notifications
func (h *WebSocketHandler) GetNotifications(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse query parameters
	filter := &models.NotificationFilter{}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		filter.Type = models.NotificationType(typeStr)
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		filter.Priority = models.NotificationPriority(priorityStr)
	}

	if unreadStr := c.Query("unread"); unreadStr != "" {
		if unread, err := strconv.ParseBool(unreadStr); err == nil {
			filter.Unread = &unread
		}
	}

	// Get notifications
	notifications, total, err := h.notificationService.GetNotifications(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve notifications",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"pagination": gin.H{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// GetAdminNotifications handles GET /api/admin/notifications
func (h *WebSocketHandler) GetAdminNotifications(c *gin.Context) {
	// Parse query parameters
	filter := &models.NotificationFilter{}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		filter.Type = models.NotificationType(typeStr)
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		filter.Priority = models.NotificationPriority(priorityStr)
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		filter.Channel = models.NotificationChannel(channelStr)
	}

	// Get admin notifications
	notifications, total, err := h.notificationService.GetAdminNotifications(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve notifications",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"pagination": gin.H{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// MarkNotificationAsRead handles PUT /api/notifications/:id/read
func (h *WebSocketHandler) MarkNotificationAsRead(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkAsRead(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark notification as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsAsRead handles PUT /api/notifications/read-all
func (h *WebSocketHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark all notifications as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification handles DELETE /api/notifications/:id
func (h *WebSocketHandler) DeleteNotification(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.DeleteNotification(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}

// GetNotificationStats handles GET /api/notifications/stats
func (h *WebSocketHandler) GetNotificationStats(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	stats, err := h.notificationService.GetNotificationStats(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get notification statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetAdminNotificationStats handles GET /api/admin/notifications/stats
func (h *WebSocketHandler) GetAdminNotificationStats(c *gin.Context) {
	stats, err := h.notificationService.GetNotificationStats(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get notification statistics",
			"details": err.Error(),
		})
		return
	}

	// Add WebSocket statistics
	wsStats := h.websocketService.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"notifications": stats,
			"websocket":     wsStats,
		},
	})
}

// TestNotification handles POST /api/admin/notifications/test (for testing)
func (h *WebSocketHandler) TestNotification(c *gin.Context) {
	var request struct {
		Type     string                 `json:"type" binding:"required"`
		Title    string                 `json:"title" binding:"required"`
		Message  string                 `json:"message" binding:"required"`
		Priority string                 `json:"priority"`
		UserID   *uint                  `json:"user_id"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := models.PriorityNormal
	if request.Priority != "" {
		priority = models.NotificationPriority(request.Priority)
	}

	// Create test notification
	if request.UserID != nil {
		h.notificationService.CreateUserNotification(*request.UserID,
			models.NotificationType(request.Type), request.Title, request.Message, request.Data)
	} else {
		h.notificationService.CreateSystemNotification(
			models.NotificationType(request.Type), request.Title, request.Message, priority, request.Data)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test notification sent"})
}
