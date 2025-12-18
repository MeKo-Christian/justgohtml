package stream

import (
	"testing"
)

func TestNewConfigDefaults(t *testing.T) {
	cfg := newConfig()

	if cfg.encoding != "" {
		t.Errorf("default encoding = %q, want empty string", cfg.encoding)
	}
}

func TestWithEncoding(t *testing.T) {
	cfg := newConfig(WithEncoding("utf-8"))

	if cfg.encoding != "utf-8" {
		t.Errorf("encoding = %q, want %q", cfg.encoding, "utf-8")
	}
}

func TestMultipleOptions(t *testing.T) {
	// First option
	cfg := newConfig(WithEncoding("iso-8859-1"))
	if cfg.encoding != "iso-8859-1" {
		t.Errorf("encoding = %q, want %q", cfg.encoding, "iso-8859-1")
	}

	// Override with second call
	cfg = newConfig(WithEncoding("iso-8859-1"), WithEncoding("utf-8"))
	if cfg.encoding != "utf-8" {
		t.Errorf("encoding after override = %q, want %q", cfg.encoding, "utf-8")
	}
}

func TestNoOptions(t *testing.T) {
	cfg := newConfig()

	// Should work without any options
	if cfg == nil {
		t.Error("newConfig() returned nil")
	}
}
