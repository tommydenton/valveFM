package ui

import "testing"

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		maxLen   int
		expected string
	}{
		// Basic cases
		{"short text fits", "Hello", 10, "Hello"},
		{"exact length", "Hello", 5, "Hello"},
		{"needs truncation", "Hello World", 8, "Hello..."},

		// Edge cases
		{"empty string", "", 10, ""},
		{"max zero", "Hello", 0, ""},
		{"max negative", "Hello", -5, ""},
		{"max 1", "Hello", 1, "H"},
		{"max 2", "Hello", 2, "He"},
		{"max 3", "Hello", 3, "Hel"},
		{"max 4 truncates", "Hello World", 4, "H..."},

		// Whitespace handling
		{"leading space trimmed", "  Hello", 10, "Hello"},
		{"trailing space trimmed", "Hello  ", 10, "Hello"},
		{"both spaces trimmed", "  Hello  ", 10, "Hello"},

		// Unicode (note: truncateText uses byte length, not rune count)
		// Multi-byte characters that fit within the limit work correctly
		{"unicode fits", "日本語", 10, "日本語"},
		// Note: truncateText may produce invalid UTF-8 when truncating multi-byte chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateText(tt.value, tt.maxLen)
			if got != tt.expected {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.value, tt.maxLen, got, tt.expected)
			}
		})
	}
}

func TestFallback(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		alt      string
		expected string
	}{
		{"non-empty value", "hello", "default", "hello"},
		{"empty value", "", "default", "default"},
		{"whitespace only", "   ", "default", "default"},
		{"tabs only", "\t\t", "default", "default"},
		{"newlines only", "\n\n", "default", "default"},
		{"mixed whitespace", " \t\n ", "default", "default"},
		{"value with spaces", "  hello  ", "default", "  hello  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fallback(tt.value, tt.alt)
			if got != tt.expected {
				t.Errorf("fallback(%q, %q) = %q, want %q", tt.value, tt.alt, got, tt.expected)
			}
		})
	}
}

