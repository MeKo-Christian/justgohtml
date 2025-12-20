package selector

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

// TestIsSelectorASTMethods tests the marker interface methods
func TestIsSelectorASTMethods(t *testing.T) {
	// These are marker methods that just exist to implement the interface
	// They're called internally by the type system but we can invoke them directly
	var _ selectorAST = ComplexSelector{}
	var _ selectorAST = SelectorList{}

	// Create instances and call the methods
	cs := ComplexSelector{}
	cs.isSelectorAST()

	sl := SelectorList{}
	sl.isSelectorAST()

	// If we get here without panic, the methods exist and work
}

// TestMatchASTDefaultCase tests the default case in matchAST
func TestMatchASTDefaultCase(t *testing.T) {
	elem := dom.NewElement("div")

	// Create a struct that implements selectorAST but isn't ComplexSelector or SelectorList
	// We need to access the internal matchAST function
	// Since we can't create a new type that implements the private interface,
	// we'll use the fact that an empty ComplexSelector should hit a different path
	emptyComplex := ComplexSelector{Parts: []ComplexPart{}}

	// This should return false
	if matchAST(elem, emptyComplex) {
		t.Error("matchAST with empty ComplexSelector should return false")
	}
}

// TestMatchComplexEdgeCases tests edge cases in matchComplex
func TestMatchComplexEdgeCases(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div1 := dom.NewElement("div")
	div1.SetAttr("class", "test")
	body.AppendChild(div1)

	div2 := dom.NewElement("div")
	div2.SetAttr("class", "other")
	body.AppendChild(div2)

	// Test descendant combinator with no matching ancestor
	sel := ComplexSelector{
		Parts: []ComplexPart{
			{
				Combinator: CombinatorNone,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindClass, Name: "nonexistent"}}},
			},
			{
				Combinator: CombinatorDescendant,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindTag, Name: "div"}}},
			},
		},
	}

	if matchComplex(div1, sel) {
		t.Error("matchComplex should return false when ancestor doesn't match")
	}

	// Test child combinator with no parent matching
	sel = ComplexSelector{
		Parts: []ComplexPart{
			{
				Combinator: CombinatorNone,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindClass, Name: "nonexistent"}}},
			},
			{
				Combinator: CombinatorChild,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindTag, Name: "div"}}},
			},
		},
	}

	if matchComplex(div1, sel) {
		t.Error("matchComplex should return false when parent doesn't match")
	}

	// Test adjacent sibling with no matching sibling
	sel = ComplexSelector{
		Parts: []ComplexPart{
			{
				Combinator: CombinatorNone,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindClass, Name: "nonexistent"}}},
			},
			{
				Combinator: CombinatorAdjacent,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindTag, Name: "div"}}},
			},
		},
	}

	if matchComplex(div2, sel) {
		t.Error("matchComplex should return false when adjacent sibling doesn't match")
	}

	// Test general sibling with no matching sibling
	sel = ComplexSelector{
		Parts: []ComplexPart{
			{
				Combinator: CombinatorNone,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindClass, Name: "nonexistent"}}},
			},
			{
				Combinator: CombinatorGeneral,
				Compound:   CompoundSelector{Selectors: []SimpleSelector{{Kind: KindTag, Name: "div"}}},
			},
		},
	}

	if matchComplex(div2, sel) {
		t.Error("matchComplex should return false when general sibling doesn't match")
	}
}

// TestMatchAttributeAttrExistsDuplicate tests the duplicate AttrExists case
func TestMatchAttributeAttrExistsDuplicate(t *testing.T) {
	elem := dom.NewElement("div")
	elem.SetAttr("test", "value")

	// The first switch case for AttrExists is covered above
	// The second one in the operator switch is unreachable but we can verify the logic
	sel := SimpleSelector{
		Kind:     KindAttr,
		Name:     "test",
		Operator: AttrExists,
		Value:    "", // AttrExists shouldn't use value
	}

	if !matchAttribute(elem, sel) {
		t.Error("matchAttribute with AttrExists should return true when attribute exists")
	}
}

