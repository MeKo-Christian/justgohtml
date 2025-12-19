// Package serialize provides HTML serialization for DOM nodes and token streams.
package serialize

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SerializeTokens serializes a stream of html5lib test tokens to HTML.
// Each token is a json.RawMessage array in the html5lib format.
func SerializeTokens(tokens []json.RawMessage) (string, error) {
	var sb strings.Builder
	var openRawText string // tracks if we're inside script/style/etc.

	for _, raw := range tokens {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return "", fmt.Errorf("invalid token format: %w", err)
		}
		if len(arr) == 0 {
			continue
		}

		var tokenType string
		if err := json.Unmarshal(arr[0], &tokenType); err != nil {
			return "", fmt.Errorf("invalid token type: %w", err)
		}

		switch tokenType {
		case "StartTag":
			if err := serializeStartTag(&sb, arr); err != nil {
				return "", err
			}
			// Track raw text elements (content not escaped)
			if len(arr) > 2 {
				var tagName string
				if err := json.Unmarshal(arr[2], &tagName); err == nil {
					if isRawTextElement(tagName) {
						openRawText = tagName
					}
				}
			}

		case "EndTag":
			if err := serializeEndTag(&sb, arr); err != nil {
				return "", err
			}
			// Clear raw text tracking
			if len(arr) > 2 {
				var tagName string
				if err := json.Unmarshal(arr[2], &tagName); err == nil {
					if tagName == openRawText {
						openRawText = ""
					}
				}
			}

		case "EmptyTag":
			if err := serializeEmptyTag(&sb, arr); err != nil {
				return "", err
			}

		case "Characters":
			if err := serializeCharacters(&sb, arr, openRawText != ""); err != nil {
				return "", err
			}

		case "Comment":
			if err := serializeComment(&sb, arr); err != nil {
				return "", err
			}

		case "Doctype":
			if err := serializeDoctypeToken(&sb, arr); err != nil {
				return "", err
			}

		default:
			return "", fmt.Errorf("unknown token type: %s", tokenType)
		}
	}

	return sb.String(), nil
}

// serializeStartTag handles ["StartTag", namespace, tagName, attrs]
func serializeStartTag(sb *strings.Builder, arr []json.RawMessage) error {
	if len(arr) < 3 {
		return fmt.Errorf("StartTag needs at least 3 elements")
	}

	var tagName string
	if err := json.Unmarshal(arr[2], &tagName); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}

	sb.WriteByte('<')
	sb.WriteString(tagName)

	// Parse attributes if present
	if len(arr) > 3 {
		if err := serializeAttrs(sb, arr[3]); err != nil {
			return err
		}
	}

	sb.WriteByte('>')
	return nil
}

// serializeEndTag handles ["EndTag", namespace, tagName]
func serializeEndTag(sb *strings.Builder, arr []json.RawMessage) error {
	if len(arr) < 3 {
		return fmt.Errorf("EndTag needs at least 3 elements")
	}

	var tagName string
	if err := json.Unmarshal(arr[2], &tagName); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}

	sb.WriteString("</")
	sb.WriteString(tagName)
	sb.WriteByte('>')
	return nil
}

// serializeEmptyTag handles ["EmptyTag", tagName, attrs]
func serializeEmptyTag(sb *strings.Builder, arr []json.RawMessage) error {
	if len(arr) < 2 {
		return fmt.Errorf("EmptyTag needs at least 2 elements")
	}

	var tagName string
	if err := json.Unmarshal(arr[1], &tagName); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}

	sb.WriteByte('<')
	sb.WriteString(tagName)

	// Parse attributes if present
	if len(arr) > 2 {
		if err := serializeAttrs(sb, arr[2]); err != nil {
			return err
		}
	}

	sb.WriteByte('>')
	return nil
}

// serializeAttrs serializes attributes from either array or object format.
func serializeAttrs(sb *strings.Builder, raw json.RawMessage) error {
	// Try array format first: [{namespace, name, value}, ...]
	var attrArray []struct {
		Namespace *string `json:"namespace"`
		Name      string  `json:"name"`
		Value     string  `json:"value"`
	}
	if err := json.Unmarshal(raw, &attrArray); err == nil && len(attrArray) > 0 {
		for _, attr := range attrArray {
			sb.WriteByte(' ')
			sb.WriteString(attr.Name)
			serializeAttrValue(sb, attr.Value)
		}
		return nil
	}

	// Try object format: {name: value, ...}
	var attrObj map[string]string
	if err := json.Unmarshal(raw, &attrObj); err == nil && len(attrObj) > 0 {
		for name, value := range attrObj {
			sb.WriteByte(' ')
			sb.WriteString(name)
			serializeAttrValue(sb, value)
		}
	}

	return nil
}

