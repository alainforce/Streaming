// internal/cache/movie_cache.go
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	// trendingCacheKey is the Redis key for the trending movies list.
	// Using a constant prevents typos scattering across the codebase.
	trendingCacheKey = "movies:trending"

	// trendingCacheTTL controls how long the data lives in Redis.
	// Trending movies don't change second-to-second — 5 minutes is
	// a sensible balance between freshness and API call reduction.
	trendingCacheTTL = 5 * time.Minute

	genresCacheKey = "movies:genres"
	// Genres change maybe once a year — 24 hours is safe.
	// Long TTL saves TMDB API calls on every page load that needs the genre list.
	genresCacheTTL = 24 * time.Hour

	// Search results are more volatile than trending — different queries,
	// different results. 10 minutes balances freshness with API savings.
	// A popular query like "inception" hit by 100 users in 10 minutes
	// becomes 1 TMDB API call instead of 100.
	searchCacheTTL = 10 * time.Minute
)

// MovieCache handles all Redis operations for movie data.
// Keeping cache logic in its own type means the service layer stays
// clean — it just calls SetTrending/GetTrending without knowing
// anything about Redis keys, TTLs, or serialization.
type MovieCache struct {
	client *redis.Client
}

// NewMovieCache is the constructor.
func NewMovieCache(client *redis.Client) *MovieCache {
	return &MovieCache{client: client}
}

// SetTrending serializes a movie list to JSON and stores it in Redis.
func (c *MovieCache) SetTrending(ctx context.Context, movies []models.Movie) error {
	// We must serialize to JSON because Redis stores bytes, not Go structs.
	data, err := json.Marshal(movies)
	if err != nil {
		return fmt.Errorf("cache: failed to marshal trending movies: %w", err)
	}

	// Set the key with an expiry (TTL). After trendingCacheTTL elapses,
	// Redis automatically deletes the key — no manual cleanup needed.
	if err := c.client.Set(ctx, trendingCacheKey, data, trendingCacheTTL).Err(); err != nil {
		return fmt.Errorf("cache: failed to set trending movies: %w", err)
	}

	return nil
}

// GetTrending retrieves and deserializes the cached trending movies.
// Returns (nil, nil) on a cache miss — the caller treats nil as "not cached".
func (c *MovieCache) GetTrending(ctx context.Context) ([]models.Movie, error) {
	data, err := c.client.Get(ctx, trendingCacheKey).Bytes()
	if err != nil {
		// redis.Nil is the specific sentinel error go-redis returns on a
		// cache miss (key doesn't exist or has expired).
		// This is NOT a real error — it's expected behaviour.
		// We return (nil, nil) so the caller knows to fetch from the source.
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss — not an error
		}
		// Any other error IS unexpected (connection lost, OOM, etc.)
		return nil, fmt.Errorf("cache: failed to get trending movies: %w", err)
	}

	var movies []models.Movie
	if err := json.Unmarshal(data, &movies); err != nil {
		return nil, fmt.Errorf("cache: failed to unmarshal trending movies: %w", err)
	}

	return movies, nil
}

// SetGenres caches the full genre list.
func (c *MovieCache) SetGenres(ctx context.Context, genres []models.Genre) error {
	data, err := json.Marshal(genres)
	if err != nil {
		return fmt.Errorf("cache: failed to marshal genres: %w", err)
	}
	return c.client.Set(ctx, genresCacheKey, data, genresCacheTTL).Err()
}

// GetGenres retrieves cached genres. Returns nil on miss — not an error.
func (c *MovieCache) GetGenres(ctx context.Context) ([]models.Genre, error) {
	data, err := c.client.Get(ctx, genresCacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // cache miss
		}
		return nil, fmt.Errorf("cache: failed to get genres: %w", err)
	}
	var genres []models.Genre
	if err := json.Unmarshal(data, &genres); err != nil {
		return nil, fmt.Errorf("cache: failed to unmarshal genres: %w", err)
	}
	return genres, nil
}

// SetSearch stores search results under a params-derived cache key.
// Each unique combination of filters gets its own Redis key.
func (c *MovieCache) SetSearch(ctx context.Context, params models.SearchParams, result *models.SearchResponse) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("cache: failed to marshal search results: %w", err)
	}
	key := params.CacheKey()
	return c.client.Set(ctx, key, data, searchCacheTTL).Err()
}

// GetSearch retrieves cached search results for the given params.
// Returns nil on miss — not an error.
func (c *MovieCache) GetSearch(ctx context.Context, params models.SearchParams) (*models.SearchResponse, error) {
	key := params.CacheKey()
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // cache miss
		}
		return nil, fmt.Errorf("cache: failed to get search results: %w", err)
	}
	var result models.SearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("cache: failed to unmarshal search results: %w", err)
	}
	return &result, nil
}
