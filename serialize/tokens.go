// Package serialize provides HTML serialization for DOM nodes and token streams.
package serialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Sentinel errors for token serialization.
var (
	ErrUnknownTokenType      = errors.New("unknown token type")
	ErrInvalidTokenFormat    = errors.New("invalid token format")
	ErrStartTagMissingFields = errors.New("startTag needs at least 3 elements")
	ErrEndTagMissingFields   = errors.New("endTag needs at least 3 elements")
	ErrEmptyTagMissingFields = errors.New("emptyTag needs at least 2 elements")
	ErrCharactersMissing     = errors.New("characters token needs at least 2 elements")
	ErrCommentMissing        = errors.New("comment token needs at least 2 elements")
	ErrDoctypeMissing        = errors.New("doctype token needs at least 2 elements")
)

// SerializeTokenOptions controls token serialization behavior.
type SerializeTokenOptions struct {
	// QuoteChar sets the preferred quote character for attributes (' or ")
	QuoteChar rune
	// UseTrailingSolidus adds trailing slash to void elements (e.g., <img />)
	UseTrailingSolidus bool
	// MinimizeBooleanAttributes omits value for boolean attributes (default true)
	MinimizeBooleanAttributes bool
	// EscapeLtInAttrs escapes < in attribute values
	EscapeLtInAttrs bool
	// EscapeRcdata escapes content in rcdata elements (script, style)
	EscapeRcdata bool
	// StripWhitespace collapses whitespace in text nodes
	StripWhitespace bool
	// OmitOptionalTags omits optional start/end tags per HTML5 spec
	OmitOptionalTags bool
	// InjectMetaCharset injects charset meta tag
	InjectMetaCharset bool
	// Encoding specifies the encoding for inject_meta_charset
	Encoding string
}

// DefaultSerializeTokenOptions returns default serialization options.
func DefaultSerializeTokenOptions() SerializeTokenOptions {
	return SerializeTokenOptions{
		QuoteChar:                 '"',
		MinimizeBooleanAttributes: true,
		OmitOptionalTags:          true,
	}
}

// SerializeTokens serializes a stream of html5lib test tokens to HTML.
// Each token is a json.RawMessage array in the html5lib format.
func SerializeTokens(tokens []json.RawMessage) (string, error) {
	opts := DefaultSerializeTokenOptions()
	return SerializeTokensWithOptions(tokens, opts)
}

// SerializeTokensWithOptions serializes tokens with custom options.
func SerializeTokensWithOptions(tokens []json.RawMessage, opts SerializeTokenOptions) (string, error) {
	var sb strings.Builder
	var rawTextDepth int
	var preformattedDepth int
	var inHead bool
	var headHasCharsetMeta bool
	var injectedMeta bool

	for i, raw := range tokens {
		if inHead && opts.InjectMetaCharset && opts.Encoding != "" && !headHasCharsetMeta && !injectedMeta {
			typ, tag := getTokenInfo(raw)
			if typ == "EndTag" && tag == "head" {
				serializeInjectedMeta(&sb, opts)
				injectedMeta = true
			}
		}

		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidTokenFormat, err)
		}
		if len(arr) == 0 {
			continue
		}

		var tokenType string
		if err := json.Unmarshal(arr[0], &tokenType); err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidTokenFormat, err)
		}

		var err error
		switch tokenType {
		case "StartTag":
			err = serializeStartTagToken(&sb, arr, opts, tokens, i)
			if err == nil {
				tagName := tokenTagName(tokenType, arr)
				if tagName == "head" {
					inHead = true
					injectedMeta = false
					if opts.InjectMetaCharset && opts.Encoding != "" {
						headHasCharsetMeta = hasCharsetMetaAhead(tokens, i)
						if !headHasCharsetMeta {
							serializeInjectedMeta(&sb, opts)
							injectedMeta = true
						}
					}
				}
				if tagName == "pre" || tagName == "textarea" {
					preformattedDepth++
				}
				if isRawTextElement(tagName) {
					rawTextDepth++
				}
			}
		case "EndTag":
			if inHead && opts.InjectMetaCharset && opts.Encoding != "" && !headHasCharsetMeta && !injectedMeta {
				tagName := tokenTagName(tokenType, arr)
				if tagName == "head" {
					serializeInjectedMeta(&sb, opts)
					injectedMeta = true
				}
			}
			err = serializeEndTagToken(&sb, arr, opts, tokens, i)
			if err == nil {
				tagName := tokenTagName(tokenType, arr)
				if tagName == "head" {
					inHead = false
					headHasCharsetMeta = false
					injectedMeta = false
				}
				if tagName == "pre" || tagName == "textarea" {
					if preformattedDepth > 0 {
						preformattedDepth--
					}
				}
				if isRawTextElement(tagName) && rawTextDepth > 0 {
					rawTextDepth--
				}
			}
		case "EmptyTag":
			err = serializeEmptyTagToken(&sb, arr, opts)
		case "Characters":
			inRawText := rawTextDepth > 0
			inPreformatted := preformattedDepth > 0
			err = serializeCharactersToken(&sb, arr, inRawText, inPreformatted, opts)
		case "Comment":
			err = serializeCommentToken(&sb, arr)
		case "Doctype":
			err = serializeDoctypeToken(&sb, arr)
		default:
			return "", fmt.Errorf("%w: %s", ErrUnknownTokenType, tokenType)
		}
		if err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