// serializeAttrValue serializes an attribute value with proper quoting.
// Per html5lib serialization spec:
// - Unquoted if value contains no special characters
// - Single quotes if value contains " but not '
// - Double quotes otherwise, escaping " as &quot;
func serializeAttrValue(sb *strings.Builder, value string) {
	if value == "" {
		// Empty value still needs =value
		sb.WriteString("=\"\"")
		return
	}

	// Check what characters are in the value
	hasDoubleQuote := strings.ContainsRune(value, '"')
	hasSingleQuote := strings.ContainsRune(value, '\'')
	needsQuoting := needsAttrQuoting(value)

	if !needsQuoting {
		// Unquoted attribute value
		sb.WriteByte('=')
		sb.WriteString(value)
	} else if hasDoubleQuote && !hasSingleQuote {
		// Use single quotes
		sb.WriteString("='")
		sb.WriteString(value)
		sb.WriteByte('\'')
	} else {
		// Use double quotes, escape " as &quot;
		sb.WriteString("=\"")
		for _, r := range value {
			if r == '"' {
				sb.WriteString("&quot;")
			} else if r == '&' {
				sb.WriteString("&amp;")
			} else {
				sb.WriteRune(r)
			}
		}
		sb.WriteByte('"')
	}
}

// needsAttrQuoting returns true if the attribute value needs quoting.
func needsAttrQuoting(value string) bool {
	for _, r := range value {
		switch r {
		case ' ', '\t', '\n', '\f', '\r', '"', '\'', '=', '>', '`':
			return true
		}
	}
	return false
}

// serializeCharacters handles ["Characters", data]
func serializeCharacters(sb *strings.Builder, arr []json.RawMessage, inRawText bool) error {
	if len(arr) < 2 {
		return fmt.Errorf("Characters needs at least 2 elements")
	}

	var data string
	if err := json.Unmarshal(arr[1], &data); err != nil {
		return fmt.Errorf("invalid character data: %w", err)
	}

	if inRawText {
		// Don't escape content in script/style/etc.
		sb.WriteString(data)
	} else {
		// Escape special characters
		for _, r := range data {
			switch r {
			case '&':
				sb.WriteString("&amp;")
			case '<':
				sb.WriteString("&lt;")
			case '>':
				sb.WriteString("&gt;")
			default:
				sb.WriteRune(r)
			}
		}
	}
	return nil
}

// serializeComment handles ["Comment", data]
func serializeComment(sb *strings.Builder, arr []json.RawMessage) error {
	if len(arr) < 2 {
		return fmt.Errorf("Comment needs at least 2 elements")
	}

	var data string
	if err := json.Unmarshal(arr[1], &data); err != nil {
		return fmt.Errorf("invalid comment data: %w", err)
	}

	sb.WriteString("<!--")
	sb.WriteString(data)
	sb.WriteString("-->")
	return nil
}

// serializeDoctypeToken handles ["Doctype", name, publicId?, systemId?]
func serializeDoctypeToken(sb *strings.Builder, arr []json.RawMessage) error {
	if len(arr) < 2 {
		return fmt.Errorf("Doctype needs at least 2 elements")
	}

	var name string
	if err := json.Unmarshal(arr[1], &name); err != nil {
		return fmt.Errorf("invalid doctype name: %w", err)
	}

	sb.WriteString("<!DOCTYPE ")
	sb.WriteString(name)

	// Parse optional public ID
	var publicID string
	if len(arr) > 2 {
		// Can be null or string
		if err := json.Unmarshal(arr[2], &publicID); err != nil {
			// Might be null, which is fine
			publicID = ""
		}
	}

	// Parse optional system ID
	var systemID string
	if len(arr) > 3 {
		if err := json.Unmarshal(arr[3], &systemID); err != nil {
			systemID = ""
		}
	}

	if publicID != "" {
		sb.WriteString(" PUBLIC \"")
		sb.WriteString(publicID)
		sb.WriteByte('"')
		if systemID != "" {
			sb.WriteString(" \"")
			sb.WriteString(systemID)
			sb.WriteByte('"')
		}
	} else if systemID != "" {
		sb.WriteString(" SYSTEM \"")
		sb.WriteString(systemID)
		sb.WriteByte('"')
	}

	sb.WriteByte('>')
	return nil
}

// isRawTextElement returns true for elements whose content is not escaped.
func isRawTextElement(tag string) bool {
	switch tag {
	case "script", "style", "xmp", "iframe", "noembed", "noframes", "plaintext":
		return true
	}
	return false
}
