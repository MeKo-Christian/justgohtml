package tokenizer

import "testing"

func TestTokenKindString(t *testing.T) {
	tests := []struct {
		kind TokenKind
		want string
	}{
		{Error, "Error"},
		{DOCTYPE, "DOCTYPE"},
		{StartTag, "StartTag"},
		{EndTag, "EndTag"},
		{Comment, "Comment"},
		{Character, "Character"},
		{EOF, "EOF"},
		{TokenKind(-1), "Unknown"},
		{TokenKind(123), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Fatalf("TokenKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestTokenAttrHelpers(t *testing.T) {
	var tok Token
	if got := tok.AttrVal("id"); got != "" {
		t.Fatalf("AttrVal on nil Attrs = %q, want empty", got)
	}
	if got := tok.HasAttr("id"); got {
		t.Fatalf("HasAttr on nil Attrs = true, want false")
	}

	tok.Attrs = map[string]string{"id": "x"}
	if got := tok.AttrVal("id"); got != "x" {
		t.Fatalf("AttrVal(id) = %q, want %q", got, "x")
	}
	if got := tok.AttrVal("class"); got != "" {
		t.Fatalf("AttrVal(class) = %q, want empty", got)
	}
	if got := tok.HasAttr("id"); !got {
		t.Fatalf("HasAttr(id) = false, want true")
	}
	if got := tok.HasAttr("class"); got {
		t.Fatalf("HasAttr(class) = true, want false")
	}
}

func TestTokenConstructors(t *testing.T) {
	if got := NewStartTagToken("div"); got.Type != StartTag || got.Name != "div" {
		t.Fatalf("NewStartTagToken = %#v, want Type=StartTag Name=div", got)
	}
	if got := NewEndTagToken("div"); got.Type != EndTag || got.Name != "div" {
		t.Fatalf("NewEndTagToken = %#v, want Type=EndTag Name=div", got)
	}
	if got := NewCharacterToken("x"); got.Type != Character || got.Data != "x" {
		t.Fatalf("NewCharacterToken = %#v, want Type=Character Data=x", got)
	}
	if got := NewCommentToken("x"); got.Type != Comment || got.Data != "x" {
		t.Fatalf("NewCommentToken = %#v, want Type=Comment Data=x", got)
	}

	pub := "pub"
	sys := "sys"
	got := NewDoctypeToken("html", &pub, &sys, true)
	if got.Type != DOCTYPE || got.Name != "html" || got.PublicID != &pub || got.SystemID != &sys || !got.ForceQuirks {
		t.Fatalf("NewDoctypeToken = %#v, want fields set", got)
	}
}
