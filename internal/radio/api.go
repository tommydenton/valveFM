package radio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://all.api.radio-browser.info"
	requestTimeout = 12 * time.Second
	maxPageSize    = 500
)

type Client struct {
	baseURL   string
	userAgent string
	http      *http.Client
}

type serverInfo struct {
	Name string `json:"name"`
}

// NewClient creates a Radio Browser API client.
func NewClient(userAgent string) (*Client, error) {
	if strings.TrimSpace(userAgent) == "" {
		return nil, errors.New("user agent is required")
	}

	client := &Client{
		baseURL:   defaultBaseURL,
		userAgent: userAgent,
		http:      &http.Client{Timeout: requestTimeout},
	}

	baseURL, err := client.pickRandomServer()
	if err == nil && baseURL != "" {
		client.baseURL = baseURL
	}

	return client, nil
}

// StationsByCountry fetches stations by ISO country code.
func (c *Client) StationsByCountry(ctx context.Context, countryCode string) ([]Station, error) {
	return c.StationsByCountryPage(ctx, countryCode, 200, 0)
}

// StationsByCountryPage fetches stations for a country with pagination.
func (c *Client) StationsByCountryPage(ctx context.Context, countryCode string, limit int, offset int) ([]Station, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode == "" {
		return nil, errors.New("country code is required")
	}

	limit, offset, err := sanitizePage(limit, offset)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/json/stations/bycountrycodeexact/%s", url.PathEscape(countryCode))
	query := stationQuery(limit, offset)

	reqURL := c.baseURL + endpoint + "?" + query.Encode()
	var stations []Station
	if err := c.doJSON(ctx, reqURL, &stations); err != nil {
		return nil, err
	}
	return stations, nil
}

// SearchStationsByCountry searches stations by name and tags within a country, with pagination.
func (c *Client) SearchStationsByCountry(ctx context.Context, countryCode string, search string, limit int, offset int) ([]Station, error) {
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode == "" {
		return nil, errors.New("country code is required")
	}

	search = strings.TrimSpace(search)
	if search == "" {
		return nil, errors.New("search query is required")
	}

	limit, offset, err := sanitizePage(limit, offset)
	if err != nil {
		return nil, err
	}

	stations, err := c.searchStationsByCountry(ctx, countryCode, "name", search, limit, offset)
	if err != nil {
		return nil, err
	}
	if len(stations) > 0 || offset > 0 {
		return stations, nil
	}

	// Fallback for tag-only discoverability when no name matches on the first page.
	stations, err = c.searchStationsByCountry(ctx, countryCode, "tag", search, limit, offset)
	if err != nil {
		return nil, err
	}
	return stations, nil
}

// Countries fetches available countries from the API.
func (c *Client) Countries(ctx context.Context) ([]Country, error) {
	reqURL := c.baseURL + "/json/countries"
	var countries []Country
	if err := c.doJSON(ctx, reqURL, &countries); err != nil {
		return nil, err
	}

	sort.Slice(countries, func(i, j int) bool {
		return strings.ToLower(countries[i].Name) < strings.ToLower(countries[j].Name)
	})
	return countries, nil
}

// ResolveStationURL calls /json/url/{stationuuid} and returns a resolved stream URL.
func (c *Client) ResolveStationURL(ctx context.Context, uuid string) (string, error) {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return "", errors.New("station uuid is required")
	}

	endpoint := fmt.Sprintf("/json/url/%s", url.PathEscape(uuid))
	reqURL := c.baseURL + endpoint

	data, err := c.getBytes(ctx, reqURL)
	if err != nil {
		return "", err
	}

	var station Station
	if err := json.Unmarshal(data, &station); err == nil && station.UUID != "" {
		return resolvedURL(station)
	}

	var stations []Station
	if err := json.Unmarshal(data, &stations); err != nil {
		return "", err
	}
	if len(stations) == 0 {
		return "", errors.New("no station data returned")
	}
	return resolvedURL(stations[0])
}

func resolvedURL(station Station) (string, error) {
	if strings.TrimSpace(station.URLResolved) != "" {
		return station.URLResolved, nil
	}
	if strings.TrimSpace(station.URL) != "" {
		return station.URL, nil
	}
	return "", errors.New("station has no stream url")
}

func (c *Client) doJSON(ctx context.Context, reqURL string, target any) error {
	data, err := c.getBytes(ctx, reqURL)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func (c *Client) getBytes(ctx context.Context, reqURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	// Limit response size to 10MB to prevent OOM on malformed responses
	return io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
}

func (c *Client) pickRandomServer() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	reqURL := defaultBaseURL + "/json/servers"
	var servers []serverInfo
	if err := c.doJSON(ctx, reqURL, &servers); err != nil {
		return "", err
	}
	if len(servers) == 0 {
		return "", errors.New("no api servers returned")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	choice := servers[r.Intn(len(servers))].Name
	if strings.TrimSpace(choice) == "" {
		return "", errors.New("empty server name")
	}
	return "https://" + choice, nil
}

func stationQuery(limit int, offset int) url.Values {
	query := url.Values{}
	query.Set("hidebroken", "true")
	query.Set("order", "clickcount")
	query.Set("reverse", "true")
	query.Set("limit", fmt.Sprintf("%d", limit))
	query.Set("offset", fmt.Sprintf("%d", offset))
	return query
}

func sanitizePage(limit int, offset int) (int, int, error) {
	if limit <= 0 {
		return 0, 0, errors.New("limit must be greater than zero")
	}
	if limit > maxPageSize {
		return 0, 0, fmt.Errorf("limit must be <= %d", maxPageSize)
	}
	if offset < 0 {
		return 0, 0, errors.New("offset must be >= 0")
	}
	return limit, offset, nil
}

func (c *Client) searchStationsByCountry(ctx context.Context, countryCode string, field string, value string, limit int, offset int) ([]Station, error) {
	query := stationQuery(limit, offset)
	query.Set("countrycodeexact", countryCode)
	query.Set(field, value)

	reqURL := c.baseURL + "/json/stations/search?" + query.Encode()
	var stations []Station
	if err := c.doJSON(ctx, reqURL, &stations); err != nil {
		return nil, err
	}
	return stations, nil
}
