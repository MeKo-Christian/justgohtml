package treebuilder

import (
	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/constants"
)

// adoptionAgency implements the adoption agency algorithm for handling misnested
// formatting elements, per WHATWG HTML §13.2.5.2.5.
func (tb *TreeBuilder) adoptionAgency(subject string) {
	// 1. If the current node is the subject, and it is not in the active formatting elements list...
	if tb.currentElement() != nil && tb.currentElement().TagName == subject {
		if !tb.hasActiveFormattingEntry(subject) {
			tb.popUntil(subject)
			return
		}
	}

	// 2. Outer loop (at most 8 iterations).
	for outer := 0; outer < 8; outer++ {
		// 3. Find formatting element.
		formattingIndex, ok := tb.findActiveFormattingIndex(subject)
		if !ok {
			return
		}
		fmtEntry := tb.activeFormatting[formattingIndex]
		formattingElement := fmtEntry.node
		if formattingElement == nil {
			tb.removeFormattingEntry(formattingIndex)
			return
		}

		// 4. If formatting element is not in open elements, remove entry and abort.
		formattingInOpenIndex, ok := tb.indexOfOpenElement(formattingElement)
		if !ok {
			tb.removeFormattingEntry(formattingIndex)
			return
		}

		// 5. If formatting element is in open elements but not in scope, abort.
		if !tb.hasElementInScope(formattingElement.TagName, constants.DefaultScope) {
			return
		}

		// 7. Find furthest block: first special element after formatting element.
		var furthestBlock *dom.Element
		for i := formattingInOpenIndex + 1; i < len(tb.openElements); i++ {
			if isSpecialElement(tb.openElements[i]) {
				furthestBlock = tb.openElements[i]
				break
			}
		}

		if furthestBlock == nil {
			// Pop elements until formatting element has been popped.
			for len(tb.openElements) > 0 {
				popped := tb.popCurrent()
				if popped == formattingElement {
					break
				}
			}
			tb.removeFormattingEntry(formattingIndex)
			return
		}

		// 8. Bookmark.
		bookmark := formattingIndex + 1

		// 9. Node and last node.
		node := furthestBlock
		lastNode := furthestBlock

		// 10. Inner loop.
		innerCounter := 0
		for {
			innerCounter++

			// 10.1 Node = element above node.
			nodeIndex, ok := tb.indexOfOpenElement(node)
			if !ok || nodeIndex == 0 {
				return
			}
			node = tb.openElements[nodeIndex-1]

			// 10.2 If node is formatting element, break.
			if node == formattingElement {
				break
			}

			// 10.3 Find active formatting entry for node.
			nodeFormattingIndex, hasNodeFormatting := tb.findActiveFormattingIndexByNode(node)
			if innerCounter > 3 && hasNodeFormatting {
				tb.removeFormattingEntry(nodeFormattingIndex)
				if nodeFormattingIndex < bookmark {
					bookmark--
				}
				hasNodeFormatting = false
			}

			if !hasNodeFormatting {
				// Remove node from open elements.
				idx, ok := tb.indexOfOpenElement(node)
				if !ok {
					return
				}
				tb.removeOpenElementAt(idx)
				if idx < len(tb.openElements) {
					node = tb.openElements[idx]
				}
				continue
			}

			// 10.4 Replace entry with new element.
			entry := tb.activeFormatting[nodeFormattingIndex]
			newElement := dom.NewElement(entry.name)
			for k, v := range entry.attrs {
				newElement.SetAttr(k, v)
			}
			tb.activeFormatting[nodeFormattingIndex].node = newElement
			tb.openElements[tb.mustIndexOfOpenElement(node)] = newElement
			node = newElement

			// 10.5 If last node is furthest block, update bookmark.
			if lastNode == furthestBlock {
				bookmark = nodeFormattingIndex + 1
			}

			// 10.6 Reparent last_node.
			if p := lastNode.Parent(); p != nil {
				p.RemoveChild(lastNode)
			}
			node.AppendChild(lastNode)

			// 10.7
			lastNode = node
		}

		// 11. Insert last_node into common ancestor.
		commonAncestor := tb.openElements[formattingInOpenIndex-1]
		if p := lastNode.Parent(); p != nil {
			p.RemoveChild(lastNode)
		}
		if shouldFosterParent(commonAncestor) {
			tb.insertFosterNode(lastNode)
		} else {
			commonAncestor.AppendChild(lastNode)
		}

		// 12. Create new formatting element (clone of formatting element).
		entry := tb.activeFormatting[formattingIndex]
		newFormattingElement := dom.NewElement(entry.name)
		for k, v := range entry.attrs {
			newFormattingElement.SetAttr(k, v)
		}
		tb.activeFormatting[formattingIndex].node = newFormattingElement

		// 13. Move children of furthest block into new formatting element.
		for {
			children := furthestBlock.Children()
			if len(children) == 0 {
				break
			}
			child := children[0]
			furthestBlock.RemoveChild(child)
			newFormattingElement.AppendChild(child)
		}
		furthestBlock.AppendChild(newFormattingElement)

		// 14. Remove formatting entry and reinsert at bookmark.
		entryToMove := tb.activeFormatting[formattingIndex]
		tb.removeFormattingEntry(formattingIndex)
		bookmark--
		if bookmark < 0 {
			bookmark = 0
		}
		if bookmark > len(tb.activeFormatting) {
			bookmark = len(tb.activeFormatting)
		}
		tb.activeFormatting = append(tb.activeFormatting, formattingEntry{})
		copy(tb.activeFormatting[bookmark+1:], tb.activeFormatting[bookmark:])
		tb.activeFormatting[bookmark] = entryToMove

		// 15. Remove formatting element from open elements and insert new one after furthest block.
		if idx, ok := tb.indexOfOpenElement(formattingElement); ok {
			tb.removeOpenElementAt(idx)
		}
		furthestIdx := tb.mustIndexOfOpenElement(furthestBlock)
		tb.insertOpenElementAt(furthestIdx+1, newFormattingElement)
	}
}

