// pkg/tmdb/client.go
package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
)

const baseURL = "https://api.themoviedb.org/3"

// Client holds the dependencies needed to talk to the TMDB API.
// By storing the apiKey and httpClient here, we avoid global variables
// and make this client easy to test by injecting a mock http.Client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient is a constructor function. This is the standard Go pattern
// for initializing a struct with dependencies.
// Always use constructors — never let callers build the struct manually,
// because they might forget to set required fields like the timeout.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Always set a timeout on HTTP clients.
		},
	}
}

// GetTrending fetches the trending movies for the day from TMDB.
// "day" and "week" are valid time windows — we default to "day".
func (c *Client) GetTrending() (*models.TMDBTrendingResponse, error) {
	url := fmt.Sprintf("%s/trending/movie/day?api_key=%s&language=en-US", baseURL, c.apiKey)

	// Build the request explicitly instead of using http.Get().
	// This gives us control to add headers, set context, etc. later on.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("tmdb: failed to build request: %w", err)
	}

	// Execute the request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tmdb: request failed: %w", err)
	}
	// CRITICAL: Always close the response body. If you forget this,
	// you will leak memory and exhaust your connection pool under load.
	defer resp.Body.Close()

	// Check for non-200 status codes BEFORE trying to decode the body.
	// A 401 or 404 response body is an error message, not valid data.
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tmdb: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Decode the JSON response body directly into our struct.
	// This is more memory-efficient than reading the body into a []byte first.
	var result models.TMDBTrendingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("tmdb: failed to decode response: %w", err)
	}

	return &result, nil
}

// SearchMovies calls TMDB's /search/movie endpoint.
// Best for: keyword/title searches. Supports year filter.
// Does NOT support: genre filtering, custom sort order.
func (c *Client) SearchMovies(params models.SearchParams) (*models.TMDBSearchResponse, error) {
	// url.Values handles URL encoding — spaces become %20, special
	// characters are escaped. Never build query strings with fmt.Sprintf.
	query := url.Values{}
	query.Set("api_key", c.apiKey)
	query.Set("language", "en-US")
	query.Set("include_adult", "false")
	query.Set("query", params.Query)

	if params.Year != "" {
		query.Set("primary_release_year", params.Year)
	}
	if params.Page != "" {
		query.Set("page", params.Page)
	} else {
		query.Set("page", "1")
	}

	endpoint := fmt.Sprintf("%s/search/movie?%s", baseURL, query.Encode())

	var result models.TMDBSearchResponse
	if err := c.get(endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DiscoverMovies calls TMDB's /discover/movie endpoint.
// Best for: browsing by genre, sorting, year filtering — no keyword.
// Does NOT support: free-text keyword search.
func (c *Client) DiscoverMovies(params models.SearchParams) (*models.TMDBSearchResponse, error) {
	query := url.Values{}
	query.Set("api_key", c.apiKey)
	query.Set("language", "en-US")
	query.Set("include_adult", "false")

	if params.GenreID != "" {
		query.Set("with_genres", params.GenreID)
	}
	if params.Year != "" {
		query.Set("primary_release_year", params.Year)
	}

	// Default sort is popularity descending — most relevant results first.
	// Valid TMDB sort values: popularity.asc, popularity.desc,
	// revenue.desc, primary_release_date.desc, vote_average.desc,
	// vote_count.desc
	sortBy := "popularity.desc"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	query.Set("sort_by", sortBy)

	if params.Page != "" {
		query.Set("page", params.Page)
	} else {
		query.Set("page", "1")
	}

	endpoint := fmt.Sprintf("%s/discover/movie?%s", baseURL, query.Encode())

	var result models.TMDBSearchResponse
	if err := c.get(endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGenres fetches the complete TMDB genre list.
// This is used to populate a genre dropdown in the frontend.
// Genres change extremely rarely — cache with a long TTL.
func (c *Client) GetGenres() (*models.TMDBGenreResponse, error) {
	endpoint := fmt.Sprintf("%s/genre/movie/list?api_key=%s&language=en-US", baseURL, c.apiKey)
	var result models.TMDBGenreResponse
	if err := c.get(endpoint, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// get is a private helper that executes a GET request and decodes
// the JSON response into the provided target.
// Centralising this eliminates the repetitive request/response/decode
// boilerplate from every public method.
func (c *Client) get(endpoint string, target any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("tmdb: failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tmdb: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tmdb: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("tmdb: failed to decode response: %w", err)
	}

	return nil
}