// serializeStartTagToken handles ["StartTag", namespace, tagName, attrs]
func serializeStartTagToken(sb *strings.Builder, arr []json.RawMessage, opts SerializeTokenOptions, tokens []json.RawMessage, idx int) error {
	if len(arr) < 3 {
		return ErrStartTagMissingFields
	}

	var tagName string
	if err := json.Unmarshal(arr[2], &tagName); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}

	// Check if this start tag should be omitted
	if opts.OmitOptionalTags && shouldOmitStartTag(tagName, arr, tokens, idx) {
		return nil
	}

	sb.WriteByte('<')
	sb.WriteString(tagName)

	// Parse attributes if present
	if len(arr) > 3 {
		if err := serializeTokenAttrs(sb, arr[3], opts, tagName); err != nil {
			return err
		}
	}

	// Add trailing solidus for void elements if requested
	if opts.UseTrailingSolidus && isVoidElement(tagName) {
		sb.WriteString(" /")
	}

	sb.WriteByte('>')

	return nil
}

// serializeEndTagToken handles ["EndTag", namespace, tagName]
func serializeEndTagToken(sb *strings.Builder, arr []json.RawMessage, opts SerializeTokenOptions, tokens []json.RawMessage, idx int) error {
	if len(arr) < 3 {
		return ErrEndTagMissingFields
	}

	var tagName string
	if err := json.Unmarshal(arr[2], &tagName); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}

	// Check if this end tag should be omitted
	if opts.OmitOptionalTags && shouldOmitEndTag(tagName, tokens, idx) {
		return nil
	}

	sb.WriteString("</")
	sb.WriteString(tagName)
	sb.WriteByte('>')

	return nil
}

// serializeEmptyTagToken handles ["EmptyTag", tagName, attrs]
func serializeEmptyTagToken(sb *strings.Builder, arr []json.RawMessage, opts SerializeTokenOptions) error {
	if len(arr) < 2 {
		return ErrEmptyTagMissingFields
	}

	var tagName string
	if err := json.Unmarshal(arr[1], &tagName); err != nil {
		return fmt.Errorf("invalid tag name: %w", err)
	}

	sb.WriteByte('<')
	sb.WriteString(tagName)

	// Parse attributes if present
	if len(arr) > 2 {
		if err := serializeTokenAttrs(sb, arr[2], opts, tagName); err != nil {
			return err
		}
	}

	// Add trailing solidus if requested
	if opts.UseTrailingSolidus {
		sb.WriteString(" /")
	}

	sb.WriteByte('>')
	return nil
}

// serializeTokenAttrs serializes attributes from either array or object format.
func serializeTokenAttrs(sb *strings.Builder, raw json.RawMessage, opts SerializeTokenOptions, tagName string) error {
	// Try array format first: [{namespace, name, value}, ...]
	attrs, err := parseTokenAttrs(raw)
	if err != nil {
		return err
	}

	if opts.InjectMetaCharset && opts.Encoding != "" && tagName == "meta" {
		attrs = normalizeMetaCharsetAttrs(attrs, opts.Encoding)
	}

	if len(attrs) == 0 {
		return nil
	}

	sortTokenAttrs(attrs)
	for _, attr := range attrs {
		sb.WriteByte(' ')
		sb.WriteString(attr.Name)
		serializeTokenAttrValue(sb, attr.Name, attr.Value, opts)
	}

	return nil
}

