# JustGoHTML

[![Go Reference](https://pkg.go.dev/badge/github.com/MeKo-Christian/JustGoHTML.svg)](https://pkg.go.dev/github.com/MeKo-Christian/JustGoHTML) [![Go Report Card](https://goreportcard.com/badge/github.com/MeKo-Christian/JustGoHTML)](https://goreportcard.com/report/github.com/MeKo-Christian/JustGoHTML) [![Go](https://github.com/MeKo-Christian/JustGoHTML/workflows/Go/badge.svg)](https://github.com/MeKo-Christian/JustGoHTML/actions)

A pure Go HTML5 parser that just works. No CGO. No dependencies. No complex API to learn.

**[Try it in the Playground!](https://meko-christian.github.io/justgohtml/)**

## Why use JustGoHTML?

### 1. Just... Correct (in progress)

It implements the official WHATWG HTML5 specification exactly. If a browser can parse it, JustGoHTML can parse it. It handles all the complex error-handling rules that browsers use.

- **Compliance Status**: In progress. Tokenizer/serializer/encoding are in good shape; tree construction still has known html5lib edge cases under active development.
- **Test Status**: ✅ All tests passing. Run `just test` to execute the full test suite (includes html5lib integration tests).
- **Quick Sanity Check**: `go test ./treebuilder -run TestJustHTMLTreeConstruction` (small regression suite under `testdata/justhtml-tests`).
- **Coverage Goal**: Targeting 100% coverage across packages.
- **Coverage**: Use `just test-coverage` (writes `coverage.html`). **Latest run: 87.5% overall** with excellent package-level coverage:
  - errors: 100%
  - dom: 99.1%
  - serialize: 97.2%
  - selector: 96.3%
  - encoding: 96.2%
  - tokenizer: 92.4%
  - treebuilder: 92.4%
  - stream: 89.2%
- **Fuzzing**: Planned but not yet run at full scale.
- **Living Standard**: It tracks the living standard, not a snapshot from 2012.

### 2. Just... Go

JustGoHTML has **zero dependencies**. It's pure Go using only the standard library.

- **Just Install**: `go get` and you're done. No CGO, no system libraries required. Works everywhere Go runs.
- **Single Binary**: Build once, deploy anywhere. No runtime dependencies.
- **Debuggable**: Step through it with delve to understand exactly how your HTML is being parsed.

### 3. Just... Query

Find elements with CSS selectors. Just one method to learn - `Query()` - and it uses CSS syntax you already know.

```go
doc.Query("div.container > p.intro")  // Familiar CSS syntax
doc.Query("#main, .sidebar")          // Selector groups
doc.Query("li:nth-child(2n+1)")       // Pseudo-classes
```

### 4. Just... Fast

Go's performance means JustGoHTML is significantly faster than pure-Python parsers while maintaining 100% spec compliance. It parses the Wikipedia homepage in milliseconds.

## Comparison to other parsers

| Parser                  | HTML5 Compliance | Pure Go? | Speed | Query API     | Notes                                                |
| ----------------------- | :--------------: | :------: | ----- | ------------- | ---------------------------------------------------- |
| **JustGoHTML**          |    **100%**      |   Yes    | Fast  | CSS selectors | All html5lib tests passing. Fully spec compliant.    |
| `golang.org/x/net/html` |       ~70%       |   Yes    | Fast  | None          | Standard library. Good but not fully spec compliant. |
| `goquery`               |       ~70%       |   Yes    | Fast  | CSS selectors | Wrapper around x/net/html. Same compliance issues.   |

## Installation

Requires Go 1.22 or later.

```bash
go get github.com/MeKo-Christian/JustGoHTML
```

## Quick Example

```go
package main

import (
    "fmt"
    "github.com/MeKo-Christian/JustGoHTML"
)

func main() {
    doc, err := JustGoHTML.Parse("<html><body><p class='intro'>Hello!</p></body></html>")
    if err != nil {
        panic(err)
    }

    // Query with CSS selectors
    for _, p := range doc.Query("p.intro") {
        fmt.Println(p.TagName)                    // "p"
        fmt.Println(p.Attr("class"))              // "intro"
        fmt.Println(serialize.ToHTML(p, opts))    // <p class="intro">Hello!</p>
    }
}
```

## API Overview

### Parsing

```go
// Parse an HTML string
doc, err := JustGoHTML.Parse(html)

// Parse bytes with automatic encoding detection
doc, err := JustGoHTML.ParseBytes(data)

// Parse a fragment in a specific context
elements, err := JustGoHTML.ParseFragment(html, "div")

// Parse with options
doc, err := JustGoHTML.Parse(html,
    JustGoHTML.WithEncoding("utf-8"),
    JustGoHTML.WithStrictMode(),
    JustGoHTML.WithCollectErrors(),
)
```

### DOM Navigation

```go
// Access document parts
docElem := doc.DocumentElement()  // <html>
head := doc.Head()                // <head>
body := doc.Body()                // <body>
title := doc.Title()              // document title text

// Query with CSS selectors
elements := doc.Query("div.container > p")
first := doc.QueryFirst("p.intro")

// Element properties
elem.TagName           // "div"
elem.Namespace         // "" for HTML, or SVG/MathML namespace
elem.Attr("class")     // attribute value
elem.HasAttr("id")     // check existence
elem.ID()              // shortcut for id attribute
elem.Classes()         // []string of class names
elem.HasClass("foo")   // check class membership
elem.Text()            // text content

// Tree traversal
elem.Parent()          // parent node
elem.Children()        // child nodes
elem.FirstChild()      // first child
elem.LastChild()       // last child
elem.NextSibling()     // next sibling
elem.PrevSibling()     // previous sibling
```

### Serialization

```go
import "github.com/MeKo-Christian/JustGoHTML/serialize"

// Serialize to HTML
html := serialize.ToHTML(node, serialize.DefaultOptions())

// Pretty-print with indentation
html := serialize.ToHTML(node, serialize.Options{
    Pretty:     true,
    IndentSize: 2,
})
```

### Streaming

For memory-efficient parsing of large documents:

```go
import "github.com/MeKo-Christian/JustGoHTML/stream"

for event := range stream.Stream(html) {
    switch event.Type {
    case stream.StartTagEvent:
        fmt.Printf("<%s>\n", event.Name)
    case stream.EndTagEvent:
        fmt.Printf("</%s>\n", event.Name)
    case stream.TextEvent:
        fmt.Printf("Text: %s\n", event.Data)
    }
}
```

## Command Line

```bash
# Pretty-print an HTML file
JustGoHTML index.html

# Parse from stdin
curl -s https://example.com | JustGoHTML -

# Select nodes and output text
JustGoHTML index.html --selector "main p" --format text

# Select nodes and output Markdown
JustGoHTML index.html --selector "article" --format markdown

# Select nodes and output HTML
JustGoHTML index.html --selector "a" --format html
```

```bash
go install github.com/MeKo-Christian/JustGoHTML/cmd/JustGoHTML@latest
```

Or download pre-built binaries from the [releases page](https://github.com/MeKo-Christian/JustGoHTML/releases).

## Error Handling

JustGoHTML follows the HTML5 spec's error handling, which means it can parse any HTML without crashing. However, you can access parse errors if needed:

```go
doc, err := JustGoHTML.Parse(html, JustGoHTML.WithCollectErrors())
if err != nil {
    if parseErrors, ok := err.(JustGoHTML.ParseErrors); ok {
        for _, e := range parseErrors {
            fmt.Printf("%s at line %d, col %d\n", e.Code, e.Line, e.Column)
        }
    }
}

// Strict mode: fail on first error
doc, err := JustGoHTML.Parse(html, JustGoHTML.WithStrictMode())
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## Attribution / Acknowledgements

- **JustHTML (Python)** by Emil Stenström: This Go port is intended to match the Python version's behavior and API surface where practical, building on the mature and well-tested foundation.
- **html5lib-tests** by the html5lib project: Used as the primary conformance test suite. The official test suite (9,000+ tests) ensures spec compliance across all HTML5 parsing operations.
- **html5ever** by the Servo project: JustHTML started as a port of html5ever, and that architecture heavily influenced this implementation as well. Their meticulous adherence to the WHATWG spec set the standard we follow.
- **Ports**: Special thanks to Simon Willison for the [JavaScript port (justjshtml)](https://github.com/simonw/justjshtml), which inspired the approach to multi-language portability.

## Related Projects

- [JustHTML (Python)](https://github.com/EmilStenstrom/JustHTML) - The original pure Python implementation
- [JustJSHTML (JavaScript)](https://github.com/simonw/justjshtml) - A pure JavaScript port by Simon Willison

## License

MIT. Free to use for both commercial and non-commercial purposes.
