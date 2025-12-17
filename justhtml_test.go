package JustGoHTML

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestParse_NotImplemented(t *testing.T) {
	doc, err := Parse("<html><body><p>Hello</p></body></html>")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if doc == nil || doc.DocumentElement() == nil || doc.DocumentElement().TagName != "html" {
		t.Fatalf("Parse returned invalid document: %#v", doc)
	}
}

func TestParseBytes_NotImplemented(t *testing.T) {
	doc, err := ParseBytes([]byte("<html><body><p>Hello</p></body></html>"))
	if err != nil {
		t.Fatalf("ParseBytes returned error: %v", err)
	}
	if doc == nil || doc.DocumentElement() == nil || doc.DocumentElement().TagName != "html" {
		t.Fatalf("ParseBytes returned invalid document: %#v", doc)
	}
}

func TestParseFragment_NotImplemented(t *testing.T) {
	nodes, err := ParseFragment("<td>Cell</td>", "tr")
	if err != nil {
		t.Fatalf("ParseFragment returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].TagName != "td" {
		t.Fatalf("ParseFragment nodes = %#v, want single <td>", nodes)
	}
}
