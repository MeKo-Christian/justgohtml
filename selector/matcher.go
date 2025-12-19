package selector

import (
	"strconv"
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// matchAST checks if an element matches a parsed selector AST.
func matchAST(elem *dom.Element, sel selectorAST) bool {
	switch s := sel.(type) {
	case ComplexSelector:
		return matchComplex(elem, s)
	case SelectorList:
		return matchSelectorList(elem, s)
	default:
		return false
	}
}

// matchSelectorList checks if an element matches any selector in the list.
func matchSelectorList(elem *dom.Element, list SelectorList) bool {
	for _, sel := range list.Selectors {
		if matchComplex(elem, sel) {
			return true
		}
	}
	return false
}

// matchComplex checks if an element matches a complex selector.
// Uses right-to-left matching strategy for efficiency.
func matchComplex(elem *dom.Element, sel ComplexSelector) bool {
	if len(sel.Parts) == 0 {
		return false
	}

	// Start with the rightmost compound selector
	lastIdx := len(sel.Parts) - 1
	if !matchCompound(elem, sel.Parts[lastIdx].Compound) {
		return false
	}

	// Work backwards through the remaining parts
	current := elem
	for i := lastIdx - 1; i >= 0; i-- {
		part := sel.Parts[i+1] // Get the combinator from the next part
		compound := sel.Parts[i].Compound

		switch part.Combinator {
		case CombinatorNone:
			// CombinatorNone should not appear in valid selector parts after the first
			return false

		case CombinatorDescendant:
			// Find any ancestor that matches
			found := false
			for ancestor := getParentElement(current); ancestor != nil; ancestor = getParentElement(ancestor) {
				if matchCompound(ancestor, compound) {
					current = ancestor
					found = true
					break
				}
			}
			if !found {
				return false
			}

		case CombinatorChild:
			// Check immediate parent
			parent := getParentElement(current)
			if parent == nil || !matchCompound(parent, compound) {
				return false
			}
			current = parent

		case CombinatorAdjacent:
			// Check immediately preceding sibling
			prev := getPreviousElementSibling(current)
			if prev == nil || !matchCompound(prev, compound) {
				return false
			}
			current = prev

		case CombinatorGeneral:
			// Find any preceding sibling that matches
			found := false
			for sib := getPreviousElementSibling(current); sib != nil; sib = getPreviousElementSibling(sib) {
				if matchCompound(sib, compound) {
					current = sib
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// matchCompound checks if an element matches all simple selectors in a compound.
func matchCompound(elem *dom.Element, compound CompoundSelector) bool {
	for _, sel := range compound.Selectors {
		if !matchSimple(elem, sel) {
			return false
		}
	}
	return true
}

// matchSimple checks if an element matches a single simple selector.
func matchSimple(elem *dom.Element, sel SimpleSelector) bool {
	switch sel.Kind {
	case KindTag:
		// Case-insensitive for HTML, case-sensitive for SVG/MathML
		if elem.Namespace == dom.NamespaceHTML {
			return strings.EqualFold(elem.TagName, sel.Name)
		}
		return elem.TagName == sel.Name

	case KindUniversal:
		return true

	case KindID:
		return elem.ID() == sel.Name

	case KindClass:
		return elem.HasClass(sel.Name)

	case KindAttr:
		return matchAttribute(elem, sel)

	case KindPseudo:
		return matchPseudo(elem, sel)

	default:
		return false
	}
}

// matchAttribute checks if an element matches an attribute selector.
func matchAttribute(elem *dom.Element, sel SimpleSelector) bool {
	if sel.Operator == AttrExists {
		return elem.HasAttr(sel.Name)
	}

	val := elem.Attr(sel.Name)
	if !elem.HasAttr(sel.Name) {
		return false
	}

	switch sel.Operator {
	case AttrExists:
		// Already handled above
		return true

	case AttrEquals:
		return val == sel.Value

	case AttrIncludes:
		// Word match (space-separated)
		words := strings.Fields(val)
		for _, w := range words {
			if w == sel.Value {
				return true
			}
		}
		return false

	case AttrDashPrefix:
		// Exact match or prefix followed by hyphen
		return val == sel.Value || strings.HasPrefix(val, sel.Value+"-")

	case AttrPrefixMatch:
		if sel.Value == "" {
			return false
		}
		return strings.HasPrefix(val, sel.Value)

	case AttrSuffixMatch:
		if sel.Value == "" {
			return false
		}
		return strings.HasSuffix(val, sel.Value)

	case AttrSubstring:
		if sel.Value == "" {
			return false
		}
		return strings.Contains(val, sel.Value)

	default:
		return false
	}
}

// matchPseudo checks if an element matches a pseudo-class selector.
func matchPseudo(elem *dom.Element, sel SimpleSelector) bool {
	switch sel.Name {
	case "first-child":
		return isFirstChild(elem)

	case "last-child":
		return isLastChild(elem)

	case "only-child":
		return isOnlyChild(elem)

	case "nth-child":
		a, b, ok := parseNthExpression(sel.Value)
		if !ok {
			return false
		}
		return isNthChild(elem, a, b)

	case "nth-last-child":
		a, b, ok := parseNthExpression(sel.Value)
		if !ok {
			return false
		}
		return isNthLastChild(elem, a, b)

	case "first-of-type":
		return isFirstOfType(elem)

	case "last-of-type":
		return isLastOfType(elem)

	case "only-of-type":
		return isOnlyOfType(elem)

	case "nth-of-type":
		a, b, ok := parseNthExpression(sel.Value)
		if !ok {
			return false
		}
		return isNthOfType(elem, a, b)

	case "nth-last-of-type":
		a, b, ok := parseNthExpression(sel.Value)
		if !ok {
			return false
		}
		return isNthLastOfType(elem, a, b)

	case "empty":
		return isEmpty(elem)

	case "root":
		return isRoot(elem)

	case "not":
		return matchNot(elem, sel.Value)

	default:
		// Unsupported pseudo-class
		return false
	}
}

// getParentElement returns the parent if it's an Element, nil otherwise.
func getParentElement(elem *dom.Element) *dom.Element {
	parent := elem.Parent()
	if parent == nil {
		return nil
	}
	if e, ok := parent.(*dom.Element); ok {
		return e
	}
	return nil
}

// getElementSiblings returns all element siblings (including the element itself).
func getElementSiblings(elem *dom.Element) []*dom.Element {
	parent := elem.Parent()
	if parent == nil {
		return []*dom.Element{elem}
	}

	var siblings []*dom.Element
	for _, child := range parent.Children() {
		if e, ok := child.(*dom.Element); ok {
			siblings = append(siblings, e)
		}
	}
	return siblings
}

// getElementIndex returns the 1-based index of the element among its siblings.
func getElementIndex(elem *dom.Element, siblings []*dom.Element) int {
	for i, sib := range siblings {
		if sib == elem {
			return i + 1 // 1-based
		}
	}
	return 0
}

// getPreviousElementSibling returns the previous element sibling or nil.
func getPreviousElementSibling(elem *dom.Element) *dom.Element {
	parent := elem.Parent()
	if parent == nil {
		return nil
	}

	var prev *dom.Element
	for _, child := range parent.Children() {
		if child == elem {
			return prev
		}
		if e, ok := child.(*dom.Element); ok {
			prev = e
		}
	}
	return nil
}

// getSiblingsOfSameType returns all element siblings with the same tag name.
func getSiblingsOfSameType(elem *dom.Element) []*dom.Element {
	parent := elem.Parent()
	if parent == nil {
		return []*dom.Element{elem}
	}

	var siblings []*dom.Element
	for _, child := range parent.Children() {
		if e, ok := child.(*dom.Element); ok {
			if strings.EqualFold(e.TagName, elem.TagName) {
				siblings = append(siblings, e)
			}
		}
	}
	return siblings
}

// isFirstChild checks if element is the first child among siblings.
func isFirstChild(elem *dom.Element) bool {
	siblings := getElementSiblings(elem)
	return len(siblings) > 0 && siblings[0] == elem
}

// isLastChild checks if element is the last child among siblings.
func isLastChild(elem *dom.Element) bool {
	siblings := getElementSiblings(elem)
	return len(siblings) > 0 && siblings[len(siblings)-1] == elem
}

// isOnlyChild checks if element is the only child.
func isOnlyChild(elem *dom.Element) bool {
	siblings := getElementSiblings(elem)
	return len(siblings) == 1 && siblings[0] == elem
}

// isNthChild checks if element matches :nth-child(An+B).
func isNthChild(elem *dom.Element, a, b int) bool {
	siblings := getElementSiblings(elem)
	index := getElementIndex(elem, siblings)
	if index == 0 {
		return false
	}
	return matchesNth(index, a, b)
}

// isNthLastChild checks if element matches :nth-last-child(An+B).
func isNthLastChild(elem *dom.Element, a, b int) bool {
	siblings := getElementSiblings(elem)
	index := getElementIndex(elem, siblings)
	if index == 0 {
		return false
	}
	// Convert to index from end
	indexFromEnd := len(siblings) - index + 1
	return matchesNth(indexFromEnd, a, b)
}

// isFirstOfType checks if element is the first of its type among siblings.
func isFirstOfType(elem *dom.Element) bool {
	siblings := getSiblingsOfSameType(elem)
	return len(siblings) > 0 && siblings[0] == elem
}

// isLastOfType checks if element is the last of its type among siblings.
func isLastOfType(elem *dom.Element) bool {
	siblings := getSiblingsOfSameType(elem)
	return len(siblings) > 0 && siblings[len(siblings)-1] == elem
}

// isOnlyOfType checks if element is the only one of its type.
func isOnlyOfType(elem *dom.Element) bool {
	siblings := getSiblingsOfSameType(elem)
	return len(siblings) == 1 && siblings[0] == elem
}

// isNthOfType checks if element matches :nth-of-type(An+B).
func isNthOfType(elem *dom.Element, a, b int) bool {
	siblings := getSiblingsOfSameType(elem)
	var index int
	for i, sib := range siblings {
		if sib == elem {
			index = i + 1
			break
		}
	}
	if index == 0 {
		return false
	}
	return matchesNth(index, a, b)
}

// isNthLastOfType checks if element matches :nth-last-of-type(An+B).
func isNthLastOfType(elem *dom.Element, a, b int) bool {
	siblings := getSiblingsOfSameType(elem)
	var index int
	for i, sib := range siblings {
		if sib == elem {
			index = i + 1
			break
		}
	}
	if index == 0 {
		return false
	}
	indexFromEnd := len(siblings) - index + 1
	return matchesNth(indexFromEnd, a, b)
}

// isEmpty checks if element has no element children and no non-whitespace text.
func isEmpty(elem *dom.Element) bool {
	for _, child := range elem.Children() {
		switch c := child.(type) {
		case *dom.Element:
			return false
		case *dom.Text:
			if strings.TrimSpace(c.Data) != "" {
				return false
			}
		}
	}
	return true
}

// isRoot checks if element is the root (parent is Document or DocumentFragment).
func isRoot(elem *dom.Element) bool {
	parent := elem.Parent()
	if parent == nil {
		return false
	}
	switch parent.(type) {
	case *dom.Document, *dom.DocumentFragment:
		return true
	}
	return false
}

// matchNot checks if element does NOT match the inner selector.
func matchNot(elem *dom.Element, arg string) bool {
	if arg == "" {
		return true
	}
	// Parse the inner selector
	innerSel, err := Parse(arg)
	if err != nil {
		// Parse error means we can't evaluate the inner selector,
		// so :not() doesn't match (returns false to not match the element)
		return false
	}
	matched := innerSel.Match(elem)
	return !matched
}

// parseNthExpression parses an An+B expression.
// Returns (a, b, ok) where the formula is index matches if (index - b) % a == 0 and (index - b) / a >= 0
func parseNthExpression(expr string) (int, int, bool) {
	expr = strings.TrimSpace(strings.ToLower(expr))

	// Handle special keywords
	switch expr {
	case "odd":
		return 2, 1, true
	case "even":
		return 2, 0, true
	}

	// Handle simple number
	if n, err := strconv.Atoi(expr); err == nil {
		return 0, n, true
	}

	// Parse An+B format
	// Examples: n, 2n, 2n+1, -n+3, n+5, -2n-1

	// Find position of 'n'
	nIdx := strings.Index(expr, "n")
	if nIdx == -1 {
		return 0, 0, false
	}

	// Parse 'a' part (before n)
	var a int
	aStr := expr[:nIdx]
	switch aStr {
	case "", "+":
		a = 1
	case "-":
		a = -1
	default:
		var err error
		a, err = strconv.Atoi(aStr)
		if err != nil {
			return 0, 0, false
		}
	}

	// Parse 'b' part (after n)
	var b int
	bStr := strings.TrimSpace(expr[nIdx+1:])
	if bStr == "" {
		b = 0
	} else {
		// Remove leading + if present
		bStr = strings.TrimPrefix(bStr, "+")
		var err error
		b, err = strconv.Atoi(bStr)
		if err != nil {
			return 0, 0, false
		}
	}

	return a, b, true
}

// matchesNth checks if index (1-based) matches the An+B formula.
func matchesNth(index, a, b int) bool {
	if a == 0 {
		// Just matches exact index
		return index == b
	}

	// Check if (index - b) is a non-negative multiple of a
	diff := index - b
	if a > 0 {
		return diff >= 0 && diff%a == 0
	}
	// a < 0: need diff <= 0 and divisible
	return diff <= 0 && diff%a == 0
}
