// internal/handlers/movie_handler.go
package handlers

import (
	"net/http"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/services"
	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	movieService *services.MovieService
}

func NewMovieHandler(movieService *services.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

func (h *MovieHandler) GetTrending(c *gin.Context) {
	// c.Request.Context() carries the request's deadline and cancellation
	// signal. If the client disconnects, this context is cancelled, which
	// propagates down through the service and cache layers automatically.
	movies, err := h.movieService.GetTrending(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch trending movies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(movies),
		"data":   movies,
	})
}

// SearchMovies handles GET /movies/search
//
// Query parameters:
//
//	q         string  keyword search (e.g. "inception")
//	genre     string  TMDB genre ID (e.g. "28" for Action)
//	year      string  release year (e.g. "2023")
//	sort_by   string  sort order (e.g. "popularity.desc", "vote_average.desc")
//	page      string  page number, defaults to 1
//
// At least one of q, genre, or year must be provided.
func (h *MovieHandler) SearchMovies(c *gin.Context) {
	params := models.SearchParams{
		Query:   c.Query("q"),
		GenreID: c.Query("genre"),
		Year:    c.Query("year"),
		SortBy:  c.Query("sort_by"),
		Page:    c.DefaultQuery("page", "1"),
	}

	// Require at least one search parameter — returning the entire
	// TMDB catalogue without any filter is not useful and wastes
	// API quota.
	if params.Query == "" && params.GenreID == "" && params.Year == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one search parameter is required: q, genre, or year",
		})
		return
	}

	// Validate sort_by against known valid values.
	// Passing an unknown sort value to TMDB returns a 422 error —
	// we catch it here with a clear message instead.
	if params.SortBy != "" && !isValidSortBy(params.SortBy) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid sort_by value",
			"allowed": validSortByValues(),
		})
		return
	}

	result, err := h.movieService.SearchMovies(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to search movies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetGenres handles GET /movies/genres
// Returns the full TMDB genre list for populating dropdowns.
func (h *MovieHandler) GetGenres(c *gin.Context) {
	genres, err := h.movieService.GetGenres(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch genres",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(genres),
		"data":   genres,
	})
}

// validSortOptions is the complete list of values TMDB's /discover endpoint accepts.
var validSortOptions = map[string]bool{
	"popularity.asc":            true,
	"popularity.desc":           true,
	"revenue.desc":              true,
	"primary_release_date.asc":  true,
	"primary_release_date.desc": true,
	"vote_average.asc":          true,
	"vote_average.desc":         true,
	"vote_count.desc":           true,
}

func isValidSortBy(s string) bool {
	return validSortOptions[s]
}

func validSortByValues() []string {
	values := make([]string, 0, len(validSortOptions))
	for k := range validSortOptions {
		values = append(values, k)
	}
	return values
}
