package treebuilder

import (
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

// These handlers are a growing implementation of the HTML5 tree construction
// insertion modes. They focus on correctness and non-panicking behavior.

func (tb *TreeBuilder) processInitial(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			return false
		}
		tb.document.QuirksMode = dom.Quirks
		tb.mode = BeforeHTML
		return true
	case tokenizer.Comment:
		tb.document.AppendChild(dom.NewComment(tok.Data))
		return false
	case tokenizer.DOCTYPE:
		tb.document.Doctype = dom.NewDocumentType(tok.Name, ptrToString(tok.PublicID), ptrToString(tok.SystemID))
		tb.setQuirksModeFromDoctype(tok.Name, tok.PublicID, tok.SystemID, tok.ForceQuirks)
		tb.mode = BeforeHTML
		return false
	default:
		tb.document.QuirksMode = dom.Quirks
		tb.mode = BeforeHTML
		return true
	}
}

func (tb *TreeBuilder) processBeforeHTML(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			return false
		}
		// Strip leading whitespace so that implicit root creation behaves like the spec.
		tok.Data = strings.TrimLeft(tok.Data, "\t\n\f\r ")
	case tokenizer.Comment:
		tb.document.AppendChild(dom.NewComment(tok.Data))
		return false
	case tokenizer.StartTag:
		if tok.Name == "html" {
			tb.insertElement("html", tok.Attrs)
			tb.mode = BeforeHead
			return false
		}
	case tokenizer.EndTag:
		// "head", "body", "html", "br" trigger implicit root creation and reprocess.
		if tok.Name == "head" || tok.Name == "body" || tok.Name == "html" || tok.Name == "br" {
			tb.insertElement("html", nil)
			tb.mode = BeforeHead
			return true
		}
		return false
	case tokenizer.EOF:
		tb.insertElement("html", nil)
		tb.mode = BeforeHead
		return true
	}

	// Create implicit <html> element.
	tb.insertElement("html", nil)
	tb.mode = BeforeHead
	return true
}

func (tb *TreeBuilder) processBeforeHead(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			return false
		}
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			// Duplicate <html>: merge attributes into the existing root.
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case "head":
			tb.headElement = tb.insertElement("head", tok.Attrs)
			tb.mode = InHead
			return false
		}
	case tokenizer.EndTag:
		// Ignore most end tags here.
		return false
	}

	// Implicit <head>.
	tb.headElement = tb.insertElement("head", nil)
	tb.mode = InHead
	return true
}

func (tb *TreeBuilder) processInHead(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.insertText(tok.Data)
			return false
		}
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			// Delegate to InBody rules for attribute merge behavior.
			tb.mode = InBody
			return true
		case "title", "textarea":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			tb.tokenizer.SetState(tokenizer.RCDATAState)
			return false
		case "script", "style", "xmp", "iframe", "noembed", "noframes":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			if tok.Name == "script" {
				tb.tokenizer.SetState(tokenizer.ScriptDataState)
			} else {
				tb.tokenizer.SetState(tokenizer.RAWTEXTState)
			}
			return false
		case "noscript":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.mode = InHeadNoscript
			return false
		case "base", "basefont", "bgsound", "link", "meta":
			// Void-ish head elements; do not stay on stack.
			tb.insertElement(tok.Name, tok.Attrs)
			tb.popCurrent()
			return false
		case "template":
			tb.insertElement("template", tok.Attrs)
			tb.mode = InTemplate
			return false
		case "head":
			// Ignore additional heads.
			return false
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "head":
			tb.popUntil("head")
			tb.mode = AfterHead
			return false
		case "template":
			// If no template element is open, ignore.
			if !tb.elementInStack("template") {
				return false
			}
			tb.popUntil("template")
			tb.mode = InHead
			return false
		}
	case tokenizer.EOF:
		tb.popUntil("head")
		tb.mode = AfterHead
		return true
	}

	// Anything else: close head and reprocess in after head.
	tb.popUntil("head")
	tb.mode = AfterHead
	return true
}

