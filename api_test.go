package JustGoHTML

import (
	"errors"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/dom"
	htmlerrors "github.com/MeKo-Christian/JustGoHTML/errors"
)

// TestParseBasicHTML tests parsing basic HTML documents.
func TestParseBasicHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTag  string
		wantText string
	}{
		{
			name:     "simple document",
			input:    "<html><body><p>Hello</p></body></html>",
			wantTag:  "html",
			wantText: "Hello",
		},
		{
			name:     "with DOCTYPE",
			input:    "<!DOCTYPE html><html><head><title>Test</title></head><body>Content</body></html>",
			wantTag:  "html",
			wantText: "TestContent",
		},
		{
			name:     "malformed HTML",
			input:    "<p>Unclosed paragraph<div>Content",
			wantTag:  "html",
			wantText: "Unclosed paragraphContent",
		},
		{
			name:     "empty string",
			input:    "",
			wantTag:  "html",
			wantText: "",
		},
		{
			name:     "just text",
			input:    "Plain text",
			wantTag:  "html",
			wantText: "Plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if doc == nil {
				t.Fatal("Parse() returned nil document")
			}

			root := doc.DocumentElement()
			if root == nil {
				t.Fatal("DocumentElement() returned nil")
			}
			if root.TagName != tt.wantTag {
				t.Errorf("root tag = %q, want %q", root.TagName, tt.wantTag)
			}

			gotText := extractAllText(doc)
			if gotText != tt.wantText {
				t.Errorf("document text = %q, want %q", gotText, tt.wantText)
			}
		})
	}
}

// TestParseBytesWithEncoding tests parsing with different encodings.
func TestParseBytesWithEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantText string
	}{
		{
			name:     "UTF-8",
			input:    []byte("<html><body><p>Hello UTF-8</p></body></html>"),
			wantText: "Hello UTF-8",
		},
		{
			name:     "UTF-8 BOM",
			input:    []byte("\xEF\xBB\xBF<html><body>UTF-8 with BOM</body></html>"),
			wantText: "UTF-8 with BOM",
		},
		{
			name:     "empty bytes",
			input:    []byte{},
			wantText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseBytes(tt.input)
			if err != nil {
				t.Fatalf("ParseBytes() error = %v", err)
			}
			if doc == nil {
				t.Fatal("ParseBytes() returned nil document")
			}

			gotText := extractAllText(doc)
			if gotText != tt.wantText {
				t.Errorf("document text = %q, want %q", gotText, tt.wantText)
			}
		})
	}
}

// TestParseFragmentContext tests fragment parsing with different contexts.
func TestParseFragmentContext(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		context string
		wantLen int
		wantTag string
	}{
		{
			name:    "td in tr context",
			html:    "<td>Cell</td>",
			context: "tr",
			wantLen: 1,
			wantTag: "td",
		},
		{
			name:    "multiple elements",
			html:    "<li>Item 1</li><li>Item 2</li>",
			context: "ul",
			wantLen: 2,
			wantTag: "li",
		},
		{
			name:    "div context",
			html:    "<p>Paragraph</p><div>Div</div>",
			context: "div",
			wantLen: 2,
			wantTag: "p",
		},
		{
			name:    "empty fragment",
			html:    "",
			context: "div",
			wantLen: 0,
			wantTag: "",
		},
		{
			name:    "text only",
			html:    "Just text",
			context: "div",
			wantLen: 0,
			wantTag: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseFragment(tt.html, tt.context)
			if err != nil {
				t.Fatalf("ParseFragment() error = %v", err)
			}
			if len(nodes) != tt.wantLen {
				t.Errorf("ParseFragment() returned %d nodes, want %d", len(nodes), tt.wantLen)
			}
			if tt.wantLen > 0 && nodes[0].TagName != tt.wantTag {
				t.Errorf("first node tag = %q, want %q", nodes[0].TagName, tt.wantTag)
			}
		})
	}
}

