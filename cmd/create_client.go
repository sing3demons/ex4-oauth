package main

import (
	"log"

	"ex4-oauth2/internal/config"
	"ex4-oauth2/internal/database"
	"ex4-oauth2/internal/models"

	"github.com/joho/godotenv"
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

	// Setup OAuth2 client repository
	clientRepo := database.NewMongoOAuth2ClientRepository(mongodb.GetDatabase())

	// Create OAuth2 client for Node.js application
	client := &models.OAuth2Client{
		Name:        "NodeJS OAuth Client",
		Description: "OAuth2 client for Node.js application",
		RedirectURIs: []string{
			"http://localhost:3001/auth/callback",
			"http://localhost:3000/auth/callback", // Alternative port
		},
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"read",
		},
		GrantTypes: []string{
			"authorization_code",
			"refresh_token",
		},
		IsPublic: false, // Confidential client (has client secret)
		IsActive: true,
	}

	// Set defaults (generates client ID and secret)
	client.SetDefaults()

	// Check if client with this name already exists
	existingClients, err := clientRepo.List(100, 0) // Get up to 100 clients
	if err != nil {
		log.Fatalf("Failed to check existing clients: %v", err)
	}

	var existingClient *models.OAuth2Client
	for _, c := range existingClients {
		if c.Name == client.Name {
			existingClient = c
			break
		}
	}

	if existingClient != nil {
		log.Println("✅ OAuth2 client already exists!")
		log.Printf("🆔 Client ID: %s", existingClient.ClientID)
		log.Printf("🔐 Client Secret: %s", existingClient.ClientSecret)
		log.Printf("📝 Name: %s", existingClient.Name)
		log.Printf("📖 Description: %s", existingClient.Description)
		log.Printf("🔗 Redirect URIs: %v", existingClient.RedirectURIs)
		log.Printf("🔧 Scopes: %v", existingClient.Scopes)
		log.Printf("🎯 Grant Types: %v", existingClient.GrantTypes)
		log.Printf("🔒 Is Public: %v", existingClient.IsPublic)
		log.Printf("✅ Is Active: %v", existingClient.IsActive)
	} else {
		// Create new client
		if err := clientRepo.Create(client); err != nil {
			log.Fatalf("Failed to create OAuth2 client: %v", err)
		}

		log.Println("✅ OAuth2 client created successfully!")
		log.Printf("🆔 Client ID: %s", client.ClientID)
		log.Printf("🔐 Client Secret: %s", client.ClientSecret)
		log.Printf("📝 Name: %s", client.Name)
		log.Printf("📖 Description: %s", client.Description)
		log.Printf("🔗 Redirect URIs: %v", client.RedirectURIs)
		log.Printf("🔧 Scopes: %v", client.Scopes)
		log.Printf("🎯 Grant Types: %v", client.GrantTypes)
		log.Printf("🔒 Is Public: %v", client.IsPublic)
		log.Printf("✅ Is Active: %v", client.IsActive)

		log.Println("")
		log.Println("📋 Configuration for your .env file:")
		log.Printf("OAUTH_CLIENT_ID=%s", client.ClientID)
		log.Printf("OAUTH_CLIENT_SECRET=%s", client.ClientSecret)
	}

	log.Println("")
	log.Println("🚀 OAuth2 Client setup completed!")
	log.Println("💡 You can now use these credentials in your Node.js application")
}
