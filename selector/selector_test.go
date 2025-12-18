package selector

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

const testMainID = "main"

// Helper to create a test DOM tree
func createTestDOM() *dom.Document {
	doc := dom.NewDocument()

	html := dom.NewElement("html")
	doc.AppendChild(html)

	head := dom.NewElement("head")
	html.AppendChild(head)

	title := dom.NewElement("title")
	title.AppendChild(dom.NewText("Test"))
	head.AppendChild(title)

	body := dom.NewElement("body")
	html.AppendChild(body)

	// Create a div with id and class
	div1 := dom.NewElement("div")
	div1.SetAttr("id", testMainID)
	div1.SetAttr("class", "container active")
	body.AppendChild(div1)

	// Create nested p elements
	p1 := dom.NewElement("p")
	p1.SetAttr("class", "intro")
	p1.AppendChild(dom.NewText("First paragraph"))
	div1.AppendChild(p1)

	p2 := dom.NewElement("p")
	p2.SetAttr("class", "content")
	p2.AppendChild(dom.NewText("Second paragraph"))
	div1.AppendChild(p2)

	// Create a span inside p1
	span := dom.NewElement("span")
	span.SetAttr("class", "highlight")
	span.AppendChild(dom.NewText("highlighted"))
	p1.AppendChild(span)

	// Create a second div
	div2 := dom.NewElement("div")
	div2.SetAttr("id", "sidebar")
	div2.SetAttr("class", "container")
	body.AppendChild(div2)

	// Add some list items
	ul := dom.NewElement("ul")
	div2.AppendChild(ul)

	for i := range 5 {
		li := dom.NewElement("li")
		if i%2 == 0 {
			li.SetAttr("class", "odd")
		} else {
			li.SetAttr("class", "even")
		}
		ul.AppendChild(li)
	}

	// Create an empty div
	emptyDiv := dom.NewElement("div")
	emptyDiv.SetAttr("class", "empty")
	body.AppendChild(emptyDiv)

	// Create a div with data attribute
	dataDiv := dom.NewElement("div")
	dataDiv.SetAttr("data-value", "test-value")
	dataDiv.SetAttr("data-lang", "en-US")
	body.AppendChild(dataDiv)

	return doc
}

// TestParseBasicSelectors tests parsing of basic selectors
func TestParseBasicSelectors(t *testing.T) {
	tests := []struct {
		selector string
		wantErr  bool
	}{
		{"div", false},
		{"*", false},
		{"#main", false},
		{".container", false},
		{"div.container", false},
		{"div#main.container", false},
		{"[href]", false},
		{"[href=\"test\"]", false},
		{"div p", false},
		{"div > p", false},
		{"div + p", false},
		{"div ~ p", false},
		{":first-child", false},
		{":nth-child(2n+1)", false},
		{":not(.active)", false},
		{"", true},
		{"   ", true},
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

// TestTagSelector tests tag name matching
func TestTagSelector(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "div")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 4 { // main, sidebar, empty, data
		t.Errorf("Expected 4 divs, got %d", len(results))
	}

	results, err = Match(body, "p")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 p elements, got %d", len(results))
	}
}

// TestUniversalSelector tests universal selector
func TestUniversalSelector(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "*")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	// Should match all elements under body
	if len(results) < 10 {
		t.Errorf("Expected many elements, got %d", len(results))
	}
}

// TestIDSelector tests ID selector
func TestIDSelector(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "#main")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 element with #main, got %d", len(results))
	}
	if results[0].ID() != testMainID {
		t.Errorf("Expected id='main', got %q", results[0].ID())
	}

	results, err = Match(body, "#sidebar")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 element with #sidebar, got %d", len(results))
	}
}

// TestClassSelector tests class selector
func TestClassSelector(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, ".container")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 { // main and sidebar both have .container
		t.Errorf("Expected 2 elements with .container, got %d", len(results))
	}

	results, err = Match(body, ".active")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 element with .active, got %d", len(results))
	}
}

// TestCompoundSelector tests compound selectors
func TestCompoundSelector(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// div.container
	results, err := Match(body, "div.container")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 div.container, got %d", len(results))
	}

	// div#main
	results, err = Match(body, "div#main")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 div#main, got %d", len(results))
	}

	// div.container.active
	results, err = Match(body, "div.container.active")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 div.container.active, got %d", len(results))
	}
}

