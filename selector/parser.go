package selector

import (
	"strings"
	"unicode"

	"github.com/MeKo-Christian/JustGoHTML/errors"
)

// tokenType represents the type of a lexical token.
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenTag
	tokenID
	tokenClass
	tokenUniversal
	tokenAttrStart  // [
	tokenAttrEnd    // ]
	tokenAttrOp     // =, ~=, |=, ^=, $=, *=
	tokenString     // "value" or 'value' or unquoted
	tokenCombinator // >, +, ~, or whitespace (descendant)
	tokenComma      // ,
	tokenColon      // :
	tokenParenOpen  // (
	tokenParenClose // )
)

// token represents a lexical token.
type token struct {
	typ   tokenType
	value string
}

// tokenizer scans a CSS selector string into tokens.
type tokenizer struct {
	input           string
	pos             int
	length          int
	selectorStr     string
	inAttr          bool // inside attribute selector
	afterAttrName   bool // after attribute name, expecting operator or ]
	afterAttrOp     bool // after attribute operator, expecting value
	afterAttrValue  bool // after attribute value, expecting ]
	inPseudoArgs    bool // inside pseudo-class arguments
	parenDepth      int  // track nested parentheses
	afterSimpleSel  bool // after a simple selector (tag, id, class, etc.)
	afterCombinator bool // after an explicit combinator
}

func newTokenizer(input string) *tokenizer {
	return &tokenizer{
		input:       input,
		pos:         0,
		length:      len(input),
		selectorStr: input,
	}
}

func (t *tokenizer) peek() rune {
	if t.pos >= t.length {
		return 0
	}
	// Handle multi-byte runes properly
	for _, r := range t.input[t.pos:] {
		return r
	}
	return 0
}

func (t *tokenizer) advance() rune {
	if t.pos >= t.length {
		return 0
	}
	r := rune(t.input[t.pos])
	// Handle multi-byte runes
	for _, ch := range t.input[t.pos:] {
		r = ch
		break
	}
	t.pos += len(string(r))
	return r
}

func (t *tokenizer) skipWhitespace() bool {
	hadWS := false
	for t.pos < t.length {
		ch := t.peek()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f' {
			t.advance()
			hadWS = true
		} else {
			break
		}
	}
	return hadWS
}

func (t *tokenizer) isNameStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '-' || ch > 127
}

func (t *tokenizer) isNameChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch > 127
}

func (t *tokenizer) readName() string {
	start := t.pos
	for t.pos < t.length {
		ch := t.peek()
		switch {
		case t.isNameChar(ch):
			t.advance()
		case ch == '\\':
			// Handle escape sequences
			t.advance()
			if t.pos < t.length {
				t.advance()
			}
		default:
			return t.input[start:t.pos]
		}
	}
	return t.input[start:t.pos]
}

func (t *tokenizer) readString(quote rune) (string, error) {
	var sb strings.Builder
	t.advance() // consume opening quote
	for t.pos < t.length {
		ch := t.advance()
		if ch == quote {
			return sb.String(), nil
		}
		if ch == '\\' {
			if t.pos < t.length {
				escaped := t.advance()
				sb.WriteRune(escaped)
			}
		} else {
			sb.WriteRune(ch)
		}
	}
	return "", &errors.SelectorError{
		Selector: t.selectorStr,
		Position: t.pos,
		Message:  "unclosed string",
	}
}

func (t *tokenizer) readUnquotedAttrValue() string {
	var sb strings.Builder
	for t.pos < t.length {
		ch := t.peek()
		if ch == ']' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			break
		}
		if ch == '\\' {
			t.advance()
			if t.pos < t.length {
				sb.WriteRune(t.advance())
			}
		} else {
			sb.WriteRune(t.advance())
		}
	}
	return sb.String()
}

