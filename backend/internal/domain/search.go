package domain

type MediaParams struct {
	Query        string
	IncludeAdult bool

	Type     string
	Language string
	Page     int

	SortBy         string
	YearFrom       int
	YearTo         int
	VoteAverageMin float32
	VoteAverageMax float32
	VoteCountMin   int

	WithGenres     string
	WatchProviders string
	WatchRegion    string
}

const (
	SortByPopularityDesc  = "popularity.desc"
	SortByPopularityAsc   = "popularity.asc"
	SortByVoteAverageDesc = "vote_average.desc"
	SortByVoteAverageAsc  = "vote_average.asc"
	SortByReleaseDateDesc = "release_date.desc"
	SortByReleaseDateAsc  = "release_date.asc"
)
