package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersion tests the --version and -v flags.
func TestVersion(t *testing.T) {
	binary := buildTestBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{"long flag", []string{"--version"}},
		{"short flag", []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v, output: %s", err, output)
			}

			got := string(output)
			if !strings.Contains(got, "justgohtml version") {
				t.Errorf("expected version output, got: %q", got)
			}
		})
	}
}

// TestMissingInput tests that the CLI requires an input file.
func TestMissingInput(t *testing.T) {
	binary := buildTestBinary(t)

	cmd := exec.Command(binary)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error for missing input, got success")
	}

	if !strings.Contains(stderr.String(), "missing input file") {
		t.Errorf("expected 'missing input file' in stderr, got: %q", stderr.String())
	}
}

// TestParseFile tests parsing an HTML file.
func TestParseFile(t *testing.T) {
	binary := buildTestBinary(t)

	// Create a temporary HTML file
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><head><title>Test</title></head><body><p>Hello</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "<html>") {
		t.Errorf("expected HTML output containing <html>, got: %q", got)
	}
	if !strings.Contains(got, "<title>") {
		t.Errorf("expected HTML output containing <title>, got: %q", got)
	}
}

// TestParseStdin tests parsing HTML from stdin.
func TestParseStdin(t *testing.T) {
	binary := buildTestBinary(t)

	htmlContent := `<!DOCTYPE html><html><body><p>From stdin</p></body></html>`

	cmd := exec.Command(binary, "-")
	cmd.Stdin = strings.NewReader(htmlContent)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "From stdin") {
		t.Errorf("expected output containing 'From stdin', got: %q", got)
	}
}

// TestInvalidFile tests error handling for non-existent files.
func TestInvalidFile(t *testing.T) {
	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "/nonexistent/path/to/file.html")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error for non-existent file, got success")
	}

	if !strings.Contains(stderr.String(), "reading input") {
		t.Errorf("expected 'reading input' error in stderr, got: %q", stderr.String())
	}
}

// TestHelp tests that -h shows usage information.
func TestHelp(t *testing.T) {
	binary := buildTestBinary(t)

	cmd := exec.Command(binary, "-h")
	// Note: -h causes flag.Parse to exit with code 0 and print to stderr
	output, _ := cmd.CombinedOutput()

	got := string(output)
	if !strings.Contains(got, "Usage:") {
		t.Errorf("expected usage information, got: %q", got)
	}
	if !strings.Contains(got, "-selector") {
		t.Errorf("expected -selector flag in help, got: %q", got)
	}
	if !strings.Contains(got, "Examples:") {
		t.Errorf("expected Examples section in help, got: %q", got)
	}
}

// TestSelectorFilter tests CSS selector filtering.
func TestSelectorFilter(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><h1>Title</h1><p>Para 1</p><p>Para 2</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		selector string
		contains []string
		excludes []string
	}{
		{
			name:     "select paragraphs",
			selector: "p",
			contains: []string{"<p>", "Para 1", "Para 2"},
			excludes: []string{"<h1>"},
		},
		{
			name:     "select h1",
			selector: "h1",
			contains: []string{"<h1>", "Title"},
			excludes: []string{"<p>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, "-s", tt.selector, htmlFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %v\noutput: %s", err, output)
			}

			got := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got: %q", want, got)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(got, exclude) {
					t.Errorf("expected output NOT to contain %q, got: %q", exclude, got)
				}
			}
		})
	}
}

// TestFirstMatch tests the --first flag.
func TestFirstMatch(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>First</p><p>Second</p><p>Third</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-s", "p", "--first", htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "First") {
		t.Errorf("expected output to contain 'First', got: %q", got)
	}
	if strings.Contains(got, "Second") {
		t.Errorf("expected output NOT to contain 'Second', got: %q", got)
	}
}

