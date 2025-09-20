package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"
	"time"

	"ex4-oauth2/internal/models"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// EmailService handles email operations
type EmailService struct {
	config                      *EmailConfig
	emailVerificationRepository models.EmailVerificationRepository
	emailTemplateRepository     models.EmailTemplateRepository
	userRepository              models.UserRepository
}

// EmailData represents data for email templates
type EmailData struct {
	UserName        string
	Email           string
	VerificationURL string
	Token           string
	ExpirationTime  string
	SiteName        string
	SiteURL         string
	SupportEmail    string
	CompanyName     string
	Year            int
}

// NewEmailService creates a new email service
func NewEmailService(
	emailVerificationRepo models.EmailVerificationRepository,
	emailTemplateRepo models.EmailTemplateRepository,
	userRepo models.UserRepository,
) *EmailService {
	config := &EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@yourapp.com"),
		FromName:     getEnv("FROM_NAME", "Your App"),
	}

	return &EmailService{
		config:                      config,
		emailVerificationRepository: emailVerificationRepo,
		emailTemplateRepository:     emailTemplateRepo,
		userRepository:              userRepo,
	}
}

// SendVerificationEmail sends email verification
func (s *EmailService) SendVerificationEmail(user *models.User, baseURL string) error {
	// Generate verification token
	token, err := s.generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Create verification record
	verification := &models.EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		Type:      "registration",
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiration
	}

	if err := s.emailVerificationRepository.Create(verification); err != nil {
		return fmt.Errorf("failed to create verification record: %w", err)
	}

	// Prepare email data
	emailData := &EmailData{
		UserName:        fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Email:           user.Email,
		VerificationURL: fmt.Sprintf("%s/api/auth/verify-email?token=%s", baseURL, token),
		Token:           token,
		ExpirationTime:  verification.ExpiresAt.Format("January 2, 2006 at 3:04 PM"),
		SiteName:        getEnv("SITE_NAME", "Your App"),
		SiteURL:         baseURL,
		SupportEmail:    getEnv("SUPPORT_EMAIL", "support@yourapp.com"),
		CompanyName:     getEnv("COMPANY_NAME", "Your Company"),
		Year:            time.Now().Year(),
	}

	// Get email template
	template, err := s.emailTemplateRepository.GetByName("verification")
	if err != nil {
		// Use default template if not found
		return s.sendDefaultVerificationEmail(user.Email, emailData)
	}

	// Send email using template
	return s.sendTemplatedEmail(user.Email, template, emailData)
}

// SendPasswordResetEmail sends password reset email
func (s *EmailService) SendPasswordResetEmail(user *models.User, baseURL string) error {
	// Generate reset token
	token, err := s.generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Delete existing password reset tokens for this user
	s.emailVerificationRepository.DeleteByUserID(user.ID, "password_reset")

	// Create reset record
	verification := &models.EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiration
	}

	if err := s.emailVerificationRepository.Create(verification); err != nil {
		return fmt.Errorf("failed to create reset record: %w", err)
	}

	// Prepare email data
	emailData := &EmailData{
		UserName:        fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Email:           user.Email,
		VerificationURL: fmt.Sprintf("%s/reset-password?token=%s", baseURL, token),
		Token:           token,
		ExpirationTime:  verification.ExpiresAt.Format("January 2, 2006 at 3:04 PM"),
		SiteName:        getEnv("SITE_NAME", "Your App"),
		SiteURL:         baseURL,
		SupportEmail:    getEnv("SUPPORT_EMAIL", "support@yourapp.com"),
		CompanyName:     getEnv("COMPANY_NAME", "Your Company"),
		Year:            time.Now().Year(),
	}

	// Get email template
	template, err := s.emailTemplateRepository.GetByName("password_reset")
	if err != nil {
		// Use default template if not found
		return s.sendDefaultPasswordResetEmail(user.Email, emailData)
	}

	// Send email using template
	return s.sendTemplatedEmail(user.Email, template, emailData)
}

