// Package serialize provides HTML serialization for DOM nodes.
package serialize

import (
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// Options configures serialization behavior.
type Options struct {
	// Pretty enables pretty-printing with indentation.
	Pretty bool

	// IndentSize is the number of spaces per indentation level.
	IndentSize int
}

// DefaultOptions returns the default serialization options.
func DefaultOptions() Options {
	return Options{
		Pretty:     false,
		IndentSize: 2,
	}
}

// ToHTML serializes a node to HTML.
func ToHTML(node dom.Node, opts Options) string {
	var sb strings.Builder
	serializeNode(&sb, node, opts, 0)
	return sb.String()
}

func serializeNode(sb *strings.Builder, node dom.Node, opts Options, depth int) {
	serializeNodeWithInline(sb, node, opts, depth, false)
}

func serializeNodeWithInline(sb *strings.Builder, node dom.Node, opts Options, depth int, inline bool) {
	switch n := node.(type) {
	case *dom.Document:
		serializeDocument(sb, n, opts, depth)
	case *dom.DocumentType:
		serializeDoctype(sb, n)
	case *dom.Element:
		serializeElement(sb, n, opts, depth, inline)
	case *dom.Text:
		serializeText(sb, n, opts, depth)
	case *dom.Comment:
		serializeComment(sb, n, opts, depth, inline)
	}
}

func serializeDocument(sb *strings.Builder, doc *dom.Document, opts Options, depth int) {
	if doc.Doctype != nil {
		serializeDoctype(sb, doc.Doctype)
		if opts.Pretty {
			sb.WriteByte('\n')
		}
	}
	for _, child := range doc.Children() {
		serializeNode(sb, child, opts, depth)
	}
}

func serializeDoctype(sb *strings.Builder, dt *dom.DocumentType) {
	sb.WriteString("<!DOCTYPE ")
	sb.WriteString(dt.Name)
	if dt.PublicID != "" {
		sb.WriteString(" PUBLIC \"")
		sb.WriteString(dt.PublicID)
		sb.WriteByte('"')
		if dt.SystemID != "" {
			sb.WriteString(" \"")
			sb.WriteString(dt.SystemID)
			sb.WriteByte('"')
		}
	} else if dt.SystemID != "" {
		sb.WriteString(" SYSTEM \"")
		sb.WriteString(dt.SystemID)
		sb.WriteByte('"')
	}
	sb.WriteByte('>')
}

func serializeElement(sb *strings.Builder, elem *dom.Element, opts Options, depth int, inline bool) {
	// Only add indentation for block elements on their own line, not inline elements
	if opts.Pretty && depth > 0 && !inline {
		sb.WriteString(strings.Repeat(" ", depth*opts.IndentSize))
	}

	sb.WriteByte('<')
	sb.WriteString(elem.TagName)

	for _, attr := range elem.Attributes.All() {
		sb.WriteByte(' ')
		sb.WriteString(attr.Name)
		sb.WriteString("=\"")
		sb.WriteString(escapeAttr(attr.Value))
		sb.WriteByte('"')
	}

	if isVoidElement(elem.TagName) {
		sb.WriteByte('>')
		return
	}

	sb.WriteByte('>')

	children := elem.Children()

	if opts.Pretty {
		serializeChildrenPretty(sb, children, opts, depth)
	} else {
		for _, child := range children {
			serializeNode(sb, child, opts, depth+1)
		}
	}

	sb.WriteString("</")
	sb.WriteString(elem.TagName)
	sb.WriteByte('>')
}

// serializeChildrenPretty handles pretty-printing of element children.
// It filters out whitespace-only text nodes and properly indents content.
func serializeChildrenPretty(sb *strings.Builder, children []dom.Node, opts Options, depth int) {
	// Filter to get significant children (skip whitespace-only text nodes)
	var significantChildren []dom.Node
	for _, child := range children {
		if text, ok := child.(*dom.Text); ok {
			if isWhitespaceOnly(text.Data) {
				continue
			}
		}
		significantChildren = append(significantChildren, child)
	}

	if len(significantChildren) == 0 {
		return
	}

	// Check if any child is a block element
	hasBlock := false
	for _, child := range significantChildren {
		if elem, ok := child.(*dom.Element); ok {
			if isBlockElement(elem.TagName) {
				hasBlock = true
				break
			}
		}
	}

	for _, child := range significantChildren {
		if hasBlock {
			sb.WriteByte('\n')
			// Only increment depth for block content (indented on new lines)
			serializeNode(sb, child, opts, depth+1)
		} else {
			// Inline content: don't increment depth, no extra indentation
			serializeNode(sb, child, opts, depth)
		}
	}

	if hasBlock {
		sb.WriteByte('\n')
		sb.WriteString(strings.Repeat(" ", depth*opts.IndentSize))
	}
}

// serializeText serializes a text node.
// In pretty mode, whitespace-only text nodes between block elements are skipped
// since the pretty printer handles formatting.
func serializeText(sb *strings.Builder, text *dom.Text, opts Options, depth int) {
	data := text.Data

	// In pretty mode, skip whitespace-only text nodes (they're just formatting noise)
	if opts.Pretty && isWhitespaceOnly(data) {
		return
	}

	// In pretty mode, collapse runs of whitespace but preserve leading/trailing
	// single spaces for inline content like "text <b>bold</b> more"
	if opts.Pretty {
		data = collapseWhitespace(data)
	}

	sb.WriteString(escapeText(data))
}

// serializeComment serializes a comment node.
func serializeComment(sb *strings.Builder, comment *dom.Comment, opts Options, depth int, inline bool) {
	if opts.Pretty && depth > 0 && !inline {
		sb.WriteString(strings.Repeat(" ", depth*opts.IndentSize))
	}
	sb.WriteString("<!--")
	sb.WriteString(comment.Data)
	sb.WriteString("-->")
}

// isWhitespaceOnly returns true if the string contains only whitespace characters.
func isWhitespaceOnly(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' && r != '\f' {
			return false
		}
	}
	return true
}

