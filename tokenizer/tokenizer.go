package tokenizer

import (
	"strings"
	"sync"
	"unicode"

	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
)

// attrMapPool pools attribute index maps to reduce allocations.
var attrMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]struct{}, 8) // Pre-allocate for typical attribute count
	},
}

// getAttrMap retrieves a map from the pool and clears it.
func getAttrMap() map[string]struct{} {
	m := attrMapPool.Get().(map[string]struct{})
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// putAttrMap returns a map to the pool.
func putAttrMap(m map[string]struct{}) {
	if m != nil {
		attrMapPool.Put(m)
	}
}

// Tokenizer implements the HTML5 tokenization algorithm (port of the Python reference).
//
// It produces a stream of tokens and collects parse errors.
type Tokenizer struct {
	opts Options

	origInput string

	buf []rune
	pos int

	state    State
	textMode State

	reconsume bool
	ignoreLF  bool

	line   int
	column int

	// Current tag token being built.
	currentTagKind        TokenKind
	currentTagName        []rune
	currentTagAttrs       []Attr
	currentTagAttrIndex   map[string]struct{}
	currentTagSelfClosing bool

	currentAttrName           []rune
	currentAttrValue          []rune
	currentAttrValueHasAmp    bool
	currentComment            []rune
	commentEOF                bool
	currentDoctypeName        []rune
	currentDoctypePublic      *[]rune // nil = not set, empty slice = empty string
	currentDoctypeSystem      *[]rune
	currentDoctypeForceQuirks bool

	// For rawtext/rcdata/script end-tag matching.
	rawtextTagName  string
	originalTagName []rune
	tempBuffer      []rune

	lastStartTagName string

	textBuffer strings.Builder
	textHasAmp bool

	pendingTokens []Token
	errors        []ParseError

	allowCDATA bool
}

// ParseError represents a tokenizer parse error.
type ParseError struct {
	Code    string
	Message string
	Line    int
	Column  int
}

// New creates a new tokenizer for the given input.
func New(input string) *Tokenizer {
	return NewWithOptions(input, defaultOptions())
}

// NewWithOptions creates a new tokenizer for the given input and options.
func NewWithOptions(input string, opts Options) *Tokenizer {
	t := &Tokenizer{
		opts:     opts,
		state:    DataState,
		textMode: DataState,
		line:     1,
		column:   0,
	}
	t.origInput = input
	t.reset(input)
	return t
}

func (t *Tokenizer) reset(input string) {
	if input != "" && t.opts.DiscardBOM {
		r := []rune(input)
		if len(r) > 0 && r[0] == 0xFEFF {
			r = r[1:]
		}
		t.buf = r
	} else {
		t.buf = []rune(input)
	}

	t.pos = 0
	t.reconsume = false
	t.ignoreLF = false
	t.line = 1
	t.column = 0
	t.textMode = t.state

	t.currentTagKind = StartTag
	t.currentTagName = t.currentTagName[:0]
	t.currentTagAttrs = t.currentTagAttrs[:0]
	// Return old map to pool and get a fresh one
	putAttrMap(t.currentTagAttrIndex)
	t.currentTagAttrIndex = getAttrMap()
	t.currentTagSelfClosing = false
	t.currentAttrName = t.currentAttrName[:0]
	t.currentAttrValue = t.currentAttrValue[:0]
	t.currentAttrValueHasAmp = false
	t.currentComment = t.currentComment[:0]
	t.currentDoctypeName = t.currentDoctypeName[:0]
	t.currentDoctypePublic = nil
	t.currentDoctypeSystem = nil
	t.currentDoctypeForceQuirks = false

	t.rawtextTagName = ""
	t.originalTagName = t.originalTagName[:0]
	t.tempBuffer = t.tempBuffer[:0]

	t.textBuffer.Reset()
	t.textHasAmp = false

	t.pendingTokens = nil
	t.errors = nil
}

// SetDiscardBOM controls whether the leading U+FEFF BOM is discarded.
// For correctness, this should be called before consuming tokens.
func (t *Tokenizer) SetDiscardBOM(discard bool) {
	if t.opts.DiscardBOM == discard {
		return
	}
	t.opts.DiscardBOM = discard
	// Re-initialize the input buffer since BOM handling affects the rune stream.
	t.reset(t.origInput)
}

// SetXMLCoercion enables/disables XML coercion for text/comment output.
func (t *Tokenizer) SetXMLCoercion(enabled bool) {
	t.opts.XMLCoercion = enabled
}

// SetAllowCDATA toggles CDATA section parsing for foreign content.
func (t *Tokenizer) SetAllowCDATA(enabled bool) {
	t.allowCDATA = enabled
}

// SetState sets the tokenizer state.
// This is used by the tree builder to switch to RCDATA, RAWTEXT, etc.
func (t *Tokenizer) SetState(state State) {
	t.state = state
	//nolint:exhaustive // Only specific states affect textMode; others use default behavior
	switch state {
	case DataState, RCDATAState, RAWTEXTState, ScriptDataState, PLAINTEXTState, CDATASectionState:
		t.textMode = state
	default:
		// Other states do not change textMode
	}
	// Ensure rawtext end-tag matching has a tag name.
	if (state == RCDATAState || state == RAWTEXTState || state == ScriptDataState) && t.rawtextTagName == "" && t.lastStartTagName != "" {
		t.rawtextTagName = t.lastStartTagName
	}
}

// SetLastStartTag sets the last start tag name.
// This is used for appropriate end tag matching in RCDATA/RAWTEXT/script states.
func (t *Tokenizer) SetLastStartTag(name string) {
	t.lastStartTagName = name
	// For tokenizer tests, we use this as the current rawtext tag name as well.
	t.rawtextTagName = name
}

// Errors returns the parse errors encountered during tokenization.
func (t *Tokenizer) Errors() []ParseError {
	return t.errors
}

// Next returns the next token.
// Returns a token with Type == EOF when input is exhausted.
func (t *Tokenizer) Next() Token {
	if len(t.pendingTokens) > 0 {
		token := t.pendingTokens[0]
		t.pendingTokens = t.pendingTokens[1:]
		return token
	}

	for len(t.pendingTokens) == 0 {
		t.step()
	}
	token := t.pendingTokens[0]
	t.pendingTokens = t.pendingTokens[1:]
	return token
}

//nolint:gocyclo,exhaustive // HTML5 tokenizer state machine dispatcher - complexity mandated by spec
func (t *Tokenizer) step() {
	//nolint:exhaustive // Only active tokenizer states dispatched; others indicate implementation errors
	switch t.state {
	case DataState:
		t.stateData()
	case TagOpenState:
		t.stateTagOpen()
	case EndTagOpenState:
		t.stateEndTagOpen()
	case TagNameState:
		t.stateTagName()
	case BeforeAttributeNameState:
		t.stateBeforeAttributeName()
	case AttributeNameState:
		t.stateAttributeName()
	case AfterAttributeNameState:
		t.stateAfterAttributeName()
	case BeforeAttributeValueState:
		t.stateBeforeAttributeValue()
	case AttributeValueDoubleQuotedState:
		t.stateAttributeValueDoubleQuoted()
	case AttributeValueSingleQuotedState:
		t.stateAttributeValueSingleQuoted()
	case AttributeValueUnquotedState:
		t.stateAttributeValueUnquoted()
	case AfterAttributeValueQuotedState:
		t.stateAfterAttributeValueQuoted()
	case SelfClosingStartTagState:
		t.stateSelfClosingStartTag()
	case MarkupDeclarationOpenState:
		t.stateMarkupDeclarationOpen()
	case CommentStartState:
		t.stateCommentStart()
	case CommentStartDashState:
		t.stateCommentStartDash()
	case CommentState:
		t.stateComment()
	case CommentEndDashState:
		t.stateCommentEndDash()
	case CommentEndState:
		t.stateCommentEnd()
	case CommentEndBangState:
		t.stateCommentEndBang()
	case BogusCommentState:
		t.stateBogusComment()
	case DOCTYPEState:
		t.stateDoctype()
	case BeforeDOCTYPENameState:
		t.stateBeforeDoctypeName()
	case DOCTYPENameState:
		t.stateDoctypeName()
	case AfterDOCTYPENameState:
		t.stateAfterDoctypeName()
	case BogusDOCTYPEState:
		t.stateBogusDoctype()
	case AfterDOCTYPEPublicKeywordState:
		t.stateAfterDoctypePublicKeyword()
	case AfterDOCTYPESystemKeywordState:
		t.stateAfterDoctypeSystemKeyword()
	case BeforeDOCTYPEPublicIdentifierState:
		t.stateBeforeDoctypePublicIdentifier()
	case DOCTYPEPublicIdentifierDoubleQuotedState:
		t.stateDoctypePublicIdentifierDoubleQuoted()
	case DOCTYPEPublicIdentifierSingleQuotedState:
		t.stateDoctypePublicIdentifierSingleQuoted()
	case AfterDOCTYPEPublicIdentifierState:
		t.stateAfterDoctypePublicIdentifier()
	case BetweenDOCTYPEPublicAndSystemIdentifiersState:
		t.stateBetweenDoctypePublicAndSystemIdentifiers()
	case BeforeDOCTYPESystemIdentifierState:
		t.stateBeforeDoctypeSystemIdentifier()
	case DOCTYPESystemIdentifierDoubleQuotedState:
		t.stateDoctypeSystemIdentifierDoubleQuoted()
	case DOCTYPESystemIdentifierSingleQuotedState:
		t.stateDoctypeSystemIdentifierSingleQuoted()
	case AfterDOCTYPESystemIdentifierState:
		t.stateAfterDoctypeSystemIdentifier()
	case CDATASectionState:
		t.stateCDATASection()
	case CDATASectionBracketState:
		t.stateCDATASectionBracket()
	case CDATASectionEndState:
		t.stateCDATASectionEnd()
	case RCDATAState:
		t.stateRCDATA()
	case RCDATALessThanSignState:
		t.stateRCDATALessThanSign()
	case RCDATAEndTagOpenState:
		t.stateRCDATAEndTagOpen()
	case RCDATAEndTagNameState:
		t.stateRCDATAEndTagName()
	case RAWTEXTState:
		t.stateRAWTEXT()
	case ScriptDataState:
		t.stateRAWTEXT() // Script data behaves like rawtext with extra handling.
	case RAWTEXTLessThanSignState:
		t.stateRAWTEXTLessThanSign()
	case RAWTEXTEndTagOpenState:
		t.stateRAWTEXTEndTagOpen()
	case RAWTEXTEndTagNameState:
		t.stateRAWTEXTEndTagName()
	case PLAINTEXTState:
		t.statePLAINTEXT()
	case ScriptDataEscapedState:
		t.stateScriptDataEscaped()
	case ScriptDataEscapedDashState:
		t.stateScriptDataEscapedDash()
	case ScriptDataEscapedDashDashState:
		t.stateScriptDataEscapedDashDash()
	case ScriptDataEscapedLessThanSignState:
		t.stateScriptDataEscapedLessThanSign()
	case ScriptDataEscapedEndTagOpenState:
		t.stateScriptDataEscapedEndTagOpen()
	case ScriptDataEscapedEndTagNameState:
		t.stateScriptDataEscapedEndTagName()
	case ScriptDataDoubleEscapeStartState:
		t.stateScriptDataDoubleEscapeStart()
	case ScriptDataDoubleEscapedState:
		t.stateScriptDataDoubleEscaped()
	case ScriptDataDoubleEscapedDashState:
		t.stateScriptDataDoubleEscapedDash()
	case ScriptDataDoubleEscapedDashDashState:
		t.stateScriptDataDoubleEscapedDashDash()
	case ScriptDataDoubleEscapedLessThanSignState:
		t.stateScriptDataDoubleEscapedLessThanSign()
	case ScriptDataDoubleEscapeEndState:
		t.stateScriptDataDoubleEscapeEnd()
	default:
		// Unimplemented states behave like Data for now.
		t.state = DataState
	}
}

func (t *Tokenizer) getChar() (rune, bool) {
	if t.reconsume {
		t.reconsume = false
		if t.pos == 0 {
			return 0, false
		}
		t.pos--
	}

	for {
		if t.pos >= len(t.buf) {
			return 0, false
		}

		c := t.buf[t.pos]
		t.pos++

		if c == '\r' {
			t.ignoreLF = true
			t.advance('\n')
			return '\n', true
		}
		if c == '\n' {
			if t.ignoreLF {
				t.ignoreLF = false
				continue
			}
			t.advance('\n')
			return '\n', true
		}

		t.ignoreLF = false
		t.advance(c)
		return c, true
	}
}

func (t *Tokenizer) peek(offset int) (rune, bool) {
	i := t.pos + offset
	if t.reconsume {
		i--
	}
	if i < 0 || i >= len(t.buf) {
		return 0, false
	}
	return t.buf[i], true
}

func (t *Tokenizer) advance(c rune) {
	if c == '\n' {
		t.line++
		t.column = 0
		return
	}
	t.column++
}

func (t *Tokenizer) emit(tok Token) {
	t.pendingTokens = append(t.pendingTokens, tok)
}

func (t *Tokenizer) emitEOF() {
	t.flushText()
	t.emit(Token{Type: EOF})
}

func (t *Tokenizer) emitError(code string) {
	t.errors = append(t.errors, ParseError{
		Code:   code,
		Line:   t.line,
		Column: max(1, t.column),
	})
}

func (t *Tokenizer) reconsumeCurrent() {
	t.reconsume = true
}

func (t *Tokenizer) appendTextRune(r rune) {
	if r == '&' {
		t.textHasAmp = true
	}
	t.textBuffer.WriteRune(r)
}

func (t *Tokenizer) flushText() {
	if t.textBuffer.Len() == 0 {
		return
	}
	data := t.textBuffer.String()
	t.textBuffer.Reset()

	// Decode character references in Data/RCDATA modes (including their helper states).
	if (t.textMode == DataState || t.textMode == RCDATAState) && t.textHasAmp {
		data = decodeEntitiesInText(data, false)
	}
	t.textHasAmp = false

	if t.opts.XMLCoercion {
		data = coerceTextForXML(data)
	}

	t.emit(Token{Type: Character, Data: data})
}

func (t *Tokenizer) finishAttribute() {
	if len(t.currentAttrName) == 0 {
		return
	}
	name := constants.InternAttributeName(string(t.currentAttrName))
	t.currentAttrName = t.currentAttrName[:0]

	if _, exists := t.currentTagAttrIndex[name]; exists {
		t.emitError("duplicate-attribute")
		t.currentAttrValue = t.currentAttrValue[:0]
		t.currentAttrValueHasAmp = false
		return
	}

	value := ""
	if len(t.currentAttrValue) > 0 {
		value = string(t.currentAttrValue)
	}
	if t.currentAttrValueHasAmp {
		value = decodeEntitiesInText(value, true)
	}
	t.currentTagAttrs = append(t.currentTagAttrs, Attr{Name: name, Value: value})
	t.currentTagAttrIndex[name] = struct{}{}

	t.currentAttrValue = t.currentAttrValue[:0]
	t.currentAttrValueHasAmp = false
}

func (t *Tokenizer) emitCurrentTag() bool {
	var switchedTextMode bool
	name := constants.InternTagName(string(t.currentTagName))
	attrs := append([]Attr(nil), t.currentTagAttrs...)
	tok := Token{
		Type:        t.currentTagKind,
		Name:        name,
		Attrs:       attrs,
		SelfClosing: t.currentTagSelfClosing,
	}

	// Tokenizer-side state switching for rawtext/rcdata elements.
	// In the full HTML parsing pipeline, the tree builder controls these switches.
	// The reference Python implementation performs this switch when emitting the
	// tag into the sink; tokenizer tests in this repo expect the same behavior.
	if tok.Type == StartTag {
		t.lastStartTagName = name
		switch name {
		case "title", "textarea":
			t.state = RCDATAState
			t.textMode = RCDATAState
			t.rawtextTagName = name
			switchedTextMode = true
		case "script":
			t.state = ScriptDataState
			t.textMode = RAWTEXTState
			t.rawtextTagName = name
			switchedTextMode = true
		case "style", "xmp", "iframe", "noembed", "noframes":
			t.state = RAWTEXTState
			t.textMode = RAWTEXTState
			t.rawtextTagName = name
			switchedTextMode = true
		case "plaintext":
			t.state = PLAINTEXTState
			t.textMode = PLAINTEXTState
			t.rawtextTagName = name
			switchedTextMode = true
		}
	}

	t.currentTagName = t.currentTagName[:0]
	t.currentTagAttrs = t.currentTagAttrs[:0]
	// Return old map to pool and get a fresh one
	putAttrMap(t.currentTagAttrIndex)
	t.currentTagAttrIndex = getAttrMap()
	t.currentAttrName = t.currentAttrName[:0]
	t.currentAttrValue = t.currentAttrValue[:0]
	t.currentAttrValueHasAmp = false
	t.currentTagSelfClosing = false
	t.currentTagKind = StartTag

	t.emit(tok)
	return switchedTextMode
}

func (t *Tokenizer) emitComment() {
	data := string(t.currentComment)
	t.currentComment = t.currentComment[:0]
	if t.opts.XMLCoercion {
		data = coerceCommentForXML(data)
	}
	t.emit(Token{Type: Comment, Data: data, CommentEOF: t.commentEOF})
	t.commentEOF = false
}

func (t *Tokenizer) emitDoctype() {
	name := string(t.currentDoctypeName)
	var publicID *string
	var systemID *string
	if t.currentDoctypePublic != nil {
		s := string(*t.currentDoctypePublic)
		publicID = &s
	}
	if t.currentDoctypeSystem != nil {
		s := string(*t.currentDoctypeSystem)
		systemID = &s
	}

	t.emit(Token{
		Type:        DOCTYPE,
		Name:        name,
		PublicID:    publicID,
		SystemID:    systemID,
		ForceQuirks: t.currentDoctypeForceQuirks,
	})
}

func (t *Tokenizer) consumeIf(lit string) bool {
	r := []rune(lit)
	if t.pos+len(r) > len(t.buf) {
		return false
	}
	for i := range r {
		if t.buf[t.pos+i] != r[i] {
			return false
		}
	}
	t.pos += len(r)
	// Update column as if consumed (best-effort; these literals are ASCII).
	t.column += len(r)
	return true
}

func (t *Tokenizer) consumeCaseInsensitive(lit string) bool {
	r := []rune(lit)
	if t.pos+len(r) > len(t.buf) {
		return false
	}
	for i := range r {
		a := t.buf[t.pos+i]
		b := r[i]
		if unicode.ToLower(a) != unicode.ToLower(b) {
			return false
		}
	}
	t.pos += len(r)
	t.column += len(r)
	return true
}

func (t *Tokenizer) stateData() {
	t.textMode = DataState
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitEOF()
			return
		}
		switch c {
		case '<':
			t.flushText()
			t.state = TagOpenState
			return
		case 0:
			t.emitError("unexpected-null-character")
			// The Python reference emits the error but keeps U+0000 in the output.
			t.appendTextRune(0)
		default:
			t.appendTextRune(c)
		}
	}
}

