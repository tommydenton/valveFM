package ui

import "testing"

func TestParseIPCCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Valid commands
		{"simple command", "play", "PLAY", false},
		{"uppercase command", "STOP", "STOP", false},
		{"mixed case", "PlAy", "PLAY", false},
		{"with leading space", "  play", "PLAY", false},
		{"with trailing space", "stop  ", "STOP", false},
		{"with both spaces", "  toggle  ", "TOGGLE", false},

		// Invalid commands
		{"empty string", "", "", true},
		{"whitespace only", "   ", "", true},
		{"tabs only", "\t\t", "", true},
		{"newlines only", "\n\n", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIPCCommand(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("parseIPCCommand() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseIPCCommand() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("parseIPCCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseIPCCommand_CommonCommands(t *testing.T) {
	// Test common IPC commands that the app might receive
	commands := []string{
		"PLAY",
		"STOP",
		"TOGGLE",
		"NEXT",
		"PREV",
		"STATUS",
		"QUIT",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			result, err := parseIPCCommand(cmd)
			if err != nil {
				t.Fatalf("parseIPCCommand(%q) error = %v", cmd, err)
			}
			if result != cmd {
				t.Errorf("parseIPCCommand(%q) = %q, want %q", cmd, result, cmd)
			}
		})
	}
}

func TestIPCReply_Struct(t *testing.T) {
	// Test ipcReply struct construction
	tests := []struct {
		name   string
		reply  ipcReply
		isOK   bool
		hasErr bool
	}{
		{
			name:   "success reply",
			reply:  ipcReply{ok: true, data: "Station: Test FM"},
			isOK:   true,
			hasErr: false,
		},
		{
			name:   "error reply",
			reply:  ipcReply{ok: false, err: "not playing"},
			isOK:   false,
			hasErr: true,
		},
		{
			name:   "success with empty data",
			reply:  ipcReply{ok: true, data: ""},
			isOK:   true,
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reply.ok != tt.isOK {
				t.Errorf("reply.ok = %v, want %v", tt.reply.ok, tt.isOK)
			}
			hasErr := tt.reply.err != ""
			if hasErr != tt.hasErr {
				t.Errorf("reply has error = %v, want %v", hasErr, tt.hasErr)
			}
		})
	}
}

func TestIPCMsg_Struct(t *testing.T) {
	// Test ipcMsg struct construction
	replyChan := make(chan ipcReply, 1)
	msg := ipcMsg{
		cmd:   "PLAY",
		reply: replyChan,
	}

	if msg.cmd != "PLAY" {
		t.Errorf("msg.cmd = %q, want %q", msg.cmd, "PLAY")
	}

	// Test channel works
	go func() {
		msg.reply <- ipcReply{ok: true, data: "Playing"}
	}()

	reply := <-msg.reply
	if !reply.ok {
		t.Error("reply should be ok")
	}
	if reply.data != "Playing" {
		t.Errorf("reply.data = %q, want %q", reply.data, "Playing")
	}
}
