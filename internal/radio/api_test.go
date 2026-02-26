package radio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient_RequiresUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		wantErr   bool
	}{
		{"valid user agent", "TestApp/1.0", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tabs only", "\t\t", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.userAgent)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient(%q) error = %v, wantErr %v", tt.userAgent, err, tt.wantErr)
			}
		})
	}
}

func TestClient_StationsByCountry(t *testing.T) {
	stations := []Station{
		{UUID: "uuid-1", Name: "Station 1", Country: "United States", CountryCode: "US", Bitrate: 128},
		{UUID: "uuid-2", Name: "Station 2", Country: "United States", CountryCode: "US", Bitrate: 256},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if !strings.Contains(r.URL.Path, "/json/stations/bycountrycodeexact/US") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("hidebroken") != "true" {
			t.Error("hidebroken should be true")
		}
		if r.URL.Query().Get("order") != "clickcount" {
			t.Error("order should be clickcount")
		}
		if r.URL.Query().Get("limit") != "200" {
			t.Error("limit should be 200")
		}
		if r.URL.Query().Get("offset") != "0" {
			t.Error("offset should default to 0")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stations)
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	result, err := client.StationsByCountry(ctx, "US")
	if err != nil {
		t.Fatalf("StationsByCountry() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("got %d stations, want 2", len(result))
	}
	if result[0].Name != "Station 1" {
		t.Errorf("first station name = %q, want %q", result[0].Name, "Station 1")
	}
}

func TestClient_StationsByCountryPage(t *testing.T) {
	var capturedPath string
	var capturedQuery = make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		for key := range r.URL.Query() {
			capturedQuery[key] = r.URL.Query().Get(key)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Station{})
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	_, err := client.StationsByCountryPage(context.Background(), "us", 125, 250)
	if err != nil {
		t.Fatalf("StationsByCountryPage() error = %v", err)
	}

	if !strings.Contains(capturedPath, "/json/stations/bycountrycodeexact/US") {
		t.Fatalf("unexpected path: %s", capturedPath)
	}
	if capturedQuery["limit"] != "125" {
		t.Errorf("limit = %q, want %q", capturedQuery["limit"], "125")
	}
	if capturedQuery["offset"] != "250" {
		t.Errorf("offset = %q, want %q", capturedQuery["offset"], "250")
	}
}

func TestClient_StationsByCountryPage_InvalidPaging(t *testing.T) {
	client := &Client{
		baseURL:   "http://example.com",
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	_, err := client.StationsByCountryPage(context.Background(), "US", 0, 0)
	if err == nil {
		t.Fatal("expected error for zero limit")
	}

	_, err = client.StationsByCountryPage(context.Background(), "US", 200, -1)
	if err == nil {
		t.Fatal("expected error for negative offset")
	}
}

func TestClient_SearchStationsByCountry(t *testing.T) {
	var capturedPath string
	var capturedQuery = make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		for key := range r.URL.Query() {
			capturedQuery[key] = r.URL.Query().Get(key)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Station{})
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	_, err := client.SearchStationsByCountry(context.Background(), "us", "rock", 100, 200)
	if err != nil {
		t.Fatalf("SearchStationsByCountry() error = %v", err)
	}

	if capturedPath != "/json/stations/search" {
		t.Fatalf("unexpected path: %s", capturedPath)
	}
	if capturedQuery["countrycodeexact"] != "US" {
		t.Errorf("countrycodeexact = %q, want %q", capturedQuery["countrycodeexact"], "US")
	}
	if capturedQuery["name"] != "rock" {
		t.Errorf("name = %q, want %q", capturedQuery["name"], "rock")
	}
	if capturedQuery["tag"] != "" {
		t.Errorf("tag should be empty for name-first search, got %q", capturedQuery["tag"])
	}
	if capturedQuery["limit"] != "100" {
		t.Errorf("limit = %q, want %q", capturedQuery["limit"], "100")
	}
	if capturedQuery["offset"] != "200" {
		t.Errorf("offset = %q, want %q", capturedQuery["offset"], "200")
	}
}

func TestClient_SearchStationsByCountry_FallbackToTag(t *testing.T) {
	requests := make([]map[string]string, 0, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := map[string]string{}
		for key := range r.URL.Query() {
			query[key] = r.URL.Query().Get(key)
		}
		requests = append(requests, query)

		w.Header().Set("Content-Type", "application/json")
		if query["name"] != "" {
			json.NewEncoder(w).Encode([]Station{})
			return
		}
		json.NewEncoder(w).Encode([]Station{{UUID: "tag-hit", Name: "Tag Match"}})
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	result, err := client.SearchStationsByCountry(context.Background(), "US", "classic", 50, 0)
	if err != nil {
		t.Fatalf("SearchStationsByCountry() error = %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
	if requests[0]["name"] != "classic" {
		t.Errorf("first request should be name search, got name=%q", requests[0]["name"])
	}
	if requests[1]["tag"] != "classic" {
		t.Errorf("second request should be tag search, got tag=%q", requests[1]["tag"])
	}
	if len(result) != 1 || result[0].UUID != "tag-hit" {
		t.Fatalf("unexpected fallback result: %+v", result)
	}
}

func TestClient_SearchStationsByCountry_RequiresSearch(t *testing.T) {
	client := &Client{
		baseURL:   "http://example.com",
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	_, err := client.SearchStationsByCountry(context.Background(), "US", "  ", 100, 0)
	if err == nil {
		t.Fatal("expected error for empty search query")
	}
}

func TestClient_StationsByCountry_EmptyCountryCode(t *testing.T) {
	client := &Client{
		baseURL:   "http://example.com",
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	_, err := client.StationsByCountry(ctx, "")
	if err == nil {
		t.Error("StationsByCountry() should return error for empty country code")
	}
}

func TestClient_StationsByCountry_NormalizesCountryCode(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Station{})
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	_, _ = client.StationsByCountry(ctx, "  us  ") // lowercase with spaces

	if !strings.Contains(capturedPath, "/US") {
		t.Errorf("country code should be normalized to uppercase, got path: %s", capturedPath)
	}
}

func TestClient_Countries(t *testing.T) {
	countries := []Country{
		{Code: "US", Name: "United States"},
		{Code: "GB", Name: "United Kingdom"},
		{Code: "AU", Name: "Australia"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/countries" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(countries)
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	result, err := client.Countries(ctx)
	if err != nil {
		t.Fatalf("Countries() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("got %d countries, want 3", len(result))
	}

	// Should be sorted alphabetically by name
	if result[0].Name != "Australia" {
		t.Errorf("first country should be Australia (sorted), got %q", result[0].Name)
	}
}

func TestClient_ResolveStationURL(t *testing.T) {
	tests := []struct {
		name        string
		response    interface{}
		expectedURL string
		wantErr     bool
	}{
		{
			name: "single station object",
			response: Station{
				UUID:        "test-uuid",
				URLResolved: "http://stream.example.com/live",
				URL:         "http://example.com",
			},
			expectedURL: "http://stream.example.com/live",
		},
		{
			name: "station array",
			response: []Station{{
				UUID:        "test-uuid",
				URLResolved: "http://stream.example.com/live2",
				URL:         "http://example.com",
			}},
			expectedURL: "http://stream.example.com/live2",
		},
		{
			name: "fallback to URL when URLResolved is empty",
			response: Station{
				UUID:        "test-uuid",
				URLResolved: "",
				URL:         "http://fallback.example.com",
			},
			expectedURL: "http://fallback.example.com",
		},
		{
			name:     "empty array",
			response: []Station{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := &Client{
				baseURL:   server.URL,
				userAgent: "TestApp/1.0",
				http:      &http.Client{Timeout: 5 * time.Second},
			}

			ctx := context.Background()
			url, err := client.ResolveStationURL(ctx, "test-uuid")

			if tt.wantErr {
				if err == nil {
					t.Error("ResolveStationURL() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveStationURL() error = %v", err)
			}
			if url != tt.expectedURL {
				t.Errorf("ResolveStationURL() = %q, want %q", url, tt.expectedURL)
			}
		})
	}
}

func TestClient_ResolveStationURL_EmptyUUID(t *testing.T) {
	client := &Client{
		baseURL:   "http://example.com",
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	_, err := client.ResolveStationURL(ctx, "")
	if err == nil {
		t.Error("ResolveStationURL() should return error for empty UUID")
	}
}

func TestClient_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	_, err := client.StationsByCountry(ctx, "US")
	if err == nil {
		t.Error("StationsByCountry() should return error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code, got: %v", err)
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	_, err := client.StationsByCountry(ctx, "US")
	if err == nil {
		t.Error("StationsByCountry() should return error for invalid JSON")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Station{})
	}))
	defer server.Close()

	client := &Client{
		baseURL:   server.URL,
		userAgent: "TestApp/1.0",
		http:      &http.Client{Timeout: 5 * time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.StationsByCountry(ctx, "US")
	if err == nil {
		t.Error("StationsByCountry() should return error when context is cancelled")
	}
}

func TestResolvedURL(t *testing.T) {
	tests := []struct {
		name        string
		station     Station
		expectedURL string
		wantErr     bool
	}{
		{
			name:        "prefers URLResolved",
			station:     Station{URLResolved: "http://resolved.com", URL: "http://original.com"},
			expectedURL: "http://resolved.com",
		},
		{
			name:        "falls back to URL",
			station:     Station{URLResolved: "", URL: "http://original.com"},
			expectedURL: "http://original.com",
		},
		{
			name:        "whitespace URLResolved falls back",
			station:     Station{URLResolved: "   ", URL: "http://original.com"},
			expectedURL: "http://original.com",
		},
		{
			name:    "no URLs",
			station: Station{URLResolved: "", URL: ""},
			wantErr: true,
		},
		{
			name:    "both whitespace",
			station: Station{URLResolved: "  ", URL: "  "},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := resolvedURL(tt.station)
			if tt.wantErr {
				if err == nil {
					t.Error("resolvedURL() should return error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolvedURL() error = %v", err)
			}
			if url != tt.expectedURL {
				t.Errorf("resolvedURL() = %q, want %q", url, tt.expectedURL)
			}
		})
	}
}
