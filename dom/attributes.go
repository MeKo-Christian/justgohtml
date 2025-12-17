package dom

import (
	"strings"
)

// Attribute represents a single HTML attribute.
type Attribute struct {
	// Namespace is the attribute namespace (usually empty for HTML attributes).
	Namespace string

	// Name is the attribute name (lowercase for HTML attributes).
	Name string

	// Value is the attribute value.
	Value string
}

// Attributes holds a collection of attributes for an element.
// Attributes are stored in insertion order and accessed case-insensitively for HTML.
type Attributes struct {
	items []Attribute
}

// NewAttributes creates a new empty Attributes collection.
func NewAttributes() *Attributes {
	return &Attributes{}
}

// Get returns the value of an attribute by name.
// For HTML attributes, the lookup is case-insensitive.
// Returns the value and true if found, or empty string and false if not.
func (a *Attributes) Get(name string) (string, bool) {
	lowerName := strings.ToLower(name)
	for _, attr := range a.items {
		if strings.ToLower(attr.Name) == lowerName && attr.Namespace == "" {
			return attr.Value, true
		}
	}
	return "", false
}

// GetNS returns the value of a namespaced attribute.
func (a *Attributes) GetNS(namespace, name string) (string, bool) {
	for _, attr := range a.items {
		if attr.Namespace == namespace && attr.Name == name {
			return attr.Value, true
		}
	}
	return "", false
}

// Set sets or updates an attribute value.
// For HTML attributes, callers should pass a lowercase name (the tokenizer already does).
func (a *Attributes) Set(name, value string) {
	a.SetNS("", strings.ToLower(name), value)
}

// SetNS sets or updates a namespaced attribute value.
func (a *Attributes) SetNS(namespace, name, value string) {
	// Try to update existing attribute
	for i := range a.items {
		if a.items[i].Namespace == namespace && strings.EqualFold(a.items[i].Name, name) {
			a.items[i].Value = value
			return
		}
	}

	// Add new attribute
	a.items = append(a.items, Attribute{
		Namespace: namespace,
		Name:      name,
		Value:     value,
	})
}

// Has returns true if an attribute with the given name exists.
func (a *Attributes) Has(name string) bool {
	_, found := a.Get(name)
	return found
}

// HasNS returns true if a namespaced attribute exists.
func (a *Attributes) HasNS(namespace, name string) bool {
	_, found := a.GetNS(namespace, name)
	return found
}

// Remove removes an attribute by name.
func (a *Attributes) Remove(name string) {
	a.RemoveNS("", name)
}

// RemoveNS removes a namespaced attribute.
func (a *Attributes) RemoveNS(namespace, name string) {
	lowerName := strings.ToLower(name)
	for i := range a.items {
		if a.items[i].Namespace == namespace && strings.ToLower(a.items[i].Name) == lowerName {
			a.items = append(a.items[:i], a.items[i+1:]...)
			return
		}
	}
}

// All returns all attributes in insertion order.
func (a *Attributes) All() []Attribute {
	result := make([]Attribute, len(a.items))
	copy(result, a.items)
	return result
}

// Len returns the number of attributes.
func (a *Attributes) Len() int {
	return len(a.items)
}

// Clone creates a copy of the attributes.
func (a *Attributes) Clone() *Attributes {
	clone := &Attributes{
		items: make([]Attribute, len(a.items)),
	}
	copy(clone.items, a.items)
	return clone
}