//nolint:gocognit,gocritic,gocyclo,nestif,cyclop,funlen,maintidx // tokenize is a state machine with inherently high complexity
func (t *tokenizer) tokenize() ([]token, error) {
	var tokens []token

	for t.pos < t.length {
		// Handle whitespace
		hadWS := t.skipWhitespace()
		if t.pos >= t.length {
			break
		}

		ch := t.peek()

		// Whitespace becomes descendant combinator only when:
		// - We're after a simple selector
		// - Not after an explicit combinator
		// - Not inside attribute selector or pseudo-args
		// - Next char is start of a new selector (not comma, ], ), or combinator)
		if hadWS && t.afterSimpleSel && !t.afterCombinator && !t.inAttr && !t.inPseudoArgs {
			if ch != ',' && ch != ']' && ch != ')' && ch != '>' && ch != '+' && ch != '~' {
				tokens = append(tokens, token{typ: tokenCombinator, value: " "})
				t.afterCombinator = true
				t.afterSimpleSel = false
			}
		}

		ch = t.peek()

		switch ch {
		case '*':
			if t.inAttr && !t.afterAttrOp {
				// Inside attribute selector, this could be *=
				t.advance()
				if t.peek() == '=' {
					t.advance()
					tokens = append(tokens, token{typ: tokenAttrOp, value: "*="})
					t.afterAttrOp = true
				} else {
					return nil, &errors.SelectorError{
						Selector: t.selectorStr,
						Position: t.pos,
						Message:  "expected = after * in attribute selector",
					}
				}
			} else {
				t.advance()
				tokens = append(tokens, token{typ: tokenUniversal, value: "*"})
				t.afterSimpleSel = true
				t.afterCombinator = false
			}

		case '#':
			t.advance()
			name := t.readName()
			if name == "" {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "expected identifier after #",
				}
			}
			tokens = append(tokens, token{typ: tokenID, value: name})
			t.afterSimpleSel = true
			t.afterCombinator = false

		case '.':
			t.advance()
			name := t.readName()
			if name == "" {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "expected identifier after .",
				}
			}
			tokens = append(tokens, token{typ: tokenClass, value: name})
			t.afterSimpleSel = true
			t.afterCombinator = false

		case '[':
			t.advance()
			tokens = append(tokens, token{typ: tokenAttrStart, value: "["})
			t.inAttr = true
			t.afterAttrName = false
			t.afterAttrOp = false
			t.afterAttrValue = false
			t.afterSimpleSel = false
			t.afterCombinator = false

		case ']':
			t.advance()
			tokens = append(tokens, token{typ: tokenAttrEnd, value: "]"})
			t.inAttr = false
			t.afterAttrName = false
			t.afterAttrOp = false
			t.afterAttrValue = false
			t.afterSimpleSel = true
			t.afterCombinator = false

		case ':':
			t.advance()
			name := t.readName()
			if name == "" {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "expected pseudo-class name after :",
				}
			}
			tokens = append(tokens, token{typ: tokenColon, value: name})
			t.afterSimpleSel = true
			t.afterCombinator = false

		case '(':
			t.advance()
			tokens = append(tokens, token{typ: tokenParenOpen, value: "("})
			t.inPseudoArgs = true
			t.parenDepth++
			t.afterSimpleSel = false
			t.afterCombinator = false

		case ')':
			t.advance()
			tokens = append(tokens, token{typ: tokenParenClose, value: ")"})
			t.parenDepth--
			if t.parenDepth <= 0 {
				t.inPseudoArgs = false
				t.parenDepth = 0
			}
			t.afterSimpleSel = true
			t.afterCombinator = false

		case ',':
			t.advance()
			tokens = append(tokens, token{typ: tokenComma, value: ","})
			t.afterSimpleSel = false
			t.afterCombinator = false

		case '>':
			t.advance()
			tokens = append(tokens, token{typ: tokenCombinator, value: ">"})
			t.afterCombinator = true
			t.afterSimpleSel = false

		case '+':
			t.advance()
			tokens = append(tokens, token{typ: tokenCombinator, value: "+"})
			t.afterCombinator = true
			t.afterSimpleSel = false

		case '~':
			if t.inAttr && !t.afterAttrOp {
				// Inside attribute selector, this could be ~=
				t.advance()
				if t.peek() == '=' {
					t.advance()
					tokens = append(tokens, token{typ: tokenAttrOp, value: "~="})
					t.afterAttrOp = true
				} else {
					return nil, &errors.SelectorError{
						Selector: t.selectorStr,
						Position: t.pos,
						Message:  "unexpected ~ in attribute selector",
					}
				}
			} else {
				t.advance()
				tokens = append(tokens, token{typ: tokenCombinator, value: "~"})
				t.afterCombinator = true
				t.afterSimpleSel = false
			}

		case '=':
			if t.inAttr {
				t.advance()
				tokens = append(tokens, token{typ: tokenAttrOp, value: "="})
				t.afterAttrOp = true
			} else {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "unexpected = outside attribute selector",
				}
			}

		case '^':
			if t.inAttr {
				t.advance()
				if t.peek() == '=' {
					t.advance()
					tokens = append(tokens, token{typ: tokenAttrOp, value: "^="})
					t.afterAttrOp = true
				} else {
					return nil, &errors.SelectorError{
						Selector: t.selectorStr,
						Position: t.pos,
						Message:  "expected = after ^",
					}
				}
			} else {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "unexpected ^ outside attribute selector",
				}
			}

		case '$':
			if t.inAttr {
				t.advance()
				if t.peek() == '=' {
					t.advance()
					tokens = append(tokens, token{typ: tokenAttrOp, value: "$="})
					t.afterAttrOp = true
				} else {
					return nil, &errors.SelectorError{
						Selector: t.selectorStr,
						Position: t.pos,
						Message:  "expected = after $",
					}
				}
			} else {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "unexpected $ outside attribute selector",
				}
			}

		case '|':
			if t.inAttr {
				t.advance()
				if t.peek() == '=' {
					t.advance()
					tokens = append(tokens, token{typ: tokenAttrOp, value: "|="})
					t.afterAttrOp = true
				} else {
					return nil, &errors.SelectorError{
						Selector: t.selectorStr,
						Position: t.pos,
						Message:  "expected = after |",
					}
				}
			} else {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "unexpected | outside attribute selector",
				}
			}

		case '"', '\'':
			str, err := t.readString(ch)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{typ: tokenString, value: str})
			if t.inAttr {
				t.afterAttrValue = true
			}

		default:
			if t.inAttr && t.afterAttrOp && !t.afterAttrValue {
				// Read unquoted attribute value
				val := t.readUnquotedAttrValue()
				if val != "" {
					tokens = append(tokens, token{typ: tokenString, value: val})
					t.afterAttrValue = true
				}
			} else if t.isNameStart(ch) || (t.inAttr && !t.afterAttrName) {
				name := t.readName()
				if name != "" {
					if t.inAttr && !t.afterAttrName {
						// Attribute name
						tokens = append(tokens, token{typ: tokenTag, value: name})
						t.afterAttrName = true
					} else if t.inPseudoArgs {
						// Pseudo-class argument (like "odd", "even", or selector for :not)
						tokens = append(tokens, token{typ: tokenString, value: name})
					} else {
						// Tag name
						tokens = append(tokens, token{typ: tokenTag, value: strings.ToLower(name)})
						t.afterSimpleSel = true
						t.afterCombinator = false
					}
				}
			} else if t.inPseudoArgs && (unicode.IsDigit(ch) || ch == '-' || ch == 'n') {
				// Read An+B expression or number
				var sb strings.Builder
				for t.pos < t.length {
					c := t.peek()
					if unicode.IsDigit(c) || c == 'n' || c == '+' || c == '-' {
						sb.WriteRune(t.advance())
					} else {
						break
					}
				}
				if sb.Len() > 0 {
					tokens = append(tokens, token{typ: tokenString, value: sb.String()})
				}
			} else {
				return nil, &errors.SelectorError{
					Selector: t.selectorStr,
					Position: t.pos,
					Message:  "unexpected character: " + string(ch),
				}
			}
		}
	}

	tokens = append(tokens, token{typ: tokenEOF})
	return tokens, nil
}

