// Package serialize provides HTML serialization for DOM nodes.
package serialize

import (
	"strconv"
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

// ToMarkdown serializes a node to Markdown.
func ToMarkdown(node dom.Node) string {
	var sb strings.Builder
	serializeMarkdown(&sb, node, 0, false)
	return strings.TrimSpace(sb.String())
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
	significantChildren := make([]dom.Node, 0, len(children))
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
			serializeNodeWithInline(sb, child, opts, depth+1, false)
		} else {
			// Inline content: mark as inline so elements don't add indentation
			serializeNodeWithInline(sb, child, opts, depth, true)
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
func serializeText(sb *strings.Builder, text *dom.Text, opts Options, _ int) {
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
		result += " "
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

// isBlockElement returns true if the tag is typically block-level.
func isBlockElement(tag string) bool {
	switch tag {
	case "address", "article", "aside", "blockquote", "body", "canvas", "dd", "div", //nolint:goconst
		"dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
		"h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hr", "html", "li", "main",
		"nav", "noscript", "ol", "p", "pre", "section", "table", "tbody", "td", "tfoot",
		"th", "thead", "title", "tr", "ul", "video":
		return true
	}
	return false
}

// serializeMarkdown converts DOM nodes to Markdown format.
func serializeMarkdown(sb *strings.Builder, node dom.Node, listDepth int, inList bool) {
	switch n := node.(type) {
	case *dom.Document:
		for _, child := range n.Children() {
			serializeMarkdown(sb, child, listDepth, inList)
		}
	case *dom.Element:
		serializeElementMarkdown(sb, n, listDepth, inList)
	case *dom.Text:
		// Normalize whitespace - collapse multiple spaces/newlines into single spaces
		text := collapseMarkdownWhitespace(n.Data)
		if text != "" {
			sb.WriteString(text)
		}
	case *dom.Comment:
		// Comments are omitted in markdown
	}
}

// collapseMarkdownWhitespace collapses runs of whitespace including newlines.
func collapseMarkdownWhitespace(s string) string {
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	inWhitespace := true

	for _, r := range s {
		if isWhitespaceChar(r) {
			if !inWhitespace {
				result.WriteByte(' ')
				inWhitespace = true
			}
		} else {
			result.WriteRune(r)
			inWhitespace = false
		}
	}

	return strings.TrimSpace(result.String())
}

// serializeElementMarkdown converts an element to Markdown.
//
//nolint:funlen // Markdown serialization requires many element type cases
func serializeElementMarkdown(sb *strings.Builder, elem *dom.Element, listDepth int, inList bool) {
	switch elem.TagName {
	case "h1":
		sb.WriteString("# ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "h2":
		sb.WriteString("## ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "h3":
		sb.WriteString("### ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "h4":
		sb.WriteString("#### ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "h5":
		sb.WriteString("##### ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "h6":
		sb.WriteString("###### ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "p":
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "br":
		sb.WriteString("  \n")
	case "hr":
		sb.WriteString("---\n\n")
	case "strong", "b":
		sb.WriteString(" **")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("** ")
	case "em", "i":
		sb.WriteString(" *")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("* ")
	case "code":
		sb.WriteString(" `")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("` ")
	case "pre":
		sb.WriteString("```\n")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n```\n\n")
	case "a":
		href, _ := elem.Attributes.Get("href")
		sb.WriteString(" [")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("](")
		sb.WriteString(href)
		sb.WriteString(") ")
	case "img":
		alt, _ := elem.Attributes.Get("alt")
		src, _ := elem.Attributes.Get("src")
		sb.WriteString("![")
		sb.WriteString(alt)
		sb.WriteString("](")
		sb.WriteString(src)
		sb.WriteString(")")
	case "ul":
		for _, child := range elem.Children() {
			serializeMarkdown(sb, child, listDepth, true)
		}
		if listDepth == 0 {
			sb.WriteString("\n")
		}
	case "ol":
		index := 1
		for _, child := range elem.Children() {
			if li, ok := child.(*dom.Element); ok && li.TagName == "li" {
				sb.WriteString(strings.Repeat("  ", listDepth))
				sb.WriteString(strconv.Itoa(index))
				sb.WriteString(". ")
				serializeChildrenMarkdown(sb, li, listDepth+1, true)
				sb.WriteString("\n")
				index++
			}
		}
		if listDepth == 0 {
			sb.WriteString("\n")
		}
	case "li":
		// List items are handled by parent ul/ol elements
		sb.WriteString(strings.Repeat("  ", listDepth))
		sb.WriteString("- ")
		serializeChildrenMarkdown(sb, elem, listDepth+1, true)
		sb.WriteString("\n")
	case "blockquote":
		sb.WriteString("> ")
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		sb.WriteString("\n\n")
	case "table", "thead", "tbody", "tr", "th", "td":
		// Basic table support - just extract text for now
		serializeChildrenMarkdown(sb, elem, listDepth, false)
		if elem.TagName == "tr" {
			sb.WriteString("\n")
		}
	case "head", "title", "meta", "link", "script", "style":
		// Skip head elements in markdown
		return
	default:
		// For other elements, just serialize children
		serializeChildrenMarkdown(sb, elem, listDepth, inList)
	}
}

// serializeChildrenMarkdown serializes all children of an element to Markdown.
func serializeChildrenMarkdown(sb *strings.Builder, elem *dom.Element, listDepth int, inList bool) {
	for _, child := range elem.Children() {
		serializeMarkdown(sb, child, listDepth, inList)
	}
}