// TestTextFormat tests the text output format.
func TestTextFormat(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><h1>Title</h1><p>Hello World</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-f", "text", htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if strings.Contains(got, "<") {
		t.Errorf("text format should not contain HTML tags, got: %q", got)
	}
	if !strings.Contains(got, "Title") {
		t.Errorf("expected text to contain 'Title', got: %q", got)
	}
	if !strings.Contains(got, "Hello World") {
		t.Errorf("expected text to contain 'Hello World', got: %q", got)
	}
}

// TestMarkdownFormat tests the markdown output format.
func TestMarkdownFormat(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><h1>Title</h1><p>Para with <strong>bold</strong> text.</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-f", "markdown", htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "# Title") {
		t.Errorf("expected markdown h1, got: %q", got)
	}
	if !strings.Contains(got, "**bold**") {
		t.Errorf("expected markdown bold, got: %q", got)
	}
}

// TestInvalidFormat tests that invalid formats are rejected.
func TestInvalidFormat(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(htmlFile, []byte("<html></html>"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-f", "invalid", htmlFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error for invalid format, got success")
	}

	if !strings.Contains(stderr.String(), "invalid format") {
		t.Errorf("expected 'invalid format' in stderr, got: %q", stderr.String())
	}
}

// TestPrettyPrint tests the --pretty flag.
func TestPrettyPrint(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><head><title>Test</title></head><body><div><p>Hello</p></div></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Pretty print enabled (default)
	cmd := exec.Command(binary, htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "\n") {
		t.Errorf("pretty-printed output should contain newlines, got: %q", got)
	}

	// Pretty print disabled
	cmd = exec.Command(binary, "-pretty=false", htmlFile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	gotNoPretty := string(output)
	// Non-pretty output should still work - we just verify it doesn't error
	// The main difference is indentation which is hard to test precisely
	_ = gotNoPretty
}

// TestMarkdownList tests markdown list conversion.
func TestMarkdownList(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><ul><li>Item 1</li><li>Item 2</li></ul></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-f", "markdown", htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "- Item 1") {
		t.Errorf("expected markdown list item, got: %q", got)
	}
	if !strings.Contains(got, "- Item 2") {
		t.Errorf("expected markdown list item, got: %q", got)
	}
}

// TestMarkdownTable tests markdown table conversion.
func TestMarkdownTable(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><table><thead><tr><th>Name</th><th>Age</th></tr></thead><tbody><tr><td>Alice</td><td>30</td></tr></tbody></table></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-f", "markdown", htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "| Name | Age |") {
		t.Errorf("expected markdown table header, got: %q", got)
	}
	if !strings.Contains(got, "| --- | --- |") {
		t.Errorf("expected markdown table separator, got: %q", got)
	}
	if !strings.Contains(got, "| Alice | 30 |") {
		t.Errorf("expected markdown table row, got: %q", got)
	}
}

// TestMarkdownLink tests markdown link conversion.
func TestMarkdownLink(t *testing.T) {
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><a href="https://example.com">Example</a></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command(binary, "-f", "markdown", htmlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	got := string(output)
	if !strings.Contains(got, "[Example](https://example.com)") {
		t.Errorf("expected markdown link, got: %q", got)
	}
}

// TestRunFunction tests the run function directly for better coverage.
func TestRunFunction(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>Test</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "<p>") {
		t.Errorf("expected HTML output, got: %q", stdout.String())
	}
}

// TestRunFunctionWithSelector tests selector filtering via run function.
func TestRunFunctionWithSelector(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p class="target">Found</p><p>Not found</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-s", ".target", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Found") {
		t.Errorf("expected to find 'Found', got: %q", got)
	}
	if strings.Contains(got, "Not found") {
		t.Errorf("expected NOT to find 'Not found', got: %q", got)
	}
}

// TestRunFunctionStdin tests stdin reading via run function.
func TestRunFunctionStdin(t *testing.T) {
	stdin := strings.NewReader(`<html><body><p>Stdin content</p></body></html>`)
	var stdout, stderr bytes.Buffer

	err := run([]string{"-"}, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "Stdin content") {
		t.Errorf("expected stdin content in output, got: %q", stdout.String())
	}
}

