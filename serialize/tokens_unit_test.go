package serialize

import (
	"encoding/json"
	"errors"
	"testing"
)

func rawToken(t *testing.T, token any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("marshal token: %v", err)
	}
	return data
}

func TestSerializeTokensInvalidJSON(t *testing.T) {
	_, err := SerializeTokensWithOptions([]json.RawMessage{json.RawMessage("invalid")}, DefaultSerializeTokenOptions())
	if err == nil || !errors.Is(err, ErrInvalidTokenFormat) {
		t.Fatalf("expected invalid token format error, got %v", err)
	}
}

func TestSerializeTokensUnknownType(t *testing.T) {
	tokens := []json.RawMessage{
		rawToken(t, []any{"Bogus"}),
	}
	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil || !errors.Is(err, ErrUnknownTokenType) {
		t.Fatalf("expected unknown token type error, got %v", err)
	}
}

func TestSerializeTokensMissingFields(t *testing.T) {
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag"}),
	}
	_, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err == nil || !errors.Is(err, ErrStartTagMissingFields) {
		t.Fatalf("expected missing fields error, got %v", err)
	}
}

func TestSerializeTokensAttributeQuoting(t *testing.T) {
	attrs := []map[string]any{
		{"namespace": nil, "name": "title", "value": `foo"bar`},
	}
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "span", attrs}),
		rawToken(t, []any{"EndTag", "html", "span"}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<span title='foo\"bar'></span>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestSerializeTokensQuoteCharOption(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.QuoteChar = '\''

	attrs := []map[string]any{
		{"namespace": nil, "name": "title", "value": "foo'bar"},
	}
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "span", attrs}),
		rawToken(t, []any{"EndTag", "html", "span"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<span title='foo&#39;bar'></span>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestSerializeTokensMinimizeBooleanAttributes(t *testing.T) {
	attrs := []map[string]any{
		{"namespace": nil, "name": "disabled", "value": ""},
	}
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "input", attrs}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<input disabled>" {
		t.Fatalf("unexpected output: %q", out)
	}

	opts := DefaultSerializeTokenOptions()
	opts.MinimizeBooleanAttributes = false
	out, err = SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<input disabled=\"\">" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestSerializeTokensRawTextEscaping(t *testing.T) {
	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "script", []any{}}),
		rawToken(t, []any{"Characters", "<b>"}),
		rawToken(t, []any{"EndTag", "html", "script"}),
	}

	out, err := SerializeTokensWithOptions(tokens, DefaultSerializeTokenOptions())
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<script><b></script>" {
		t.Fatalf("unexpected output: %q", out)
	}

	opts := DefaultSerializeTokenOptions()
	opts.EscapeRcdata = true
	out, err = SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<script>&lt;b&gt;</script>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestSerializeTokensOmitOptionalTags(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.OmitOptionalTags = true

	tokens := []json.RawMessage{
		rawToken(t, []any{"StartTag", "html", "html", []any{}}),
		rawToken(t, []any{"EndTag", "html", "html"}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestSerializeTokensTrailingSolidus(t *testing.T) {
	opts := DefaultSerializeTokenOptions()
	opts.UseTrailingSolidus = true

	tokens := []json.RawMessage{
		rawToken(t, []any{"EmptyTag", "img", []any{}}),
	}

	out, err := SerializeTokensWithOptions(tokens, opts)
	if err != nil {
		t.Fatalf("SerializeTokensWithOptions error: %v", err)
	}
	if out != "<img />" {
		t.Fatalf("unexpected output: %q", out)
	}
}
