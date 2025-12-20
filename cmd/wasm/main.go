//go:build js && wasm

// Package main provides WebAssembly bindings for JustGoHTML.
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/dom"
	_ "github.com/MeKo-Christian/JustGoHTML/selector" // Register selector functions with dom
	"github.com/MeKo-Christian/JustGoHTML/serialize"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

func main() {
	// Register functions
	js.Global().Set("JustGoHTML", js.ValueOf(map[string]any{
		"parse":         js.FuncOf(parse),
		"parseFragment": js.FuncOf(parseFragment),
		"tokenize":      js.FuncOf(tokenize),
		"query":         js.FuncOf(query),
		"version":       js.ValueOf(JustGoHTML.Version),
	}))

	// Keep the program running
	select {}
}

// parse parses HTML and returns a serialized result.
// Arguments: html (string), options (object)
// Options: { format: "html"|"text"|"tree", selector: string, pretty: bool }
func parse(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errorResult("parse requires an HTML string argument")
	}

	html := args[0].String()

	// Parse options
	opts := parseOptions{}
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		opts = getParseOptions(args[1])
	}

	// Parse HTML
	doc, err := JustGoHTML.Parse(html)
	if err != nil {
		return errorResult("parse error: " + err.Error())
	}

	return formatOutput(doc, opts)
}

// parseFragment parses an HTML fragment in a context.
// Arguments: html (string), context (string), options (object)
func parseFragment(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errorResult("parseFragment requires html and context arguments")
	}

	html := args[0].String()
	context := args[1].String()

	opts := parseOptions{}
	if len(args) > 2 && !args[2].IsUndefined() && !args[2].IsNull() {
		opts = getParseOptions(args[2])
	}

	nodes, err := JustGoHTML.ParseFragment(html, context)
	if err != nil {
		return errorResult("parse error: " + err.Error())
	}

	return formatFragmentOutput(nodes, opts)
}

// tokenize tokenizes HTML and returns tokens as an array.
// Arguments: html (string)
// Returns: array of token objects with type, data, and other properties
func tokenize(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errorResult("tokenize requires an HTML string argument")
	}

	html := args[0].String()
	tok := tokenizer.New(html)

	var tokens []map[string]any
	for {
		tt := tok.Next()
		token := tokenToJS(&tt)
		tokens = append(tokens, token)
		if tt.Type == tokenizer.EOF {
			break
		}
	}

	// Convert to JSON and back to JS value
	data, err := json.Marshal(map[string]any{
		"success": true,
		"tokens":  tokens,
		"errors":  errorsToJS(tok.Errors()),
	})
	if err != nil {
		return errorResult("JSON encoding error: " + err.Error())
	}

	return js.Global().Get("JSON").Call("parse", string(data))
}

// query parses HTML and runs a CSS selector query.
// Arguments: html (string), selector (string), options (object)
// Options: { format: "html"|"text", pretty: bool }
// Returns: array of matching elements serialized according to format
func query(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errorResult("query requires html and selector arguments")
	}

	html := args[0].String()
	selectorStr := args[1].String()

	if selectorStr == "" {
		return errorResult("selector cannot be empty")
	}

	// Parse options
	opts := parseOptions{Format: "html", Pretty: true}
	if len(args) > 2 && !args[2].IsUndefined() && !args[2].IsNull() {
		opts = getParseOptions(args[2])
	}

	// Parse HTML
	doc, err := JustGoHTML.Parse(html)
	if err != nil {
		return errorResult("parse error: " + err.Error())
	}

	// Run selector query
	matches, err := doc.Query(selectorStr)
	if err != nil {
		return errorResult("selector error: " + err.Error())
	}

	// Format results
	var results []map[string]any
	for i, elem := range matches {
		var serialized string
		switch opts.Format {
		case "text":
			serialized = extractElementText(elem)
		default:
			serialized = serialize.ToHTML(elem, serialize.Options{
				Pretty:     opts.Pretty,
				IndentSize: 2,
			})
		}

		results = append(results, map[string]any{
			"index":   i,
			"tagName": elem.TagName,
			"html":    serialized,
			"tree":    nodeToTree(elem),
		})
	}

	data, err := json.Marshal(map[string]any{
		"success": true,
		"count":   len(matches),
		"matches": results,
	})
	if err != nil {
		return errorResult("JSON encoding error: " + err.Error())
	}

	return js.Global().Get("JSON").Call("parse", string(data))
}

type parseOptions struct {
	Format   string
	Selector string
	Pretty   bool
}

func getParseOptions(v js.Value) parseOptions {
	opts := parseOptions{
		Format: "html",
		Pretty: false,
	}

	if format := v.Get("format"); !format.IsUndefined() {
		opts.Format = format.String()
	}
	if selector := v.Get("selector"); !selector.IsUndefined() {
		opts.Selector = selector.String()
	}
	if pretty := v.Get("pretty"); !pretty.IsUndefined() {
		opts.Pretty = pretty.Bool()
	}

	return opts
}

