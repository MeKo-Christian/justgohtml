package tokenizer

// State represents the tokenizer state.
// The tokenizer is a state machine that transitions between these states.
type State int

// InvalidState is used to indicate an unknown or invalid state.
const InvalidState State = -1

// State aliases for html5lib-tests compatibility.
const (
	PlaintextState = PLAINTEXTState
	RawtextState   = RAWTEXTState
)

// Tokenizer states as defined by the HTML5 specification.
// See: https://html.spec.whatwg.org/multipage/parsing.html#tokenization
const (
	DataState State = iota
	RCDATAState
	RAWTEXTState
	ScriptDataState
	PLAINTEXTState
	TagOpenState
	EndTagOpenState
	TagNameState
	RCDATALessThanSignState
	RCDATAEndTagOpenState
	RCDATAEndTagNameState
	RAWTEXTLessThanSignState
	RAWTEXTEndTagOpenState
	RAWTEXTEndTagNameState
	ScriptDataLessThanSignState
	ScriptDataEndTagOpenState
	ScriptDataEndTagNameState
	ScriptDataEscapeStartState
	ScriptDataEscapeStartDashState
	ScriptDataEscapedState
	ScriptDataEscapedDashState
	ScriptDataEscapedDashDashState
	ScriptDataEscapedLessThanSignState
	ScriptDataEscapedEndTagOpenState
	ScriptDataEscapedEndTagNameState
	ScriptDataDoubleEscapeStartState
	ScriptDataDoubleEscapedState
	ScriptDataDoubleEscapedDashState
	ScriptDataDoubleEscapedDashDashState
	ScriptDataDoubleEscapedLessThanSignState
	ScriptDataDoubleEscapeEndState
	BeforeAttributeNameState
	AttributeNameState
	AfterAttributeNameState
	BeforeAttributeValueState
	AttributeValueDoubleQuotedState
	AttributeValueSingleQuotedState
	AttributeValueUnquotedState
	AfterAttributeValueQuotedState
	SelfClosingStartTagState
	BogusCommentState
	MarkupDeclarationOpenState
	CommentStartState
	CommentStartDashState
	CommentState
	CommentLessThanSignState
	CommentLessThanSignBangState
	CommentLessThanSignBangDashState
	CommentLessThanSignBangDashDashState
	CommentEndDashState
	CommentEndState
	CommentEndBangState
	DOCTYPEState
	BeforeDOCTYPENameState
	DOCTYPENameState
	AfterDOCTYPENameState
	AfterDOCTYPEPublicKeywordState
	BeforeDOCTYPEPublicIdentifierState
	DOCTYPEPublicIdentifierDoubleQuotedState
	DOCTYPEPublicIdentifierSingleQuotedState
	AfterDOCTYPEPublicIdentifierState
	BetweenDOCTYPEPublicAndSystemIdentifiersState
	AfterDOCTYPESystemKeywordState
	BeforeDOCTYPESystemIdentifierState
	DOCTYPESystemIdentifierDoubleQuotedState
	DOCTYPESystemIdentifierSingleQuotedState
	AfterDOCTYPESystemIdentifierState
	BogusDOCTYPEState
	CDATASectionState
	CDATASectionBracketState
	CDATASectionEndState
	CharacterReferenceState
	NamedCharacterReferenceState
	AmbiguousAmpersandState
	NumericCharacterReferenceState
	HexadecimalCharacterReferenceStartState
	DecimalCharacterReferenceStartState
	HexadecimalCharacterReferenceState
	DecimalCharacterReferenceState
	NumericCharacterReferenceEndState
)

// String returns the name of the state for debugging.
func (s State) String() string {
	names := [...]string{
		"Data",
		"RCDATA",
		"RAWTEXT",
		"ScriptData",
		"PLAINTEXT",
		"TagOpen",
		"EndTagOpen",
		"TagName",
		"RCDATALessThanSign",
		"RCDATAEndTagOpen",
		"RCDATAEndTagName",
		"RAWTEXTLessThanSign",
		"RAWTEXTEndTagOpen",
		"RAWTEXTEndTagName",
		"ScriptDataLessThanSign",
		"ScriptDataEndTagOpen",
		"ScriptDataEndTagName",
		"ScriptDataEscapeStart",
		"ScriptDataEscapeStartDash",
		"ScriptDataEscaped",
		"ScriptDataEscapedDash",
		"ScriptDataEscapedDashDash",
		"ScriptDataEscapedLessThanSign",
		"ScriptDataEscapedEndTagOpen",
		"ScriptDataEscapedEndTagName",
		"ScriptDataDoubleEscapeStart",
		"ScriptDataDoubleEscaped",
		"ScriptDataDoubleEscapedDash",
		"ScriptDataDoubleEscapedDashDash",
		"ScriptDataDoubleEscapedLessThanSign",
		"ScriptDataDoubleEscapeEnd",
		"BeforeAttributeName",
		"AttributeName",
		"AfterAttributeName",
		"BeforeAttributeValue",
		"AttributeValueDoubleQuoted",
		"AttributeValueSingleQuoted",
		"AttributeValueUnquoted",
		"AfterAttributeValueQuoted",
		"SelfClosingStartTag",
		"BogusComment",
		"MarkupDeclarationOpen",
		"CommentStart",
		"CommentStartDash",
		"Comment",
		"CommentLessThanSign",
		"CommentLessThanSignBang",
		"CommentLessThanSignBangDash",
		"CommentLessThanSignBangDashDash",
		"CommentEndDash",
		"CommentEnd",
		"CommentEndBang",
		"DOCTYPE",
		"BeforeDOCTYPEName",
		"DOCTYPEName",
		"AfterDOCTYPEName",
		"AfterDOCTYPEPublicKeyword",
		"BeforeDOCTYPEPublicIdentifier",
		"DOCTYPEPublicIdentifierDoubleQuoted",
		"DOCTYPEPublicIdentifierSingleQuoted",
		"AfterDOCTYPEPublicIdentifier",
		"BetweenDOCTYPEPublicAndSystemIdentifiers",
		"AfterDOCTYPESystemKeyword",
		"BeforeDOCTYPESystemIdentifier",
		"DOCTYPESystemIdentifierDoubleQuoted",
		"DOCTYPESystemIdentifierSingleQuoted",
		"AfterDOCTYPESystemIdentifier",
		"BogusDOCTYPE",
		"CDATASection",
		"CDATASectionBracket",
		"CDATASectionEnd",
		"CharacterReference",
		"NamedCharacterReference",
		"AmbiguousAmpersand",
		"NumericCharacterReference",
		"HexadecimalCharacterReferenceStart",
		"DecimalCharacterReferenceStart",
		"HexadecimalCharacterReference",
		"DecimalCharacterReference",
		"NumericCharacterReferenceEnd",
	}
	if int(s) < len(names) {
		return names[s]
	}
	return "Unknown"
}
