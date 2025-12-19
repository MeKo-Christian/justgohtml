package encoding

import (
	"errors"
	"testing"
)

func TestDecodeWithEncodingInvalid(t *testing.T) {
	_, err := decodeWithEncoding([]byte("x"), &Encoding{Name: "bogus"})
	if !errors.Is(err, ErrInvalidEncoding) {
		t.Fatalf("expected ErrInvalidEncoding, got %v", err)
	}
}

func TestNormalizeMetaDeclaredEncoding(t *testing.T) {
	enc := normalizeMetaDeclaredEncoding([]byte("utf-16"))
	if enc == nil || enc.Name != "UTF-8" {
		t.Fatalf("expected UTF-8, got %#v", enc)
	}

	enc = normalizeMetaDeclaredEncoding([]byte("utf-32"))
	if enc != nil {
		t.Fatalf("expected nil for unsupported utf-32, got %#v", enc)
	}

	enc = normalizeMetaDeclaredEncoding([]byte("iso-8859-2"))
	if enc == nil || enc.Name != "iso-8859-2" {
		t.Fatalf("expected iso-8859-2, got %#v", enc)
	}
}

func TestPrescanForMetaCharset(t *testing.T) {
	data := []byte("<!-- comment --><meta charset=\"utf-8\">")
	enc := prescanForMetaCharset(data)
	if enc == nil || enc.Name != "UTF-8" {
		t.Fatalf("expected UTF-8, got %#v", enc)
	}

	data = []byte("<meta http-equiv=\"content-type\" content=\"text/html; charset=ascii\">")
	enc = prescanForMetaCharset(data)
	if enc == nil || enc.Name != "windows-1252" {
		t.Fatalf("expected windows-1252, got %#v", enc)
	}
}

func TestASCIIHelpers(t *testing.T) {
	if !isASCIIWhitespace('\t') {
		t.Fatal("expected tab to be ASCII whitespace")
	}
	if isASCIIWhitespace('A') {
		t.Fatal("expected 'A' to not be ASCII whitespace")
	}
	if !isASCIIAlpha('Z') {
		t.Fatal("expected 'Z' to be ASCII alpha")
	}
	if isASCIIAlpha('1') {
		t.Fatal("expected '1' to not be ASCII alpha")
	}
	if asciiLower('Z') != 'z' {
		t.Fatalf("expected asciiLower('Z') to be 'z', got %q", asciiLower('Z'))
	}
	if asciiLower('!') != '!' {
		t.Fatalf("expected asciiLower('!') to be '!', got %q", asciiLower('!'))
	}
}
