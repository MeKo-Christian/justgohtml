package testutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// SerializeHTML5LibTree serializes a parsed document to the html5lib tree-construction
// test "document" format.
//
// Format reference: https://github.com/html5lib/html5lib-tests
func SerializeHTML5LibTree(doc *dom.Document) string {
	var sb strings.Builder

	if doc.Doctype != nil {
		sb.WriteString("| <!DOCTYPE ")
		if doc.Doctype.Name == "" {
			sb.WriteString(">")
		} else {
			sb.WriteString(doc.Doctype.Name)
			if doc.Doctype.PublicID != "" || doc.Doctype.SystemID != "" {
				sb.WriteString(" \"")
				sb.WriteString(doc.Doctype.PublicID)
				sb.WriteString("\" \"")
				sb.WriteString(doc.Doctype.SystemID)
				sb.WriteString("\">")
			} else {
				sb.WriteString(">")
			}
		}
		sb.WriteByte('\n')
	}

	sb.WriteString(SerializeHTML5LibNodes(doc.Children()))

	return strings.TrimRight(sb.String(), "\n")
}

// SerializeHTML5LibNodes serializes a list of nodes using the html5lib tree-construction
// test format (used for document fragments).
func SerializeHTML5LibNodes(nodes []dom.Node) string {
	var sb strings.Builder
	for _, child := range nodes {
		serializeHTML5LibNode(&sb, child, 0)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func serializeHTML5LibNode(sb *strings.Builder, node dom.Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *dom.Element:
		sb.WriteString("| ")
		sb.WriteString(indent)
		sb.WriteString("<")
		sb.WriteString(formatHTML5LibTagName(n))
		sb.WriteString(">")
		sb.WriteByte('\n')

		attrs := n.Attributes.All()
		sort.Slice(attrs, func(i, j int) bool {
			return formatHTML5LibAttributeName(attrs[i]) < formatHTML5LibAttributeName(attrs[j])
		})
		for _, attr := range attrs {
			sb.WriteString("| ")
			sb.WriteString(indent)
			sb.WriteString("  ")
			sb.WriteString(formatHTML5LibAttributeName(attr))
			sb.WriteString("=\"")
			sb.WriteString(escapeHTML5LibString(attr.Value))
			sb.WriteString("\"")
			sb.WriteByte('\n')
		}

		if n.TemplateContent != nil {
			sb.WriteString("| ")
			sb.WriteString(strings.Repeat("  ", depth+1))
			sb.WriteString("content")
			sb.WriteByte('\n')
			for _, child := range n.TemplateContent.Children() {
				serializeHTML5LibNode(sb, child, depth+2)
			}
		}

		for _, child := range n.Children() {
			serializeHTML5LibNode(sb, child, depth+1)
		}

	case *dom.Text:
		sb.WriteString("| ")
		sb.WriteString(indent)
		sb.WriteString("\"")
		sb.WriteString(escapeHTML5LibString(n.Data))
		sb.WriteString("\"")
		sb.WriteByte('\n')

	case *dom.Comment:
		sb.WriteString("| ")
		sb.WriteString(indent)
		sb.WriteString("<!-- ")
		sb.WriteString(n.Data)
		sb.WriteString(" -->")
		sb.WriteByte('\n')

	case *dom.DocumentType:
		// DocumentType nodes are represented via doc.Doctype; ignore here.
		return

	default:
		// Unknown node types are ignored in this representation.
		return
	}
}

func formatHTML5LibTagName(el *dom.Element) string {
	switch el.Namespace {
	case "", dom.NamespaceHTML:
		return el.TagName
	case dom.NamespaceSVG:
		return "svg " + el.TagName
	case dom.NamespaceMathML:
		return "math " + el.TagName
	default:
		// If we ever end up with an unexpected namespace, keep the output stable
		// and obvious rather than silently discarding the namespace information.
		return fmt.Sprintf("%s %s", el.Namespace, el.TagName)
	}
}

func formatHTML5LibAttributeName(attr dom.Attribute) string {
	var designator string
	switch attr.Namespace {
	case "":
		designator = ""
	case "http://www.w3.org/1999/xlink":
		designator = "xlink "
	case "http://www.w3.org/XML/1998/namespace":
		designator = "xml "
	case "http://www.w3.org/2000/xmlns/":
		designator = "xmlns "
	default:
		// Unknown namespace - keep it explicit (and test-visible).
		designator = attr.Namespace + " "
	}

	if designator == "" {
		return attr.Name
	}

	local := attr.Name
	if idx := strings.IndexByte(local, ':'); idx >= 0 {
		local = local[idx+1:]
	}
	return designator + local
}

func escapeHTML5LibString(s string) string {
	return s
}
