package serialize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// TestCollapseWhitespaceLeadingAndTrailing tests both leading and trailing space restoration
func TestCollapseWhitespaceLeadingAndTrailing(t *testing.T) {
	// This specifically tests the path where both leading and trailing spaces are restored
	result := collapseWhitespace(" a ")

	expected := " a "
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestCollapseWhitespaceComplexCase tests a complex case with multiple spaces
func TestCollapseWhitespaceComplexCase(t *testing.T) {
	// This tests the trimming of trailing space from collapsed content
	result := collapseWhitespace("  a   b   c  ")

	expected := " a b c "
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestSerializeStartTagTokenVoidWithoutTrailingSolidus tests void element without trailing solidus
func TestSerializeStartTagTokenVoidWithoutTrailingSolidus(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.UseTrailingSolidus = false

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "img", []any{}}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<img>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeEmptyTagTokenWithoutTrailingSolidus tests EmptyTag without trailing solidus
func TestSerializeEmptyTagTokenWithoutTrailingSolidus(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.UseTrailingSolidus = false

	attrs := []map[string]any{
		{"namespace": nil, "name": "src", "value": "image.png"},
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"EmptyTag", "img", attrs}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<img src=image.png>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestGetNextTokenInfoCharactersToken tests getNextTokenInfo with Characters token
func TestGetNextTokenInfoCharactersToken(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`["Characters", "test"]`),
	}

	typ, tag := getNextTokenInfo(tokens, 0)

	// Characters tokens don't have a tag name
	if typ != "Characters" || tag != "" {
		t.Fatalf("expected type=Characters, tag=empty, got type=%q, tag=%q", typ, tag)
	}
}

// TestHasCharsetMetaAheadNonMetaStartTag tests hasCharsetMetaAhead skipping non-meta start tags
func TestHasCharsetMetaAheadNonMetaStartTag(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["StartTag", "html", "title", []]`), // Non-meta tag, should skip
		json.RawMessage(`["StartTag", "html", "meta", [{"namespace": null, "name": "charset", "value": "UTF-8"}]]`),
		json.RawMessage(`["EndTag", "html", "head"]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)

	if !result {
		t.Fatal("expected true, should find meta charset after skipping title")
	}
}

// TestHasCharsetMetaAheadReturnsOnEmptyType tests hasCharsetMetaAhead early return on empty type
func TestHasCharsetMetaAheadReturnsOnEmptyType(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`invalid`), // This will cause getTokenInfo to return empty type
	}

	result := hasCharsetMetaAhead(tokens, 0)

	if result {
		t.Fatal("expected false when encountering invalid token (empty type)")
	}
}

// TestSerializeTokensMetaInjectionBeforeEndHead tests meta injection before </head>
func TestSerializeTokensMetaInjectionBeforeEndHead(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.InjectMetaCharset = true
	opts.Encoding = "UTF-8"
	opts.OmitOptionalTags = false

	// Test the path where meta is injected right before </head> when no charset is found
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"EndTag", "html", "head"}), // Meta should be injected here
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Meta should be injected at start of head
	expected := `<head><meta charset=UTF-8></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokensNoMetaInjectionWhenCharsetExists tests that meta is not injected when charset exists
func TestSerializeTokensNoMetaInjectionWhenCharsetExists(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.InjectMetaCharset = true
	opts.Encoding = "UTF-8"
	opts.OmitOptionalTags = false

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"EmptyTag", "meta", []map[string]any{{"namespace": nil, "name": "charset", "value": "ISO-8859-1"}}}),
		rawToken(t, []any{"EndTag", "html", "head"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT inject meta, but should normalize existing charset
	// There should be only ONE meta tag
	expected := `<head><meta charset=UTF-8></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokensEndTagBeforeMetaInjection tests the specific branch where
// we check for end head tag before injecting meta
func TestSerializeTokensEndTagBeforeMetaInjection(t *testing.T) {
	opts := DefaultSerializeTokenOptions()

	opts.InjectMetaCharset = true
	opts.Encoding = "UTF-8"
	opts.OmitOptionalTags = false

	// This tests line 121-126 in SerializeTokensWithOptions
	// where we check for EndTag head before processing other EndTags
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"StartTag", "html", "div", []any{}}), // Non-charset content
		rawToken(t, []any{"EndTag", "html", "div"}),
		rawToken(t, []any{"EndTag", "html", "head"}), // This triggers the meta injection check
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Meta should be injected at start of head
	expected := `<head><meta charset=UTF-8><div></div></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestHasAttributesWithNonEmptyArray tests hasAttributes with non-empty array
func TestHasAttributesWithNonEmptyArray(t *testing.T) {
	// This specifically tests the path where unmarshal succeeds AND len > 0
	arr := []json.RawMessage{
		json.RawMessage(`"StartTag"`),
		json.RawMessage(`"html"`),
		json.RawMessage(`"div"`),
		json.RawMessage(`[{"namespace": null, "name": "class", "value": "test"}]`),
	}

	if !hasAttributes(arr) {
		t.Fatal("expected true for non-empty attribute array")
	}
}

// TestHasAttributesWithNonEmptyObject tests hasAttributes with non-empty object
func TestHasAttributesWithNonEmptyObject(t *testing.T) {
	// This specifically tests the path where unmarshal to object succeeds AND len > 0
	arr := []json.RawMessage{
		json.RawMessage(`"StartTag"`),
		json.RawMessage(`"html"`),
		json.RawMessage(`"div"`),
		json.RawMessage(`{"class": "test"}`),
	}

	if !hasAttributes(arr) {
		t.Fatal("expected true for non-empty attribute object")
	}
}

// TestSerializeDocumentFragmentNode tests serializing an unhandled node type
func TestSerializeDocumentFragmentNode(t *testing.T) {
	// DocumentFragment is not handled in serializeNodeWithInline switch
	fragment := &dom.DocumentFragment{}

	var sb strings.Builder
	opts := DefaultOptions()

	// Call serializeNode which calls serializeNodeWithInline
	serializeNode(&sb, fragment, opts, 0)

	// Should produce empty output (unhandled node type does nothing)

	if sb.String() != "" {
		t.Fatalf("expected empty output for DocumentFragment, got %q", sb.String())
	}
}
