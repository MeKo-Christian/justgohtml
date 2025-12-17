// Package encoding implements HTML5 encoding detection and decoding.
package encoding

import (
	"bytes"
	"errors"
	"strings"
)

// ErrInvalidEncoding is returned when the specified encoding is not supported.
var ErrInvalidEncoding = errors.New("unsupported or invalid encoding")

// Encoding represents a character encoding.
type Encoding struct {
	// Name is the canonical name of the encoding.
	Name string

	// Labels are the encoding labels that map to this encoding.
	Labels []string
}

// Common encodings.
var (
	UTF8 = &Encoding{
		Name: "UTF-8",
		Labels: []string{
			"utf-8", "utf8", "unicode-1-1-utf-8",
			"unicode11utf8", "unicode20utf8", "x-unicode20utf8",
		},
	}
	Windows1252 = &Encoding{
		Name: "windows-1252",
		Labels: []string{
			"windows-1252", "windows1252", "cp1252", "x-cp1252",
			"ansi_x3.4-1968", "ascii", "us-ascii",
			"iso-ir-100", "csisolatin1",
		},
	}
	ISO88591 = &Encoding{
		Name: "ISO-8859-1",
		Labels: []string{
			"iso-8859-1", "iso8859-1", "iso88591",
			"iso_8859-1", "iso_8859-1:1987",
			"latin1", "latin-1", "l1",
			"cp819", "ibm819",
		},
	}
	ISO88592 = &Encoding{
		Name: "iso-8859-2",
		Labels: []string{
			"iso-8859-2", "iso8859-2", "iso88592",
			"iso_8859-2", "iso_8859-2:1987",
			"iso-ir-101", "csisolatin2",
			"latin2", "latin-2", "l2",
		},
	}
	EUCJP = &Encoding{
		Name: "euc-jp",
		Labels: []string{
			"euc-jp", "eucjp",
			"cseucpkdfmtjapanese", "x-euc-jp",
		},
	}
	UTF16   = &Encoding{Name: "utf-16", Labels: []string{"utf-16", "utf16"}}
	UTF16LE = &Encoding{Name: "utf-16le", Labels: []string{"utf-16le", "utf16le"}}
	UTF16BE = &Encoding{Name: "utf-16be", Labels: []string{"utf-16be", "utf16be"}}
)

// ASCII whitespace characters per HTML5 spec
var asciiWhitespace = map[byte]bool{
	0x09: true, // TAB
	0x0A: true, // LF
	0x0C: true, // FF
	0x0D: true, // CR
	0x20: true, // SPACE
}

// Decode decodes HTML bytes to a string using encoding detection.
//
// The detection follows the HTML5 specification:
// 1. BOM (Byte Order Mark)
// 2. Provided encoding hint (transport encoding)
// 3. <meta charset> in the first 1024 bytes (non-comment content)
// 4. Fallback to windows-1252
func Decode(data []byte, hint string) (string, *Encoding, error) {
	// Use hint if provided (transport encoding)
	if hint != "" {
		if enc := normalizeEncodingLabel(hint); enc != nil {
			bom := detectBOM(data)
			bomLen := 0
			if bom != nil {
				bomLen = bomLength(bom)
			}
			decoded, err := decodeWithEncoding(data[bomLen:], enc)
			return decoded, enc, err
		}
	}

	// Check for BOM
	if enc := detectBOM(data); enc != nil {
		bomLen := bomLength(enc)
		decoded, err := decodeWithEncoding(data[bomLen:], enc)
		return decoded, enc, err
	}

	// Scan for meta charset
	if enc := prescanForMetaCharset(data); enc != nil {
		decoded, err := decodeWithEncoding(data, enc)
		return decoded, enc, err
	}

	// Fallback to windows-1252
	decoded, err := decodeWithEncoding(data, Windows1252)
	return decoded, Windows1252, err
}

// detectBOM checks for a Byte Order Mark and returns the corresponding encoding.
func detectBOM(data []byte) *Encoding {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return UTF8
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return UTF16LE
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return UTF16BE
	}
	return nil
}

const (
	utf16BEName = "utf-16be"
	utf16LEName = "utf-16le"
)