// collapseWhitespace collapses runs of whitespace into single spaces
// but preserves a single leading/trailing space if present.
// This is important for inline content like "text <b>bold</b> more".
func collapseWhitespace(s string) string {
	if len(s) == 0 {
		return s
	}

	var sb strings.Builder
	hasLeadingSpace := isWhitespaceChar(rune(s[0]))
	hasTrailingSpace := isWhitespaceChar(rune(s[len(s)-1]))

	inWhitespace := true // Start true to skip leading whitespace in loop
	for _, r := range s {
		if isWhitespaceChar(r) {
			if !inWhitespace {
				sb.WriteByte(' ')
				inWhitespace = true
			}
		} else {
			sb.WriteRune(r)
			inWhitespace = false
		}
	}

	result := sb.String()
	// Trim trailing space from collapsed content
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}

	// Restore leading/trailing spaces if original had them
	if hasLeadingSpace && len(result) > 0 {
		result = " " + result
	}
	if hasTrailingSpace && len(result) > 0 {
		result = result + " "
	}

	return result
}

// isWhitespaceChar returns true if r is an HTML whitespace character.
func isWhitespaceChar(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f'
}

// escapeText escapes text content for HTML.
func escapeText(s string) string {
	var sb strings.Builder
	for _, r := range s {
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
	return sb.String()
}

// escapeAttr escapes an attribute value.
func escapeAttr(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			sb.WriteString("&amp;")
		case '"':
			sb.WriteString("&quot;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// isVoidElement returns true if the tag is a void element.
func isVoidElement(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

// hasBlockContent returns true if the element contains block-level content.
func hasBlockContent(elem *dom.Element) bool {
	for _, child := range elem.Children() {
		if childElem, ok := child.(*dom.Element); ok {
			if isBlockElement(childElem.TagName) {
				return true
			}
		}
	}
	return false
}

// isBlockElement returns true if the tag is typically block-level.
func isBlockElement(tag string) bool {
	switch tag {
	case "address", "article", "aside", "blockquote", "body", "canvas", "dd", "div",
		"dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
		"h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hr", "html", "li", "main",
		"nav", "noscript", "ol", "p", "pre", "section", "table", "tbody", "td", "tfoot",
		"th", "thead", "title", "tr", "ul", "video":
		return true
	}
	return false
}