func formatOutput(doc *dom.Document, opts parseOptions) any {
	var result string

	switch opts.Format {
	case "html":
		result = serialize.ToHTML(doc, serialize.Options{
			Pretty:     opts.Pretty,
			IndentSize: 2,
		})
	case "text":
		result = extractText(doc)
	case "tree":
		return treeToJS(doc)
	default:
		result = serialize.ToHTML(doc, serialize.DefaultOptions())
	}

	data, err := json.Marshal(map[string]any{
		"success": true,
		"result":  result,
	})
	if err != nil {
		return errorResult("JSON encoding error: " + err.Error())
	}

	return js.Global().Get("JSON").Call("parse", string(data))
}

func formatFragmentOutput(nodes []*dom.Element, opts parseOptions) any {
	var results []string
	for _, node := range nodes {
		switch opts.Format {
		case "html":
			results = append(results, serialize.ToHTML(node, serialize.Options{
				Pretty:     opts.Pretty,
				IndentSize: 2,
			}))
		case "text":
			results = append(results, extractElementText(node))
		default:
			results = append(results, serialize.ToHTML(node, serialize.DefaultOptions()))
		}
	}

	data, err := json.Marshal(map[string]any{
		"success": true,
		"results": results,
	})
	if err != nil {
		return errorResult("JSON encoding error: " + err.Error())
	}

	return js.Global().Get("JSON").Call("parse", string(data))
}

func errorResult(msg string) any {
	data, _ := json.Marshal(map[string]any{
		"success": false,
		"error":   msg,
	})
	return js.Global().Get("JSON").Call("parse", string(data))
}

func tokenToJS(t *tokenizer.Token) map[string]any {
	result := map[string]any{
		"type": t.Type.String(),
	}

	switch t.Type {
	case tokenizer.DOCTYPE:
		result["name"] = t.Name
		if t.PublicID != nil {
			result["publicId"] = *t.PublicID
		}
		if t.SystemID != nil {
			result["systemId"] = *t.SystemID
		}
		result["forceQuirks"] = t.ForceQuirks
	case tokenizer.StartTag, tokenizer.EndTag:
		result["name"] = t.Name
		result["selfClosing"] = t.SelfClosing
		if len(t.Attrs) > 0 {
			result["attributes"] = tokenizer.AttrsToMap(t.Attrs)
		}
	case tokenizer.Comment:
		result["data"] = t.Data
	case tokenizer.Character:
		result["data"] = t.Data
	}

	return result
}

func errorsToJS(errs []tokenizer.ParseError) []map[string]any {
	if len(errs) == 0 {
		return nil
	}
	result := make([]map[string]any, len(errs))
	for i, e := range errs {
		result[i] = map[string]any{
			"code":   e.Code,
			"line":   e.Line,
			"column": e.Column,
		}
	}
	return result
}

func extractText(doc *dom.Document) string {
	var sb []byte
	for _, child := range doc.Children() {
		extractNodeText(&sb, child)
	}
	return string(sb)
}

func extractElementText(elem *dom.Element) string {
	var sb []byte
	for _, child := range elem.Children() {
		extractNodeText(&sb, child)
	}
	return string(sb)
}

func extractNodeText(sb *[]byte, node dom.Node) {
	switch n := node.(type) {
	case *dom.Text:
		*sb = append(*sb, n.Data...)
	case *dom.Element:
		for _, child := range n.Children() {
			extractNodeText(sb, child)
		}
	case *dom.Document:
		for _, child := range n.Children() {
			extractNodeText(sb, child)
		}
	}
}

func treeToJS(doc *dom.Document) any {
	tree := nodeToTree(doc)
	data, err := json.Marshal(map[string]any{
		"success": true,
		"tree":    tree,
	})
	if err != nil {
		return errorResult("JSON encoding error: " + err.Error())
	}
	return js.Global().Get("JSON").Call("parse", string(data))
}

func nodeToTree(node dom.Node) map[string]any {
	switch n := node.(type) {
	case *dom.Document:
		children := make([]map[string]any, 0)
		for _, child := range n.Children() {
			children = append(children, nodeToTree(child))
		}
		return map[string]any{
			"type":     "document",
			"children": children,
		}
	case *dom.DocumentType:
		return map[string]any{
			"type":     "doctype",
			"name":     n.Name,
			"publicId": n.PublicID,
			"systemId": n.SystemID,
		}
	case *dom.Element:
		children := make([]map[string]any, 0)
		for _, child := range n.Children() {
			children = append(children, nodeToTree(child))
		}
		attrs := make(map[string]string)
		for _, attr := range n.Attributes.All() {
			attrs[attr.Name] = attr.Value
		}
		return map[string]any{
			"type":       "element",
			"tagName":    n.TagName,
			"namespace":  n.Namespace,
			"attributes": attrs,
			"children":   children,
		}
	case *dom.Text:
		return map[string]any{
			"type": "text",
			"data": n.Data,
		}
	case *dom.Comment:
		return map[string]any{
			"type": "comment",
			"data": n.Data,
		}
	default:
		return map[string]any{
			"type": "unknown",
		}
	}
}
