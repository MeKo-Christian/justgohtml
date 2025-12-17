package treebuilder

import (
	"sort"
	"strings"

	"github.com/MeKo-Christian/JustGoHTML/dom"
)

type formattingEntry struct {
	marker    bool
	name      string
	attrs     map[string]string
	node      *dom.Element
	signature string
}

func (tb *TreeBuilder) pushFormattingMarker() {
	tb.activeFormatting = append(tb.activeFormatting, formattingEntry{marker: true})
}

func (tb *TreeBuilder) clearActiveFormattingUpToMarker() {
	for len(tb.activeFormatting) > 0 {
		last := tb.activeFormatting[len(tb.activeFormatting)-1]
		tb.activeFormatting = tb.activeFormatting[:len(tb.activeFormatting)-1]
		if last.marker {
			return
		}
	}
}

func (tb *TreeBuilder) appendActiveFormattingEntry(name string, attrs map[string]string, node *dom.Element) {
	entryAttrs := cloneAttrs(attrs)
	tb.activeFormatting = append(tb.activeFormatting, formattingEntry{
		name:      name,
		attrs:     entryAttrs,
		node:      node,
		signature: attrsSignature(entryAttrs),
	})
}

func (tb *TreeBuilder) findActiveFormattingIndex(name string) (int, bool) {
	for i := len(tb.activeFormatting) - 1; i >= 0; i-- {
		entry := tb.activeFormatting[i]
		if entry.marker {
			break
		}
		if entry.name == name {
			return i, true
		}
	}
	return -1, false
}

func (tb *TreeBuilder) findActiveFormattingIndexByNode(node *dom.Element) (int, bool) {
	for i := len(tb.activeFormatting) - 1; i >= 0; i-- {
		entry := tb.activeFormatting[i]
		if !entry.marker && entry.node == node {
			return i, true
		}
	}
	return -1, false
}

func (tb *TreeBuilder) findActiveFormattingDuplicate(name string, attrs map[string]string) (int, bool) {
	sig := attrsSignature(attrs)
	var matches []int
	for i, entry := range tb.activeFormatting {
		if entry.marker {
			matches = matches[:0]
			continue
		}
		if entry.name == name && entry.signature == sig {
			matches = append(matches, i)
		}
	}
	if len(matches) >= 3 {
		return matches[0], true
	}
	return -1, false
}

func (tb *TreeBuilder) hasActiveFormattingEntry(name string) bool {
	_, ok := tb.findActiveFormattingIndex(name)
	return ok
}

func (tb *TreeBuilder) removeFormattingEntry(index int) {
	if index < 0 || index >= len(tb.activeFormatting) {
		return
	}
	copy(tb.activeFormatting[index:], tb.activeFormatting[index+1:])
	tb.activeFormatting = tb.activeFormatting[:len(tb.activeFormatting)-1]
}

func (tb *TreeBuilder) removeLastActiveFormattingByName(name string) {
	for i := len(tb.activeFormatting) - 1; i >= 0; i-- {
		entry := tb.activeFormatting[i]
		if entry.marker {
			break
		}
		if entry.name == name {
			tb.removeFormattingEntry(i)
			return
		}
	}
}

func (tb *TreeBuilder) removeLastOpenElementByName(name string) {
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		if tb.openElements[i].TagName == name {
			copy(tb.openElements[i:], tb.openElements[i+1:])
			tb.openElements = tb.openElements[:len(tb.openElements)-1]
			return
		}
	}
}

func (tb *TreeBuilder) reconstructActiveFormattingElements() {
	// Per WHATWG HTML §13.2.5.2.1 (reconstruct the active formatting elements).
	if len(tb.activeFormatting) == 0 {
		return
	}
	last := tb.activeFormatting[len(tb.activeFormatting)-1]
	if last.marker || tb.elementInOpenElements(last.node) {
		return
	}

	index := len(tb.activeFormatting) - 1
	for {
		index--
		if index < 0 {
			index = 0
			break
		}
		entry := tb.activeFormatting[index]
		if entry.marker || tb.elementInOpenElements(entry.node) {
			index++
			break
		}
	}

	for index < len(tb.activeFormatting) {
		entry := tb.activeFormatting[index]
		el := tb.insertElement(entry.name, cloneAttrs(entry.attrs))
		tb.activeFormatting[index].node = el
		index++
	}
}

func (tb *TreeBuilder) elementInOpenElements(node *dom.Element) bool {
	for _, el := range tb.openElements {
		if el == node {
			return true
		}
	}
	return false
}

func cloneAttrs(attrs map[string]string) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]string, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}

func attrsSignature(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(attrs[k])
		sb.WriteByte(0)
	}
	return sb.String()
}