func (tb *TreeBuilder) processInHeadNoscript(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			return tb.processInHead(tok)
		}
		tb.popUntil("noscript")
		tb.mode = InHead
		return true
	case tokenizer.Comment:
		return tb.processInHead(tok)
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			tb.mode = InBody
			return true
		case "basefont", "bgsound", "link", "meta", "noframes", "style":
			return tb.processInHead(tok)
		case "head", "noscript":
			return false
		default:
			tb.popUntil("noscript")
			tb.mode = InHead
			return true
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "noscript":
			tb.popUntil("noscript")
			tb.mode = InHead
			return false
		case "br":
			tb.popUntil("noscript")
			tb.mode = InHead
			return true
		default:
			return false
		}
	case tokenizer.EOF:
		tb.popUntil("noscript")
		tb.mode = InHead
		return true
	default:
		return false
	}
}

func (tb *TreeBuilder) processAfterHead(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.insertText(tok.Data)
			return false
		}
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			tb.mode = InBody
			return true
		case "body":
			tb.insertElement("body", tok.Attrs)
			tb.framesetOK = false
			tb.mode = InBody
			return false
		case "frameset":
			tb.insertElement("frameset", tok.Attrs)
			tb.mode = InFrameset
			return false
		case "head":
			// Parse error; ignore token.
			return false
		}
	case tokenizer.EndTag:
		if tok.Name == "html" {
			tb.mode = InBody
			return true
		}
	case tokenizer.EOF:
		tb.insertElement("body", nil)
		tb.mode = InBody
		return true
	}

	// Implicit <body>.
	tb.insertElement("body", nil)
	tb.framesetOK = false
	tb.mode = InBody
	return true
}

func (tb *TreeBuilder) processText(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		tb.insertText(tok.Data)
		return false
	case tokenizer.EndTag:
		tb.popUntil(tok.Name)
		tb.mode = tb.originalMode
		tb.tokenizer.SetState(tokenizer.DataState)
		return false
	case tokenizer.EOF:
		tb.mode = tb.originalMode
		tb.tokenizer.SetState(tokenizer.DataState)
		return true
	default:
		return false
	}
}

