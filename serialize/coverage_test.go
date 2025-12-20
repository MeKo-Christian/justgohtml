package serialize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

const testEncodingUTF8 = "UTF-8"

// TestSerializeTokens tests the wrapper function (currently 0% coverage)
func TestSerializeTokens(t *testing.T) {
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", []any{}}),
		rawToken(t, []any{"Characters", "Hello"}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokens(tokens)
	if err != nil {
		t.Fatalf("SerializeTokens error: %v", err)
	}
	if out != "<div>Hello</div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

// TestSerializeDoctypePublicID tests DOCTYPE with PUBLIC ID (currently 33.3% coverage)
func TestSerializeDoctypePublicID(t *testing.T) {
	dt := dom.NewDocumentType("html", "-//W3C//DTD HTML 4.01//EN", "http://www.w3.org/TR/html4/strict.dtd")

	var sb strings.Builder
	serializeDoctype(&sb, dt)

	expected := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">`
	if sb.String() != expected {
		t.Fatalf("unexpected doctype: %q, want %q", sb.String(), expected)
	}
}

// TestSerializeDoctypePublicIDOnly tests DOCTYPE with PUBLIC ID but no SYSTEM ID
func TestSerializeDoctypePublicIDOnly(t *testing.T) {
	dt := dom.NewDocumentType("html", "-//W3C//DTD HTML 4.01//EN", "")

	var sb strings.Builder
	serializeDoctype(&sb, dt)

	expected := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN">`
	if sb.String() != expected {
		t.Fatalf("unexpected doctype: %q, want %q", sb.String(), expected)
	}
}

// TestSerializeDoctypeSystemID tests DOCTYPE with SYSTEM ID only
func TestSerializeDoctypeSystemID(t *testing.T) {
	dt := dom.NewDocumentType("html", "", "http://www.w3.org/TR/html4/strict.dtd")

	var sb strings.Builder
	serializeDoctype(&sb, dt)

	expected := `<!DOCTYPE html SYSTEM "http://www.w3.org/TR/html4/strict.dtd">`
	if sb.String() != expected {
		t.Fatalf("unexpected doctype: %q, want %q", sb.String(), expected)
	}
}

// TestSerializeTextPrettyModeCollapseWhitespace tests text serialization in pretty mode
func TestSerializeTextPrettyModeCollapseWhitespace(t *testing.T) {
	text := dom.NewText("  hello   world  ")

	var sb strings.Builder
	serializeText(&sb, text, Options{Pretty: true, IndentSize: 2}, 0)

	// Pretty mode should collapse whitespace
	expected := " hello world "
	if sb.String() != expected {
		t.Fatalf("unexpected text output: %q, want %q", sb.String(), expected)
	}
}

// TestSerializeCommentPrettyModeWithDepth tests comment serialization in pretty mode with indentation
func TestSerializeCommentPrettyModeWithDepth(t *testing.T) {
	comment := dom.NewComment("test comment")

	var sb strings.Builder
	serializeComment(&sb, comment, Options{Pretty: true, IndentSize: 2}, 2, false)

	// Should have indentation (depth 2, indent size 2 = 4 spaces)
	expected := "    <!--test comment-->"
	if sb.String() != expected {
		t.Fatalf("unexpected comment output: %q, want %q", sb.String(), expected)
	}
}

// TestSerializeCommentInlineMode tests comment serialization in inline mode (no indentation)
func TestSerializeCommentInlineMode(t *testing.T) {
	comment := dom.NewComment("inline")

	var sb strings.Builder
	serializeComment(&sb, comment, Options{Pretty: true, IndentSize: 2}, 2, true)

	// Inline mode should not add indentation even in pretty mode
	expected := "<!--inline-->"
	if sb.String() != expected {
		t.Fatalf("unexpected comment output: %q, want %q", sb.String(), expected)
	}
}

// TestParseTokenAttrsObjectFormat tests parsing attributes in object format
func TestParseTokenAttrsObjectFormat(t *testing.T) {
	// Test object format: {"name": "value"}
	raw := json.RawMessage(`{"id": "foo", "class": "bar"}`)
	attrs := parseTokenAttrs(raw)

	if len(attrs) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(attrs))
	}

	// Check that both attributes are present (order may vary due to map iteration)
	hasID := false
	hasClass := false
	for _, attr := range attrs {
		if attr.Name == "id" && attr.Value == "foo" {
			hasID = true
		}
		if attr.Name == "class" && attr.Value == "bar" {
			hasClass = true
		}
	}
	if !hasID || !hasClass {
		t.Fatalf("missing expected attributes: %+v", attrs)
	}
}