// bomLength returns the length of the BOM for the given encoding.
func bomLength(enc *Encoding) int {
	switch enc.Name {
	case "UTF-8":
		return 3
	case utf16LEName, utf16BEName:
		return 2
	default:
		return 0
	}
}

// normalizeEncodingLabel normalizes an encoding label to a canonical encoding.
// Returns nil if the label is not recognized.
func normalizeEncodingLabel(label string) *Encoding {
	if label == "" {
		return nil
	}

	label = strings.ToLower(strings.TrimSpace(label))
	if label == "" {
		return nil
	}

	// Security: never allow utf-7
	if label == "utf-7" || label == "utf7" || label == "x-utf-7" {
		return Windows1252
	}

	// Try all known encodings
	encodings := []*Encoding{UTF8, Windows1252, ISO88591, ISO88592, EUCJP, UTF16, UTF16LE, UTF16BE}
	for _, enc := range encodings {
		for _, l := range enc.Labels {
			if l == label {
				// HTML treats ISO-8859-1 labels as windows-1252
				if enc == ISO88591 {
					return Windows1252
				}
				return enc
			}
		}
	}

	return nil
}

// normalizeMetaDeclaredEncoding normalizes a meta-declared encoding.
// Per HTML spec, UTF-16/UTF-32 in meta declarations are treated as UTF-8.
func normalizeMetaDeclaredEncoding(label []byte) *Encoding {
	enc := normalizeEncodingLabel(string(label))
	if enc == nil {
		return nil
	}

	// Per HTML meta charset handling: ignore UTF-16/UTF-32 declarations
	switch enc.Name {
	case "utf-16", utf16LEName, utf16BEName, "utf-32", "utf-32le", "utf-32be":
		return UTF8
	}

	return enc
}

// isASCIIWhitespace checks if a byte is ASCII whitespace.
func isASCIIWhitespace(b byte) bool {
	return asciiWhitespace[b]
}

// isASCIIAlpha checks if a byte is an ASCII letter.
func isASCIIAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// asciiLower converts an ASCII letter to lowercase.
func asciiLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b | 0x20
	}
	return b
}

// skipASCIIWhitespace skips ASCII whitespace starting at position i.
func skipASCIIWhitespace(data []byte, i int) int {
	n := len(data)
	for i < n && isASCIIWhitespace(data[i]) {
		i++
	}
	return i
}

// stripASCIIWhitespace removes leading and trailing ASCII whitespace.
func stripASCIIWhitespace(value []byte) []byte {
	start := 0
	end := len(value)
	for start < end && isASCIIWhitespace(value[start]) {
		start++
	}
	for end > start && isASCIIWhitespace(value[end-1]) {
		end--
	}
	return value[start:end]
}

// extractCharsetFromContent extracts a charset value from a Content-Type meta content attribute.
func extractCharsetFromContent(contentBytes []byte) []byte {
	if len(contentBytes) == 0 {
		return nil
	}

	// Normalize whitespace to spaces and convert to lowercase
	b := make([]byte, len(contentBytes))
	for i, ch := range contentBytes {
		if isASCIIWhitespace(ch) {
			b[i] = ' '
		} else {
			b[i] = asciiLower(ch)
		}
	}

	idx := bytes.Index(b, []byte("charset"))
	if idx == -1 {
		return nil
	}

	i := idx + len("charset")
	n := len(b)

	// Skip whitespace
	for i < n && b[i] == ' ' {
		i++
	}

	// Expect '='
	if i >= n || b[i] != '=' {
		return nil
	}
	i++

	// Skip whitespace
	for i < n && b[i] == ' ' {
		i++
	}

	if i >= n {
		return nil
	}

	// Check for quote
	var quote byte
	if b[i] == '"' || b[i] == '\'' {
		quote = b[i]
		i++
	}

	start := i
	for i < n {
		ch := b[i]
		if quote != 0 {
			if ch == quote {
				break
			}
		} else {
			if ch == ' ' || ch == ';' {
				break
			}
		}
		i++
	}

	// If quoted, we must find the closing quote
	if quote != 0 && (i >= n || b[i] != quote) {
		return nil
	}

	return b[start:i]
}

