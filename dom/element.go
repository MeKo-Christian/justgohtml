package dom

import (
	"strings"
)

// Namespace constants for HTML, SVG, and MathML.
const (
	NamespaceHTML   = "http://www.w3.org/1999/xhtml"
	NamespaceSVG    = "http://www.w3.org/2000/svg"
	NamespaceMathML = "http://www.w3.org/1998/Math/MathML"
)

// Element represents an HTML, SVG, or MathML element.
type Element struct {
	baseNode

	// TagName is the element's tag name (lowercase for HTML elements).
	TagName string

	// Namespace is the element's namespace URI.
	// For HTML elements, this is NamespaceHTML.
	Namespace string

	// Attributes contains the element's attributes.
	Attributes *Attributes

	// TemplateContent holds the content of <template> elements.
	// This is nil for non-template elements.
	TemplateContent *DocumentFragment
}

// NewElement creates a new element with the given tag name.
func NewElement(tagName string) *Element {
	e := &Element{
		TagName:    strings.ToLower(tagName),
		Namespace:  NamespaceHTML,
		Attributes: NewAttributes(),
	}
	e.baseNode.init(e)
	return e
}

// NewElementNS creates a new element with the given tag name and namespace.
func NewElementNS(tagName, namespace string) *Element {
	e := &Element{
		TagName:    tagName, // Don't lowercase for foreign elements
		Namespace:  namespace,
		Attributes: NewAttributes(),
	}
	e.baseNode.init(e)
	return e
}

// Type implements Node.
func (e *Element) Type() NodeType {
	return ElementNodeType
}

// Clone implements Node.
func (e *Element) Clone(deep bool) Node {
	clone := &Element{
		TagName:    e.TagName,
		Namespace:  e.Namespace,
		Attributes: e.Attributes.Clone(),
	}
	clone.baseNode.init(clone)

	if deep {
		for _, child := range e.children {
			clonedChild := child.Clone(true)
			clone.AppendChild(clonedChild)
		}
		if e.TemplateContent != nil {
			clone.TemplateContent = e.TemplateContent.Clone(true).(*DocumentFragment)
		}
	}

	return clone
}

// AppendChild adds a child node, properly setting the parent.
func (e *Element) AppendChild(child Node) {
	child.SetParent(e)
	e.children = append(e.children, child)
}

// InsertBefore inserts a new child before a reference child.
func (e *Element) InsertBefore(newChild, refChild Node) {
	if refChild == nil {
		e.AppendChild(newChild)
		return
	}

	for i, child := range e.children {
		if child == refChild {
			newChild.SetParent(e)
			e.children = append(e.children[:i], append([]Node{newChild}, e.children[i:]...)...)
			return
		}
	}
	e.AppendChild(newChild)
}

// RemoveChild removes a child node.
func (e *Element) RemoveChild(child Node) {
	for i, c := range e.children {
		if c == child {
			child.SetParent(nil)
			e.children = append(e.children[:i], e.children[i+1:]...)
			return
		}
	}
}

// Query finds all descendant elements matching the CSS selector.
func (e *Element) Query(selector string) ([]*Element, error) {
	// TODO: Implement selector parsing and matching
	_ = selector
	return nil, nil
}

// QueryFirst finds the first descendant element matching the CSS selector.
func (e *Element) QueryFirst(selector string) (*Element, error) {
	results, err := e.Query(selector)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Text returns the text content of this element and its descendants.
func (e *Element) Text() string {
	var sb strings.Builder
	e.collectText(&sb)
	return sb.String()
}

func (e *Element) collectText(sb *strings.Builder) {
	for _, child := range e.children {
		switch c := child.(type) {
		case *Text:
			sb.WriteString(c.Data)
		case *Element:
			c.collectText(sb)
		}
	}
}

// Attr returns the value of an attribute, or empty string if not present.
func (e *Element) Attr(name string) string {
	val, _ := e.Attributes.Get(name)
	return val
}

// HasAttr returns true if the element has the given attribute.
func (e *Element) HasAttr(name string) bool {
	return e.Attributes.Has(name)
}

// SetAttr sets an attribute value.
func (e *Element) SetAttr(name, value string) {
	e.Attributes.Set(name, value)
}

// RemoveAttr removes an attribute.
func (e *Element) RemoveAttr(name string) {
	e.Attributes.Remove(name)
}

// ID returns the value of the id attribute.
func (e *Element) ID() string {
	return e.Attr("id")
}

// Classes returns the list of CSS classes on this element.
func (e *Element) Classes() []string {
	class := e.Attr("class")
	if class == "" {
		return nil
	}
	return strings.Fields(class)
}

// HasClass returns true if the element has the given CSS class.
func (e *Element) HasClass(class string) bool {
	for _, c := range e.Classes() {
		if c == class {
			return true
		}
	}
	return false
}
