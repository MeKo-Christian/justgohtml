package treebuilder_test

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
)

func TestTreeBuilder_Smoke_Comments01(t *testing.T) {
	doc, err := JustGoHTML.Parse("FOO<!-- BAR -->BAZ")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     "FOO"
|     <!--  BAR  -->
|     "BAZ"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestTreeBuilder_Smoke_Entities02AttrDecoding(t *testing.T) {
	doc, err := JustGoHTML.Parse(`<div bar="ZZ&gt;YY"></div>`)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <div>
|       bar="ZZ>YY"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}