func (t *Tokenizer) startTag(kind TokenKind, first rune) {
	t.currentTagKind = kind
	t.currentTagName = t.currentTagName[:0]
	t.currentTagAttrs = t.currentTagAttrs[:0]
	// Return old map to pool and get a fresh one
	putAttrMap(t.currentTagAttrIndex)
	t.currentTagAttrIndex = getAttrMap()
	t.currentAttrName = t.currentAttrName[:0]
	t.currentAttrValue = t.currentAttrValue[:0]
	t.currentAttrValueHasAmp = false
	t.currentTagSelfClosing = false

	if first >= 'A' && first <= 'Z' {
		first += 32
	}
	t.currentTagName = append(t.currentTagName, first)
}

func (t *Tokenizer) stateTagOpen() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-before-tag-name")
		t.appendTextRune('<')
		t.emitEOF()
		return
	}
	switch c {
	case '!':
		t.state = MarkupDeclarationOpenState
	case '/':
		t.state = EndTagOpenState
	case '?':
		t.emitError("unexpected-question-mark-instead-of-tag-name")
		t.currentComment = t.currentComment[:0]
		t.reconsumeCurrent()
		t.state = BogusCommentState
	default:
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			t.startTag(StartTag, c)
			t.state = TagNameState
			return
		}
		t.emitError("invalid-first-character-of-tag-name")
		t.appendTextRune('<')
		t.reconsumeCurrent()
		t.state = DataState
	}
}