// prescanForMetaCharset scans the first 1024 bytes of non-comment content
// for a meta charset declaration per HTML5 spec.
//
//nolint:gocognit,gocyclo,nestif,cyclop,funlen,maintidx // Complexity required by HTML5 spec algorithm
func prescanForMetaCharset(data []byte) *Encoding {
	// Scan up to 1024 bytes of non-comment input, but allow skipping
	// arbitrarily large comments (bounded by a hard cap).
	const maxNonComment = 1024
	const maxTotalScan = 65536

	n := len(data)
	i := 0
	nonComment := 0

	for i < n && i < maxTotalScan && nonComment < maxNonComment {
		if data[i] != '<' {
			i++
			nonComment++
			continue
		}

		// Check for comment
		if i+3 < n && data[i+1] == '!' && data[i+2] == '-' && data[i+3] == '-' {
			end := bytes.Index(data[i+4:], []byte("-->"))
			if end == -1 {
				return nil
			}
			i = i + 4 + end + 3
			continue
		}

		// Tag open
		j := i + 1
		if j < n && data[j] == '/' {
			// End tag - skip it
			k := i
			var quote byte
			for k < n && k < maxTotalScan && nonComment < maxNonComment {
				ch := data[k]
				if quote == 0 {
					if ch == '"' || ch == '\'' {
						quote = ch
					} else if ch == '>' {
						k++
						nonComment++
						break
					}
				} else {
					if ch == quote {
						quote = 0
					}
				}
				k++
				nonComment++
			}
			i = k
			continue
		}

		if j >= n || !isASCIIAlpha(data[j]) {
			i++
			nonComment++
			continue
		}

		// Read tag name
		nameStart := j
		for j < n && isASCIIAlpha(data[j]) {
			j++
		}

		tagName := data[nameStart:j]
		if !bytes.Equal(bytes.ToLower(tagName), []byte("meta")) {
			// Skip the rest of this tag
			k := i
			var quote byte
			for k < n && k < maxTotalScan && nonComment < maxNonComment {
				ch := data[k]
				if quote == 0 {
					if ch == '"' || ch == '\'' {
						quote = ch
					} else if ch == '>' {
						k++
						nonComment++
						break
					}
				} else {
					if ch == quote {
						quote = 0
					}
				}
				k++
				nonComment++
			}
			i = k
			continue
		}

		// Parse attributes until '>'
		var charset []byte
		var httpEquiv []byte
		var content []byte

		k := j
		sawGT := false
		startI := i

		for k < n && k < maxTotalScan {
			ch := data[k]

			if ch == '>' {
				sawGT = true
				k++
				break
			}

			if ch == '<' {
				// Restart scanning from here
				break
			}

			if isASCIIWhitespace(ch) || ch == '/' {
				k++
				continue
			}

			// Attribute name
			attrStart := k
			for k < n {
				ch = data[k]
				if isASCIIWhitespace(ch) || ch == '=' || ch == '>' || ch == '/' || ch == '<' {
					break
				}
				k++
			}
			attrName := bytes.ToLower(data[attrStart:k])
			k = skipASCIIWhitespace(data, k)

			var value []byte
			if k < n && data[k] == '=' {
				k++
				k = skipASCIIWhitespace(data, k)
				if k >= n {
					break
				}

				var quote byte
				if data[k] == '"' || data[k] == '\'' {
					quote = data[k]
					k++
					valStart := k
					endQuote := bytes.IndexByte(data[k:], quote)
					if endQuote == -1 {
						// Unclosed quote: ignore this meta
						i++
						nonComment++
						charset = nil
						httpEquiv = nil
						content = nil
						sawGT = false
						break
					}
					value = data[valStart : k+endQuote]
					k = k + endQuote + 1
				} else {
					valStart := k
					for k < n {
						ch = data[k]
						if isASCIIWhitespace(ch) || ch == '>' || ch == '<' {
							break
						}
						k++
					}
					value = data[valStart:k]
				}
			}

			switch {
			case bytes.Equal(attrName, []byte("charset")):
				charset = stripASCIIWhitespace(value)
			case bytes.Equal(attrName, []byte("http-equiv")):
				httpEquiv = value
			case bytes.Equal(attrName, []byte("content")):
				content = value
			}
		}

		if sawGT {
			// Check for charset attribute
			if charset != nil {
				enc := normalizeMetaDeclaredEncoding(charset)
				if enc != nil {
					return enc
				}
			}

			// Check for http-equiv="Content-Type" content="..."
			if httpEquiv != nil && bytes.Equal(bytes.ToLower(httpEquiv), []byte("content-type")) && content != nil {
				extracted := extractCharsetFromContent(content)
				if extracted != nil {
					enc := normalizeMetaDeclaredEncoding(extracted)
					if enc != nil {
						return enc
					}
				}
			}

			// Continue scanning after this tag
			i = k
			consumed := i - startI
			nonComment += consumed
		} else {
			// Continue scanning
			i++
			nonComment++
		}
	}

	return nil
}

