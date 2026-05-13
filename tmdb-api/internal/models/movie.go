// internal/models/movie.go
package models

// TMDBTrendingResponse maps the top-level JSON response from TMDB's
// trending endpoint. We only map the fields we actually need.
// TMDB returns many more fields — we intentionally ignore them.
type TMDBTrendingResponse struct {
	Page    int     `json:"page"`
	Results []Movie `json:"results"`
}

// Movie represents a single movie from TMDB.
// The `json:"..."` tags control how Go serializes/deserializes the field.
// The tag name must exactly match the key in the TMDB JSON response.
type Movie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Overview     string  `json:"overview"`
	ReleaseDate  string  `json:"release_date"`
	PosterPath   string  `json:"poster_path"`
	VoteAverage  float64 `json:"vote_average"`
	VoteCount    int     `json:"vote_count"`
	Popularity   float64 `json:"popularity"`
	OriginalLang string  `json:"original_language"`
}

// Genre maps a TMDB genre ID to its human-readable name.
// e.g. {ID: 28, Name: "Action"}
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TMDBGenreResponse is the top-level response from /genre/movie/list
type TMDBGenreResponse struct {
	Genres []Genre `json:"genres"`
}

// TMDBSearchResponse is shared between /search/movie and /discover/movie —
// both return the same shape.
type TMDBSearchResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

// SearchParams holds every filter the user can pass as query parameters.
// We parse these once in the handler and pass the struct down through
// every layer — cleaner than passing 5 individual string arguments.
type SearchParams struct {
	Query   string // ?q=         free text keyword
	GenreID string // ?genre=     TMDB genre ID e.g. "28" for Action
	Year    string // ?year=      primary release year e.g. "2023"
	SortBy  string // ?sort_by=   e.g. "popularity.desc"
	Page    string // ?page=      pagination
}

// IsKeywordSearch returns true when the user provided a text query.
// This determines which TMDB endpoint we call.
func (p SearchParams) IsKeywordSearch() bool {
	return p.Query != ""
}

// HasFilters returns true when genre or sort filters were provided.
// Used to build the "filters ignored" warning in keyword search mode.
func (p SearchParams) HasFilters() bool {
	return p.GenreID != "" || p.SortBy != ""
}

// CacheKey builds a deterministic, sorted cache key from the params.
// Sorting the components means that params in different orders
// produce the same cache key.
// e.g. "search:genre=28&page=1&q=inception&sort_by=popularity.desc&year=2010"
func (p SearchParams) CacheKey() string {
	parts := []string{}
	if p.Query != "" {
		parts = append(parts, "q="+p.Query)
	}
	if p.GenreID != "" {
		parts = append(parts, "genre="+p.GenreID)
	}
	if p.Year != "" {
		parts = append(parts, "year="+p.Year)
	}
	if p.SortBy != "" {
		parts = append(parts, "sort_by="+p.SortBy)
	}
	if p.Page != "" {
		parts = append(parts, "page="+p.Page)
	}

	key := "search:"
	for i, part := range parts {
		if i > 0 {
			key += "&"
		}
		key += part
	}
	return key
}

// SearchResponse is what we return to the client.
// It wraps the results with pagination metadata and an optional
// warning when filters were ignored due to TMDB API limitations.
type SearchResponse struct {
	Page         int     `json:"page"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
	Results      []Movie `json:"results"`
	// Warning is non-empty when the user passed filters that couldn't
	// be applied. Being transparent about API limitations builds trust.
	Warning string `json:"warning,omitempty"`
}