func (t *Tokenizer) stateEndTagOpen() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-before-tag-name")
		t.appendTextRune('<')
		t.appendTextRune('/')
		t.emitEOF()
		return
	}
	if c == '>' {
		t.emitError("empty-end-tag")
		t.state = DataState
		return
	}
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
		t.startTag(EndTag, c)
		t.state = TagNameState
		return
	}
	t.emitError("invalid-first-character-of-tag-name")
	t.currentComment = t.currentComment[:0]
	t.reconsumeCurrent()
	t.state = BogusCommentState
}

func (t *Tokenizer) stateTagName() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}

		switch c {
		case '\t', '\n', '\f', ' ':
			t.state = BeforeAttributeNameState
			return
		case '/':
			t.state = SelfClosingStartTagState
			return
		case '>':
			t.finishAttribute()
			if !t.emitCurrentTag() {
				t.state = DataState
			}
			return
		case 0:
			t.emitError("unexpected-null-character")
			t.currentTagName = append(t.currentTagName, unicode.ReplacementChar)
		default:
			if c >= 'A' && c <= 'Z' {
				c += 32
			}
			t.currentTagName = append(t.currentTagName, c)
		}
	}
}

func (t *Tokenizer) stateBeforeAttributeName() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '/':
			t.finishAttribute()
			t.state = SelfClosingStartTagState
			return
		case '>':
			t.finishAttribute()
			if !t.emitCurrentTag() {
				t.state = DataState
			}
			return
		default:
			t.finishAttribute()
			t.currentAttrName = t.currentAttrName[:0]
			t.currentAttrValue = t.currentAttrValue[:0]
			t.currentAttrValueHasAmp = false
			switch {
			case c == 0:
				t.emitError("unexpected-null-character")
				c = unicode.ReplacementChar
			case c >= 'A' && c <= 'Z':
				c += 32
			case c == '=':
				t.emitError("unexpected-equals-sign-before-attribute-name")
			}
			t.currentAttrName = append(t.currentAttrName, c)
			t.state = AttributeNameState
			return
		}
	}
}

