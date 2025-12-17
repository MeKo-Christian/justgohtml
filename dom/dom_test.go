package dom

import "testing"

func TestDocumentInsertBeforeSetsParent(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")
	head := NewElement("head")
	body := NewElement("body")

	doc.AppendChild(html)
	html.AppendChild(body)
	html.InsertBefore(head, body)

	if head.Parent() != html {
		t.Fatalf("head.Parent() = %T, want html element", head.Parent())
	}
	if body.Parent() != html {
		t.Fatalf("body.Parent() = %T, want html element", body.Parent())
	}
	if doc.Parent() != nil {
		t.Fatalf("doc.Parent() = %T, want nil", doc.Parent())
	}
}

func TestDocumentFragmentAppendChildSetsParent(t *testing.T) {
	df := NewDocumentFragment()
	div := NewElement("div")
	df.AppendChild(div)
	if div.Parent() != df {
		t.Fatalf("div.Parent() = %T, want DocumentFragment", div.Parent())
	}
}