// TestAttributeSelectors tests attribute selectors
func TestAttributeSelectors(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	tests := []struct {
		selector string
		expected int
	}{
		{"[id]", 2},                     // main and sidebar
		{"[id=\"main\"]", 1},            // exact match
		{"[class~=\"container\"]", 2},   // word match
		{"[data-lang|=\"en\"]", 1},      // prefix match
		{"[data-value^=\"test\"]", 1},   // starts with
		{"[data-value$=\"value\"]", 1},  // ends with
		{"[data-value*=\"-\"]", 1},      // contains
		{"[data-nonexistent]", 0},       // no match
		{"[class~=\"nonexistent\"]", 0}, // no match
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

// TestDescendantCombinator tests descendant combinator (space)
func TestDescendantCombinator(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// div p (p descendants of div)
	results, err := Match(body, "div p")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 p elements in divs, got %d", len(results))
	}

	// body div (div descendants of body)
	// body is the root, so we need to start from a parent
	// Actually, starting from body, it won't match 'body' itself
	// Let's test from html
	html := doc.DocumentElement()
	results, err = Match(html, "body div")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 div descendants of body, got %d", len(results))
	}
}

// TestChildCombinator tests child combinator (>)
func TestChildCombinator(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// div > p (direct children only)
	results, err := Match(body, "div > p")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 direct p children of div, got %d", len(results))
	}

	// body > div (direct div children of body)
	html := doc.DocumentElement()
	results, err = Match(html, "body > div")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 direct div children of body, got %d", len(results))
	}
}

// TestAdjacentSiblingCombinator tests adjacent sibling combinator (+)
func TestAdjacentSiblingCombinator(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// p + p (p immediately after p)
	results, err := Match(body, "p + p")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 p immediately after p, got %d", len(results))
	}

	// li + li (li immediately after li)
	results, err = Match(body, "li + li")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 4 { // 4 li elements have a preceding li
		t.Errorf("Expected 4 li+li, got %d", len(results))
	}
}

// TestGeneralSiblingCombinator tests general sibling combinator (~)
func TestGeneralSiblingCombinator(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// p.intro ~ p (p after p.intro)
	results, err := Match(body, "p.intro ~ p")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 p after p.intro, got %d", len(results))
	}
}

// TestPseudoClassFirstChild tests :first-child
func TestPseudoClassFirstChild(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "li:first-child")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 li:first-child, got %d", len(results))
	}

	results, err = Match(body, "p:first-child")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 p:first-child, got %d", len(results))
	}
}

// TestPseudoClassLastChild tests :last-child
func TestPseudoClassLastChild(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "li:last-child")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 li:last-child, got %d", len(results))
	}

	results, err = Match(body, "p:last-child")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 p:last-child, got %d", len(results))
	}
}

// TestPseudoClassOnlyChild tests :only-child
func TestPseudoClassOnlyChild(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// ul is the only child of div#sidebar
	results, err := Match(body, "ul:only-child")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 ul:only-child, got %d", len(results))
	}
}

// TestPseudoClassNthChild tests :nth-child
func TestPseudoClassNthChild(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	tests := []struct {
		selector string
		expected int
	}{
		{"li:nth-child(1)", 1},
		{"li:nth-child(2)", 1},
		{"li:nth-child(odd)", 3},  // 1, 3, 5
		{"li:nth-child(even)", 2}, // 2, 4
		{"li:nth-child(2n)", 2},   // 2, 4
		{"li:nth-child(2n+1)", 3}, // 1, 3, 5
		{"li:nth-child(3n)", 1},   // 3
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

// TestPseudoClassNthLastChild tests :nth-last-child
func TestPseudoClassNthLastChild(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "li:nth-last-child(1)")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 li:nth-last-child(1), got %d", len(results))
	}

	results, err = Match(body, "li:nth-last-child(2)")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 li:nth-last-child(2), got %d", len(results))
	}
}

// TestPseudoClassFirstOfType tests :first-of-type
func TestPseudoClassFirstOfType(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "p:first-of-type")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 p:first-of-type, got %d", len(results))
	}

	results, err = Match(body, "div:first-of-type")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 div:first-of-type, got %d", len(results))
	}
}

