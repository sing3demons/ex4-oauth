package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ex4-oauth2/internal/models"
	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	adminService *services.AdminService
	userRepo     models.UserRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *services.AdminService, userRepo models.UserRepository) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		userRepo:     userRepo,
	}
}

// UpdateUserStatusRequest represents the request for updating user status
type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

// UpdateUserRoleRequest represents the request for updating user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user moderator admin"`
}

// GetDashboardStats returns admin dashboard statistics
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch dashboard stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetUsers returns paginated user list with filters
func (h *AdminHandler) GetUsers(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse filters
	filters := make(map[string]interface{})

	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters["is_active"] = isActive
		}
	}

	if emailVerifiedStr := c.Query("email_verified"); emailVerifiedStr != "" {
		if emailVerified, err := strconv.ParseBool(emailVerifiedStr); err == nil {
			filters["email_verified"] = emailVerified
		}
	}

	if provider := c.Query("provider"); provider != "" {
		filters["provider"] = provider
	}

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	users, total, err := h.adminService.GetUsers(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUserActivity returns user activity logs
func (h *AdminHandler) GetUserActivity(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	activities, err := h.adminService.GetUserActivity(uint(userID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch user activity",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"user_id":    userID,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetSystemEvents returns system event logs
func (h *AdminHandler) GetSystemEvents(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse filters
	filters := make(map[string]interface{})

	if eventType := c.Query("event_type"); eventType != "" {
		filters["event_type"] = eventType
	}

	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	if fromDate := c.Query("from_date"); fromDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", fromDate); err == nil {
			filters["from_date"] = parsedDate
		}
	}

	if toDate := c.Query("to_date"); toDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", toDate); err == nil {
			filters["to_date"] = parsedDate.Add(24*time.Hour - time.Second) // End of day
		}
	}

	events, err := h.adminService.GetSystemEvents(limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch system events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":  events,
		"limit":   limit,
		"offset":  offset,
		"filters": filters,
	})
}

// UpdateUserStatus updates user active status
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Update user status
	if err := h.adminService.UpdateUserStatus(uint(userID), req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user status",
			"details": err.Error(),
		})
		return
	}

	// Log admin activity
	adminUserID := c.GetUint("user_id")
	h.adminService.LogUserActivity(
		adminUserID,
		"admin_update_user_status",
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		true,
		map[string]interface{}{
			"target_user_id": userID,
			"is_active":      req.IsActive,
		},
		"",
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "User status updated successfully",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"is_active": req.IsActive,
		},
	})
}

// UpdateUserRole updates user role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Prevent self-demotion for admins
	adminUserID := c.GetUint("user_id")
	if adminUserID == uint(userID) && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot change your own admin role",
		})
		return
	}

	// Update user role
	if err := h.adminService.UpdateUserRole(uint(userID), req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user role",
			"details": err.Error(),
		})
		return
	}

	// Log admin activity
	h.adminService.LogUserActivity(
		adminUserID,
		"admin_update_user_role",
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		true,
		map[string]interface{}{
			"target_user_id": userID,
			"new_role":       req.Role,
			"old_role":       user.Role,
		},
		"",
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  req.Role,
		},
	})
}

// DeleteUser soft deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Prevent self-deletion for admins
	adminUserID := c.GetUint("user_id")
	if adminUserID == uint(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete your own account",
		})
		return
	}

	// Delete user
	if err := h.adminService.DeleteUser(uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": err.Error(),
		})
		return
	}

	// Log admin activity
	h.adminService.LogUserActivity(
		adminUserID,
		"admin_delete_user",
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		true,
		map[string]interface{}{
			"target_user_id": userID,
			"target_email":   user.Email,
		},
		"",
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// CleanupLogs removes old activity logs and events
func (h *AdminHandler) CleanupLogs(c *gin.Context) {
	retentionDaysStr := c.DefaultQuery("retention_days", "90")
	retentionDays, err := strconv.Atoi(retentionDaysStr)
	if err != nil || retentionDays < 1 {
		retentionDays = 90
	}

	if err := h.adminService.CleanupOldLogs(retentionDays); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup logs",
			"details": err.Error(),
		})
		return
	}

	// Log admin activity
	adminUserID := c.GetUint("user_id")
	h.adminService.LogUserActivity(
		adminUserID,
		"admin_cleanup_logs",
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		true,
		map[string]interface{}{
			"retention_days": retentionDays,
		},
		"",
	)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Logs cleaned up successfully",
		"retention_days": retentionDays,
	})
}

// GetSystemInfo returns system information
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"system": gin.H{
			"name":        "OAuth2 Authorization Server",
			"version":     "1.0.0",
			"environment": "development", // Should come from config
			"timestamp":   time.Now().Unix(),
		},
		"admin_features": []string{
			"User Management",
			"Role Management",
			"Activity Logging",
			"System Events",
			"Dashboard Statistics",
			"Log Cleanup",
		},
		"valid_roles":       h.adminService.GetValidRoles(),
		"valid_event_types": h.adminService.GetValidEventTypes(),
		"valid_severities":  h.adminService.GetValidSeverityLevels(),
	})
}
