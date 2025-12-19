package treebuilder

import "github.com/MeKo-Christian/JustGoHTML/dom"

func (tb *TreeBuilder) populateSelectedContent(root dom.Node) {
	selects := []*dom.Element{}
	findElements(root, "select", &selects)

	for _, sel := range selects {
		selectedcontent := findElement(sel, "selectedcontent")
		if selectedcontent == nil {
			continue
		}

		options := []*dom.Element{}
		findElements(sel, "option", &options)
		if len(options) == 0 {
			continue
		}

		var selected *dom.Element
		for _, opt := range options {
			if opt.Namespace == dom.NamespaceHTML && opt.HasAttr("selected") {
				selected = opt
				break
			}
		}
		if selected == nil {
			selected = options[0]
		}

		cloneChildren(selected, selectedcontent)
	}
}

func findElements(node dom.Node, name string, out *[]*dom.Element) {
	if el, ok := node.(*dom.Element); ok {
		if el.Namespace == dom.NamespaceHTML && el.TagName == name {
			*out = append(*out, el)
		}
		if el.TemplateContent != nil {
			for _, child := range el.TemplateContent.Children() {
				findElements(child, name, out)
			}
		}
	}
	for _, child := range node.Children() {
		findElements(child, name, out)
	}
}

func findElement(node dom.Node, name string) *dom.Element {
	if el, ok := node.(*dom.Element); ok {
		if el.Namespace == dom.NamespaceHTML && el.TagName == name {
			return el
		}
		if el.TemplateContent != nil {
			for _, child := range el.TemplateContent.Children() {
				if found := findElement(child, name); found != nil {
					return found
				}
			}
		}
	}
	for _, child := range node.Children() {
		if found := findElement(child, name); found != nil {
			return found
		}
	}
	return nil
}

func cloneChildren(source, target *dom.Element) {
	for _, child := range append([]dom.Node(nil), target.Children()...) {
		target.RemoveChild(child)
	}
	for _, child := range source.Children() {
		target.AppendChild(child.Clone(true))
	}
}