func (t *Tokenizer) stateAttributeName() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			t.finishAttribute()
			t.state = AfterAttributeNameState
			return
		case '/':
			t.finishAttribute()
			t.state = SelfClosingStartTagState
			return
		case '=':
			t.state = BeforeAttributeValueState
			return
		case '>':
			t.finishAttribute()
			if !t.emitCurrentTag() {
				t.state = DataState
			}
			return
		case 0:
			t.emitError("unexpected-null-character")
			t.currentAttrName = append(t.currentAttrName, unicode.ReplacementChar)
		default:
			if c == '"' || c == '\'' || c == '<' {
				t.emitError("unexpected-character-in-attribute-name")
			}
			if c >= 'A' && c <= 'Z' {
				c += 32
			}
			t.currentAttrName = append(t.currentAttrName, c)
		}
	}
}

func (t *Tokenizer) stateAfterAttributeName() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '/':
			t.finishAttribute()
			t.state = SelfClosingStartTagState
			return
		case '=':
			t.state = BeforeAttributeValueState
			return
		case '>':
			t.finishAttribute()
			if !t.emitCurrentTag() {
				t.state = DataState
			}
			return
		default:
			t.finishAttribute()
			t.currentAttrName = t.currentAttrName[:0]
			t.currentAttrValue = t.currentAttrValue[:0]
			t.currentAttrValueHasAmp = false
			if c == 0 {
				t.emitError("unexpected-null-character")
				c = unicode.ReplacementChar
			} else if c >= 'A' && c <= 'Z' {
				c += 32
			}
			t.currentAttrName = append(t.currentAttrName, c)
			t.state = AttributeNameState
			return
		}
	}
}

func (t *Tokenizer) stateBeforeAttributeValue() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '"':
			t.state = AttributeValueDoubleQuotedState
			return
		case '\'':
			t.state = AttributeValueSingleQuotedState
			return
		case '>':
			t.emitError("missing-attribute-value")
			t.finishAttribute()
			if !t.emitCurrentTag() {
				t.state = DataState
			}
			return
		default:
			t.reconsumeCurrent()
			t.state = AttributeValueUnquotedState
			return
		}
	}
}

func (t *Tokenizer) stateAttributeValueDoubleQuoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}
		switch c {
		case '"':
			t.state = AfterAttributeValueQuotedState
			return
		case '&':
			t.currentAttrValueHasAmp = true
			t.currentAttrValue = append(t.currentAttrValue, '&')
		case 0:
			t.emitError("unexpected-null-character")
			t.currentAttrValue = append(t.currentAttrValue, unicode.ReplacementChar)
		default:
			t.currentAttrValue = append(t.currentAttrValue, c)
		}
	}
}

func (t *Tokenizer) stateAttributeValueSingleQuoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emitEOF()
			return
		}
		switch c {
		case '\'':
			t.state = AfterAttributeValueQuotedState
			return
		case '&':
			t.currentAttrValueHasAmp = true
			t.currentAttrValue = append(t.currentAttrValue, '&')
		case 0:
			t.emitError("unexpected-null-character")
			t.currentAttrValue = append(t.currentAttrValue, unicode.ReplacementChar)
		default:
			t.currentAttrValue = append(t.currentAttrValue, c)
		}
	}
}

func (t *Tokenizer) stateAttributeValueUnquoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-tag")
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			t.finishAttribute()
			t.state = BeforeAttributeNameState
			return
		case '>':
			t.finishAttribute()
			t.emitCurrentTag()
			t.state = DataState
			return
		case '&':
			t.currentAttrValueHasAmp = true
			t.currentAttrValue = append(t.currentAttrValue, '&')
		case 0:
			t.emitError("unexpected-null-character")
			t.currentAttrValue = append(t.currentAttrValue, unicode.ReplacementChar)
		default:
			if c == '"' || c == '\'' || c == '<' || c == '=' || c == '`' {
				t.emitError("unexpected-character-in-unquoted-attribute-value")
			}
			t.currentAttrValue = append(t.currentAttrValue, c)
		}
	}
}

func (t *Tokenizer) stateAfterAttributeValueQuoted() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-tag")
		t.emitEOF()
		return
	}
	switch c {
	case '\t', '\n', '\f', ' ':
		t.finishAttribute()
		t.state = BeforeAttributeNameState
	case '/':
		t.finishAttribute()
		t.state = SelfClosingStartTagState
	case '>':
		t.finishAttribute()
		if !t.emitCurrentTag() {
			t.state = DataState
		}
	default:
		t.emitError("missing-whitespace-between-attributes")
		t.finishAttribute()
		t.reconsumeCurrent()
		t.state = BeforeAttributeNameState
	}
}

func (t *Tokenizer) stateSelfClosingStartTag() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-tag")
		t.emitEOF()
		return
	}
	if c == '>' {
		t.currentTagSelfClosing = true
		if !t.emitCurrentTag() {
			t.state = DataState
		}
		return
	}
	t.emitError("unexpected-character-after-solidus-in-tag")
	t.reconsumeCurrent()
	t.state = BeforeAttributeNameState
}

func (t *Tokenizer) stateMarkupDeclarationOpen() {
	if t.consumeIf("--") {
		t.currentComment = t.currentComment[:0]
		t.state = CommentStartState
		return
	}
	if t.consumeCaseInsensitive("DOCTYPE") {
		t.currentDoctypeName = t.currentDoctypeName[:0]
		t.currentDoctypePublic = nil
		t.currentDoctypeSystem = nil
		t.currentDoctypeForceQuirks = false
		t.state = DOCTYPEState
		return
	}
	if t.consumeIf("[CDATA[") {
		if t.allowCDATA {
			t.state = CDATASectionState
		} else {
			t.emitError("cdata-in-html-content")
			t.currentComment = t.currentComment[:0]
			t.currentComment = append(t.currentComment, []rune("[CDATA[")...)
			t.state = BogusCommentState
		}
		return
	}

	t.emitError("incorrectly-opened-comment")
	t.currentComment = t.currentComment[:0]
	t.state = BogusCommentState
}

