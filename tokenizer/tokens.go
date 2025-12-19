// Package tokenizer implements the HTML5 tokenization algorithm.
package tokenizer

// TokenKind represents the type of a token.
//
// This is a "tag" for the Token sum type, matching the HTML5 tokenizer output
// tokens (StartTag, EndTag, Character, Comment, DOCTYPE, EOF).
type TokenKind int

// Token types produced by the tokenizer.
const (
	// Error indicates a parse error. The tokenizer records parse errors separately,
	// but this kind exists for test hooks and tooling.
	Error TokenKind = iota

	// DOCTYPE represents a DOCTYPE declaration.
	DOCTYPE

	// StartTag represents a start tag.
	StartTag

	// EndTag represents an end tag.
	EndTag

	// Comment represents a comment.
	Comment

	// Character represents character data.
	Character

	// EOF indicates end of input.
	EOF
)

// String returns the name of the token kind.
func (t TokenKind) String() string {
	names := [...]string{
		"Error",
		"DOCTYPE",
		"StartTag",
		"EndTag",
		"Comment",
		"Character",
		"EOF",
	}
	if t >= 0 && int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

// Tag represents the common data carried by start/end tag tokens.
type Tag struct {
	Name        string
	Attrs       []Attr
	SelfClosing bool
}

// CharacterToken represents a run of character data.
type CharacterToken struct {
	Data string
}

// CommentToken represents a comment token.
type CommentToken struct {
	Data string
}

// DoctypeToken represents a DOCTYPE token.
type DoctypeToken struct {
	Name        string
	PublicID    *string
	SystemID    *string
	ForceQuirks bool
}

// Token represents a token produced by the tokenizer.
//
// Token is a sum type (tagged union) keyed by Type.
type Token struct {
	Type TokenKind

	// Name is the tag name for StartTag/EndTag, or DOCTYPE name.
	Name string

	// Data is the content for Comment or Character tokens.
	Data string

	// Attrs holds attributes for StartTag tokens.
	Attrs []Attr

	// SelfClosing is true for self-closing tags (e.g., <br/>).
	SelfClosing bool

	// PublicID is the public identifier for DOCTYPE.
	PublicID *string

	// SystemID is the system identifier for DOCTYPE.
	SystemID *string

	// ForceQuirks is true if the DOCTYPE triggers quirks mode.
	ForceQuirks bool

	// ErrorCode is set for Error.
	ErrorCode string

	// CommentEOF indicates a bogus comment ended at EOF.
	CommentEOF bool
}

// Attr represents an HTML attribute.
type Attr struct {
	// Namespace for foreign attributes (usually empty for HTML).
	Namespace string

	// Name is the attribute name (lowercase for HTML).
	Name string

	// Value is the attribute value.
	Value string
}

// NewStartTagToken creates a StartTag token.
func NewStartTagToken(name string) Token {
	return Token{Type: StartTag, Name: name}
}

// NewEndTagToken creates an EndTag token.
func NewEndTagToken(name string) Token {
	return Token{Type: EndTag, Name: name}
}

// NewCharacterToken creates a Character token.
func NewCharacterToken(data string) Token {
	return Token{Type: Character, Data: data}
}

// NewCommentToken creates a Comment token.
func NewCommentToken(data string) Token {
	return Token{Type: Comment, Data: data}
}

// NewDoctypeToken creates a DOCTYPE token.
func NewDoctypeToken(name string, publicID, systemID *string, forceQuirks bool) Token {
	return Token{
		Type:        DOCTYPE,
		Name:        name,
		PublicID:    publicID,
		SystemID:    systemID,
		ForceQuirks: forceQuirks,
	}
}

// AttrVal returns the value of an attribute by name, or empty string if not found.
func (t *Token) AttrVal(name string) string {
	if len(t.Attrs) == 0 {
		return ""
	}
	for _, a := range t.Attrs {
		if a.Namespace == "" && a.Name == name {
			return a.Value
		}
	}
	return ""
}

// HasAttr returns true if the token has an attribute with the given name.
func (t *Token) HasAttr(name string) bool {
	if len(t.Attrs) == 0 {
		return false
	}
	for _, a := range t.Attrs {
		if a.Namespace == "" && a.Name == name {
			return true
		}
	}
	return false
}

func AttrsToMap(attrs []Attr) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, a := range attrs {
		if a.Namespace != "" {
			continue
		}
		out[a.Name] = a.Value
	}
	return out
}
