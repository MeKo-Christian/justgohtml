package constants

// Character classification lookup tables for fast tokenizer hot path.
// These tables provide O(1) character classification for ASCII characters,
// avoiding switch statements with multiple case values.

// isWhitespace provides fast lookup for HTML whitespace characters.
// HTML5 whitespace: U+0009 TAB, U+000A LF, U+000C FF, U+0020 SPACE
// Per §13.2.6.4.1 of WHATWG HTML5 spec.
var isWhitespace [256]bool

// isASCIIUpper provides fast lookup for uppercase ASCII letters (A-Z).
var isASCIIUpper [256]bool

// isASCIILower provides fast lookup for lowercase ASCII letters (a-z).
var isASCIILower [256]bool

// isASCIIAlpha provides fast lookup for ASCII letters (A-Z, a-z).
var isASCIIAlpha [256]bool

// isASCIIAlphaNum provides fast lookup for ASCII alphanumeric characters (0-9, A-Z, a-z).
var isASCIIAlphaNum [256]bool

func init() {
	// HTML5 whitespace characters
	isWhitespace['\t'] = true // U+0009 TAB
	isWhitespace['\n'] = true // U+000A LF
	isWhitespace['\f'] = true // U+000C FF
	isWhitespace[' '] = true  // U+0020 SPACE

	// Uppercase ASCII letters (A-Z)
	for c := 'A'; c <= 'Z'; c++ {
		isASCIIUpper[c] = true
		isASCIIAlpha[c] = true
		isASCIIAlphaNum[c] = true
	}

	// Lowercase ASCII letters (a-z)
	for c := 'a'; c <= 'z'; c++ {
		isASCIILower[c] = true
		isASCIIAlpha[c] = true
		isASCIIAlphaNum[c] = true
	}

	// ASCII digits (0-9)
	for c := '0'; c <= '9'; c++ {
		isASCIIAlphaNum[c] = true
	}
}

// IsWhitespace returns true if c is an HTML5 whitespace character.
// Fast path for ASCII, correct handling for all Unicode values.
func IsWhitespace(c rune) bool {
	if c < 256 {
		return isWhitespace[c]
	}
	return false
}

// IsASCIIUpper returns true if c is an uppercase ASCII letter (A-Z).
func IsASCIIUpper(c rune) bool {
	if c < 256 {
		return isASCIIUpper[c]
	}
	return false
}

// IsASCIILower returns true if c is a lowercase ASCII letter (a-z).
func IsASCIILower(c rune) bool {
	if c < 256 {
		return isASCIILower[c]
	}
	return false
}

// IsASCIIAlpha returns true if c is an ASCII letter (A-Z or a-z).
func IsASCIIAlpha(c rune) bool {
	if c < 256 {
		return isASCIIAlpha[c]
	}
	return false
}

// IsASCIIAlphaNum returns true if c is an ASCII alphanumeric character (0-9, A-Z, a-z).
func IsASCIIAlphaNum(c rune) bool {
	if c < 256 {
		return isASCIIAlphaNum[c]
	}
	return false
}

// ToLower converts an ASCII uppercase letter to lowercase.
// For non-ASCII or non-uppercase, returns the character unchanged.
// This is faster than unicode.ToLower for ASCII characters.
func ToLower(c rune) rune {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
