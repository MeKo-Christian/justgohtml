package tokenizer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MeKo-Christian/JustGoHTML/internal/testutil"
	"github.com/MeKo-Christian/JustGoHTML/tokenizer"
)

const (
	html5libTestsDir = "../testdata/html5lib-tests/tokenizer"
	JustHTMLTestsDir = "../testdata/justhtml-tests"
)

// TestJustGoHTMLTokenizer runs JustGoHTML-specific tokenizer tests.
func TestJustGoHTMLTokenizer(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(JustHTMLTestsDir); os.IsNotExist(err) {
		t.Skip("JustGoHTML-tests not found")
	}

	files, err := testutil.CollectTestFiles(JustHTMLTestsDir, "*.test")
	if err != nil {
		t.Fatalf("Failed to collect test files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No JustGoHTML tokenizer test files found")
	}

	for _, file := range files {
		// capture for parallel
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTokenizerTestFile(t, file)
		})
	}
}

func runTokenizerTestFile(t *testing.T, path string) {
	t.Helper()
	testFile, err := testutil.ParseTokenizerFile(path)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	tests := testFile.Tests
	xmlViolation := false
	if len(tests) == 0 {
		tests = testFile.XMLViolationTests
		xmlViolation = true
	}

	for i, test := range tests {
		testName := test.Description
		if testName == "" {
			testName = "test"
		}
		t.Run(testName, func(t *testing.T) {
			runSingleTokenizerTest(t, test, i, xmlViolation)
		})
	}
}

func runSingleTokenizerTest(t *testing.T, test testutil.TokenizerTest, _ int, xmlViolation bool) {
	t.Helper()
	input := test.Input
	expectedTokens := test.Output

	// Handle double-escaped tests
	if test.DoubleEscaped {
		input = testutil.UnescapeUnicode(input)
		expectedTokens = unescapeTokens(expectedTokens)
	}

	initialStates := test.InitialStates
	if len(initialStates) == 0 {
		initialStates = []string{"Data state"}
	}

	for _, stateName := range initialStates {
		state := mapInitialState(stateName)
		if state == tokenizer.InvalidState {
			t.Skipf("Unknown initial state: %s", stateName)
			continue
		}

		tok := tokenizer.New(input)
		tok.SetDiscardBOM(test.DiscardBOM)
		tok.SetXMLCoercion(xmlViolation)
		tok.SetState(state)
		if test.LastStartTag != "" {
			tok.SetLastStartTag(test.LastStartTag)
		}

		var actualTokens []interface{}
		for {
			token := tok.Next()
			if token.Type == tokenizer.EOF {
				break
			}
			actualTokens = append(actualTokens, tokenToTestFormat(token))
		}

		// Collapse consecutive character tokens
		actualTokens = collapseCharacterTokens(actualTokens)

		// Compare tokens
		if !compareTokens(expectedTokens, actualTokens) {
			t.Errorf("State %q:\nInput: %q\nExpected: %v\nActual: %v",
				stateName, input, formatExpected(expectedTokens), actualTokens)
		}
	}
}

// mapInitialState converts html5lib state name to our State type.
func mapInitialState(name string) tokenizer.State {
	switch name {
	case "Data state":
		return tokenizer.DataState
	case "PLAINTEXT state":
		return tokenizer.PlaintextState
	case "RCDATA state":
		return tokenizer.RCDATAState
	case "RAWTEXT state":
		return tokenizer.RawtextState
	case "Script data state":
		return tokenizer.ScriptDataState
	case "CDATA section state":
		return tokenizer.CDATASectionState
	default:
		return tokenizer.InvalidState
	}
}