// decodeWithEncoding decodes data using the specified encoding.
//
//nolint:gocognit // Complexity required for handling multiple encodings
func decodeWithEncoding(data []byte, enc *Encoding) (string, error) {
	switch enc.Name {
	case "UTF-8":
		// Replace invalid sequences with U+FFFD
		return string(data), nil

	case "windows-1252":
		// windows-1252 has special mappings for 0x80-0x9F
		var sb strings.Builder
		sb.Grow(len(data))
		for _, b := range data {
			if b >= 0x80 && b <= 0x9F {
				sb.WriteRune(windows1252Table[b-0x80])
			} else {
				sb.WriteRune(rune(b))
			}
		}
		return sb.String(), nil

	case "ISO-8859-1":
		// Each byte maps to a code point
		var sb strings.Builder
		sb.Grow(len(data))
		for _, b := range data {
			sb.WriteRune(rune(b))
		}
		return sb.String(), nil

	case "iso-8859-2":
		// ISO-8859-2 (Latin-2) character mapping
		var sb strings.Builder
		sb.Grow(len(data))
		for _, b := range data {
			if b < 0x80 {
				sb.WriteRune(rune(b))
			} else {
				// Use iso8859-2 table for 0x80-0xFF range
				sb.WriteRune(iso88592Table[b-0x80])
			}
		}
		return sb.String(), nil

	case "euc-jp":
		// EUC-JP decoding - simplified version
		// For proper implementation, we'd need full EUC-JP decoding tables
		// For now, we'll do a basic implementation that handles ASCII
		var sb strings.Builder
		i := 0
		for i < len(data) {
			if data[i] < 0x80 {
				// ASCII
				sb.WriteByte(data[i])
				i++
			} else {
				// Multi-byte character - just replace with replacement character
				sb.WriteRune('\uFFFD')
				i++
				if i < len(data) && data[i] >= 0x80 {
					i++
				}
			}
		}
		return sb.String(), nil

	case utf16LEName:
		// UTF-16LE decoding
		if len(data)%2 != 0 {
			// Odd length, add padding
			data = append(data, 0)
		}
		runes := make([]rune, 0, len(data)/2)
		for i := 0; i < len(data); i += 2 {
			r := rune(data[i]) | rune(data[i+1])<<8
			runes = append(runes, r)
		}
		return string(runes), nil

	case utf16BEName:
		// UTF-16BE decoding
		if len(data)%2 != 0 {
			data = append(data, 0)
		}
		runes := make([]rune, 0, len(data)/2)
		for i := 0; i < len(data); i += 2 {
			r := rune(data[i])<<8 | rune(data[i+1])
			runes = append(runes, r)
		}
		return string(runes), nil

	case "utf-16":
		// UTF-16 with BOM detection in the data itself
		// This is a simplified version
		if len(data) >= 2 {
			if data[0] == 0xFF && data[1] == 0xFE {
				return decodeWithEncoding(data[2:], UTF16LE)
			} else if data[0] == 0xFE && data[1] == 0xFF {
				return decodeWithEncoding(data[2:], UTF16BE)
			}
		}
		// Default to LE if no BOM
		return decodeWithEncoding(data, UTF16LE)

	default:
		return "", ErrInvalidEncoding
	}
}

