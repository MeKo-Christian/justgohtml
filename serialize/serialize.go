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
	switch n := node.(type) {
	case *dom.Document:
		serializeDocument(sb, n, opts, depth)
	case *dom.DocumentType:
		serializeDoctype(sb, n)
	case *dom.Element:
		serializeElement(sb, n, opts, depth)
	case *dom.Text:
		sb.WriteString(escapeText(n.Data))
	case *dom.Comment:
		sb.WriteString("<!--")
		sb.WriteString(n.Data)
		sb.WriteString("-->")
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

func serializeElement(sb *strings.Builder, elem *dom.Element, opts Options, depth int) {
	if opts.Pretty && depth > 0 {
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

	hasBlockChildren := hasBlockContent(elem)

	for _, child := range elem.Children() {
		if opts.Pretty && hasBlockChildren {
			sb.WriteByte('\n')
		}
		serializeNode(sb, child, opts, depth+1)
	}

	if opts.Pretty && hasBlockChildren && len(elem.Children()) > 0 {
		sb.WriteByte('\n')
		sb.WriteString(strings.Repeat(" ", depth*opts.IndentSize))
	}

	sb.WriteString("</")
	sb.WriteString(elem.TagName)
	sb.WriteByte('>')
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
	case "address", "article", "aside", "blockquote", "canvas", "dd", "div",
		"dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
		"h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "li", "main",
		"nav", "noscript", "ol", "p", "pre", "section", "table", "tfoot",
		"ul", "video":
		return true
	}
	return false
}
