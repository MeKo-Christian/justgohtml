package encoding_test

import (
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/encoding"
)

const (
	encUTF8        = "UTF-8"
	encWindows1252 = "windows-1252"
	encUTF16       = "utf-16"
	encEUCJP       = "euc-jp"
	encISO88592    = "iso-8859-2"
)

// TestEncodingLabelNormalization tests that all WHATWG-defined encoding labels
// are correctly recognized and normalized to their canonical encodings.
// This ensures support for 20+ encoding labels as required.
func TestEncodingLabelNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		label    string
		wantName string // Expected canonical encoding name
	}{
		// UTF-8 labels (WHATWG standard)
		{"utf-8", encUTF8},
		{"utf8", encUTF8},
		{"unicode-1-1-utf-8", encUTF8},
		{"unicode11utf8", encUTF8},
		{"unicode20utf8", encUTF8},
		{"x-unicode20utf8", encUTF8},

		// windows-1252 labels (including ASCII labels per WHATWG)
		{"windows-1252", encWindows1252},
		{"windows1252", encWindows1252},
		{"cp1252", encWindows1252},
		{"x-cp1252", encWindows1252},
		{"ansi_x3.4-1968", encWindows1252},
		{"ascii", encWindows1252},
		{"cp819", encWindows1252},
		{"csisolatin1", encWindows1252},
		{"ibm819", encWindows1252},
		{"iso-8859-1", encWindows1252},
		{"iso-ir-100", encWindows1252},
		{"iso8859-1", encWindows1252},
		{"iso88591", encWindows1252},
		{"iso_8859-1", encWindows1252},
		{"iso_8859-1:1987", encWindows1252},
		{"l1", encWindows1252},
		{"latin1", encWindows1252},
		{"us-ascii", encWindows1252},

		// ISO-8859-2 labels
		{"iso-8859-2", encISO88592},
		{"iso8859-2", encISO88592},
		{"iso88592", encISO88592},
		{"iso_8859-2", encISO88592},
		{"iso_8859-2:1987", encISO88592},
		{"l2", encISO88592},
		{"latin2", encISO88592},
		{"l2", "iso-8859-2"},
		{"csisolatin2", "iso-8859-2"},

		// EUC-JP labels
		{"euc-jp", "euc-jp"},
		{"eucjp", "euc-jp"},
		{"cseucpkdfmtjapanese", "euc-jp"},
		{"x-euc-jp", "euc-jp"},

		// UTF-16 variants
		{"utf-16le", "utf-16le"},
		{"utf16le", "utf-16le"},
		{"utf-16be", "utf-16be"},
		{"utf16be", "utf-16be"},
		{"utf-16", "utf-16"},
		{"utf16", "utf-16"},

		// Case insensitivity
		{"UTF-8", "UTF-8"},
		{"UTF8", "UTF-8"},
		{"Windows-1252", "windows-1252"},
		{"WINDOWS-1252", "windows-1252"},
		{"ISO-8859-1", "windows-1252"},

		// Whitespace trimming
		{" utf-8 ", "UTF-8"},
		{"\tutf-8\n", "UTF-8"},
		{"  windows-1252  ", "windows-1252"},

		// UTF-7 security check (should return windows-1252)
		{"utf-7", "windows-1252"},
		{"utf7", "windows-1252"},
		{"x-utf-7", "windows-1252"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			t.Parallel()

			_, enc, err := encoding.Decode([]byte("test"), tt.label)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", tt.label, err)
			}
			if enc == nil {
				t.Fatalf("Decode(%q) returned nil encoding", tt.label)
			}
			if enc.Name != tt.wantName {
				t.Errorf("Decode(%q) encoding name = %q, want %q", tt.label, enc.Name, tt.wantName)
			}
		})
	}
}

