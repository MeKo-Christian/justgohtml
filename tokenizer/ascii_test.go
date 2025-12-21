package tokenizer

import (
	"testing"
)

// Test-only accessor for isASCIIOnly field
func (t *Tokenizer) IsASCIIOnly() bool {
	return t.isASCIIOnly
}

// Test-only method to force rune mode (for comparison benchmarks)
func (t *Tokenizer) ForceRuneMode() {
	t.isASCIIOnly = false
}

func TestASCIIDetection(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantASCII bool
	}{
		{"empty", "", true},
		{"simple ASCII", "<div>hello</div>", true},
		{"with entities", "&amp;&lt;&gt;", true},
		{"numbers and symbols", "<!DOCTYPE html>", true},
		{"complex ASCII", `<!DOCTYPE html><html><body><p class="test">Hello World!</p></body></html>`, true},
		{"unicode emoji", "<div>😀</div>", false},
		//nolint:gosmopolitan
		{"unicode Japanese", "<div>こんにちは</div>", false},
		//nolint:gosmopolitan
		{"unicode Chinese", "<div>你好世界</div>", false},
		//nolint:gosmopolitan
		{"mixed ASCII and unicode", "<div>hello 世界</div>", false},
		{"high ASCII", "<div>\x80</div>", false},
		{"latin extended", "<div>café</div>", false},
		{"BOM", "\xFE\xFF<div>test</div>", false}, // BOM is non-ASCII
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := New(tt.input)
			if got := tok.IsASCIIOnly(); got != tt.wantASCII {
				t.Errorf("IsASCIIOnly() = %v, want %v", got, tt.wantASCII)
			}
		})
	}
}

func TestASCIIDetectionWithOptions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		discardBOM bool
		wantASCII  bool
	}{
		{"ASCII with BOM discarded", "\xFE\xFF<div>test</div>", true, false}, // BOM still in original input
		{"ASCII without BOM", "<div>test</div>", false, true},
		{"ASCII without BOM, discard enabled", "<div>test</div>", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{DiscardBOM: tt.discardBOM}
			tok := NewWithOptions(tt.input, opts)
			if got := tok.IsASCIIOnly(); got != tt.wantASCII {
				t.Errorf("IsASCIIOnly() = %v, want %v", got, tt.wantASCII)
			}
		})
	}
}
