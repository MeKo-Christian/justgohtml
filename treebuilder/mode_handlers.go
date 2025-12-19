package treebuilder

import (
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

const (
	tagBase     = "base"
	tagBasefont = "basefont"
	tagBgsound  = "bgsound"
	tagLink     = "link"
	tagMeta     = "meta"
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
	case tokenizer.Error, tokenizer.StartTag, tokenizer.EndTag, tokenizer.EOF:
		tb.document.QuirksMode = dom.Quirks
		tb.mode = BeforeHTML
		return true
	}
	return false
}

func (tb *TreeBuilder) processBeforeHTML(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			return false
		}
		// Strip leading whitespace so that implicit root creation behaves like the spec.
		tok.Data = strings.TrimLeft(tok.Data, "\t\n\f\r ")
		if tok.Data == "" {
			return false
		}
		tb.insertElement("html", nil)
		tb.mode = BeforeHead
		tb.ProcessToken(tok)
		return false
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
	case tokenizer.Error, tokenizer.DOCTYPE:
		// Fall through to implicit <html> creation
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
	case tokenizer.Error, tokenizer.DOCTYPE, tokenizer.EOF:
		// Fall through to implicit <head> creation
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
			// Per "in body" insertion mode: merge attributes into the existing root.
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case "title":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			tb.tokenizer.SetState(tokenizer.RCDATAState)
			return false
		case "script", "style", "noframes":
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
		case tagBase, tagBasefont, tagBgsound, tagLink, tagMeta:
			// Void-ish head elements; do not stay on stack.
			tb.insertElement(tok.Name, tok.Attrs)
			tb.popCurrent()
			return false
		case "template":
			tb.insertElement("template", tok.Attrs)
			tb.pushActiveFormattingMarker()
			tb.framesetOK = false
			tb.mode = InTemplate
			tb.templateModes = append(tb.templateModes, InTemplate)
			return false
		case "head":
			// Ignore additional heads.
			return false
		}
		tb.popUntil("head")
		tb.mode = AfterHead
		return true
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
			tb.generateImpliedEndTags("")
			tb.popUntil("template")
			tb.clearActiveFormattingElements()
			if len(tb.templateModes) > 0 {
				tb.templateModes = tb.templateModes[:len(tb.templateModes)-1]
			}
			tb.resetInsertionModeAppropriately()
			return false
		case "body", "html", "br":
			tb.popUntil("head")
			tb.mode = AfterHead
			return true
		}
	case tokenizer.EOF:
		tb.popUntil("head")
		tb.mode = AfterHead
		return true
	case tokenizer.Error, tokenizer.DOCTYPE:
		// Fall through to implicit head closure
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
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr", "td", "th":
			// Table-structure elements are ignored in "in body".
			return false
		case "html":
			if len(tb.templateModes) > 0 {
				return false
			}
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case tagBasefont, tagBgsound, tagLink, tagMeta, "noframes", "style":
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
	case tokenizer.Error, tokenizer.DOCTYPE:
		return false
	}
	return false
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
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr", "td", "th":
			// Table-structure elements are ignored in "in body".
			return false
		case "html":
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case "body":
			tb.insertElement("body", tok.Attrs)
			tb.framesetOK = false
			tb.mode = InBody
			return false
		case "frameset":
			tb.insertElement("frameset", tok.Attrs)
			tb.mode = InFrameset
			return false
		case tagBase, tagBasefont, tagBgsound, tagLink, tagMeta, "noframes", "script", "style", "title", "noscript":
			if tb.headElement != nil {
				tb.openElements = append(tb.openElements, tb.headElement)
			}
			reprocess := tb.processInHead(tok)
			for i := len(tb.openElements) - 1; i >= 0; i-- {
				if tb.openElements[i] == tb.headElement {
					tb.openElements = append(tb.openElements[:i], tb.openElements[i+1:]...)
					break
				}
			}
			return reprocess
		case "template":
			if tb.headElement != nil {
				tb.openElements = append(tb.openElements, tb.headElement)
			}
			tb.mode = InHead
			return true
		case "head":
			// Parse error; ignore token.
			return false
		}
	case tokenizer.EndTag:
		if tok.Name == "body" || tok.Name == "html" || tok.Name == "br" {
			// Act as if a start tag "body" was seen, then reprocess.
			tb.insertElement("body", nil)
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
	case tokenizer.Error, tokenizer.DOCTYPE, tokenizer.StartTag, tokenizer.Comment:
		return false
	}
	return false
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
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr", "td", "th":
			// Table-structure elements are ignored in "in body".
			return false
		case "html":
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case "address", "article", "aside", "blockquote", "center", "details", "dialog", "dir", "div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "main", "menu", "nav", "ol", "section", "summary", "ul":
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			tb.insertElement(tok.Name, tok.Attrs)
			tb.framesetOK = false
			return false
		case "h1", "h2", "h3", "h4", "h5", "h6":
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			// Close any open heading element to avoid nested headings.
			for _, el := range tb.openElements {
				if el.TagName == "h1" || el.TagName == "h2" || el.TagName == "h3" || el.TagName == "h4" || el.TagName == "h5" || el.TagName == "h6" {
					tb.popUntil(el.TagName)
					break
				}
			}
			tb.insertElement(tok.Name, tok.Attrs)
			tb.framesetOK = false
			return false
		case "pre":
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			tb.insertElement(tok.Name, tok.Attrs)
			tb.framesetOK = false
			return false
		case "listing":
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			tb.insertElement(tok.Name, tok.Attrs)
			tb.framesetOK = false
			return false
		case tagBase, tagBasefont, tagBgsound, tagLink, tagMeta:
			// Per spec §13.2.6.4.7: process using the rules for "in head".
			// These are void elements - insert and immediately pop.
			tb.insertElement(tok.Name, tok.Attrs)
			tb.popCurrent()
			return false
		case "template":
			return tb.processInHead(tok)
		case "frameset":
			if !tb.framesetOK {
				return false
			}
			// Pop everything up to, but not including, the html element.
			for len(tb.openElements) > 0 && tb.currentElement().TagName != "html" {
				tb.popCurrent()
			}
			tb.insertElement("frameset", tok.Attrs)
			tb.mode = InFrameset
			return false
		case "frame":
			if tb.fragmentContext != nil {
				tb.insertElement("frame", tok.Attrs)
				tb.popCurrent()
			}
			return false
		case "body":
			if len(tb.templateModes) > 0 {
				return false
			}
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
		case "rp", "rt":
			tb.generateImpliedEndTags("rtc")
			tb.insertElement(tok.Name, tok.Attrs)
			return false
		case "rb", "rtc":
			if tb.currentElement() != nil {
				switch tb.currentElement().TagName {
				case "rb", "rp", "rt", "rtc":
					tb.generateImpliedEndTags("")
				}
			}
			tb.insertElement(tok.Name, tok.Attrs)
			return false
		case "textarea", "title":
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			tb.tokenizer.SetState(tokenizer.RCDATAState)
			return false
		case "xmp":
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			tb.insertElement(tok.Name, tok.Attrs)
			tb.originalMode = tb.mode
			tb.mode = Text
			tb.tokenizer.SetLastStartTag(tok.Name)
			tb.tokenizer.SetState(tokenizer.RAWTEXTState)
			tb.framesetOK = false
			return false
		case "plaintext":
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			tb.insertElement(tok.Name, tok.Attrs)
			tb.tokenizer.SetLastStartTag(tok.Name)
			tb.tokenizer.SetState(tokenizer.PLAINTEXTState)
			tb.framesetOK = false
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
			if tb.hasPElementInButtonScope() {
				tb.popUntil("p")
			}
			tb.insertElement("p", tok.Attrs)
			return false
		case "image":
			tb.insertElement("img", tok.Attrs)
			tb.popCurrent()
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
				tb.mode = AfterBody
			}
			return false
		case "html":
			if tb.hasElementInScope("body", constants.DefaultScope) {
				tb.mode = AfterBody
				return true
			}
			return false
		case "br":
			tb.insertElement("br", nil)
			tb.popCurrent()
			tb.framesetOK = false
			return false
		case "p":
			if !tb.hasPElementInButtonScope() {
				tb.insertElement("p", nil)
			}
			tb.popUntil("p")
			return false
		case "template":
			if !tb.elementInStack("template") {
				return false
			}
			tb.generateImpliedEndTags("")
			tb.popUntil("template")
			tb.clearActiveFormattingElements()
			if len(tb.templateModes) > 0 {
				tb.templateModes = tb.templateModes[:len(tb.templateModes)-1]
			}
			tb.resetInsertionModeAppropriately()
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
		if len(tb.templateModes) > 0 {
			return tb.processInTemplate(tok)
		}
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
			tb.clearStackUntil(map[string]bool{"table": true, "template": true, "html": true})
			tb.insertElement("caption", tok.Attrs)
			tb.mode = InCaption
			return false
		case "col":
			tb.clearStackUntil(map[string]bool{"table": true, "template": true, "html": true})
			tb.insertElement("colgroup", nil)
			tb.mode = InColumnGroup
			return true
		case "colgroup":
			tb.clearStackUntil(map[string]bool{"table": true, "template": true, "html": true})
			tb.insertElement("colgroup", tok.Attrs)
			tb.mode = InColumnGroup
			return false
		case "tbody", "thead", "tfoot":
			tb.clearStackUntil(map[string]bool{"table": true, "template": true, "html": true})
			tb.insertElement(tok.Name, tok.Attrs)
			tb.mode = InTableBody
			return false
		case "tr":
			tb.clearStackUntil(map[string]bool{"table": true, "template": true, "html": true})
			tb.insertElement("tbody", nil)
			tb.mode = InTableBody
			return true
		case "td", "th":
			tb.clearStackUntil(map[string]bool{"table": true, "template": true, "html": true})
			tb.insertElement("tbody", nil)
			tb.mode = InTableBody
			return true
		case "table":
			if !tb.hasElementInTableScope("table") {
				return false
			}
			tb.popUntil("table")
			tb.mode = InBody
			return true
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
		case "template":
			return tb.processInHead(tok)
		}
		// Default: parse error; foster parent and process using "in body" rules.
		return tb.withFosterParenting(func() bool {
			return tb.processInBody(tok)
		})
	case tokenizer.EndTag:
		switch tok.Name {
		case "table":
			if !tb.hasElementInTableScope("table") {
				return false
			}
			tb.popUntil("table")
			tb.mode = InBody
			return false
		case "body", "caption", "col", "colgroup", "html", "tbody", "tfoot", "thead", "tr", "td", "th":
			return false
		default:
			// Default: parse error; foster parent and process using "in body" rules.
			return tb.withFosterParenting(func() bool {
				return tb.processInBody(tok)
			})
		}
	case tokenizer.EOF:
		if len(tb.templateModes) > 0 {
			return tb.processInTemplate(tok)
		}
		return false
	case tokenizer.Error:
		return false
	}
	return false
}

