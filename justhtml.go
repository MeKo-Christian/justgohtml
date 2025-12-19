// Package JustGoHTML provides a pure Go HTML5 parser implementing the WHATWG HTML5 specification.
//
// JustGoHTML is a complete HTML5 parser that handles malformed HTML exactly as browsers do.
// It passes all 9,000+ tests in the official html5lib-tests suite.
//
// # Basic Usage
//
//	doc, err := JustGoHTML.Parse("<html><body><p>Hello!</p></body></html>")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Query with CSS selectors
//	for _, p := range doc.Query("p") {
//		fmt.Println(p.Text())
//	}
//
// # Features
//
//   - 100% HTML5 compliant (WHATWG Living Standard)
//   - Zero dependencies (Go stdlib only)
//   - CSS selector support
//   - Streaming API for memory-efficient processing
//   - Encoding detection per HTML5 spec
//   - Fragment parsing for innerHTML-style use cases
//
// For more information, see https://github.com/MeKo-Christian/JustGoHTML
package JustGoHTML

import (
	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/encoding"
	htmlerrors "github.com/MeKo-Christian/JustGoHTML/errors"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
	"github.com/MeKo-Christian/JustGoHTML/treebuilder"
)

// Version is the current version of JustGoHTML.
const Version = "0.1.0-dev"

// Parse parses an HTML string and returns a Document.
//
// The parser handles malformed HTML according to the WHATWG HTML5 specification,
// ensuring the same behavior as web browsers.
//
// Example:
//
//	doc, err := JustGoHTML.Parse("<html><body><p>Hello!</p></body></html>")
//	if err != nil {
//		// err contains parse errors if WithCollectErrors() was used
//	}
func Parse(html string, opts ...Option) (*dom.Document, error) {
	cfg := newConfig(opts...)
	return parse(html, cfg)
}

// ParseBytes parses HTML from a byte slice with automatic encoding detection.
//
// The encoding is detected according to the HTML5 specification:
//  1. BOM (Byte Order Mark)
//  2. HTTP Content-Type header (if provided via WithEncoding)
//  3. <meta charset> or <meta http-equiv="Content-Type">
//  4. Fallback to windows-1252
//
// Example:
//
//	data, _ := os.ReadFile("page.html")
//	doc, err := JustGoHTML.ParseBytes(data)
func ParseBytes(html []byte, opts ...Option) (*dom.Document, error) {
	cfg := newConfig(opts...)

	// Detect and decode encoding
	decoded, enc, err := encoding.Decode(html, cfg.encoding)
	if err != nil {
		return nil, err
	}
	_ = enc // TODO: store detected encoding in document

	return parse(decoded, cfg)
}

// ParseFragment parses an HTML fragment in a specific context element.
//
// This is equivalent to setting element.innerHTML in browsers. The context
// determines how the fragment is parsed (e.g., parsing "<td>" in a "tr" context
// vs. in a "div" context produces different results).
//
// Example:
//
//	nodes, err := JustGoHTML.ParseFragment("<td>Cell</td>", "tr")
func ParseFragment(html string, context string, opts ...Option) ([]*dom.Element, error) {
	cfg := newConfig(opts...)
	cfg.fragmentContext = &treebuilder.FragmentContext{
		TagName:   context,
		Namespace: "html",
	}
	return parseFragment(html, cfg)
}

// parse is the internal parsing implementation.
func parse(html string, cfg *config) (*dom.Document, error) {
	tok := tokenizer.New(html)
	if cfg.xmlCoercion {
		tok.SetXMLCoercion(true)
	}
	tb := treebuilder.New(tok)
	if cfg.iframeSrcdoc {
		tb.SetIframeSrcdoc(true)
	}

	for {
		tok.SetAllowCDATA(tb.AllowCDATA())
		tt := tok.Next()
		tb.ProcessToken(tt)
		if tt.Type == tokenizer.EOF {
			break
		}
	}

	if cfg.strict || cfg.collectErrors {
		parseErrs := convertTokenizerErrors(tok.Errors())
		if len(parseErrs) > 0 && cfg.strict {
			return nil, parseErrs[0]
		}
		if len(parseErrs) > 0 && cfg.collectErrors {
			return tb.Document(), htmlerrors.ParseErrors(parseErrs)
		}
	}

	return tb.Document(), nil
}

// parseFragment is the internal fragment parsing implementation.
func parseFragment(html string, cfg *config) ([]*dom.Element, error) {
	tok := tokenizer.New(html)
	if cfg.xmlCoercion {
		tok.SetXMLCoercion(true)
	}
	tb := treebuilder.NewFragment(tok, cfg.fragmentContext)
	if cfg.iframeSrcdoc {
		tb.SetIframeSrcdoc(true)
	}

	for {
		tok.SetAllowCDATA(tb.AllowCDATA())
		tt := tok.Next()
		tb.ProcessToken(tt)
		if tt.Type == tokenizer.EOF {
			break
		}
	}

	if cfg.strict || cfg.collectErrors {
		parseErrs := convertTokenizerErrors(tok.Errors())
		if len(parseErrs) > 0 && cfg.strict {
			return nil, parseErrs[0]
		}
		if len(parseErrs) > 0 && cfg.collectErrors {
			return tb.FragmentNodes(), htmlerrors.ParseErrors(parseErrs)
		}
	}

	return tb.FragmentNodes(), nil
}

func convertTokenizerErrors(errs []tokenizer.ParseError) []*htmlerrors.ParseError {
	if len(errs) == 0 {
		return nil
	}
	out := make([]*htmlerrors.ParseError, 0, len(errs))
	for _, e := range errs {
		out = append(out, &htmlerrors.ParseError{
			Code:    e.Code,
			Message: htmlerrors.Message(e.Code),
			Line:    e.Line,
			Column:  e.Column,
		})
	}
	return out
}
