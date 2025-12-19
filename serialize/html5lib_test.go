package serialize_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
	"github.com/MeKo-Christian/JustGoHTML/serialize"
)

const html5libTestsDir = "../testdata/html5lib-tests/serializer"

// TestHTML5LibSerializer runs all html5lib serializer tests.
func TestHTML5LibSerializer(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(html5libTestsDir); os.IsNotExist(err) {
		t.Skip("html5lib-tests not found - run 'git submodule update --init'")
	}

	files, err := testutil.CollectTestFiles(html5libTestsDir, "*.test")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No serializer test files found")
	}

	for _, file := range files {
		// capture for parallel
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runSerializerTestFile(t, file)
		})
	}
}

func runSerializerTestFile(t *testing.T, path string) {
	t.Helper()
	testFile, err := testutil.ParseSerializerFile(path)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	for i, test := range testFile.Tests {
		testName := test.Description
		if testName == "" {
			testName = "test"
		}
		t.Run(testName, func(t *testing.T) {
			runSingleSerializerTest(t, test, i)
		})
	}
}

func runSingleSerializerTest(t *testing.T, test testutil.SerializerTest, _ int) {
	t.Helper()

	// Skip tests that only have XHTML expected output (no HTML5 expected)
	if len(test.Expected) == 0 {
		t.Skip("No expected output")
		return
	}

	// Serialize the token stream
	actual, err := serialize.SerializeTokens(test.Input)
	if err != nil {
		t.Fatalf("Serialization error: %v", err)
	}

	// Check if actual matches any of the expected outputs
	for _, expected := range test.Expected {
		if actual == expected {
			return // Success!
		}
	}

	// No match found - report failure
	t.Errorf("Serialization mismatch\nExpected: %q\nActual:   %q", test.Expected[0], actual)
}
