package treebuilder_test

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
)

func TestForeignContent_SVGTagAndAttrAdjustment(t *testing.T) {
	doc, err := JustGoHTML.Parse(`<svg viewbox="0 0 1 1"></svg>`)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <svg svg>
|       viewBox="0 0 1 1"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestForeignContent_SVGTagNameCaseAdjustment(t *testing.T) {
	doc, err := JustGoHTML.Parse(`<svg><lineargradient></lineargradient></svg>`)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <svg svg>
|       <svg linearGradient>`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestForeignContent_HTMLIntegrationPoint_ForeignObject(t *testing.T) {
	doc, err := JustGoHTML.Parse(`<svg><foreignObject><p>Hi</p></foreignObject></svg>`)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <svg svg>
|       <svg foreignObject>
|         <p>
|           "Hi"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}