func (tb *TreeBuilder) processInTableText(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if strings.ContainsRune(tok.Data, 0) {
			tok.Data = strings.ReplaceAll(tok.Data, "\x00", "")
			if tok.Data == "" {
				return false
			}
		}
		tb.pendingTableText = append(tb.pendingTableText, tok.Data)
		return false
	case tokenizer.Error, tokenizer.DOCTYPE, tokenizer.StartTag, tokenizer.EndTag, tokenizer.Comment, tokenizer.EOF:
		// Flush pending table text.
		for _, s := range tb.pendingTableText {
			if isAllWhitespace(s) {
				tb.insertText(s)
			} else {
				_ = tb.withFosterParenting(func() bool {
					tb.reconstructActiveFormattingElements()
					tb.insertText(s)
					return false
				})
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
	return false
}

func (tb *TreeBuilder) processInCaption(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		return tb.processInBody(tok)
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
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
		switch tok.Name {
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr", "td", "th":
			tb.popUntil("caption")
			tb.mode = InTable
			return true
		}
		if tok.Name == "table" {
			tb.popUntil("caption")
			tb.mode = InTable
			return true
		}
	case tokenizer.Error, tokenizer.DOCTYPE, tokenizer.EOF:
		// Fall through to processInBody
	}
	return tb.processInBody(tok)
}

func (tb *TreeBuilder) processInColumnGroup(tok tokenizer.Token) bool {
	current := tb.currentElement()
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.insertText(tok.Data)
			return false
		}
		if current != nil && current.TagName == "template" {
			return false
		}
		if current != nil && current.TagName == "html" {
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
			return tb.processInHead(tok)
		case "colgroup":
			if current != nil && current.TagName == "colgroup" {
				tb.popUntil("colgroup")
				tb.mode = InTable
				return true
			}
			return false
		}
	case tokenizer.EndTag:
		if tok.Name == "colgroup" {
			if current != nil && current.TagName == "colgroup" {
				tb.popUntil("colgroup")
				tb.mode = InTable
			}
			return false
		}
		if tok.Name == "col" {
			return false
		}
		if tok.Name == "template" {
			return tb.processInHead(tok)
		}
	case tokenizer.EOF:
		if current != nil && current.TagName == "colgroup" {
			tb.popUntil("colgroup")
			tb.mode = InTable
			return true
		}
		if current != nil && current.TagName == "template" {
			return tb.processInTemplate(tok)
		}
		return false
	case tokenizer.Error, tokenizer.DOCTYPE:
		// Fall through to implicit colgroup closure
	}

	// Close colgroup and reprocess in table.
	if current != nil && current.TagName == "colgroup" {
		tb.popUntil("colgroup")
		tb.mode = InTable
		return true
	}
	if current != nil && current.TagName == "template" {
		return false
	}
	if current != nil && current.TagName == "html" {
		return false
	}
	tb.mode = InTable
	return true
}