// TestParseTokenAttrsEmptyArray tests parsing empty attribute array
func TestParseTokenAttrsEmptyArray(t *testing.T) {
	raw := json.RawMessage(`[]`)
	attrs := parseTokenAttrs(raw)

	if attrs != nil {
		t.Fatalf("expected nil for empty array, got %+v", attrs)
	}
}

// TestParseTokenAttrsEmptyObject tests parsing empty attribute object
func TestParseTokenAttrsEmptyObject(t *testing.T) {
	raw := json.RawMessage(`{}`)
	attrs := parseTokenAttrs(raw)

	if attrs != nil {
		t.Fatalf("expected nil for empty object, got %+v", attrs)
	}
}

// TestParseTokenAttrsInvalidJSON tests parsing invalid JSON
func TestParseTokenAttrsInvalidJSON(t *testing.T) {
	raw := json.RawMessage(`invalid`)
	attrs := parseTokenAttrs(raw)

	if attrs != nil {
		t.Fatalf("expected nil for invalid JSON, got %+v", attrs)
	}
}

// TestGetTokenInfoInvalidJSON tests getTokenInfo with invalid JSON
func TestGetTokenInfoInvalidJSON(t *testing.T) {
	raw := json.RawMessage(`invalid`)
	typ, tag := getTokenInfo(raw)

	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for invalid JSON, got type=%q, tag=%q", typ, tag)
	}
}

// TestGetTokenInfoEmptyArray tests getTokenInfo with empty array
func TestGetTokenInfoEmptyArray(t *testing.T) {
	raw := json.RawMessage(`[]`)
	typ, tag := getTokenInfo(raw)

	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for empty array, got type=%q, tag=%q", typ, tag)
	}
}

// TestGetTokenInfoInvalidTokenType tests getTokenInfo with invalid token type
func TestGetTokenInfoInvalidTokenType(t *testing.T) {
	raw := json.RawMessage(`[123]`)
	typ, tag := getTokenInfo(raw)

	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for invalid token type, got type=%q, tag=%q", typ, tag)
	}
}

// TestStartsWithSpaceErrorCases tests startsWithSpace error handling
func TestStartsWithSpaceErrorCases(t *testing.T) {
	// Test with invalid index
	if startsWithSpace([]json.RawMessage{}, 0) {
		t.Fatal("expected false for out of bounds index")
	}

	// Test with invalid JSON
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`invalid`),
	}
	if startsWithSpace(tokens, 0) {
		t.Fatal("expected false for invalid JSON")
	}

	// Test with array too short
	tokens = []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`["Characters"]`),
	}
	if startsWithSpace(tokens, 0) {
		t.Fatal("expected false for array too short")
	}

	// Test with invalid data field
	tokens = []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`["Characters", 123]`),
	}
	if startsWithSpace(tokens, 0) {
		t.Fatal("expected false for invalid data field")
	}

	// Test with empty data
	tokens = []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`["Characters", ""]`),
	}
	if startsWithSpace(tokens, 0) {
		t.Fatal("expected false for empty data")
	}
}

// TestGetPrevTokenInfoErrorCases tests getPrevTokenInfo error handling
func TestGetPrevTokenInfoErrorCases(t *testing.T) {
	// Test with idx at 0 (no previous token)
	tokens := []json.RawMessage{json.RawMessage(`["StartTag"]`)}
	typ, tag := getPrevTokenInfo(tokens, 0)
	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for idx=0, got type=%q, tag=%q", typ, tag)
	}

	// Test with invalid JSON
	tokens = []json.RawMessage{
		json.RawMessage(`invalid`),
		json.RawMessage(`["StartTag"]`),
	}
	typ, tag = getPrevTokenInfo(tokens, 1)
	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for invalid JSON, got type=%q, tag=%q", typ, tag)
	}

	// Test with empty array
	tokens = []json.RawMessage{
		json.RawMessage(`[]`),
		json.RawMessage(`["StartTag"]`),
	}
	typ, tag = getPrevTokenInfo(tokens, 1)
	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for empty array, got type=%q, tag=%q", typ, tag)
	}

	// Test with invalid token type
	tokens = []json.RawMessage{
		json.RawMessage(`[123]`),
		json.RawMessage(`["StartTag"]`),
	}
	typ, tag = getPrevTokenInfo(tokens, 1)
	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings for invalid token type, got type=%q, tag=%q", typ, tag)
	}
}

