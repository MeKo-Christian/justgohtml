package treebuilder_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML"
	"github.com/MeKo-Christian/JustGoHTML/dom"
	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
	"github.com/MeKo-Christian/JustGoHTML/treebuilder"
)

const (
	html5libTestsDir = "../testdata/html5lib-tests/tree-construction"
	JustHTMLTestsDir = "../testdata/justhtml-tests"
)

// TestHTML5LibTreeConstruction runs all html5lib tree-construction tests.
func TestHTML5LibTreeConstruction(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(html5libTestsDir); os.IsNotExist(err) {
		t.Skip("html5lib-tests not found - run 'git submodule update --init'")
	}

	files, err := testutil.CollectTestFiles(html5libTestsDir, "*.dat")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No tree-construction test files found")
	}

	strict := os.Getenv("JustHTML_HTML5LIB_STRICT") == "1" || os.Getenv("JustHTML_HTML5LIB_TREE_STRICT") == "1"

	for _, file := range files {
		// capture for parallel
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTreeConstructionTestFile(t, file, strict)
		})
	}
}

// TestJustHTMLTreeConstruction runs JustHTML-specific tree-construction tests.
func TestJustHTMLTreeConstruction(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(JustHTMLTestsDir); os.IsNotExist(err) {
		t.Skip("JustHTML-tests not found")
	}

	files, err := testutil.CollectTestFiles(JustHTMLTestsDir, "*.dat")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No JustHTML tree-construction test files found")
	}

	strict := os.Getenv("JustHTML_JustHTML_TREE_STRICT") == "1" || os.Getenv("JustHTML_HTML5LIB_STRICT") == "1"

	for _, file := range files {
		// capture for parallel
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTreeConstructionTestFile(t, file, strict)
		})
	}
}

func runTreeConstructionTestFile(t *testing.T, path string, strict bool) {
	t.Helper()
	tests, err := testutil.ParseTreeConstructionFile(path)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	if strict {
		var passed, failed, skipped int64
		var mu sync.Mutex
		var examples []string

		for i, test := range tests {
			testName := truncate(test.Data, 40)
			if testName == "" {
				testName = "empty"
			}
			testIndex := i
			t.Run(testName, func(t *testing.T) {
				got, want, skipReason, err := runSingleTreeConstructionTest(test)
				if skipReason != "" {
					atomic.AddInt64(&skipped, 1)
					t.Skip(skipReason)
				}
				if err != nil {
					atomic.AddInt64(&failed, 1)
					t.Fatalf("parse error: %v\nInput: %q", err, truncate(test.Data, 100))
				}

				if got == want {
					atomic.AddInt64(&passed, 1)
					return
				}

				atomic.AddInt64(&failed, 1)
				t.Errorf("tree mismatch\ninput: %q\n\nwant:\n%s\n\ngot:\n%s", truncate(test.Data, 200), want, got)

				mu.Lock()
				if len(examples) < 3 {
					examples = append(examples, fmt.Sprintf("case %d input %q\nwant:\n%s\n\ngot:\n%s", testIndex, truncate(test.Data, 120), want, got))
				}
				mu.Unlock()
			})
		}

		if testing.Verbose() {
			t.Logf("summary: %d passed, %d failed, %d skipped", passed, failed, skipped)
			if len(examples) > 0 {
				t.Logf("examples:\n%s", strings.Join(examples, "\n\n"))
			}
		}
		return
	}

	var passed, failed, skipped int
	var examples []string

	for _, test := range tests {
		got, want, skipReason, err := runSingleTreeConstructionTest(test)
		if skipReason != "" {
			skipped++
			continue
		}
		if err != nil {
			failed++
			if len(examples) < 3 {
				examples = append(examples, fmt.Sprintf("parse error: %v\ninput: %q", err, truncate(test.Data, 120)))
			}
			continue
		}
		if got == want {
			passed++
			continue
		}

		failed++
		if len(examples) < 3 {
			examples = append(examples, fmt.Sprintf("input %q\nwant:\n%s\n\ngot:\n%s", truncate(test.Data, 120), want, got))
		}
	}

	if testing.Verbose() {
		t.Logf("summary: %d passed, %d failed, %d skipped (run 'just test-spec-strict' to fail on mismatches)", passed, failed, skipped)
		if len(examples) > 0 {
			t.Logf("examples:\n%s", strings.Join(examples, "\n\n"))
		}
	}
}

