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

	// Email repositories
	emailVerificationRepo := database.NewEmailVerificationRepository(db.GetDB())
	emailTemplateRepo := database.NewEmailTemplateRepository(db.GetDB())

	// Admin repository
	adminRepo := database.NewAdminRepository(db.GetDB())

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

	// Setup email service
	emailService := services.NewEmailService(
		emailVerificationRepo,
		emailTemplateRepo,
		userRepo,
	)

	// Setup admin service
	adminService := services.NewAdminService(
		adminRepo,
		userRepo,
	)

	// Setup audit service
	auditService := services.NewAuditService(db)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService, emailService)
	emailHandler := handlers.NewEmailHandler(emailService, userRepo)
	adminHandler := handlers.NewAdminHandler(adminService, userRepo)
	auditHandler := handlers.NewAuditHandler(auditService)
	oauthHandler := handlers.NewOAuthHandler(
		userRepo,
		clientRepo,
		authCodeRepo,
		accessTokenRepo,
		jwtService,
		"http://localhost:"+cfg.Port, // baseURL
	)
	oauthClientHandler := handlers.NewOAuth2ClientHandler(clientRepo)

	// Setup middleware
	auditMiddleware := middleware.NewAuditMiddleware(auditService)

	// Setup router
	router := setupRouter(cfg, authHandler, emailHandler, adminHandler, auditHandler, oauthHandler, oauthClientHandler, jwtService, userRepo, auditMiddleware)

	// Start server
	log.Printf("Server starting on %s", cfg.GetServerAddress())
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, emailHandler *handlers.EmailHandler, adminHandler *handlers.AdminHandler, auditHandler *handlers.AuditHandler, oauthHandler *handlers.OAuthHandler, oauthClientHandler *handlers.OAuth2ClientHandler, jwtService *auth.JWTService, userRepo models.UserRepository, auditMiddleware *middleware.AuditMiddleware) *gin.Engine {
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
	router.Use(auditMiddleware.LogRequest())                // Add audit logging
	router.Use(auditMiddleware.LogSecurityEvents())         // Add security event logging

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

			// Email verification routes
			auth.POST("/send-verification", emailHandler.SendVerificationEmail)
			auth.GET("/verify-email", emailHandler.VerifyEmail)
			auth.POST("/verify-email", emailHandler.VerifyEmail)
			auth.POST("/password-reset", emailHandler.SendPasswordResetEmail)
			auth.POST("/reset-password", emailHandler.ResetPassword)

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

		// Email management routes (protected)
		email := api.Group("/email")
		email.Use(middleware.AuthMiddleware(jwtService))
		{
			email.GET("/stats", emailHandler.GetEmailStats)
			email.POST("/cleanup", emailHandler.CleanupExpiredTokens)
		}

		// Admin routes (admin only)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(jwtService))
		admin.Use(middleware.AdminMiddleware(userRepo))
		{
			admin.GET("/stats", adminHandler.GetDashboardStats)
			admin.GET("/users", adminHandler.GetUsers)
			admin.GET("/users/:id/activity", adminHandler.GetUserActivity)
			admin.GET("/events", adminHandler.GetSystemEvents)
			admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.POST("/cleanup-logs", adminHandler.CleanupLogs)
			admin.GET("/system-info", adminHandler.GetSystemInfo)

			// Audit & Compliance routes
			audit := admin.Group("/audit")
			{
				audit.GET("/logs", auditHandler.GetAuditLogs)
				audit.GET("/search", auditHandler.SearchAuditLogs)
				audit.GET("/users/:id/trail", auditHandler.GetUserAuditTrail)
				audit.GET("/resources/:resource_type/:resource_id/trail", auditHandler.GetResourceAuditTrail)
				audit.GET("/high-risk", auditHandler.GetHighRiskActivity)
				audit.GET("/security-alerts", auditHandler.GetSecurityAlerts)
				audit.POST("/security-alerts/:id/resolve", auditHandler.ResolveSecurityAlert)
				audit.POST("/compliance/reports", auditHandler.GenerateComplianceReport)
				audit.POST("/cleanup", auditHandler.CleanupOldLogs)
			}
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
					"Email Verification",
					"Password Reset via Email",
					"Admin Dashboard",
					"User Management",
					"Role-Based Access Control",
					"Activity Logging",
					"Comprehensive Audit Trail",
					"Security Event Monitoring",
					"Compliance Reporting",
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
						"register":          "POST /api/auth/register",
						"login":             "POST /api/auth/login",
						"refresh":           "POST /api/auth/refresh",
						"logout":            "POST /api/auth/logout",
						"profile":           "GET /api/auth/profile",
						"update_profile":    "PUT /api/auth/profile",
						"change_password":   "POST /api/auth/change-password",
						"logout_all":        "POST /api/auth/logout-all",
						"send_verification": "POST /api/auth/send-verification",
						"verify_email":      "GET|POST /api/auth/verify-email",
						"password_reset":    "POST /api/auth/password-reset",
						"reset_password":    "POST /api/auth/reset-password",
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
					"email": gin.H{
						"stats":   "GET /api/email/stats",
						"cleanup": "POST /api/email/cleanup",
					},
					"admin": gin.H{
						"dashboard_stats": "GET /api/admin/stats",
						"users":           "GET /api/admin/users",
						"user_activity":   "GET /api/admin/users/:id/activity",
						"system_events":   "GET /api/admin/events",
						"update_status":   "PUT /api/admin/users/:id/status",
						"update_role":     "PUT /api/admin/users/:id/role",
						"delete_user":     "DELETE /api/admin/users/:id",
						"cleanup_logs":    "POST /api/admin/cleanup-logs",
						"system_info":     "GET /api/admin/system-info",
					},
					"audit": gin.H{
						"logs":               "GET /api/admin/audit/logs",
						"search":             "GET /api/admin/audit/search",
						"user_trail":         "GET /api/admin/audit/users/:id/trail",
						"resource_trail":     "GET /api/admin/audit/resources/:type/:id/trail",
						"high_risk":          "GET /api/admin/audit/high-risk",
						"security_alerts":    "GET /api/admin/audit/security-alerts",
						"resolve_alert":      "POST /api/admin/audit/security-alerts/:id/resolve",
						"compliance_reports": "POST /api/admin/audit/compliance/reports",
						"cleanup":            "POST /api/admin/audit/cleanup",
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
