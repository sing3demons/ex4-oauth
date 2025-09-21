package main

import (
	"log"

	"ex4-oauth2/internal/config"
	"ex4-oauth2/internal/database"
	"ex4-oauth2/internal/models"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup MongoDB
	mongodb, err := database.NewMongoDB(cfg.DatabaseURL, "oauth2_db")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	// Setup email template repository
	emailTemplateRepo := database.NewMongoEmailTemplateRepository(mongodb.GetDatabase())

	// Create default templates
	templates := []models.EmailTemplate{
		{
			ID:       "verification-template",
			Name:     "verification",
			Subject:  "Verify Your Email Address",
			TextBody: "Please verify your email by clicking this link: {{.VerificationURL}}",
			HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <title>Email Verification</title>
</head>
<body>
    <h2>Email Verification</h2>
    <p>Hello {{.UserName}},</p>
    <p>Please verify your email address by clicking the button below:</p>
    <a href="{{.VerificationURL}}" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; border-radius: 4px;">Verify Email</a>
    <p>If the button doesn't work, you can copy and paste this link: {{.VerificationURL}}</p>
    <p>This verification link will expire in 24 hours.</p>
    <p>Best regards,<br>OAuth2 Server Team</p>
</body>
</html>`,
			Variables: `["Username", "VerificationURL"]`,
			IsActive:  true,
		},
		{
			ID:       "password-reset-template",
			Name:     "password_reset",
			Subject:  "Reset Your Password",
			TextBody: "Reset your password by clicking this link: {{.ResetURL}}",
			HTMLBody: `
<!DOCTYPE html>
<html>
<head>
    <title>Password Reset</title>
</head>
<body>
    <h2>Password Reset</h2>
    <p>Hello {{.UserName}},</p>
    <p>You requested to reset your password. Click the button below to reset it:</p>
    <a href="{{.VerificationURL}}" style="background-color: #f44336; color: white; padding: 14px 20px; text-decoration: none; border-radius: 4px;">Reset Password</a>
    <p>If the button doesn't work, you can copy and paste this link: {{.VerificationURL}}</p>
    <p>This reset link will expire in 1 hour.</p>
    <p>If you didn't request this password reset, please ignore this email.</p>
    <p>Best regards,<br>OAuth2 Server Team</p>
</body>
</html>`,
			Variables: `["Username", "ResetURL"]`,
			IsActive:  true,
		},
	}

	// Insert templates
	for _, template := range templates {
		// Check if template already exists
		existing, err := emailTemplateRepo.GetByName(template.Name)
		if err != nil || existing == nil {
			// Template doesn't exist, create it
			if err := emailTemplateRepo.Create(&template); err != nil {
				log.Printf("Failed to create template %s: %v", template.Name, err)
			} else {
				log.Printf("Created email template: %s", template.Name)
			}
		} else {
			// Template exists, update it
			template.ID = existing.ID // Keep the same ID
			if err := emailTemplateRepo.Update(&template); err != nil {
				log.Printf("Failed to update template %s: %v", template.Name, err)
			} else {
				log.Printf("Updated email template: %s", template.Name)
			}
		}
	}

	log.Println("Email template setup completed!")
}