func runSingleTreeConstructionTest(test testutil.TreeConstructionTest) (got string, want string, skipReason string, err error) {
	// Skip script tests for now.
	if test.ScriptDirective == "script-on" {
		return "", "", "script-on tests not yet supported", nil
	}

	want = strings.TrimRight(test.Document, "\n")

	if test.FragmentContext != "" {
		got, err = parseHTML5LibFragment(test.Data, test.FragmentContext)
		return got, want, "", err
	}

	doc, err := JustGoHTML.Parse(test.Data)
	if err != nil {
		return "", want, "", err
	}
	return testutil.SerializeHTML5LibTree(doc), want, "", nil
}

func parseHTML5LibFragment(input string, ctx string) (string, error) {
	fc, err := parseFragmentContext(ctx)
	if err != nil {
		return "", err
	}

	tok := tokenizer.New(input)
	tb := treebuilder.NewFragment(tok, fc)

	for {
		tt := tok.Next()
		tb.ProcessToken(tt)
		if tt.Type == tokenizer.EOF {
			break
		}
	}

	doc := tb.Document()
	contextEl, err := firstChildElement(doc.DocumentElement())
	if err != nil {
		return "", err
	}
	return testutil.SerializeHTML5LibNodes(contextEl.Children()), nil
}

func parseFragmentContext(s string) (*treebuilder.FragmentContext, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty fragment context")
	}
	if len(fields) == 1 {
		return &treebuilder.FragmentContext{TagName: fields[0], Namespace: "html"}, nil
	}

	ns := fields[0]
	tag := strings.Join(fields[1:], " ")
	switch ns {
	case "svg":
		return &treebuilder.FragmentContext{TagName: tag, Namespace: "svg"}, nil
	case "math":
		return &treebuilder.FragmentContext{TagName: tag, Namespace: "mathml"}, nil
	default:
		// Unknown namespace designator; treat the whole context as an HTML tag name.
		return &treebuilder.FragmentContext{TagName: s, Namespace: "html"}, nil
	}
}

func firstChildElement(el *dom.Element) (*dom.Element, error) {
	if el == nil {
		return nil, fmt.Errorf("missing document element")
	}
	for _, child := range el.Children() {
		if e, ok := child.(*dom.Element); ok {
			return e, nil
		}
	}
	return nil, fmt.Errorf("missing context element")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// BenchmarkTreeBuilder benchmarks the full parsing pipeline.
func BenchmarkTreeBuilder(b *testing.B) {
	// Create a moderately complex HTML document.
	html := strings.Repeat("<div class='test'><p>Hello, <b>world</b>!</p><ul><li>Item 1</li><li>Item 2</li></ul></div>", 1000)

	b.ResetTimer()
	for range b.N {
		_, _ = JustGoHTML.Parse(html)
	}
}

// BenchmarkTreeBuilderSmall benchmarks parsing a small HTML document.
func BenchmarkTreeBuilderSmall(b *testing.B) {
	html := "<!DOCTYPE html><html><head><title>Test</title></head><body><p>Hello</p></body></html>"

	b.ResetTimer()
	for range b.N {
		_, _ = JustGoHTML.Parse(html)
	}
}

// BenchmarkTreeBuilderTables benchmarks parsing HTML with tables.
func BenchmarkTreeBuilderTables(b *testing.B) {
	row := "<tr><td>Cell 1</td><td>Cell 2</td><td>Cell 3</td></tr>"
	html := "<!DOCTYPE html><table>" + strings.Repeat(row, 100) + "</table>"

	b.ResetTimer()
	for range b.N {
		_, _ = JustGoHTML.Parse(html)
	}
}

// BenchmarkTreeBuilderNested benchmarks parsing deeply nested HTML.
func BenchmarkTreeBuilderNested(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("<div>")
	}
	sb.WriteString("content")
	for i := 0; i < 100; i++ {
		sb.WriteString("</div>")
	}
	html := sb.String()

	b.ResetTimer()
	for range b.N {
		_, _ = JustGoHTML.Parse(html)
	}
}

// BenchmarkTreeBuilderForeign benchmarks parsing HTML with SVG content.
func BenchmarkTreeBuilderForeign(b *testing.B) {
	html := strings.Repeat("<div><svg><circle r='10'/><rect width='20' height='10'/></svg></div>", 100)

	b.ResetTimer()
	for range b.N {
		_, _ = JustGoHTML.Parse(html)
	}
}
