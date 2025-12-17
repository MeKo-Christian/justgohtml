package errors

// Error codes as defined by the WHATWG HTML5 specification.
// See: https://html.spec.whatwg.org/multipage/parsing.html#parse-errors
const (
	// Tokenizer errors
	AbruptClosingOfEmptyComment                               = "abrupt-closing-of-empty-comment"
	AbruptDoctypePublicIdentifier                             = "abrupt-doctype-public-identifier"
	AbruptDoctypeSystemIdentifier                             = "abrupt-doctype-system-identifier"
	AbsenceOfDigitsInNumericCharReference                     = "absence-of-digits-in-numeric-character-reference"
	CDATAInHTMLContent                                        = "cdata-in-html-content"
	CharacterReferenceOutsideUnicodeRange                     = "character-reference-outside-unicode-range"
	ControlCharacterInInputStream                             = "control-character-in-input-stream"
	ControlCharacterReference                                 = "control-character-reference"
	DuplicateAttribute                                        = "duplicate-attribute"
	EndTagWithAttributes                                      = "end-tag-with-attributes"
	EndTagWithTrailingSolidus                                 = "end-tag-with-trailing-solidus"
	EOFBeforeTagName                                          = "eof-before-tag-name"
	EOFInCDATA                                                = "eof-in-cdata"
	EOFInComment                                              = "eof-in-comment"
	EOFInDoctype                                              = "eof-in-doctype"
	EOFInScriptHTMLCommentLikeText                            = "eof-in-script-html-comment-like-text"
	EOFInTag                                                  = "eof-in-tag"
	IncorrectlyClosedComment                                  = "incorrectly-closed-comment"
	IncorrectlyOpenedComment                                  = "incorrectly-opened-comment"
	InvalidCharacterSequenceAfterDoctypeName                  = "invalid-character-sequence-after-doctype-name"
	InvalidFirstCharacterOfTagName                            = "invalid-first-character-of-tag-name"
	MissingAttributeValue                                     = "missing-attribute-value"
	MissingDoctypeName                                        = "missing-doctype-name"
	MissingDoctypePublicIdentifier                            = "missing-doctype-public-identifier"
	MissingDoctypeSystemIdentifier                            = "missing-doctype-system-identifier"
	MissingEndTagName                                         = "missing-end-tag-name"
	MissingQuoteBeforeDoctypePublicIdentifier                 = "missing-quote-before-doctype-public-identifier"
	MissingQuoteBeforeDoctypeSystemIdentifier                 = "missing-quote-before-doctype-system-identifier"
	MissingSemicolonAfterCharacterReference                   = "missing-semicolon-after-character-reference"
	MissingWhitespaceAfterDoctypePublicKeyword                = "missing-whitespace-after-doctype-public-keyword"
	MissingWhitespaceAfterDoctypeSystemKeyword                = "missing-whitespace-after-doctype-system-keyword"
	MissingWhitespaceBeforeDoctypeName                        = "missing-whitespace-before-doctype-name"
	MissingWhitespaceBetweenAttributes                        = "missing-whitespace-between-attributes"
	MissingWhitespaceBetweenDoctypePublicAndSystemIdentifiers = "missing-whitespace-between-doctype-public-and-system-identifiers"
	NestedComment                                             = "nested-comment"
	NoncharacterCharacterReference                            = "noncharacter-character-reference"
	NoncharacterInInputStream                                 = "noncharacter-in-input-stream"
	NonVoidHTMLElementStartTagWithTrailingSolidus             = "non-void-html-element-start-tag-with-trailing-solidus"
	NullCharacterReference                                    = "null-character-reference"
	SurrogateCharacterReference                               = "surrogate-character-reference"
	SurrogateInInputStream                                    = "surrogate-in-input-stream"
	UnexpectedCharacterAfterDoctypeSystemIdentifier           = "unexpected-character-after-doctype-system-identifier"
	UnexpectedCharacterInAttributeName                        = "unexpected-character-in-attribute-name"
	UnexpectedCharacterInUnquotedAttributeValue               = "unexpected-character-in-unquoted-attribute-value"
	UnexpectedEqualsSignBeforeAttributeName                   = "unexpected-equals-sign-before-attribute-name"
	UnexpectedNullCharacter                                   = "unexpected-null-character"
	UnexpectedQuestionMarkInsteadOfTagName                    = "unexpected-question-mark-instead-of-tag-name"
	UnexpectedSolidusInTag                                    = "unexpected-solidus-in-tag"
	UnknownNamedCharacterReference                            = "unknown-named-character-reference"

	// Tree construction errors
	NonSpaceCharacterInTableText = "non-space-character-in-table-text"
	FosterParentedCharacter      = "foster-parented-character"
)