func (t *Tokenizer) stateCommentStart() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-comment")
		t.emitComment()
		t.emit(Token{Type: EOF})
		return
	}
	switch c {
	case '-':
		t.state = CommentStartDashState
	case '>':
		t.emitError("abrupt-closing-of-empty-comment")
		t.emitComment()
		t.state = DataState
	case 0:
		t.emitError("unexpected-null-character")
		t.currentComment = append(t.currentComment, unicode.ReplacementChar)
		t.state = CommentState
	default:
		t.currentComment = append(t.currentComment, c)
		t.state = CommentState
	}
}

func (t *Tokenizer) stateCommentStartDash() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-comment")
		t.emitComment()
		t.emit(Token{Type: EOF})
		return
	}
	switch c {
	case '-':
		t.state = CommentEndState
	case '>':
		t.emitError("abrupt-closing-of-empty-comment")
		t.emitComment()
		t.state = DataState
	case 0:
		t.emitError("unexpected-null-character")
		t.currentComment = append(t.currentComment, '-', unicode.ReplacementChar)
		t.state = CommentState
	default:
		t.currentComment = append(t.currentComment, '-', c)
		t.state = CommentState
	}
}

func (t *Tokenizer) stateComment() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-comment")
			t.emitComment()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '-' {
			t.state = CommentEndDashState
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			t.currentComment = append(t.currentComment, unicode.ReplacementChar)
			continue
		}
		t.currentComment = append(t.currentComment, c)
	}
}

func (t *Tokenizer) stateCommentEndDash() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-comment")
		t.emitComment()
		t.emit(Token{Type: EOF})
		return
	}
	switch c {
	case '-':
		t.state = CommentEndState
	case 0:
		t.emitError("unexpected-null-character")
		t.currentComment = append(t.currentComment, '-', unicode.ReplacementChar)
		t.state = CommentState
	default:
		t.currentComment = append(t.currentComment, '-', c)
		t.state = CommentState
	}
}

func (t *Tokenizer) stateCommentEnd() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-comment")
		t.emitComment()
		t.emit(Token{Type: EOF})
		return
	}
	switch c {
	case '>':
		t.emitComment()
		t.state = DataState
	case '!':
		t.state = CommentEndBangState
	case '-':
		t.currentComment = append(t.currentComment, '-')
	default:
		if c == 0 {
			t.emitError("unexpected-null-character")
			t.currentComment = append(t.currentComment, '-', '-', unicode.ReplacementChar)
		} else {
			t.emitError("incorrectly-closed-comment")
			t.currentComment = append(t.currentComment, '-', '-', c)
		}
		t.state = CommentState
	}
}

func (t *Tokenizer) stateCommentEndBang() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-comment")
		t.emitComment()
		t.emit(Token{Type: EOF})
		return
	}
	switch c {
	case '-':
		t.currentComment = append(t.currentComment, '-', '-', '!')
		t.state = CommentEndDashState
	case '>':
		t.emitError("incorrectly-closed-comment")
		t.emitComment()
		t.state = DataState
	case 0:
		t.emitError("unexpected-null-character")
		t.currentComment = append(t.currentComment, '-', '-', '!', unicode.ReplacementChar)
		t.state = CommentState
	default:
		t.currentComment = append(t.currentComment, '-', '-', '!', c)
		t.state = CommentState
	}
}

func (t *Tokenizer) stateBogusComment() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.commentEOF = true
			t.emitComment()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '>' {
			t.commentEOF = false
			t.emitComment()
			t.state = DataState
			return
		}
		if c == 0 {
			t.currentComment = append(t.currentComment, unicode.ReplacementChar)
			continue
		}
		t.currentComment = append(t.currentComment, c)
	}
}

func (t *Tokenizer) stateDoctype() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-doctype")
		t.currentDoctypeForceQuirks = true
		t.emitDoctype()
		t.emit(Token{Type: EOF})
		return
	}
	switch c {
	case '\t', '\n', '\f', ' ':
		t.state = BeforeDOCTYPENameState
	case '>':
		t.emitError("expected-doctype-name-but-got-right-bracket")
		t.currentDoctypeForceQuirks = true
		t.emitDoctype()
		t.state = DataState
	default:
		t.emitError("missing-whitespace-before-doctype-name")
		t.reconsumeCurrent()
		t.state = BeforeDOCTYPENameState
	}
}

func (t *Tokenizer) stateBeforeDoctypeName() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype-name")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '\t' || c == '\n' || c == '\f' || c == ' ' {
			continue
		}
		if c == '>' {
			t.emitError("expected-doctype-name-but-got-right-bracket")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		}
		if c >= 'A' && c <= 'Z' {
			c += 32
		} else if c == 0 {
			t.emitError("unexpected-null-character")
			c = unicode.ReplacementChar
		}
		t.currentDoctypeName = append(t.currentDoctypeName, c)
		t.state = DOCTYPENameState
		return
	}
}

func (t *Tokenizer) stateDoctypeName() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype-name")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			t.state = AfterDOCTYPENameState
			return
		case '>':
			t.emitDoctype()
			t.state = DataState
			return
		default:
			if c >= 'A' && c <= 'Z' {
				c += 32
			} else if c == 0 {
				t.emitError("unexpected-null-character")
				c = unicode.ReplacementChar
			}
			t.currentDoctypeName = append(t.currentDoctypeName, c)
		}
	}
}

func (t *Tokenizer) stateAfterDoctypeName() {
	if t.consumeCaseInsensitive("PUBLIC") {
		t.state = AfterDOCTYPEPublicKeywordState
		return
	}
	if t.consumeCaseInsensitive("SYSTEM") {
		t.state = AfterDOCTYPESystemKeywordState
		return
	}

	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '\t' || c == '\n' || c == '\f' || c == ' ' {
			continue
		}
		if c == '>' {
			t.emitDoctype()
			t.state = DataState
			return
		}
		t.emitError("missing-whitespace-after-doctype-name")
		t.currentDoctypeForceQuirks = true
		t.reconsumeCurrent()
		t.state = BogusDOCTYPEState
		return
	}
}

