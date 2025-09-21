package services

import (
	"fmt"

	"ex4-oauth2/internal/models"
)

// AdminService handles admin operations
type AdminService struct {
	userRepository models.UserRepository
}

// NewAdminService creates a new admin service
func NewAdminService(userRepo models.UserRepository) *AdminService {
	return &AdminService{
		userRepository: userRepo,
	}
}

// GetDashboardStats returns dashboard statistics
func (s *AdminService) GetDashboardStats() (*models.AdminStats, error) {
	// Simplified implementation - return mock stats for now
	return &models.AdminStats{
		TotalUsers:   100,
		ActiveUsers:  80,
		TotalClients: 10,
		TotalTokens:  50,
	}, nil
}

// GetUsers returns paginated user list with filters
func (s *AdminService) GetUsers(limit, offset int, filters map[string]interface{}) ([]*models.User, int64, error) {
	// Simplified implementation - return empty list for now
	return []*models.User{}, 0, nil
}

// GetUserActivity returns user activity logs
func (s *AdminService) GetUserActivity(userID string, limit, offset int) ([]*models.UserActivity, error) {
	// Simplified implementation - return empty list
	return []*models.UserActivity{}, nil
}

// GetSystemEvents returns system event logs
func (s *AdminService) GetSystemEvents(limit, offset int, filters map[string]interface{}) ([]*models.SystemEvent, error) {
	// Simplified implementation - return empty list
	return []*models.SystemEvent{}, nil
}

// UpdateUserStatus updates user active status
func (s *AdminService) UpdateUserStatus(userID string, isActive bool) error {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsActive = isActive
	return s.userRepository.Update(user)
}

// UpdateUserRole updates user role
func (s *AdminService) UpdateUserRole(userID string, role string) error {
	// Validate role
	validRoles := map[string]bool{
		"user":      true,
		"moderator": true,
		"admin":     true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}

	user.Role = role
	return s.userRepository.Update(user)
}

// DeleteUser soft deletes a user
func (s *AdminService) DeleteUser(userID string) error {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Soft delete by setting IsActive to false
	user.IsActive = false
	return s.userRepository.Update(user)
}

// LogUserActivity logs user activity (simplified implementation)
func (s *AdminService) LogUserActivity(userID string, action, ipAddress, userAgent string, success bool, details map[string]interface{}, errorMessage string) error {
	// Simplified implementation - just log to console for now
	fmt.Printf("User Activity: UserID=%s, Action=%s, Success=%t\n", userID, action, success)
	return nil
}

// LogSystemEvent logs system events (simplified implementation)
func (s *AdminService) LogSystemEvent(eventType, severity, message string, details map[string]interface{}, ipAddress, userAgent string, userID *string) error {
	// Simplified implementation - just log to console for now
	fmt.Printf("System Event: Type=%s, Severity=%s, Message=%s\n", eventType, severity, message)
	return nil
}

// CleanupOldLogs removes old activity logs and events (simplified implementation)
func (s *AdminService) CleanupOldLogs(retentionDays int) error {
	// Simplified implementation - no-op for now
	fmt.Printf("Cleanup: Would remove logs older than %d days\n", retentionDays)
	return nil
}

// IsUserAdmin checks if user has admin role
func (s *AdminService) IsUserAdmin(userID string) (bool, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return false, err
	}

	return user.Role == "admin", nil
}

// IsUserModerator checks if user has moderator or admin role
func (s *AdminService) IsUserModerator(userID string) (bool, error) {
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
