package selector

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// TestASTStringMethods tests String() methods on AST types
func TestASTStringMethods(t *testing.T) {
	// Test SelectorKind.String()
	kindTests := []struct {
		kind SelectorKind
		want string
	}{
		{KindTag, "Tag"},
		{KindUniversal, "Universal"},
		{KindID, "ID"},
		{KindClass, "Class"},
		{KindAttr, "Attr"},
		{KindPseudo, "Pseudo"},
		{SelectorKind(999), "Unknown"},
	}

	for _, tt := range kindTests {
		got := tt.kind.String()
		if got != tt.want {
			t.Errorf("SelectorKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}

	// Test AttrOperator.String()
	opTests := []struct {
		op   AttrOperator
		want string
	}{
		{AttrExists, ""},
		{AttrEquals, "="},
		{AttrIncludes, "~="},
		{AttrDashPrefix, "|="},
		{AttrPrefixMatch, "^="},
		{AttrSuffixMatch, "$="},
		{AttrSubstring, "*="},
		{AttrOperator(999), "?"},
	}

	for _, tt := range opTests {
		got := tt.op.String()
		if got != tt.want {
			t.Errorf("AttrOperator(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}

	// Test Combinator.String()
	combTests := []struct {
		comb Combinator
		want string
	}{
		{CombinatorNone, ""},
		{CombinatorDescendant, " "},
		{CombinatorChild, ">"},
		{CombinatorAdjacent, "+"},
		{CombinatorGeneral, "~"},
		{Combinator(999), "?"},
	}

	for _, tt := range combTests {
		got := tt.comb.String()
		if got != tt.want {
			t.Errorf("Combinator(%d).String() = %q, want %q", tt.comb, got, tt.want)
		}
	}
}

// TestNthOfTypePseudoClasses tests :nth-of-type and :nth-last-of-type
func TestNthOfTypePseudoClasses(t *testing.T) {
	// Create a DOM with multiple elements of the same type
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Add paragraphs with mixed siblings
	div := dom.NewElement("div")
	body.AppendChild(div)

	// First p (1st of type)
	p1 := dom.NewElement("p")
	p1.SetAttr("id", "p1")
	div.AppendChild(p1)

	// span (not a p)
	span1 := dom.NewElement("span")
	div.AppendChild(span1)

	// Second p (2nd of type)
	p2 := dom.NewElement("p")
	p2.SetAttr("id", "p2")
	div.AppendChild(p2)

	// Another span
	span2 := dom.NewElement("span")
	div.AppendChild(span2)

	// Third p (3rd of type)
	p3 := dom.NewElement("p")
	p3.SetAttr("id", "p3")
	div.AppendChild(p3)

	tests := []struct {
		selector string
		wantIDs  []string
	}{
		{"p:nth-of-type(1)", []string{"p1"}},
		{"p:nth-of-type(2)", []string{"p2"}},
		{"p:nth-of-type(3)", []string{"p3"}},
		{"p:nth-of-type(odd)", []string{"p1", "p3"}},
		{"p:nth-of-type(even)", []string{"p2"}},
		{"p:nth-of-type(2n)", []string{"p2"}},
		{"p:nth-of-type(2n+1)", []string{"p1", "p3"}},
		{"p:nth-last-of-type(1)", []string{"p3"}},
		{"p:nth-last-of-type(2)", []string{"p2"}},
		{"p:nth-last-of-type(3)", []string{"p1"}},
	}

	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			results, err := Match(body, tt.selector)
			if err != nil {
				t.Fatalf("Match(%q) error: %v", tt.selector, err)
			}

			if len(results) != len(tt.wantIDs) {
				t.Errorf("Match(%q) = %d elements, want %d", tt.selector, len(results), len(tt.wantIDs))
				return
			}

			gotIDs := make([]string, len(results))
			for i, elem := range results {
				gotIDs[i] = elem.ID()
			}

			for i, wantID := range tt.wantIDs {
				if gotIDs[i] != wantID {
					t.Errorf("Match(%q)[%d].ID() = %q, want %q", tt.selector, i, gotIDs[i], wantID)
				}
			}
		})
	}
}

// TestUnquotedAttributeValues tests parsing of unquoted attribute values
func TestUnquotedAttributeValues(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Create elements with various attributes
	div1 := dom.NewElement("div")
	div1.SetAttr("data-value", "test123")
	body.AppendChild(div1)

	div2 := dom.NewElement("div")
	div2.SetAttr("data-value", "other")
	body.AppendChild(div2)

	tests := []struct {
		selector string
		expected int
	}{
		// Unquoted attribute values
		{"[data-value=test123]", 1},
		{"[data-value=other]", 1},
		// Quoted for comparison
		{"[data-value=\"test123\"]", 1},
		{"[data-value='other']", 1},
	}

	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			results, err := Match(body, tt.selector)
			if err != nil {
				t.Fatalf("Match(%q) error: %v", tt.selector, err)
			}
			if len(results) != tt.expected {
				t.Errorf("Match(%q) = %d elements, want %d", tt.selector, len(results), tt.expected)
			}
		})
	}
}