func (tb *TreeBuilder) processInBody(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		tb.reconstructActiveFormattingElements()
		if tok.Data != "" {
			if !isAllWhitespace(tok.Data) {
				tb.framesetOK = false
			}
			tb.insertText(tok.Data)
		}
		return false
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.DOCTYPE:
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case "base", "basefont", "bgsound", "link", "meta":
			// Per spec §13.2.6.4.7: process using the rules for "in head".
			// These are void elements - insert and immediately pop.
			tb.insertElement(tok.Name, tok.Attrs)
			tb.popCurrent()
			return false
		case "body":
			// If a body element already exists, merge attrs.
			if body := tb.document.Body(); body != nil {
				tb.addMissingAttributes(body, tok.Attrs)
				tb.framesetOK = false
				return false
			}
			tb.insertElement("body", tok.Attrs)
			tb.framesetOK = false
			return false
		case "svg":
			tb.reconstructActiveFormattingElements()
			tb.insertForeignElement("svg", dom.NamespaceSVG, prepareForeignAttributes(dom.NamespaceSVG, tok.Attrs), tok.SelfClosing)
			tb.framesetOK = false
			return false
		case "math":
			tb.reconstructActiveFormattingElements()
			tb.insertForeignElement("math", dom.NamespaceMathML, prepareForeignAttributes(dom.NamespaceMathML, tok.Attrs), tok.SelfClosing)
			tb.framesetOK = false
			return false
		case "a":
			if tb.hasActiveFormattingEntry("a") {
				tb.adoptionAgency("a")
				tb.removeLastActiveFormattingByName("a")
				tb.removeLastOpenElementByName("a")
			}
			tb.reconstructActiveFormattingElements()
			node := tb.insertElement("a", tok.Attrs)
			tb.appendActiveFormattingEntry("a", tok.Attrs, node)
			tb.framesetOK = false
			return false
		case "table":
			tb.insertElement("table", tok.Attrs)
			tb.framesetOK = false
			tb.mode = InTable
			return false
		case "select":
			tb.reconstructActiveFormattingElements()
			tb.insertElement("select", tok.Attrs)
			tb.framesetOK = false
			tb.mode = InSelect
			return false
		case "textarea", "title":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			tb.tokenizer.SetState(tokenizer.RCDATAState)
			return false
		case "script", "style":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			if tok.Name == "script" {
				tb.tokenizer.SetState(tokenizer.ScriptDataState)
			} else {
				tb.tokenizer.SetState(tokenizer.RAWTEXTState)
			}
			return false
		case "p":
			if tb.hasElementInScope("p", constants.ButtonScope) {
				tb.popUntil("p")
			}
			tb.reconstructActiveFormattingElements()
			tb.insertElement("p", tok.Attrs)
			tb.framesetOK = false
			return false
		case "br":
			tb.insertElement("br", tok.Attrs)
			tb.popCurrent()
			tb.framesetOK = false
			return false
		}

		if constants.FormattingElements[tok.Name] {
			if tok.Name == "nobr" && tb.hasElementInScope("nobr", constants.DefaultScope) {
				tb.adoptionAgency("nobr")
				tb.removeLastActiveFormattingByName("nobr")
				tb.removeLastOpenElementByName("nobr")
			}
			tb.reconstructActiveFormattingElements()
			if dup, ok := tb.findActiveFormattingDuplicate(tok.Name, tok.Attrs); ok {
				tb.removeFormattingEntry(dup)
			}
			node := tb.insertElement(tok.Name, tok.Attrs)
			tb.appendActiveFormattingEntry(tok.Name, tok.Attrs, node)
			tb.framesetOK = false
			return false
		}

		tb.reconstructActiveFormattingElements()
		el := tb.insertElement(tok.Name, tok.Attrs)
		if tok.SelfClosing || constants.VoidElements[tok.Name] {
			tb.popCurrent()
			_ = el
		} else if tok.Name != "" && !isAllWhitespace(tok.Name) {
			tb.framesetOK = false
		}
		return false
	case tokenizer.EndTag:
		switch tok.Name {
		case "body":
			if tb.hasElementInScope("body", constants.DefaultScope) {
				tb.popUntil("body")
				tb.mode = AfterBody
			}
			return false
		case "html":
			if tb.hasElementInScope("body", constants.DefaultScope) {
				tb.mode = AfterBody
				return true
			}
			return false
		case "p":
			if !tb.hasElementInScope("p", constants.ButtonScope) {
				tb.insertElement("p", nil)
			}
			tb.popUntil("p")
			return false
		default:
			if constants.FormattingElements[tok.Name] {
				tb.adoptionAgency(tok.Name)
				return false
			}
			tb.popUntilCaseInsensitive(tok.Name)
			return false
		}
	case tokenizer.EOF:
		return false
	default:
		return false
	}
}

func (tb *TreeBuilder) processInTable(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		// Switch to "in table text" and reprocess.
		mode := tb.mode
		tb.tableTextOriginalMode = &mode
		tb.pendingTableText = tb.pendingTableText[:0]
		tb.mode = InTableText
		return true
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "caption":
			tb.insertElement("caption", tok.Attrs)
			tb.mode = InCaption
			return false
		case "colgroup":
			tb.insertElement("colgroup", tok.Attrs)
			tb.mode = InColumnGroup
			return false
		case "tbody", "thead", "tfoot":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.mode = InTableBody
			return false
		case "tr":
			tb.insertElement("tbody", nil)
			tb.mode = InTableBody
			return true
		case "td", "th":
			tb.insertElement("tbody", nil)
			tb.mode = InTableBody
			return true
		case "table":
			tb.popUntil("table")
			tb.mode = InBody
			return true
		case "select":
			tb.insertElement("select", tok.Attrs)
			tb.mode = InSelectInTable
			return false
		case "template":
			tb.insertElement("template", tok.Attrs)
			tb.mode = InTemplate
			return false
		}
		// Default: foster parent character data handling is not implemented yet,
		// so fall back to InBody processing.
		tb.mode = InBody
		return true
	case tokenizer.EndTag:
		switch tok.Name {
		case "table":
			tb.popUntil("table")
			tb.mode = InBody
			return false
		case "body", "caption", "col", "colgroup", "html", "tbody", "tfoot", "thead", "tr", "td", "th":
			return false
		}
	case tokenizer.EOF:
		return false
	}
	return false
}