// VerifyEmail verifies an email verification token
func (s *EmailService) VerifyEmail(token string) (*models.User, error) {
	// Get verification record
	verification, err := s.emailVerificationRepository.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid verification token")
	}

	// Check if token is expired
	if verification.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("verification token has expired")
	}

	// Check if already verified
	if verification.VerifiedAt != nil {
		return nil, fmt.Errorf("email already verified")
	}

	// Get user
	user, err := s.userRepository.GetByID(verification.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Mark as verified
	if err := s.emailVerificationRepository.MarkAsVerified(token); err != nil {
		return nil, fmt.Errorf("failed to mark as verified: %w", err)
	}

	// Update user email verified status
	user.EmailVerified = true
	if err := s.userRepository.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// generateVerificationToken generates a secure random token
func (s *EmailService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// sendTemplatedEmail sends email using template
func (s *EmailService) sendTemplatedEmail(to string, emailTemplate *models.EmailTemplate, data *EmailData) error {
	// Parse and execute subject template
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return fmt.Errorf("failed to parse subject template: %w", err)
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return fmt.Errorf("failed to execute subject template: %w", err)
	}

	// Parse and execute HTML body template
	htmlTmpl, err := template.New("html").Parse(emailTemplate.HTMLBody)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	// Parse and execute text body template
	var textBuf bytes.Buffer
	if emailTemplate.TextBody != "" {
		textTmpl, err := template.New("text").Parse(emailTemplate.TextBody)
		if err != nil {
			return fmt.Errorf("failed to parse text template: %w", err)
		}

		if err := textTmpl.Execute(&textBuf, data); err != nil {
			return fmt.Errorf("failed to execute text template: %w", err)
		}
	}

	// Send email
	return s.sendEmail(to, subjectBuf.String(), htmlBuf.String(), textBuf.String())
}

// sendDefaultVerificationEmail sends default verification email
func (s *EmailService) sendDefaultVerificationEmail(to string, data *EmailData) error {
	subject := fmt.Sprintf("Verify your email for %s", data.SiteName)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c5aa0;">Email Verification</h2>
        <p>Hello %s,</p>
        <p>Thank you for registering with %s. To complete your registration, please verify your email address by clicking the button below:</p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Verify Email</a>
        </div>
        
        <p>Or copy and paste this link in your browser:</p>
        <p style="word-break: break-all; color: #666;">%s</p>
        
        <p><strong>This verification link will expire on %s.</strong></p>
        
        <p>If you didn't create an account with us, please ignore this email.</p>
        
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="font-size: 12px; color: #666;">
            Best regards,<br>
            The %s Team<br>
            <a href="mailto:%s">%s</a>
        </p>
    </div>
</body>
</html>
`, data.UserName, data.SiteName, data.VerificationURL, data.VerificationURL, data.ExpirationTime, data.CompanyName, data.SupportEmail, data.SupportEmail)

	textBody := fmt.Sprintf(`
Email Verification

Hello %s,

Thank you for registering with %s. To complete your registration, please verify your email address by visiting this link:

%s

This verification link will expire on %s.

If you didn't create an account with us, please ignore this email.

Best regards,
The %s Team
%s
`, data.UserName, data.SiteName, data.VerificationURL, data.ExpirationTime, data.CompanyName, data.SupportEmail)

	return s.sendEmail(to, subject, htmlBody, textBody)
}

// sendDefaultPasswordResetEmail sends default password reset email
func (s *EmailService) sendDefaultPasswordResetEmail(to string, data *EmailData) error {
	subject := fmt.Sprintf("Reset your password for %s", data.SiteName)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #dc3545;">Password Reset Request</h2>
        <p>Hello %s,</p>
        <p>We received a request to reset your password for your %s account. Click the button below to reset your password:</p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #dc3545; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        
        <p>Or copy and paste this link in your browser:</p>
        <p style="word-break: break-all; color: #666;">%s</p>
        
        <p><strong>This reset link will expire on %s.</strong></p>
        
        <p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="font-size: 12px; color: #666;">
            Best regards,<br>
            The %s Team<br>
            <a href="mailto:%s">%s</a>
        </p>
    </div>
</body>
</html>
`, data.UserName, data.SiteName, data.VerificationURL, data.VerificationURL, data.ExpirationTime, data.CompanyName, data.SupportEmail, data.SupportEmail)

	textBody := fmt.Sprintf(`
Password Reset Request

Hello %s,

We received a request to reset your password for your %s account. To reset your password, please visit this link:

%s

This reset link will expire on %s.

If you didn't request a password reset, please ignore this email or contact support if you have concerns.

Best regards,
The %s Team
%s
`, data.UserName, data.SiteName, data.VerificationURL, data.ExpirationTime, data.CompanyName, data.SupportEmail)

	return s.sendEmail(to, subject, htmlBody, textBody)
}

// sendEmail sends email via SMTP
func (s *EmailService) sendEmail(to, subject, htmlBody, textBody string) error {
	// SMTP server configuration
	smtpAddr := s.config.SMTPHost + ":" + s.config.SMTPPort
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Create message
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	// Build email message
	msg := s.buildEmailMessage(from, to, subject, htmlBody, textBody)

	// Send email
	err := smtp.SendMail(smtpAddr, auth, s.config.FromEmail, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildEmailMessage builds the complete email message
func (s *EmailService) buildEmailMessage(from, to, subject, htmlBody, textBody string) string {
	boundary := "boundary-" + generateBoundary()

	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"", boundary),
		"",
	}

	body := []string{
		fmt.Sprintf("--%s", boundary),
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		textBody,
		"",
		fmt.Sprintf("--%s", boundary),
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		htmlBody,
		"",
		fmt.Sprintf("--%s--", boundary),
	}

	return strings.Join(headers, "\r\n") + strings.Join(body, "\r\n")
}

// generateBoundary generates a unique boundary for multipart email
func generateBoundary() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CleanupExpiredTokens removes expired verification tokens
func (s *EmailService) CleanupExpiredTokens() error {
	return s.emailVerificationRepository.CleanupExpiredTokens()
}

// GetVerificationByToken gets verification record by token
func (s *EmailService) GetVerificationByToken(token string) (*models.EmailVerification, error) {
	return s.emailVerificationRepository.GetByToken(token)
}

// MarkTokenAsVerified marks a token as verified
func (s *EmailService) MarkTokenAsVerified(token string) error {
	return s.emailVerificationRepository.MarkAsVerified(token)
}
