package treebuilder

// InsertionMode represents the current insertion mode of the tree builder.
// These modes determine how tokens are processed during tree construction.
type InsertionMode int

// Insertion modes as defined by the HTML5 specification.
// See: https://html.spec.whatwg.org/multipage/parsing.html#insertion-mode
const (
	Initial InsertionMode = iota
	BeforeHTML
	BeforeHead
	InHead
	InHeadNoscript
	AfterHead
	InBody
	Text
	InTable
	InTableText
	InCaption
	InColumnGroup
	InTableBody
	InRow
	InCell
	InSelect
	InSelectInTable
	InTemplate
	AfterBody
	InFrameset
	AfterFrameset
	AfterAfterBody
	AfterAfterFrameset
)

// String returns the name of the insertion mode for debugging.
func (m InsertionMode) String() string {
	names := [...]string{
		"initial",
		"before html",
		"before head",
		"in head",
		"in head noscript",
		"after head",
		"in body",
		"text",
		"in table",
		"in table text",
		"in caption",
		"in column group",
		"in table body",
		"in row",
		"in cell",
		"in select",
		"in select in table",
		"in template",
		"after body",
		"in frameset",
		"after frameset",
		"after after body",
		"after after frameset",
	}
	if m >= 0 && int(m) < len(names) {
		return names[m]
	}
	return "unknown"
}
