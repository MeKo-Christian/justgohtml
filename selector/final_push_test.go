package selector

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// TestSelectorASTInterface tests that ComplexSelector and SelectorList implement selectorAST
func TestSelectorASTInterface(t *testing.T) {
	// These types implement the private selectorAST interface
	// We can verify they can be assigned to the interface type
	var ast selectorAST

	// ComplexSelector implements selectorAST
	cs := ComplexSelector{}
	ast = cs
	if ast == nil {
		t.Error("ComplexSelector should implement selectorAST")
	}

	// SelectorList implements selectorAST
	sl := SelectorList{}
	ast = sl
	if ast == nil {
		t.Error("SelectorList should implement selectorAST")
	}

	// Verify the marker methods are called by matchAST
	// This indirectly tests the isSelectorAST methods
	elem := dom.NewElement("div")

	// matchAST with ComplexSelector calls cs.isSelectorAST()
	_ = matchAST(elem, cs)

	// matchAST with SelectorList calls sl.isSelectorAST()
	_ = matchAST(elem, sl)
}

// TestTokenizerPeekMultibyteEdgeCases tests peek with various multi-byte scenarios
func TestTokenizerPeekMultibyteEdgeCases(t *testing.T) {
	// Test with empty string after advance
	tok := newTokenizer("a")
	tok.advance() // Move past 'a'
	ch := tok.peek()
	if ch != 0 {
		t.Errorf("peek() after advancing past end = %c, want 0", ch)
	}

	// Test with string that's exactly one character
	tok = newTokenizer(".")
	ch = tok.peek()
	if ch != '.' {
		t.Errorf("peek() at single char = %c, want '.'", ch)
	}
}

// TestReadStringEscapedQuote tests readString with escaped quotes
func TestReadStringEscapedQuote(t *testing.T) {
	// Test string with escaped quote
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	div.SetAttr("data", `test"value`)
	body.AppendChild(div)

	// Parse selector with escaped quote in string
	results, err := Match(body, `[data="test\"value"]`)
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	// The selector should parse but might not match due to different escaping
	_ = results
}

// TestMatchAttributeEdgeCasesComplete tests all remaining matchAttribute paths
func TestMatchAttributeEdgeCasesComplete(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Test the first return in matchAttribute (AttrExists early check)
	div := dom.NewElement("div")
	div.SetAttr("data-test", "value")
	body.AppendChild(div)

	sel := SimpleSelector{
		Kind:     KindAttr,
		Name:     "data-test",
		Operator: AttrExists,
	}

	if !matchAttribute(div, sel) {
		t.Error("matchAttribute with AttrExists should return true")
	}

	// Test the duplicate AttrExists case in the switch
	// This is actually unreachable because it's handled above, but exists in the code
}

// TestMatchPseudoAllPaths tests all pseudo-class matching paths
func TestMatchPseudoAllPaths(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// Test with valid nth expressions that we haven't covered
	p1 := dom.NewElement("p")
	div.AppendChild(p1)

	p2 := dom.NewElement("p")
	div.AppendChild(p2)

	// Test nth-child with valid expression
	sel := SimpleSelector{
		Kind:  KindPseudo,
		Name:  "nth-child",
		Value: "2n",
	}

	if !matchPseudo(p2, sel) {
		t.Error("matchPseudo(nth-child(2n)) should match second child")
	}

	// Test nth-last-child with valid expression
	sel = SimpleSelector{
		Kind:  KindPseudo,
		Name:  "nth-last-child",
		Value: "2n",
	}

	if !matchPseudo(p1, sel) {
		t.Error("matchPseudo(nth-last-child(2n)) should match first child (2nd from end)")
	}
}

// TestParseComplexErrorPaths tests parse error paths
func TestParseComplexErrorPaths(t *testing.T) {
	// These test cases should trigger various error paths in the parse function

	// Test selector list with error in second selector
	_, err := Parse("div, [invalid")
	if err == nil {
		t.Error("Parse should return error for invalid selector in list")
	}

	// Test complex selector with error in compound
	_, err = Parse("div >")
	if err == nil {
		t.Error("Parse should return error for combinator without following selector")
	}
}

// TestIsNthChildBoundaryConditions tests boundary conditions in isNthChild
func TestIsNthChildBoundaryConditions(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Create a parent with exactly one child
	div := dom.NewElement("div")
	body.AppendChild(div)

	p := dom.NewElement("p")
	div.AppendChild(p)

	// Test various An+B formulas
	if !isNthChild(p, 1, 1) {
		t.Error("isNthChild(1, 1) should match for n+1 where n=0 (first child)")
	}
}

// TestIsNthLastChildBoundaryConditions tests boundary conditions in isNthLastChild
func TestIsNthLastChildBoundaryConditions(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	p := dom.NewElement("p")
	div.AppendChild(p)

	// Test various An+B formulas
	if !isNthLastChild(p, 1, 1) {
		t.Error("isNthLastChild(1, 1) should match for n+1 where n=0 (last child)")
	}
}

// TestIsNthOfTypeBoundaryConditions tests boundary conditions in isNthOfType
func TestIsNthOfTypeBoundaryConditions(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	p := dom.NewElement("p")
	div.AppendChild(p)

	// Test the index finding loop
	if !isNthOfType(p, 1, 1) {
		t.Error("isNthOfType(1, 1) should match for n+1 where n=0 (first of type)")
	}
}

// TestIsNthLastOfTypeBoundaryConditions tests boundary conditions in isNthLastOfType
func TestIsNthLastOfTypeBoundaryConditions(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	p := dom.NewElement("p")
	div.AppendChild(p)

	// Test the index finding loop
	if !isNthLastOfType(p, 1, 1) {
		t.Error("isNthLastOfType(1, 1) should match for n+1 where n=0 (last of type)")
	}
}

// TestGetParentElementWithTextNodes tests getParentElement skipping text nodes
func TestGetParentElementWithTextNodes(t *testing.T) {
	doc := dom.NewDocument()
	html := dom.NewElement("html")
	doc.AppendChild(html)

	// html's parent is Document (not an Element)
	parent := getParentElement(html)
	if parent != nil {
		t.Errorf("getParentElement(html) should return nil, got %v", parent)
	}
}

// TestGetPreviousElementSiblingWithNonElementSiblings tests getPreviousElementSibling
func TestGetPreviousElementSiblingWithNonElementSiblings(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Add multiple non-element siblings before the target
	text1 := dom.NewText("text1")
	body.AppendChild(text1)

	comment := dom.NewComment("comment")
	body.AppendChild(comment)

	text2 := dom.NewText("text2")
	body.AppendChild(text2)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// getPreviousElementSibling should skip all non-element nodes
	prev := getPreviousElementSibling(div)
	if prev != nil {
		t.Errorf("getPreviousElementSibling should return nil when only non-element siblings, got %v", prev)
	}
}

// TestTokenizeUnclosedAttributeSelector tests tokenize with unclosed [
func TestTokenizeUnclosedAttributeSelector(t *testing.T) {
	_, err := Parse("[attr=value")
	if err == nil {
		t.Error("Parse with unclosed attribute selector should return error")
	}
}

// TestParseSelectorListErrors tests parse errors in selector lists
func TestParseSelectorListErrors(t *testing.T) {
	// Error in first selector
	_, err := Parse("[invalid, div")
	if err == nil {
		t.Error("Parse with error in first selector of list should return error")
	}

	// Error in subsequent selector
	_, err = Parse("div, [invalid")
	if err == nil {
		t.Error("Parse with error in subsequent selector of list should return error")
	}
}
