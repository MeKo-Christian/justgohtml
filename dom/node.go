// Package dom provides DOM node types for the HTML5 parser.
package dom

// NodeType represents the type of a DOM node.
type NodeType int

// Node types as defined by the DOM specification.
const (
	ElementNodeType  NodeType = 1
	TextNodeType     NodeType = 3
	CommentNodeType  NodeType = 8
	DocumentNodeType NodeType = 9
	DoctypeNodeType  NodeType = 10
)

// Node is the interface implemented by all DOM node types.
type Node interface {
	// Type returns the node type.
	Type() NodeType

	// Parent returns the parent node, or nil if this is the root.
	Parent() Node

	// SetParent sets the parent node.
	SetParent(parent Node)

	// Children returns the child nodes.
	Children() []Node

	// AppendChild adds a child node.
	AppendChild(child Node)

	// InsertBefore inserts a new child before a reference child.
	InsertBefore(newChild, refChild Node)

	// RemoveChild removes a child node.
	RemoveChild(child Node)

	// ReplaceChild replaces an old child with a new child.
	// Returns the replaced child (oldChild).
	ReplaceChild(newChild, oldChild Node) Node

	// HasChildNodes returns true if this node has any children.
	HasChildNodes() bool

	// Clone creates a copy of this node.
	// If deep is true, all descendants are also cloned.
	Clone(deep bool) Node
}

// baseNode provides common functionality for all node types.
type baseNode struct {
	self     Node
	parent   Node
	children []Node
}

func (n *baseNode) init(self Node) {
	n.self = self
}

func (n *baseNode) Parent() Node {
	return n.parent
}

func (n *baseNode) SetParent(parent Node) {
	n.parent = parent
}

func (n *baseNode) Children() []Node {
	return n.children
}

func (n *baseNode) AppendChild(child Node) {
	if n.self != nil {
		child.SetParent(n.self)
	}
	n.children = append(n.children, child)
}

func (n *baseNode) InsertBefore(newChild, refChild Node) {
	if refChild == nil {
		n.AppendChild(newChild)
		return
	}

	for i, child := range n.children {
		if child == refChild {
			if n.self != nil {
				newChild.SetParent(n.self)
			}
			n.children = append(n.children[:i], append([]Node{newChild}, n.children[i:]...)...)
			return
		}
	}
	// refChild not found, append
	n.AppendChild(newChild)
}

func (n *baseNode) RemoveChild(child Node) {
	for i, c := range n.children {
		if c == child {
			child.SetParent(nil)
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
}

func (n *baseNode) ReplaceChild(newChild, oldChild Node) Node {
	for i, c := range n.children {
		if c == oldChild {
			if n.self != nil {
				newChild.SetParent(n.self)
			}
			oldChild.SetParent(nil)
			n.children[i] = newChild
			return oldChild
		}
	}
	return nil
}

func (n *baseNode) HasChildNodes() bool {
	return len(n.children) > 0
}
