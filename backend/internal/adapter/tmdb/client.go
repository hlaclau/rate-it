package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
)

const baseURL = "https://api.themoviedb.org/3"

// Client implements port.MediaFetcher against the TMDB API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type movieResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	PosterPath  *string `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
}

type seriesResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	PosterPath   *string `json:"poster_path"`
	FirstAirDate string  `json:"first_air_date"`
}

func (c *Client) FetchMovie(id string) ([]byte, *domain.Media, error) {
	raw, err := c.get("/movie/" + id)
	if err != nil {
		return nil, nil, err
	}

	var r movieResponse
	if err = json.Unmarshal(raw, &r); err != nil {
		return nil, nil, fmt.Errorf("decode movie: %w", err)
	}

	m := &domain.Media{
		ExternalID:  strconv.Itoa(r.ID),
		Source:      domain.SourceTMDB,
		Type:        domain.TypeMovie,
		Title:       r.Title,
		PosterPath:  r.PosterPath,
		ReleaseYear: parseYear(r.ReleaseDate),
	}
	return raw, m, nil
}

func (c *Client) FetchSeries(id string) ([]byte, *domain.Media, error) {
	raw, err := c.get("/tv/" + id)
	if err != nil {
		return nil, nil, err
	}

	var r seriesResponse
	if err = json.Unmarshal(raw, &r); err != nil {
		return nil, nil, fmt.Errorf("decode series: %w", err)
	}

	m := &domain.Media{
		ExternalID:  strconv.Itoa(r.ID),
		Source:      domain.SourceTMDB,
		Type:        domain.TypeSeries,
		Title:       r.Name,
		PosterPath:  r.PosterPath,
		ReleaseYear: parseYear(r.FirstAirDate),
	}
	return raw, m, nil
}

// mediaResponse is used internally to parse, annotate, and re-serialize list results.
type mediaResponse struct {
	Page         int                      `json:"page"`
	Results      []map[string]interface{} `json:"results"`
	TotalPages   int                      `json:"total_pages"`
	TotalResults int                      `json:"total_results"`
}

// FetchMedia routes to TMDB search or discover based on whether a query is present.
func (c *Client) FetchMedia(p domain.MediaParams) ([]byte, error) {
	if p.Query != "" {
		return c.search(p)
	}
	return c.discover(p)
}

func (c *Client) search(p domain.MediaParams) ([]byte, error) {
	v := url.Values{"query": {p.Query}}
	if p.Page > 1 {
		v.Set("page", strconv.Itoa(p.Page))
	}
	if p.Language != "" {
		v.Set("language", p.Language)
	}
	if p.IncludeAdult {
		v.Set("include_adult", "true")
	}
	return c.get("/search/multi?" + v.Encode())
}

func (c *Client) discover(p domain.MediaParams) ([]byte, error) {
	switch p.Type {
	case "movie":
		return c.discoverOne("/discover/movie", "movie", p)
	case "series", "tv":
		return c.discoverOne("/discover/tv", "tv", p)
	default: // "" or "all"
		return c.discoverAll(p)
	}
}

func (c *Client) discoverOne(endpoint, mediaType string, p domain.MediaParams) ([]byte, error) {
	raw, err := c.get(endpoint + "?" + buildDiscoverValues(endpoint, p).Encode())
	if err != nil {
		return nil, err
	}
	var resp mediaResponse
	if err = json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode discover response: %w", err)
	}
	for i := range resp.Results {
		resp.Results[i]["media_type"] = mediaType
	}
	return json.Marshal(resp)
}

func (c *Client) discoverAll(p domain.MediaParams) ([]byte, error) {
	type res struct {
		data []byte
		err  error
	}
	movieCh := make(chan res, 1)
	tvCh := make(chan res, 1)

	go func() {
		raw, err := c.get("/discover/movie?" + buildDiscoverValues("/discover/movie", p).Encode())
		movieCh <- res{raw, err}
	}()
	go func() {
		raw, err := c.get("/discover/tv?" + buildDiscoverValues("/discover/tv", p).Encode())
		tvCh <- res{raw, err}
	}()

	mr, tr := <-movieCh, <-tvCh
	if mr.err != nil {
		return nil, mr.err
	}
	if tr.err != nil {
		return nil, tr.err
	}

	var movieResp, tvResp mediaResponse
	if err := json.Unmarshal(mr.data, &movieResp); err != nil {
		return nil, fmt.Errorf("decode movie discover: %w", err)
	}
	if err := json.Unmarshal(tr.data, &tvResp); err != nil {
		return nil, fmt.Errorf("decode tv discover: %w", err)
	}
	for i := range movieResp.Results {
		movieResp.Results[i]["media_type"] = "movie"
	}
	for i := range tvResp.Results {
		tvResp.Results[i]["media_type"] = "tv"
	}
	merged := mediaResponse{
		Page:         movieResp.Page,
		Results:      append(movieResp.Results, tvResp.Results...),
		TotalPages:   max(movieResp.TotalPages, tvResp.TotalPages),
		TotalResults: movieResp.TotalResults + tvResp.TotalResults,
	}
	return json.Marshal(merged)
}

// buildDiscoverValues constructs url.Values for a discover endpoint.
// release_date sort keys are remapped to first_air_date for the TV endpoint.
func buildDiscoverValues(endpoint string, p domain.MediaParams) url.Values {
	isTV := endpoint == "/discover/tv"
	v := url.Values{}

	sortBy := p.SortBy
	if sortBy == "" {
		sortBy = domain.SortByPopularityDesc
	}
	if isTV {
		if sortBy == domain.SortByReleaseDateDesc {
			sortBy = "first_air_date.desc"
		} else if sortBy == domain.SortByReleaseDateAsc {
			sortBy = "first_air_date.asc"
		}
	}
	v.Set("sort_by", sortBy)

	dateGTE, dateLTE := "primary_release_date.gte", "primary_release_date.lte"
	if isTV {
		dateGTE, dateLTE = "first_air_date.gte", "first_air_date.lte"
	}
	if p.YearFrom > 0 {
		v.Set(dateGTE, fmt.Sprintf("%d-01-01", p.YearFrom))
	}
	if p.YearTo > 0 {
		v.Set(dateLTE, fmt.Sprintf("%d-12-31", p.YearTo))
	}
	if p.VoteAverageMin > 0 {
		v.Set("vote_average.gte", strconv.FormatFloat(float64(p.VoteAverageMin), 'f', 1, 32))
	}
	if p.VoteAverageMax > 0 {
		v.Set("vote_average.lte", strconv.FormatFloat(float64(p.VoteAverageMax), 'f', 1, 32))
	}
	if p.VoteCountMin > 0 {
		v.Set("vote_count.gte", strconv.Itoa(p.VoteCountMin))
	}
	if p.Language != "" {
		v.Set("with_original_language", p.Language)
	}
	if p.Page > 1 {
		v.Set("page", strconv.Itoa(p.Page))
	}
	return v
}

func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tmdb request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, port.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tmdb status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

// parseYear extracts the year from a "YYYY-MM-DD" string, returning nil on failure.
func parseYear(date string) *int16 {
	if len(date) < 4 {
		return nil
	}
	y, err := strconv.Atoi(date[:4])
	if err != nil {
		return nil
	}
	year := int16(y)
	return &year
}
