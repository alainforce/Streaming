// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RateLimiterConfig defines limits for different route groups.
type RateLimiterConfig struct {
	// General API limit: 60 requests per minute per IP
	GeneralRate string
	// Auth endpoints limit: 10 requests per minute per IP
	// Tighter because brute-force attacks target login endpoints.
	AuthRate string
}

// DefaultRateLimiterConfig returns sensible defaults.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		GeneralRate: "60-M", // 60 per minute
		AuthRate:    "10-M", // 10 per minute
	}
}

// NewRateLimiter creates a Redis-backed rate limiter middleware for Gin.
// Storing state in Redis means the limit works correctly even if you
// run multiple instances of your API (horizontal scaling).
//
// Rate format: "<limit>-<period>"
//
//	S = second, M = minute, H = hour, D = day
//	Examples: "100-S", "60-M", "1000-H"
func NewRateLimiter(redisClient *redis.Client, rate string) (gin.HandlerFunc, error) {
	// Parse the rate string into a limiter.Rate struct.
	parsedRate, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		return nil, err
	}

	// Create the Redis store. This is where counters are persisted.
	store, err := redisstore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		// All rate limit keys are prefixed to avoid collisions with
		// your other Redis keys (like "movies:trending").
		Prefix: "ratelimit",
	})
	if err != nil {
		return nil, err
	}

	// Build the limiter instance.
	instance := limiter.New(store, parsedRate,
		// When the limit is exceeded, return 429 Too Many Requests.
		limiter.WithTrustForwardHeader(true), // Respect X-Forwarded-For behind a proxy
	)

	// Wrap it in the Gin adapter with a custom error handler.
	middleware := ginlimiter.NewMiddleware(instance,
		ginlimiter.WithLimitReachedHandler(func(c *gin.Context) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests, please slow down",
			})
			c.Abort()
		}),
	)

	return middleware, nil
}
