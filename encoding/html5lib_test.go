package encoding_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/encoding"
	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
)

const html5libTestsDir = "../testdata/html5lib-tests/encoding"

// TestHTML5LibEncoding runs all html5lib encoding tests.
func TestHTML5LibEncoding(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(html5libTestsDir); os.IsNotExist(err) {
		t.Skip("html5lib-tests not found - run 'git submodule update --init'")
	}

	files, err := testutil.CollectTestFiles(html5libTestsDir, "*.dat")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No encoding test files found")
	}

	for _, file := range files {
		// Skip "scripted" tests - these test JavaScript-generated meta tags
		// which aren't part of the byte-level prescan algorithm
		if strings.Contains(file, "/scripted/") {
			continue
		}
		// capture for parallel
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runEncodingTestFile(t, file)
		})
	}
}

func runEncodingTestFile(t *testing.T, path string) {
	t.Helper()
	tests, err := testutil.ParseEncodingFile(path)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	for i, test := range tests {
		testName := test.ExpectedEncoding
		if testName == "" {
			testName = "empty"
		}
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			runSingleEncodingTest(t, test, i)
		})
	}
}

func runSingleEncodingTest(t *testing.T, test testutil.EncodingTest, _ int) {
	t.Helper()
	_, enc, err := encoding.Decode(test.Data, "")
	if err != nil {
		t.Errorf("Decode error: %v", err)
		return
	}

	actualEncoding := ""
	if enc != nil {
		actualEncoding = enc.Name
	}

	// Normalize encoding names for comparison
	expected := normalizeEncodingName(test.ExpectedEncoding)
	actual := normalizeEncodingName(actualEncoding)

	if expected != actual {
		t.Errorf("Encoding mismatch:\nExpected: %s\nActual: %s\nInput (first 100 bytes): %q",
			test.ExpectedEncoding, actualEncoding, truncateBytes(test.Data, 100))
	}
}

func normalizeEncodingName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// Map common aliases to canonical names
	switch name {
	case encWindows1252, "cp1252", "x-cp1252":
		return encWindows1252
	case "iso-8859-1", "iso8859-1", "latin1":
		return "iso-8859-1"
	case "utf-8", "utf8":
		return "utf-8"
	}
	return name
}

func truncateBytes(b []byte, maxLen int) []byte {
	if len(b) <= maxLen {
		return b
	}
	return b[:maxLen]
}
