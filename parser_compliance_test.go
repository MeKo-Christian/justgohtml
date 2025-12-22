// Package JustGoHTML provides comprehensive compliance and benchmark comparisons
// between Go HTML parsers using the official html5lib-tests suite.
package JustGoHTML

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"golang.org/x/net/html"

	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
)

const (
	html5libTreeTestsDir = "testdata/html5lib-tests/tree-construction"
)

// =============================================================================
// HTML5Lib Tree Serialization for golang.org/x/net/html
// =============================================================================

// serializeNetHTMLTree converts a net/html parsed document to html5lib format.
func serializeNetHTMLTree(doc *html.Node) string {
	var sb strings.Builder

	// Find doctype and children
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			serializeNetHTMLDoctype(&sb, c)
		} else {
			serializeNetHTMLNode(&sb, c, 0)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func serializeNetHTMLDoctype(sb *strings.Builder, n *html.Node) {
	sb.WriteString("| <!DOCTYPE ")
	if n.Data == "" {
		sb.WriteString(">")
		sb.WriteByte('\n')
		return
	}

	sb.WriteString(n.Data)

	// net/html stores public/system IDs in Attr
	publicID, systemID := "", ""
	for _, a := range n.Attr {
		switch a.Key {
		case "public":
			publicID = a.Val
		case "system":
			systemID = a.Val
		}
	}

	if publicID != "" || systemID != "" {
		sb.WriteString(" \"")
		sb.WriteString(publicID)
		sb.WriteString("\" \"")
		sb.WriteString(systemID)
		sb.WriteString("\">")
	} else {
		sb.WriteString(">")
	}
	sb.WriteByte('\n')
}

func serializeNetHTMLNode(sb *strings.Builder, n *html.Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n.Type { //nolint:exhaustive // Only handling node types that appear in tree output
	case html.ElementNode:
		sb.WriteString("| ")
		sb.WriteString(indent)
		sb.WriteString("<")
		// Handle namespace
		switch n.Namespace {
		case "", "html":
			sb.WriteString(n.Data)
		case "svg":
			sb.WriteString("svg ")
			sb.WriteString(n.Data)
		case "math":
			sb.WriteString("math ")
			sb.WriteString(n.Data)
		default:
			sb.WriteString(n.Namespace)
			sb.WriteString(" ")
			sb.WriteString(n.Data)
		}
		sb.WriteString(">")
		sb.WriteByte('\n')

		// Sort attributes alphabetically (html5lib format requirement)
		attrs := make([]html.Attribute, len(n.Attr))
		copy(attrs, n.Attr)
		sort.Slice(attrs, func(i, j int) bool {
			return formatNetHTMLAttrName(attrs[i]) < formatNetHTMLAttrName(attrs[j])
		})
		for _, attr := range attrs {
			sb.WriteString("| ")
			sb.WriteString(indent)
			sb.WriteString("  ")
			sb.WriteString(formatNetHTMLAttrName(attr))
			sb.WriteString("=\"")
			sb.WriteString(attr.Val)
			sb.WriteString("\"")
			sb.WriteByte('\n')
		}

		// Handle template content
		if n.Data == "template" && n.Namespace == "" {
			sb.WriteString("| ")
			sb.WriteString(strings.Repeat("  ", depth+1))
			sb.WriteString("content")
			sb.WriteByte('\n')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				serializeNetHTMLNode(sb, c, depth+2)
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				serializeNetHTMLNode(sb, c, depth+1)
			}
		}

	case html.TextNode:
		sb.WriteString("| ")
		sb.WriteString(indent)
		sb.WriteString("\"")
		sb.WriteString(n.Data)
		sb.WriteString("\"")
		sb.WriteByte('\n')

	case html.CommentNode:
		sb.WriteString("| ")
		sb.WriteString(indent)
		sb.WriteString("<!-- ")
		sb.WriteString(n.Data)
		sb.WriteString(" -->")
		sb.WriteByte('\n')

	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			serializeNetHTMLNode(sb, c, depth)
		}
	}
}

