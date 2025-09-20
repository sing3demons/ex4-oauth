package handlers

import (
	"fmt"
	"net/http"

	"ex4-oauth2/internal/models"
	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
)

// EmailHandler handles email-related operations
type EmailHandler struct {
	emailService   *services.EmailService
	userRepository models.UserRepository
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(
	emailService *services.EmailService,
	userRepository models.UserRepository,
) *EmailHandler {
	return &EmailHandler{
		emailService:   emailService,
		userRepository: userRepository,
	}
}

// SendVerificationEmailRequest represents the request for sending verification email
type SendVerificationEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmailRequest represents the request for email verification
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// PasswordResetRequest represents the request for password reset
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the request for resetting password
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// SendVerificationEmail sends verification email to user
func (h *EmailHandler) SendVerificationEmail(c *gin.Context) {
	var req SendVerificationEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get user by email
	user, err := h.userRepository.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Check if email is already verified
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email is already verified",
		})
		return
	}

	// Get base URL from request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	if ssl := c.GetHeader("X-Forwarded-Ssl"); ssl == "on" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	// Send verification email
	if err := h.emailService.SendVerificationEmail(user, baseURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send verification email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully",
	})
}

// VerifyEmail verifies email with token
func (h *EmailHandler) VerifyEmail(c *gin.Context) {
	// Get token from query parameter or body
	token := c.Query("token")
	if token == "" {
		var req VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request",
				"details": err.Error(),
			})
			return
		}
		token = req.Token
	}

	// Verify email
	user, err := h.emailService.VerifyEmail(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
		"user": gin.H{
			"id":             user.ID,
			"email":          user.Email,
			"username":       user.Username,
			"email_verified": user.EmailVerified,
		},
	})
}

// SendPasswordResetEmail sends password reset email
func (h *EmailHandler) SendPasswordResetEmail(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get user by email
	user, err := h.userRepository.GetByEmail(req.Email)
	if err != nil {
		// Don't reveal if user exists or not for security
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account with that email exists, we've sent a password reset link",
		})
		return
	}

	// Only send reset email for local accounts
	if user.Provider != "local" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password reset is not available for social media accounts",
		})
		return
	}

	// Get base URL from request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	if ssl := c.GetHeader("X-Forwarded-Ssl"); ssl == "on" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	// Send password reset email
	if err := h.emailService.SendPasswordResetEmail(user, baseURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send password reset email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with that email exists, we've sent a password reset link",
	})
}

// ResetPassword resets password with token
func (h *EmailHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get verification record
	verification, err := h.emailService.GetVerificationByToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired reset token",
		})
		return
	}

	// Check if it's a password reset token
	if verification.Type != "password_reset" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token type",
		})
		return
	}

	// Get user
	user, err := h.userRepository.GetByID(verification.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Update password (will be hashed by BeforeCreate hook)
	user.Password = req.NewPassword
	if err := h.userRepository.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update password",
			"details": err.Error(),
		})
		return
	}

	// Mark token as verified (to prevent reuse)
	if err := h.emailService.MarkTokenAsVerified(req.Token); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to mark token as verified: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetEmailStats returns email verification statistics
func (h *EmailHandler) GetEmailStats(c *gin.Context) {
	// This would require additional repository methods to get stats
	// For now, return basic info
	c.JSON(http.StatusOK, gin.H{
		"message": "Email verification system is active",
		"features": []string{
			"Email verification for new registrations",
			"Password reset via email",
			"Resend verification emails",
			"Token expiration handling",
		},
	})
}

// CleanupExpiredTokens manually triggers cleanup of expired tokens
func (h *EmailHandler) CleanupExpiredTokens(c *gin.Context) {
	if err := h.emailService.CleanupExpiredTokens(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup expired tokens",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expired tokens cleaned up successfully",
	})
}
