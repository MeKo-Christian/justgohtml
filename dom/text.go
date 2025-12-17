package dom

// Text represents a text node.
type Text struct {
	parent Node

	// Data is the text content.
	Data string
}

// NewText creates a new text node.
func NewText(data string) *Text {
	return &Text{Data: data}
}

// Type implements Node.
func (t *Text) Type() NodeType {
	return TextNodeType
}

// Parent implements Node.
func (t *Text) Parent() Node {
	return t.parent
}

// SetParent implements Node.
func (t *Text) SetParent(parent Node) {
	t.parent = parent
}

// Children implements Node (text nodes have no children).
func (t *Text) Children() []Node {
	return nil
}

// AppendChild implements Node (no-op for text nodes).
func (t *Text) AppendChild(_ Node) {}

// InsertBefore implements Node (no-op for text nodes).
func (t *Text) InsertBefore(_, _ Node) {}

// RemoveChild implements Node (no-op for text nodes).
func (t *Text) RemoveChild(_ Node) {}

// Clone implements Node.
func (t *Text) Clone(_ bool) Node {
	return &Text{Data: t.Data}
}

// Comment represents a comment node.
type Comment struct {
	parent Node

	// Data is the comment content (without <!-- and -->).
	Data string
}

// NewComment creates a new comment node.
func NewComment(data string) *Comment {
	return &Comment{Data: data}
}

// Type implements Node.
func (c *Comment) Type() NodeType {
	return CommentNodeType
}

// Parent implements Node.
func (c *Comment) Parent() Node {
	return c.parent
}

// SetParent implements Node.
func (c *Comment) SetParent(parent Node) {
	c.parent = parent
}

// Children implements Node (comment nodes have no children).
func (c *Comment) Children() []Node {
	return nil
}

// AppendChild implements Node (no-op for comment nodes).
func (c *Comment) AppendChild(_ Node) {}

// InsertBefore implements Node (no-op for comment nodes).
func (c *Comment) InsertBefore(_, _ Node) {}

// RemoveChild implements Node (no-op for comment nodes).
func (c *Comment) RemoveChild(_ Node) {}

// Clone implements Node.
func (c *Comment) Clone(_ bool) Node {
	return &Comment{Data: c.Data}
}
