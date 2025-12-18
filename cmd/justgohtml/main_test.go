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
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("command failed: %v", err)
			}

			got := string(output)
			if !strings.Contains(got, "justhtml version") {
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
// Note: This test is skipped until the parser is implemented.
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
	// Currently the parser returns "not implemented", so the CLI fails
	// Once the parser is implemented, this test should pass
	if err != nil {
		if strings.Contains(string(output), "not implemented") {
			t.Skip("parser not implemented yet")
		}
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	// The CLI should succeed (exit 0) when parsing valid HTML
	got := string(output)
	if got == "" {
		t.Error("expected some output from parsing")
	}
}

// TestParseStdin tests parsing HTML from stdin.
// Note: This test is skipped until the parser is implemented.
func TestParseStdin(t *testing.T) {
	binary := buildTestBinary(t)

	htmlContent := `<!DOCTYPE html><html><body><p>From stdin</p></body></html>`

	cmd := exec.Command(binary, "-")
	cmd.Stdin = strings.NewReader(htmlContent)

	output, err := cmd.CombinedOutput()
	// Currently the parser returns "not implemented", so the CLI fails
	// Once the parser is implemented, this test should pass
	if err != nil {
		if strings.Contains(string(output), "not implemented") {
			t.Skip("parser not implemented yet")
		}
		t.Fatalf("command failed: %v\noutput: %s", err, output)
	}

	// The CLI should succeed when reading from stdin
	got := string(output)
	if got == "" {
		t.Error("expected some output from parsing stdin")
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
}

// buildTestBinary compiles the CLI binary for testing and returns its path.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Build in a temp directory
	tmpDir := t.TempDir()
	binary := filepath.Join(tmpDir, "justhtml")

	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = filepath.Dir(mustFindGoMod(t))
	cmd.Dir = filepath.Join(cmd.Dir, "cmd", "justhtml")

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
