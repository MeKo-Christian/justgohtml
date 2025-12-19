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
