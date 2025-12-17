// Package selector implements CSS selector parsing and matching.
package selector

import (
	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/errors"
)

// Selector represents a parsed CSS selector.
type Selector interface {
	// Match returns true if the element matches this selector.
	Match(element *dom.Element) bool

	// String returns the original selector string.
	String() string
}

// Parse parses a CSS selector string.
func Parse(selector string) (Selector, error) {
	// TODO: Implement selector parsing
	_ = selector
	return nil, &errors.SelectorError{
		Selector: selector,
		Position: 0,
		Message:  "selector parsing not yet implemented",
	}
}

// Match returns all elements in the subtree that match the selector.
func Match(root *dom.Element, selector string) ([]*dom.Element, error) {
	sel, err := Parse(selector)
	if err != nil {
		return nil, err
	}

	var results []*dom.Element
	matchDescendants(root, sel, &results)
	return results, nil
}

// MatchFirst returns the first element that matches the selector.
func MatchFirst(root *dom.Element, selector string) (*dom.Element, error) {
	sel, err := Parse(selector)
	if err != nil {
		return nil, err
	}

	return findFirst(root, sel), nil
}

func matchDescendants(elem *dom.Element, sel Selector, results *[]*dom.Element) {
	if sel.Match(elem) {
		*results = append(*results, elem)
	}
	for _, child := range elem.Children() {
		if childElem, ok := child.(*dom.Element); ok {
			matchDescendants(childElem, sel, results)
		}
	}
}

func findFirst(elem *dom.Element, sel Selector) *dom.Element {
	if sel.Match(elem) {
		return elem
	}
	for _, child := range elem.Children() {
		if childElem, ok := child.(*dom.Element); ok {
			if found := findFirst(childElem, sel); found != nil {
				return found
			}
		}
	}
	return nil
}
