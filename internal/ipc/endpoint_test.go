//go:build !windows

package ipc

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndpoint_Struct(t *testing.T) {
	ep := Endpoint{
		Network: "unix",
		Address: "/tmp/test.sock",
	}

	if ep.Network != "unix" {
		t.Errorf("Network = %q, want %q", ep.Network, "unix")
	}
	if ep.Address != "/tmp/test.sock" {
		t.Errorf("Address = %q, want %q", ep.Address, "/tmp/test.sock")
	}
}

func TestResolveEndpoint(t *testing.T) {
	ep, err := ResolveEndpoint()
	if err != nil {
		t.Fatalf("ResolveEndpoint() error = %v", err)
	}

	if ep.Network != "unix" {
		t.Errorf("Network = %q, want %q", ep.Network, "unix")
	}

	if !strings.Contains(ep.Address, "valvefm") {
		t.Errorf("Address should contain 'valvefm', got %q", ep.Address)
	}

	if !strings.HasSuffix(ep.Address, "ctl.sock") {
		t.Errorf("Address should end with 'ctl.sock', got %q", ep.Address)
	}
}

func TestListen(t *testing.T) {
	listener, ep, err := Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()
	defer Cleanup(ep)

	// Verify listener is working
	if listener == nil {
		t.Fatal("Listen() returned nil listener")
	}

	// Verify endpoint
	if ep.Network != "unix" {
		t.Errorf("Network = %q, want %q", ep.Network, "unix")
	}

	// Verify socket file exists
	if _, err := os.Stat(ep.Address); os.IsNotExist(err) {
		t.Errorf("Socket file should exist at %q", ep.Address)
	}

	// Test we can connect to the listener
	conn, err := net.Dial(ep.Network, ep.Address)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	conn.Close()
}

func TestListen_CreatesDirectory(t *testing.T) {
	listener, ep, err := Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()
	defer Cleanup(ep)

	// Verify parent directory exists
	dir := filepath.Dir(ep.Address)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", dir, err)
	}
	if !info.IsDir() {
		t.Errorf("%q should be a directory", dir)
	}
}

func TestListen_RemovesExistingSocket(t *testing.T) {
	// First listen creates socket
	listener1, ep1, err := Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	listener1.Close()

	// Second listen should work (removes old socket)
	listener2, ep2, err := Listen()
	if err != nil {
		t.Fatalf("Listen() second call error = %v", err)
	}
	defer listener2.Close()
	defer Cleanup(ep2)

	// Endpoints should be the same
	if ep1.Address != ep2.Address {
		t.Errorf("Addresses should match: %q vs %q", ep1.Address, ep2.Address)
	}
}

func TestCleanup(t *testing.T) {
	listener, ep, err := Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	// Don't close the listener yet - socket exists while listener is open
	// Verify socket exists while listener is active
	if _, err := os.Stat(ep.Address); os.IsNotExist(err) {
		t.Fatal("Socket should exist while listener is open")
	}

	// Close listener first
	listener.Close()

	// Cleanup should work regardless of socket state
	err = Cleanup(ep)
	if err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Verify socket is removed (or was already removed by Close)
	if _, err := os.Stat(ep.Address); err == nil {
		t.Error("Socket should be removed after cleanup")
	}
}

func TestCleanup_EmptyAddress(t *testing.T) {
	ep := Endpoint{
		Network: "unix",
		Address: "",
	}

	// Should not error for empty address
	err := Cleanup(ep)
	if err != nil {
		t.Errorf("Cleanup() error = %v, want nil", err)
	}
}

func TestCleanup_NonExistentFile(t *testing.T) {
	ep := Endpoint{
		Network: "unix",
		Address: "/tmp/nonexistent_socket_12345.sock",
	}

	// Should not error for non-existent file
	err := Cleanup(ep)
	if err != nil {
		t.Errorf("Cleanup() error = %v, want nil", err)
	}
}

func TestSocketPermissions(t *testing.T) {
	listener, ep, err := Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()
	defer Cleanup(ep)

	info, err := os.Stat(ep.Address)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Socket should have restricted permissions (0600)
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("Socket permissions = %o, want %o", perm, 0o600)
	}
}
