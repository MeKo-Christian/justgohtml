package tokenizer

import "testing"

func collectTokens(html string, initial State) []Token {
	tok := New(html)
	tok.SetState(initial)
	var out []Token
	for {
		t := tok.Next()
		if t.Type == EOF {
			break
		}
		out = append(out, t)
	}
	return out
}

func TestTokenizer_BOMDiscard(t *testing.T) {
	tok := New("\ufeff<div>")
	tok.SetDiscardBOM(true)
	var tokens []Token
	for {
		tt := tok.Next()
		if tt.Type == EOF {
			break
		}
		tokens = append(tokens, tt)
	}
	if len(tokens) != 1 || tokens[0].Type != StartTag || tokens[0].Name != "div" {
		t.Fatalf("tokens = %#v, want single StartTag(div)", tokens)
	}
}

func TestTokenizer_CRLFNormalization(t *testing.T) {
	tokens := collectTokens("a\r\nb\rc", DataState)
	if len(tokens) != 1 || tokens[0].Type != Character {
		t.Fatalf("tokens = %#v, want single Character", tokens)
	}
	if tokens[0].Data != "a\nb\nc" {
		t.Fatalf("data = %q, want %q", tokens[0].Data, "a\nb\nc")
	}
}

func TestTokenizer_XMLCoercion(t *testing.T) {
	tok := New("\f\uFDD0")
	tok.SetXMLCoercion(true)
	var tokens []Token
	for {
		tt := tok.Next()
		if tt.Type == EOF {
			break
		}
		tokens = append(tokens, tt)
	}
	if len(tokens) != 1 || tokens[0].Type != Character {
		t.Fatalf("tokens = %#v, want single Character", tokens)
	}
	if tokens[0].Data != " \ufffd" {
		t.Fatalf("data = %q, want %q", tokens[0].Data, " \ufffd")
	}
}

func TestTokenizer_NullInAttrNameAndValue(t *testing.T) {
	tokens := collectTokens("<div a\u0000b='b\u0000c'>", DataState)
	if len(tokens) != 1 || tokens[0].Type != StartTag {
		t.Fatalf("tokens = %#v, want single StartTag", tokens)
	}
	if got := tokens[0].Attrs["a\ufffdb"]; got != "b\ufffdc" {
		t.Fatalf("attrs = %#v, want a\\ufffdb=b\\ufffdc", tokens[0].Attrs)
	}
}

func TestTokenizer_MissingAttrValue(t *testing.T) {
	tokens := collectTokens("<div a=>", DataState)
	if len(tokens) != 1 || tokens[0].Type != StartTag {
		t.Fatalf("tokens = %#v, want StartTag", tokens)
	}
	if got := tokens[0].Attrs["a"]; got != "" {
		t.Fatalf("attrs[a] = %q, want empty", got)
	}
}

func TestTokenizer_SwitchToRCDATAForTitle(t *testing.T) {
	tok := New("<title>Hi &amp; bye</title>")
	var kinds []TokenKind
	var datas []string
	for {
		t := tok.Next()
		if t.Type == EOF {
			break
		}
		kinds = append(kinds, t.Type)
		datas = append(datas, t.Data)
	}
	if len(kinds) != 3 || kinds[0] != StartTag || kinds[1] != Character || kinds[2] != EndTag {
		t.Fatalf("kinds = %#v, want [StartTag Character EndTag]", kinds)
	}
	if datas[1] != "Hi & bye" {
		t.Fatalf("data = %q, want entity-decoded text", datas[1])
	}
}