func (tb *TreeBuilder) processInTableBody(tok tokenizer.Token) bool {
	current := tb.currentElement()
	switch tok.Type {
	case tokenizer.Character, tokenizer.Comment:
		return tb.processInTable(tok)
	case tokenizer.StartTag:
		switch tok.Name {
		case "tr":
			tb.clearStackUntil(map[string]bool{"tbody": true, "tfoot": true, "thead": true, "template": true, "html": true})
			tb.insertElement("tr", tok.Attrs)
			tb.mode = InRow
			return false
		case "td", "th":
			tb.clearStackUntil(map[string]bool{"tbody": true, "tfoot": true, "thead": true, "template": true, "html": true})
			tb.insertElement("tr", nil)
			tb.mode = InRow
			return true
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "table":
			if current != nil && current.TagName == "template" {
				return false
			}
			if current != nil && (current.TagName == "tbody" || current.TagName == "thead" || current.TagName == "tfoot") {
				tb.popCurrent()
			}
			tb.mode = InTable
			return true
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "tbody", "thead", "tfoot":
			if current != nil && current.TagName == "template" {
				return false
			}
			if !tb.hasElementInTableScope(tok.Name) {
				return false
			}
			tb.popUntil(tok.Name)
			tb.mode = InTable
			return false
		case "table":
			if current != nil && current.TagName == "template" {
				return false
			}
			if !tb.hasElementInTableScope("table") {
				return false
			}
			tb.popUntil("tbody")
			tb.mode = InTable
			return true
		}
	}
	tb.mode = InTable
	return true
}

