package treebuilder

import (
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

func (tb *TreeBuilder) shouldUseForeignContent(tok tokenizer.Token) bool {
	current := tb.currentElement()
	if current == nil {
		return false
	}
	if current.Namespace == dom.NamespaceHTML {
		return false
	}
	if tok.Type == tokenizer.EOF {
		return false
	}

	// MathML text integration points.
	if tb.isMathMLTextIntegrationPoint(current) {
		if tok.Type == tokenizer.Character {
			return false
		}
		if tok.Type == tokenizer.StartTag {
			if tok.Name != "mglyph" && tok.Name != "malignmark" {
				return false
			}
		}
	}

	// MathML annotation-xml special-case.
	if current.Namespace == dom.NamespaceMathML && strings.EqualFold(current.TagName, "annotation-xml") {
		if tok.Type == tokenizer.StartTag && tok.Name == "svg" {
			return false
		}
	}

	// HTML integration points.
	if tb.isHTMLIntegrationPoint(current) {
		if tok.Type == tokenizer.Character {
			return false
		}
		if tok.Type == tokenizer.StartTag {
			return false
		}
	}

	return true
}

func (tb *TreeBuilder) processForeignContent(tok tokenizer.Token) bool {
	current := tb.currentElement()
	if current == nil {
		return false
	}

	switch tok.Type {
	case tokenizer.Character:
		if tok.Data == "" {
			return false
		}
		data := strings.ReplaceAll(tok.Data, "\x00", string('\uFFFD'))
		if !isAllWhitespace(data) {
			tb.framesetOK = false
		}
		tb.insertText(data)
		return false
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		nameLower := tok.Name
		if constants.ForeignBreakoutElements[nameLower] || (nameLower == "font" && foreignBreakoutFont(tok.Attrs)) {
			tb.popUntilHTMLOrIntegrationPoint()
			tb.resetInsertionModeAppropriately()
			tb.forceHTMLMode = true
			return true
		}

		namespace := current.Namespace
		adjustedName := tok.Name
		if namespace == dom.NamespaceSVG {
			adjustedName = adjustSVGTagName(tok.Name)
		}
		attrs := prepareForeignAttributes(namespace, tok.Attrs)
		tb.insertForeignElement(adjustedName, namespace, attrs, tok.SelfClosing)
		return false
	case tokenizer.EndTag:
		nameLower := tok.Name
		if nameLower == "br" || nameLower == "p" {
			tb.popUntilHTMLOrIntegrationPoint()
			tb.resetInsertionModeAppropriately()
			tb.forceHTMLMode = true
			return true
		}

		// Walk stack backwards looking for a matching element (ASCII case-insensitive).
		// Per WHATWG HTML §13.2.6.5 (parsing main foreign content), end tag handling.
		for i := len(tb.openElements) - 1; i >= 0; i-- {
			node := tb.openElements[i]
			isHTML := node.Namespace == dom.NamespaceHTML

			if strings.EqualFold(node.TagName, nameLower) {
				if tb.fragmentElement != nil && node == tb.fragmentElement {
					return false
				}
				// If the matched element is in HTML namespace, reprocess using
				// the current insertion mode (which will handle the end tag).
				if isHTML {
					tb.forceHTMLMode = true
					return true
				}
				// Foreign element - pop everything from this point up.
				tb.openElements = tb.openElements[:i]
				return false
			}

			// If we hit an HTML element that doesn't match, reprocess using
			// the current insertion mode.
			if isHTML {
				tb.forceHTMLMode = true
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (tb *TreeBuilder) popUntilHTMLOrIntegrationPoint() {
	for len(tb.openElements) > 0 {
		node := tb.currentElement()
		if node == nil {
			return
		}
		if node.Namespace == dom.NamespaceHTML {
			return
		}
		if tb.isHTMLIntegrationPoint(node) {
			return
		}
		tb.popCurrent()
	}
}

func (tb *TreeBuilder) isHTMLIntegrationPoint(node *dom.Element) bool {
	if node == nil {
		return false
	}
	// annotation-xml only counts with certain encoding values.
	if node.Namespace == dom.NamespaceMathML && node.TagName == "annotation-xml" {
		if enc, ok := node.Attributes.Get("encoding"); ok {
			switch strings.ToLower(enc) {
			case "text/html", "application/xhtml+xml":
				return true
			default:
				return false
			}
		}
		return false
	}
	ip := constants.IntegrationPoint{Namespace: node.Namespace, LocalName: node.TagName}
	return constants.HTMLIntegrationPoints[ip]
}

func (tb *TreeBuilder) isMathMLTextIntegrationPoint(node *dom.Element) bool {
	if node == nil {
		return false
	}
	ip := constants.IntegrationPoint{Namespace: node.Namespace, LocalName: node.TagName}
	return constants.MathMLTextIntegrationPoints[ip]
}

func foreignBreakoutFont(attrs map[string]string) bool {
	for k := range attrs {
		switch strings.ToLower(k) {
		case "color", "face", "size":
			return true
		}
	}
	return false
}

func adjustSVGTagName(name string) string {
	if adjusted, ok := constants.SVGTagNameAdjustments[strings.ToLower(name)]; ok {
		return adjusted
	}
	return name
}

func prepareForeignAttributes(namespace string, attrs map[string]string) []dom.Attribute {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]dom.Attribute, 0, len(attrs))
	for name, value := range attrs {
		lower := strings.ToLower(name)
		adjustedName := name

		switch namespace {
		case dom.NamespaceMathML:
			if adj, ok := constants.MathMLAttributeAdjustments[lower]; ok {
				adjustedName = adj
				lower = strings.ToLower(adjustedName)
			}
		case dom.NamespaceSVG:
			if adj, ok := constants.SVGAttributeAdjustments[lower]; ok {
				adjustedName = adj
				lower = strings.ToLower(adjustedName)
			}
		}

		if foreignAdj, ok := constants.ForeignAttributeAdjustments[lower]; ok {
			prefix := foreignAdj.Prefix
			local := foreignAdj.LocalName
			if prefix != "" {
				adjustedName = prefix + ":" + local
			} else {
				adjustedName = local
			}
			out = append(out, dom.Attribute{Namespace: foreignAdj.NamespaceURL, Name: adjustedName, Value: value})
			continue
		}

		out = append(out, dom.Attribute{Name: adjustedName, Value: value})
	}
	return out
}

func (tb *TreeBuilder) insertForeignElement(name, namespace string, attrs []dom.Attribute, selfClosing bool) *dom.Element {
	el := dom.NewElementNS(name, namespace)
	for _, a := range attrs {
		el.Attributes.SetNS(a.Namespace, a.Name, a.Value)
	}
	tb.currentNode().AppendChild(el)
	if !selfClosing {
		tb.openElements = append(tb.openElements, el)
	}
	return el
}