// TestEmptyAttributeMatcherEdgeCases tests edge cases in attribute matching
func TestEmptyAttributeMatcherEdgeCases(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	div.SetAttr("data-value", "test")
	div.SetAttr("data-empty", "")
	body.AppendChild(div)

	tests := []struct {
		selector string
		expected int
		desc     string
	}{
		{"[data-value^=\"\"]", 0, "empty prefix should not match"},
		{"[data-value$=\"\"]", 0, "empty suffix should not match"},
		{"[data-value*=\"\"]", 0, "empty substring should not match"},
		{"[data-empty]", 1, "attribute exists with empty value"},
		{"[data-empty=\"\"]", 1, "exact match for empty value"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			results, err := Match(body, tt.selector)
			if err != nil {
				t.Fatalf("Match(%q) error: %v", tt.selector, err)
			}
			if len(results) != tt.expected {
				t.Errorf("Match(%q) = %d elements, want %d (%s)", tt.selector, len(results), tt.expected, tt.desc)
			}
		})
	}
}

// TestGetParentElementNilCases tests edge cases in getParentElement
func TestGetParentElementNilCases(t *testing.T) {
	doc := dom.NewDocument()
	html := dom.NewElement("html")
	doc.AppendChild(html)

	// html's parent is Document, not Element
	parent := getParentElement(html)
	if parent != nil {
		t.Errorf("getParentElement(html) = %v, want nil (parent is Document)", parent)
	}
}

// TestGetElementIndexEdgeCases tests edge cases in getElementIndex
func TestGetElementIndexEdgeCases(t *testing.T) {
	elem := dom.NewElement("div")
	siblings := []*dom.Element{dom.NewElement("p"), dom.NewElement("span")}

	// Element not in siblings list
	index := getElementIndex(elem, siblings)
	if index != 0 {
		t.Errorf("getElementIndex(elem not in list) = %d, want 0", index)
	}
}

// TestGetPreviousElementSiblingEdgeCases tests edge cases
func TestGetPreviousElementSiblingEdgeCases(t *testing.T) {
	// Element with no parent
	elem := dom.NewElement("div")
	prev := getPreviousElementSibling(elem)
	if prev != nil {
		t.Errorf("getPreviousElementSibling(no parent) = %v, want nil", prev)
	}

	// First element sibling
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	prev = getPreviousElementSibling(div)
	if prev != nil {
		t.Errorf("getPreviousElementSibling(first child) = %v, want nil", prev)
	}
}

// TestGetSiblingsOfSameTypeEdgeCases tests edge cases
func TestGetSiblingsOfSameTypeEdgeCases(t *testing.T) {
	// Element with no parent
	elem := dom.NewElement("div")
	siblings := getSiblingsOfSameType(elem)
	if len(siblings) != 1 || siblings[0] != elem {
		t.Errorf("getSiblingsOfSameType(no parent) = %v, want [elem]", siblings)
	}
}

