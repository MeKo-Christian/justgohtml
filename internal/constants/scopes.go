package constants

// Scope terminators for the tree builder.
// These define which elements terminate various scopes during parsing.

// DefaultScope elements terminate the default scope.
var DefaultScope = map[string]bool{
	"applet":   true,
	"caption":  true,
	"html":     true,
	"table":    true,
	"td":       true,
	"th":       true,
	"marquee":  true,
	"object":   true,
	"template": true,
	// MathML elements
	"mi":             true,
	"mo":             true,
	"mn":             true,
	"ms":             true,
	"mtext":          true,
	"annotation-xml": true,
	// SVG elements
	"foreignObject": true,
	"desc":          true,
	"title":         true,
}

// ListItemScope elements terminate list item scope.
var ListItemScope = map[string]bool{
	"applet":   true,
	"caption":  true,
	"html":     true,
	"table":    true,
	"td":       true,
	"th":       true,
	"marquee":  true,
	"object":   true,
	"template": true,
	"ol":       true,
	"ul":       true,
	// MathML elements
	"mi":             true,
	"mo":             true,
	"mn":             true,
	"ms":             true,
	"mtext":          true,
	"annotation-xml": true,
	// SVG elements
	"foreignObject": true,
	"desc":          true,
	"title":         true,
}

// ButtonScope elements terminate button scope.
var ButtonScope = map[string]bool{
	"applet":   true,
	"caption":  true,
	"html":     true,
	"table":    true,
	"td":       true,
	"th":       true,
	"marquee":  true,
	"object":   true,
	"template": true,
	"button":   true,
	// MathML elements
	"mi":             true,
	"mo":             true,
	"mn":             true,
	"ms":             true,
	"mtext":          true,
	"annotation-xml": true,
	// SVG elements
	"foreignObject": true,
	"desc":          true,
	"title":         true,
}

// TableScope elements terminate table scope.
var TableScope = map[string]bool{
	"html":     true,
	"table":    true,
	"template": true,
}

// TableBodyScope elements terminate table body scope.
var TableBodyScope = map[string]bool{
	"html":     true,
	"table":    true,
	"template": true,
	"tbody":    true,
	"tfoot":    true,
	"thead":    true,
}

// TableRowScope elements terminate table row scope.
var TableRowScope = map[string]bool{
	"html":     true,
	"table":    true,
	"template": true,
	"tbody":    true,
	"tfoot":    true,
	"thead":    true,
	"tr":       true,
}

// SelectScope elements are NOT scope terminators for select (everything except these).
var SelectScope = map[string]bool{
	"optgroup": true,
	"option":   true,
}
