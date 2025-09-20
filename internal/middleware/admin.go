package middleware

import (
	"net/http"

	"ex4-oauth2/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware checks if user has admin role
func AdminMiddleware(userRepo models.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from auth middleware
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Get user details
		user, err := userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user is admin
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		// Set user role in context
		c.Set("user_role", user.Role)
		c.Next()
	}
}

// ModeratorMiddleware checks if user has moderator or admin role
func ModeratorMiddleware(userRepo models.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from auth middleware
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Get user details
		user, err := userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user is moderator or admin
		if user.Role != "admin" && user.Role != "moderator" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Moderator or admin access required",
			})
			c.Abort()
			return
		}

		// Set user role in context
		c.Set("user_role", user.Role)
		c.Next()
	}
}

// RoleRequiredMiddleware checks if user has specific role
func RoleRequiredMiddleware(userRepo models.UserRepository, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from auth middleware
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			c.Abort()
			return
		}

		// Get user details
		user, err := userRepo.GetByID(userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range requiredRoles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient permissions",
				"required_roles": requiredRoles,
				"user_role":      user.Role,
			})
			c.Abort()
			return
		}

		// Set user role in context
		c.Set("user_role", user.Role)
		c.Next()
	}
}
