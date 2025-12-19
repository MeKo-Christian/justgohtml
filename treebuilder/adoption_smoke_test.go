package treebuilder_test

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
)

func TestAdoptionAgency_A_P_Misnesting(t *testing.T) {
	doc, err := JustGoHTML.Parse("<a><p></a></p>")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <a>
|     <p>
|       <a>`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestAdoptionAgency_NestedAnchors(t *testing.T) {
	doc, err := JustGoHTML.Parse("<a><p>X<a>Y</a>Z</p></a>")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <a>
|     <p>
|       <a>
|         "X"
|       <a>
|         "Y"
|       "Z"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestAdoptionAgency_A_B_Misnesting(t *testing.T) {
	doc, err := JustGoHTML.Parse("<a>1<b>2</a>3</b>")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <a>
|       "1"
|       <b>
|         "2"
|     <b>
|       "3"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestAdoptionAgency_TableFormatting(t *testing.T) {
	doc, err := JustGoHTML.Parse("<a><table><td><a><table></table><a></tr><a></table><b>X</b>C<a>Y")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <a>
|       <a>
|       <table>
|         <tbody>
|           <tr>
|             <td>
|               <a>
|                 <table>
|               <a>
|     <a>
|       <b>
|         "X"
|       "C"
|     <a>
|       "Y"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}

func TestReconstructActiveFormattingElements_ReopensAfterPClosed(t *testing.T) {
	doc, err := JustGoHTML.Parse("<p><b>1</p>2")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := testutil.SerializeHTML5LibTree(doc)
	want := `| <html>
|   <head>
|   <body>
|     <p>
|       <b>
|         "1"
|     <b>
|       "2"`
	if got != want {
		t.Fatalf("tree mismatch\ngot:\n%s\n\nwant:\n%s", got, want)
	}
}
