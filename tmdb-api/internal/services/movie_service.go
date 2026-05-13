// internal/services/movie_service.go
package services

import (
	"context"
	"log"

	"github.com/alainforce/streaming/tmdb-api/internal/cache"
	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/pkg/tmdb"
)

// MovieService now has two dependencies: the TMDB client and the cache.
// The cache is optional — if Redis is down, we fall back gracefully.
type MovieService struct {
	tmdbClient *tmdb.Client
	movieCache *cache.MovieCache
}

// NewMovieService now accepts a cache as a second argument.
func NewMovieService(tmdbClient *tmdb.Client, movieCache *cache.MovieCache) *MovieService {
	return &MovieService{
		tmdbClient: tmdbClient,
		movieCache: movieCache,
	}
}

// GetTrending implements the cache-aside pattern:
//  1. Check cache — if HIT, return immediately (fast path)
//  2. On MISS, fetch from TMDB (slow path)
//  3. Store fresh data in cache for next time
//  4. Return data to caller
//
// This pattern is also called "lazy loading" — we only populate the
// cache when someone actually asks for the data.
func (s *MovieService) GetTrending(ctx context.Context) ([]models.Movie, error) {
	// --- Step 1: Check the cache ---
	cached, err := s.movieCache.GetTrending(ctx)
	if err != nil {
		// Redis is reachable but returned an unexpected error.
		// We LOG it and continue to fetch from TMDB rather than
		// failing the entire request. The cache is an optimisation —
		// the application must work without it.
		log.Printf("WARN: cache read failed, falling back to TMDB: %v", err)
	}

	// Cache HIT: return immediately without touching TMDB
	if cached != nil {
		log.Println("INFO: trending cache HIT")
		return cached, nil
	}

	// --- Step 2: Cache MISS — fetch from TMDB ---
	log.Println("INFO: trending cache MISS, fetching from TMDB")

	response, err := s.tmdbClient.GetTrending()
	if err != nil {
		return nil, err
	}

	movies := response.Results

	// --- Step 3: Populate the cache for next time ---
	// We do this in a goroutine so the cache write doesn't slow down
	// the response to the client. Fire and forget.
	// We create a fresh context here because the request context (ctx)
	// may be cancelled by the time the goroutine runs.
	go func() {
		bgCtx := context.Background()
		if err := s.movieCache.SetTrending(bgCtx, movies); err != nil {
			log.Printf("WARN: failed to cache trending movies: %v", err)
		} else {
			log.Println("INFO: trending movies cached successfully")
		}
	}()

	// --- Step 4: Return the fresh data ---
	return movies, nil
}

// SearchMovies applies the cache-aside pattern to search.
// It routes to the correct TMDB endpoint based on whether
// a keyword query was provided.
func (s *MovieService) SearchMovies(ctx context.Context, params models.SearchParams) (*models.SearchResponse, error) {
	// --- Cache check ---
	cached, err := s.movieCache.GetSearch(ctx, params)
	if err != nil {
		log.Printf("WARN: search cache read failed: %v", err)
	}
	if cached != nil {
		log.Printf("INFO: search cache HIT for key: %s", params.CacheKey())
		return cached, nil
	}

	log.Printf("INFO: search cache MISS for key: %s", params.CacheKey())

	// --- Route to correct TMDB endpoint ---
	var tmdbResult *models.TMDBSearchResponse
	var warning string

	if params.IsKeywordSearch() {
		// Keyword search: use /search/movie
		tmdbResult, err = s.tmdbClient.SearchMovies(params)
		if err != nil {
			return nil, err
		}
		// Transparently warn the client if they passed filters that
		// TMDB's search endpoint doesn't support.
		// Hiding this would silently return wrong results.
		if params.HasFilters() {
			warning = "genre and sort filters are not supported when searching by keyword — only year filter was applied"
		}
	} else {
		// No keyword: use /discover/movie for filter-based browsing
		tmdbResult, err = s.tmdbClient.DiscoverMovies(params)
		if err != nil {
			return nil, err
		}
	}

	result := &models.SearchResponse{
		Page:         tmdbResult.Page,
		TotalPages:   tmdbResult.TotalPages,
		TotalResults: tmdbResult.TotalResults,
		Results:      tmdbResult.Results,
		Warning:      warning,
	}
	// --- Cache the result asynchronously ---
	go func() {
		if err := s.movieCache.SetSearch(context.Background(), params, result); err != nil {
			log.Printf("WARN: failed to cache search results: %v", err)
		} else {
			log.Printf("INFO: cached search results for key: %s", params.CacheKey())
		}
	}()

	return result, nil
}

// GetGenres applies cache-aside to the genre list.
// With a 24-hour TTL, this almost always hits the cache after
// the first call.
func (s *MovieService) GetGenres(ctx context.Context) ([]models.Genre, error) {
	cached, err := s.movieCache.GetGenres(ctx)
	if err != nil {
		log.Printf("WARN: genre cache read failed: %v", err)
	}
	if cached != nil {
		log.Println("INFO: genres cache HIT")
		return cached, nil
	}

	log.Println("INFO: genres cache MISS, fetching from TMDB")
	response, err := s.tmdbClient.GetGenres()
	if err != nil {
		return nil, err
	}

	genres := response.Genres
	go func() {
		if err := s.movieCache.SetGenres(context.Background(), genres); err != nil {
			log.Printf("WARN: failed to cache genres: %v", err)
		}
	}()

	return genres, nil
}
