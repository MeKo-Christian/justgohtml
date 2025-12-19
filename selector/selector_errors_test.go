package selector

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

func TestParseErrors(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"div@foo",
		"[attr",
		"[]",
		"[attr!=value]",
		"[attr=]",
		`[attr="unclosed]`,
		"div:",
		"div >",
		"div > > p",
		"#",
		".",
		"div,",
	}

	for _, sel := range tests {
		t.Run(sel, func(t *testing.T) {
			if _, err := Parse(sel); err == nil {
				t.Fatalf("Parse(%q) expected error", sel)
			}
		})
	}
}

func TestMatchAttributeCaseInsensitive(t *testing.T) {
	doc := dom.NewDocument()
	root := dom.NewElement("html")
	doc.AppendChild(root)

	elem := dom.NewElement("div")
	elem.Attributes.SetNS("", "DATA-ID", "123")
	root.AppendChild(elem)

	results, err := Match(root, "[data-id]")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestMatchAttributeEmptyValueOperators(t *testing.T) {
	doc := dom.NewDocument()
	root := dom.NewElement("html")
	doc.AppendChild(root)

	elem := dom.NewElement("div")
	elem.SetAttr("data-x", "")
	root.AppendChild(elem)

	tests := []string{
		`div[data-x^=""]`,
		`div[data-x$=""]`,
		`div[data-x*=""]`,
	}

	for _, sel := range tests {
		t.Run(sel, func(t *testing.T) {
			results, err := Match(root, sel)
			if err != nil {
				t.Fatalf("Match error: %v", err)
			}
			if len(results) != 0 {
				t.Fatalf("expected 0 matches for %q, got %d", sel, len(results))
			}
		})
	}
}

func TestMatchAttributeMissingOnElement(t *testing.T) {
	doc := dom.NewDocument()
	root := dom.NewElement("html")
	doc.AppendChild(root)

	elem := dom.NewElement("div")
	root.AppendChild(elem)

	results, err := Match(root, "[id]")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}

func TestNotInvalidInnerSelector(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "div:not(div >)")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}

func TestNotEmptyArgMatchesAll(t *testing.T) {
	doc := createTestDOM()
	body := doc.Body()

	results, err := Match(body, "div:not()")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected matches for div:not()")
	}
}

func TestPseudoClassesNoParent(t *testing.T) {
	elem := dom.NewElement("div")

	results, err := Match(elem, ":first-child")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected element to match :first-child, got %d", len(results))
	}

	results, err = Match(elem, ":root")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected element to not match :root, got %d", len(results))
	}
}

func TestNthChildInvalidExpression(t *testing.T) {
	doc := dom.NewDocument()
	root := dom.NewElement("html")
	doc.AppendChild(root)
	ul := dom.NewElement("ul")
	root.AppendChild(ul)
	for i := 0; i < 2; i++ {
		ul.AppendChild(dom.NewElement("li"))
	}

	results, err := Match(root, "li:nth-child(xn+1)")
	if err != nil {
		t.Fatalf("Match error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(results))
	}
}