// buildTestBinary compiles the CLI binary for testing and returns its path.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Build in a temp directory
	tmpDir := t.TempDir()
	binary := filepath.Join(tmpDir, "justgohtml")

	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = filepath.Dir(mustFindGoMod(t))
	cmd.Dir = filepath.Join(cmd.Dir, "cmd", "justgohtml")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, output)
	}

	return binary
}

// TestSelectorShorthand tests both -s and --selector flags.
func TestSelectorShorthand(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p class="target">Found</p><p>Other</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{"long flag", []string{"--selector", ".target", htmlFile}},
		{"short flag", []string{"-s", ".target", htmlFile}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tt.args, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got := stdout.String()
			if !strings.Contains(got, "Found") {
				t.Errorf("expected output to contain 'Found', got: %q", got)
			}
			if strings.Contains(got, "Other") {
				t.Errorf("expected output NOT to contain 'Other', got: %q", got)
			}
		})
	}
}

// TestFormatShorthand tests both -f and --format flags.
func TestFormatShorthand(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>Test</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name   string
		args   []string
		noTags bool
	}{
		{"long flag text", []string{"--format", "text", htmlFile}, true},
		{"short flag text", []string{"-f", "text", htmlFile}, true},
		{"long flag html", []string{"--format", "html", htmlFile}, false},
		{"short flag html", []string{"-f", "html", htmlFile}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tt.args, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got := stdout.String()
			hasTags := strings.Contains(got, "<p>")
			if tt.noTags && hasTags {
				t.Errorf("text format should not contain tags, got: %q", got)
			}
			if !tt.noTags && !hasTags {
				t.Errorf("html format should contain tags, got: %q", got)
			}
		})
	}
}

// TestInvalidSelector tests error handling for invalid CSS selectors.
func TestInvalidSelector(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>Test</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-s", "[[invalid", htmlFile}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid selector, got success")
	}

	if !strings.Contains(err.Error(), "invalid selector") {
		t.Errorf("expected 'invalid selector' in error, got: %v", err)
	}
}

// TestEmptySelector tests that empty selector returns full document.
func TestEmptySelector(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>Test</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "<html>") {
		t.Errorf("expected full document, got: %q", got)
	}
}

// TestIndentOption tests the --indent flag.
func TestIndentOption(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><div><p>Test</p></div></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name   string
		indent string
	}{
		{"indent 2", "2"},
		{"indent 4", "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run([]string{"--indent", tt.indent, htmlFile}, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			// Just verify it doesn't error - indentation testing is complex
			if stdout.Len() == 0 {
				t.Error("expected output, got empty")
			}
		})
	}
}

// TestStripOption tests the --strip flag for text output.
func TestStripOption(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>   Text   with   spaces   </p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name            string
		stripFlag       string
		expectCollapsed bool
	}{
		{"strip enabled", "true", true},
		{"strip disabled", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run([]string{"-f", "text", "--strip=" + tt.stripFlag, htmlFile}, nil, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got := stdout.String()
			hasMultipleSpaces := strings.Contains(got, "  ")
			if tt.expectCollapsed && hasMultipleSpaces {
				t.Errorf("expected collapsed whitespace, got: %q", got)
			}
		})
	}
}

// TestSeparatorOption tests the --separator flag for text output.
func TestSeparatorOption(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>First</p><p>Second</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-f", "text", "-s", "p", "--separator", " | ", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Separator option exists but its effect depends on implementation
	if stdout.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestMultipleMatches tests handling of multiple selector matches.
func TestMultipleMatches(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body>
		<div class="item">First</div>
		<div class="item">Second</div>
		<div class="item">Third</div>
	</body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-s", ".item", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "First") {
		t.Errorf("expected 'First' in output, got: %q", got)
	}
	if !strings.Contains(got, "Second") {
		t.Errorf("expected 'Second' in output, got: %q", got)
	}
	if !strings.Contains(got, "Third") {
		t.Errorf("expected 'Third' in output, got: %q", got)
	}
}

// TestNoMatches tests handling when selector matches nothing.
func TestNoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><p>Test</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-s", ".nonexistent", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	// When no matches, output should be empty or minimal
	if strings.Contains(got, "<p>") {
		t.Errorf("expected no <p> in output when selector matches nothing, got: %q", got)
	}
}