// parser builds AST from tokens.
type parser struct {
	tokens      []token
	pos         int
	selectorStr string
}

func newParser(tokens []token, selectorStr string) *parser {
	return &parser{
		tokens:      tokens,
		pos:         0,
		selectorStr: selectorStr,
	}
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *parser) parse() (selectorAST, error) {
	sel, err := p.parseComplexSelector()
	if err != nil {
		return nil, err
	}

	// Check for comma-separated selector list
	if p.peek().typ == tokenComma {
		list := SelectorList{Selectors: []ComplexSelector{*sel}}
		for p.peek().typ == tokenComma {
			p.advance() // consume comma
			next, err := p.parseComplexSelector()
			if err != nil {
				return nil, err
			}
			list.Selectors = append(list.Selectors, *next)
		}
		if p.peek().typ != tokenEOF {
			return nil, &errors.SelectorError{
				Selector: p.selectorStr,
				Position: p.pos,
				Message:  "unexpected token after selector list",
			}
		}
		return list, nil
	}

	if p.peek().typ != tokenEOF {
		return nil, &errors.SelectorError{
			Selector: p.selectorStr,
			Position: p.pos,
			Message:  "unexpected token: " + p.peek().value,
		}
	}

	return *sel, nil
}

func (p *parser) parseComplexSelector() (*ComplexSelector, error) {
	compound, err := p.parseCompoundSelector()
	if err != nil {
		return nil, err
	}

	sel := &ComplexSelector{
		Parts: []ComplexPart{{
			Combinator: CombinatorNone,
			Compound:   *compound,
		}},
	}

	for {
		tok := p.peek()
		if tok.typ != tokenCombinator {
			break
		}

		p.advance()
		var comb Combinator
		switch tok.value {
		case " ":
			comb = CombinatorDescendant
		case ">":
			comb = CombinatorChild
		case "+":
			comb = CombinatorAdjacent
		case "~":
			comb = CombinatorGeneral
		}

		compound, err := p.parseCompoundSelector()
		if err != nil {
			return nil, err
		}

		sel.Parts = append(sel.Parts, ComplexPart{
			Combinator: comb,
			Compound:   *compound,
		})
	}

	return sel, nil
}

