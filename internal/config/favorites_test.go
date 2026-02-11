package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"radio-tui/internal/radio"
)

// newTestFavorites creates a Favorites instance with a temp file for testing.
func newTestFavorites(t *testing.T) *Favorites {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "favorites.json")
	return &Favorites{
		path:  path,
		items: make(map[string]Favorite),
	}
}

func TestFavorites_Toggle_Add(t *testing.T) {
	favs := newTestFavorites(t)

	station := radio.Station{
		UUID:    "test-uuid-1",
		Name:    "Test FM",
		Country: "US",
		Tags:    "rock,pop",
	}

	added, err := favs.Toggle(station)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}
	if !added {
		t.Error("Toggle() should return true when adding a favorite")
	}

	if !favs.IsFavorite(station.UUID) {
		t.Error("IsFavorite() should return true after adding")
	}
}

func TestFavorites_Toggle_Remove(t *testing.T) {
	favs := newTestFavorites(t)

	station := radio.Station{
		UUID:    "test-uuid-2",
		Name:    "Test FM 2",
		Country: "UK",
		Tags:    "news",
	}

	// First add
	_, err := favs.Toggle(station)
	if err != nil {
		t.Fatalf("Toggle() add error = %v", err)
	}

	// Then remove
	added, err := favs.Toggle(station)
	if err != nil {
		t.Fatalf("Toggle() remove error = %v", err)
	}
	if added {
		t.Error("Toggle() should return false when removing a favorite")
	}

	if favs.IsFavorite(station.UUID) {
		t.Error("IsFavorite() should return false after removing")
	}
}

func TestFavorites_Toggle_EmptyUUID(t *testing.T) {
	favs := newTestFavorites(t)

	station := radio.Station{
		UUID: "",
		Name: "No UUID Station",
	}

	_, err := favs.Toggle(station)
	if err == nil {
		t.Error("Toggle() should return error for empty UUID")
	}
}

func TestFavorites_IsFavorite_NotFound(t *testing.T) {
	favs := newTestFavorites(t)

	if favs.IsFavorite("nonexistent-uuid") {
		t.Error("IsFavorite() should return false for unknown UUID")
	}
}

func TestFavorites_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "favorites.json")

	// Create and add favorites
	favs := &Favorites{
		path:  path,
		items: make(map[string]Favorite),
	}

	stations := []radio.Station{
		{UUID: "uuid-1", Name: "Station 1", Country: "US", Tags: "rock"},
		{UUID: "uuid-2", Name: "Station 2", Country: "UK", Tags: "pop"},
		{UUID: "uuid-3", Name: "Station 3", Country: "JP", Tags: "jazz"},
	}

	for _, s := range stations {
		if _, err := favs.Toggle(s); err != nil {
			t.Fatalf("Toggle() error = %v", err)
		}
	}

	// Verify file was written
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var stored favoritesFile
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(stored.Stations) != 3 {
		t.Errorf("Persisted %d stations, want 3", len(stored.Stations))
	}

	// Verify station data
	uuids := make(map[string]bool)
	for _, s := range stored.Stations {
		uuids[s.UUID] = true
	}
	for _, s := range stations {
		if !uuids[s.UUID] {
			t.Errorf("Station %s not persisted", s.UUID)
		}
	}
}

func TestFavorites_ConcurrentAccess(t *testing.T) {
	favs := newTestFavorites(t)

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Concurrent toggles
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			station := radio.Station{
				UUID:    "concurrent-uuid",
				Name:    "Concurrent Station",
				Country: "US",
			}
			for j := 0; j < numOperations; j++ {
				_, _ = favs.Toggle(station)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = favs.IsFavorite("concurrent-uuid")
			}
		}()
	}

	wg.Wait()
	// Test passes if no race conditions or deadlocks occurred
}

func TestFavorites_MultipleFavorites(t *testing.T) {
	favs := newTestFavorites(t)

	stations := []radio.Station{
		{UUID: "a", Name: "Station A", Country: "US", Tags: "rock"},
		{UUID: "b", Name: "Station B", Country: "UK", Tags: "pop"},
		{UUID: "c", Name: "Station C", Country: "JP", Tags: "jazz"},
	}

	// Add all
	for _, s := range stations {
		added, err := favs.Toggle(s)
		if err != nil {
			t.Fatalf("Toggle() error = %v", err)
		}
		if !added {
			t.Errorf("Expected station %s to be added", s.UUID)
		}
	}

	// Verify all are favorites
	for _, s := range stations {
		if !favs.IsFavorite(s.UUID) {
			t.Errorf("Station %s should be a favorite", s.UUID)
		}
	}

	// Remove middle one
	removed, err := favs.Toggle(stations[1])
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}
	if removed {
		t.Error("Expected station to be removed (false)")
	}

	// Verify state
	if !favs.IsFavorite("a") {
		t.Error("Station A should still be a favorite")
	}
	if favs.IsFavorite("b") {
		t.Error("Station B should not be a favorite")
	}
	if !favs.IsFavorite("c") {
		t.Error("Station C should still be a favorite")
	}
}

func TestFavorite_StoresStationData(t *testing.T) {
	favs := newTestFavorites(t)

	station := radio.Station{
		UUID:    "data-test",
		Name:    "Data Test FM",
		Country: "Mongolia",
		Tags:    "news,talk,culture",
	}

	_, err := favs.Toggle(station)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	// Read the persisted file
	data, err := os.ReadFile(favs.path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var stored favoritesFile
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(stored.Stations) != 1 {
		t.Fatalf("Expected 1 station, got %d", len(stored.Stations))
	}

	fav := stored.Stations[0]
	if fav.UUID != station.UUID {
		t.Errorf("UUID = %q, want %q", fav.UUID, station.UUID)
	}
	if fav.Name != station.Name {
		t.Errorf("Name = %q, want %q", fav.Name, station.Name)
	}
	if fav.Country != station.Country {
		t.Errorf("Country = %q, want %q", fav.Country, station.Country)
	}
	if fav.Tags != station.Tags {
		t.Errorf("Tags = %q, want %q", fav.Tags, station.Tags)
	}
}
