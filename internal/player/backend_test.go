package player

import (
	"errors"
	"sync"
	"testing"
)

// mockBackend implements Backend for testing
type mockBackend struct {
	mu        sync.Mutex
	playing   bool
	lastURL   string
	playErr   error
	stopErr   error
	playCalls int
	stopCalls int
}

func (m *mockBackend) Play(url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playCalls++
	if m.playErr != nil {
		return m.playErr
	}
	m.playing = true
	m.lastURL = url
	return nil
}

func (m *mockBackend) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalls++
	if m.stopErr != nil {
		return m.stopErr
	}
	m.playing = false
	return nil
}

func (m *mockBackend) IsPlaying() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playing
}

func (m *mockBackend) LastURL() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastURL
}

func TestCompositeBackend_Stop_WhenNotPlaying(t *testing.T) {
	cb := &CompositeBackend{}

	err := cb.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v, want nil", err)
	}
}

func TestCompositeBackend_IsPlaying_WhenNotPlaying(t *testing.T) {
	cb := &CompositeBackend{}

	if cb.IsPlaying() {
		t.Error("IsPlaying() should return false when nothing is playing")
	}
}

func TestCompositeBackend_LastURL_Empty(t *testing.T) {
	cb := &CompositeBackend{}

	if cb.LastURL() != "" {
		t.Errorf("LastURL() = %q, want empty", cb.LastURL())
	}
}

func TestCompositeBackend_LastURL_AfterPlay(t *testing.T) {
	mock := &mockBackend{}
	cb := &CompositeBackend{
		gp: &GoPlayer{}, // Will fail, but that's ok for this test
	}
	// We can't easily test with real backends, but we can test lastURL is set
	url := "http://example.com/stream"

	// This will fail because GoPlayer isn't properly initialized,
	// but lastURL should still be set
	_ = cb.Play(url)

	if cb.LastURL() != url {
		t.Errorf("LastURL() = %q, want %q", cb.LastURL(), url)
	}

	_ = mock // silence unused warning
}

func TestCompositeBackend_Play_NoBackends(t *testing.T) {
	cb := &CompositeBackend{
		gp:  nil,
		ext: nil,
	}

	err := cb.Play("http://example.com/stream")
	if err == nil {
		t.Error("Play() should return error when no backends available")
	}
}

func TestCompositeBackend_ConcurrentAccess(t *testing.T) {
	cb := &CompositeBackend{}

	done := make(chan bool)

	// Concurrent Stop calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = cb.Stop()
			done <- true
		}()
	}

	// Concurrent IsPlaying calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = cb.IsPlaying()
			done <- true
		}()
	}

	// Concurrent LastURL calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = cb.LastURL()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 30; i++ {
		<-done
	}
}

func TestNew_ReturnsBackendOrError(t *testing.T) {
	// New() should return either a valid backend or an error, never both nil
	backend, err := New()

	if backend == nil && err == nil {
		t.Error("New() should return either a backend or an error")
	}

	// If we got a backend, it should implement the interface
	if backend != nil {
		// Test that interface methods don't panic
		_ = backend.IsPlaying()
		_ = backend.LastURL()
		_ = backend.Stop()
	}
}

func TestBackendInterface(t *testing.T) {
	// Verify that our types implement the Backend interface
	var _ Backend = (*Player)(nil)
	var _ Backend = (*CompositeBackend)(nil)
	var _ Backend = (*GoPlayer)(nil)
}

func TestMockBackend_Behavior(t *testing.T) {
	// Test our mock to ensure it works correctly for other tests
	mock := &mockBackend{}

	if mock.IsPlaying() {
		t.Error("mock should not be playing initially")
	}

	if err := mock.Play("http://test.com"); err != nil {
		t.Errorf("Play() error = %v", err)
	}

	if !mock.IsPlaying() {
		t.Error("mock should be playing after Play()")
	}

	if mock.LastURL() != "http://test.com" {
		t.Errorf("LastURL() = %q, want %q", mock.LastURL(), "http://test.com")
	}

	if err := mock.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if mock.IsPlaying() {
		t.Error("mock should not be playing after Stop()")
	}

	if mock.playCalls != 1 {
		t.Errorf("playCalls = %d, want 1", mock.playCalls)
	}

	if mock.stopCalls != 1 {
		t.Errorf("stopCalls = %d, want 1", mock.stopCalls)
	}
}

func TestMockBackend_WithErrors(t *testing.T) {
	mock := &mockBackend{
		playErr: errors.New("play error"),
		stopErr: errors.New("stop error"),
	}

	if err := mock.Play("http://test.com"); err == nil {
		t.Error("Play() should return error")
	}

	if err := mock.Stop(); err == nil {
		t.Error("Stop() should return error")
	}
}
