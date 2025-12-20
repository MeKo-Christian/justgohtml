//nolint:goconst // UTF-8 strings in test JSON and assertions
package serialize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// TestEscapeTextGreaterThan tests escaping > character
func TestEscapeTextGreaterThan(t *testing.T) {
	result := escapeText("a>b")
	expected := "a&gt;b"
	if result != expected {
		t.Fatalf("unexpected escaped text: %q, want %q", result, expected)
	}
}

// TestSerializeTextNonPrettyMode tests text serialization in non-pretty mode
func TestSerializeTextNonPrettyMode(t *testing.T) {
	text := dom.NewText("  hello   world  ")

	var sb strings.Builder
	serializeText(&sb, text, Options{Pretty: false, IndentSize: 2}, 0)

	// Non-pretty mode should preserve whitespace
	expected := "  hello   world  "
	if sb.String() != expected {
		t.Fatalf("unexpected text output: %q, want %q", sb.String(), expected)
	}
}

// TestCollapseWhitespaceEmptyResult tests collapsing whitespace with empty result
func TestCollapseWhitespaceEmptyResult(t *testing.T) {
	result := collapseWhitespace("   ")
	// All whitespace collapses to empty (leading/trailing space only restored if result has content)
	expected := ""
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestCollapseWhitespaceNoLeadingSpace tests collapsing whitespace without leading space
func TestCollapseWhitespaceNoLeadingSpace(t *testing.T) {
	result := collapseWhitespace("a   b ")
	// No leading space, has trailing space
	expected := "a b "
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestCollapseWhitespaceNoTrailingSpace tests collapsing whitespace without trailing space
func TestCollapseWhitespaceNoTrailingSpace(t *testing.T) {
	result := collapseWhitespace(" a   b")
	// Has leading space, no trailing space
	expected := " a b"
	if result != expected {
		t.Fatalf("unexpected collapsed whitespace: %q, want %q", result, expected)
	}
}

// TestHasAttributesEmptyArray tests hasAttributes with empty array
func TestHasAttributesEmptyArray(t *testing.T) {
	// ["StartTag", "html", "div", []]
	arr := []json.RawMessage{
		json.RawMessage(`"StartTag"`),
		json.RawMessage(`"html"`),
		json.RawMessage(`"div"`),
		json.RawMessage(`[]`),
	}

	if hasAttributes(arr) {
		t.Fatal("expected false for empty array")
	}
}

// TestHasAttributesEmptyObject tests hasAttributes with empty object
func TestHasAttributesEmptyObject(t *testing.T) {
	// ["StartTag", "html", "div", {}]
	arr := []json.RawMessage{
		json.RawMessage(`"StartTag"`),
		json.RawMessage(`"html"`),
		json.RawMessage(`"div"`),
		json.RawMessage(`{}`),
	}

	if hasAttributes(arr) {
		t.Fatal("expected false for empty object")
	}
}

// TestSerializeNodeWithInlineDocumentFragment tests serializing a DocumentFragment
func TestSerializeNodeWithInlineDocumentFragment(t *testing.T) {
	// Create a DocumentFragment (currently not handled in serializeNodeWithInline)
	// This tests the default case where node type is not explicitly handled
	fragment := &dom.DocumentFragment{}
	fragment.AppendChild(dom.NewText("test"))

	var sb strings.Builder
	// This should not panic, just do nothing for unhandled node types
	serializeNodeWithInline(&sb, fragment, Options{}, 0, false)

	// DocumentFragment is not handled in serializeNodeWithInline, so output should be empty
	if sb.String() != "" {
		t.Fatalf("unexpected output for DocumentFragment: %q", sb.String())
	}
}

// TestGetNextTokenInfoErrorInTagNameUnmarshal tests error handling in getNextTokenInfo
func TestGetNextTokenInfoErrorInTagNameUnmarshal(t *testing.T) {
	// Create a token where tag name unmarshal will fail
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag"]`),
		json.RawMessage(`["StartTag", "html", 123]`), // tag name is a number, not a string
	}

	typ, tag := getNextTokenInfo(tokens, 0)
	// Should return type but empty tag due to unmarshal error
	if typ != "StartTag" || tag != "" {
		t.Fatalf("expected type=StartTag, tag=empty, got type=%q, tag=%q", typ, tag)
	}
}

// TestGetPrevTokenInfoErrorInTagNameUnmarshal tests error handling in getPrevTokenInfo
func TestGetPrevTokenInfoErrorInTagNameUnmarshal(t *testing.T) {
	// Create a token where tag name unmarshal will fail
	tokens := []json.RawMessage{
		json.RawMessage(`["EndTag", "html", 123]`), // tag name is a number, not a string
		json.RawMessage(`["StartTag"]`),
	}

	typ, tag := getPrevTokenInfo(tokens, 1)
	// Should return type but empty tag due to unmarshal error
	if typ != "EndTag" || tag != "" {
		t.Fatalf("expected type=EndTag, tag=empty, got type=%q, tag=%q", typ, tag)
	}
}

// TestSerializeEmptyTagTokenNoAttributes tests EmptyTag without attributes
func TestSerializeEmptyTagTokenNoAttributes(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.UseTrailingSolidus = false

	tokens := []json.RawMessage{
		rawToken(t, []any{"EmptyTag", "br"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<br>"
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestSerializeStartTagTokenInvalidTagName tests error handling for invalid tag name
func TestSerializeStartTagTokenInvalidTagName(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", 123]`), // Invalid tag name (number instead of string)
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid tag name")
	}
}

// TestSerializeEndTagTokenInvalidTagName tests error handling for invalid tag name
func TestSerializeEndTagTokenInvalidTagName(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["EndTag", "html", 123]`), // Invalid tag name (number instead of string)
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid tag name")
	}
}

// TestSerializeEmptyTagTokenInvalidTagName tests error handling for invalid tag name
func TestSerializeEmptyTagTokenInvalidTagName(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["EmptyTag", 123]`), // Invalid tag name (number instead of string)
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid tag name")
	}
}

// TestSerializeCommentTokenInvalidData tests error handling for invalid comment data
func TestSerializeCommentTokenInvalidData(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["Comment", 123]`), // Invalid data (number instead of string)
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid comment data")
	}
}

// TestSerializeCharactersTokenInvalidData tests error handling for invalid characters data
func TestSerializeCharactersTokenInvalidData(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["Characters", 123]`), // Invalid data (number instead of string)
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid characters data")
	}
}