// tokenToTestFormat converts a Token to the html5lib test format.
func tokenToTestFormat(token *tokenizer.Token) interface{} {
	switch token.Type {
	case tokenizer.DOCTYPE:
		var name interface{} = token.Name
		if token.Name == "" {
			name = nil
		}
		// ["DOCTYPE", name, publicId, systemId, correctness]
		return []interface{}{
			"DOCTYPE",
			name,
			token.PublicID,
			token.SystemID,
			!token.ForceQuirks,
		}
	case tokenizer.StartTag:
		result := []interface{}{"StartTag", token.Name, tokenizer.AttrsToMap(token.Attrs)}
		if token.SelfClosing {
			result = append(result, true)
		}
		return result
	case tokenizer.EndTag:
		return []interface{}{"EndTag", token.Name}
	case tokenizer.Comment:
		return []interface{}{"Comment", token.Data}
	case tokenizer.Character:
		return []interface{}{"Character", token.Data}
	case tokenizer.Error, tokenizer.EOF:
		return nil
	}
	return nil
}

// collapseCharacterTokens merges consecutive character tokens.
func collapseCharacterTokens(tokens []interface{}) []interface{} {
	result := make([]interface{}, 0, len(tokens))
	for _, tok := range tokens {
		arr, ok := tok.([]interface{})
		if !ok || len(arr) == 0 {
			result = append(result, tok)
			continue
		}

		if arr[0] == "Character" && len(result) > 0 {
			lastArr, ok := result[len(result)-1].([]interface{})
			if ok && len(lastArr) >= 2 && lastArr[0] == "Character" {
				// Merge with previous character token
				lastArr[1] = lastArr[1].(string) + arr[1].(string)
				continue
			}
		}
		result = append(result, tok)
	}
	return result
}

// compareTokens compares expected (json.RawMessage) with actual tokens.
func compareTokens(expected []json.RawMessage, actual []interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		var exp interface{}
		if err := json.Unmarshal(expected[i], &exp); err != nil {
			return false
		}

		if !deepEqual(exp, actual[i]) {
			return false
		}
	}
	return true
}

// deepEqual compares two interface{} values for equality.
func deepEqual(a, b interface{}) bool {
	// Normalize both to JSON and compare
	aJSON, errA := json.Marshal(a)
	bJSON, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(aJSON) == string(bJSON)
}

// formatExpected formats expected tokens for error output.
func formatExpected(tokens []json.RawMessage) []interface{} {
	result := make([]interface{}, 0, len(tokens))
	for _, raw := range tokens {
		var v interface{}
		_ = json.Unmarshal(raw, &v)
		result = append(result, v)
	}
	return result
}

// unescapeTokens handles double-escaped test data.
func unescapeTokens(tokens []json.RawMessage) []json.RawMessage {
	result := make([]json.RawMessage, 0, len(tokens))
	for _, raw := range tokens {
		result = append(result, json.RawMessage(undoubleEscapeUnicodeInJSON(raw)))
	}
	return result
}

func undoubleEscapeUnicodeInJSON(raw []byte) []byte {
	isHex := func(b byte) bool {
		return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
	}

	out := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		// Replace `\\uXXXX` (6 bytes) with `\uXXXX` (5 bytes) to undo one escaping layer.
		if raw[i] == '\\' && i+6 < len(raw) && raw[i+1] == '\\' && raw[i+2] == 'u' &&
			isHex(raw[i+3]) && isHex(raw[i+4]) && isHex(raw[i+5]) && isHex(raw[i+6]) {
			out = append(out, '\\', 'u', raw[i+3], raw[i+4], raw[i+5], raw[i+6])
			i += 6
			continue
		}
		out = append(out, raw[i])
	}
	return out
}

// BenchmarkTokenizer benchmarks the tokenizer on a simple document.
func BenchmarkTokenizer(b *testing.B) {
	html := strings.Repeat("<div class='test'>Hello, <b>world</b>!</div>", 1000)

	b.ResetTimer()
	for range b.N {
		tok := tokenizer.New(html)
		for {
			token := tok.Next()
			if token.Type == tokenizer.EOF {
				break
			}
		}
	}
}
