package serialize_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
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

func runSingleSerializerTest(t *testing.T, _ testutil.SerializerTest, _ int) {
	t.Helper()
	// TODO: Implement serializer test
	// For now, we just skip since we need to build DOM from token stream
	t.Skip("Serializer test infrastructure not yet implemented")
}
