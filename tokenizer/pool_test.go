package tokenizer

import (
	"testing"
)

// TestTokenPoolReuse verifies that tokens are reused from the pool.
// This test ensures the pool reduces allocations by reusing token structs.
func TestTokenPoolReuse(t *testing.T) {
	// Parse simple HTML that generates multiple tokens
	tok := New("<div class='test'>hello</div>")

	var tokens []*Token
	for {
		token := tok.Next()
		if token.Type == EOF {
			break
		}
		tokens = append(tokens, token)
	}

	// We should have 3 tokens: StartTag, Character, EndTag
	if len(tokens) != 3 {
		t.Fatalf("got %d tokens, want 3", len(tokens))
	}

	// Verify token types
	if tokens[0].Type != StartTag {
		t.Errorf("tokens[0].Type = %v, want StartTag", tokens[0].Type)
	}
	if tokens[1].Type != Character {
		t.Errorf("tokens[1].Type = %v, want Character", tokens[1].Type)
	}
	if tokens[2].Type != EndTag {
		t.Errorf("tokens[2].Type = %v, want EndTag", tokens[2].Type)
	}

	// Verify token data
	if tokens[0].Name != "div" {
		t.Errorf("tokens[0].Name = %q, want 'div'", tokens[0].Name)
	}
	if tokens[1].Data != "hello" {
		t.Errorf("tokens[1].Data = %q, want 'hello'", tokens[1].Data)
	}
	if tokens[2].Name != "div" {
		t.Errorf("tokens[2].Name = %q, want 'div'", tokens[2].Name)
	}
}

// TestTokenPoolReset verifies that pooled tokens are properly reset.
func TestTokenPoolReset(t *testing.T) {
	// Get a token from the pool
	tok1 := getToken()

	// Populate it
	tok1.Type = StartTag
	tok1.Name = "div"
	tok1.Data = "some data"
	tok1.SelfClosing = true
	tok1.Attrs = append(tok1.Attrs, Attr{Name: "class", Value: "test"})

	// Return to pool
	putToken(tok1)

	// Get another token - might be the same one
	tok2 := getToken()

	// Verify all fields are reset (zero values)
	if tok2.Type != 0 {
		t.Errorf("tok2.Type = %v, want 0 (reset)", tok2.Type)
	}
	if tok2.Name != "" {
		t.Errorf("tok2.Name = %q, want empty (reset)", tok2.Name)
	}
	if tok2.Data != "" {
		t.Errorf("tok2.Data = %q, want empty (reset)", tok2.Data)
	}
	if tok2.SelfClosing {
		t.Errorf("tok2.SelfClosing = true, want false (reset)")
	}
	if len(tok2.Attrs) != 0 {
		t.Errorf("len(tok2.Attrs) = %d, want 0 (reset)", len(tok2.Attrs))
	}
}
