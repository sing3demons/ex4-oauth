package main

import (
	"log"
	"time"

	"ex4-oauth2/internal/auth"
	"ex4-oauth2/internal/config"
	"ex4-oauth2/internal/database"
	"ex4-oauth2/internal/handlers"
	"ex4-oauth2/internal/middleware"
	"ex4-oauth2/internal/models"
	"ex4-oauth2/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Setup MongoDB
	mongodb, err := database.NewMongoDB(cfg.DatabaseURL, "oauth2_db")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	// Create indexes
	if err := mongodb.CreateIndexes(); err != nil {
		log.Fatalf("Failed to create MongoDB indexes: %v", err)
	}

	// Setup repositories
	userRepo := database.NewMongoUserRepository(mongodb.GetDatabase())
	refreshTokenRepo := database.NewMongoRefreshTokenRepository(mongodb.GetDatabase())
	emailVerificationRepo := database.NewMongoEmailVerificationRepository(mongodb.GetDatabase())
	emailTemplateRepo := database.NewMongoEmailTemplateRepository(mongodb.GetDatabase())

	// OAuth2 repositories
	clientRepo := database.NewMongoOAuth2ClientRepository(mongodb.GetDatabase())
	authCodeRepo := database.NewMongoOAuth2AuthorizationCodeRepository(mongodb.GetDatabase())
	accessTokenRepo := database.NewMongoOAuth2AccessTokenRepository(mongodb.GetDatabase())

	// Setup JWT service
	jwtService := auth.NewJWTService(
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		refreshTokenRepo,
	)

	// Setup email service
	emailService := services.NewEmailService(
		emailVerificationRepo,
		emailTemplateRepo, // email template repo
		userRepo,
	)

	// Setup services (simplified - using userRepo as admin repo for now)
	adminService := services.NewAdminService(userRepo)

	introspectionService := services.NewIntrospectionService(
		accessTokenRepo,
		refreshTokenRepo,
		clientRepo,
		userRepo,
		jwtService,
	)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService, emailService)
	adminHandler := handlers.NewAdminHandler(adminService, userRepo)
	introspectionHandler := handlers.NewIntrospectionHandler(introspectionService)
	oauthHandler := handlers.NewOAuthHandler(
		userRepo,
		clientRepo,
		authCodeRepo,
		accessTokenRepo,
		jwtService,
		"http://localhost:"+cfg.Port,
	)
	oauthClientHandler := handlers.NewOAuth2ClientHandler(clientRepo)

	// Setup router
	router := setupFullRouter(cfg, authHandler, oauthHandler, oauthClientHandler, adminHandler, introspectionHandler, jwtService, userRepo)

	// Start server
	log.Printf("🚀 OAuth2 Server starting on %s", cfg.GetServerAddress())
	log.Printf("📋 Available endpoints:")
	log.Printf("  • Health: GET /health")
	log.Printf("  • OAuth2: GET|POST /api/auth/oauth/authorize")
	log.Printf("  • Token: POST /api/auth/oauth/token")
	log.Printf("  • UserInfo: GET /api/auth/oauth/userinfo")
	log.Printf("  • Discovery: GET /api/auth/oauth/.well-known/openid-configuration")
	log.Printf("  • Register: POST /api/auth/register")
	log.Printf("  • Login: POST /api/auth/login")
	log.Printf("  • Verify Email: GET /api/auth/verify-email")
	log.Printf("  • Profile: GET /api/auth/profile")
	log.Printf("  • Refresh: POST /api/auth/refresh")
	log.Printf("  • Logout: POST /api/auth/logout")
	log.Printf("  • Clients: GET|POST|PUT|DELETE /api/auth/clients")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupFullRouter(cfg *config.Config, authHandler *handlers.AuthHandler, oauthHandler *handlers.OAuthHandler, oauthClientHandler *handlers.OAuth2ClientHandler, adminHandler *handlers.AdminHandler, introspectionHandler *handlers.IntrospectionHandler, jwtService *auth.JWTService, userRepo models.UserRepository) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
			"features": []string{
				"oauth2",
				"user_auth",
				"jwt_tokens",
				"mongodb",
				"email_verification",
				"client_management",
			},
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			// Local authentication endpoints
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/verify-email", authHandler.VerifyEmail)

			// Protected authentication endpoints
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(jwtService))
			{
				authProtected.GET("/profile", authHandler.GetProfile)
				authProtected.PUT("/profile", authHandler.UpdateProfile)
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.POST("/logout-all", authHandler.LogoutAll)
				authProtected.POST("/change-password", authHandler.ChangePassword)
			}

			// OAuth2 routes
			oauth := auth.Group("/oauth")
			{
				oauth.GET("/authorize", oauthHandler.Authorize)
				oauth.POST("/authorize", oauthHandler.Authorize)
				oauth.POST("/token", oauthHandler.Token)
				oauth.GET("/userinfo", oauthHandler.UserInfo)
				oauth.GET("/.well-known/openid-configuration", oauthHandler.WellKnown)
				oauth.POST("/introspect", introspectionHandler.IntrospectToken)
			}

			// OAuth2 client management
			clients := auth.Group("/clients")
			clients.Use(middleware.AuthMiddleware(jwtService))
			{
				clients.POST("", oauthClientHandler.CreateClient)
				clients.GET("/:id", oauthClientHandler.GetClient)
				clients.PUT("/:id", oauthClientHandler.UpdateClient)
				clients.DELETE("/:id", oauthClientHandler.DeleteClient)
				clients.GET("", oauthClientHandler.GetClients)
			}

			// Admin routes (require admin role)
			admin := auth.Group("/admin")
			admin.Use(middleware.AuthMiddleware(jwtService))
			admin.Use(middleware.AdminMiddleware(userRepo))
			{
				admin.GET("/dashboard", adminHandler.GetDashboardStats)
				admin.GET("/users", adminHandler.GetUsers)
				admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
				admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
				admin.DELETE("/users/:id", adminHandler.DeleteUser)
				admin.GET("/users/:id/activity", adminHandler.GetUserActivity)
				admin.GET("/events", adminHandler.GetSystemEvents)
			}
		}

		// API documentation endpoint
		api.GET("/endpoints", func(c *gin.Context) {
			endpoints := gin.H{
				"health": "GET /health",
				"auth": gin.H{
					"register":        "POST /api/auth/register",
					"login":           "POST /api/auth/login",
					"refresh":         "POST /api/auth/refresh",
					"profile":         "GET /api/auth/profile (protected)",
					"update_profile":  "PUT /api/auth/profile (protected)",
					"logout":          "POST /api/auth/logout (protected)",
					"logout_all":      "POST /api/auth/logout-all (protected)",
					"change_password": "POST /api/auth/change-password (protected)",
				},
				"oauth": gin.H{
					"authorize": "GET|POST /api/auth/oauth/authorize",
					"token":     "POST /api/auth/oauth/token",
					"userinfo":  "GET /api/auth/oauth/userinfo",
					"discovery": "GET /api/auth/oauth/.well-known/openid-configuration",
				},
				"clients": gin.H{
					"create": "POST /api/auth/clients",
					"get":    "GET /api/auth/clients/:id",
					"update": "PUT /api/auth/clients/:id",
					"delete": "DELETE /api/auth/clients/:id",
					"list":   "GET /api/auth/clients",
				},
			}
			c.JSON(200, endpoints)
		})
	}

	return router
}
