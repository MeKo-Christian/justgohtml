package treebuilder

import (
	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

// TreeBuilder implements a (work-in-progress) HTML5 tree construction stage.
//
// This is a direct porting target of the Python reference implementation and is
// intended to be driven by the tokenizer token stream.
type TreeBuilder struct {
	document *dom.Document

	openElements []*dom.Element

	mode         InsertionMode
	originalMode InsertionMode

	headElement *dom.Element

	activeFormatting []formattingEntry

	// Template insertion modes stack.
	templateModes []InsertionMode

	// Table parsing support.
	pendingTableText      []string
	tableTextOriginalMode *InsertionMode
	framesetOK            bool
	fosterParenting       bool

	fragmentContext *FragmentContext
	fragmentRoot    *dom.Element
	fragmentElement *dom.Element

	tokenizer *tokenizer.Tokenizer

	// forceHTMLMode is set by processForeignContent when it encounters a token
	// that should be reprocessed using normal HTML insertion mode rules rather
	// than foreign content rules. This prevents infinite loops when foreign
	// content contains tokens that trigger breakout to HTML mode.
	forceHTMLMode bool

	iframeSrcdoc bool
}

// New creates a new tree builder for full document parsing.
func New(tok *tokenizer.Tokenizer) *TreeBuilder {
	return &TreeBuilder{
		document:         dom.NewDocument(),
		mode:             Initial,
		originalMode:     Initial,
		openElements:     nil,
		activeFormatting: nil,
		templateModes:    nil,
		pendingTableText: nil,
		framesetOK:       true,
		fragmentRoot:     nil,
		fragmentContext:  nil,
		tokenizer:        tok,
	}
}

// NewFragment creates a new tree builder for fragment parsing.
func NewFragment(tok *tokenizer.Tokenizer, ctx *FragmentContext) *TreeBuilder {
	tb := &TreeBuilder{
		document:         dom.NewDocument(),
		mode:             Initial,
		originalMode:     Initial,
		openElements:     nil,
		activeFormatting: nil,
		templateModes:    nil,
		pendingTableText: nil,
		framesetOK:       false,
		fragmentContext:  ctx,
		tokenizer:        tok,
	}

	// Minimal fragment setup: create an <html> root and a context element.
	html := dom.NewElement("html")
	tb.document.AppendChild(html)
	tb.openElements = append(tb.openElements, html)
	tb.fragmentRoot = html

	if ctx != nil && ctx.TagName != "" {
		contextEl := dom.NewElement(ctx.TagName)
		switch ctx.Namespace {
		case "svg":
			contextEl = dom.NewElementNS(ctx.TagName, dom.NamespaceSVG)
		case "mathml":
			contextEl = dom.NewElementNS(ctx.TagName, dom.NamespaceMathML)
		}
		html.AppendChild(contextEl)
		tb.openElements = append(tb.openElements, contextEl)
		tb.fragmentElement = contextEl

		// Set the initial insertion mode based on the context element, per HTML5 fragment parsing.
		tag := contextEl.TagName
		if ctx.Namespace != "" && ctx.Namespace != "html" {
			tb.mode = InBody
		} else {
			switch tag {
			case "html":
				tb.mode = BeforeHead
			case "tbody", "thead", "tfoot":
				tb.mode = InTableBody
			case "tr":
				tb.mode = InRow
			case "td", "th":
				tb.mode = InCell
			case "caption":
				tb.mode = InCaption
			case "colgroup":
				tb.mode = InColumnGroup
			case "table":
				tb.mode = InTable
			case "select":
				tb.mode = InSelect
			default:
				tb.mode = InBody
			}
		}
		tb.originalMode = tb.mode

		// Adjust tokenizer state based on the fragment context element, per HTML5 fragment parsing.
		// This is necessary because the fragment setup does not emit the context start tag token.
		if ctx.Namespace == "" || ctx.Namespace == "html" {
			switch tag {
			case "title", "textarea":
				tb.tokenizer.SetLastStartTag(tag)
				tb.tokenizer.SetState(tokenizer.RCDATAState)
			case "style", "xmp", "iframe", "noembed", "noframes":
				tb.tokenizer.SetLastStartTag(tag)
				tb.tokenizer.SetState(tokenizer.RAWTEXTState)
			case "script":
				tb.tokenizer.SetLastStartTag(tag)
				tb.tokenizer.SetState(tokenizer.ScriptDataState)
			case "plaintext":
				tb.tokenizer.SetLastStartTag(tag)
				tb.tokenizer.SetState(tokenizer.PLAINTEXTState)
			}
		}
	}

	return tb
}

// SetIframeSrcdoc toggles iframe srcdoc parsing behavior (affects quirks mode decisions).
func (tb *TreeBuilder) SetIframeSrcdoc(enabled bool) {
	tb.iframeSrcdoc = enabled
}

// Document returns the constructed document.
func (tb *TreeBuilder) Document() *dom.Document {
	return tb.document
}

// FragmentNodes returns the fragment's top-level element children.
func (tb *TreeBuilder) FragmentNodes() []*dom.Element {
	root := tb.fragmentElement
	if root == nil {
		root = tb.fragmentRoot
	}
	if root == nil {
		return nil
	}
	var out []*dom.Element
	for _, child := range root.Children() {
		if el, ok := child.(*dom.Element); ok {
			out = append(out, el)
		}
	}
	return out
}

// ProcessToken consumes a tokenizer token and updates the DOM tree.
func (tb *TreeBuilder) ProcessToken(tok tokenizer.Token) {
	// The full HTML5 algorithm is implemented incrementally; keep the current
	// behavior non-panicking and deterministic.
	for {
		// Check if we should use foreign content rules.
		// forceHTMLMode bypasses this check when reprocessing a token that
		// triggered breakout from foreign content.
		if !tb.forceHTMLMode && tb.shouldUseForeignContent(tok) {
			reprocess := tb.processForeignContent(tok)
			if !reprocess {
				return
			}
			continue
		}
		tb.forceHTMLMode = false
		var reprocess bool
		switch tb.mode {
		case Initial:
			reprocess = tb.processInitial(tok)
		case BeforeHTML:
			reprocess = tb.processBeforeHTML(tok)
		case BeforeHead:
			reprocess = tb.processBeforeHead(tok)
		case InHead:
			reprocess = tb.processInHead(tok)
		case InHeadNoscript:
			reprocess = tb.processInHeadNoscript(tok)
		case AfterHead:
			reprocess = tb.processAfterHead(tok)
		case Text:
			reprocess = tb.processText(tok)
		case InBody:
			reprocess = tb.processInBody(tok)
		case InTable:
			reprocess = tb.processInTable(tok)
		case InTableText:
			reprocess = tb.processInTableText(tok)
		case InCaption:
			reprocess = tb.processInCaption(tok)
		case InColumnGroup:
			reprocess = tb.processInColumnGroup(tok)
		case InTableBody:
			reprocess = tb.processInTableBody(tok)
		case InRow:
			reprocess = tb.processInRow(tok)
		case InCell:
			reprocess = tb.processInCell(tok)
		case InSelect:
			reprocess = tb.processInSelect(tok)
		case InSelectInTable:
			reprocess = tb.processInSelectInTable(tok)
		case InTemplate:
			reprocess = tb.processInTemplate(tok)
		case AfterBody:
			reprocess = tb.processAfterBody(tok)
		case InFrameset:
			reprocess = tb.processInFrameset(tok)
		case AfterFrameset:
			reprocess = tb.processAfterFrameset(tok)
		case AfterAfterBody:
			reprocess = tb.processAfterAfterBody(tok)
		case AfterAfterFrameset:
			reprocess = tb.processAfterAfterFrameset(tok)
		default:
			// Fallback: treat as InBody for now.
			reprocess = tb.processInBody(tok)
		}
		if !reprocess {
			return
		}
	}
}

func (tb *TreeBuilder) currentNode() dom.Node {
	if len(tb.openElements) == 0 {
		return tb.document
	}
	return tb.openElements[len(tb.openElements)-1]
}

func (tb *TreeBuilder) currentElement() *dom.Element {
	if len(tb.openElements) == 0 {
		return nil
	}
	return tb.openElements[len(tb.openElements)-1]
}

func (tb *TreeBuilder) insertComment(data string) {
	tb.insertNode(dom.NewComment(data), nil)
}

func (tb *TreeBuilder) insertText(data string) {
	if data == "" {
		return
	}
	parent, before := tb.appropriateInsertionLocation()
	tb.insertNode(dom.NewText(data), &insertionLocation{parent: parent, before: before})
}

func (tb *TreeBuilder) insertElement(name string, attrs []tokenizer.Attr) *dom.Element {
	el := dom.NewElement(name)
	if el.TagName == "template" && el.Namespace == dom.NamespaceHTML && el.TemplateContent == nil {
		el.TemplateContent = dom.NewDocumentFragment()
	}
	for _, a := range attrs {
		if a.Namespace != "" {
			// HTML namespace attributes are handled later (foreign content).
			el.Attributes.SetNS(a.Namespace, a.Name, a.Value)
			continue
		}
		el.SetAttr(a.Name, a.Value)
	}
	tb.insertNode(el, nil)
	tb.openElements = append(tb.openElements, el)
	return el
}

func (tb *TreeBuilder) addMissingAttributes(el *dom.Element, attrs []tokenizer.Attr) {
	if el == nil {
		return
	}
	if len(tb.templateModes) > 0 {
		return
	}
	for _, a := range attrs {
		if a.Namespace != "" {
			if !el.Attributes.HasNS(a.Namespace, a.Name) {
				el.Attributes.SetNS(a.Namespace, a.Name, a.Value)
			}
			continue
		}
		if !el.HasAttr(a.Name) {
			el.SetAttr(a.Name, a.Value)
		}
	}
}

func (tb *TreeBuilder) popCurrent() *dom.Element {
	if len(tb.openElements) == 0 {
		return nil
	}
	el := tb.openElements[len(tb.openElements)-1]
	tb.openElements = tb.openElements[:len(tb.openElements)-1]
	return el
}

func (tb *TreeBuilder) popUntil(name string) {
	for len(tb.openElements) > 0 {
		el := tb.openElements[len(tb.openElements)-1]
		tb.openElements = tb.openElements[:len(tb.openElements)-1]
		if el.TagName == name {
			return
		}
	}
}

func (tb *TreeBuilder) elementInStack(name string) bool {
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		if tb.openElements[i].TagName == name {
			return true
		}
	}
	return false
}

