package constants

import "testing"

func TestIsWhitespace(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		{'\t', true},  // U+0009 TAB
		{'\n', true},  // U+000A LF
		{'\f', true},  // U+000C FF
		{' ', true},   // U+0020 SPACE
		{'\r', false}, // U+000D CR (not HTML5 whitespace)
		{'a', false},
		{'Z', false},
		{'0', false},
		{0x00, false},
		{0xFF, false},
		{0x100, false}, // Beyond ASCII
	}

	for _, tt := range tests {
		got := IsWhitespace(tt.char)
		if got != tt.want {
			t.Errorf("IsWhitespace(%q) = %v, want %v", tt.char, got, tt.want)
		}
	}
}

func TestIsASCIIUpper(t *testing.T) {
	// Test all uppercase letters
	for c := 'A'; c <= 'Z'; c++ {
		if !IsASCIIUpper(c) {
			t.Errorf("IsASCIIUpper(%q) = false, want true", c)
		}
	}

	// Test non-uppercase characters
	tests := []rune{'a', 'z', '0', '9', ' ', '\t', '@', '[', '`', '{', 0x100}
	for _, c := range tests {
		if IsASCIIUpper(c) {
			t.Errorf("IsASCIIUpper(%q) = true, want false", c)
		}
	}
}

func TestIsASCIILower(t *testing.T) {
	// Test all lowercase letters
	for c := 'a'; c <= 'z'; c++ {
		if !IsASCIILower(c) {
			t.Errorf("IsASCIILower(%q) = false, want true", c)
		}
	}

	// Test non-lowercase characters
	tests := []rune{'A', 'Z', '0', '9', ' ', '\t', '@', '[', '`', '{', 0x100}
	for _, c := range tests {
		if IsASCIILower(c) {
			t.Errorf("IsASCIILower(%q) = true, want false", c)
		}
	}
}

func TestIsASCIIAlpha(t *testing.T) {
	// Test all letters
	for c := 'A'; c <= 'Z'; c++ {
		if !IsASCIIAlpha(c) {
			t.Errorf("IsASCIIAlpha(%q) = false, want true", c)
		}
	}
	for c := 'a'; c <= 'z'; c++ {
		if !IsASCIIAlpha(c) {
			t.Errorf("IsASCIIAlpha(%q) = false, want true", c)
		}
	}

	// Test non-alpha characters
	tests := []rune{'0', '9', ' ', '\t', '@', '[', '`', '{', 0x100}
	for _, c := range tests {
		if IsASCIIAlpha(c) {
			t.Errorf("IsASCIIAlpha(%q) = true, want false", c)
		}
	}
}

func TestIsASCIIAlphaNum(t *testing.T) {
	// Test all alphanumeric
	for c := 'A'; c <= 'Z'; c++ {
		if !IsASCIIAlphaNum(c) {
			t.Errorf("IsASCIIAlphaNum(%q) = false, want true", c)
		}
	}
	for c := 'a'; c <= 'z'; c++ {
		if !IsASCIIAlphaNum(c) {
			t.Errorf("IsASCIIAlphaNum(%q) = false, want true", c)
		}
	}
	for c := '0'; c <= '9'; c++ {
		if !IsASCIIAlphaNum(c) {
			t.Errorf("IsASCIIAlphaNum(%q) = false, want true", c)
		}
	}

	// Test non-alphanumeric characters
	tests := []rune{' ', '\t', '@', '[', '`', '{', '/', ':', 0x100}
	for _, c := range tests {
		if IsASCIIAlphaNum(c) {
			t.Errorf("IsASCIIAlphaNum(%q) = true, want false", c)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input rune
		want  rune
	}{
		{'A', 'a'},
		{'Z', 'z'},
		{'M', 'm'},
		{'a', 'a'}, // Already lowercase
		{'z', 'z'},
		{'0', '0'}, // Not a letter
		{' ', ' '},
		{0x100, 0x100}, // Non-ASCII
	}

	for _, tt := range tests {
		got := ToLower(tt.input)
		if got != tt.want {
			t.Errorf("ToLower(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// Benchmarks to ensure lookup tables are actually faster

func BenchmarkIsWhitespace(b *testing.B) {
	chars := []rune{' ', '\t', '\n', '\f', 'a', 'Z', '0'}
	for i := 0; i < b.N; i++ {
		for _, c := range chars {
			_ = IsWhitespace(c)
		}
	}
}

func BenchmarkIsWhitespaceSwitch(b *testing.B) {
	chars := []rune{' ', '\t', '\n', '\f', 'a', 'Z', '0'}
	for i := 0; i < b.N; i++ {
		for _, c := range chars {
			switch c {
			case '\t', '\n', '\f', ' ':
				_ = true
			default:
				_ = false
			}
		}
	}
}

func BenchmarkIsASCIIUpper(b *testing.B) {
	chars := []rune{'A', 'M', 'Z', 'a', 'z', '0', ' '}
	for i := 0; i < b.N; i++ {
		for _, c := range chars {
			_ = IsASCIIUpper(c)
		}
	}
}

func BenchmarkIsASCIIUpperRange(b *testing.B) {
	chars := []rune{'A', 'M', 'Z', 'a', 'z', '0', ' '}
	for i := 0; i < b.N; i++ {
		for _, c := range chars {
			_ = c >= 'A' && c <= 'Z'
		}
	}
}