func formatNetHTMLAttrName(attr html.Attribute) string {
	switch attr.Namespace {
	case "":
		return attr.Key
	case "http://www.w3.org/1999/xlink":
		local := attr.Key
		if idx := strings.IndexByte(local, ':'); idx >= 0 {
			local = local[idx+1:]
		}
		return "xlink " + local
	case "http://www.w3.org/XML/1998/namespace":
		local := attr.Key
		if idx := strings.IndexByte(local, ':'); idx >= 0 {
			local = local[idx+1:]
		}
		return "xml " + local
	case "http://www.w3.org/2000/xmlns/":
		local := attr.Key
		if idx := strings.IndexByte(local, ':'); idx >= 0 {
			local = local[idx+1:]
		}
		return "xmlns " + local
	default:
		return attr.Namespace + " " + attr.Key
	}
}

// =============================================================================
// Compliance Test Types and Common Runner
// =============================================================================

// ComplianceResult holds the results of running html5lib tests against a parser.
type ComplianceResult struct {
	ParserName string
	Passed     int
	Failed     int
	Skipped    int
	Total      int
	Percentage float64
	Failures   []ComplianceFailure
}

// ComplianceFailure records a single test failure.
type ComplianceFailure struct {
	File     string
	Input    string
	Expected string
	Got      string
	Reason   string
}

// parseAndSerializeFunc is a function that parses HTML and returns the serialized tree.
type parseAndSerializeFunc func(test testutil.TreeConstructionTest) (got string, err error, skip bool)

// runComplianceTests is a generic compliance test runner.
func runComplianceTests(t *testing.T, parserName string, parseFunc parseAndSerializeFunc) ComplianceResult {
	t.Helper()

	result := ComplianceResult{ParserName: parserName}

	if _, err := os.Stat(html5libTreeTestsDir); os.IsNotExist(err) {
		t.Skip("html5lib-tests not found - run 'git submodule update --init'")
	}

	files, err := testutil.CollectTestFiles(html5libTreeTestsDir, "*.dat")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			passed, failed, skipped, failures := runComplianceFile(path, parseFunc)

			mu.Lock()
			result.Passed += passed
			result.Failed += failed
			result.Skipped += skipped
			result.Failures = append(result.Failures, failures...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	result.Total = result.Passed + result.Failed + result.Skipped
	if result.Passed+result.Failed > 0 {
		result.Percentage = float64(result.Passed) * 100 / float64(result.Passed+result.Failed)
	}

	return result
}

//nolint:nonamedreturns // Named returns provide clarity for multiple return values
func runComplianceFile(path string, parseFunc parseAndSerializeFunc) (passed, failed, skipped int, failures []ComplianceFailure) {
	tests, err := testutil.ParseTreeConstructionFile(path)
	if err != nil {
		return 0, 0, 0, nil
	}

	fileName := filepath.Base(path)

	for _, test := range tests {
		// Skip script tests (no JS engine)
		if test.ScriptDirective == "script-on" {
			skipped++
			continue
		}

		// Skip fragment parsing tests (requires special handling)
		if test.FragmentContext != "" {
			skipped++
			continue
		}

		got, err, skip := parseFunc(test)
		if skip {
			skipped++
			continue
		}
		if err != nil {
			failed++
			failures = append(failures, ComplianceFailure{
				File:   fileName,
				Input:  test.Data,
				Reason: fmt.Sprintf("parse error: %v", err),
			})
			continue
		}

		want := strings.TrimRight(test.Document, "\n")

		if got == want {
			passed++
		} else {
			failed++
			failures = append(failures, ComplianceFailure{
				File:     fileName,
				Input:    test.Data,
				Expected: want,
				Got:      got,
			})
		}
	}

	return passed, failed, skipped, failures
}

// =============================================================================
// Parser-specific implementations
// =============================================================================

func netHTMLParseFunc(test testutil.TreeConstructionTest) (string, error, bool) {
	// Skip iframe-srcdoc tests (net/html doesn't support this mode)
	if test.IframeSrcdoc {
		return "", nil, true
	}

	doc, err := html.Parse(strings.NewReader(test.Data))
	if err != nil {
		return "", err, false
	}

	return serializeNetHTMLTree(doc), nil, false
}

func justGoHTMLParseFunc(test testutil.TreeConstructionTest) (string, error, bool) {
	opts := []Option{}
	if test.IframeSrcdoc {
		opts = append(opts, WithIframeSrcdoc())
	}
	if test.XMLCoercion {
		opts = append(opts, WithXMLCoercion())
	}

	doc, err := Parse(test.Data, opts...)
	if err != nil {
		return "", err, false
	}

	return testutil.SerializeHTML5LibTree(doc), nil, false
}

// =============================================================================
// Public Test Functions
// =============================================================================

// TestNetHTMLCompliance runs html5lib tree-construction tests against net/html.
func TestNetHTMLCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compliance test in short mode")
	}

	result := runComplianceTests(t, "golang.org/x/net/html", netHTMLParseFunc)

	t.Logf("\n=== golang.org/x/net/html HTML5 Compliance ===")
	t.Logf("Passed:     %d", result.Passed)
	t.Logf("Failed:     %d", result.Failed)
	t.Logf("Skipped:    %d", result.Skipped)
	t.Logf("Total:      %d", result.Total)
	t.Logf("Compliance: %.2f%%", result.Percentage)

	// Show sample failures (first 10)
	if len(result.Failures) > 0 {
		t.Logf("\nSample failures (first 10):")
		for i, f := range result.Failures {
			if i >= 10 {
				break
			}
			t.Logf("\n--- Failure %d: %s ---", i+1, f.File)
			t.Logf("Input: %q", truncateStr(f.Input, 100))
			if f.Reason != "" {
				t.Logf("Reason: %s", f.Reason)
			}
		}
	}
}