func isAllWhitespace(s string) bool {
	for _, r := range s {
		switch r {
		case '\t', '\n', '\f', '\r', ' ':
			continue
		default:
			return false
		}
	}
	return true
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type insertionLocation struct {
	parent dom.Node
	before dom.Node
}

func (tb *TreeBuilder) withFosterParenting(fn func() bool) bool {
	prev := tb.fosterParenting
	tb.fosterParenting = true
	defer func() { tb.fosterParenting = prev }()
	return fn()
}

func (tb *TreeBuilder) appropriateInsertionLocation() (dom.Node, dom.Node) {
	if current := tb.currentElement(); current != nil && current.Namespace == dom.NamespaceHTML && current.TagName == "template" {
		if current.TemplateContent == nil {
			current.TemplateContent = dom.NewDocumentFragment()
		}
		return current.TemplateContent, nil
	}
	if !tb.fosterParenting || !shouldFosterForNode(tb.currentElement()) {
		return tb.currentNode(), nil
	}
	return tb.fosterInsertionLocation()
}

func shouldFosterForNode(el *dom.Element) bool {
	if el == nil || el.Namespace != dom.NamespaceHTML {
		return false
	}
	return constants.TableFosterTargets[el.TagName]
}

func (tb *TreeBuilder) shouldFosterParenting(target *dom.Element, forTag string, isText bool) bool {
	if !tb.fosterParenting {
		return false
	}
	if target == nil || target.Namespace != dom.NamespaceHTML {
		return false
	}
	if !constants.TableFosterTargets[target.TagName] {
		return false
	}
	if isText {
		return true
	}
	if forTag != "" && constants.TableAllowedChildren[forTag] {
		return false
	}
	return true
}

func (tb *TreeBuilder) fosterInsertionLocation() (dom.Node, dom.Node) {
	tableEl, tableIndex := tb.lastTableElement()
	templateEl, templateIndex := tb.lastTemplateElement()
	if templateEl != nil && (tableEl == nil || templateIndex > tableIndex) {
		if templateEl.TemplateContent == nil {
			templateEl.TemplateContent = dom.NewDocumentFragment()
		}
		return templateEl.TemplateContent, nil
	}
	if tableEl == nil {
		return tb.currentNode(), nil
	}
	if p := tableEl.Parent(); p != nil {
		return p, tableEl
	}

	// If the table element has no parent, insert into the element immediately above it in the stack.
	if tableIndex > 0 {
		return tb.openElements[tableIndex-1], nil
	}
	return tb.document, nil
}

func (tb *TreeBuilder) lastTableElement() (*dom.Element, int) {
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		el := tb.openElements[i]
		if el != nil && el.Namespace == dom.NamespaceHTML && el.TagName == "table" {
			return el, i
		}
	}
	return nil, -1
}