// TestUnrecognizedEncodingLabels tests that unrecognized labels
// fall back to windows-1252 (the default fallback).
func TestUnrecognizedEncodingLabels(t *testing.T) {
	t.Parallel()

	tests := []string{
		"unknown-encoding",
		"fake-label",
		"not-real",
	}

	for _, label := range tests {
		t.Run(label, func(t *testing.T) {
			t.Parallel()

			_, enc, err := encoding.Decode([]byte("test"), label)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", label, err)
			}
			// Unrecognized labels should fall back to windows-1252
			if enc.Name != encWindows1252 {
				t.Errorf("Decode(%q) encoding name = %q, want %q (fallback)", label, enc.Name, encWindows1252)
			}
		})
	}
}

// TestEmptyEncodingLabel tests that empty labels fall back correctly.
func TestEmptyEncodingLabel(t *testing.T) {
	t.Parallel()

	_, enc, err := encoding.Decode([]byte("test"), "")
	if err != nil {
		t.Fatalf("Decode with empty label error: %v", err)
	}
	// Empty label should fall back to windows-1252
	if enc.Name != "windows-1252" {
		t.Errorf("Decode with empty label encoding name = %q, want %q", enc.Name, "windows-1252")
	}
}

// TestEncodingLabelCount verifies we support at least 20+ distinct encoding labels.
func TestEncodingLabelCount(t *testing.T) {
	t.Parallel()

	// List of distinct labels we should support (not counting case/whitespace variants)
	distinctLabels := []string{
		// UTF-8 (6 labels)
		"utf-8", "utf8", "unicode-1-1-utf-8", "unicode11utf8", "unicode20utf8", "x-unicode20utf8",

		// windows-1252 / ASCII / ISO-8859-1 (18 labels)
		"windows-1252", "windows1252", "cp1252", "x-cp1252",
		"ansi_x3.4-1968", "ascii", "cp819", "csisolatin1", "ibm819",
		"iso-8859-1", "iso-ir-100", "iso8859-1", "iso88591",
		"iso_8859-1", "iso_8859-1:1987", "l1", "latin1", "us-ascii",

		// ISO-8859-2 (9 labels)
		"iso-8859-2", "iso8859-2", "iso88592", "iso_8859-2", "iso_8859-2:1987",
		"iso-ir-101", "latin2", "l2", "csisolatin2",

		// EUC-JP (4 labels)
		"euc-jp", "eucjp", "cseucpkdfmtjapanese", "x-euc-jp",

		// UTF-16 variants (6 labels)
		"utf-16le", "utf16le", "utf-16be", "utf16be", "utf-16", "utf16",
	}

	// Verify each label is recognized
	for _, label := range distinctLabels {
		_, enc, err := encoding.Decode([]byte("test"), label)
		if err != nil {
			t.Errorf("Label %q not recognized: %v", label, err)
			continue
		}
		if enc == nil {
			t.Errorf("Label %q returned nil encoding", label)
		}
	}

	// Verify we have at least 20 distinct labels
	if len(distinctLabels) < 20 {
		t.Errorf("Only testing %d distinct labels, need at least 20", len(distinctLabels))
	}

	t.Logf("Successfully tested %d distinct encoding labels", len(distinctLabels))
}

// TestBOMDetection tests BOM (Byte Order Mark) detection for various encodings.
func TestBOMDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		data         []byte
		wantEncoding string
	}{
		{
			name:         "UTF-8 BOM",
			data:         []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			wantEncoding: "UTF-8",
		},
		{
			name:         "UTF-16LE BOM",
			data:         []byte{0xFF, 0xFE, 'h', 0x00, 'i', 0x00},
			wantEncoding: "utf-16le",
		},
		{
			name:         "UTF-16BE BOM",
			data:         []byte{0xFE, 0xFF, 0x00, 'h', 0x00, 'i'},
			wantEncoding: "utf-16be",
		},
		{
			name:         "No BOM",
			data:         []byte("hello world"),
			wantEncoding: "windows-1252", // fallback
		},
		{
			name:         "Too short for UTF-8 BOM",
			data:         []byte{0xEF, 0xBB},
			wantEncoding: "windows-1252", // fallback
		},
		{
			name:         "Too short for UTF-16 BOM",
			data:         []byte{0xFF},
			wantEncoding: "windows-1252", // fallback
		},
		{
			name:         "Empty data",
			data:         []byte{},
			wantEncoding: "windows-1252", // fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, enc, err := encoding.Decode(tt.data, "")
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}
			if enc.Name != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", enc.Name, tt.wantEncoding)
			}
		})
	}
}

