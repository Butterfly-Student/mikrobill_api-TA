package gin_inbound_adapter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"MikrOps/utils/redis"
)

// PublicRegistrationRateLimit - Rate limiting for public registration endpoint
// Limit: 5 requests per hour per IP address
func (h *middlewareAdapter) PublicRegistrationRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:register:%s", clientIP)
		ctx := c.Request.Context()

		limit := int64(5)       // 5 requests
		window := 1 * time.Hour // per hour

		// Increment counter
		count, err := redis.Incr(ctx, key)
		if err != nil {
			// Log error but allow request to proceed
			c.Next()
			return
		}

		// Set TTL on first request
		if count == 1 {
			redis.Expire(ctx, key, window)
		}

		// Check if limit exceeded
		if count > limit {
			ttl, _ := redis.TTL(ctx, key)
			resetTime := time.Now().Add(ttl).Unix()

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			c.Header("Retry-After", fmt.Sprintf("%d", int64(ttl.Seconds())))

			SendAbort(c, http.StatusTooManyRequests,
				"Registration rate limit exceeded. Please try again later.")
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))

		ttl, _ := redis.TTL(ctx, key)
		resetTime := time.Now().Add(ttl).Unix()
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		c.Next()
	}
}

// CustomerLoginRateLimit - Rate limiting for customer login endpoint
// Limit: 10 requests per hour per IP address (prevents brute force attacks)
func (h *middlewareAdapter) CustomerLoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:customer:login:%s", clientIP)
		ctx := c.Request.Context()

		limit := int64(10)      // 10 requests
		window := 1 * time.Hour // per hour

		// Increment counter
		count, err := redis.Incr(ctx, key)
		if err != nil {
			// Log error but allow request to proceed
			c.Next()
			return
		}

		// Set TTL on first request
		if count == 1 {
			redis.Expire(ctx, key, window)
		}

		// Check if limit exceeded
		if count > limit {
			ttl, _ := redis.TTL(ctx, key)
			resetTime := time.Now().Add(ttl).Unix()

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			c.Header("Retry-After", fmt.Sprintf("%d", int64(ttl.Seconds())))

			SendAbort(c, http.StatusTooManyRequests,
				"Login rate limit exceeded. Too many login attempts. Please try again later.")
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))

		ttl, _ := redis.TTL(ctx, key)
		resetTime := time.Now().Add(ttl).Unix()
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		c.Next()
	}
}

// CustomerPortalRateLimit - Rate limiting for customer portal endpoints
// Limit: 100 requests per minute per authenticated customer
func (h *middlewareAdapter) CustomerPortalRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get customer ID from context (set by CustomerAuth middleware)
		customerIDValue, exists := c.Get("customer_id")
		if !exists {
			// If no customer ID, fall back to IP-based limiting
			clientIP := c.ClientIP()
			customerIDValue = clientIP
		}

		customerID := fmt.Sprintf("%v", customerIDValue)
		key := fmt.Sprintf("ratelimit:customer:portal:%s", customerID)
		ctx := c.Request.Context()

		limit := int64(100)       // 100 requests
		window := 1 * time.Minute // per minute

		// Increment counter
		count, err := redis.Incr(ctx, key)
		if err != nil {
			// Log error but allow request to proceed
			c.Next()
			return
		}

		// Set TTL on first request
		if count == 1 {
			redis.Expire(ctx, key, window)
		}

		// Check if limit exceeded
		if count > limit {
			ttl, _ := redis.TTL(ctx, key)
			resetTime := time.Now().Add(ttl).Unix()

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			c.Header("Retry-After", fmt.Sprintf("%d", int64(ttl.Seconds())))

			SendAbort(c, http.StatusTooManyRequests,
				"Too many requests. Please slow down and try again later.")
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))

		ttl, _ := redis.TTL(ctx, key)
		resetTime := time.Now().Add(ttl).Unix()
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))

		c.Next()
	}
}
