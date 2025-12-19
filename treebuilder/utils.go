package treebuilder

import (
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

func (tb *TreeBuilder) hasElementInScope(tagName string, scope map[string]bool) bool {
	return tb.hasElementInScopeInternal(tagName, scope, true)
}

func (tb *TreeBuilder) hasPElementInButtonScope() bool {
	return tb.hasElementInScope("p", constants.ButtonScope)
}

func (tb *TreeBuilder) hasElementInTableScope(tagName string) bool {
	return tb.hasElementInScopeInternal(tagName, constants.TableScope, false)
}

func (tb *TreeBuilder) hasElementInListItemScope(tagName string) bool {
	return tb.hasElementInScope(tagName, constants.ListItemScope)
}

func (tb *TreeBuilder) hasElementInDefinitionScope(tagName string) bool {
	return tb.hasElementInScope(tagName, constants.DefinitionScope)
}

func (tb *TreeBuilder) hasForeignElementOnStack() bool {
	for _, node := range tb.openElements {
		if node.Namespace != dom.NamespaceHTML {
			return true
		}
	}
	return false
}

func (tb *TreeBuilder) hasElementInScopeInternal(tagName string, scope map[string]bool, checkIntegrationPoints bool) bool {
	// Per WHATWG HTML §13.2.5.2.5 (has an element in scope).
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		node := tb.openElements[i]
		if node.Namespace == dom.NamespaceHTML && node.TagName == tagName {
			return true
		}
		if node.Namespace == dom.NamespaceHTML {
			if scope[node.TagName] {
				return false
			}
			continue
		}
		if checkIntegrationPoints && (tb.isHTMLIntegrationPoint(node) || tb.isMathMLTextIntegrationPoint(node)) {
			return false
		}
	}
	return false
}

func (tb *TreeBuilder) hasAnyElementInScope(tagSet map[string]bool, scope map[string]bool) bool {
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		node := tb.openElements[i]
		if node.Namespace == dom.NamespaceHTML && tagSet[node.TagName] {
			return true
		}
		if node.Namespace == dom.NamespaceHTML {
			if scope[node.TagName] {
				return false
			}
			continue
		}
		if tb.isHTMLIntegrationPoint(node) || tb.isMathMLTextIntegrationPoint(node) {
			return false
		}
	}
	return false
}

var headingElements = map[string]bool{
	"h1": true,
	"h2": true,
	"h3": true,
	"h4": true,
	"h5": true,
	"h6": true,
}

func isHeadingElement(tag string) bool {
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

func (tb *TreeBuilder) generateImpliedEndTags(except string) {
	// Per WHATWG HTML §13.2.5.3 (generate implied end tags).
	for len(tb.openElements) > 0 {
		node := tb.currentElement()
		if node == nil || node.Namespace != dom.NamespaceHTML {
			return
		}
		if constants.ImpliedEndTagElements[node.TagName] && node.TagName != except {
			tb.popCurrent()
			continue
		}
		return
	}
}

func (tb *TreeBuilder) clearStackUntil(tagNames map[string]bool) {
	// Per WHATWG HTML §13.2.6.4.9 (clear the stack back to a table context), generalized.
	for len(tb.openElements) > 0 {
		node := tb.currentElement()
		if node == nil {
			return
		}
		if node.Namespace == dom.NamespaceHTML && tagNames[node.TagName] {
			return
		}
		tb.popCurrent()
	}
}

func (tb *TreeBuilder) closeCaptionElement() bool {
	if !tb.hasElementInTableScope("caption") {
		return false
	}
	tb.generateImpliedEndTags("")
	for len(tb.openElements) > 0 {
		node := tb.popCurrent()
		if node.TagName == "caption" {
			break
		}
	}
	tb.clearActiveFormattingUpToMarker()
	tb.mode = InTable
	return true
}

func (tb *TreeBuilder) closeTableCell() bool {
	if !tb.hasElementInTableScope("td") && !tb.hasElementInTableScope("th") {
		return false
	}
	tb.popUntilAnyCell()
	tb.clearActiveFormattingElements()
	tb.mode = InRow
	return true
}

func (tb *TreeBuilder) resetInsertionModeAppropriately() {
	// Per WHATWG HTML §13.2.5.2.4 (reset the insertion mode appropriately).
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		node := tb.openElements[i]
		// Only HTML namespace elements participate in insertion-mode selection.
		// Foreign elements like SVG <tr>/<th> must not switch us into table modes.
		if node.Namespace != dom.NamespaceHTML {
			continue
		}
		switch strings.ToLower(node.TagName) {
		case "select":
			tb.mode = InSelect
			return
		case "td", "th":
			tb.mode = InCell
			return
		case "tr":
			tb.mode = InRow
			return
		case "tbody", "tfoot", "thead":
			tb.mode = InTableBody
			return
		case "caption":
			tb.mode = InCaption
			return
		case "colgroup":
			tb.mode = InColumnGroup
			return
		case "table":
			tb.mode = InTable
			return
		case "template":
			if len(tb.templateModes) > 0 {
				tb.mode = tb.templateModes[len(tb.templateModes)-1]
				return
			}
		case "head":
			tb.mode = InHead
			return
		case "body", "html":
			tb.mode = InBody
			return
		}
	}
	tb.mode = InBody
}