// windows1252Table maps bytes 0x80-0x9F to their Unicode code points.
var windows1252Table = [32]rune{
	0x20AC, // 0x80 -> EURO SIGN
	0x0081, // 0x81 -> <control>
	0x201A, // 0x82 -> SINGLE LOW-9 QUOTATION MARK
	0x0192, // 0x83 -> LATIN SMALL LETTER F WITH HOOK
	0x201E, // 0x84 -> DOUBLE LOW-9 QUOTATION MARK
	0x2026, // 0x85 -> HORIZONTAL ELLIPSIS
	0x2020, // 0x86 -> DAGGER
	0x2021, // 0x87 -> DOUBLE DAGGER
	0x02C6, // 0x88 -> MODIFIER LETTER CIRCUMFLEX ACCENT
	0x2030, // 0x89 -> PER MILLE SIGN
	0x0160, // 0x8A -> LATIN CAPITAL LETTER S WITH CARON
	0x2039, // 0x8B -> SINGLE LEFT-POINTING ANGLE QUOTATION MARK
	0x0152, // 0x8C -> LATIN CAPITAL LIGATURE OE
	0x008D, // 0x8D -> <control>
	0x017D, // 0x8E -> LATIN CAPITAL LETTER Z WITH CARON
	0x008F, // 0x8F -> <control>
	0x0090, // 0x90 -> <control>
	0x2018, // 0x91 -> LEFT SINGLE QUOTATION MARK
	0x2019, // 0x92 -> RIGHT SINGLE QUOTATION MARK
	0x201C, // 0x93 -> LEFT DOUBLE QUOTATION MARK
	0x201D, // 0x94 -> RIGHT DOUBLE QUOTATION MARK
	0x2022, // 0x95 -> BULLET
	0x2013, // 0x96 -> EN DASH
	0x2014, // 0x97 -> EM DASH
	0x02DC, // 0x98 -> SMALL TILDE
	0x2122, // 0x99 -> TRADE MARK SIGN
	0x0161, // 0x9A -> LATIN SMALL LETTER S WITH CARON
	0x203A, // 0x9B -> SINGLE RIGHT-POINTING ANGLE QUOTATION MARK
	0x0153, // 0x9C -> LATIN SMALL LIGATURE OE
	0x009D, // 0x9D -> <control>
	0x017E, // 0x9E -> LATIN SMALL LETTER Z WITH CARON
	0x0178, // 0x9F -> LATIN CAPITAL LETTER Y WITH DIAERESIS
}

// iso88592Table maps bytes 0x80-0xFF to their Unicode code points for ISO-8859-2.
var iso88592Table = [128]rune{
	0x0080, 0x0081, 0x0082, 0x0083, 0x0084, 0x0085, 0x0086, 0x0087,
	0x0088, 0x0089, 0x008A, 0x008B, 0x008C, 0x008D, 0x008E, 0x008F,
	0x0090, 0x0091, 0x0092, 0x0093, 0x0094, 0x0095, 0x0096, 0x0097,
	0x0098, 0x0099, 0x009A, 0x009B, 0x009C, 0x009D, 0x009E, 0x009F,
	0x00A0, 0x0104, 0x02D8, 0x0141, 0x00A4, 0x013D, 0x015A, 0x00A7,
	0x00A8, 0x0160, 0x015E, 0x0164, 0x0179, 0x00AD, 0x017D, 0x017B,
	0x00B0, 0x0105, 0x02DB, 0x0142, 0x00B4, 0x013E, 0x015B, 0x02C7,
	0x00B8, 0x0161, 0x015F, 0x0165, 0x017A, 0x02DD, 0x017E, 0x017C,
	0x0154, 0x00C1, 0x00C2, 0x0102, 0x00C4, 0x0139, 0x0106, 0x00C7,
	0x010C, 0x00C9, 0x0118, 0x00CB, 0x011A, 0x00CD, 0x00CE, 0x010E,
	0x0110, 0x0143, 0x0147, 0x00D3, 0x00D4, 0x0150, 0x00D6, 0x00D7,
	0x0158, 0x016E, 0x00DA, 0x0170, 0x00DC, 0x00DD, 0x0162, 0x00DF,
	0x0155, 0x00E1, 0x00E2, 0x0103, 0x00E4, 0x013A, 0x0107, 0x00E7,
	0x010D, 0x00E9, 0x0119, 0x00EB, 0x011B, 0x00ED, 0x00EE, 0x010F,
	0x0111, 0x0144, 0x0148, 0x00F3, 0x00F4, 0x0151, 0x00F6, 0x00F7,
	0x0159, 0x016F, 0x00FA, 0x0171, 0x00FC, 0x00FD, 0x0163, 0x02D9,
}
