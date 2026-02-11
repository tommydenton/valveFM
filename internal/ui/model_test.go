package ui

import (
	"testing"

	"radio-tui/internal/radio"
)

// createTestModel creates a Model with test data for testing
func createTestModel() *Model {
	m := &Model{
		stations: []radio.Station{
			{UUID: "1", Name: "Rock FM", Tags: "rock,metal", Frequency: 98.5},
			{UUID: "2", Name: "Pop Radio", Tags: "pop,hits", Frequency: 101.3},
			{UUID: "3", Name: "Jazz Station", Tags: "jazz,smooth", Frequency: 104.7},
			{UUID: "4", Name: "News Talk", Tags: "news,talk", Frequency: 88.1},
			{UUID: "5", Name: "Classical Music", Tags: "classical", Frequency: 106.9},
		},
		countries: []radio.Country{
			{Code: "US", Name: "United States"},
			{Code: "GB", Name: "United Kingdom"},
			{Code: "JP", Name: "Japan"},
			{Code: "DE", Name: "Germany"},
			{Code: "FR", Name: "France"},
		},
		country:  "US",
		selected: 0,
	}
	return m
}

func TestModel_VisibleStations_NoFilter(t *testing.T) {
	m := createTestModel()

	stations := m.visibleStations()
	if len(stations) != 5 {
		t.Errorf("visibleStations() = %d stations, want 5", len(stations))
	}
}

func TestModel_VisibleStations_WithFilter(t *testing.T) {
	m := createTestModel()
	m.filtered = []radio.Station{m.stations[0], m.stations[2]} // Rock FM, Jazz Station
	m.search.SetValue("test")                                  // Non-empty to trigger filter

	stations := m.visibleStations()
	if len(stations) != 2 {
		t.Errorf("visibleStations() with filter = %d stations, want 2", len(stations))
	}
}

func TestModel_CurrentStation(t *testing.T) {
	m := createTestModel()
	m.selected = 2

	station, ok := m.currentStation()
	if !ok {
		t.Fatal("currentStation() should return ok=true")
	}
	if station.Name != "Jazz Station" {
		t.Errorf("currentStation().Name = %q, want %q", station.Name, "Jazz Station")
	}
}

func TestModel_CurrentStation_EmptyList(t *testing.T) {
	m := &Model{
		stations: []radio.Station{},
		selected: 0,
	}

	_, ok := m.currentStation()
	if ok {
		t.Error("currentStation() should return ok=false for empty list")
	}
}

func TestModel_CurrentStation_OutOfBounds(t *testing.T) {
	m := createTestModel()
	m.selected = 100 // Out of bounds

	_, ok := m.currentStation()
	if ok {
		t.Error("currentStation() should return ok=false for out of bounds selection")
	}
}

func TestModel_MoveSelection(t *testing.T) {
	m := createTestModel()
	m.selected = 2

	// Move down
	changed := m.moveSelection(1)
	if !changed {
		t.Error("moveSelection(1) should return true when position changes")
	}
	if m.selected != 3 {
		t.Errorf("selected = %d, want 3", m.selected)
	}

	// Move up
	changed = m.moveSelection(-1)
	if !changed {
		t.Error("moveSelection(-1) should return true when position changes")
	}
	if m.selected != 2 {
		t.Errorf("selected = %d, want 2", m.selected)
	}
}

func TestModel_MoveSelection_AtBoundary(t *testing.T) {
	m := createTestModel()

	// At start, try to move up
	m.selected = 0
	changed := m.moveSelection(-1)
	if changed {
		t.Error("moveSelection(-1) should return false at start")
	}
	if m.selected != 0 {
		t.Errorf("selected = %d, want 0", m.selected)
	}

	// At end, try to move down
	m.selected = len(m.stations) - 1
	changed = m.moveSelection(1)
	if changed {
		t.Error("moveSelection(1) should return false at end")
	}
	if m.selected != len(m.stations)-1 {
		t.Errorf("selected = %d, want %d", m.selected, len(m.stations)-1)
	}
}

func TestModel_MoveSelection_EmptyList(t *testing.T) {
	m := &Model{
		stations: []radio.Station{},
		selected: 0,
	}

	changed := m.moveSelection(1)
	if changed {
		t.Error("moveSelection() should return false for empty list")
	}
}

