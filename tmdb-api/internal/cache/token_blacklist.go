// internal/cache/token_blacklist.go
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// blacklistPrefix namespaces token blacklist keys.
	// Full key looks like: "blacklist:f47ac10b-58cc-4372-a567-0e02b2c3d479"
	blacklistPrefix = "blacklist:"

	// banPrefix namespaces user ban status keys.
	// Full key looks like: "user_banned:42"
	// No TTL — ban persists until an admin explicitly unbans.
	banPrefix = "user_banned:"
)

// TokenBlacklist manages two Redis-backed security features:
//  1. Revoked token JTIs (for logout)
//  2. Banned user IDs (for ban enforcement on every request)
//
// Keeping both here is a deliberate choice — they're both
// request-time security checks in the auth middleware, they both
// use the same Redis client, and they're both simple key lookups.
type TokenBlacklist struct {
	client *redis.Client
}

func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

// BlacklistToken stores a jti in Redis until the token would have expired.
// After ttl elapses, Redis automatically removes the key — zero maintenance.
func (b *TokenBlacklist) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	// If remaining time is zero or negative, the token is already expired —
	// no need to blacklist it, just return quietly.
	if ttl <= 0 {
		return nil
	}

	key := blacklistPrefix + jti
	if err := b.client.Set(ctx, key, 1, ttl).Err(); err != nil {
		return fmt.Errorf("blacklist: failed to blacklist token %s: %w", jti, err)
	}

	return nil
}

// IsTokenBlacklisted returns true if the jti has been revoked.
// A return of (false, nil) means the token is clean — let it through.
func (b *TokenBlacklist) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := blacklistPrefix + jti

	_, err := b.client.Get(ctx, key).Result()
	if err != nil {
		// redis.Nil means the key doesn't exist → token is NOT blacklisted.
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("blacklist: failed to check token %s: %w", jti, err)
	}

	// Key exists → token has been revoked.
	return true, nil
}

// BanUser stores a ban marker for the given userID with no expiry.
// The marker persists until UnbanUser is called explicitly.
func (b *TokenBlacklist) BanUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", banPrefix, userID)
	// 0 TTL means no expiration — the key lives until explicitly deleted.
	if err := b.client.Set(ctx, key, 1, 0).Err(); err != nil {
		return fmt.Errorf("blacklist: failed to ban user %s: %w", userID, err)
	}
	return nil
}

// UnbanUser removes the ban marker, immediately restoring access.
func (b *TokenBlacklist) UnbanUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", banPrefix, userID)
	if err := b.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("blacklist: failed to unban user %s: %w", userID, err)
	}
	return nil
}

// IsUserBanned returns true if the user currently has an active ban.
func (b *TokenBlacklist) IsUserBanned(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("%s%s", banPrefix, userID)

	_, err := b.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // Key doesn't exist → user is not banned
		}
		return false, fmt.Errorf("blacklist: failed to check ban for user %s: %w", userID, err)
	}

	return true, nil
}
