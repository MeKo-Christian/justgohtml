package treebuilder

import (
	"github.com/MeKo-Christian/JustGoHTML/dom"
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

	// Table parsing support.
	pendingTableText      []string
	tableTextOriginalMode *InsertionMode
	framesetOK            bool

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
		reprocess := false
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
	tb.currentNode().AppendChild(dom.NewComment(data))
}

func (tb *TreeBuilder) insertText(data string) {
	if data == "" {
		return
	}
	parent := tb.currentNode()
	children := parent.Children()
	if len(children) > 0 {
		if last, ok := children[len(children)-1].(*dom.Text); ok {
			last.Data += data
			return
		}
	}
	parent.AppendChild(dom.NewText(data))
}

func (tb *TreeBuilder) insertElement(name string, attrs map[string]string) *dom.Element {
	el := dom.NewElement(name)
	for k, v := range attrs {
		el.SetAttr(k, v)
	}
	tb.currentNode().AppendChild(el)
	tb.openElements = append(tb.openElements, el)
	return el
}

func (tb *TreeBuilder) addMissingAttributes(el *dom.Element, attrs map[string]string) {
	if el == nil {
		return
	}
	for k, v := range attrs {
		if !el.HasAttr(k) {
			el.SetAttr(k, v)
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

func (tb *TreeBuilder) insertFosterText(data string) {
	if data == "" {
		return
	}
	var tableEl *dom.Element
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		if tb.openElements[i].TagName == "table" && tb.openElements[i].Namespace == dom.NamespaceHTML {
			tableEl = tb.openElements[i]
			break
		}
	}
	if tableEl == nil {
		tb.insertText(data)
		return
	}
	parent := tableEl.Parent()
	if parent == nil {
		tb.document.AppendChild(dom.NewText(data))
		return
	}
	parent.InsertBefore(dom.NewText(data), tableEl)
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