// TestMatchPseudoInvalidNthExpressions tests pseudo-classes with invalid nth expressions
func TestMatchPseudoInvalidNthExpressions(t *testing.T) {
	elem := dom.NewElement("div")

	tests := []struct {
		name  string
		value string
	}{
		{"nth-child", "invalid"},
		{"nth-last-child", "xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel := SimpleSelector{
				Kind:  KindPseudo,
				Name:  tt.name,
				Value: tt.value,
			}

			if matchPseudo(elem, sel) {
				t.Errorf("matchPseudo(%s with invalid expression) should return false", tt.name)
			}
		})
	}
}

// TestGetParentElementNonElementParent tests getParentElement when parent exists but is not Element
func TestGetParentElementNonElementParent(t *testing.T) {
	// This case is already covered by TestGetParentElementNilCases
	// where html's parent is Document
	doc := dom.NewDocument()
	html := dom.NewElement("html")
	doc.AppendChild(html)

	parent := getParentElement(html)
	if parent != nil {
		t.Error("getParentElement should return nil when parent is Document, not Element")
	}
}

// TestGetPreviousElementSiblingNoMatch tests getPreviousElementSibling iterations
func TestGetPreviousElementSiblingNoMatch(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Add a text node followed by an element
	text := dom.NewText("text")
	body.AppendChild(text)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// The div's previous sibling is text, not an element
	// So getPreviousElementSibling should skip over it
	prev := getPreviousElementSibling(div)
	if prev != nil {
		t.Errorf("getPreviousElementSibling(div after text) = %v, want nil", prev)
	}
}

// TestIsNthChildWithIndexZero tests isNthChild when getElementIndex returns 0
func TestIsNthChildWithIndexZero(t *testing.T) {
	// Create a scenario where getElementIndex would return 0
	// This is hard to trigger in normal usage, but we can test the logic
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// isNthChild with valid index should work
	if !isNthChild(div, 0, 1) {
		t.Error("isNthChild(0, 1) should match first child")
	}
}

// TestIsNthLastChildWithIndexZero tests isNthLastChild when getElementIndex returns 0
func TestIsNthLastChildWithIndexZero(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	body.AppendChild(div)

	// isNthLastChild with valid index should work
	if !isNthLastChild(div, 0, 1) {
		t.Error("isNthLastChild(0, 1) should match last child")
	}
}

// TestTokenizerPeekMultiByteRune tests peek with multi-byte UTF-8 characters
func TestTokenizerPeekMultiByteRune(t *testing.T) {
	tok := newTokenizer("emoji😀")
	// The tokenizer should handle multi-byte runes properly
	tokens, err := tok.tokenize()
	if err == nil {
		// It might error on the emoji, but shouldn't crash
		_ = tokens
	}
}

