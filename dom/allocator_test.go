package dom

import "testing"

func TestNodeAllocatorElements(t *testing.T) {
	alloc := NewNodeAllocator()

	el := alloc.NewElement("DiV")
	if el.TagName != "div" {
		t.Fatalf("TagName = %q, want %q", el.TagName, "div")
	}
	if el.Namespace != NamespaceHTML {
		t.Fatalf("Namespace = %q, want %q", el.Namespace, NamespaceHTML)
	}
	if el.Attributes == nil {
		t.Fatal("Attributes is nil")
	}
	if el.TemplateContent != nil {
		t.Fatalf("TemplateContent = %v, want nil", el.TemplateContent)
	}

	elNS := alloc.NewElementNS("foreignObject", NamespaceSVG)
	if elNS.TagName != "foreignObject" {
		t.Fatalf("TagName = %q, want %q", elNS.TagName, "foreignObject")
	}
	if elNS.Namespace != NamespaceSVG {
		t.Fatalf("Namespace = %q, want %q", elNS.Namespace, NamespaceSVG)
	}

	el.SetAttr("class", "one")
	elNS.SetAttr("class", "two")
	if el.Attr("class") == elNS.Attr("class") {
		t.Fatalf("attributes shared unexpectedly: %q", el.Attr("class"))
	}
}

func TestNodeAllocatorTextComment(t *testing.T) {
	alloc := NewNodeAllocator()

	txt := alloc.NewText("hello")
	if txt.Data != "hello" {
		t.Fatalf("Data = %q, want %q", txt.Data, "hello")
	}
	if txt.Parent() != nil {
		t.Fatal("Text parent should be nil")
	}

	comment := alloc.NewComment("note")
	if comment.Data != "note" {
		t.Fatalf("Data = %q, want %q", comment.Data, "note")
	}
	if comment.Parent() != nil {
		t.Fatal("Comment parent should be nil")
	}
}

func TestNodeAllocatorDocumentTypes(t *testing.T) {
	alloc := NewNodeAllocator()

	dt := alloc.NewDocumentType("html", "pub", "sys")
	if dt.Name != "html" || dt.PublicID != "pub" || dt.SystemID != "sys" {
		t.Fatalf("doctype fields mismatch: %+v", dt)
	}
	if dt.Parent() != nil {
		t.Fatal("DocumentType parent should be nil")
	}
}

func TestNodeAllocatorDocumentAndFragment(t *testing.T) {
	alloc := NewNodeAllocator()

	doc := alloc.NewDocument()
	el := alloc.NewElement("html")
	doc.AppendChild(el)
	if el.Parent() != doc {
		t.Fatal("element parent should be document")
	}

	frag := alloc.NewDocumentFragment()
	child := alloc.NewElement("span")
	frag.AppendChild(child)
	if child.Parent() != frag {
		t.Fatal("element parent should be fragment")
	}
}
