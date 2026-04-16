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

func (c *Client) SearchMovies(query string) ([]byte, error) {
	return c.get("/search/movie?query=" + url.QueryEscape(query) + "&page=1")
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
