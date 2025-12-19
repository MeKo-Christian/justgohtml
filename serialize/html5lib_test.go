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

	// Convert test options to SerializeTokenOptions
	opts := serialize.DefaultSerializeTokenOptions()

	// Apply test-specific options
	if test.Options != nil {
		if v, ok := test.Options["quote_char"].(string); ok && len(v) > 0 {
			opts.QuoteChar = rune(v[0])
		}
		if v, ok := test.Options["use_trailing_solidus"].(bool); ok {
			opts.UseTrailingSolidus = v
		}
		if v, ok := test.Options["minimize_boolean_attributes"].(bool); ok {
			opts.MinimizeBooleanAttributes = v
		}
		if v, ok := test.Options["quote_attr_values"].(bool); ok && v {
			// quote_attr_values=true means minimize boolean attrs (omit =value)
			opts.MinimizeBooleanAttributes = true
		}
		if v, ok := test.Options["escape_lt_in_attrs"].(bool); ok {
			opts.EscapeLtInAttrs = v
		}
		if v, ok := test.Options["escape_rcdata"].(bool); ok {
			opts.EscapeRcdata = v
		}
		if v, ok := test.Options["strip_whitespace"].(bool); ok {
			opts.StripWhitespace = v
		}
		if v, ok := test.Options["omit_optional_tags"].(bool); ok {
			opts.OmitOptionalTags = v
		}
		if v, ok := test.Options["inject_meta_charset"].(bool); ok {
			opts.InjectMetaCharset = v
			// inject_meta_charset implies omit_optional_tags
			opts.OmitOptionalTags = true
		}
		if v, ok := test.Options["encoding"].(string); ok {
			opts.Encoding = v
		}
	}

	// Serialize the token stream with options
	actual, err := serialize.SerializeTokensWithOptions(test.Input, opts)
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