func (tb *TreeBuilder) processInTableText(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		tb.pendingTableText = append(tb.pendingTableText, tok.Data)
		return false
	default:
		// Flush pending table text.
		for _, s := range tb.pendingTableText {
			if isAllWhitespace(s) {
				tb.insertText(s)
			} else {
				tb.insertFosterText(s)
			}
		}
		tb.pendingTableText = tb.pendingTableText[:0]
		if tb.tableTextOriginalMode != nil {
			tb.mode = *tb.tableTextOriginalMode
			tb.tableTextOriginalMode = nil
		} else {
			tb.mode = InTable
		}
		return true
	}
}

func (tb *TreeBuilder) processInCaption(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.EndTag:
		if tok.Name == "caption" {
			tb.popUntil("caption")
			tb.mode = InTable
			return false
		}
		if tok.Name == "table" {
			tb.popUntil("caption")
			tb.mode = InTable
			return true
		}
	case tokenizer.StartTag:
		if tok.Name == "table" {
			tb.popUntil("caption")
			tb.mode = InTable
			return true
		}
	}
	tb.mode = InBody
	return true
}

func (tb *TreeBuilder) processInColumnGroup(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.insertText(tok.Data)
			return false
		}
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "col":
			tb.insertElement("col", tok.Attrs)
			tb.popCurrent()
			return false
		case "template":
			tb.insertElement("template", tok.Attrs)
			tb.mode = InTemplate
			return false
		}
	case tokenizer.EndTag:
		if tok.Name == "colgroup" {
			tb.popUntil("colgroup")
			tb.mode = InTable
			return false
		}
	case tokenizer.EOF:
		return false
	}

	// Close colgroup and reprocess in table.
	tb.popUntil("colgroup")
	tb.mode = InTable
	return true
}

func (tb *TreeBuilder) processInTableBody(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.StartTag:
		switch tok.Name {
		case "tr":
			tb.insertElement("tr", tok.Attrs)
			tb.mode = InRow
			return false
		case "td", "th":
			tb.insertElement("tr", nil)
			tb.mode = InRow
			return true
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "tbody", "thead", "tfoot":
			tb.popUntil(tok.Name)
			tb.mode = InTable
			return false
		case "table":
			tb.popUntil("tbody")
			tb.mode = InTable
			return true
		}
	}
	tb.mode = InTable
	return true
}

func (tb *TreeBuilder) processInRow(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.StartTag:
		if tok.Name == "td" || tok.Name == "th" {
			tb.insertElement(tok.Name, tok.Attrs)
			tb.mode = InCell
			return false
		}
		if tok.Name == "tr" {
			tb.popUntil("tr")
			tb.mode = InTableBody
			return true
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "tr":
			tb.popUntil("tr")
			tb.mode = InTableBody
			return false
		case "table":
			tb.popUntil("tr")
			tb.mode = InTableBody
			return true
		}
	}
	tb.mode = InTableBody
	return true
}

func (tb *TreeBuilder) processInCell(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.EndTag:
		if tok.Name == "td" || tok.Name == "th" {
			tb.popUntil(tok.Name)
			tb.mode = InRow
			return false
		}
		if tok.Name == "tr" || tok.Name == "table" {
			tb.popUntilAnyCell()
			tb.mode = InRow
			return true
		}
	case tokenizer.StartTag:
		if tok.Name == "td" || tok.Name == "th" {
			tb.popUntilAnyCell()
			tb.mode = InRow
			return true
		}
	}
	tb.mode = InBody
	return true
}

func (tb *TreeBuilder) popUntilAnyCell() {
	for len(tb.openElements) > 0 {
		name := tb.currentElement().TagName
		tb.popCurrent()
		if name == "td" || name == "th" {
			return
		}
	}
}

func (tb *TreeBuilder) processInSelect(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		tb.insertText(tok.Data)
		return false
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			tb.mode = InBody
			return true
		case "option":
			// If current node is option, pop it.
			if tb.currentElement() != nil && tb.currentElement().TagName == "option" {
				tb.popCurrent()
			}
			tb.insertElement("option", tok.Attrs)
			return false
		case "optgroup":
			if tb.currentElement() != nil && tb.currentElement().TagName == "option" {
				tb.popCurrent()
			}
			if tb.currentElement() != nil && tb.currentElement().TagName == "optgroup" {
				tb.popCurrent()
			}
			tb.insertElement("optgroup", tok.Attrs)
			return false
		case "select":
			// Close the current select.
			tb.popUntil("select")
			tb.mode = InBody
			return true
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "option":
			if tb.currentElement() != nil && tb.currentElement().TagName == "option" {
				tb.popCurrent()
			}
			return false
		case "optgroup":
			if tb.currentElement() != nil && tb.currentElement().TagName == "option" {
				tb.popCurrent()
			}
			if tb.currentElement() != nil && tb.currentElement().TagName == "optgroup" {
				tb.popCurrent()
			}
			return false
		case "select":
			tb.popUntil("select")
			tb.mode = InBody
			return false
		}
	case tokenizer.EOF:
		return false
	}
	return false
}