// TestIsEmptyWithComment tests isEmpty with comment nodes
func TestIsEmptyWithComment(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Div with only a comment (should be considered empty)
	div1 := dom.NewElement("div")
	div1.SetAttr("id", "with-comment")
	div1.AppendChild(dom.NewComment("just a comment"))
	body.AppendChild(div1)

	// Div with whitespace text (should be considered empty)
	div2 := dom.NewElement("div")
	div2.SetAttr("id", "with-whitespace")
	div2.AppendChild(dom.NewText("   \n\t  "))
	body.AppendChild(div2)

	// Div with non-whitespace text (should NOT be empty)
	div3 := dom.NewElement("div")
	div3.SetAttr("id", "with-text")
	div3.AppendChild(dom.NewText("text"))
	body.AppendChild(div3)

	tests := []struct {
		selector string
		wantIDs  []string
	}{
		{"div:empty", []string{"with-comment", "with-whitespace"}},
	}

	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			results, err := Match(body, tt.selector)
			if err != nil {
				t.Fatalf("Match(%q) error: %v", tt.selector, err)
			}

			if len(results) != len(tt.wantIDs) {
				t.Errorf("Match(%q) = %d elements, want %d", tt.selector, len(results), len(tt.wantIDs))
			}
		})
	}
}

// TestIsRootWithDocumentFragment tests :root with DocumentFragment
func TestIsRootWithDocumentFragment(t *testing.T) {
	frag := dom.NewDocumentFragment()
	div := dom.NewElement("div")
	frag.AppendChild(div)

	// Element with DocumentFragment as parent should match :root
	if !isRoot(div) {
		t.Error("isRoot(elem with DocumentFragment parent) = false, want true")
	}
}

// TestMatchNotEdgeCases tests :not() edge cases
func TestMatchNotEdgeCases(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	div.SetAttr("class", "test")
	body.AppendChild(div)

	tests := []struct {
		selector string
		expected int
		desc     string
	}{
		{"div:not()", 1, ":not() with empty arg should match"},
		{"div:not([invalid selector)", 0, ":not() with invalid selector should not match"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			results, err := Match(body, tt.selector)
			if err != nil {
				t.Fatalf("Match(%q) error: %v", tt.selector, err)
			}
			if len(results) != tt.expected {
				t.Errorf("Match(%q) = %d elements, want %d (%s)", tt.selector, len(results), tt.expected, tt.desc)
			}
		})
	}
}

// TestMatchASTUnknownType tests matchAST with unknown type
func TestMatchASTUnknownType(t *testing.T) {
	elem := dom.NewElement("div")

	// SimpleSelector doesn't implement selectorAST, so this should return false
	// We need to use an invalid type
	var invalidAST selectorAST = ComplexSelector{Parts: []ComplexPart{}}

	// ComplexSelector with empty parts should not match
	if matchAST(elem, invalidAST) {
		t.Error("matchAST with empty ComplexSelector should return false")
	}
}

// TestComplexSelectorCombinatorNone tests invalid CombinatorNone in non-first position
func TestComplexSelectorCombinatorNone(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// Manually construct an invalid selector with CombinatorNone in non-first position
	invalidSel := ComplexSelector{
		Parts: []ComplexPart{
			{
				Combinator: CombinatorNone,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindTag, Name: "body"}}},
			},
			{
				Combinator: CombinatorNone, // Invalid: CombinatorNone in non-first position
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindTag, Name: "div"}}},
			},
		},
	}

	if matchComplex(div, invalidSel) {
		t.Error("matchComplex with CombinatorNone in non-first position should return false")
	}
}

// TestTokenizerPeekEdgeCases tests tokenizer peek at EOF
func TestTokenizerPeekEdgeCases(t *testing.T) {
	tok := newTokenizer("")
	ch := tok.peek()
	if ch != 0 {
		t.Errorf("peek() at EOF = %c, want 0", ch)
	}

	ch = tok.advance()
	if ch != 0 {
		t.Errorf("advance() at EOF = %c, want 0", ch)
	}
}

// TestTokenizerReadNameWithEscapes tests escape sequences in names
func TestTokenizerReadNameWithEscapes(t *testing.T) {
	// Test parsing selectors with escaped characters in class/id names
	tests := []struct {
		selector string
		wantErr  bool
	}{
		{`.class\:name`, false}, // Escaped : in class name
		{`#id\-name`, false},    // Escaped - in id name
	}

	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			_, err := Parse(tt.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.selector, err, tt.wantErr)
			}
		})
	}
}

// TestParserPeekEdgeCases tests parser peek with no tokens
func TestParserPeekEdgeCases(t *testing.T) {
	p := newParser([]token{}, "")
	tok := p.peek()
	if tok.typ != tokenEOF {
		t.Errorf("peek() with empty tokens = %v, want EOF", tok.typ)
	}
}

