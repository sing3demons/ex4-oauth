package services

import (
	"encoding/json"
	"fmt"
	"time"

	"ex4-oauth2/internal/models"
)

// AdminService handles admin operations
type AdminService struct {
	adminRepository models.AdminRepository
	userRepository  models.UserRepository
}

// NewAdminService creates a new admin service
func NewAdminService(
	adminRepo models.AdminRepository,
	userRepo models.UserRepository,
) *AdminService {
	return &AdminService{
		adminRepository: adminRepo,
		userRepository:  userRepo,
	}
}

// GetDashboardStats returns dashboard statistics
func (s *AdminService) GetDashboardStats() (*models.AdminStats, error) {
	return s.adminRepository.GetStats()
}

// GetUsers returns paginated user list with filters
func (s *AdminService) GetUsers(limit, offset int, filters map[string]interface{}) ([]*models.User, int64, error) {
	return s.adminRepository.GetUsers(limit, offset, filters)
}

// GetUserActivity returns user activity logs
func (s *AdminService) GetUserActivity(userID uint, limit, offset int) ([]*models.UserActivity, error) {
	return s.adminRepository.GetUserActivity(userID, limit, offset)
}

// GetSystemEvents returns system event logs
func (s *AdminService) GetSystemEvents(limit, offset int, filters map[string]interface{}) ([]*models.SystemEvent, error) {
	return s.adminRepository.GetSystemEvents(limit, offset, filters)
}

// UpdateUserStatus updates user active status
func (s *AdminService) UpdateUserStatus(userID uint, isActive bool) error {
	// Log the admin action
	details := map[string]interface{}{
		"user_id":   userID,
		"is_active": isActive,
	}
	detailsJSON, _ := json.Marshal(details)

	event := &models.SystemEvent{
		EventType: "admin",
		Severity:  "info",
		Message:   fmt.Sprintf("User %d status changed to active: %v", userID, isActive),
		Details:   string(detailsJSON),
	}

	// Create system event (non-blocking)
	go s.adminRepository.CreateSystemEvent(event)

	return s.adminRepository.UpdateUserStatus(userID, isActive)
}

// UpdateUserRole updates user role
func (s *AdminService) UpdateUserRole(userID uint, role string) error {
	// Validate role
	validRoles := map[string]bool{
		"user":      true,
		"moderator": true,
		"admin":     true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Log the admin action
	details := map[string]interface{}{
		"user_id":  userID,
		"new_role": role,
	}
	detailsJSON, _ := json.Marshal(details)

	event := &models.SystemEvent{
		EventType: "admin",
		Severity:  "info",
		Message:   fmt.Sprintf("User %d role changed to: %s", userID, role),
		Details:   string(detailsJSON),
	}

	// Create system event (non-blocking)
	go s.adminRepository.CreateSystemEvent(event)

	return s.adminRepository.UpdateUserRole(userID, role)
}

// DeleteUser soft deletes a user
func (s *AdminService) DeleteUser(userID uint) error {
	// Get user details for logging
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Log the admin action
	details := map[string]interface{}{
		"user_id":  userID,
		"email":    user.Email,
		"username": user.Username,
	}
	detailsJSON, _ := json.Marshal(details)

	event := &models.SystemEvent{
		EventType: "admin",
		Severity:  "warning",
		Message:   fmt.Sprintf("User %d (%s) deleted", userID, user.Email),
		Details:   string(detailsJSON),
	}

	// Create system event (non-blocking)
	go s.adminRepository.CreateSystemEvent(event)

	return s.adminRepository.DeleteUser(userID)
}

// LogUserActivity logs user activity
func (s *AdminService) LogUserActivity(userID uint, action, ipAddress, userAgent string, success bool, details map[string]interface{}, errorMessage string) error {
	detailsJSON := ""
	if details != nil {
		detailsBytes, _ := json.Marshal(details)
		detailsJSON = string(detailsBytes)
	}

	activity := &models.UserActivity{
		UserID:       userID,
		Action:       action,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      detailsJSON,
		Success:      success,
		ErrorMessage: errorMessage,
	}

	return s.adminRepository.CreateUserActivity(activity)
}

// LogSystemEvent logs system events
func (s *AdminService) LogSystemEvent(eventType, severity, message string, details map[string]interface{}, ipAddress, userAgent string, userID *uint) error {
	detailsJSON := ""
	if details != nil {
		detailsBytes, _ := json.Marshal(details)
		detailsJSON = string(detailsBytes)
	}

	event := &models.SystemEvent{
		EventType: eventType,
		Severity:  severity,
		Message:   message,
		Details:   detailsJSON,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		UserID:    userID,
	}

	return s.adminRepository.CreateSystemEvent(event)
}

// CleanupOldLogs removes old activity logs and events
func (s *AdminService) CleanupOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Cleanup activities
	if err := s.adminRepository.CleanupOldActivities(cutoffDate); err != nil {
		return fmt.Errorf("failed to cleanup old activities: %w", err)
	}

	// Cleanup events
	if err := s.adminRepository.CleanupOldEvents(cutoffDate); err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	// Log cleanup action
	details := map[string]interface{}{
		"retention_days": retentionDays,
		"cutoff_date":    cutoffDate.Format(time.RFC3339),
	}

	s.LogSystemEvent("system", "info", fmt.Sprintf("Cleaned up logs older than %d days", retentionDays), details, "", "", nil)

	return nil
}

// IsUserAdmin checks if user has admin role
func (s *AdminService) IsUserAdmin(userID uint) (bool, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return false, err
	}

	return user.Role == "admin", nil
}

// IsUserModerator checks if user has moderator or admin role
func (s *AdminService) IsUserModerator(userID uint) (bool, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return false, err
	}

	return user.Role == "admin" || user.Role == "moderator", nil
}

// GetValidRoles returns list of valid user roles
func (s *AdminService) GetValidRoles() []string {
	return []string{"user", "moderator", "admin"}
}

// GetValidEventTypes returns list of valid event types
func (s *AdminService) GetValidEventTypes() []string {
	return []string{"security", "system", "admin", "oauth", "auth"}
}

// GetValidSeverityLevels returns list of valid severity levels
func (s *AdminService) GetValidSeverityLevels() []string {
	return []string{"info", "warning", "error", "critical"}
}