//nolint:dupl // stateAfterDoctypePublicKeyword and stateAfterDoctypeSystemKeyword follow same HTML5 spec pattern
func (t *Tokenizer) stateAfterDoctypePublicKeyword() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("missing-quote-before-doctype-public-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			t.state = BeforeDOCTYPEPublicIdentifierState
			return
		case '"':
			t.emitError("missing-whitespace-before-doctype-public-identifier")
			empty := []rune{}
			t.currentDoctypePublic = &empty
			t.state = DOCTYPEPublicIdentifierDoubleQuotedState
			return
		case '\'':
			t.emitError("missing-whitespace-before-doctype-public-identifier")
			empty := []rune{}
			t.currentDoctypePublic = &empty
			t.state = DOCTYPEPublicIdentifierSingleQuotedState
			return
		case '>':
			t.emitError("missing-doctype-public-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		default:
			t.emitError("unexpected-character-after-doctype-public-keyword")
			t.currentDoctypeForceQuirks = true
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

//nolint:dupl // stateAfterDoctypePublicKeyword and stateAfterDoctypeSystemKeyword follow same HTML5 spec pattern
func (t *Tokenizer) stateAfterDoctypeSystemKeyword() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("missing-quote-before-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			t.state = BeforeDOCTYPESystemIdentifierState
			return
		case '"':
			t.emitError("missing-whitespace-after-doctype-public-identifier")
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierDoubleQuotedState
			return
		case '\'':
			t.emitError("missing-whitespace-after-doctype-public-identifier")
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierSingleQuotedState
			return
		case '>':
			t.emitError("missing-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		default:
			t.emitError("unexpected-character-after-doctype-system-keyword")
			t.currentDoctypeForceQuirks = true
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

func (t *Tokenizer) stateBeforeDoctypePublicIdentifier() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '"':
			empty := []rune{}
			t.currentDoctypePublic = &empty
			t.state = DOCTYPEPublicIdentifierDoubleQuotedState
			return
		case '\'':
			empty := []rune{}
			t.currentDoctypePublic = &empty
			t.state = DOCTYPEPublicIdentifierSingleQuotedState
			return
		case '>':
			t.emitError("missing-doctype-public-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		default:
			t.emitError("missing-quote-before-doctype-public-identifier")
			t.currentDoctypeForceQuirks = true
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

func (t *Tokenizer) stateDoctypePublicIdentifierDoubleQuoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '"' {
			t.state = AfterDOCTYPEPublicIdentifierState
			return
		}
		if c == '>' {
			t.emitError("abrupt-doctype-public-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			c = unicode.ReplacementChar
		}
		*t.currentDoctypePublic = append(*t.currentDoctypePublic, c)
	}
}

func (t *Tokenizer) stateDoctypePublicIdentifierSingleQuoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '\'' {
			t.state = AfterDOCTYPEPublicIdentifierState
			return
		}
		if c == '>' {
			t.emitError("abrupt-doctype-public-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			c = unicode.ReplacementChar
		}
		*t.currentDoctypePublic = append(*t.currentDoctypePublic, c)
	}
}

func (t *Tokenizer) stateAfterDoctypePublicIdentifier() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			t.state = BetweenDOCTYPEPublicAndSystemIdentifiersState
			return
		case '>':
			t.emitDoctype()
			t.state = DataState
			return
		case '"':
			t.emitError("missing-whitespace-between-doctype-public-and-system-identifiers")
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierDoubleQuotedState
			return
		case '\'':
			t.emitError("missing-whitespace-between-doctype-public-and-system-identifiers")
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierSingleQuotedState
			return
		default:
			t.emitError("missing-quote-before-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

func (t *Tokenizer) stateBetweenDoctypePublicAndSystemIdentifiers() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '>':
			t.emitDoctype()
			t.state = DataState
			return
		case '"':
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierDoubleQuotedState
			return
		case '\'':
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierSingleQuotedState
			return
		default:
			t.emitError("missing-quote-before-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

func (t *Tokenizer) stateBeforeDoctypeSystemIdentifier() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '"':
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierDoubleQuotedState
			return
		case '\'':
			empty := []rune{}
			t.currentDoctypeSystem = &empty
			t.state = DOCTYPESystemIdentifierSingleQuotedState
			return
		case '>':
			t.emitError("missing-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		default:
			t.emitError("missing-quote-before-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

func (t *Tokenizer) stateDoctypeSystemIdentifierDoubleQuoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '"' {
			t.state = AfterDOCTYPESystemIdentifierState
			return
		}
		if c == '>' {
			t.emitError("abrupt-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			c = unicode.ReplacementChar
		}
		*t.currentDoctypeSystem = append(*t.currentDoctypeSystem, c)
	}
}

func (t *Tokenizer) stateDoctypeSystemIdentifierSingleQuoted() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '\'' {
			t.state = AfterDOCTYPESystemIdentifierState
			return
		}
		if c == '>' {
			t.emitError("abrupt-doctype-system-identifier")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.state = DataState
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			c = unicode.ReplacementChar
		}
		*t.currentDoctypeSystem = append(*t.currentDoctypeSystem, c)
	}
}

func (t *Tokenizer) stateAfterDoctypeSystemIdentifier() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitError("eof-in-doctype")
			t.currentDoctypeForceQuirks = true
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		switch c {
		case '\t', '\n', '\f', ' ':
			continue
		case '>':
			t.emitDoctype()
			t.state = DataState
			return
		default:
			t.emitError("unexpected-character-after-doctype-system-identifier")
			t.reconsumeCurrent()
			t.state = BogusDOCTYPEState
			return
		}
	}
}

func (t *Tokenizer) stateBogusDoctype() {
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitDoctype()
			t.emit(Token{Type: EOF})
			return
		}
		if c == '>' {
			t.emitDoctype()
			t.state = DataState
			return
		}
	}
}

func (t *Tokenizer) stateCDATASection() {
	t.textMode = CDATASectionState
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-cdata")
		t.emitEOF()
		return
	}
	if c == ']' {
		t.state = CDATASectionBracketState
		return
	}
	t.appendTextRune(c)
}

func (t *Tokenizer) stateCDATASectionBracket() {
	c, ok := t.getChar()
	if !ok {
		t.emitError("eof-in-cdata")
		t.appendTextRune(']')
		t.emitEOF()
		return
	}
	if c == ']' {
		t.state = CDATASectionEndState
		return
	}
	t.appendTextRune(']')
	t.reconsumeCurrent()
	t.state = CDATASectionState
}

func (t *Tokenizer) stateCDATASectionEnd() {
	c, ok := t.getChar()
	if ok && c == '>' {
		t.flushText()
		t.state = DataState
		return
	}
	t.appendTextRune(']')
	if !ok {
		t.appendTextRune(']')
		t.emitError("eof-in-cdata")
		t.emitEOF()
		return
	}
	if c == ']' {
		return
	}
	t.appendTextRune(']')
	t.reconsumeCurrent()
	t.state = CDATASectionState
}

func (t *Tokenizer) stateRCDATA() {
	t.textMode = RCDATAState
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitEOF()
			return
		}
		switch c {
		case '<':
			t.state = RCDATALessThanSignState
			return
		case 0:
			t.emitError("unexpected-null-character")
			t.appendTextRune(unicode.ReplacementChar)
		default:
			t.appendTextRune(c)
		}
	}
}

func (t *Tokenizer) stateRCDATALessThanSign() {
	c, ok := t.getChar()
	if ok && c == '/' {
		t.currentTagName = t.currentTagName[:0]
		t.originalTagName = t.originalTagName[:0]
		t.state = RCDATAEndTagOpenState
		return
	}
	t.appendTextRune('<')
	if ok {
		t.reconsumeCurrent()
	}
	t.state = RCDATAState
}

func (t *Tokenizer) stateRCDATAEndTagOpen() {
	c, ok := t.getChar()
	if ok && ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
		t.currentTagName = append(t.currentTagName, unicode.ToLower(c))
		t.originalTagName = append(t.originalTagName, c)
		t.state = RCDATAEndTagNameState
		return
	}
	t.appendTextRune('<')
	t.appendTextRune('/')
	if ok {
		t.reconsumeCurrent()
	}
	t.state = RCDATAState
}