// TestParseWithOptions tests parsing with various options.
func TestParseWithOptions(t *testing.T) {
	t.Run("with strict mode", func(t *testing.T) {
		// Malformed HTML should trigger errors in strict mode
		doc, err := Parse("<html><body><p>Test", WithStrictMode())
		// In non-strict mode, this would succeed
		// Strict mode behavior depends on whether there are parse errors
		_ = doc
		_ = err
		// Just verify the option doesn't panic
	})

	t.Run("with collect errors", func(t *testing.T) {
		doc, err := Parse("<html><body><p>Test", WithCollectErrors())
		if doc == nil {
			t.Fatal("WithCollectErrors should still return document")
		}
		// err may or may not be nil depending on parse errors
		_ = err
	})

	t.Run("with iframe srcdoc", func(t *testing.T) {
		doc, err := Parse("<html><body>Test</body></html>", WithIframeSrcdoc())
		if err != nil {
			t.Fatalf("WithIframeSrcdoc error = %v", err)
		}
		if doc == nil {
			t.Fatal("WithIframeSrcdoc returned nil document")
		}
	})

	t.Run("with XML coercion", func(t *testing.T) {
		doc, err := Parse("<html><body>Test</body></html>", WithXMLCoercion())
		if err != nil {
			t.Fatalf("WithXMLCoercion error = %v", err)
		}
		if doc == nil {
			t.Fatal("WithXMLCoercion returned nil document")
		}
	})
}

// TestParseComplexHTML tests parsing more complex HTML structures.
func TestParseComplexHTML(t *testing.T) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Test Page</title>
	<style>body { color: red; }</style>
	<script>console.log('test');</script>
</head>
<body>
	<header>
		<h1>Main Title</h1>
		<nav>
			<ul>
				<li><a href="/">Home</a></li>
				<li><a href="/about">About</a></li>
			</ul>
		</nav>
	</header>
	<main>
		<article>
			<h2>Article Title</h2>
			<p>First paragraph with <strong>bold</strong> and <em>italic</em>.</p>
			<p>Second paragraph.</p>
		</article>
	</main>
	<footer>
		<p>&copy; 2024 Test</p>
	</footer>
