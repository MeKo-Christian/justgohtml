package serialize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// TestSerializeTextPrettyWhitespaceOnly tests early return for whitespace-only text in pretty mode
func TestSerializeTextPrettyWhitespaceOnly(t *testing.T) {
	text := dom.NewText("   \n\t  ")

	var sb strings.Builder
	serializeText(&sb, text, Options{Pretty: true, IndentSize: 2}, 0)

	// Should return early without writing anything
	if sb.String() != "" {
		t.Fatalf("unexpected output for whitespace-only text in pretty mode: %q", sb.String())
	}
}

// TestCollapseWhitespaceTrimTrailing tests trimming trailing space after collapse
func TestCollapseWhitespaceTrimTrailing(t *testing.T) {
	// Input with internal whitespace that will create trailing space during collapse
	result := collapseWhitespace("a  ")
	// Should be "a " (trailing space preserved from original)
	expected := "a "
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestCollapseWhitespaceInternalWhitespace tests internal whitespace handling
func TestCollapseWhitespaceInternalWhitespace(t *testing.T) {
	// This should trigger the path where we trim trailing space from collapsed content
	// then restore it based on hasTrailingSpace
	result := collapseWhitespace("a b  ")
	expected := "a b "
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestSerializeTokensInvalidTokenTypeJSON tests invalid JSON in token type field
func TestSerializeTokensInvalidTokenTypeJSON(t *testing.T) {
	// Array with invalid token type field (can't unmarshal)
	tokens := []json.RawMessage{
		json.RawMessage(`[{"invalid": "object"}]`), // Token type is an object, not a string
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid token type JSON")
	}
}

// TestSerializeEmptyTagTokenWithAttributes tests EmptyTag with attributes
func TestSerializeEmptyTagTokenWithAttributes(t *testing.T) {
	attrs := []map[string]any{
		{"namespace": nil, "name": "id", "value": "test"},
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"EmptyTag", "img", attrs}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<img id=test>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeStartTagTokenWithAttributes tests StartTag serialization with attributes
func TestSerializeStartTagTokenWithAttributes(t *testing.T) {
	attrs := []map[string]any{
		{"namespace": nil, "name": "class", "value": "test"},
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", attrs}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<div class=test></div>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokensPreformattedDepthTracking tests <pre> and <textarea> depth tracking
func TestSerializeTokensPreformattedDepthTracking(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.StripWhitespace = true

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "textarea", []any{}}),
		rawToken(t, []any{"Characters", "  preserve  "}),
		rawToken(t, []any{"EndTag", "html", "textarea"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Whitespace should be preserved inside textarea
	expected := "<textarea>  preserve  </textarea>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestGetNextTokenInfoEmptyTagFormat tests getNextTokenInfo with EmptyTag format
func TestGetNextTokenInfoEmptyTagFormat(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`["EmptyTag", "br"]`),
	}

	typ, tag := getNextTokenInfo(tokens, 0)
	if typ != "EmptyTag" || tag != "br" {
		t.Fatalf("expected type=EmptyTag, tag=br, got type=%q, tag=%q", typ, tag)
	}
}

// TestGetPrevTokenInfoEmptyTagFormat tests getPrevTokenInfo with EmptyTag format
func TestGetPrevTokenInfoEmptyTagFormat(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["EmptyTag", "br"]`),
		json.RawMessage(`["StartTag"]`),
	}

	typ, tag := getPrevTokenInfo(tokens, 1)
	if typ != "EmptyTag" || tag != "br" {
		t.Fatalf("expected type=EmptyTag, tag=br, got type=%q, tag=%q", typ, tag)
	}
}

// TestHasCharsetMetaAheadEmptyTagMeta tests hasCharsetMetaAhead with EmptyTag meta
func TestHasCharsetMetaAheadEmptyTagMeta(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["EmptyTag", "meta", [{"namespace": null, "name": "charset", "value": "UTF-8"}]]`),
		json.RawMessage(`["EndTag", "html", "head"]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)
	if !result {
		t.Fatal("expected true for EmptyTag meta with charset")
	}
}

// TestHasCharsetMetaAheadHTTPEquivContentType tests http-equiv detection
func TestHasCharsetMetaAheadHTTPEquivContentType(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["StartTag", "html", "meta", [{"namespace": null, "name": "http-equiv", "value": "content-type"}]]`),
		json.RawMessage(`["EndTag", "html", "head"]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)
	if !result {
		t.Fatal("expected true for http-equiv content-type meta")
	}
}

// TestSerializeDoctypeTokenWithInvalidPublicID tests doctype with invalid publicID JSON
func TestSerializeDoctypeTokenWithInvalidPublicID(t *testing.T) {
	// publicID field is an object instead of string/null
	tokens := []json.RawMessage{
		json.RawMessage(`["Doctype", "html", {"invalid": "object"}, null]`),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should treat invalid publicID as empty string
	expected := "<!DOCTYPE html>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeDoctypeTokenWithInvalidSystemID tests doctype with invalid systemID JSON
func TestSerializeDoctypeTokenWithInvalidSystemID(t *testing.T) {
	// systemID field is an object instead of string/null
	tokens := []json.RawMessage{
		json.RawMessage(`["Doctype", "html", "-//W3C//DTD HTML 4.01//EN", {"invalid": "object"}]`),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should treat invalid systemID as empty string
	expected := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN">`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokenAttrValueQuoteCharZero tests default quote char when QuoteChar is zero
func TestSerializeTokenAttrValueQuoteCharZero(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.QuoteChar = 0 // Zero value should default to double quote

	attrs := []map[string]any{
		{"namespace": nil, "name": "title", "value": "foo bar"}, // Needs quoting due to space
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", attrs}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use double quotes when QuoteChar is 0
	expected := `<div title="foo bar"></div>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeTokenAttrValueAmpersandEscaping tests & escaping in single-quoted attributes
func TestSerializeTokenAttrValueAmpersandEscaping(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.QuoteChar = '\''

	attrs := []map[string]any{
		{"namespace": nil, "name": "data", "value": "foo&bar"},
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", attrs}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should escape & even when using single quotes
	expected := `<div data='foo&amp;bar'></div>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestHasCharsetMetaAheadReturnsOnEndTag tests early return when encountering end head tag
func TestHasCharsetMetaAheadReturnsOnEndTag(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["StartTag", "html", "title", []]`),
		json.RawMessage(`["EndTag", "html", "head"]`), // Should stop here and return false
		json.RawMessage(`["EmptyTag", "meta", [{"namespace": null, "name": "charset", "value": "UTF-8"}]]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)
	if result {
		t.Fatal("expected false when charset meta comes after head end tag")
	}
}

// TestHasCharsetMetaAheadEmptyRawAttrs tests handling of empty rawAttrs in hasCharsetMetaAhead
func TestHasCharsetMetaAheadEmptyRawAttrs(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["StartTag", "html", "meta"]`), // No attrs field (too short array)
		json.RawMessage(`["EndTag", "html", "head"]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)
	if result {
		t.Fatal("expected false when meta tag has no attributes")
	}
}
