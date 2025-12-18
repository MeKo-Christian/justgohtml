package dom

// selectorMatch is implemented by the selector package and set via SetSelectorMatch.
// This breaks the circular dependency between dom and selector packages.
var selectorMatch func(root *Element, selector string) ([]*Element, error)

// selectorMatchFirst is implemented by the selector package and set via SetSelectorMatchFirst.
var selectorMatchFirst func(root *Element, selector string) (*Element, error)

// SetSelectorMatch sets the function used by Element.Query.
// This is called by the selector package during initialization.
func SetSelectorMatch(fn func(root *Element, selector string) ([]*Element, error)) {
	selectorMatch = fn
}

// SetSelectorMatchFirst sets the function used by Element.QueryFirst.
// This is called by the selector package during initialization.
func SetSelectorMatchFirst(fn func(root *Element, selector string) (*Element, error)) {
	selectorMatchFirst = fn
}

func init() {
	// Set default implementations that return errors if selector package not imported
	selectorMatch = func(_ *Element, _ string) ([]*Element, error) {
		return nil, nil
	}
	selectorMatchFirst = func(_ *Element, _ string) (*Element, error) {
		return nil, nil
	}
}