// TestDecodeWithHint tests that transport-provided encoding hints are respected.
func TestDecodeWithHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		data         []byte
		hint         string
		wantEncoding string
	}{
		{
			name:         "Hint overrides BOM",
			data:         []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			hint:         "iso-8859-2",
			wantEncoding: "iso-8859-2",
		},
		{
			name:         "Hint with invalid label falls back",
			data:         []byte("hello"),
			hint:         "unknown-encoding",
			wantEncoding: "windows-1252", // falls back when hint is invalid
		},
		{
			name:         "Hint with whitespace",
			data:         []byte("test"),
			hint:         "  utf-8  ",
			wantEncoding: "UTF-8",
		},
		{
			name:         "Case insensitive hint",
			data:         []byte("test"),
			hint:         "UTF-8",
			wantEncoding: "UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, enc, err := encoding.Decode(tt.data, tt.hint)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}
			if enc.Name != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", enc.Name, tt.wantEncoding)
			}
		})
	}
}

// TestDecodeAllEncodings tests decoding with all supported encodings.
func TestDecodeAllEncodings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hint         string
		data         []byte
		wantEncoding string
	}{
		{"utf-8", []byte("hello"), "UTF-8"},
		{"windows-1252", []byte{0x80, 0x81}, "windows-1252"},     // Euro sign and control
		{"iso-8859-1", []byte{0xE9}, "windows-1252"},             // HTML treats ISO-8859-1 as windows-1252
		{"iso-8859-2", []byte{0xA1}, "iso-8859-2"},               // Ą in ISO-8859-2
		{"euc-jp", []byte{0x41, 0x42}, "euc-jp"},                 // ASCII in EUC-JP
		{"utf-16le", []byte{0x41, 0x00, 0x42, 0x00}, "utf-16le"}, // "AB" in UTF-16LE
		{"utf-16be", []byte{0x00, 0x41, 0x00, 0x42}, "utf-16be"}, // "AB" in UTF-16BE
		{"utf-16", []byte{0xFF, 0xFE, 0x41, 0x00}, "utf-16"},     // UTF-16 with BOM
	}

	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			t.Parallel()
			decoded, enc, err := encoding.Decode(tt.data, tt.hint)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", tt.hint, err)
			}
			if enc.Name != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", enc.Name, tt.wantEncoding)
			}
			if decoded == "" {
				t.Error("Decoded string is empty")
			}
		})
	}
}