func TestModel_EnsureSelection(t *testing.T) {
	m := createTestModel()

	// Negative selection
	m.selected = -5
	m.ensureSelection()
	if m.selected != 0 {
		t.Errorf("ensureSelection() with negative = %d, want 0", m.selected)
	}

	// Over bounds
	m.selected = 100
	m.ensureSelection()
	if m.selected != 4 { // len-1
		t.Errorf("ensureSelection() with over bounds = %d, want 4", m.selected)
	}

	// Valid selection (should not change)
	m.selected = 2
	m.ensureSelection()
	if m.selected != 2 {
		t.Errorf("ensureSelection() with valid = %d, want 2", m.selected)
	}
}

func TestModel_EnsureSelection_EmptyList(t *testing.T) {
	m := &Model{
		stations: []radio.Station{},
		selected: 5,
	}

	m.ensureSelection()
	if m.selected != 0 {
		t.Errorf("ensureSelection() with empty list = %d, want 0", m.selected)
	}
}

func TestModel_CurrentCountry(t *testing.T) {
	m := createTestModel()
	m.countryIndex = 1

	country, ok := m.currentCountry()
	if !ok {
		t.Fatal("currentCountry() should return ok=true")
	}
	if country.Code != "GB" {
		t.Errorf("currentCountry().Code = %q, want %q", country.Code, "GB")
	}
}

func TestModel_CurrentCountry_EmptyList(t *testing.T) {
	m := &Model{
		countries:    []radio.Country{},
		countryIndex: 0,
	}

	_, ok := m.currentCountry()
	if ok {
		t.Error("currentCountry() should return ok=false for empty list")
	}
}

func TestModel_MoveCountrySelection(t *testing.T) {
	m := createTestModel()
	m.countryIndex = 2

	m.moveCountrySelection(1)
	if m.countryIndex != 3 {
		t.Errorf("countryIndex = %d, want 3", m.countryIndex)
	}

	m.moveCountrySelection(-1)
	if m.countryIndex != 2 {
		t.Errorf("countryIndex = %d, want 2", m.countryIndex)
	}
}

func TestModel_MoveCountrySelection_AtBoundary(t *testing.T) {
	m := createTestModel()

	// At start
	m.countryIndex = 0
	m.moveCountrySelection(-1)
	if m.countryIndex != 0 {
		t.Errorf("countryIndex = %d, want 0", m.countryIndex)
	}

	// At end
	m.countryIndex = len(m.countries) - 1
	m.moveCountrySelection(1)
	if m.countryIndex != len(m.countries)-1 {
		t.Errorf("countryIndex = %d, want %d", m.countryIndex, len(m.countries)-1)
	}
}

func TestModel_EnsureCountrySelection(t *testing.T) {
	m := createTestModel()

	// Negative
	m.countryIndex = -5
	m.ensureCountrySelection()
	if m.countryIndex != 0 {
		t.Errorf("ensureCountrySelection() = %d, want 0", m.countryIndex)
	}

	// Over bounds
	m.countryIndex = 100
	m.ensureCountrySelection()
	if m.countryIndex != 4 {
		t.Errorf("ensureCountrySelection() = %d, want 4", m.countryIndex)
	}
}

func TestModel_VisibleCountries_NoFilter(t *testing.T) {
	m := createTestModel()

	countries := m.visibleCountries()
	if len(countries) != 5 {
		t.Errorf("visibleCountries() = %d, want 5", len(countries))
	}
}

func TestModel_VisibleCountries_WithFilter(t *testing.T) {
	m := createTestModel()
	m.filteredCountries = []radio.Country{m.countries[0], m.countries[1]}
	m.countrySearch.SetValue("test") // Non-empty to trigger filter

	countries := m.visibleCountries()
	if len(countries) != 2 {
		t.Errorf("visibleCountries() with filter = %d, want 2", len(countries))
	}
}

func TestModel_UpdateDialRange(t *testing.T) {
	m := createTestModel()

	m.updateDialRange()

	if !m.dialUseFreq {
		t.Error("dialUseFreq should be true when stations have frequencies")
	}
	if m.dialMin != 88.1 {
		t.Errorf("dialMin = %v, want 88.1", m.dialMin)
	}
	if m.dialMax != 106.9 {
		t.Errorf("dialMax = %v, want 106.9", m.dialMax)
	}
}

func TestModel_UpdateDialRange_NoFrequencies(t *testing.T) {
	m := &Model{
		stations: []radio.Station{
			{UUID: "1", Name: "Station 1", Frequency: 0},
			{UUID: "2", Name: "Station 2", Frequency: 0},
		},
	}

	m.updateDialRange()

	if m.dialUseFreq {
		t.Error("dialUseFreq should be false when no frequencies")
	}
}