func isSpecialElement(el *dom.Element) bool {
	if el == nil || el.Namespace != dom.NamespaceHTML {
		return false
	}
	return constants.SpecialElements[el.TagName]
}

func shouldFosterParent(commonAncestor *dom.Element) bool {
	if commonAncestor == nil {
		return false
	}
	switch commonAncestor.TagName {
	case "table", "tbody", "tfoot", "thead", "tr":
		return true
	default:
		return false
	}
}

func (tb *TreeBuilder) insertFosterNode(node dom.Node) {
	// Minimal foster parenting insertion location: insert before the last table
	// element on the stack, otherwise append to the current node.
	var tableEl *dom.Element
	for i := len(tb.openElements) - 1; i >= 0; i-- {
		if tb.openElements[i].TagName == "table" && tb.openElements[i].Namespace == dom.NamespaceHTML {
			tableEl = tb.openElements[i]
			break
		}
	}
	if tableEl == nil {
		tb.currentNode().AppendChild(node)
		return
	}
	parent := tableEl.Parent()
	if parent == nil {
		tb.document.AppendChild(node)
		return
	}
	parent.InsertBefore(node, tableEl)
}

func (tb *TreeBuilder) indexOfOpenElement(target *dom.Element) (int, bool) {
	for i, el := range tb.openElements {
		if el == target {
			return i, true
		}
	}
	return -1, false
}

func (tb *TreeBuilder) mustIndexOfOpenElement(target *dom.Element) int {
	idx, ok := tb.indexOfOpenElement(target)
	if !ok {
		panic("treebuilder: expected element on open element stack")
	}
	return idx
}

func (tb *TreeBuilder) removeOpenElementAt(index int) {
	if index < 0 || index >= len(tb.openElements) {
		return
	}
	copy(tb.openElements[index:], tb.openElements[index+1:])
	tb.openElements = tb.openElements[:len(tb.openElements)-1]
}

func (tb *TreeBuilder) insertOpenElementAt(index int, el *dom.Element) {
	if index < 0 {
		index = 0
	}
	if index > len(tb.openElements) {
		index = len(tb.openElements)
	}
	tb.openElements = append(tb.openElements, nil)
	copy(tb.openElements[index+1:], tb.openElements[index:])
	tb.openElements[index] = el
}