func (p *parser) parseCompoundSelector() (*CompoundSelector, error) {
	compound := &CompoundSelector{}

	for {
		tok := p.peek()

		switch tok.typ {
		case tokenTag:
			p.advance()
			compound.Selectors = append(compound.Selectors, SimpleSelector{
				Kind: KindTag,
				Name: tok.value,
			})

		case tokenUniversal:
			p.advance()
			compound.Selectors = append(compound.Selectors, SimpleSelector{
				Kind: KindUniversal,
				Name: "*",
			})

		case tokenID:
			p.advance()
			compound.Selectors = append(compound.Selectors, SimpleSelector{
				Kind: KindID,
				Name: tok.value,
			})

		case tokenClass:
			p.advance()
			compound.Selectors = append(compound.Selectors, SimpleSelector{
				Kind: KindClass,
				Name: tok.value,
			})

		case tokenAttrStart:
			sel, err := p.parseAttributeSelector()
			if err != nil {
				return nil, err
			}
			compound.Selectors = append(compound.Selectors, *sel)

		case tokenColon:
			sel := p.parsePseudoSelector()
			compound.Selectors = append(compound.Selectors, *sel)

		case tokenEOF, tokenAttrEnd, tokenAttrOp, tokenString, tokenCombinator, tokenComma, tokenParenOpen, tokenParenClose:
			// These tokens end a compound selector
			if len(compound.Selectors) == 0 {
				return nil, &errors.SelectorError{
					Selector: p.selectorStr,
					Position: p.pos,
					Message:  "expected selector",
				}
			}
			return compound, nil
		}
	}
}

func (p *parser) parseAttributeSelector() (*SimpleSelector, error) {
	p.advance() // consume [

	nameTok := p.peek()
	if nameTok.typ != tokenTag {
		return nil, &errors.SelectorError{
			Selector: p.selectorStr,
			Position: p.pos,
			Message:  "expected attribute name",
		}
	}
	p.advance()

	sel := &SimpleSelector{
		Kind:     KindAttr,
		Name:     nameTok.value,
		Operator: AttrExists,
	}

	// Check for operator
	opTok := p.peek()
	if opTok.typ == tokenAttrOp {
		p.advance()
		switch opTok.value {
		case "=":
			sel.Operator = AttrEquals
		case "~=":
			sel.Operator = AttrIncludes
		case "|=":
			sel.Operator = AttrDashPrefix
		case "^=":
			sel.Operator = AttrPrefixMatch
		case "$=":
			sel.Operator = AttrSuffixMatch
		case "*=":
			sel.Operator = AttrSubstring
		}

		// Read value
		valTok := p.peek()
		if valTok.typ == tokenString {
			p.advance()
			sel.Value = valTok.value
		} else {
			return nil, &errors.SelectorError{
				Selector: p.selectorStr,
				Position: p.pos,
				Message:  "expected attribute value",
			}
		}
	}

	// Expect ]
	if p.peek().typ != tokenAttrEnd {
		return nil, &errors.SelectorError{
			Selector: p.selectorStr,
			Position: p.pos,
			Message:  "expected ]",
		}
	}
	p.advance()

	return sel, nil
}

func (p *parser) parsePseudoSelector() *SimpleSelector {
	nameTok := p.advance() // tokenColon already has the name

	sel := &SimpleSelector{
		Kind: KindPseudo,
		Name: nameTok.value,
	}

	// Check for functional pseudo-class arguments
	if p.peek().typ == tokenParenOpen {
		p.advance() // consume (

		// Read everything until matching )
		var args strings.Builder
		depth := 1
		for depth > 0 && p.peek().typ != tokenEOF {
			tok := p.advance()
			// Reconstruct the original selector syntax from tokens
			switch tok.typ {
			case tokenParenOpen:
				depth++
				args.WriteString("(")
			case tokenParenClose:
				depth--
				if depth > 0 {
					args.WriteString(")")
				}
			case tokenID:
				args.WriteString("#")
				args.WriteString(tok.value)
			case tokenClass:
				args.WriteString(".")
				args.WriteString(tok.value)
			case tokenUniversal:
				args.WriteString("*")
			case tokenColon:
				args.WriteString(":")
				args.WriteString(tok.value)
			case tokenAttrStart:
				args.WriteString("[")
			case tokenAttrEnd:
				args.WriteString("]")
			case tokenAttrOp:
				args.WriteString(tok.value)
			case tokenCombinator:
				args.WriteString(tok.value)
			case tokenComma:
				args.WriteString(",")
			case tokenEOF, tokenTag, tokenString:
				args.WriteString(tok.value)
			}
		}
		sel.Value = strings.TrimSpace(args.String())
	}

	return sel
}