//nolint:dupl // stateRCDATAEndTagName and stateRAWTEXTEndTagName follow same HTML5 spec pattern with different fallback states
func (t *Tokenizer) stateRCDATAEndTagName() {
	for {
		c, ok := t.getChar()
		if ok && ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			t.currentTagName = append(t.currentTagName, unicode.ToLower(c))
			t.originalTagName = append(t.originalTagName, c)
			continue
		}

		tagName := string(t.currentTagName)
		if tagName == t.rawtextTagName {
			if ok && c == '>' {
				t.flushText()
				t.emit(Token{Type: EndTag, Name: tagName})
				t.state = DataState
				t.rawtextTagName = ""
				t.currentTagName = t.currentTagName[:0]
				t.originalTagName = t.originalTagName[:0]
				return
			}
			if ok && (c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f') {
				t.flushText()
				t.currentTagKind = EndTag
				t.currentTagName = []rune(tagName)
				t.currentTagAttrs = t.currentTagAttrs[:0]
				putAttrMap(t.currentTagAttrIndex)
		t.currentTagAttrIndex = getAttrMap()
				t.state = BeforeAttributeNameState
				return
			}
			if ok && c == '/' {
				t.flushText()
				t.currentTagKind = EndTag
				t.currentTagName = []rune(tagName)
				t.currentTagAttrs = t.currentTagAttrs[:0]
				putAttrMap(t.currentTagAttrIndex)
		t.currentTagAttrIndex = getAttrMap()
				t.state = SelfClosingStartTagState
				return
			}
		}

		// Not a matching end tag.
		t.appendTextRune('<')
		t.appendTextRune('/')
		for _, r := range t.originalTagName {
			t.appendTextRune(r)
		}
		t.currentTagName = t.currentTagName[:0]
		t.originalTagName = t.originalTagName[:0]
		if ok {
			t.reconsumeCurrent()
		}
		t.state = RCDATAState
		return
	}
}

func (t *Tokenizer) stateRAWTEXT() {
	t.textMode = RAWTEXTState
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitEOF()
			return
		}
		if c == '<' {
			// Script special-cases for "<!--" starting escape.
			if t.rawtextTagName == "script" {
				n1, ok1 := t.peek(0)
				n2, ok2 := t.peek(1)
				n3, ok3 := t.peek(2)
				if ok1 && ok2 && ok3 && n1 == '!' && n2 == '-' && n3 == '-' {
					t.appendTextRune('<')
					t.appendTextRune('!')
					t.appendTextRune('-')
					t.appendTextRune('-')
					_, _ = t.getChar()
					_, _ = t.getChar()
					_, _ = t.getChar()
					t.state = ScriptDataEscapedState
					return
				}
			}
			t.state = RAWTEXTLessThanSignState
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			t.appendTextRune(unicode.ReplacementChar)
			continue
		}
		t.appendTextRune(c)
	}
}

func (t *Tokenizer) stateRAWTEXTLessThanSign() {
	c, ok := t.getChar()
	if ok && c == '/' {
		t.currentTagName = t.currentTagName[:0]
		t.originalTagName = t.originalTagName[:0]
		t.state = RAWTEXTEndTagOpenState
		return
	}
	t.appendTextRune('<')
	if ok {
		t.reconsumeCurrent()
	}
	if t.rawtextTagName == "script" {
		t.state = ScriptDataState
	} else {
		t.state = RAWTEXTState
	}
}

func (t *Tokenizer) stateRAWTEXTEndTagOpen() {
	c, ok := t.getChar()
	if ok && ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
		t.currentTagName = append(t.currentTagName, unicode.ToLower(c))
		t.originalTagName = append(t.originalTagName, c)
		t.state = RAWTEXTEndTagNameState
		return
	}
	t.appendTextRune('<')
	t.appendTextRune('/')
	if ok {
		t.reconsumeCurrent()
	}
	if t.rawtextTagName == "script" {
		t.state = ScriptDataState
	} else {
		t.state = RAWTEXTState
	}
}

//nolint:dupl // stateRCDATAEndTagName and stateRAWTEXTEndTagName follow same HTML5 spec pattern with different fallback states
func (t *Tokenizer) stateRAWTEXTEndTagName() {
	for {
		c, ok := t.getChar()
		if ok && ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			t.currentTagName = append(t.currentTagName, unicode.ToLower(c))
			t.originalTagName = append(t.originalTagName, c)
			continue
		}
		tagName := string(t.currentTagName)
		if tagName == t.rawtextTagName {
			if ok && c == '>' {
				t.flushText()
				t.emit(Token{Type: EndTag, Name: tagName})
				t.state = DataState
				t.rawtextTagName = ""
				t.currentTagName = t.currentTagName[:0]
				t.originalTagName = t.originalTagName[:0]
				return
			}
			if ok && (c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f') {
				t.flushText()
				t.currentTagKind = EndTag
				t.currentTagName = []rune(tagName)
				t.currentTagAttrs = t.currentTagAttrs[:0]
				putAttrMap(t.currentTagAttrIndex)
		t.currentTagAttrIndex = getAttrMap()
				t.state = BeforeAttributeNameState
				return
			}
			if ok && c == '/' {
				t.flushText()
				t.currentTagKind = EndTag
				t.currentTagName = []rune(tagName)
				t.currentTagAttrs = t.currentTagAttrs[:0]
				putAttrMap(t.currentTagAttrIndex)
		t.currentTagAttrIndex = getAttrMap()
				t.state = SelfClosingStartTagState
				return
			}
		}

		// Not a matching end tag.
		t.appendTextRune('<')
		t.appendTextRune('/')
		for _, r := range t.originalTagName {
			t.appendTextRune(r)
		}
		t.currentTagName = t.currentTagName[:0]
		t.originalTagName = t.originalTagName[:0]
		if !ok {
			t.emitEOF()
			return
		}
		t.reconsumeCurrent()
		if t.rawtextTagName == "script" {
			t.state = ScriptDataState
		} else {
			t.state = RAWTEXTState
		}
		return
	}
}

func (t *Tokenizer) statePLAINTEXT() {
	t.textMode = PLAINTEXTState
	for {
		c, ok := t.getChar()
		if !ok {
			t.emitEOF()
			return
		}
		if c == 0 {
			t.emitError("unexpected-null-character")
			t.appendTextRune(unicode.ReplacementChar)
			continue
		}
		t.appendTextRune(c)
	}
}

func (t *Tokenizer) stateScriptDataEscaped() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	switch c {
	case '-':
		t.appendTextRune('-')
		t.state = ScriptDataEscapedDashState
	case '<':
		t.state = ScriptDataEscapedLessThanSignState
	case 0:
		t.emitError("unexpected-null-character")
		t.appendTextRune(unicode.ReplacementChar)
	default:
		t.appendTextRune(c)
	}
}

func (t *Tokenizer) stateScriptDataEscapedDash() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	switch c {
	case '-':
		t.appendTextRune('-')
		t.state = ScriptDataEscapedDashDashState
	case '<':
		t.state = ScriptDataEscapedLessThanSignState
	case 0:
		t.emitError("unexpected-null-character")
		t.appendTextRune(unicode.ReplacementChar)
		t.state = ScriptDataEscapedState
	default:
		t.appendTextRune(c)
		t.state = ScriptDataEscapedState
	}
}

func (t *Tokenizer) stateScriptDataEscapedDashDash() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	switch c {
	case '-':
		t.appendTextRune('-')
	case '<':
		t.appendTextRune('<')
		t.state = ScriptDataEscapedLessThanSignState
	case '>':
		t.appendTextRune('>')
		t.state = ScriptDataState
	case 0:
		t.emitError("unexpected-null-character")
		t.appendTextRune(unicode.ReplacementChar)
		t.state = ScriptDataEscapedState
	default:
		t.appendTextRune(c)
		t.state = ScriptDataEscapedState
	}
}