func (tb *TreeBuilder) clearActiveFormattingElements() {
	// Per WHATWG HTML §13.2.5.2.2 (clear the list of active formatting elements up to the last marker).
	tb.clearActiveFormattingUpToMarker()
}

func (tb *TreeBuilder) pushActiveFormattingMarker() {
	// Per WHATWG HTML §13.2.5.2.3 (push a marker onto the list of active formatting elements).
	tb.pushFormattingMarker()
}

func (tb *TreeBuilder) setQuirksModeFromDoctype(name string, publicID, systemID *string, forceQuirks bool) {
	_, mode := doctypeErrorAndQuirks(name, publicID, systemID, forceQuirks, tb.iframeSrcdoc)
	tb.document.QuirksMode = mode
}

func (tb *TreeBuilder) anyOtherEndTag(name string) {
	target := strings.ToLower(name)
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		node := tb.openElements[i]
		if strings.ToLower(node.TagName) == target {
			tb.generateImpliedEndTags(name)
			tb.openElements = tb.openElements[:i]
			return
		}
		if isSpecialElement(node) {
			return
		}
	}
}

func (tb *TreeBuilder) removeFromOpenElements(target *dom.Element) bool {
	for i, el := range tb.openElements {
		if el == target {
			tb.openElements = append(tb.openElements[:i], tb.openElements[i+1:]...)
			return true
		}
	}
	return false
}

func filterWhitespace(data string) string {
	if data == "" {
		return ""
	}
	var sb strings.Builder
	for _, r := range data {
		switch r {
		case '\t', '\n', '\f', '\r', ' ':
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func doctypeErrorAndQuirks(name string, publicID, systemID *string, forceQuirks bool, iframeSrcdoc bool) (bool, dom.QuirksMode) {
	nameLower := strings.ToLower(name)
	public := ptrToString(publicID)
	system := ptrToString(systemID)

	acceptable := map[[3]string]bool{
		{"html", "", ""}:                         true,
		{"html", "", "about:legacy-compat"}:      true,
		{"html", "-//W3C//DTD HTML 4.0//EN", ""}: true,
		{"html", "-//W3C//DTD HTML 4.0//EN", "http://www.w3.org/TR/REC-html40/strict.dtd"}:                true,
		{"html", "-//W3C//DTD HTML 4.01//EN", ""}:                                                         true,
		{"html", "-//W3C//DTD HTML 4.01//EN", "http://www.w3.org/TR/html4/strict.dtd"}:                    true,
		{"html", "-//W3C//DTD XHTML 1.0 Strict//EN", "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"}: true,
		{"html", "-//W3C//DTD XHTML 1.1//EN", "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"}:             true,
	}

	parseError := !acceptable[[3]string{nameLower, public, system}]

	publicLower := strings.ToLower(public)
	systemLower := strings.ToLower(system)

	if forceQuirks {
		return parseError, dom.Quirks
	}
	if iframeSrcdoc {
		return parseError, dom.NoQuirks
	}
	if nameLower != "html" {
		return parseError, dom.Quirks
	}
	if constants.QuirkyPublicMatches[publicLower] {
		return parseError, dom.Quirks
	}
	if constants.QuirkySystemMatches[systemLower] {
		return parseError, dom.Quirks
	}
	if publicLower != "" && hasAnyPrefix(publicLower, constants.QuirkyPublicPrefixes) {
		return parseError, dom.Quirks
	}
	if publicLower != "" && hasAnyPrefix(publicLower, constants.LimitedQuirkyPublicPrefixes) {
		return parseError, dom.LimitedQuirks
	}
	if publicLower != "" && hasAnyPrefix(publicLower, constants.HTML4PublicPrefixes) {
		if systemID == nil {
			return parseError, dom.Quirks
		}
		return parseError, dom.LimitedQuirks
	}
	return parseError, dom.NoQuirks
}

func hasAnyPrefix(needle string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(needle, prefix) {
			return true
		}
	}
	return false
}

func isHiddenInput(attrs []tokenizer.Attr) bool {
	for _, attr := range attrs {
		if attr.Namespace != "" {
			continue
		}
		if strings.EqualFold(attr.Name, "type") && strings.EqualFold(attr.Value, "hidden") {
			return true
		}
	}
	return false
}
