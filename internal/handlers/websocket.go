package handlers

import (
	"net/http"
	"strconv"

	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler handles websocket and notification operations (simplified)
type WebSocketHandler struct {
	websocketService     *services.WebSocketService
	notificationService  *services.NotificationService
}

// NewWebSocketHandler creates a new websocket handler
func NewWebSocketHandler(websocketService *services.WebSocketService, notificationService *services.NotificationService) *WebSocketHandler {
	return &WebSocketHandler{
		websocketService:     websocketService,
		notificationService:  notificationService,
	}
}

// HandleWebSocket handles websocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	h.websocketService.HandleWebSocket(c)
}

// GetNotifications returns user notifications
func (h *WebSocketHandler) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get pagination parameters
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get notifications
	notifications, err := h.notificationService.GetUserNotifications(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get notifications",
			"details": err.Error(),
		})
		return
	}

	// Get unread count
	unreadCount, _ := h.notificationService.GetUnreadCount(userID)

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(notifications),
		},
	})
}

// MarkNotificationAsRead marks a notification as read
func (h *WebSocketHandler) MarkNotificationAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Notification ID is required",
		})
		return
	}

	if err := h.notificationService.MarkNotificationAsRead(notificationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark notification as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// DeleteNotification deletes a notification
func (h *WebSocketHandler) DeleteNotification(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Notification ID is required",
		})
		return
	}

	if err := h.notificationService.DeleteNotification(notificationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification deleted",
	})
}

// GetConnectionStats returns websocket connection statistics
func (h *WebSocketHandler) GetConnectionStats(c *gin.Context) {
	stats := h.websocketService.GetConnectionStats()

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// BroadcastMessage broadcasts a message to all connected users (admin only)
func (h *WebSocketHandler) BroadcastMessage(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.websocketService.BroadcastToAll([]byte(req.Message)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to broadcast message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message broadcasted successfully",
	})
}