// serializeTokenAttrValue serializes an attribute value with proper quoting.
// Per html5lib serialization spec:
// - Unquoted if value contains no special characters
// - Single quotes if value contains " but not '
// - Double quotes otherwise, escaping " as &quot;
func serializeTokenAttrValue(sb *strings.Builder, name, value string, opts SerializeTokenOptions) {
	// Handle boolean attribute minimization
	// When MinimizeBooleanAttributes is true, omit value if:
	// 1. It's a known boolean attribute (disabled, checked, etc.)
	// 2. The value equals the attribute name (e.g., irrelevant="irrelevant")
	if opts.MinimizeBooleanAttributes && (value == "" || value == name) {
		// Don't write =value for boolean attributes
		return
	}

	if value == "" {
		// Empty value still needs =value
		sb.WriteString("=\"\"")
		return
	}

	// Check what characters are in the value
	hasDoubleQuote := strings.ContainsRune(value, '"')
	hasSingleQuote := strings.ContainsRune(value, '\'')
	needsQuoting := needsTokenAttrQuoting(value)

	// Determine quote character based on options
	useQuoteChar := opts.QuoteChar
	if useQuoteChar == 0 {
		useQuoteChar = '"'
	}

	// If quote_char option forces single quotes
	if useQuoteChar == '\'' {
		sb.WriteString("='")
		for _, r := range value {
			switch r {
			case '\'':
				sb.WriteString("&#39;")
			case '&':
				sb.WriteString("&amp;")
			default:
				sb.WriteRune(r)
			}
		}
		sb.WriteByte('\'')
		return
	}

	switch {
	case !needsQuoting:
		// Unquoted attribute value
		sb.WriteByte('=')
		sb.WriteString(value)
	case hasDoubleQuote && !hasSingleQuote:
		// Use single quotes, escape & as &amp;
		sb.WriteString("='")
		for _, r := range value {
			if r == '&' {
				sb.WriteString("&amp;")
			} else {
				sb.WriteRune(r)
			}
		}
		sb.WriteByte('\'')
	default:
		// Use double quotes, escape " as &quot;
		sb.WriteString("=\"")
		for _, r := range value {
			switch r {
			case '"':
				sb.WriteString("&quot;")
			case '&':
				sb.WriteString("&amp;")
			case '<':
				if opts.EscapeLtInAttrs {
					sb.WriteString("&lt;")
				} else {
					sb.WriteRune(r)
				}
			default:
				sb.WriteRune(r)
			}
		}
		sb.WriteByte('"')
	}
}

// needsTokenAttrQuoting returns true if the attribute value needs quoting.
func needsTokenAttrQuoting(value string) bool {
	for _, r := range value {
		switch r {
		case ' ', '\t', '\n', '\f', '\r', '"', '\'', '=', '>', '`':
			return true
		}
	}
	return false
}

