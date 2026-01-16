package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultLoginRateLimit is the default number of login attempts per minute
	DefaultLoginRateLimit = 5

	// DefaultRateLimitWindow is the default time window for rate limiting
	DefaultRateLimitWindow = time.Minute
)

// RateLimiter handles rate limiting using Redis
type RateLimiter struct {
	client    *redis.Client
	limit     int
	window    time.Duration
	keyPrefix string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *redis.Client, limit int, window time.Duration, keyPrefix string) *RateLimiter {
	if limit <= 0 {
		limit = DefaultLoginRateLimit
	}
	if window <= 0 {
		window = DefaultRateLimitWindow
	}
	if keyPrefix == "" {
		keyPrefix = "rate_limit"
	}

	return &RateLimiter{
		client:    client,
		limit:     limit,
		window:    window,
		keyPrefix: keyPrefix,
	}
}

// Allow checks if the request should be allowed based on rate limiting
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, int, time.Duration, error) {
	fullKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)

	// Use a pipeline for atomic operations
	pipe := rl.client.Pipeline()

	// Increment the counter
	incrCmd := pipe.Incr(ctx, fullKey)

	// Set expiration on first request
	pipe.Expire(ctx, fullKey, rl.window)

	// Get TTL
	ttlCmd := pipe.TTL(ctx, fullKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := int(incrCmd.Val())
	ttl := ttlCmd.Val()

	remaining := rl.limit - count
	if remaining < 0 {
		remaining = 0
	}

	return count <= rl.limit, remaining, ttl, nil
}

// Reset resets the rate limit counter for a key
func (rl *RateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", rl.keyPrefix, key)
	return rl.client.Del(ctx, fullKey).Err()
}

// LoginRateLimit creates a middleware that limits login attempts
func LoginRateLimit(redisClient *redis.Client, limit int) gin.HandlerFunc {
	limiter := NewRateLimiter(redisClient, limit, DefaultRateLimitWindow, "login_rate_limit")

	return func(c *gin.Context) {
		// Use client IP as the rate limit key
		clientIP := c.ClientIP()

		allowed, remaining, retryAfter, err := limiter.Allow(c.Request.Context(), clientIP)
		if err != nil {
			// If Redis is unavailable, allow the request but log the error
			c.Set("rate_limit_error", err.Error())
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(retryAfter).Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many login attempts",
				"message":     fmt.Sprintf("rate limit exceeded, please try again in %d seconds", int(retryAfter.Seconds())),
				"retry_after": int(retryAfter.Seconds()),
			})
			return
		}

		c.Next()
	}
}

// ResetLoginRateLimit resets the login rate limit for an IP
func ResetLoginRateLimit(redisClient *redis.Client, clientIP string) error {
	limiter := NewRateLimiter(redisClient, DefaultLoginRateLimit, DefaultRateLimitWindow, "login_rate_limit")
	return limiter.Reset(context.Background(), clientIP)
}
