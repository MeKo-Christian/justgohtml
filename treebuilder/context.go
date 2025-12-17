// Package treebuilder implements the HTML5 tree construction algorithm.
package treebuilder

// FragmentContext specifies the context for fragment parsing.
// This is used when parsing HTML in a specific context (like innerHTML).
type FragmentContext struct {
	// TagName is the context element's tag name (e.g., "div", "tr", "body").
	TagName string

	// Namespace is the context element's namespace.
	// Usually "html", but can be "svg" or "mathml" for foreign elements.
	Namespace string
}
