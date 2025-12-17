package tokenizer

import (
	"strconv"
	"unicode"

	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
)

func decodeNumericEntity(text string, isHex bool) rune {
	base := 10
	if isHex {
		base = 16
	}
	codepoint, err := strconv.ParseInt(text, base, 32)
	if err != nil {
		return unicode.ReplacementChar
	}

	cp := int(codepoint)
	if replacement, ok := constants.NumericReplacements[cp]; ok {
		return replacement
	}

	// Invalid ranges per HTML5 spec.
	if cp > 0x10FFFF {
		return unicode.ReplacementChar
	}
	if cp >= 0xD800 && cp <= 0xDFFF {
		return unicode.ReplacementChar
	}
	return rune(cp)
}

// decodeEntitiesInText decodes HTML entities in a string.
//
// This follows the behavior of the Python reference implementation and is used
// when flushing text and attribute values.
func decodeEntitiesInText(text string, inAttribute bool) string {
	var out []rune
	out = make([]rune, 0, len(text))

	runes := []rune(text)
	i := 0
	for i < len(runes) {
		// Find next '&'
		nextAmp := -1
		for j := i; j < len(runes); j++ {
			if runes[j] == '&' {
				nextAmp = j
				break
			}
		}
		if nextAmp == -1 {
			out = append(out, runes[i:]...)
			break
		}
		if nextAmp > i {
			out = append(out, runes[i:nextAmp]...)
		}

		i = nextAmp
		j := i + 1
		if j < len(runes) && runes[j] == '#' {
			j++
			isHex := false
			if j < len(runes) && (runes[j] == 'x' || runes[j] == 'X') {
				isHex = true
				j++
			}

			digitStart := j
			if isHex {
				for j < len(runes) && ((runes[j] >= '0' && runes[j] <= '9') || (runes[j] >= 'a' && runes[j] <= 'f') || (runes[j] >= 'A' && runes[j] <= 'F')) {
					j++
				}
			} else {
				for j < len(runes) && (runes[j] >= '0' && runes[j] <= '9') {
					j++
				}
			}

			hasSemicolon := j < len(runes) && runes[j] == ';'
			digitText := string(runes[digitStart:j])
			if digitText != "" {
				out = append(out, decodeNumericEntity(digitText, isHex))
				if hasSemicolon {
					i = j + 1
				} else {
					i = j
				}
				continue
			}

			// Invalid numeric entity, keep as-is.
			if hasSemicolon && j < len(runes) {
				out = append(out, runes[i:j+1]...)
				i = j + 1
			} else {
				out = append(out, runes[i:j]...)
				i = j
			}
			continue
		}

		// Named entity: collect alphanumeric.
		for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j])) {
			j++
		}
		entityName := string(runes[i+1 : j])
		hasSemicolon := j < len(runes) && runes[j] == ';'

		if entityName == "" {
			out = append(out, '&')
			i++
			continue
		}

		// Exact match with semicolon.
		if hasSemicolon {
			if value, ok := constants.NamedEntities[entityName]; ok {
				out = append(out, []rune(value)...)
				i = j + 1
				continue
			}

			// Legacy prefix match in text.
			if !inAttribute {
				bestLen := 0
				best := ""
				for k := len(entityName); k > 0; k-- {
					prefix := entityName[:k]
					if constants.LegacyEntities[prefix] {
						if v, ok := constants.NamedEntities[prefix]; ok {
							best = v
							bestLen = k
							break
						}
					}
				}
				if bestLen > 0 {
					out = append(out, []rune(best)...)
					i = i + 1 + bestLen
					continue
				}
			}
		}

		// Without semicolon for legacy.
		if constants.LegacyEntities[entityName] {
			if value, ok := constants.NamedEntities[entityName]; ok {
				nextChar := rune(0)
				if j < len(runes) {
					nextChar = runes[j]
				}
				if inAttribute && nextChar != 0 && (unicode.IsLetter(nextChar) || unicode.IsDigit(nextChar) || nextChar == '=') {
					out = append(out, '&')
					i++
					continue
				}
				out = append(out, []rune(value)...)
				i = j
				continue
			}
		}

		// Longest legacy prefix match.
		bestLen := 0
		best := ""
		for k := len(entityName); k > 0; k-- {
			prefix := entityName[:k]
			if constants.LegacyEntities[prefix] {
				if v, ok := constants.NamedEntities[prefix]; ok {
					best = v
					bestLen = k
					break
				}
			}
		}
		if bestLen > 0 {
			if inAttribute {
				out = append(out, '&')
				i++
				continue
			}
			out = append(out, []rune(best)...)
			i = i + 1 + bestLen
			continue
		}

		// No match.
		if hasSemicolon {
			out = append(out, runes[i:j+1]...)
			i = j + 1
		} else {
			out = append(out, '&')
			i++
		}
	}

	return string(out)
}
