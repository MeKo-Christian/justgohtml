package treebuilder

import "testing"

func TestFragmentContextFields(t *testing.T) {
	ctx := FragmentContext{
		TagName:   "div",
		Namespace: "html",
	}
	if ctx.TagName != "div" || ctx.Namespace != "html" {
		t.Fatalf("FragmentContext = %#v, want TagName=div Namespace=html", ctx)
	}
}