// TestSerializeDoctypeTokenInvalidName tests error handling for invalid doctype name
func TestSerializeDoctypeTokenInvalidName(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["Doctype", 123]`), // Invalid name (number instead of string)
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for invalid doctype name")
	}
}

// TestSerializeTokenAttrValueEmptyValue tests attribute with truly empty value
func TestSerializeTokenAttrValueEmptyValue(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.MinimizeBooleanAttributes = false // Don't minimize

	attrs := []map[string]any{
		{"namespace": nil, "name": "data-val", "value": ""},
	}

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "div", attrs}),
		rawToken(t, []any{"EndTag", "html", "div"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty value should serialize as =""
	expected := `<div data-val=""></div>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestHasCharsetMetaAheadInvalidTokenJSON tests error handling in hasCharsetMetaAhead
func TestHasCharsetMetaAheadInvalidTokenJSON(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`invalid`), // Invalid JSON
	}

	result := hasCharsetMetaAhead(tokens, 0)
	if result {
		t.Fatal("expected false when token JSON is invalid")
	}
}

// TestHasCharsetMetaAheadCaseInsensitiveCharset tests case-insensitive charset detection
func TestHasCharsetMetaAheadCaseInsensitiveCharset(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`["StartTag", "html", "head", []]`),
		json.RawMessage(`["StartTag", "html", "meta", [{"namespace": null, "name": "CHARSET", "value": "UTF-8"}]]`),
		json.RawMessage(`["EndTag", "html", "head"]`),
	}

	result := hasCharsetMetaAhead(tokens, 0)
	if !result {
		t.Fatal("expected true for case-insensitive charset attribute")
	}
}

// TestNormalizeMetaCharsetAttrsEmptyAttrs tests normalizing with empty attrs
func TestNormalizeMetaCharsetAttrsEmptyAttrs(t *testing.T) {
	attrs := []tokenAttr{}
	result := normalizeMetaCharsetAttrs(attrs, "UTF-8")

	if len(result) != 0 {
		t.Fatalf("expected empty result for empty attrs, got %+v", result)
	}
}

// TestNormalizeMetaCharsetAttrsCaseInsensitive tests case-insensitive attribute matching
func TestNormalizeMetaCharsetAttrsCaseInsensitive(t *testing.T) {
	attrs := []tokenAttr{
		{Name: "CHARSET", Value: "ISO-8859-1"},
	}

	result := normalizeMetaCharsetAttrs(attrs, "UTF-8")

	if len(result) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(result))
	}

	if result[0].Value != "UTF-8" {
		t.Fatalf("expected charset normalized to UTF-8, got %q", result[0].Value)
	}
}

// TestMetaCharsetInjectionAtEndOfHead tests meta injection when there's no charset ahead
func TestMetaCharsetInjectionAtEndOfHead(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.InjectMetaCharset = true
	opts.Encoding = "UTF-8"
	opts.OmitOptionalTags = false

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "head", []any{}}),
		rawToken(t, []any{"StartTag", "html", "title", []any{}}),
		rawToken(t, []any{"Characters", "Test"}),
		rawToken(t, []any{"EndTag", "html", "title"}),
		rawToken(t, []any{"EndTag", "html", "head"}), // Meta should be injected before this
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Meta should be injected at the start of head since there's no charset ahead
	expected := `<head><meta charset=UTF-8><title>Test</title></head>`
	if out != expected {
		t.Fatalf("unexpected output: %q, want %q", out, expected)
	}
}

// TestInvalidTokenTypeInArray tests handling of invalid token type in array element
func TestInvalidTokenTypeInArray(t *testing.T) {
	tokens := []json.RawMessage{
		json.RawMessage(`[null]`), // Token type is null
	}

	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil {
		t.Fatal("expected error for null token type")
	}
}