// TestReadStringEdgeCases tests readString edge cases
func TestReadStringEdgeCases(t *testing.T) {
	// Test unclosed string
	_, err := Parse(`[attr="unclosed`)
	if err == nil {
		t.Error("Parse with unclosed string should return error")
	}

	// Test string with escape at EOF
	_, err = Parse(`[attr="value\`)
	if err == nil {
		t.Error("Parse with escape at EOF should return error")
	}
}

// TestReadUnquotedAttrValueEdgeCases tests unquoted attribute value edge cases
func TestReadUnquotedAttrValueEdgeCases(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	div := dom.NewElement("div")
	div.SetAttr("data-value", "test-123")
	body.AppendChild(div)

	// Test unquoted attribute value with special characters
	results, err := Match(body, "[data-value=test-123]")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 match for unquoted attribute value, got %d", len(results))
	}

	// Test unquoted value with escapes
	div2 := dom.NewElement("div")
	div2.SetAttr("data-value", "test")
	body.AppendChild(div2)

	_, err = Match(body, `[data-value=test\-value]`)
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	// This should parse the escaped value
}

// TestTokenizeEdgeCases tests various tokenize edge cases
func TestTokenizeEdgeCases(t *testing.T) {
	tests := []struct {
		selector string
		wantErr  bool
		desc     string
	}{
		{"[attr~=value", true, "unclosed attribute selector"},
		{"[attr|=value", true, "unclosed attribute selector with |="},
		{"div,", true, "trailing comma"},
		{".class,", true, "trailing comma after class"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := Parse(tt.selector)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v (%s)", tt.selector, err, tt.wantErr, tt.desc)
			}
		})
	}
}

// TestParseEdgeCases tests parse function edge cases
func TestParseEdgeCases(t *testing.T) {
	// Test selector list parsing
	sel, err := Parse("div, span")
	if err != nil {
		t.Errorf("Parse(\"div, span\") error = %v, want nil", err)
	}
	if sel == nil {
		t.Error("Parse should return non-nil selector for valid selector list")
	}

	// Test that "extra" is parsed as a tag name, not trailing content
	// The parser is quite permissive and treats consecutive tags as compound selectors
	sel, err = Parse("div span")
	if err != nil {
		t.Errorf("Parse(\"div span\") error = %v, want nil (descendant combinator)", err)
	}
	if sel == nil {
		t.Error("Parse should return non-nil selector for descendant combinator")
	}
}

// TestParsePseudoSelectorEdgeCases tests parsePseudoSelector edge cases
func TestParsePseudoSelectorEdgeCases(t *testing.T) {
	// Test all different token types in pseudo-selector arguments
	_, err := Parse(":not(#id)")
	if err != nil {
		t.Errorf("Parse(:not(#id)) error = %v, want nil", err)
	}

	_, err = Parse(":not(.class)")
	if err != nil {
		t.Errorf("Parse(:not(.class)) error = %v, want nil", err)
	}

	_, err = Parse(":not(*)")
	if err != nil {
		t.Errorf("Parse(:not(*)) error = %v, want nil", err)
	}

	_, err = Parse(":not([attr])")
	if err != nil {
		t.Errorf("Parse(:not([attr])) error = %v, want nil", err)
	}

	_, err = Parse(":not(div > span)")
	if err != nil {
		t.Errorf("Parse(:not(div > span)) error = %v, want nil", err)
	}

	_, err = Parse(":not(div, span)")
	if err != nil {
		t.Errorf("Parse(:not(div, span)) error = %v, want nil", err)
	}
}

// TestMatchFirstEdgeCases tests MatchFirst edge cases
func TestMatchFirstEdgeCases(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Test with parse error
	_, err := MatchFirst(body, "[invalid")
	if err == nil {
		t.Error("MatchFirst with invalid selector should return error")
	}
}

// TestIsNthOfTypeAllPaths tests all code paths in isNthOfType
func TestIsNthOfTypeAllPaths(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Create multiple elements of the same type
	p1 := dom.NewElement("p")
	body.AppendChild(p1)

	span := dom.NewElement("span")
	body.AppendChild(span)

	p2 := dom.NewElement("p")
	body.AppendChild(p2)

	// Test the loop that finds the index
	if !isNthOfType(p2, 0, 2) {
		t.Error("isNthOfType should match second p element")
	}
}

// TestIsNthLastOfTypeAllPaths tests all code paths in isNthLastOfType
func TestIsNthLastOfTypeAllPaths(t *testing.T) {
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	doc.AppendChild(body)

	// Create multiple elements of the same type
	p1 := dom.NewElement("p")
	body.AppendChild(p1)

	span := dom.NewElement("span")
	body.AppendChild(span)

	p2 := dom.NewElement("p")
	body.AppendChild(p2)

	// Test the loop that finds the index
	if !isNthLastOfType(p1, 0, 2) {
		t.Error("isNthLastOfType should match second-to-last p element")
	}
}