func TestListWindow(t *testing.T) {
	tests := []struct {
		name          string
		length        int
		selected      int
		max           int
		expectedStart int
		expectedEnd   int
	}{
		// Small list (fits entirely)
		{"small list", 5, 2, 10, 0, 5},
		{"small list at start", 3, 0, 10, 0, 3},

		// Large list
		{"large list start", 20, 0, 5, 0, 5},
		{"large list middle", 20, 10, 5, 8, 13},
		{"large list end", 20, 19, 5, 15, 20},
		{"large list near end", 20, 18, 5, 15, 20},

		// Edge cases
		{"single item", 1, 0, 5, 0, 1},
		{"max equals length", 10, 5, 10, 0, 10},
		{"selected at boundary", 10, 2, 5, 0, 5},

		// Window centering
		{"center window", 100, 50, 10, 45, 55},
		{"window shifted at start", 100, 3, 10, 0, 10},
		{"window shifted at end", 100, 97, 10, 90, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := listWindow(tt.length, tt.selected, tt.max)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("listWindow(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.length, tt.selected, tt.max, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}

func TestJoinHeader(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		width    int
		expected string
	}{
		{"basic join", "LEFT", "RIGHT", 20, "LEFT           RIGHT"},
		{"zero width", "LEFT", "RIGHT", 0, ""},
		{"right exceeds width", "L", "VERYLONGRIGHT", 5, "VE..."},
		{"exact fit", "AB", "CD", 5, "AB CD"},
		{"left truncated", "VERYLONGLEFT", "R", 10, "VERYL... R"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinHeader(tt.left, tt.right, tt.width)
			if got != tt.expected {
				t.Errorf("joinHeader(%q, %q, %d) = %q, want %q",
					tt.left, tt.right, tt.width, got, tt.expected)
			}
		})
	}
}

func TestInnerWidthForPanel(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected int
	}{
		{"standard width", 80, 74},
		{"narrow width", 20, 14},
		{"very narrow", 8, 6},
		{"tiny", 4, 2},
		{"minimum", 2, 2},
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := innerWidthForPanel(tt.width)
			if got != tt.expected {
				t.Errorf("innerWidthForPanel(%d) = %d, want %d", tt.width, got, tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a greater", 10, 5, 10},
		{"b greater", 5, 10, 10},
		{"equal", 7, 7, 7},
		{"negative a", -5, 3, 3},
		{"negative b", 3, -5, 3},
		{"both negative", -5, -3, -3},
		{"zero and positive", 0, 5, 5},
		{"zero and negative", 0, -5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := max(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestFormatFreqLabel(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"whole number", 98.0, "98"},
		{"half", 98.5, "98.5"},
		{"very close to whole", 98.01, "98"},
		{"close to whole", 97.99, "98"},
		{"third", 98.3, "98.3"},
		{"two thirds", 98.7, "98.7"},
		{"large", 108.0, "108"},
		{"small", 88.0, "88"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFreqLabel(tt.value)
			if got != tt.expected {
				t.Errorf("formatFreqLabel(%v) = %q, want %q", tt.value, got, tt.expected)
			}
		})
	}
}

func TestPlaceLabel(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		pos      int
		label    string
		expected string
	}{
		{"center", 10, 5, "AB", "    AB    "},
		{"at start", 10, 1, "AB", "AB        "},
		{"at end", 10, 9, "AB", "        AB"},
		{"too long for start", 10, 0, "ABCD", "ABCD      "},
		{"overflow", 5, 2, "ABCDEFGH", "     "}, // label doesn't fit, unchanged
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := make([]rune, tt.width)
			for i := range line {
				line[i] = ' '
			}
			placeLabel(line, tt.pos, tt.label)
			got := string(line)
			if got != tt.expected {
				t.Errorf("placeLabel(width=%d, pos=%d, %q) = %q, want %q",
					tt.width, tt.pos, tt.label, got, tt.expected)
			}
		})
	}
}

func TestBuildDialScale(t *testing.T) {
	// Test that the function produces valid output
	tests := []struct {
		name    string
		width   int
		min     float64
		max     float64
		dynamic bool
	}{
		{"fixed scale narrow", 20, 88, 108, false},
		{"fixed scale wide", 80, 88, 108, false},
		{"dynamic scale", 60, 90, 100, true},
		{"dynamic narrow", 20, 95, 105, true},
		{"minimum width", 10, 88, 108, false},
		{"below minimum", 5, 88, 108, false}, // should be clamped to 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels, bar, minor := buildDialScale(tt.width, tt.min, tt.max, tt.dynamic)

			expectedWidth := tt.width
			if expectedWidth < 10 {
				expectedWidth = 10
			}

			// Verify bar has correct length
			if len(bar) != expectedWidth {
				t.Errorf("bar length = %d, want %d", len(bar), expectedWidth)
			}

			// Verify labels has correct length
			if len(labels) != expectedWidth {
				t.Errorf("labels length = %d, want %d", len(labels), expectedWidth)
			}

			// Verify minor line (can be empty for narrow widths)
			if minor != "" && len(minor) != expectedWidth {
				t.Errorf("minor length = %d, want %d or empty", len(minor), expectedWidth)
			}
		})
	}
}

func TestBuildDialScaleFixed(t *testing.T) {
	labels, bar, _ := buildDialScaleFixed(60)

	// Check that bar contains pipe characters at label positions
	pipeCount := 0
	for _, r := range bar {
		if r == '|' {
			pipeCount++
		}
	}

	// Fixed scale has 11 labels (88, 90, 92, ..., 108)
	if pipeCount != 11 {
		t.Errorf("bar should have 11 pipe characters, got %d", pipeCount)
	}

	// Labels should contain frequency values
	if labels == "" {
		t.Error("labels should not be empty")
	}
}
