package domain

// MediaParams is the unified parameter set for both text search and media discovery.
// When Query is non-empty the TMDB search endpoint is used and discover-only params
// (SortBy, YearFrom, YearTo, VoteAverageMin, VoteAverageMax) are ignored by TMDB.
// When Query is empty the TMDB discover endpoint is used.
type MediaParams struct {
	// Text search
	Query        string
	IncludeAdult bool

	// Shared by both modes
	Type     string // "movie", "series", or "" / "all" for both
	Language string // ISO 639-1, e.g. "fr"
	Page     int    // 1–1000

	// Discover mode only (ignored when Query is set)
	SortBy         string  // see SortBy* constants below
	YearFrom       int
	YearTo         int
	VoteAverageMin float32 // 0–10
	VoteAverageMax float32 // 0–10
	VoteCountMin   int     // minimum number of votes; 0 means no filter
}

// Supported SortBy values for discover mode.
const (
	SortByPopularityDesc  = "popularity.desc"
	SortByPopularityAsc   = "popularity.asc"
	SortByVoteAverageDesc = "vote_average.desc"
	SortByVoteAverageAsc  = "vote_average.asc"
	SortByReleaseDateDesc = "release_date.desc"
	SortByReleaseDateAsc  = "release_date.asc"
)
