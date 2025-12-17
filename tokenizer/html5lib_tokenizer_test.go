package tokenizer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
)

// TestHTML5LibTokenizer runs all html5lib tokenizer tests.
func TestHTML5LibTokenizer(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(html5libTestsDir); os.IsNotExist(err) {
		t.Skip("html5lib-tests not found - run 'git submodule update --init'")
	}

	files, err := testutil.CollectTestFiles(html5libTestsDir, "*.test")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No tokenizer test files found")
	}

	for _, file := range files {
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTokenizerTestFile(t, file)
		})
	}
}