func TestModel_UpdateDialRange_EmptyList(t *testing.T) {
	m := &Model{
		stations: []radio.Station{},
	}

	m.updateDialRange()

	if m.dialUseFreq {
		t.Error("dialUseFreq should be false for empty list")
	}
	if m.dialMin != 0 || m.dialMax != 0 {
		t.Errorf("dialMin/dialMax = %v/%v, want 0/0", m.dialMin, m.dialMax)
	}
}

func TestModel_DialValueForIndex(t *testing.T) {
	m := createTestModel()
	m.updateDialRange()

	// Index 0 should return frequency of first station (sorted by position, not freq)
	val := m.dialValueForIndex(0)
	if val != 98.5 { // Rock FM
		t.Errorf("dialValueForIndex(0) = %v, want 98.5", val)
	}

	// Index 3 should return frequency of fourth station
	val = m.dialValueForIndex(3)
	if val != 88.1 { // News Talk
		t.Errorf("dialValueForIndex(3) = %v, want 88.1", val)
	}
}

func TestModel_DialValueForIndex_OutOfBounds(t *testing.T) {
	m := createTestModel()

	val := m.dialValueForIndex(-1)
	if val != 0 {
		t.Errorf("dialValueForIndex(-1) = %v, want 0", val)
	}

	val = m.dialValueForIndex(100)
	if val != 0 {
		t.Errorf("dialValueForIndex(100) = %v, want 0", val)
	}
}

func TestModel_SnapDial(t *testing.T) {
	m := createTestModel()
	m.updateDialRange()
	m.selected = 2

	m.snapDial()

	// dialPos and dialTarget should match the frequency of selected station
	expectedFreq := m.stations[2].Frequency.Float64() // Jazz Station = 104.7
	if m.dialPos != expectedFreq {
		t.Errorf("dialPos = %v, want %v", m.dialPos, expectedFreq)
	}
	if m.dialTarget != expectedFreq {
		t.Errorf("dialTarget = %v, want %v", m.dialTarget, expectedFreq)
	}
}

func TestModel_UpdateInputWidths(t *testing.T) {
	m := createTestModel()

	// Wide screen
	m.width = 100
	m.updateInputWidths()
	if m.search.Width != 32 { // Max width
		t.Errorf("search.Width = %d, want 32 (max)", m.search.Width)
	}

	// Narrow screen
	m.width = 30
	m.updateInputWidths()
	if m.search.Width != 10 { // width - 20 = 10
		t.Errorf("search.Width = %d, want 10", m.search.Width)
	}

	// Very narrow
	m.width = 15
	m.updateInputWidths()
	if m.search.Width != 10 { // Minimum
		t.Errorf("search.Width = %d, want 10 (min)", m.search.Width)
	}
}

func TestModel_IPCStatus(t *testing.T) {
	m := createTestModel()
	m.country = "US"
	m.playing = false

	status := m.ipcStatus()

	if status == "" {
		t.Error("ipcStatus() should not be empty")
	}

	// Should contain JSON-like structure
	expectedParts := []string{`"playing":false`, `"country":"US"`}
	for _, part := range expectedParts {
		if !contains(status, part) {
			t.Errorf("ipcStatus() missing %q, got %q", part, status)
		}
	}
}

func TestModel_IPCStatus_Playing(t *testing.T) {
	m := createTestModel()
	m.playing = true
	m.country = "JP"

	status := m.ipcStatus()

	if !contains(status, `"playing":true`) {
		t.Errorf("ipcStatus() should show playing:true, got %q", status)
	}
	if !contains(status, `"country":"JP"`) {
		t.Errorf("ipcStatus() should show country:JP, got %q", status)
	}
}

func TestSendIPCReply_NilChannel(t *testing.T) {
	// Should not panic with nil channel
	sendIPCReply(nil, ipcReply{ok: true})
}

func TestSendIPCReply_BufferedChannel(t *testing.T) {
	ch := make(chan ipcReply, 1)
	sendIPCReply(ch, ipcReply{ok: true, data: "test"})

	reply := <-ch
	if !reply.ok {
		t.Error("reply.ok should be true")
	}
	if reply.data != "test" {
		t.Errorf("reply.data = %q, want %q", reply.data, "test")
	}
}

func TestInputMode_Constants(t *testing.T) {
	// Verify input mode constants are distinct
	modes := []inputMode{inputNone, inputLocation, inputSearch, inputCountrySelect}
	seen := make(map[inputMode]bool)

	for _, mode := range modes {
		if seen[mode] {
			t.Errorf("duplicate input mode: %v", mode)
		}
		seen[mode] = true
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
