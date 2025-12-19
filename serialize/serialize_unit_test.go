package serialize

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

func TestToHTMLDocumentWithDoctypePretty(t *testing.T) {
	doc := dom.NewDocument()
	doc.Doctype = dom.NewDocumentType("html", "", "")

	html := dom.NewElement("html")
	doc.AppendChild(html)

	out := ToHTML(doc, Options{Pretty: true, IndentSize: 2})
	if out != "<!DOCTYPE html>\n<html></html>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestToHTMLTextEscaping(t *testing.T) {
	div := dom.NewElement("div")
	div.AppendChild(dom.NewText("a<b&c"))

	out := ToHTML(div, DefaultOptions())
	if out != "<div>a&lt;b&amp;c</div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestToHTMLAttributeEscaping(t *testing.T) {
	div := dom.NewElement("div")
	div.SetAttr("data-val", `a&"b`)

	out := ToHTML(div, DefaultOptions())
	if out != "<div data-val=\"a&amp;&quot;b\"></div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestToHTMLVoidElement(t *testing.T) {
	br := dom.NewElement("br")
	out := ToHTML(br, DefaultOptions())
	if out != "<br>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPrettyInlineChildren(t *testing.T) {
	div := dom.NewElement("div")
	div.AppendChild(dom.NewElement("span"))

	out := ToHTML(div, Options{Pretty: true, IndentSize: 2})
	if out != "<div><span></span></div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPrettyBlockIndent(t *testing.T) {
	div := dom.NewElement("div")
	div.AppendChild(dom.NewElement("p"))

	out := ToHTML(div, Options{Pretty: true, IndentSize: 2})
	if out != "<div>\n  <p></p>\n</div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPrettySkipsWhitespaceTextNodes(t *testing.T) {
	div := dom.NewElement("div")
	div.AppendChild(dom.NewText("\n  "))
	div.AppendChild(dom.NewElement("p"))
	div.AppendChild(dom.NewText("\n"))

	out := ToHTML(div, Options{Pretty: true, IndentSize: 2})
	if out != "<div>\n  <p></p>\n</div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestPrettyCommentInline(t *testing.T) {
	div := dom.NewElement("div")
	div.AppendChild(dom.NewComment("x"))

	out := ToHTML(div, Options{Pretty: true, IndentSize: 2})
	if out != "<div><!--x--></div>" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCollapseWhitespace(t *testing.T) {
	got := collapseWhitespace("  a   b  ")
	if got != " a b " {
		t.Fatalf("unexpected collapsed whitespace: %q", got)
	}
}

func TestIsWhitespaceOnly(t *testing.T) {
	if !isWhitespaceOnly(" \n\t\r") {
		t.Fatal("expected whitespace-only string to be true")
	}
	if isWhitespaceOnly(" a ") {
		t.Fatal("expected non-whitespace string to be false")
	}
}

func TestIsVoidAndBlockElements(t *testing.T) {
	if !isVoidElement("img") {
		t.Fatal("expected img to be void element")
	}
	if isVoidElement("div") {
		t.Fatal("expected div to not be void element")
	}
	if !isBlockElement("div") {
		t.Fatal("expected div to be block element")
	}
	if isBlockElement("span") {
		t.Fatal("expected span to not be block element")
	}
}
