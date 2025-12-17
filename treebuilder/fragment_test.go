package treebuilder

import (
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

func TestNewFragment_ContextElementNamespace(t *testing.T) {
	tok := tokenizer.New("")

	tbSVG := NewFragment(tok, &FragmentContext{TagName: "foreignObject", Namespace: "svg"})
	if tbSVG.fragmentElement == nil {
		t.Fatal("missing fragment context element")
	}
	if tbSVG.fragmentElement.Namespace != dom.NamespaceSVG {
		t.Fatalf("svg context element namespace = %q, want %q", tbSVG.fragmentElement.Namespace, dom.NamespaceSVG)
	}
	if tbSVG.fragmentElement.TagName != "foreignObject" {
		t.Fatalf("svg context element tag = %q, want %q", tbSVG.fragmentElement.TagName, "foreignObject")
	}

	tbMath := NewFragment(tok, &FragmentContext{TagName: "mi", Namespace: "mathml"})
	if tbMath.fragmentElement == nil {
		t.Fatal("missing fragment context element")
	}
	if tbMath.fragmentElement.Namespace != dom.NamespaceMathML {
		t.Fatalf("mathml context element namespace = %q, want %q", tbMath.fragmentElement.Namespace, dom.NamespaceMathML)
	}
	if tbMath.fragmentElement.TagName != "mi" {
		t.Fatalf("mathml context element tag = %q, want %q", tbMath.fragmentElement.TagName, "mi")
	}
}
