package treebuilder

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

func newTBWithStack(t *testing.T, tagNames ...string) *TreeBuilder {
	t.Helper()
	tb := New(tokenizer.New(""))
	var parent dom.Node = tb.document
	for _, name := range tagNames {
		el := dom.NewElement(name)
		parent.AppendChild(el)
		tb.openElements = append(tb.openElements, el)
		parent = el
		if name == "head" {
			tb.headElement = el
		}
	}
	return tb
}

func TestInBody_TableSwitchesMode(t *testing.T) {
	tb := newTBWithStack(t, "html", "body")
	tb.mode = InBody

	tb.processInBody(tokenizer.Token{Type: tokenizer.StartTag, Name: "table"})

	if tb.mode != InTable {
		t.Fatalf("mode = %v, want %v", tb.mode, InTable)
	}
	if got := tb.currentElement(); got == nil || got.TagName != "table" {
		t.Fatalf("current element = %v, want table", got)
	}
}

func TestInTable_CharacterSwitchesToTableText(t *testing.T) {
	tb := newTBWithStack(t, "html", "body", "table")
	tb.mode = InTable

	reprocess := tb.processInTable(tokenizer.Token{Type: tokenizer.Character, Data: "X"})

	if !reprocess {
		t.Fatalf("reprocess = false, want true")
	}
	if tb.mode != InTableText {
		t.Fatalf("mode = %v, want %v", tb.mode, InTableText)
	}
	if tb.tableTextOriginalMode == nil || *tb.tableTextOriginalMode != InTable {
		t.Fatalf("tableTextOriginalMode = %v, want %v", tb.tableTextOriginalMode, InTable)
	}
}

func TestInTableText_FosterParentingInsertsBeforeTable(t *testing.T) {
	tb := newTBWithStack(t, "html", "body", "table")
	tb.mode = InTableText
	orig := InTable
	tb.tableTextOriginalMode = &orig
	tb.pendingTableText = []string{"X"}

	reprocess := tb.processInTableText(tokenizer.Token{Type: tokenizer.EndTag, Name: "table"})
	if !reprocess {
		t.Fatalf("reprocess = false, want true")
	}
	if tb.mode != InTable {
		t.Fatalf("mode = %v, want %v", tb.mode, InTable)
	}

	body := tb.document.Body()
	if body == nil {
		t.Fatalf("missing body")
	}
	children := body.Children()
	if len(children) != 2 {
		t.Fatalf("body children = %d, want 2", len(children))
	}
	if txt, ok := children[0].(*dom.Text); !ok || txt.Data != "X" {
		t.Fatalf("first child = %#v, want Text(\"X\")", children[0])
	}
	if el, ok := children[1].(*dom.Element); !ok || el.TagName != "table" {
		t.Fatalf("second child = %#v, want <table>", children[1])
	}
}

func TestNewFragment_SelectsContextInsertionMode(t *testing.T) {
	tb := NewFragment(tokenizer.New(""), &FragmentContext{TagName: "tr", Namespace: "html"})
	if tb.mode != InRow {
		t.Fatalf("mode = %v, want %v", tb.mode, InRow)
	}
	if tb.originalMode != InRow {
		t.Fatalf("originalMode = %v, want %v", tb.originalMode, InRow)
	}
}

func TestAfterBody_CommentAttachesToHTML(t *testing.T) {
	tb := newTBWithStack(t, "html")
	tb.mode = AfterBody

	tb.processAfterBody(tokenizer.Token{Type: tokenizer.Comment, Data: "hi"})

	html := tb.document.DocumentElement()
	if html == nil {
		t.Fatalf("missing html element")
	}
	children := html.Children()
	if len(children) != 1 {
		t.Fatalf("html children = %d, want 1", len(children))
	}
	if c, ok := children[0].(*dom.Comment); !ok || c.Data != "hi" {
		t.Fatalf("child = %#v, want Comment(\"hi\")", children[0])
	}
}
