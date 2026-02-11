package radio

import (
	"encoding/json"
	"testing"
)

func TestFrequency_Float64(t *testing.T) {
	tests := []struct {
		name     string
		freq     Frequency
		expected float64
	}{
		{"zero", Frequency(0), 0},
		{"positive", Frequency(98.5), 98.5},
		{"negative", Frequency(-1.5), -1.5},
		{"large", Frequency(108.0), 108.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.freq.Float64()
			if got != tt.expected {
				t.Errorf("Float64() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFrequency_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Frequency
	}{
		// Number inputs
		{"number int", `98`, Frequency(98)},
		{"number float", `98.5`, Frequency(98.5)},
		{"number zero", `0`, Frequency(0)},
		{"number negative", `-5.5`, Frequency(-5.5)},

		// String inputs
		{"string int", `"98"`, Frequency(98)},
		{"string float", `"98.5"`, Frequency(98.5)},
		{"string zero", `"0"`, Frequency(0)},
		{"string with spaces", `"  98.5  "`, Frequency(98.5)},
		{"string empty", `""`, Frequency(0)},
		{"string whitespace only", `"   "`, Frequency(0)},

		// Invalid string values (should default to 0)
		{"string invalid", `"not a number"`, Frequency(0)},
		{"string partial", `"98.5abc"`, Frequency(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Frequency
			err := json.Unmarshal([]byte(tt.json), &f)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if f != tt.expected {
				t.Errorf("UnmarshalJSON() = %v, want %v", f, tt.expected)
			}
		})
	}
}

func TestFrequency_UnmarshalJSON_Null(t *testing.T) {
	var f Frequency
	err := json.Unmarshal([]byte(`null`), &f)
	// null should not error, frequency stays at zero value
	if err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if f != 0 {
		t.Errorf("UnmarshalJSON(null) = %v, want 0", f)
	}
}

func TestStation_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		expectedFreq Frequency
		expectedName string
	}{
		{
			name: "frequency as number",
			json: `{
				"stationuuid": "abc123",
				"name": "Test FM",
				"country": "US",
				"countrycode": "US",
				"tags": "rock,pop",
				"bitrate": 128,
				"frequency": 98.5,
				"url_resolved": "http://example.com/stream",
				"url": "http://example.com",
				"favicon": "",
				"clickcount": 100,
				"is_broken": false
			}`,
			expectedFreq: Frequency(98.5),
			expectedName: "Test FM",
		},
		{
			name: "frequency as string",
			json: `{
				"stationuuid": "abc123",
				"name": "Radio Awesome",
				"country": "MN",
				"countrycode": "MN",
				"tags": "news",
				"bitrate": 64,
				"frequency": "101.5",
				"url_resolved": "http://example.com/stream2",
				"url": "http://example.com",
				"favicon": "",
				"clickcount": 50,
				"is_broken": false
			}`,
			expectedFreq: Frequency(101.5),
			expectedName: "Radio Awesome",
		},
		{
			name: "frequency empty string",
			json: `{
				"stationuuid": "xyz789",
				"name": "No Freq Radio",
				"country": "JP",
				"countrycode": "JP",
				"tags": "",
				"bitrate": 192,
				"frequency": "",
				"url_resolved": "http://example.com/stream3",
				"url": "http://example.com",
				"favicon": "",
				"clickcount": 200,
				"is_broken": false
			}`,
			expectedFreq: Frequency(0),
			expectedName: "No Freq Radio",
		},
		{
			name: "frequency missing",
			json: `{
				"stationuuid": "missing",
				"name": "Missing Freq",
				"country": "UK",
				"countrycode": "GB",
				"tags": "talk",
				"bitrate": 96,
				"url_resolved": "http://example.com/stream4",
				"url": "http://example.com",
				"favicon": "",
				"clickcount": 10,
				"is_broken": false
			}`,
			expectedFreq: Frequency(0),
			expectedName: "Missing Freq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Station
			err := json.Unmarshal([]byte(tt.json), &s)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if s.Frequency != tt.expectedFreq {
				t.Errorf("Station.Frequency = %v, want %v", s.Frequency, tt.expectedFreq)
			}
			if s.Name != tt.expectedName {
				t.Errorf("Station.Name = %v, want %v", s.Name, tt.expectedName)
			}
		})
	}
}
