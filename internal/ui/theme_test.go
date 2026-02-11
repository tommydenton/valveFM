package ui

import "testing"

func TestThemeBySlug(t *testing.T) {
	tests := []struct {
		name         string
		slug         string
		expectedName string
	}{
		// Valid slugs
		{"vintage", "vintage", "Vintage"},
		{"tokyo-night", "tokyo-night", "Tokyo Night"},
		{"nord", "nord", "Nord"},
		{"catppuccin-mocha", "catppuccin-mocha", "Catppuccin Mocha"},
		{"catppuccin-latte", "catppuccin-latte", "Catppuccin Latte"},
		{"gruvbox-dark", "gruvbox-dark", "Gruvbox Dark"},
		{"dracula", "dracula", "Dracula"},
		{"solarized-dark", "solarized-dark", "Solarized Dark"},
		{"one-dark", "one-dark", "One Dark"},
		{"rose-pine", "rose-pine", "Rose Pine"},
		{"kanagawa", "kanagawa", "Kanagawa"},
		{"everforest", "everforest", "Everforest"},

		// Invalid slugs (should fallback to Vintage)
		{"empty slug", "", "Vintage"},
		{"unknown slug", "nonexistent-theme", "Vintage"},
		{"typo in slug", "vintge", "Vintage"},
		{"case mismatch", "VINTAGE", "Vintage"},
		{"partial match", "dark", "Vintage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := ThemeBySlug(tt.slug)
			if theme.Name != tt.expectedName {
				t.Errorf("ThemeBySlug(%q).Name = %q, want %q", tt.slug, theme.Name, tt.expectedName)
			}
		})
	}
}

func TestThemesNotEmpty(t *testing.T) {
	if len(Themes) == 0 {
		t.Error("Themes slice should not be empty")
	}

	// Verify first theme is Vintage (used as fallback)
	if Themes[0].Slug != "vintage" {
		t.Errorf("First theme should be 'vintage', got %q", Themes[0].Slug)
	}
}

func TestThemesHaveRequiredFields(t *testing.T) {
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			if theme.Name == "" {
				t.Error("Theme.Name should not be empty")
			}
			if theme.Slug == "" {
				t.Error("Theme.Slug should not be empty")
			}
			if theme.Fg == "" {
				t.Error("Theme.Fg (foreground) should not be empty")
			}
			if theme.Accent == "" {
				t.Error("Theme.Accent should not be empty")
			}
			if theme.Secondary == "" {
				t.Error("Theme.Secondary should not be empty")
			}
			if theme.Bg == "" {
				t.Error("Theme.Bg (background) should not be empty")
			}
			if theme.Success == "" {
				t.Error("Theme.Success should not be empty")
			}
			if theme.Muted == "" {
				t.Error("Theme.Muted should not be empty")
			}
			if theme.Error == "" {
				t.Error("Theme.Error should not be empty")
			}
		})
	}
}

func TestThemeSlugsAreUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, theme := range Themes {
		if seen[theme.Slug] {
			t.Errorf("Duplicate theme slug: %q", theme.Slug)
		}
		seen[theme.Slug] = true
	}
}

func TestBuildStyles(t *testing.T) {
	// Test that BuildStyles produces valid Styles for each theme
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			styles := BuildStyles(theme)

			// Verify that styles are not zero values by checking some properties
			// We can't easily check lipgloss.Style internals, but we can verify
			// the function doesn't panic and returns something

			// Just verify we got a Styles struct back (no panic)
			_ = styles.App
			_ = styles.Header
			_ = styles.Panel
			_ = styles.InsetPanel
			_ = styles.DialLabel
			_ = styles.DialScale
			_ = styles.DialPointer
			_ = styles.StationName
			_ = styles.Meta
			_ = styles.ListHeader
			_ = styles.ListItem
			_ = styles.ListActive
			_ = styles.KeyHint
			_ = styles.HelpBox
			_ = styles.Error
			_ = styles.Accent
			_ = styles.Muted
		})
	}
}

func TestThemeColorFormats(t *testing.T) {
	// All colors should be valid hex colors starting with #
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			colors := map[string]string{
				"Fg":        theme.Fg,
				"Accent":    theme.Accent,
				"Secondary": theme.Secondary,
				"Bg":        theme.Bg,
				"Success":   theme.Success,
				"Muted":     theme.Muted,
				"Error":     theme.Error,
			}

			for name, color := range colors {
				if len(color) != 7 {
					t.Errorf("Theme.%s = %q, want 7-character hex color", name, color)
				}
				if color[0] != '#' {
					t.Errorf("Theme.%s = %q, should start with #", name, color)
				}
			}
		})
	}
}