func (tb *TreeBuilder) processInRow(tok tokenizer.Token) bool {
	current := tb.currentElement()
	switch tok.Type {
	case tokenizer.Character, tokenizer.Comment:
		return tb.processInTable(tok)
	case tokenizer.StartTag:
		if tok.Name == "td" || tok.Name == "th" {
			tb.clearStackUntil(map[string]bool{"tr": true, "template": true, "html": true})
			tb.insertElement(tok.Name, tok.Attrs)
			tb.pushActiveFormattingMarker()
			tb.mode = InCell
			return false
		}
		if tok.Name == "tr" {
			tb.popUntil("tr")
			tb.mode = InTableBody
			return true
		}
		if tok.Name == "caption" || tok.Name == "col" || tok.Name == "colgroup" || tok.Name == "tbody" || tok.Name == "tfoot" || tok.Name == "thead" || tok.Name == "table" {
			if current != nil && current.TagName == "template" {
				return false
			}
			if !tb.hasElementInTableScope("tr") {
				return false
			}
			tb.popUntil("tr")
			tb.mode = InTableBody
			return true
		}
	case tokenizer.EndTag:
		switch tok.Name {
		case "tr":
			if !tb.hasElementInTableScope("tr") {
				return false
			}
			tb.popUntil("tr")
			tb.mode = InTableBody
			return false
		case "table":
			if !tb.hasElementInTableScope("tr") {
				return false
			}
			tb.popUntil("tr")
			tb.mode = InTableBody
			return true
		}
	}
	if !tb.hasElementInTableScope("tr") {
		return false
	}
	tb.mode = InTableBody
	return true
}