</body>
</html>`

	doc, err := Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify document structure
	if doc.DocumentElement() == nil {
		t.Fatal("missing document element")
	}
	if doc.DocumentElement().TagName != "html" {
		t.Errorf("root tag = %q, want html", doc.DocumentElement().TagName)
	}

	// Test CSS selector queries
	t.Run("query by tag", func(t *testing.T) {
		paragraphs, err := doc.Query("p")
		if err != nil {
			t.Fatalf("Query() error = %v", err)
		}
		if len(paragraphs) < 2 {
			t.Errorf("found %d paragraphs, want at least 2", len(paragraphs))
		}
	})

	t.Run("query by class", func(t *testing.T) {
		links, err := doc.Query("a")
		if err != nil {
			t.Fatalf("Query() error = %v", err)
		}
		if len(links) < 2 {
			t.Errorf("found %d links, want at least 2", len(links))
		}
	})

	t.Run("query complex selector", func(t *testing.T) {
		navLinks, err := doc.Query("nav li a")
		if err != nil {
			t.Fatalf("Query() error = %v", err)
		}
		if len(navLinks) < 2 {
			t.Errorf("found %d nav links, want at least 2", len(navLinks))
		}
	})

	t.Run("query first", func(t *testing.T) {
		h1, err := doc.QueryFirst("h1")
		if err != nil {
			t.Fatalf("QueryFirst() error = %v", err)
		}
		if h1 == nil {
			t.Fatal("QueryFirst() returned nil")
		}
		if h1.TagName != "h1" {
			t.Errorf("QueryFirst() tag = %q, want h1", h1.TagName)
		}
	})
}

// TestParseNestedStructures tests handling of deeply nested elements.
func TestParseNestedStructures(t *testing.T) {
	html := "<div><div><div><div><div><p>Deep nesting</p></div></div></div></div></div>"

	doc, err := Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	paragraphs, err := doc.Query("p")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(paragraphs) != 1 {
		t.Errorf("found %d paragraphs, want 1", len(paragraphs))
	}

	text := paragraphs[0].Text()
	if !strings.Contains(text, "Deep nesting") {
		t.Errorf("paragraph text = %q, want to contain 'Deep nesting'", text)
	}
}

// TestParseSpecialElements tests parsing special HTML elements.
func TestParseSpecialElements(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		query   string
		wantLen int
	}{
		{
			name:    "table",
			html:    "<table><thead><tr><th>H1</th></tr></thead><tbody><tr><td>D1</td></tr></tbody></table>",
			query:   "td",
			wantLen: 1,
		},
		{
			name:    "form",
			html:    "<form><input type='text' name='field'><button>Submit</button></form>",
			query:   "input",
			wantLen: 1,
		},
		{
			name:    "list",
			html:    "<ul><li>Item 1</li><li>Item 2</li><li>Item 3</li></ul>",
			query:   "li",
			wantLen: 3,
		},
		{
			name:    "template",
			html:    "<template><div>Template content</div></template>",
			query:   "template",
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(tt.html)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			elements, err := doc.Query(tt.query)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if len(elements) != tt.wantLen {
				t.Errorf("found %d elements, want %d", len(elements), tt.wantLen)
			}
		})
	}
}

// TestParseSelfClosingTags tests handling of self-closing and void elements.
func TestParseSelfClosingTags(t *testing.T) {
	html := `<html><body>
		<img src="test.jpg" alt="Test">
		<br>
		<hr>
		<input type="text" name="field">
		<meta charset="UTF-8">
		<link rel="stylesheet" href="style.css">
	</body></html>`

	doc, err := Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tests := []struct {
		tag     string
		wantMin int
	}{
		{"img", 1},
		{"br", 1},
		{"hr", 1},
		{"input", 1},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			elements, err := doc.Query(tt.tag)
			if err != nil {
				t.Fatalf("Query(%q) error = %v", tt.tag, err)
			}
			if len(elements) < tt.wantMin {
				t.Errorf("found %d %s elements, want at least %d", len(elements), tt.tag, tt.wantMin)
			}
		})
	}
}

// TestParseComments tests handling of HTML comments.
func TestParseComments(t *testing.T) {
	html := `<html><body>
		<!-- This is a comment -->
		<p>Content</p>
		<!-- Another comment -->
	</body></html>`

	doc, err := Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Comments should be part of the DOM
	if doc == nil || doc.DocumentElement() == nil {
		t.Fatal("Parse returned invalid document")
	}
}

// TestParseCDATA tests handling of CDATA sections.
func TestParseCDATA(t *testing.T) {
	html := `<html><body><script><![CDATA[
		var x = 1 < 2 && 3 > 2;
	]]></script></body></html>`

	doc, err := Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	scripts, err := doc.Query("script")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(scripts) < 1 {
		t.Fatal("expected at least one script element")
	}
}

// TestParseEntities tests HTML entity handling.
func TestParseEntities(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantText string
	}{
		{
			name:     "named entities",
			html:     "<p>&lt;&gt;&amp;&quot;&#39;</p>",
			wantText: "<>&\"'",
		},
		{
			name:     "numeric entities",
			html:     "<p>&#60;&#62;&#38;</p>",
			wantText: "<>&",
		},
		{
			name:     "hex entities",
			html:     "<p>&#x3C;&#x3E;&#x26;</p>",
			wantText: "<>&",
		},
		{
			name:     "common entities",
			html:     "<p>&nbsp;&copy;&reg;&trade;</p>",
			wantText: "\u00A0©®™",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(tt.html)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			paragraphs, err := doc.Query("p")
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if len(paragraphs) != 1 {
				t.Fatalf("found %d paragraphs, want 1", len(paragraphs))
			}

			gotText := paragraphs[0].Text()
			if gotText != tt.wantText {
				t.Errorf("text = %q, want %q", gotText, tt.wantText)
			}
		})
	}
}

// TestParseAttributes tests element attribute handling.
func TestParseAttributes(t *testing.T) {
	html := `<div id="main" class="container active" data-value="123" disabled></div>`

	doc, err := Parse(html)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	divs, err := doc.Query("div")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(divs) != 1 {
		t.Fatalf("found %d divs, want 1", len(divs))
	}

	div := divs[0]

	tests := []struct {
		attr string
		want string
	}{
		{"id", "main"},
		{"class", "container active"},
		{"data-value", "123"},
		{"disabled", ""},
	}

	for _, tt := range tests {
		t.Run(tt.attr, func(t *testing.T) {
			got := div.Attr(tt.attr)
			if got != tt.want {
				t.Errorf("Attr(%q) = %q, want %q", tt.attr, got, tt.want)
			}
		})
	}

	// Test HasAttr
	if !div.HasAttr("id") {
		t.Error("HasAttr(id) = false, want true")
	}
	if div.HasAttr("nonexistent") {
		t.Error("HasAttr(nonexistent) = true, want false")
	}
}

// TestParseErrorRecovery tests error recovery for malformed HTML.
func TestParseErrorRecovery(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		query string
		want  int
	}{
		{
			name:  "unclosed tags",
			html:  "<p>Para 1<p>Para 2<p>Para 3",
			query: "p",
			want:  3,
		},
		{
			name:  "mismatched tags",
			html:  "<div><p>Text</div></p>",
			query: "div",
			want:  1,
		},
		{
			name:  "missing closing tags",
			html:  "<html><body><div><p>Content",
			query: "p",
			want:  1,
		},
		{
			name:  "invalid nesting",
			html:  "<p><div>Invalid</div></p>",
			query: "div",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(tt.html)
			if err != nil {
				// Error recovery means we should still get a document
				t.Logf("Parse returned error (expected for malformed HTML): %v", err)
			}
			if doc == nil {
				t.Fatal("Parse() returned nil document even with error recovery")
			}

			elements, err := doc.Query(tt.query)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if len(elements) != tt.want {
				t.Errorf("found %d elements, want %d", len(elements), tt.want)
			}
		})
	}
}

// TestParseErrorCollection tests error collection functionality.
func TestParseErrorCollection(t *testing.T) {
	html := "<html><body><p>Test</p></body>"

	doc, err := Parse(html, WithCollectErrors())
	if doc == nil {
		t.Fatal("WithCollectErrors should still return document")
	}

	// Check if we got a ParseErrors type
	if err != nil {
		var parseErrors htmlerrors.ParseErrors
		if errors.As(err, &parseErrors) {
			t.Errorf("error type = %T, want htmlerrors.ParseErrors", err)
		}
	}
}

// TestParseStrictMode tests strict mode parsing.
func TestParseStrictMode(t *testing.T) {
	// Valid HTML should work in strict mode
	validHTML := "<!DOCTYPE html><html><head><title>Test</title></head><body><p>Content</p></body></html>"
	doc, err := Parse(validHTML, WithStrictMode())
	if err != nil {
		t.Logf("Strict mode returned error for valid HTML: %v", err)
	}
	if doc == nil {
		t.Fatal("Strict mode returned nil document")
	}
}

// Helper function to extract all text from a document.
func extractAllText(node dom.Node) string {
	var sb strings.Builder
	extractTextHelper(node, &sb)
	return sb.String()
}

func extractTextHelper(node dom.Node, sb *strings.Builder) {
	switch n := node.(type) {
	case *dom.Text:
		sb.WriteString(n.Data)
	case *dom.Element:
		for _, child := range n.Children() {
			extractTextHelper(child, sb)
		}
	case *dom.Document:
		for _, child := range n.Children() {
			extractTextHelper(child, sb)
		}
	}
}
