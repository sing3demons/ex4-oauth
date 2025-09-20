package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a token bucket rate limiter
type RateLimiter struct {
	tokens     int           // Current number of tokens
	maxTokens  int           // Maximum number of tokens
	refillRate time.Duration // Time between token refills
	lastRefill time.Time     // Last refill time
	mutex      sync.Mutex    // Mutex for thread safety
}

// ClientRateLimiters holds rate limiters for different clients
type ClientRateLimiters struct {
	limiters map[string]*RateLimiter
	mutex    sync.RWMutex
}

// Global rate limiter storage
var (
	globalLimiters = &ClientRateLimiters{
		limiters: make(map[string]*RateLimiter),
	}
)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Refill tokens based on time elapsed
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// GetRateLimiter gets or creates a rate limiter for a client
func (crl *ClientRateLimiters) GetRateLimiter(clientID string, maxTokens int, refillRate time.Duration) *RateLimiter {
	crl.mutex.RLock()
	limiter, exists := crl.limiters[clientID]
	crl.mutex.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	crl.mutex.Lock()
	defer crl.mutex.Unlock()

	// Double-check in case another goroutine created it
	if limiter, exists := crl.limiters[clientID]; exists {
		return limiter
	}

	limiter = NewRateLimiter(maxTokens, refillRate)
	crl.limiters[clientID] = limiter
	return limiter
}

// TokenBucketRateLimitMiddleware creates a rate limiting middleware with token bucket algorithm
func TokenBucketRateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	refillRate := window / time.Duration(maxRequests)

	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientID := c.ClientIP()

		// Special handling for localhost/development
		if clientID == "::1" || clientID == "127.0.0.1" {
			clientID = "localhost"
		}

		// Get rate limiter for this client
		limiter := globalLimiters.GetRateLimiter(clientID, maxRequests, refillRate)

		// Check if request is allowed
		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", string(rune(maxRequests)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", string(rune(time.Now().Add(refillRate).Unix())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":             "rate_limit_exceeded",
				"error_description": "Too many requests. Please try again later.",
				"retry_after":       int(refillRate.Seconds()),
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(maxRequests)))

		c.Next()
	}
}

// StrictLoginRateLimitMiddleware creates a stricter rate limiter for login attempts
func StrictLoginRateLimitMiddleware() gin.HandlerFunc {
	// Allow 5 login attempts per minute per IP
	return TokenBucketRateLimitMiddleware(5, time.Minute)
}

// EnhancedAPIRateLimitMiddleware creates a general API rate limiter
func EnhancedAPIRateLimitMiddleware() gin.HandlerFunc {
	// Allow 100 requests per minute per IP
	return TokenBucketRateLimitMiddleware(100, time.Minute)
}

// OAuthRateLimitMiddleware creates rate limiter for OAuth2 endpoints
func OAuthRateLimitMiddleware() gin.HandlerFunc {
	// Allow 20 OAuth requests per minute per IP
	return TokenBucketRateLimitMiddleware(20, time.Minute)
}

// CleanupExpiredLimiters removes old rate limiters to prevent memory leaks
func CleanupExpiredLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			globalLimiters.mutex.Lock()
			now := time.Now()

			for clientID, limiter := range globalLimiters.limiters {
				limiter.mutex.Lock()

				// Remove limiters that haven't been used for 10 minutes
				if now.Sub(limiter.lastRefill) > 10*time.Minute {
					delete(globalLimiters.limiters, clientID)
				}

				limiter.mutex.Unlock()
			}

			globalLimiters.mutex.Unlock()
		}
	}()
}

// GetRateLimitStatus returns current rate limit status for monitoring
func GetRateLimitStatus() map[string]interface{} {
	globalLimiters.mutex.RLock()
	defer globalLimiters.mutex.RUnlock()

	status := make(map[string]interface{})
	status["active_limiters"] = len(globalLimiters.limiters)

	clientStatus := make(map[string]interface{})
	for clientID, limiter := range globalLimiters.limiters {
		limiter.mutex.Lock()
		clientStatus[clientID] = map[string]interface{}{
			"tokens":      limiter.tokens,
			"max_tokens":  limiter.maxTokens,
			"last_refill": limiter.lastRefill,
		}
		limiter.mutex.Unlock()
	}
	status["clients"] = clientStatus

	return status
}