// errorMessages maps error codes to human-readable messages.
var errorMessages = map[string]string{
	AbruptClosingOfEmptyComment:                               "This error occurs if the parser encounters an empty comment that is abruptly closed by a U+003E (>) code point.",
	AbruptDoctypePublicIdentifier:                             "This error occurs if the parser encounters a U+003E (>) code point in the DOCTYPE public identifier.",
	AbruptDoctypeSystemIdentifier:                             "This error occurs if the parser encounters a U+003E (>) code point in the DOCTYPE system identifier.",
	AbsenceOfDigitsInNumericCharReference:                     "This error occurs if the parser encounters a numeric character reference that doesn't contain any digits.",
	CDATAInHTMLContent:                                        "This error occurs if the parser encounters a CDATA section outside of foreign content (SVG or MathML).",
	CharacterReferenceOutsideUnicodeRange:                     "This error occurs if the parser encounters a numeric character reference that references a code point greater than U+10FFFF.",
	ControlCharacterInInputStream:                             "This error occurs if the input stream contains a control character other than ASCII whitespace or U+0000 NULL.",
	ControlCharacterReference:                                 "This error occurs if the parser encounters a numeric character reference that references a control character.",
	DuplicateAttribute:                                        "This error occurs if the parser encounters an attribute with the same name as an earlier attribute on the same tag.",
	EndTagWithAttributes:                                      "This error occurs if the parser encounters an end tag with attributes.",
	EndTagWithTrailingSolidus:                                 "This error occurs if the parser encounters an end tag with a trailing solidus (/).",
	EOFBeforeTagName:                                          "This error occurs if the parser encounters EOF where a tag name is expected.",
	EOFInCDATA:                                                "This error occurs if the parser encounters EOF in a CDATA section.",
	EOFInComment:                                              "This error occurs if the parser encounters EOF in a comment.",
	EOFInDoctype:                                              "This error occurs if the parser encounters EOF in a DOCTYPE.",
	EOFInScriptHTMLCommentLikeText:                            "This error occurs if the parser encounters EOF in a script element in an HTML comment-like text.",
	EOFInTag:                                                  "This error occurs if the parser encounters EOF in a tag.",
	IncorrectlyClosedComment:                                  "This error occurs if the parser encounters an incorrectly closed comment.",
	IncorrectlyOpenedComment:                                  "This error occurs if the parser encounters an incorrectly opened comment.",
	InvalidCharacterSequenceAfterDoctypeName:                  "This error occurs if the parser encounters an invalid character sequence after a DOCTYPE name.",
	InvalidFirstCharacterOfTagName:                            "This error occurs if the parser encounters an invalid first character of a tag name.",
	MissingAttributeValue:                                     "This error occurs if the parser encounters an attribute name not followed by an attribute value.",
	MissingDoctypeName:                                        "This error occurs if the parser encounters a DOCTYPE without a name.",
	MissingDoctypePublicIdentifier:                            "This error occurs if the parser encounters a DOCTYPE with a missing public identifier.",
	MissingDoctypeSystemIdentifier:                            "This error occurs if the parser encounters a DOCTYPE with a missing system identifier.",
	MissingEndTagName:                                         "This error occurs if the parser encounters a missing end tag name.",
	MissingQuoteBeforeDoctypePublicIdentifier:                 "This error occurs if the parser encounters a DOCTYPE public identifier without a leading quote.",
	MissingQuoteBeforeDoctypeSystemIdentifier:                 "This error occurs if the parser encounters a DOCTYPE system identifier without a leading quote.",
	MissingSemicolonAfterCharacterReference:                   "This error occurs if the parser encounters a character reference not terminated by a semicolon.",
	MissingWhitespaceAfterDoctypePublicKeyword:                "This error occurs if the parser encounters a DOCTYPE with missing whitespace after the PUBLIC keyword.",
	MissingWhitespaceAfterDoctypeSystemKeyword:                "This error occurs if the parser encounters a DOCTYPE with missing whitespace after the SYSTEM keyword.",
	MissingWhitespaceBeforeDoctypeName:                        "This error occurs if the parser encounters a DOCTYPE without whitespace before the name.",
	MissingWhitespaceBetweenAttributes:                        "This error occurs if the parser encounters a missing whitespace between attributes.",
	MissingWhitespaceBetweenDoctypePublicAndSystemIdentifiers: "This error occurs if the parser encounters a DOCTYPE with missing whitespace between public and system identifiers.",
	NestedComment:                                             "This error occurs if the parser encounters a nested comment.",
	NoncharacterCharacterReference:                            "This error occurs if the parser encounters a numeric character reference that references a noncharacter.",
	NoncharacterInInputStream:                                 "This error occurs if the input stream contains a noncharacter.",
	NonVoidHTMLElementStartTagWithTrailingSolidus:             "This error occurs if the parser encounters a non-void HTML element start tag with a trailing solidus.",
	NullCharacterReference:                                    "This error occurs if the parser encounters a numeric character reference that references U+0000 NULL.",
	SurrogateCharacterReference:                               "This error occurs if the parser encounters a numeric character reference that references a surrogate.",
	SurrogateInInputStream:                                    "This error occurs if the input stream contains a surrogate.",
	UnexpectedCharacterAfterDoctypeSystemIdentifier:           "This error occurs if the parser encounters an unexpected character after a DOCTYPE system identifier.",
	UnexpectedCharacterInAttributeName:                        "This error occurs if the parser encounters an unexpected character in an attribute name.",
	UnexpectedCharacterInUnquotedAttributeValue:               "This error occurs if the parser encounters an unexpected character in an unquoted attribute value.",
	UnexpectedEqualsSignBeforeAttributeName:                   "This error occurs if the parser encounters an equals sign before an attribute name.",
	UnexpectedNullCharacter:                                   "This error occurs if the parser encounters an unexpected null character.",
	UnexpectedQuestionMarkInsteadOfTagName:                    "This error occurs if the parser encounters a question mark instead of a tag name.",
	UnexpectedSolidusInTag:                                    "This error occurs if the parser encounters an unexpected solidus in a tag.",
	UnknownNamedCharacterReference:                            "This error occurs if the parser encounters an unknown named character reference.",
}

// Message returns the human-readable message for an error code.
func Message(code string) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}