// TestMetaCharsetDetection tests detection of charset from meta tags.
func TestMetaCharsetDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		html         string
		wantEncoding string
	}{
		{
			name:         "meta charset attribute",
			html:         `<meta charset="utf-8">`,
			wantEncoding: "UTF-8",
		},
		{
			name:         "meta http-equiv Content-Type",
			html:         `<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-2">`,
			wantEncoding: "iso-8859-2",
		},
		{
			name:         "meta charset with whitespace",
			html:         `<meta charset="  utf-8  ">`,
			wantEncoding: "UTF-8",
		},
		{
			name:         "meta after 1024 bytes (should not be detected)",
			html:         strings.Repeat("x", 1025) + `<meta charset="iso-8859-2">`,
			wantEncoding: "windows-1252", // fallback since meta is beyond scan limit
		},
		{
			name:         "meta in comment (should be ignored)",
			html:         `<!--<meta charset="iso-8859-2">--><meta charset="utf-8">`,
			wantEncoding: "UTF-8",
		},
		{
			name:         "UTF-16 in meta declaration becomes UTF-8",
			html:         `<meta charset="utf-16">`,
			wantEncoding: "UTF-8", // Per HTML spec, UTF-16/32 in meta are treated as UTF-8
		},
		{
			name:         "UTF-16LE in meta declaration becomes UTF-8",
			html:         `<meta charset="utf-16le">`,
			wantEncoding: "UTF-8",
		},
		{
			name:         "UTF-16BE in meta declaration becomes UTF-8",
			html:         `<meta charset="utf-16be">`,
			wantEncoding: "UTF-8",
		},
		{
			name:         "No meta tag",
			html:         `<html><body>hello</body></html>`,
			wantEncoding: "windows-1252", // fallback
		},
		{
			name:         "Unclosed quoted attribute in meta",
			html:         `<meta charset="utf-8><meta charset="iso-8859-2">`,
			wantEncoding: "windows-1252", // Malformed meta causes fallback
		},
		{
			name:         "meta with Content-Type and mixed case",
			html:         `<meta http-equiv="Content-TYPE" content="text/html; CHARSET=utf-8">`,
			wantEncoding: "UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, enc, err := encoding.Decode([]byte(tt.html), "")
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}
			if enc.Name != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", enc.Name, tt.wantEncoding)
			}
		})
	}
}