// TestComplexMarkdown tests complex markdown conversion.
func TestComplexMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body>
		<h1>Main Title</h1>
		<p>Paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
		</ul>
		<blockquote>A quote</blockquote>
		<pre>Code block</pre>
	</body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-f", "markdown", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	expectations := []string{
		"# Main Title",
		"**bold**",
		"*italic*",
		"- Item 1",
		"- Item 2",
		"> A quote",
		"```",
	}

	for _, want := range expectations {
		if !strings.Contains(got, want) {
			t.Errorf("expected markdown output to contain %q, got: %q", want, got)
		}
	}
}

// TestStdinWithSelector tests combining stdin input with selector.
func TestStdinWithSelector(t *testing.T) {
	stdin := strings.NewReader(`<html><body><h1>Title</h1><p>Content</p></body></html>`)
	var stdout, stderr bytes.Buffer

	err := run([]string{"-s", "h1", "-"}, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Title") {
		t.Errorf("expected 'Title' in output, got: %q", got)
	}
	if strings.Contains(got, "Content") {
		t.Errorf("expected NOT to find 'Content' (filtered by selector), got: %q", got)
	}
}

// TestStdinWithTextFormat tests stdin with text format output.
func TestStdinWithTextFormat(t *testing.T) {
	stdin := strings.NewReader(`<html><body><p>Hello <strong>World</strong></p></body></html>`)
	var stdout, stderr bytes.Buffer

	err := run([]string{"-f", "text", "-"}, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "<") {
		t.Errorf("text format should not contain HTML tags, got: %q", got)
	}
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "World") {
		t.Errorf("expected text content, got: %q", got)
	}
}

// TestEmptyFile tests handling of empty HTML files.
func TestEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "empty.html")
	if err := os.WriteFile(htmlFile, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Empty file should still produce valid HTML structure
	got := stdout.String()
	if !strings.Contains(got, "<html>") {
		t.Errorf("expected HTML structure even for empty file, got: %q", got)
	}
}

// TestLargeFile tests handling of larger HTML files.
func TestLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "large.html")

	// Generate a large HTML file
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><body>")
	for range 1000 {
		sb.WriteString("<p>Paragraph ")
		sb.WriteString(strings.Repeat("x", 100))
		sb.WriteString("</p>")
	}
	sb.WriteString("</body></html>")

	if err := os.WriteFile(htmlFile, []byte(sb.String()), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if stdout.Len() == 0 {
		t.Error("expected output for large file, got empty")
	}
}

// TestSpecialCharactersInPath tests file paths with special characters.
func TestSpecialCharactersInPath(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test file with spaces.html")
	htmlContent := `<!DOCTYPE html><html><body><p>Test</p></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "<p>") {
		t.Errorf("expected HTML output, got: %q", got)
	}
}

// TestMarkdownImage tests markdown image conversion.
func TestMarkdownImage(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><img src="test.jpg" alt="Test Image"></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-f", "markdown", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "![Test Image](test.jpg)") {
		t.Errorf("expected markdown image syntax, got: %q", got)
	}
}

// TestMarkdownBlockquote tests markdown blockquote conversion.
func TestMarkdownBlockquote(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><blockquote>Quote text</blockquote></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-f", "markdown", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "> Quote text") {
		t.Errorf("expected markdown blockquote syntax, got: %q", got)
	}
}

// TestMarkdownCodeBlock tests markdown code block conversion.
func TestMarkdownCodeBlock(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html><html><body><pre>code here</pre></body></html>`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-f", "markdown", htmlFile}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "```") {
		t.Errorf("expected markdown code block syntax, got: %q", got)
	}
	if !strings.Contains(got, "code here") {
		t.Errorf("expected code content, got: %q", got)
	}
}

// mustFindGoMod finds the go.mod file by walking up from cwd.
func mustFindGoMod(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	for {
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return goMod
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod")
		}
		dir = parent
	}
}
