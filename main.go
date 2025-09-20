package main

import (
	"log"
	"time"

	"ex4-oauth2/internal/auth"
	"ex4-oauth2/internal/config"
	"ex4-oauth2/internal/database"
	"ex4-oauth2/internal/handlers"
	"ex4-oauth2/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Setup database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup repositories
	userRepo := database.NewUserRepository(db.GetDB())
	refreshTokenRepo := database.NewRefreshTokenRepository(db.GetDB())

	// OAuth2 repositories
	clientRepo := database.NewOAuth2ClientRepository(db.GetDB())
	authCodeRepo := database.NewOAuth2AuthorizationCodeRepository(db.GetDB())
	accessTokenRepo := database.NewOAuth2AccessTokenRepository(db.GetDB())

	// Setup JWT service
	jwtService := auth.NewJWTService(
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		refreshTokenRepo,
	)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)
	oauthHandler := handlers.NewOAuthHandler(
		userRepo,
		clientRepo,
		authCodeRepo,
		accessTokenRepo,
		jwtService,
		"http://localhost:"+cfg.Port, // baseURL
	)
	oauthClientHandler := handlers.NewOAuth2ClientHandler(clientRepo)

	// Setup router
	router := setupRouter(cfg, authHandler, oauthHandler, oauthClientHandler, jwtService)

	// Start server
	log.Printf("Server starting on %s", cfg.GetServerAddress())
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, oauthHandler *handlers.OAuthHandler, oauthClientHandler *handlers.OAuth2ClientHandler, jwtService *auth.JWTService) *gin.Engine {
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
	router.Use(middleware.EnhancedAPIRateLimitMiddleware()) // Add rate limiting

	// Start cleanup for rate limiters
	middleware.CleanupExpiredLimiters()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			// Local authentication
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", middleware.StrictLoginRateLimitMiddleware(), authHandler.Login) // Add login rate limiting
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)

			// OAuth routes
			oauth := auth.Group("/oauth")
			oauth.Use(middleware.OAuthRateLimitMiddleware()) // Add OAuth rate limiting
			{
				// OAuth2 Server endpoints
				oauth.GET("/authorize", oauthHandler.Authorize)
				oauth.POST("/token", oauthHandler.Token)
				oauth.POST("/consent", oauthHandler.Consent)
				oauth.GET("/userinfo", oauthHandler.UserInfo)
				oauth.GET("/.well-known/openid_configuration", oauthHandler.WellKnown)
				oauth.GET("/.well-known/jwks.json", oauthHandler.JWKs)
			}

			// Protected routes
			protected := auth.Group("")
			protected.Use(middleware.AuthMiddleware(jwtService))
			{
				protected.GET("/profile", authHandler.GetProfile)
				protected.PUT("/profile", authHandler.UpdateProfile)
				protected.POST("/change-password", authHandler.ChangePassword)
				protected.POST("/logout-all", authHandler.LogoutAll)
			}
		}

		// User management routes (protected)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtService))
		{
			users.GET("", authHandler.GetUsers)
		}

		// OAuth2 Client management routes (protected)
		clients := api.Group("/oauth2/clients")
		clients.Use(middleware.AuthMiddleware(jwtService))
		{
			clients.POST("", oauthClientHandler.CreateClient)
			clients.GET("", oauthClientHandler.GetClients)
			clients.GET("/:id", oauthClientHandler.GetClient)
			clients.PUT("/:id", oauthClientHandler.UpdateClient)
			clients.DELETE("/:id", oauthClientHandler.DeleteClient)
		}

		// API info
		api.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"name":        "OAuth2 Authorization Server API",
				"version":     "1.0.0",
				"description": "Complete OAuth2 Authorization Server with OIDC and PKCE support",
				"features": []string{
					"User Registration & Login",
					"JWT Authentication",
					"OAuth2 Authorization Server",
					"OIDC Support",
					"PKCE Security",
					"Profile Management",
					"Token Refresh",
					"Client Management",
					"ID Token Generation",
					"Rate Limiting",
				},
				"endpoints": gin.H{
					"auth": gin.H{
						"register":        "POST /api/auth/register",
						"login":           "POST /api/auth/login",
						"refresh":         "POST /api/auth/refresh",
						"logout":          "POST /api/auth/logout",
						"profile":         "GET /api/auth/profile",
						"update_profile":  "PUT /api/auth/profile",
						"change_password": "POST /api/auth/change-password",
						"logout_all":      "POST /api/auth/logout-all",
					},
					"oauth2": gin.H{
						"authorize": "GET /api/auth/oauth/authorize",
						"token":     "POST /api/auth/oauth/token",
						"consent":   "POST /api/auth/oauth/consent",
						"userinfo":  "GET /api/auth/oauth/userinfo",
						"discovery": "GET /api/auth/oauth/.well-known/openid_configuration",
						"jwks":      "GET /api/auth/oauth/.well-known/jwks.json",
					},
					"clients": gin.H{
						"create": "POST /api/oauth2/clients",
						"list":   "GET /api/oauth2/clients",
						"get":    "GET /api/oauth2/clients/:id",
						"update": "PUT /api/oauth2/clients/:id",
						"delete": "DELETE /api/oauth2/clients/:id",
					},
					"users": gin.H{
						"list": "GET /api/users",
					},
					"monitoring": gin.H{
						"rate_limits": "GET /api/monitoring/rate-limits",
					},
				},
			})
		})

		// Monitoring endpoints
		monitoring := api.Group("/monitoring")
		{
			monitoring.GET("/rate-limits", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"status":      "ok",
					"rate_limits": middleware.GetRateLimitStatus(),
					"timestamp":   time.Now().Unix(),
				})
			})
		}
	}

	return router
}
