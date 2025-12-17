package treebuilder

import "testing"

func TestInsertionModeString(t *testing.T) {
	tests := []struct {
		mode InsertionMode
		want string
	}{
		{Initial, "initial"},
		{BeforeHTML, "before html"},
		{BeforeHead, "before head"},
		{InHead, "in head"},
		{InHeadNoscript, "in head noscript"},
		{AfterHead, "after head"},
		{InBody, "in body"},
		{Text, "text"},
		{InTable, "in table"},
		{InTableText, "in table text"},
		{InCaption, "in caption"},
		{InColumnGroup, "in column group"},
		{InTableBody, "in table body"},
		{InRow, "in row"},
		{InCell, "in cell"},
		{InSelect, "in select"},
		{InSelectInTable, "in select in table"},
		{InTemplate, "in template"},
		{AfterBody, "after body"},
		{InFrameset, "in frameset"},
		{AfterFrameset, "after frameset"},
		{AfterAfterBody, "after after body"},
		{AfterAfterFrameset, "after after frameset"},
		{InsertionMode(-1), "unknown"},
		{InsertionMode(123), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Fatalf("InsertionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}