func (tb *TreeBuilder) processInSelectInTable(tok tokenizer.Token) bool {
	// If we see a table-affecting token, pop select and reprocess.
	if tok.Type == tokenizer.StartTag {
		switch tok.Name {
		case "caption", "table", "tbody", "tfoot", "thead", "tr", "td", "th":
			tb.popUntil("select")
			tb.mode = InTable
			return true
		}
	}
	if tok.Type == tokenizer.EndTag {
		switch tok.Name {
		case "caption", "table", "tbody", "tfoot", "thead", "tr", "td", "th":
			tb.popUntil("select")
			tb.mode = InTable
			return true
		}
	}
	return tb.processInSelect(tok)
}

func (tb *TreeBuilder) processInTemplate(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.EndTag:
		if tok.Name == "template" {
			tb.popUntil("template")
			tb.mode = InHead
			return false
		}
	case tokenizer.EOF:
		return false
	}
	// For now, treat template contents like "in body".
	tb.mode = InBody
	return true
}

func (tb *TreeBuilder) processAfterBody(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.mode = InBody
			return true
		}
	case tokenizer.Comment:
		// Comments after body attach to the <html> element.
		if len(tb.openElements) > 0 {
			tb.openElements[0].AppendChild(dom.NewComment(tok.Data))
		} else {
			tb.document.AppendChild(dom.NewComment(tok.Data))
		}
		return false
	case tokenizer.StartTag:
		if tok.Name == "html" {
			tb.mode = InBody
			return true
		}
	case tokenizer.EndTag:
		if tok.Name == "html" {
			tb.mode = AfterAfterBody
			return false
		}
	case tokenizer.EOF:
		return false
	}
	tb.mode = InBody
	return true
}

func (tb *TreeBuilder) processInFrameset(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.insertText(tok.Data)
		}
		return false
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "frameset":
			tb.insertElement("frameset", tok.Attrs)
			return false
		case "frame":
			tb.insertElement("frame", tok.Attrs)
			tb.popCurrent()
			return false
		case "noframes":
			tb.mode = InBody
			return true
		}
	case tokenizer.EndTag:
		if tok.Name == "frameset" {
			tb.popUntil("frameset")
			if !tb.elementInStack("frameset") {
				tb.mode = AfterFrameset
			}
			return false
		}
	case tokenizer.EOF:
		return false
	}
	return false
}

func (tb *TreeBuilder) processAfterFrameset(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.insertText(tok.Data)
		}
		return false
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		if tok.Name == "html" {
			tb.mode = InBody
			return true
		}
		if tok.Name == "noframes" {
			tb.mode = InBody
			return true
		}
	case tokenizer.EndTag:
		if tok.Name == "html" {
			tb.mode = AfterAfterFrameset
			return false
		}
	case tokenizer.EOF:
		return false
	}
	return false
}

func (tb *TreeBuilder) processAfterAfterBody(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Comment:
		tb.document.AppendChild(dom.NewComment(tok.Data))
		return false
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.mode = InBody
			return true
		}
	case tokenizer.StartTag:
		if tok.Name == "html" {
			tb.mode = InBody
			return true
		}
	case tokenizer.EOF:
		return false
	}
	tb.mode = InBody
	return true
}

func (tb *TreeBuilder) processAfterAfterFrameset(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Comment:
		tb.document.AppendChild(dom.NewComment(tok.Data))
		return false
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.mode = InFrameset
			return true
		}
	case tokenizer.StartTag:
		if tok.Name == "html" {
			tb.mode = InBody
			return true
		}
	case tokenizer.EOF:
		return false
	}
	tb.mode = InBody
	return true
}
