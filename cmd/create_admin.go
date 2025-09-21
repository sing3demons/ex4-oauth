package main

import (
	"log"

	"ex4-oauth2/internal/config"
	"ex4-oauth2/internal/database"
	"ex4-oauth2/internal/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Setup MongoDB
	mongodb, err := database.NewMongoDB(cfg.DatabaseURL, "oauth2_db")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	// Setup user repository
	userRepo := database.NewMongoUserRepository(mongodb.GetDatabase())

	// Create admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := &models.User{
		ID:            "admin-user-id",
		Email:         "admin@example.com",
		Username:      "admin",
		FirstName:     "Admin",
		LastName:      "User",
		Password:      string(hashedPassword),
		Role:          "admin",
		IsActive:      true,
		EmailVerified: true,
		Provider:      "local",
	}

	// Check if admin exists
	existingAdmin, err := userRepo.GetByEmail("admin@example.com")
	if err != nil || existingAdmin == nil {
		// Create admin user
		if err := userRepo.Create(admin); err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Println("✅ Admin user created successfully!")
			log.Println("📧 Email: admin@example.com")
			log.Println("🔑 Password: admin123")
			log.Println("👑 Role: admin")
		}
	} else {
		// Update existing admin to ensure admin role
		existingAdmin.Role = "admin"
		existingAdmin.IsActive = true
		existingAdmin.EmailVerified = true

		if err := userRepo.Update(existingAdmin); err != nil {
			log.Printf("Failed to update admin user: %v", err)
		} else {
			log.Println("✅ Admin user updated successfully!")
			log.Println("📧 Email: admin@example.com")
			log.Println("🔑 Password: admin123")
			log.Println("👑 Role: admin")
		}
	}

	log.Println("🚀 Admin setup completed!")
}
