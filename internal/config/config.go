package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	// Server settings
	Port string
	Host string
	Env  string

	// Database settings
	DatabaseURL string

	// JWT settings
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	// OAuth2 settings
	GoogleClientID     string
	GoogleClientSecret string
	OAuthRedirectURL   string

	// Security settings
	BcryptCost      int
	RateLimitPerMin int

	// CORS settings
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server settings
		Port: getEnv("PORT", "8080"),
		Host: getEnv("HOST", "localhost"),
		Env:  getEnv("ENV", "development"),

		// Database settings
		DatabaseURL: getEnv("DATABASE_URL", "mongodb://localhost:27017"),

		// JWT settings
		JWTSecret:       getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")),
		RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_TTL", "168h")), // 7 days

		// OAuth2 settings
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		OAuthRedirectURL:   getEnv("OAUTH_REDIRECT_URL", "http://localhost:8080/api/auth/oauth/google/callback"),

		// Security settings
		BcryptCost:      parseInt(getEnv("BCRYPT_COST", "12")),
		RateLimitPerMin: parseInt(getEnv("RATE_LIMIT_PER_MIN", "60")),

		// CORS settings
		AllowedOrigins: []string{
			getEnv("ALLOWED_ORIGIN_1", "http://localhost:3000"),
			getEnv("ALLOWED_ORIGIN_2", "http://localhost:3001"),
		},
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Host + ":" + c.Port
}

// Validate validates the configuration
func (c *Config) Validate() error {
	required := map[string]string{
		"JWT_SECRET": c.JWTSecret,
	}

	for env, value := range required {
		if value == "" {
			return &ConfigError{Field: env, Message: "is required"}
		}
	}

	if c.IsProduction() {
		prodRequired := map[string]string{
			"GOOGLE_CLIENT_ID":     c.GoogleClientID,
			"GOOGLE_CLIENT_SECRET": c.GoogleClientSecret,
		}

		for env, value := range prodRequired {
			if value == "" {
				return &ConfigError{Field: env, Message: "is required in production"}
			}
		}
	}

	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + " " + e.Message
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
