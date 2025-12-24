package dom

import "strings"

const (
	elementChunkSize   = 128
	textChunkSize      = 256
	commentChunkSize   = 64
	doctypeChunkSize   = 32
	documentChunkSize  = 8
	fragmentChunkSize  = 64
	attributeChunkSize = 128
)

// NodeAllocator provides arena-style allocation for DOM nodes.
// It reduces per-node allocations by handing out pointers from fixed-size chunks.
type NodeAllocator struct {
	elements  []Element
	elementAt int

	texts  []Text
	textAt int

	comments  []Comment
	commentAt int

	doctypes  []DocumentType
	doctypeAt int

	documents  []Document
	documentAt int

	fragments  []DocumentFragment
	fragmentAt int

	attributes  []Attributes
	attributeAt int
}

// NewNodeAllocator creates a new allocator for DOM nodes.
func NewNodeAllocator() *NodeAllocator {
	return &NodeAllocator{}
}

func (a *NodeAllocator) nextElement() *Element {
	if a.elementAt >= len(a.elements) {
		a.elements = make([]Element, elementChunkSize)
		a.elementAt = 0
	}
	e := &a.elements[a.elementAt]
	a.elementAt++
	return e
}

func (a *NodeAllocator) nextText() *Text {
	if a.textAt >= len(a.texts) {
		a.texts = make([]Text, textChunkSize)
		a.textAt = 0
	}
	t := &a.texts[a.textAt]
	a.textAt++
	return t
}

func (a *NodeAllocator) nextComment() *Comment {
	if a.commentAt >= len(a.comments) {
		a.comments = make([]Comment, commentChunkSize)
		a.commentAt = 0
	}
	c := &a.comments[a.commentAt]
	a.commentAt++
	return c
}

func (a *NodeAllocator) nextDoctype() *DocumentType {
	if a.doctypeAt >= len(a.doctypes) {
		a.doctypes = make([]DocumentType, doctypeChunkSize)
		a.doctypeAt = 0
	}
	dt := &a.doctypes[a.doctypeAt]
	a.doctypeAt++
	return dt
}

func (a *NodeAllocator) nextDocument() *Document {
	if a.documentAt >= len(a.documents) {
		a.documents = make([]Document, documentChunkSize)
		a.documentAt = 0
	}
	d := &a.documents[a.documentAt]
	a.documentAt++
	return d
}

func (a *NodeAllocator) nextFragment() *DocumentFragment {
	if a.fragmentAt >= len(a.fragments) {
		a.fragments = make([]DocumentFragment, fragmentChunkSize)
		a.fragmentAt = 0
	}
	df := &a.fragments[a.fragmentAt]
	a.fragmentAt++
	return df
}

func (a *NodeAllocator) nextAttributes() *Attributes {
	if a.attributeAt >= len(a.attributes) {
		a.attributes = make([]Attributes, attributeChunkSize)
		a.attributeAt = 0
	}
	attr := &a.attributes[a.attributeAt]
	a.attributeAt++
	return attr
}

// NewDocument creates a new document node.
func (a *NodeAllocator) NewDocument() *Document {
	d := a.nextDocument()
	d.baseNode = baseNode{}
	d.Doctype = nil
	d.QuirksMode = NoQuirks
	d.init(d)
	return d
}

// NewDocumentFragment creates a new document fragment.
func (a *NodeAllocator) NewDocumentFragment() *DocumentFragment {
	df := a.nextFragment()
	df.baseNode = baseNode{}
	df.init(df)
	return df
}

// NewElement creates a new HTML element with lowercase tag name.
func (a *NodeAllocator) NewElement(tagName string) *Element {
	e := a.nextElement()
	e.baseNode = baseNode{}
	e.TagName = strings.ToLower(tagName)
	e.Namespace = NamespaceHTML
	e.Attributes = a.newAttributes()
	e.TemplateContent = nil
	e.init(e)
	return e
}

// NewElementNS creates a new element with the given namespace.
func (a *NodeAllocator) NewElementNS(tagName, namespace string) *Element {
	e := a.nextElement()
	e.baseNode = baseNode{}
	e.TagName = tagName
	e.Namespace = namespace
	e.Attributes = a.newAttributes()
	e.TemplateContent = nil
	e.init(e)
	return e
}

// NewText creates a new text node.
func (a *NodeAllocator) NewText(data string) *Text {
	t := a.nextText()
	t.parent = nil
	t.Data = data
	return t
}

// NewComment creates a new comment node.
func (a *NodeAllocator) NewComment(data string) *Comment {
	c := a.nextComment()
	c.parent = nil
	c.Data = data
	return c
}

// NewDocumentType creates a new DOCTYPE node.
func (a *NodeAllocator) NewDocumentType(name, publicID, systemID string) *DocumentType {
	dt := a.nextDoctype()
	dt.parent = nil
	dt.Name = name
	dt.PublicID = publicID
	dt.SystemID = systemID
	return dt
}

func (a *NodeAllocator) newAttributes() *Attributes {
	attr := a.nextAttributes()
	attr.items = attr.items[:0]
	return attr
}