// TestPseudoClassLastOfType tests :last-of-type
func TestPseudoClassLastOfType(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "p:last-of-type")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 p:last-of-type, got %d", len(results))
	}
}

// TestPseudoClassOnlyOfType tests :only-of-type
func TestPseudoClassOnlyOfType(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "ul:only-of-type")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 ul:only-of-type, got %d", len(results))
	}
}

// TestPseudoClassEmpty tests :empty
func TestPseudoClassEmpty(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "div:empty")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	// Both .empty div and data-div are empty (no children)
	if len(results) != 2 {
		t.Errorf("Expected 2 div:empty, got %d", len(results))
	}

	// Test with class to be specific
	results, err = Match(body, "div.empty:empty")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 div.empty:empty, got %d", len(results))
	}
}

// TestPseudoClassRoot tests :root
func TestPseudoClassRoot(t *testing.T) {
	doc := createTestDOM()
	html := doc.DocumentElement()

	// html:root should match
	results, err := Match(html, "html:root")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 html:root, got %d", len(results))
	}

	// body:root should not match
	results, err = Match(html, "body:root")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 body:root, got %d", len(results))
	}
}

// TestPseudoClassNot tests :not()
func TestPseudoClassNot(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// div:not(.empty) - divs that don't have class empty
	// main (has class container active), sidebar (has class container),
	// data-div (has data-value and data-lang, no class)
	results, err := Match(body, "div:not(.empty)")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	// There are 4 divs: main, sidebar, empty, data-div
	// div:not(.empty) should match: main, sidebar, data-div = 3
	// But if body is included in results, count would be higher
	// Let's just count total divs first
	allDivs, _ := Match(body, "div")
	emptyDivs, _ := Match(body, "div.empty")
	expected := len(allDivs) - len(emptyDivs)
	if len(results) != expected {
		t.Errorf("Expected %d div:not(.empty), got %d (total divs=%d, empty divs=%d)",
			expected, len(results), len(allDivs), len(emptyDivs))
	}

	// li:not(:first-child)
	results, err = Match(body, "li:not(:first-child)")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	// 5 li elements total, 1 is :first-child
	if len(results) != 4 {
		t.Errorf("Expected 4 li:not(:first-child), got %d", len(results))
	}
}

// TestSelectorList tests comma-separated selectors
func TestSelectorList(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// h1, h2, h3 - none exist
	results, err := Match(body, "h1, h2, h3")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 h1,h2,h3, got %d", len(results))
	}

	// p, span
	results, err = Match(body, "p, span")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 3 { // 2 p + 1 span
		t.Errorf("Expected 3 p,span, got %d", len(results))
	}

	// #main, #sidebar
	results, err = Match(body, "#main, #sidebar")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 #main,#sidebar, got %d", len(results))
	}
}

// TestComplexSelectors tests complex selector chains
func TestComplexSelectors(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// div#main > p.intro
	results, err := Match(body, "div#main > p.intro")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 div#main > p.intro, got %d", len(results))
	}

	// body > div.container p
	html := doc.DocumentElement()
	results, err = Match(html, "body > div.container p")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 body > div.container p, got %d", len(results))
	}
}

// TestMatchFirst tests MatchFirst function
func TestMatchFirst(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	result, err := MatchFirst(body, "div")
	if err != nil {
		t.Fatalf("MatchFirst error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected to find a div")
	}
	if result.ID() != testMainID {
		t.Errorf("Expected first div to be #main, got #%s", result.ID())
	}

	result, err = MatchFirst(body, ".nonexistent")
	if err != nil {
		t.Fatalf("MatchFirst error: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil for nonexistent selector, got %v", result)
	}
}

// TestElementQuery tests Element.Query method
func TestElementQuery(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := body.Query("div")
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 divs, got %d", len(results))
	}
}

// TestElementQueryFirst tests Element.QueryFirst method
func TestElementQueryFirst(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	result, err := body.QueryFirst("div")
	if err != nil {
		t.Fatalf("QueryFirst error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected to find a div")
	}
	if result.ID() != testMainID {
		t.Errorf("Expected first div to be #main, got #%s", result.ID())
	}
}

