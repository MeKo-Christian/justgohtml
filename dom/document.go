package dom

// QuirksMode represents the document's quirks mode.
type QuirksMode int

// Quirks mode values.
const (
	NoQuirks      QuirksMode = iota // Standards mode
	Quirks                          // Quirks mode
	LimitedQuirks                   // Almost standards mode
)

// Document represents an HTML document.
type Document struct {
	baseNode

	// Doctype is the document's DOCTYPE declaration.
	Doctype *DocumentType

	// QuirksMode indicates the document's quirks mode.
	QuirksMode QuirksMode
}

// NewDocument creates a new empty document.
func NewDocument() *Document {
	d := &Document{}
	d.baseNode.init(d)
	return d
}

// Type implements Node.
func (d *Document) Type() NodeType {
	return DocumentNodeType
}

// Clone implements Node.
func (d *Document) Clone(deep bool) Node {
	clone := &Document{
		QuirksMode: d.QuirksMode,
	}
	clone.baseNode.init(clone)

	if d.Doctype != nil {
		clone.Doctype = d.Doctype.Clone(false).(*DocumentType)
	}

	if deep {
		for _, child := range d.children {
			clonedChild := child.Clone(true)
			clone.AppendChild(clonedChild)
		}
	}

	return clone
}

// AppendChild adds a child node, properly setting the parent.
func (d *Document) AppendChild(child Node) {
	child.SetParent(d)
	d.children = append(d.children, child)
}

// DocumentElement returns the root element (html element).
func (d *Document) DocumentElement() *Element {
	for _, child := range d.children {
		if elem, ok := child.(*Element); ok {
			return elem
		}
	}
	return nil
}

// Head returns the head element, or nil if not found.
func (d *Document) Head() *Element {
	html := d.DocumentElement()
	if html == nil {
		return nil
	}
	for _, child := range html.Children() {
		if elem, ok := child.(*Element); ok && elem.TagName == "head" {
			return elem
		}
	}
	return nil
}

// Body returns the body element, or nil if not found.
func (d *Document) Body() *Element {
	html := d.DocumentElement()
	if html == nil {
		return nil
	}
	for _, child := range html.Children() {
		if elem, ok := child.(*Element); ok && elem.TagName == "body" {
			return elem
		}
	}
	return nil
}

// Title returns the document title from the <title> element.
func (d *Document) Title() string {
	head := d.Head()
	if head == nil {
		return ""
	}
	for _, child := range head.Children() {
		if elem, ok := child.(*Element); ok && elem.TagName == "title" {
			return elem.Text()
		}
	}
	return ""
}

// Query finds all elements matching the CSS selector.
func (d *Document) Query(selector string) ([]*Element, error) {
	root := d.DocumentElement()
	if root == nil {
		return nil, nil
	}
	return root.Query(selector)
}

// QueryFirst finds the first element matching the CSS selector.
func (d *Document) QueryFirst(selector string) (*Element, error) {
	root := d.DocumentElement()
	if root == nil {
		return nil, nil
	}
	return root.QueryFirst(selector)
}

// DocumentType represents a DOCTYPE declaration.
type DocumentType struct {
	parent Node

	// Name is the DOCTYPE name (usually "html").
	Name string

	// PublicID is the public identifier.
	PublicID string

	// SystemID is the system identifier.
	SystemID string
}

// NewDocumentType creates a new DOCTYPE node.
func NewDocumentType(name, publicID, systemID string) *DocumentType {
	return &DocumentType{
		Name:     name,
		PublicID: publicID,
		SystemID: systemID,
	}
}

// Type implements Node.
func (dt *DocumentType) Type() NodeType {
	return DoctypeNodeType
}

// Parent implements Node.
func (dt *DocumentType) Parent() Node {
	return dt.parent
}

// SetParent implements Node.
func (dt *DocumentType) SetParent(parent Node) {
	dt.parent = parent
}

// Children implements Node (DOCTYPE nodes have no children).
func (dt *DocumentType) Children() []Node {
	return nil
}

// AppendChild implements Node (no-op for DOCTYPE nodes).
func (dt *DocumentType) AppendChild(_ Node) {}

// InsertBefore implements Node (no-op for DOCTYPE nodes).
func (dt *DocumentType) InsertBefore(_, _ Node) {}

// RemoveChild implements Node (no-op for DOCTYPE nodes).
func (dt *DocumentType) RemoveChild(_ Node) {}

// Clone implements Node.
func (dt *DocumentType) Clone(_ bool) Node {
	return &DocumentType{
		Name:     dt.Name,
		PublicID: dt.PublicID,
		SystemID: dt.SystemID,
	}
}

// DocumentFragment represents a document fragment (used for template content).
type DocumentFragment struct {
	baseNode
}

// NewDocumentFragment creates a new document fragment.
func NewDocumentFragment() *DocumentFragment {
	df := &DocumentFragment{}
	df.baseNode.init(df)
	return df
}

// Type implements Node.
func (df *DocumentFragment) Type() NodeType {
	// Document fragments don't have a standard type constant
	// Using DocumentNodeType as closest match
	return DocumentNodeType
}

// Clone implements Node.
func (df *DocumentFragment) Clone(deep bool) Node {
	clone := &DocumentFragment{}
	clone.baseNode.init(clone)

	if deep {
		for _, child := range df.children {
			clonedChild := child.Clone(true)
			clone.AppendChild(clonedChild)
		}
	}

	return clone
}

// AppendChild adds a child node, properly setting the parent.
func (df *DocumentFragment) AppendChild(child Node) {
	child.SetParent(df)
	df.children = append(df.children, child)
}
