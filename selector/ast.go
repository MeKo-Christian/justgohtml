// Package selector implements CSS selector parsing and matching.
package selector

// SelectorKind identifies the type of simple selector.
type SelectorKind int

const (
	KindTag       SelectorKind = iota // div, span, etc.
	KindUniversal                     // *
	KindID                            // #foo
	KindClass                         // .bar
	KindAttr                          // [attr], [attr="val"]
	KindPseudo                        // :first-child, :nth-child()
)

// String returns a string representation of the selector kind.
func (k SelectorKind) String() string {
	switch k {
	case KindTag:
		return "Tag"
	case KindUniversal:
		return "Universal"
	case KindID:
		return "ID"
	case KindClass:
		return "Class"
	case KindAttr:
		return "Attr"
	case KindPseudo:
		return "Pseudo"
	default:
		return "Unknown"
	}
}

// AttrOperator represents attribute comparison operators.
type AttrOperator int

const (
	AttrExists      AttrOperator = iota // [attr]
	AttrEquals                          // [attr="val"]
	AttrIncludes                        // [attr~="val"] - word match
	AttrDashPrefix                      // [attr|="val"] - prefix match (hyphen-separated)
	AttrPrefixMatch                     // [attr^="val"] - starts with
	AttrSuffixMatch                     // [attr$="val"] - ends with
	AttrSubstring                       // [attr*="val"] - contains
)

// String returns a string representation of the attribute operator.
func (op AttrOperator) String() string {
	switch op {
	case AttrExists:
		return ""
	case AttrEquals:
		return "="
	case AttrIncludes:
		return "~="
	case AttrDashPrefix:
		return "|="
	case AttrPrefixMatch:
		return "^="
	case AttrSuffixMatch:
		return "$="
	case AttrSubstring:
		return "*="
	default:
		return "?"
	}
}

// Combinator represents the relationship between compound selectors.
type Combinator int

const (
	CombinatorNone       Combinator = iota // No combinator (first in chain)
	CombinatorDescendant                   // space (descendant)
	CombinatorChild                        // > (direct child)
	CombinatorAdjacent                     // + (adjacent sibling)
	CombinatorGeneral                      // ~ (general sibling)
)

// String returns a string representation of the combinator.
func (c Combinator) String() string {
	switch c {
	case CombinatorNone:
		return ""
	case CombinatorDescendant:
		return " "
	case CombinatorChild:
		return ">"
	case CombinatorAdjacent:
		return "+"
	case CombinatorGeneral:
		return "~"
	default:
		return "?"
	}
}

// SimpleSelector represents a single atomic selector.
type SimpleSelector struct {
	Kind     SelectorKind // Type of selector
	Name     string       // Tag name, ID, class name, attr name, or pseudo-class name
	Operator AttrOperator // For attribute selectors
	Value    string       // For attribute selectors or functional pseudo-class arguments
}

// CompoundSelector is a sequence of simple selectors (e.g., div.foo#bar).
// All simple selectors must match for the compound to match.
type CompoundSelector struct {
	Selectors []SimpleSelector
}

// ComplexPart represents one step in a complex selector chain.
type ComplexPart struct {
	Combinator Combinator
	Compound   CompoundSelector
}

// ComplexSelector chains compound selectors with combinators.
// Represented as a list of (combinator, compound) pairs where the first
// combinator is always CombinatorNone.
type ComplexSelector struct {
	Parts []ComplexPart
}

// SelectorList represents comma-separated selectors.
// An element matches if it matches any selector in the list.
type SelectorList struct {
	Selectors []ComplexSelector
}

// selectorAST is a marker interface for parsed selector AST nodes.
type selectorAST interface {
	isSelectorAST()
}

func (ComplexSelector) isSelectorAST() {}
func (SelectorList) isSelectorAST()    {}