// TestJustGoHTMLCompliance runs html5lib tree-construction tests against JustGoHTML.
func TestJustGoHTMLCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compliance test in short mode")
	}

	result := runComplianceTests(t, "JustGoHTML", justGoHTMLParseFunc)

	t.Logf("\n=== JustGoHTML HTML5 Compliance ===")
	t.Logf("Passed:     %d", result.Passed)
	t.Logf("Failed:     %d", result.Failed)
	t.Logf("Skipped:    %d", result.Skipped)
	t.Logf("Total:      %d", result.Total)
	t.Logf("Compliance: %.2f%%", result.Percentage)
}

// TestParserComplianceComparison runs compliance tests for all parsers and outputs
// a comparison table.
func TestParserComplianceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compliance comparison in short mode")
	}

	// Run tests for each parser in parallel
	var justGoResult, netHTMLResult ComplianceResult
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		justGoResult = runComplianceTests(t, "JustGoHTML", justGoHTMLParseFunc)
	}()
	go func() {
		defer wg.Done()
		netHTMLResult = runComplianceTests(t, "golang.org/x/net/html", netHTMLParseFunc)
	}()
	wg.Wait()

	// Output comparison table
	t.Logf("\n")
	t.Logf("╔══════════════════════════════════════════════════════════════════════════════╗")
	t.Logf("║                    HTML5 Tree Construction Compliance                        ║")
	t.Logf("╠═══════════════════════════╦═════════╦═════════╦═════════╦═══════════════════╣")
	t.Logf("║ Parser                    ║ Passed  ║ Failed  ║ Skipped ║ Compliance        ║")
	t.Logf("╠═══════════════════════════╬═════════╬═════════╬═════════╬═══════════════════╣")
	t.Logf("║ %-25s ║ %7d ║ %7d ║ %7d ║ %6.2f%%           ║",
		"JustGoHTML", justGoResult.Passed, justGoResult.Failed, justGoResult.Skipped, justGoResult.Percentage)
	t.Logf("║ %-25s ║ %7d ║ %7d ║ %7d ║ %6.2f%%           ║",
		"golang.org/x/net/html", netHTMLResult.Passed, netHTMLResult.Failed, netHTMLResult.Skipped, netHTMLResult.Percentage)
	t.Logf("╚═══════════════════════════╩═════════╩═════════╩═════════╩═══════════════════╝")
	t.Logf("\nNote: goquery uses golang.org/x/net/html as its parser, so compliance is identical.")
	t.Logf("Tests skipped: script-on tests (no JS engine) and fragment parsing tests.")
}

// =============================================================================
// Utility Functions
// =============================================================================

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