// TestSerializeInjectedMetaEmptyEncoding tests serializeInjectedMeta with empty encoding
func TestSerializeInjectedMetaEmptyEncoding(t *testing.T) {
	var sb strings.Builder
	opts := DefaultSerializeTokenOptions()
	opts.Encoding = "" // Empty encoding should return early

	serializeInjectedMeta(&sb, opts)

	if sb.String() != "" {
		t.Fatalf("expected empty output for empty encoding, got %q", sb.String())
	}
}

// TestSerializeInjectedMetaWithEncoding tests serializeInjectedMeta with valid encoding
func TestSerializeInjectedMetaWithEncoding(t *testing.T) {
	var sb strings.Builder
	opts := DefaultSerializeTokenOptions()
	opts.Encoding = testEncodingUTF8

	serializeInjectedMeta(&sb, opts)

	// Unquoted because testEncodingUTF8 doesn't contain special chars requiring quotes
	expected := `<meta charset=UTF-8>`
	if sb.String() != expected {
		t.Fatalf("unexpected meta tag: %q, want %q", sb.String(), expected)
	}
}

// TestTokenSerializationEdgeCases tests various edge cases in token serialization
func TestTokenSerializationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []json.RawMessage
		opts     SerializeTokenOptions
		expected string
		wantErr  bool
	}{
		{
			name: "EndTag missing fields",
			tokens: []json.RawMessage{
				rawToken(t, []any{"EndTag", "html"}),
			},
			opts:    DefaultSerializeTokenOptions(),
			wantErr: true,
		},
		{
			name: "EmptyTag missing fields",
			tokens: []json.RawMessage{
				rawToken(t, []any{"EmptyTag"}),
			},
			opts:    DefaultSerializeTokenOptions(),
			wantErr: true,
		},
		{
			name: "Characters missing data",
			tokens: []json.RawMessage{
				rawToken(t, []any{"Characters"}),
			},
			opts:    DefaultSerializeTokenOptions(),
			wantErr: true,
		},
		{
			name: "Comment missing data",
			tokens: []json.RawMessage{
				rawToken(t, []any{"Comment"}),
			},
			opts:    DefaultSerializeTokenOptions(),
			wantErr: true,
		},
		{
			name: "Doctype missing name",
			tokens: []json.RawMessage{
				rawToken(t, []any{"Doctype"}),
			},
			opts:    DefaultSerializeTokenOptions(),
			wantErr: true,
		},
		{
			name: "Empty token array",
			tokens: []json.RawMessage{
				rawToken(t, []any{}),
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: "",
			wantErr:  false,
		},
		{
			name: "StartTag with void element and trailing solidus",
			tokens: []json.RawMessage{
				rawToken(t, []any{"StartTag", "html", "br", []any{}}),
			},
			opts:     SerializeTokenOptions{UseTrailingSolidus: true},
			expected: "<br />",
			wantErr:  false,
		},
		{
			name: "Doctype with null publicID and systemID",
			tokens: []json.RawMessage{
				json.RawMessage(`["Doctype", "html", null, null]`),
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: "<!DOCTYPE html>",
			wantErr:  false,
		},
		{
			name: "Doctype with publicID and null systemID",
			tokens: []json.RawMessage{
				json.RawMessage(`["Doctype", "html", "-//W3C//DTD HTML 4.01//EN", null]`),
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN">`,
			wantErr:  false,
		},
		{
			name: "Doctype with publicID and systemID",
			tokens: []json.RawMessage{
				json.RawMessage(`["Doctype", "html", "-//W3C//DTD HTML 4.01//EN", "http://www.w3.org/TR/html4/strict.dtd"]`),
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">`,
			wantErr:  false,
		},
		{
			name: "Doctype with only systemID",
			tokens: []json.RawMessage{
				json.RawMessage(`["Doctype", "html", null, "http://www.w3.org/TR/html4/strict.dtd"]`),
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: `<!DOCTYPE html SYSTEM "http://www.w3.org/TR/html4/strict.dtd">`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := SerializeTokensWithOptions(tt.tokens, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tt.expected {
				t.Fatalf("unexpected output: %q, want %q", out, tt.expected)
			}
		})
	}
}

// TestAttributeValueEscapingEdgeCases tests edge cases in attribute value escaping
func TestAttributeValueEscapingEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []map[string]any
		opts     SerializeTokenOptions
		expected string
	}{
		{
			name: "Unquoted attribute value",
			attrs: []map[string]any{
				{"namespace": nil, "name": "class", "value": "simple"},
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: "<div class=simple></div>",
		},
		{
			name: "Attribute value with ampersand in single quotes",
			attrs: []map[string]any{
				{"namespace": nil, "name": "data", "value": `foo"bar&baz`},
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: `<div data='foo"bar&amp;baz'></div>`,
		},
		{
			name: "Attribute value with less-than when EscapeLtInAttrs is true",
			attrs: []map[string]any{
				{"namespace": nil, "name": "data", "value": "foo <bar"},
			},
			opts:     SerializeTokenOptions{EscapeLtInAttrs: true},
			expected: `<div data="foo &lt;bar"></div>`,
		},
		{
			name: "Attribute value with less-than when EscapeLtInAttrs is false",
			attrs: []map[string]any{
				{"namespace": nil, "name": "data", "value": "foo <bar"},
			},
			opts:     SerializeTokenOptions{EscapeLtInAttrs: false},
			expected: `<div data="foo <bar"></div>`,
		},
		{
			name: "Boolean attribute matching name",
			attrs: []map[string]any{
				{"namespace": nil, "name": "irrelevant", "value": "irrelevant"},
			},
			opts:     DefaultSerializeTokenOptions(),
			expected: `<div irrelevant></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := []json.RawMessage{
				rawToken(t, []any{"StartTag", "html", "div", tt.attrs}),
				rawToken(t, []any{"EndTag", "html", "div"}),
			}

			out, err := SerializeTokensWithOptions(tokens, tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tt.expected {
				t.Fatalf("unexpected output: %q, want %q", out, tt.expected)
			}
		})
	}
}

// TestCharactersTokenWhitespaceStripping tests whitespace stripping in Characters tokens
func TestCharactersTokenWhitespaceStripping(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.StripWhitespace = true

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", []any{}}),
		rawToken(t, []any{"Characters", "  hello   world  "}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Whitespace should be collapsed to single spaces
	expected := "<div> hello world </div>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestCharactersTokenInPreformatted tests Characters tokens inside <pre> (no whitespace stripping)
func TestCharactersTokenInPreformatted(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.StripWhitespace = true

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "pre", []any{}}),
		rawToken(t, []any{"Characters", "  hello   world  "}),
		rawToken(t, []any{"EndTag", "html", "pre"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Whitespace should be preserved in preformatted content
	expected := "<pre>  hello   world  </pre>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestMetaCharsetInjection tests meta charset injection functionality
func TestMetaCharsetInjection(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.InjectMetaCharset = true
	opts.Encoding = testEncodingUTF8
	opts.OmitOptionalTags = false // Don't omit tags for this test

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"StartTag", "html", "title", []any{}}),
		rawToken(t, []any{"Characters", "Test"}),
		rawToken(t, []any{"EndTag", "html", "title"}),
		rawToken(t, []any{"EndTag", "html", "head"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should inject meta charset at the beginning of head (unquoted)
	expected := `<head><meta charset=UTF-8><title>Test</title></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestMetaCharsetNormalization tests normalizing existing meta charset attributes
func TestMetaCharsetNormalization(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.InjectMetaCharset = true
	opts.Encoding = testEncodingUTF8
	opts.OmitOptionalTags = false // Don't omit tags for this test

	attrs := []map[string]any{
		{"namespace": nil, "name": "charset", "value": "ISO-8859-1"},
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"StartTag", "html", "meta", attrs}),
		rawToken(t, []any{"EndTag", "html", "head"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should normalize charset to UTF-8 (unquoted)
	expected := `<head><meta charset=UTF-8></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestMetaCharsetHTTPEquiv tests normalizing http-equiv meta tags
func TestMetaCharsetHTTPEquiv(t *testing.T) {
	attrs := []tokenAttr{
		{Name: "http-equiv", Value: "content-type"},
	}

	result := normalizeMetaCharsetAttrs(attrs, testEncodingUTF8)

	// Should add content attribute
	if len(result) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(result))
	}

	hasContent := false
	for _, attr := range result {
		if attr.Name == "content" && attr.Value == "text/html; charset=UTF-8" {
			hasContent = true
		}
	}
	if !hasContent {
		t.Fatalf("expected content attribute with charset, got %+v", result)
	}
}

// TestMetaCharsetHTTPEquivUpdateExisting tests updating existing content attribute
func TestMetaCharsetHTTPEquivUpdateExisting(t *testing.T) {
	attrs := []tokenAttr{
		{Name: "http-equiv", Value: "content-type"},
		{Name: "content", Value: "text/html; charset=ISO-8859-1"},
	}

	result := normalizeMetaCharsetAttrs(attrs, testEncodingUTF8)

	// Should update existing content attribute
	if len(result) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(result))
	}

	found := false
	for _, attr := range result {
		if attr.Name == "content" {
			if attr.Value != "text/html; charset=UTF-8" {
				t.Fatalf("expected updated charset in content, got %q", attr.Value)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("content attribute not found")
	}
}