// TestPseudoSelectorWithNestedParens tests functional pseudo-class with nested parens
func TestPseudoSelectorWithNestedParens(t *testing.T) {
	// This tests the parsePseudoSelector depth tracking
	_, err := Parse(":not(:not(div))")
	if err != nil {
		t.Errorf("Parse(:not(:not(div))) error = %v, want nil", err)
	}
}

// TestNthOfTypeZeroIndex tests nth-of-type when element is not found
func TestNthOfTypeZeroIndex(t *testing.T) {
	// The actual code paths that return index=0 are unreachable in normal usage
	// because getSiblingsOfSameType always includes the element itself.
	// However, we can verify the logic works correctly:

	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// Test that nth-of-type(1) matches the only element of its type
	if !isNthOfType(div, 0, 1) {
		t.Error("isNthOfType(0, 1) should match element at index 1")
	}

	// Test that nth-last-of-type(1) matches the only element of its type
	if !isNthLastOfType(div, 0, 1) {
		t.Error("isNthLastOfType(0, 1) should match element at index 1")
	}
}

// TestMatchFirstNoMatch tests MatchFirst with no matching elements
func TestMatchFirstNoMatch(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	result, err := MatchFirst(body, ".nonexistent")
	if err != nil {
		t.Fatalf("MatchFirst error: %v", err)
	}
	if result != nil {
		t.Errorf("MatchFirst(.nonexistent) = %v, want nil", result)
	}
}

// TestMatchWithInvalidSelector tests Match with parse errors
func TestMatchWithInvalidSelector(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	_, err := Match(body, "[invalid")
	if err == nil {
		t.Error("Match([invalid) error = nil, want error")
	}
}

// TestMatchSimpleUnknownKind tests matchSimple with unknown selector kind
func TestMatchSimpleUnknownKind(t *testing.T) {
	elem := dom.NewElement("div")
	sel := SimpleSelector{Kind: SelectorKind(999), Name: "unknown"}

	if matchSimple(elem, sel) {
		t.Error("matchSimple with unknown kind should return false")
	}
}

// TestMatchPseudoUnknownPseudoClass tests matchPseudo with unsupported pseudo-class
func TestMatchPseudoUnknownPseudoClass(t *testing.T) {
	elem := dom.NewElement("div")
	sel := SimpleSelector{Kind: KindPseudo, Name: "unknown-pseudo"}

	if matchPseudo(elem, sel) {
		t.Error("matchPseudo with unknown pseudo-class should return false")
	}
}

// TestParseNthExpressionInvalidFormats tests parseNthExpression with invalid input
func TestParseNthExpressionInvalidFormats(t *testing.T) {
	tests := []string{
		"invalid",
		"xyz",
		"n+",
		"abc",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, _, ok := parseNthExpression(expr)
			if ok {
				t.Errorf("parseNthExpression(%q) ok = true, want false", expr)
			}
		})
	}
}

// TestSVGElementTagMatching tests case-sensitive tag matching for SVG elements
func TestSVGElementTagMatching(t *testing.T) {
	doc := dom.NewDocument()
	svg := dom.NewElement("svg")
	svg.Namespace = dom.NamespaceSVG
	doc.AppendChild(svg)

	circle := dom.NewElement("circle")
	circle.Namespace = dom.NamespaceSVG
	svg.AppendChild(circle)

	// SVG elements should be case-sensitive
	sel := SimpleSelector{Kind: KindTag, Name: "circle"}
	if !matchSimple(circle, sel) {
		t.Error("matchSimple(circle, 'circle') = false, want true for SVG")
	}

	sel = SimpleSelector{Kind: KindTag, Name: "Circle"}
	if matchSimple(circle, sel) {
		t.Error("matchSimple(circle, 'Circle') = true, want false for SVG (case-sensitive)")
	}
}

// TestMatchAttributeDefaultCase tests default case in matchAttribute
func TestMatchAttributeDefaultCase(t *testing.T) {
	elem := dom.NewElement("div")
	elem.SetAttr("test", "value")

	// This should hit the default case by using an invalid operator
	sel := SimpleSelector{
		Kind:     KindAttr,
		Name:     "test",
		Operator: AttrOperator(999),
		Value:    "value",
	}

	if matchAttribute(elem, sel) {
		t.Error("matchAttribute with invalid operator should return false")
	}
}