// serializeCharactersToken handles ["Characters", data]
func serializeCharactersToken(sb *strings.Builder, arr []json.RawMessage, inRawText bool, inPreformatted bool, opts SerializeTokenOptions) error {
	if len(arr) < 2 {
		return ErrCharactersMissing
	}

	var data string
	if err := json.Unmarshal(arr[1], &data); err != nil {
		return fmt.Errorf("invalid character data: %w", err)
	}

	if opts.StripWhitespace && !inRawText && !inPreformatted {
		data = collapseTokenWhitespace(data)
	}

	if inRawText && !opts.EscapeRcdata {
		// Don't escape content in script/style/etc. (unless EscapeRcdata is set)
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

// serializeCommentToken handles ["Comment", data]
func serializeCommentToken(sb *strings.Builder, arr []json.RawMessage) error {
	if len(arr) < 2 {
		return ErrCommentMissing
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
		return ErrDoctypeMissing
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

// shouldOmitStartTag checks if a start tag can be omitted per HTML5 spec.
// Per https://html.spec.whatwg.org/multipage/syntax.html#optional-tags
func shouldOmitStartTag(tagName string, arr []json.RawMessage, tokens []json.RawMessage, idx int) bool {
	// Get the next token info
	nextType, nextTag := getNextTokenInfo(tokens, idx)

	// Check if the element has attributes - if so, don't omit
	if hasAttributes(arr) {
		return false
	}

	switch tagName {
	case "html":
		// An html element's start tag can be omitted if the first thing inside
		// the html element is not a comment.
		// Also keep if followed by space character (first char is space)
		if nextType == "Comment" {
			return false
		}
		if nextType == "Characters" && startsWithSpace(tokens, idx) {
			return false
		}
		return true
	case "head":
		// A head element's start tag can be omitted if the element is empty,
		// or if the first thing inside the head element is an element.
		return nextType == "StartTag" || nextType == "EmptyTag" || nextType == "EndTag"
	case "body":
		// A body element's start tag can be omitted if the element is empty,
		// or if the first thing inside the body element is not space character
		// or a comment.
		if nextType == "Comment" {
			return false
		}
		if nextType == "Characters" && startsWithSpace(tokens, idx) {
			return false // Don't omit if followed by space character
		}
		return true
	case "colgroup":
		// A colgroup element's start tag can be omitted if the first thing inside
		// the colgroup element is a col element.
		if nextType == "StartTag" || nextType == "EmptyTag" {
			return nextTag == "col"
		}
		return false
	case "tbody":
		// A tbody element's start tag can be omitted if the first thing inside
		// the tbody element is a tr element and the tbody is the first in a table.
		if nextType == "StartTag" && nextTag == "tr" {
			prevType, prevTag := getPrevTokenInfo(tokens, idx)
			return prevType == "StartTag" && prevTag == "table"
		}
		return false
	}
	return false
}

// startsWithSpace checks if the next Characters token starts with whitespace.
func startsWithSpace(tokens []json.RawMessage, idx int) bool {
	if idx+1 >= len(tokens) {
		return false
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(tokens[idx+1], &arr); err != nil || len(arr) < 2 {
		return false
	}

	var data string
	if err := json.Unmarshal(arr[1], &data); err != nil || len(data) == 0 {
		return false
	}

	// Check if first character is a space character
	switch data[0] {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	}
	return false
}

// hasAttributes returns true if the token has any attributes.
func hasAttributes(arr []json.RawMessage) bool {
	// For StartTag: ["StartTag", namespace, tagName, attrs]
	if len(arr) <= 3 {
		return false
	}

	// Check if attrs is empty
	var attrArray []interface{}
	if err := json.Unmarshal(arr[3], &attrArray); err == nil && len(attrArray) > 0 {
		return true
	}

	var attrObj map[string]interface{}
	if err := json.Unmarshal(arr[3], &attrObj); err == nil && len(attrObj) > 0 {
		return true
	}

	return false
}

// shouldOmitEndTag checks if an end tag can be omitted per HTML5 spec.
func shouldOmitEndTag(tagName string, tokens []json.RawMessage, idx int) bool {
	nextType, nextTag := getNextTokenInfo(tokens, idx)

	switch tagName {
	case "html":
		// An html element's end tag can be omitted if the html element is not
		// immediately followed by a comment.
		if nextType == "Comment" {
			return false
		}
		// Also don't omit if followed by space character
		if nextType == "Characters" && startsWithSpace(tokens, idx) {
			return false
		}
		return true
	case "head":
		// A head element's end tag can be omitted if the head element is not
		// immediately followed by space character or a comment.
		if nextType == "Comment" || (nextType == "Characters" && startsWithSpace(tokens, idx)) {
			return false
		}
		return true
	case "body":
		// A body element's end tag can be omitted if the body element is not
		// immediately followed by a comment.
		if nextType == "Comment" || (nextType == "Characters" && startsWithSpace(tokens, idx)) {
			return false
		}
		return true
	case "li":
		// An li element's end tag can be omitted if the li element is
		// immediately followed by another li element or if there is no more
		// content in the parent element.
		return nextType == "" || (nextType == "StartTag" && nextTag == "li") || nextType == "EndTag"
	case "dt":
		// A dt element's end tag can be omitted if the dt element is
		// immediately followed by another dt element or a dd element.
		return nextType == "StartTag" && (nextTag == "dt" || nextTag == "dd")
	case "dd":
		// A dd element's end tag can be omitted if the dd element is
		// immediately followed by another dd element or a dt element, or
		// if there is no more content in the parent element.
		return nextType == "" || (nextType == "StartTag" && (nextTag == "dd" || nextTag == "dt")) || nextType == "EndTag"
	case "p":
		// A p element's end tag can be omitted if the p element is immediately
		// followed by certain elements, or if there is no more content.
		if nextType == "" || nextType == "EndTag" {
			return true
		}
		if nextType == "StartTag" || nextType == "EmptyTag" {
			switch nextTag {
			case "address", "article", "aside", "blockquote", "details", "dialog",
				"dir", "div", "dl", "fieldset", "figcaption", "figure", "footer",
				"form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hgroup",
				"hr", "main", "menu", "nav", "ol", "p", "pre", "search", "section",
				"table", "ul", "datagrid":
				return true
			}
		}
		return false
	case "optgroup":
		// An optgroup element's end tag can be omitted if the optgroup element
		// is immediately followed by another optgroup element, or if there is
		// no more content in the parent element.
		return nextType == "" || nextType == "EndTag" ||
			(nextType == "StartTag" && nextTag == "optgroup")
	case "option":
		// An option element's end tag can be omitted if the option element is
		// immediately followed by another option element, or an optgroup element,
		// or if there is no more content in the parent element.
		return nextType == "" || nextType == "EndTag" ||
			(nextType == "StartTag" && (nextTag == "option" || nextTag == "optgroup"))
	case "colgroup":
		// A colgroup element's end tag can be omitted if it is not immediately
		// followed by a comment or a space character.
		if nextType == "Comment" || (nextType == "Characters" && startsWithSpace(tokens, idx)) {
			return false
		}
		if nextType == "StartTag" && nextTag == "colgroup" {
			return false
		}
		return true
	case "thead":
		// End tag can be omitted if immediately followed by tbody or tfoot.
		return nextType == "StartTag" && (nextTag == "tbody" || nextTag == "tfoot")
	case "tbody":
		// End tag can be omitted if immediately followed by tbody or tfoot,
		// or if there is no more content.
		return nextType == "" || nextType == "EndTag" ||
			(nextType == "StartTag" && (nextTag == "tbody" || nextTag == "tfoot"))
	case "tfoot":
		// End tag can be omitted if immediately followed by tbody,
		// or if there is no more content.
		return nextType == "" || nextType == "EndTag" ||
			(nextType == "StartTag" && nextTag == "tbody")
	case "tr":
		// End tag can be omitted if immediately followed by another tr,
		// or if there is no more content.
		return nextType == "" || nextType == "EndTag" ||
			(nextType == "StartTag" && nextTag == "tr")
	case "td", "th":
		// End tag can be omitted if immediately followed by td/th,
		// or if there is no more content.
		return nextType == "" || nextType == "EndTag" ||
			(nextType == "StartTag" && (nextTag == "td" || nextTag == "th"))
	}
	return false
}

type tokenAttr struct {
	Name  string
	Value string
}

func parseTokenAttrs(raw json.RawMessage) ([]tokenAttr, error) {
	var attrArray []struct {
		Namespace *string `json:"namespace"`
		Name      string  `json:"name"`
		Value     string  `json:"value"`
	}
	if err := json.Unmarshal(raw, &attrArray); err == nil {
		if len(attrArray) == 0 {
			return nil, nil
		}
		attrs := make([]tokenAttr, 0, len(attrArray))
		for _, attr := range attrArray {
			attrs = append(attrs, tokenAttr{Name: attr.Name, Value: attr.Value})
		}
		return attrs, nil
	}

	var attrObj map[string]string
	if err := json.Unmarshal(raw, &attrObj); err == nil {
		if len(attrObj) == 0 {
			return nil, nil
		}
		attrs := make([]tokenAttr, 0, len(attrObj))
		for name, value := range attrObj {
			attrs = append(attrs, tokenAttr{Name: name, Value: value})
		}
		return attrs, nil
	}

	return nil, nil
}

func sortTokenAttrs(attrs []tokenAttr) {
	if len(attrs) < 2 {
		return
	}
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].Name < attrs[j].Name
	})
}

func normalizeMetaCharsetAttrs(attrs []tokenAttr, encoding string) []tokenAttr {
	if len(attrs) == 0 {
		return attrs
	}
	var hasHTTP bool
	var httpIdx int
	var hasContent bool
	var contentIdx int
	for i, attr := range attrs {
		if strings.EqualFold(attr.Name, "charset") {
			attrs[i].Value = encoding
			return attrs
		}
		if strings.EqualFold(attr.Name, "http-equiv") {
			hasHTTP = true
			httpIdx = i
		}
		if strings.EqualFold(attr.Name, "content") {
			hasContent = true
			contentIdx = i
		}
	}
	if hasHTTP && strings.EqualFold(attrs[httpIdx].Value, "content-type") {
		content := "text/html; charset=" + encoding
		if hasContent {
			attrs[contentIdx].Value = content
		} else {
			attrs = append(attrs, tokenAttr{Name: "content", Value: content})
		}
	}
	return attrs
}

func hasCharsetMetaAhead(tokens []json.RawMessage, idx int) bool {
	for i := idx + 1; i < len(tokens); i++ {
		typ, tag := getTokenInfo(tokens[i])
		if typ == "" {
			return false
		}
		if typ == "EndTag" && tag == "head" {
			return false
		}
		if typ == "StartTag" || typ == "EmptyTag" {
			if tag != "meta" {
				continue
			}
			var arr []json.RawMessage
			if err := json.Unmarshal(tokens[i], &arr); err != nil {
				continue
			}
			var rawAttrs json.RawMessage
			if typ == "StartTag" {
				if len(arr) > 3 {
					rawAttrs = arr[3]
				}
			} else {
				if len(arr) > 2 {
					rawAttrs = arr[2]
				}
			}
			if len(rawAttrs) == 0 {
				continue
			}
			attrs, _ := parseTokenAttrs(rawAttrs)
			for _, attr := range attrs {
				if strings.EqualFold(attr.Name, "charset") {
					return true
				}
			}
			var httpEquiv bool
			for _, attr := range attrs {
				if strings.EqualFold(attr.Name, "http-equiv") && strings.EqualFold(attr.Value, "content-type") {
					httpEquiv = true
					break
				}
			}
			if httpEquiv {
				return true
			}
		}
	}
	return false
}

func getTokenInfo(raw json.RawMessage) (string, string) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
		return "", ""
	}
	var tokenType string
	if err := json.Unmarshal(arr[0], &tokenType); err != nil {
		return "", ""
	}
	return tokenType, tokenTagName(tokenType, arr)
}

