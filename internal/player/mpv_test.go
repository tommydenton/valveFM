package player

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file (non-executable)
	regularFile := filepath.Join(tmpDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create an executable file
	execFile := filepath.Join(tmpDir, "exec.sh")
	if err := os.WriteFile(execFile, []byte("#!/bin/sh\necho hello"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create a .exe file (Windows style)
	exeFile := filepath.Join(tmpDir, "program.exe")
	if err := os.WriteFile(exeFile, []byte("fake exe"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create a directory
	dirPath := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(dirPath, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"non-existent file", filepath.Join(tmpDir, "nonexistent"), false},
		{"regular file", regularFile, false},
		{"executable file", execFile, true},
		{"exe file", exeFile, true}, // .exe files are always considered executable
		{"directory", dirPath, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExecutable(tt.path)
			if got != tt.expected {
				t.Errorf("isExecutable(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestPlayer_Play_EmptyURL(t *testing.T) {
	p := &Player{
		backend: "mpv",
		path:    "/usr/bin/mpv",
	}

	err := p.Play("")
	if err == nil {
		t.Error("Play() should return error for empty URL")
	}
}

func TestPlayer_Play_UnknownBackend(t *testing.T) {
	p := &Player{
		backend: "unknown",
		path:    "/some/path",
	}

	err := p.Play("http://example.com/stream")
	if err == nil {
		t.Error("Play() should return error for unknown backend")
	}
}

func TestPlayer_Stop_WhenNotPlaying(t *testing.T) {
	p := &Player{
		backend: "mpv",
		path:    "/usr/bin/mpv",
	}

	// Stop should not error when nothing is playing
	err := p.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v, want nil", err)
	}
}

func TestPlayer_IsPlaying_WhenNotPlaying(t *testing.T) {
	p := &Player{
		backend: "mpv",
		path:    "/usr/bin/mpv",
	}

	if p.IsPlaying() {
		t.Error("IsPlaying() should return false when nothing is playing")
	}
}

func TestPlayer_LastURL_Empty(t *testing.T) {
	p := &Player{
		backend: "mpv",
		path:    "/usr/bin/mpv",
	}

	if p.LastURL() != "" {
		t.Errorf("LastURL() = %q, want empty", p.LastURL())
	}
}

func TestPlayer_LastURL_AfterPlayAttempt(t *testing.T) {
	p := &Player{
		backend: "mpv",
		path:    "/nonexistent/mpv", // Use non-existent path to avoid actual playback
	}

	url := "http://example.com/stream"
	_ = p.Play(url) // Will fail but should still set lastURL

	if p.LastURL() != url {
		t.Errorf("LastURL() = %q, want %q", p.LastURL(), url)
	}
}

func TestPlayer_ConcurrentAccess(t *testing.T) {
	p := &Player{
		backend: "mpv",
		path:    "/usr/bin/mpv",
	}

	done := make(chan bool)

	// Concurrent Stop calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = p.Stop()
			done <- true
		}()
	}

	// Concurrent IsPlaying calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = p.IsPlaying()
			done <- true
		}()
	}

	// Concurrent LastURL calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = p.LastURL()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 30; i++ {
		<-done
	}
}

func TestNewExternal_NoPlayerAvailable(t *testing.T) {
	// This test documents the behavior when no player is found
	// In most test environments, mpv/ffplay may not be installed
	// The function should either return a player or an error, never panic

	player, err := newExternal()
	// Either player is found or error is returned
	if player == nil && err == nil {
		t.Error("newExternal() should return either a player or an error")
	}
}
