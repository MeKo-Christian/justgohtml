package serialize

import (
	"encoding/json"
	"testing"
)

// TestSerializeTokensRawTextNesting tests nested raw text elements (script in script)

func TestSerializeTokensRawTextNesting(t *testing.T) {
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "script", []any{}}),
		rawToken(t, []any{"StartTag", "html", "style", []any{}}), // Nested raw text
		rawToken(t, []any{"Characters", "<>"}),
		rawToken(t, []any{"EndTag", "html", "style"}),
		rawToken(t, []any{"EndTag", "html", "script"}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should preserve <> without escaping
	expected := "<script><style><></style></script>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokensMultiplePreformattedNesting tests nested preformatted elements
func TestSerializeTokensMultiplePreformattedNesting(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.StripWhitespace = true

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "pre", []any{}}),
		rawToken(t, []any{"StartTag", "html", "pre", []any{}}), // Nested pre
		rawToken(t, []any{"Characters", "  nested  "}),
		rawToken(t, []any{"EndTag", "html", "pre"}),
		rawToken(t, []any{"EndTag", "html", "pre"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Whitespace should be preserved due to nesting
	expected := "<pre><pre>  nested  </pre></pre>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokensDecrementDepthAtZero tests decrementing depth when already at zero
func TestSerializeTokensDecrementDepthAtZero(t *testing.T) {
	// This tests the safety check: if preformattedDepth > 0 before decrementing
	tokens := []json.RawMessage{
		rawToken(t, []any{"EndTag", "html", "pre"}), // End without start
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle gracefully (pre depth check prevents going negative)

	expected := "</pre>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokensRawTextDepthAtZero tests decrementing raw text depth at zero
func TestSerializeTokensRawTextDepthAtZero(t *testing.T) {
	// This tests the safety check for rawTextDepth > 0
	tokens := []json.RawMessage{
		rawToken(t, []any{"EndTag", "html", "script"}), // End without start
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should handle gracefully
	expected := "</script>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestHasAttributesShortArray tests hasAttributes with array too short
func TestHasAttributesShortArray(t *testing.T) {
	// Array with only 3 elements (no attrs field)

	arr := []json.RawMessage{
		json.RawMessage(`"StartTag"`),
		json.RawMessage(`"html"`),
		json.RawMessage(`"div"`),
	}

	if hasAttributes(arr) {
		t.Fatal("expected false for array too short")
	}
}

// TestSerializeStartTagTokenWithEmptyAttrs tests start tag with empty attrs object
func TestSerializeStartTagTokenWithEmptyAttrs(t *testing.T) {
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", map[string]string{}}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<div></div>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeEmptyTagTokenWithEmptyAttrs tests empty tag with attrs field present but empty
func TestSerializeEmptyTagTokenWithEmptyAttrs(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["EmptyTag", "br", {}]`),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<br>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestHasCharsetMetaAheadEmptyTagWithShortArray tests EmptyTag meta with short array (no attrs)

func TestHasCharsetMetaAheadEmptyTagWithShortArray(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["EmptyTag", "meta"]`), // Too short, no attrs
		json.RawMessage(`["EndTag", "html", "head"]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)

	if result {
		t.Fatal("expected false for EmptyTag meta with no attributes")
	}
}

// TestSerializeTokensComplexMetaInjectionScenario tests complex meta injection flow
func TestSerializeTokensComplexMetaInjectionScenario(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.InjectMetaCharset = true
	opts.Encoding = testEncodingUTF8
	opts.OmitOptionalTags = false

	// Test where we have content but no meta charset ahead
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"StartTag", "html", "link", []map[string]any{{"namespace": nil, "name": "rel", "value": "stylesheet"}}}),
		rawToken(t, []any{"EndTag", "html", "head"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Meta should be injected at start since there's no charset ahead
	expected := `<head><meta charset=UTF-8><link rel=stylesheet></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestGetNextTokenInfoAtEnd tests getNextTokenInfo when at the last token
func TestGetNextTokenInfoAtEnd(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
	}

	typ, tag := getNextTokenInfo(tokens, 0)

	// At end, should return empty strings
	if typ != "" || tag != "" {
		t.Fatalf("expected empty strings at end, got type=%q, tag=%q", typ, tag)
	}
}