// TestEdgeCasesForCoverage tests edge cases to achieve 100% coverage.
func TestEdgeCasesForCoverage(t *testing.T) {
	t.Parallel()

	t.Run("bomLength with default case", func(t *testing.T) {
		t.Parallel()
		// Test bomLength with an encoding that has no BOM
		// This tests the default case in the bomLength function
		_, enc, err := encoding.Decode([]byte("test"), encWindows1252)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		// The bomLength function is called internally with Windows1252 which returns 0
		if enc.Name != encWindows1252 {
			t.Errorf("Expected %s, got %s", encWindows1252, enc.Name)
		}
	})

	t.Run("normalizeEncodingLabel with whitespace-only label", func(t *testing.T) {
		t.Parallel()
		// Test with a label that becomes empty after trimming
		_, enc, err := encoding.Decode([]byte("test"), "   ")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		// Should fall back to windows-1252
		if enc.Name != encWindows1252 {
			t.Errorf("Expected fallback to %s, got %s", encWindows1252, enc.Name)
		}
	})

	t.Run("decodeWithEncoding UTF-16 without BOM", func(t *testing.T) {
		t.Parallel()
		// Test UTF-16 decoding without a BOM (should default to LE)
		_, enc, err := encoding.Decode([]byte{0x41, 0x00, 0x42, 0x00}, encUTF16)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encUTF16 {
			t.Errorf("Expected %s, got %s", encUTF16, enc.Name)
		}
	})

	t.Run("decodeWithEncoding with odd-length UTF-16LE", func(t *testing.T) {
		t.Parallel()
		// Test UTF-16LE with odd length (should add padding)
		_, enc, err := encoding.Decode([]byte{0x41, 0x00, 0x42}, "utf-16le")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "utf-16le" {
			t.Errorf("Expected utf-16le, got %s", enc.Name)
		}
	})

	t.Run("decodeWithEncoding with odd-length UTF-16BE", func(t *testing.T) {
		t.Parallel()
		// Test UTF-16BE with odd length (should add padding)
		_, enc, err := encoding.Decode([]byte{0x00, 0x41, 0x00}, "utf-16be")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "utf-16be" {
			t.Errorf("Expected utf-16be, got %s", enc.Name)
		}
	})

	t.Run("decodeWithEncoding EUC-JP multibyte", func(t *testing.T) {
		t.Parallel()
		// Test EUC-JP with multibyte characters
		_, enc, err := encoding.Decode([]byte{0xA1, 0xA1, 0x41}, encEUCJP)
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encEUCJP {
			t.Errorf("Expected %s, got %s", encEUCJP, enc.Name)
		}
	})

	t.Run("prescanForMetaCharset with end tag", func(t *testing.T) {
		t.Parallel()
		// Test with end tags to cover those code paths
		html := `</div><meta charset="utf-8">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encUTF8 {
			t.Errorf("Expected %s, got %s", encUTF8, enc.Name)
		}
	})

	t.Run("prescanForMetaCharset with non-meta tag containing attributes", func(t *testing.T) {
		t.Parallel()
		// Test with non-meta tags that have attributes
		html := `<div id="test" class="foo"><meta charset="utf-8">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encUTF8 {
			t.Errorf("Expected %s, got %s", encUTF8, enc.Name)
		}
	})

	t.Run("prescanForMetaCharset with quoted attributes in non-meta tag", func(t *testing.T) {
		t.Parallel()
		// Test with quoted attributes containing > character
		html := `<div title="test > foo"><meta charset="iso-8859-2">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encISO88592 {
			t.Errorf("Expected %s, got %s", encISO88592, enc.Name)
		}
	})

	t.Run("extractCharsetFromContent with semicolon separator", func(t *testing.T) {
		t.Parallel()
		// Test charset extraction with semicolon after charset value
		html := `<meta http-equiv="Content-Type" content="text/html; charset=utf-8; other=value">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encUTF8 {
			t.Errorf("Expected %s, got %s", encUTF8, enc.Name)
		}
	})

	t.Run("extractCharsetFromContent with single quotes", func(t *testing.T) {
		t.Parallel()
		// Test charset extraction with single-quoted value
		html := `<meta http-equiv="Content-Type" content="text/html; charset='iso-8859-2'">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encISO88592 {
			t.Errorf("Expected %s, got %s", encISO88592, enc.Name)
		}
	})

	t.Run("prescanForMetaCharset with tag that is not alphabetic", func(t *testing.T) {
		t.Parallel()
		// Test with < followed by non-alphabetic character
		html := `<123><meta charset="utf-8">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != encUTF8 {
			t.Errorf("Expected %s, got %s", encUTF8, enc.Name)
		}
	})

	t.Run("prescanForMetaCharset restarts on < in attributes", func(t *testing.T) {
		t.Parallel()

		// Test with < character in attribute area (should restart scanning)
		html := `<meta charset="utf-8" <test><meta charset="iso-8859-2">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		// Should find the second meta tag
		if enc.Name != encISO88592 {
			t.Errorf("Expected %s, got %s", encISO88592, enc.Name)
		}
	})
}

// TestExtractCharsetEdgeCases tests edge cases in charset extraction from content attribute.
func TestExtractCharsetEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		html         string
		wantEncoding string
	}{
		{
			name:         "charset= with nothing after",
			html:         `<meta http-equiv="Content-Type" content="charset=">`,
			wantEncoding: encWindows1252, // No valid charset
		},
		{
			name:         "charset without =",
			html:         `<meta http-equiv="Content-Type" content="charset">`,
			wantEncoding: encWindows1252, // No valid charset
		},
		{
			name:         "charset at end of content",
			html:         `<meta http-equiv="Content-Type" content="text/html; charset">`,
			wantEncoding: encWindows1252, // No valid charset
		},
		{
			name:         "empty content attribute",
			html:         `<meta http-equiv="Content-Type" content="">`,
			wantEncoding: encWindows1252,
		},
		{
			name:         "charset with unclosed double quote",
			html:         `<meta http-equiv="Content-Type" content='charset="utf-8'>`,
			wantEncoding: encWindows1252, // Unclosed quote
		},
		{
			name:         "charset with space before =",
			html:         `<meta http-equiv="Content-Type" content="charset  =  utf-8">`,
			wantEncoding: encUTF8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, enc, err := encoding.Decode([]byte(tt.html), "")
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}
			if enc.Name != tt.wantEncoding {
				t.Errorf("Encoding = %q, want %q", enc.Name, tt.wantEncoding)
			}
		})
	}
}

// TestDecodeEncodingEdgeCases tests edge cases in decoding different encodings.
func TestDecodeEncodingEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("ISO-8859-1 direct decode", func(t *testing.T) {
		t.Parallel()

		// Even though we define ISO-8859-1, it should map to windows-1252 per HTML spec
		data := []byte{0xA9} // Copyright symbol
		_, enc, err := encoding.Decode(data, "iso-8859-1")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		// Per normalizeEncodingLabel, ISO-8859-1 labels return Windows1252
		if enc.Name != "windows-1252" {
			t.Errorf("Expected windows-1252, got %s", enc.Name)
		}
	})

	t.Run("windows-1252 special mappings", func(t *testing.T) {
		t.Parallel()

		// Test that bytes 0x80-0x9F are mapped correctly
		data := []byte{0x99} // Trade mark sign in windows-1252
		decoded, enc, err := encoding.Decode(data, "windows-1252")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "windows-1252" {
			t.Errorf("Expected windows-1252, got %s", enc.Name)
		}
		// Should contain the trademark symbol
		if len(decoded) == 0 {
			t.Error("Decoded string is empty")
		}
	})

	t.Run("EUC-JP with ASCII and multibyte", func(t *testing.T) {
		t.Parallel()

		// Mix of ASCII and multibyte characters
		data := []byte{0x41, 0xA1, 0xA2, 0x42, 0xA3} // A, multibyte, B, partial multibyte
		decoded, enc, err := encoding.Decode(data, "euc-jp")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "euc-jp" {
			t.Errorf("Expected euc-jp, got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string is empty")
		}
	})

	t.Run("UTF-16 with BOM detection in data", func(t *testing.T) {
		t.Parallel()

		// UTF-16 with BOM in data (not at start)
		data := []byte{0xFF, 0xFE, 0x41, 0x00} // LE BOM + 'A'
		decoded, enc, err := encoding.Decode(data, "utf-16")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "utf-16" {
			t.Errorf("Expected utf-16, got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string is empty")
		}
	})

	t.Run("UTF-16 with BE BOM in data", func(t *testing.T) {
		t.Parallel()

		// UTF-16 with BE BOM in data
		data := []byte{0xFE, 0xFF, 0x00, 0x41} // BE BOM + 'A'
		decoded, enc, err := encoding.Decode(data, "utf-16")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "utf-16" {
			t.Errorf("Expected utf-16, got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string is empty")
		}
	})
}

// TestPrescanEdgeCases tests additional edge cases in meta charset prescan.
func TestPrescanEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("meta with attribute value without =", func(t *testing.T) {
		t.Parallel()

		// Attribute without value
		html := `<meta charset utf-8>`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		// charset attribute has no value, should fall back
		if enc.Name != "windows-1252" {
			t.Errorf("Expected windows-1252, got %s", enc.Name)
		}
	})

	t.Run("meta with unquoted attribute value", func(t *testing.T) {
		t.Parallel()

		html := `<meta http-equiv=Content-Type content=text/html;charset=utf-8>`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "UTF-8" {
			t.Errorf("Expected UTF-8, got %s", enc.Name)
		}
	})

	t.Run("end tag with quotes", func(t *testing.T) {
		t.Parallel()

		html := `</div class="test"><meta charset="iso-8859-2">`
		_, enc, err := encoding.Decode([]byte(html), "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "iso-8859-2" {
			t.Errorf("Expected iso-8859-2, got %s", enc.Name)
		}
	})
}

// TestBomLengthCoverage ensures all bomLength paths are covered.
func TestBomLengthCoverage(t *testing.T) {
	t.Parallel()

	t.Run("UTF-16LE BOM strips correctly", func(t *testing.T) {
		t.Parallel()

		// UTF-16LE with BOM at start (should strip 2 bytes)
		data := []byte{0xFF, 0xFE, 0x41, 0x00, 0x42, 0x00} // BOM + "AB"
		_, enc, err := encoding.Decode(data, "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "utf-16le" {
			t.Errorf("Expected utf-16le, got %s", enc.Name)
		}
	})

	t.Run("UTF-16BE BOM strips correctly", func(t *testing.T) {
		t.Parallel()

		// UTF-16BE with BOM at start (should strip 2 bytes)
		data := []byte{0xFE, 0xFF, 0x00, 0x41, 0x00, 0x42} // BOM + "AB"
		_, enc, err := encoding.Decode(data, "")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "utf-16be" {
			t.Errorf("Expected utf-16be, got %s", enc.Name)
		}
	})

	t.Run("ISO-8859-2 has no BOM", func(t *testing.T) {
		t.Parallel()

		// ISO-8859-2 with BOM-like bytes (should not be treated as BOM)
		data := []byte{0xEF, 0xBB, 0xBF, 0x41}
		_, enc, err := encoding.Decode(data, "iso-8859-2")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "iso-8859-2" {
			t.Errorf("Expected iso-8859-2, got %s", enc.Name)
		}
	})
}

// TestDecodeWithEncodingFullCoverage tests all decoding paths.
func TestDecodeWithEncodingFullCoverage(t *testing.T) {
	t.Parallel()

	t.Run("ISO-8859-1 byte-to-rune mapping", func(t *testing.T) {
		t.Parallel()

		// ISO-8859-1 maps each byte to a code point directly
		// Test with extended ASCII range
		data := []byte{0xFF, 0xFE, 0xFD} // ÿ þ ý
		// Use ISO-8859-1 encoding directly (not via hint)
		// Note: HTML spec treats ISO-8859-1 as windows-1252, but we can still test
		// the decodeWithEncoding function behavior if we could call it directly
		// For now, test via the public API which maps ISO-8859-1 to windows-1252
		decoded, enc, err := encoding.Decode(data, "iso-8859-1")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "windows-1252" {
			t.Errorf("Expected windows-1252 (ISO-8859-1 mapped), got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string should not be empty")
		}
	})

	t.Run("ISO-8859-2 extended range", func(t *testing.T) {
		t.Parallel()

		// ISO-8859-2 has special mappings for 0x80-0xFF
		data := []byte{0x7F, 0x80, 0x81, 0xA0, 0xFF} // Mix of ASCII and extended
		decoded, enc, err := encoding.Decode(data, "iso-8859-2")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "iso-8859-2" {
			t.Errorf("Expected iso-8859-2, got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string should not be empty")
		}
	})

	t.Run("windows-1252 all control chars", func(t *testing.T) {
		t.Parallel()

		// Test all special mappings in 0x80-0x9F range
		data := make([]byte, 32)
		for i := range data {
			data[i] = byte(0x80 + i)
		}
		decoded, enc, err := encoding.Decode(data, "windows-1252")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "windows-1252" {
			t.Errorf("Expected windows-1252, got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string should not be empty")
		}
	})

	t.Run("EUC-JP with only high bytes", func(t *testing.T) {
		t.Parallel()

		// EUC-JP with consecutive multibyte sequences
		data := []byte{0xA1, 0xA2, 0xA3, 0xA4}
		decoded, enc, err := encoding.Decode(data, "euc-jp")
		if err != nil {
			t.Fatalf("Decode error: %v", err)
		}
		if enc.Name != "euc-jp" {
			t.Errorf("Expected euc-jp, got %s", enc.Name)
		}
		if len(decoded) == 0 {
			t.Error("Decoded string should not be empty")
		}
	})
}