func tokenTagName(tokenType string, arr []json.RawMessage) string {
	var tagName string
	switch tokenType {
	case "StartTag", "EndTag":
		if len(arr) >= 3 {
			_ = json.Unmarshal(arr[2], &tagName)
		}
	case "EmptyTag":
		if len(arr) >= 2 {
			_ = json.Unmarshal(arr[1], &tagName)
		}
	}
	return tagName
}

func serializeInjectedMeta(sb *strings.Builder, opts SerializeTokenOptions) {
	if opts.Encoding == "" {
		return
	}
	sb.WriteString("<meta charset")
	serializeTokenAttrValue(sb, "charset", opts.Encoding, opts)
	sb.WriteByte('>')
}

func collapseTokenWhitespace(s string) string {
	var sb strings.Builder
	inWhitespace := false
	for _, r := range s {
		if isWhitespaceRune(r) {
			if !inWhitespace {
				sb.WriteByte(' ')
				inWhitespace = true
			}
			continue
		}
		sb.WriteRune(r)
		inWhitespace = false
	}
	return sb.String()
}

func isWhitespaceRune(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

// getNextTokenInfo returns the type and tag name of the next token.
func getNextTokenInfo(tokens []json.RawMessage, idx int) (string, string) {
	if idx+1 >= len(tokens) {
		return "", "" // EOF
	}

	var tokenType, tagName string
	var arr []json.RawMessage
	if err := json.Unmarshal(tokens[idx+1], &arr); err != nil || len(arr) == 0 {
		return "", ""
	}

	if err := json.Unmarshal(arr[0], &tokenType); err != nil {
		return "", ""
	}

	// Get tag name for StartTag, EndTag, EmptyTag
	switch tokenType {
	case "StartTag", "EndTag":
		if len(arr) >= 3 {
			_ = json.Unmarshal(arr[2], &tagName)
		}
	case "EmptyTag":
		if len(arr) >= 2 {
			_ = json.Unmarshal(arr[1], &tagName)
		}
	}

	return tokenType, tagName
}

func getPrevTokenInfo(tokens []json.RawMessage, idx int) (string, string) {
	if idx-1 < 0 {
		return "", ""
	}

	var tokenType, tagName string
	var arr []json.RawMessage
	if err := json.Unmarshal(tokens[idx-1], &arr); err != nil || len(arr) == 0 {
		return "", ""
	}

	if err := json.Unmarshal(arr[0], &tokenType); err != nil {
		return "", ""
	}

	switch tokenType {
	case "StartTag", "EndTag":
		if len(arr) >= 3 {
			_ = json.Unmarshal(arr[2], &tagName)
		}
	case "EmptyTag":
		if len(arr) >= 2 {
			_ = json.Unmarshal(arr[1], &tagName)
		}
	}

	return tokenType, tagName
}
