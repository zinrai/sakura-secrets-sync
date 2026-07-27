package main

import "testing"

func TestPathToName(t *testing.T) {
	tests := []struct {
		dir  string
		path string
		want string
	}{
		{"staging", "staging/config/database.yaml", "config/database.yaml"},
		{"staging", "staging/ssh_keys/id_ed25519", "ssh_keys/id_ed25519"},
		{"staging", "staging/app.env", "app.env"},
		{"production", "production/a/b/c.json", "a/b/c.json"},
	}

	for _, tt := range tests {
		got := PathToName(tt.dir, tt.path)
		if got != tt.want {
			t.Errorf("PathToName(%q, %q) = %q, want %q", tt.dir, tt.path, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"value\n", "value"},
		{"value\n\n", "value"},
		{"value", "value"},
		{"line1\nline2\n", "line1\nline2"},
		{"line1\nline2", "line1\nline2"},
	}

	for _, tt := range tests {
		got := Normalize(tt.in)
		if got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