// TestParseNthExpression tests An+B formula parsing
func TestParseNthExpression(t *testing.T) {
	tests := []struct {
		expr string
		a    int
		b    int
		ok   bool
	}{
		{"odd", 2, 1, true},
		{"even", 2, 0, true},
		{"1", 0, 1, true},
		{"5", 0, 5, true},
		{"n", 1, 0, true},
		{"2n", 2, 0, true},
		{"-n", -1, 0, true},
		{"n+1", 1, 1, true},
		{"n-1", 1, -1, true},
		{"2n+1", 2, 1, true},
		{"2n-1", 2, -1, true},
		{"-n+3", -1, 3, true},
		{"3n+4", 3, 4, true},
		{"+n", 1, 0, true},
		{"-2n+3", -2, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			a, b, ok := parseNthExpression(tt.expr)
			if ok != tt.ok {
				t.Errorf("parseNthExpression(%q) ok = %v, want %v", tt.expr, ok, tt.ok)
				return
			}
			if ok && (a != tt.a || b != tt.b) {
				t.Errorf("parseNthExpression(%q) = (%d, %d), want (%d, %d)", tt.expr, a, b, tt.a, tt.b)
			}
		})
	}
}

// TestMatchesNth tests the nth matching formula
func TestMatchesNth(t *testing.T) {
	tests := []struct {
		index int
		a     int
		b     int
		want  bool
	}{
		// a=0, b=n means exact match
		{1, 0, 1, true},
		{2, 0, 1, false},
		{5, 0, 5, true},

		// a=2, b=0 means even (2, 4, 6, ...)
		{1, 2, 0, false},
		{2, 2, 0, true},
		{3, 2, 0, false},
		{4, 2, 0, true},

		// a=2, b=1 means odd (1, 3, 5, ...)
		{1, 2, 1, true},
		{2, 2, 1, false},
		{3, 2, 1, true},

		// a=3, b=1 means 1, 4, 7, 10, ...
		{1, 3, 1, true},
		{2, 3, 1, false},
		{4, 3, 1, true},
		{7, 3, 1, true},

		// negative a values
		{1, -1, 3, true},
		{2, -1, 3, true},
		{3, -1, 3, true},
		{4, -1, 3, false},
	}

	for _, tt := range tests {
		var name string
		if tt.a == 0 {
			name = string(rune('0' + tt.b)) // Simple number
		} else {
			name = string(rune('0'+tt.a)) + "n+" + string(rune('0'+tt.b))
		}
		t.Run(name, func(t *testing.T) {
			got := matchesNth(tt.index, tt.a, tt.b)
			if got != tt.want {
				t.Errorf("matchesNth(%d, %d, %d) = %v, want %v", tt.index, tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestSelectorString tests Selector.String method
func TestSelectorString(t *testing.T) {
	sel, err := Parse("div.container")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if sel.String() != "div.container" {
		t.Errorf("Expected selector string 'div.container', got %q", sel.String())
	}
}

// TestCaseInsensitiveTagMatching tests that tag matching is case-insensitive for HTML
func TestCaseInsensitiveTagMatching(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	// Should match regardless of case
	results, err := Match(body, "DIV")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 DIV matches, got %d", len(results))
	}

	results, err = Match(body, "Div")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 Div matches, got %d", len(results))
	}
}

// BenchmarkParse benchmarks selector parsing
func BenchmarkParse(b *testing.B) {
	selectors := []string{
		"div",
		"#main",
		".container",
		"div.container.active",
		"div > p.intro",
		"body > div.container p span",
		":nth-child(2n+1)",
		"[data-value=\"test\"]",
	}

	for _, sel := range selectors {
		b.Run(sel, func(b *testing.B) {
			for range b.N {
				_, _ = Parse(sel)
			}
		})
	}
}

// BenchmarkMatch benchmarks selector matching
func BenchmarkMatch(b *testing.B) {
	doc := createTestDOM()
	body := doc.Body()

	selectors := []string{
		"div",
		"#main",
		".container",
		"div.container",
		"div > p",
		"body div p",
		":first-child",
		":nth-child(odd)",
	}

	for _, sel := range selectors {
		b.Run(sel, func(b *testing.B) {
			for range b.N {
				_, _ = Match(body, sel)
			}
		})
	}
}
