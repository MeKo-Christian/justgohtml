package treebuilder

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

func TestHasElementInScope_IntegrationPointTerminates(t *testing.T) {
	tb := New(tokenizer.New(""))
	html := dom.NewElement("html")
	tb.document.AppendChild(html)
	tb.openElements = append(tb.openElements, html)

	foreignObject := dom.NewElementNS("foreignObject", dom.NamespaceSVG)
	html.AppendChild(foreignObject)
	tb.openElements = append(tb.openElements, foreignObject)

	if tb.hasElementInScope("html", constants.DefaultScope) {
		t.Fatalf("hasElementInScope(html) = true, want false (integration point terminates scope)")
	}
	if !tb.hasElementInTableScope("html") {
		t.Fatalf("hasElementInTableScope(html) = false, want true (table scope ignores integration points)")
	}
}

func TestHasElementInScope_TerminatorsStopSearch(t *testing.T) {
	tb := newTBWithStack(t, "html", "body", "table")
	if tb.hasElementInScope("body", constants.DefaultScope) {
		t.Fatalf("hasElementInScope(body) = true, want false (table terminates default scope)")
	}
}

func TestGenerateImpliedEndTags(t *testing.T) {
	tb := newTBWithStack(t, "html", "body", "p", "li", "dt")
	tb.generateImpliedEndTags("")
	if got := tb.currentElement(); got == nil || got.TagName != "body" {
		t.Fatalf("currentElement = %v, want body", got)
	}
	if len(tb.openElements) != 2 {
		t.Fatalf("openElements len = %d, want 2", len(tb.openElements))
	}

	tb = newTBWithStack(t, "html", "body", "p", "li", "dt")
	tb.generateImpliedEndTags("p")
	if got := tb.currentElement(); got == nil || got.TagName != "p" {
		t.Fatalf("currentElement = %v, want p", got)
	}
	if len(tb.openElements) != 3 {
		t.Fatalf("openElements len = %d, want 3", len(tb.openElements))
	}
}

func TestResetInsertionModeAppropriately(t *testing.T) {
	tb := newTBWithStack(t, "html", "body", "table", "tbody", "tr", "td")
	tb.mode = InBody
	tb.resetInsertionModeAppropriately()
	if tb.mode != InCell {
		t.Fatalf("mode = %v, want %v", tb.mode, InCell)
	}

	tb = newTBWithStack(t, "html", "body", "table", "colgroup")
	tb.mode = InBody
	tb.resetInsertionModeAppropriately()
	if tb.mode != InColumnGroup {
		t.Fatalf("mode = %v, want %v", tb.mode, InColumnGroup)
	}

	tb = newTBWithStack(t, "html", "body", "template")
	tb.mode = InBody
	tb.resetInsertionModeAppropriately()
	if tb.mode != InTemplate {
		t.Fatalf("mode = %v, want %v", tb.mode, InTemplate)
	}
}

func TestActiveFormattingMarkers(t *testing.T) {
	tb := New(tokenizer.New(""))
	tb.activeFormatting = []formattingEntry{
		{name: "a"},
		{marker: true},
		{name: "b"},
	}

	tb.clearActiveFormattingElements()
	if len(tb.activeFormatting) != 1 || tb.activeFormatting[0].name != "a" {
		t.Fatalf("activeFormatting = %#v, want only entry a", tb.activeFormatting)
	}

	tb.pushActiveFormattingMarker()
	if len(tb.activeFormatting) != 2 || !tb.activeFormatting[1].marker {
		t.Fatalf("activeFormatting = %#v, want trailing marker", tb.activeFormatting)
	}
}

func TestQuirksMode_NoDoctypeSetsQuirks(t *testing.T) {
	tb := New(tokenizer.New(""))
	tb.ProcessToken(tokenizer.Token{Type: tokenizer.StartTag, Name: "html"})
	if tb.document.QuirksMode != dom.Quirks {
		t.Fatalf("QuirksMode = %v, want %v", tb.document.QuirksMode, dom.Quirks)
	}
}

func TestQuirksMode_DoctypeRules(t *testing.T) {
	tb := New(tokenizer.New(""))
	tb.ProcessToken(tokenizer.Token{Type: tokenizer.DOCTYPE, Name: "html"})
	if tb.document.QuirksMode != dom.NoQuirks {
		t.Fatalf("QuirksMode = %v, want %v", tb.document.QuirksMode, dom.NoQuirks)
	}

	publicID := "-//W3C//DTD XHTML 1.0 Transitional//"
	tb = New(tokenizer.New(""))
	tb.ProcessToken(tokenizer.Token{Type: tokenizer.DOCTYPE, Name: "html", PublicID: &publicID})
	if tb.document.QuirksMode != dom.LimitedQuirks {
		t.Fatalf("QuirksMode = %v, want %v", tb.document.QuirksMode, dom.LimitedQuirks)
	}

	publicID = "-//W3C//DTD HTML 4.01 Transitional//"
	tb = New(tokenizer.New(""))
	tb.ProcessToken(tokenizer.Token{Type: tokenizer.DOCTYPE, Name: "html", PublicID: &publicID})
	if tb.document.QuirksMode != dom.Quirks {
		t.Fatalf("QuirksMode = %v, want %v", tb.document.QuirksMode, dom.Quirks)
	}

	systemID := "http://example.com/strict.dtd"
	tb = New(tokenizer.New(""))
	tb.ProcessToken(tokenizer.Token{Type: tokenizer.DOCTYPE, Name: "html", PublicID: &publicID, SystemID: &systemID})
	if tb.document.QuirksMode != dom.LimitedQuirks {
		t.Fatalf("QuirksMode = %v, want %v", tb.document.QuirksMode, dom.LimitedQuirks)
	}

	tb = New(tokenizer.New(""))
	tb.ProcessToken(tokenizer.Token{Type: tokenizer.DOCTYPE, Name: "nothtml"})
	if tb.document.QuirksMode != dom.Quirks {
		t.Fatalf("QuirksMode = %v, want %v", tb.document.QuirksMode, dom.Quirks)
	}
}