func (tb *TreeBuilder) lastTemplateElement() (*dom.Element, int) {
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		el := tb.openElements[i]
		if el != nil && el.Namespace == dom.NamespaceHTML && el.TagName == "template" {
			return el, i
		}
	}
	return nil, -1
}

func (tb *TreeBuilder) insertNode(node dom.Node, loc *insertionLocation) {
	var parent dom.Node
	var before dom.Node
	if loc != nil && loc.parent != nil {
		parent = loc.parent
		before = loc.before
	} else {
		parent, before = tb.appropriateInsertionLocation()
	}

	if before == nil {
		// Append with text-node coalescing.
		children := parent.Children()
		if txt, ok := node.(*dom.Text); ok && len(children) > 0 {
			if last, ok := children[len(children)-1].(*dom.Text); ok {
				last.Data += txt.Data
				return
			}
		}
		parent.AppendChild(node)
		return
	}

	// InsertBefore with basic text-node coalescing around the insertion point.
	if txt, ok := node.(*dom.Text); ok {
		if mergeTarget := siblingTextBefore(parent, before); mergeTarget != nil {
			mergeTarget.Data += txt.Data
			return
		}
		if beforeText, ok := before.(*dom.Text); ok {
			beforeText.Data = txt.Data + beforeText.Data
			return
		}
	}
	parent.InsertBefore(node, before)
}

func siblingTextBefore(parent dom.Node, ref dom.Node) *dom.Text {
	children := parent.Children()
	for i := range children {
		if children[i] == ref {
			if i > 0 {
				if t, ok := children[i-1].(*dom.Text); ok {
					return t
				}
			}
			return nil
		}
	}
	return nil
}