func (t *Tokenizer) stateScriptDataEscapedLessThanSign() {
	c, ok := t.getChar()
	if ok && c == '/' {
		t.tempBuffer = t.tempBuffer[:0]
		t.state = ScriptDataEscapedEndTagOpenState
		return
	}
	if ok && unicode.IsLetter(c) {
		t.tempBuffer = t.tempBuffer[:0]
		t.appendTextRune('<')
		t.appendTextRune(c)
		t.tempBuffer = append(t.tempBuffer, unicode.ToLower(c))
		t.state = ScriptDataDoubleEscapeStartState
		return
	}
	t.appendTextRune('<')
	if ok {
		t.reconsumeCurrent()
	}
	t.state = ScriptDataEscapedState
}

func (t *Tokenizer) stateScriptDataEscapedEndTagOpen() {
	c, ok := t.getChar()
	if ok && unicode.IsLetter(c) {
		t.currentTagName = t.currentTagName[:0]
		t.originalTagName = t.originalTagName[:0]
		t.currentTagName = append(t.currentTagName, unicode.ToLower(c))
		t.originalTagName = append(t.originalTagName, c)
		t.state = ScriptDataEscapedEndTagNameState
		return
	}
	t.appendTextRune('<')
	t.appendTextRune('/')
	if ok {
		t.reconsumeCurrent()
	}
	t.state = ScriptDataEscapedState
}

func (t *Tokenizer) stateScriptDataEscapedEndTagName() {
	for {
		c, ok := t.getChar()
		if ok && unicode.IsLetter(c) {
			t.currentTagName = append(t.currentTagName, unicode.ToLower(c))
			t.originalTagName = append(t.originalTagName, c)
			continue
		}
		tagName := string(t.currentTagName)
		if tagName == "script" {
			if ok && (c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f') {
				t.flushText()
				t.currentTagKind = EndTag
				t.currentTagName = []rune(tagName)
				t.currentTagAttrs = t.currentTagAttrs[:0]
				putAttrMap(t.currentTagAttrIndex)
		t.currentTagAttrIndex = getAttrMap()
				t.state = BeforeAttributeNameState
				return
			}
			if ok && c == '/' {
				t.flushText()
				t.currentTagKind = EndTag
				t.currentTagName = []rune(tagName)
				t.currentTagAttrs = t.currentTagAttrs[:0]
				putAttrMap(t.currentTagAttrIndex)
		t.currentTagAttrIndex = getAttrMap()
				t.state = SelfClosingStartTagState
				return
			}
			if ok && c == '>' {
				t.flushText()
				t.emit(Token{Type: EndTag, Name: tagName})
				t.state = DataState
				return
			}
		}

		t.appendTextRune('<')
		t.appendTextRune('/')
		for _, r := range t.originalTagName {
			t.appendTextRune(r)
		}
		t.currentTagName = t.currentTagName[:0]
		t.originalTagName = t.originalTagName[:0]
		if ok {
			t.reconsumeCurrent()
		}
		t.state = ScriptDataEscapedState
		return
	}
}

func (t *Tokenizer) stateScriptDataDoubleEscapeStart() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	if unicode.IsLetter(c) {
		t.tempBuffer = append(t.tempBuffer, unicode.ToLower(c))
		t.appendTextRune(c)
		return
	}

	temp := strings.ToLower(string(t.tempBuffer))
	if temp == "script" {
		if ok && (c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '/' || c == '>') {
			t.state = ScriptDataDoubleEscapedState
		} else {
			t.state = ScriptDataEscapedState
		}
	} else {
		t.state = ScriptDataEscapedState
	}
	if ok {
		t.reconsumeCurrent()
	}
}

func (t *Tokenizer) stateScriptDataDoubleEscaped() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	switch c {
	case '-':
		t.appendTextRune('-')
		t.state = ScriptDataDoubleEscapedDashState
	case '<':
		t.appendTextRune('<')
		t.state = ScriptDataDoubleEscapedLessThanSignState
	case 0:
		t.emitError("unexpected-null-character")
		t.appendTextRune(unicode.ReplacementChar)
	default:
		t.appendTextRune(c)
	}
}

func (t *Tokenizer) stateScriptDataDoubleEscapedDash() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	switch c {
	case '-':
		t.appendTextRune('-')
		t.state = ScriptDataDoubleEscapedDashDashState
	case '<':
		t.appendTextRune('<')
		t.state = ScriptDataDoubleEscapedLessThanSignState
	case 0:
		t.emitError("unexpected-null-character")
		t.appendTextRune(unicode.ReplacementChar)
		t.state = ScriptDataDoubleEscapedState
	default:
		t.appendTextRune(c)
		t.state = ScriptDataDoubleEscapedState
	}
}

func (t *Tokenizer) stateScriptDataDoubleEscapedDashDash() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	switch c {
	case '-':
		t.appendTextRune('-')
	case '<':
		t.appendTextRune('<')
		t.state = ScriptDataDoubleEscapedLessThanSignState
	case '>':
		t.appendTextRune('>')
		t.state = ScriptDataState
	case 0:
		t.emitError("unexpected-null-character")
		t.appendTextRune(unicode.ReplacementChar)
		t.state = ScriptDataDoubleEscapedState
	default:
		t.appendTextRune(c)
		t.state = ScriptDataDoubleEscapedState
	}
}

func (t *Tokenizer) stateScriptDataDoubleEscapedLessThanSign() {
	c, ok := t.getChar()
	if ok && c == '/' {
		t.tempBuffer = t.tempBuffer[:0]
		t.appendTextRune('/')
		t.state = ScriptDataDoubleEscapeEndState
		return
	}
	if ok {
		t.reconsumeCurrent()
	}
	t.state = ScriptDataDoubleEscapedState
}

func (t *Tokenizer) stateScriptDataDoubleEscapeEnd() {
	c, ok := t.getChar()
	if !ok {
		t.emitEOF()
		return
	}
	if unicode.IsLetter(c) {
		t.tempBuffer = append(t.tempBuffer, unicode.ToLower(c))
		t.appendTextRune(c)
		return
	}
	temp := strings.ToLower(string(t.tempBuffer))
	if temp == "script" {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '/' || c == '>' {
			t.state = ScriptDataEscapedState
		} else {
			t.state = ScriptDataDoubleEscapedState
		}
	} else {
		t.state = ScriptDataDoubleEscapedState
	}
	t.reconsumeCurrent()
}

func coerceTextForXML(text string) string {
	// Fast path for ASCII.
	isASCII := true
	for _, r := range text {
		if r > 0x7f {
			isASCII = false
			break
		}
	}
	if isASCII {
		return strings.ReplaceAll(text, "\f", " ")
	}

	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if r == '\f' {
			b.WriteRune(' ')
			continue
		}
		// U+FDD0..U+FDEF
		if r >= 0xFDD0 && r <= 0xFDEF {
			b.WriteRune(unicode.ReplacementChar)
			continue
		}
		// U+FFFE/U+FFFF in any plane.
		if r&0xFFFF == 0xFFFE || r&0xFFFF == 0xFFFF {
			b.WriteRune(unicode.ReplacementChar)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func coerceCommentForXML(text string) string {
	return strings.ReplaceAll(text, "--", "- -")
}