func (tb *TreeBuilder) processInCell(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.EndTag:
		if tok.Name == "td" || tok.Name == "th" {
			if !tb.hasElementInTableScope(tok.Name) {
				return false
			}
			tb.popUntil(tok.Name)
			tb.clearActiveFormattingElements()
			tb.mode = InRow
			return false
		}
		if tok.Name == "tr" || tok.Name == "table" {
			if !tb.hasElementInTableScope(tok.Name) {
				return false
			}
			tb.popUntilAnyCell()
			tb.clearActiveFormattingElements()
			tb.mode = InRow
			return true
		}
	case tokenizer.StartTag:
		if tok.Name == "caption" || tok.Name == "col" || tok.Name == "colgroup" || tok.Name == "tbody" || tok.Name == "tfoot" || tok.Name == "thead" || tok.Name == "tr" {
			tb.popUntilAnyCell()
			tb.clearActiveFormattingElements()
			tb.mode = InRow
			return true
		}
		if tok.Name == "td" || tok.Name == "th" {
			tb.popUntilAnyCell()
			tb.clearActiveFormattingElements()
			tb.mode = InRow
			return true
		}
	}
	return tb.processInBody(tok)
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
		data := tok.Data
		if strings.ContainsRune(data, 0) || strings.ContainsRune(data, '\f') {
			data = strings.ReplaceAll(data, "\x00", "")
			data = strings.ReplaceAll(data, "\x0c", "")
		}
		if data != "" {
			tb.insertText(data)
		}
		return false
	case tokenizer.Comment:
		tb.insertComment(tok.Data)
		return false
	case tokenizer.StartTag:
		switch tok.Name {
		case "html":
			if len(tb.openElements) > 0 && tb.openElements[0].TagName == "html" {
				tb.addMissingAttributes(tb.openElements[0], tok.Attrs)
			}
			return false
		case "script", "style", "template":
			return tb.processInHead(tok)
		case "hr":
			if tb.currentElement() != nil && tb.currentElement().TagName == "option" {
				tb.popCurrent()
			}
			if tb.currentElement() != nil && tb.currentElement().TagName == "optgroup" {
				tb.popCurrent()
			}
			tb.insertElement("hr", tok.Attrs)
			tb.popCurrent()
			return false
		case "svg":
			tb.insertForeignElement("svg", dom.NamespaceSVG, prepareForeignAttributes(dom.NamespaceSVG, tok.Attrs), tok.SelfClosing)
			return false
		case "math":
			tb.insertForeignElement("math", dom.NamespaceMathML, prepareForeignAttributes(dom.NamespaceMathML, tok.Attrs), tok.SelfClosing)
			return false
		case "input", "textarea":
			tb.popUntil("select")
			tb.resetInsertionModeAppropriately()
			return true
		case "caption", "table", "tbody", "tfoot", "thead", "tr", "td", "th", "col", "colgroup":
			// Parse error; pop the select and reprocess the token.
			tb.popUntil("select")
			tb.resetInsertionModeAppropriately()
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
			tb.resetInsertionModeAppropriately()
			return false
		case "plaintext":
			tb.insertElement("plaintext", tok.Attrs)
			tb.tokenizer.SetLastStartTag("plaintext")
			tb.tokenizer.SetState(tokenizer.PLAINTEXTState)
			return false
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
			tb.resetInsertionModeAppropriately()
			return false
		case "caption", "table", "tbody", "tfoot", "thead", "tr", "td", "th", "col", "colgroup":
			// Parse error; pop the select and reprocess the token.
			tb.popUntil("select")
			tb.resetInsertionModeAppropriately()
			return true
		}
	case tokenizer.EOF:
		return tb.processInBody(tok)
	case tokenizer.Error, tokenizer.DOCTYPE:
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
	case tokenizer.Character, tokenizer.Comment:
		return tb.processInBody(tok)
	case tokenizer.StartTag:
		switch tok.Name {
		case "caption", "colgroup", "tbody", "tfoot", "thead":
			if len(tb.templateModes) > 0 {
				tb.templateModes[len(tb.templateModes)-1] = InTable
			} else {
				tb.templateModes = append(tb.templateModes, InTable)
			}
			tb.mode = InTable
			return true
		case "col":
			if len(tb.templateModes) > 0 {
				tb.templateModes[len(tb.templateModes)-1] = InColumnGroup
			} else {
				tb.templateModes = append(tb.templateModes, InColumnGroup)
			}
			tb.mode = InColumnGroup
			return true
		case "tr":
			if len(tb.templateModes) > 0 {
				tb.templateModes[len(tb.templateModes)-1] = InTableBody
			} else {
				tb.templateModes = append(tb.templateModes, InTableBody)
			}
			tb.mode = InTableBody
			return true
		case "td", "th":
			if len(tb.templateModes) > 0 {
				tb.templateModes[len(tb.templateModes)-1] = InRow
			} else {
				tb.templateModes = append(tb.templateModes, InRow)
			}
			tb.mode = InRow
			return true
		case tagBase, tagBasefont, tagBgsound, tagLink, tagMeta, "noframes", "script", "style", "template", "title":
			return tb.processInHead(tok)
		default:
			if len(tb.templateModes) > 0 {
				tb.templateModes[len(tb.templateModes)-1] = InBody
			} else {
				tb.templateModes = append(tb.templateModes, InBody)
			}
			tb.mode = InBody
			return true
		}
	case tokenizer.EndTag:
		if tok.Name == "template" {
			return tb.processInHead(tok)
		}
	case tokenizer.EOF:
		if !tb.elementInStack("template") {
			return false
		}
		tb.popUntil("template")
		tb.clearActiveFormattingElements()
		if len(tb.templateModes) > 0 {
			tb.templateModes = tb.templateModes[:len(tb.templateModes)-1]
		}
		tb.resetInsertionModeAppropriately()
		return true
	}
	return false
}

func (tb *TreeBuilder) processAfterBody(tok tokenizer.Token) bool {
	switch tok.Type {
	case tokenizer.Character:
		if isAllWhitespace(tok.Data) {
			tb.processInBody(tok)
			return false
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
			tb.processInBody(tok)
			return false
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
