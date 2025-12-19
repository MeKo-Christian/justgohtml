// Package testutil provides utilities for running html5lib-tests.
package testutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Test file section markers.
const (
	sectionData     = "data"
	sectionErrors   = "errors"
	sectionDocument = "document"
	sectionFragment = "fragment"
	sectionEncoding = "encoding"
)

// Precompile to avoid regex construction on every sort comparison.
var naturalNumberRe = regexp.MustCompile(`(\d+)`)

// TreeConstructionTest represents a single tree-construction test case.
type TreeConstructionTest struct {
	Data            string
	Errors          []string
	Document        string
	FragmentContext string // e.g., "div" or "svg path"
	ScriptDirective string // "script-on" or "script-off"
	IframeSrcdoc    bool
	XMLCoercion     bool
}

// TokenizerTestFile represents a tokenizer test file (JSON format).
type TokenizerTestFile struct {
	Tests             []TokenizerTest `json:"tests"`
	XMLViolationTests []TokenizerTest `json:"xmlViolationTests"`
}

// TokenizerTest represents a single tokenizer test case.
type TokenizerTest struct {
	Description   string            `json:"description"`
	Input         string            `json:"input"`
	Output        []json.RawMessage `json:"output"`
	Errors        []TokenizerError  `json:"errors"`
	InitialStates []string          `json:"initialStates"`
	LastStartTag  string            `json:"lastStartTag"`
	DoubleEscaped bool              `json:"doubleEscaped"`
	DiscardBOM    bool              `json:"discardBom"`
}

// TokenizerError represents a tokenizer error in the test format.
type TokenizerError struct {
	Code   string `json:"code"`
	Line   int    `json:"line"`
	Column int    `json:"col"`
}

// SerializerTestFile represents a serializer test file (JSON format).
type SerializerTestFile struct {
	Tests []SerializerTest `json:"tests"`
}

// SerializerTest represents a single serializer test case.
type SerializerTest struct {
	Description string                 `json:"description"`
	Input       []json.RawMessage      `json:"input"`
	Expected    []string               `json:"expected"`
	Options     map[string]interface{} `json:"options"`
}

// EncodingTest represents a single encoding test case.
type EncodingTest struct {
	Data             []byte
	ExpectedEncoding string
}

// ParseTreeConstructionFile parses a .dat file containing tree-construction tests.
func ParseTreeConstructionFile(path string) ([]TreeConstructionTest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tests []TreeConstructionTest
	scanner := bufio.NewScanner(file)

	var currentTest *TreeConstructionTest
	var mode string
	var dataLines, errorLines, documentLines []string

	flush := func() {
		if currentTest != nil && (len(dataLines) > 0 || len(documentLines) > 0) {
			currentTest.Data = decodeEscapes(strings.Join(dataLines, "\n"))
			currentTest.Errors = errorLines
			currentTest.Document = strings.Join(documentLines, "\n")
			tests = append(tests, *currentTest)
		}
		currentTest = &TreeConstructionTest{}
		dataLines = nil
		errorLines = nil
		documentLines = nil
		mode = ""
	}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		if strings.HasPrefix(line, "#") {
			directive := strings.TrimPrefix(line, "#")
			switch directive {
			case sectionData:
				flush()
				mode = sectionData
			case sectionErrors:
				mode = sectionErrors
			case sectionDocument:
				mode = sectionDocument
			case "document-fragment":
				mode = sectionFragment
			case "script-on", "script-off":
				if currentTest != nil {
					currentTest.ScriptDirective = directive
				}
			case "iframe-srcdoc":
				if currentTest != nil {
					currentTest.IframeSrcdoc = true
				}
			case "xml-coercion":
				if currentTest != nil {
					currentTest.XMLCoercion = true
				}
			default:
				mode = directive
			}
			continue
		}

		switch mode {
		case sectionData:
			dataLines = append(dataLines, line)
		case sectionErrors:
			if strings.TrimSpace(line) != "" {
				errorLines = append(errorLines, line)
			}
		case sectionDocument:
			documentLines = append(documentLines, line)
		case sectionFragment:
			if currentTest != nil && strings.TrimSpace(line) != "" {
				currentTest.FragmentContext = strings.TrimSpace(line)
			}
		}
	}

	flush() // Final test

	return tests, scanner.Err()
}

func decodeEscapes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		switch s[i+1] {
		case '\\':
			b.WriteByte('\\')
			i++
			continue
		case 'n':
			b.WriteByte('\n')
			i++
			continue
		case 't':
			b.WriteByte('\t')
			i++
			continue
		case 'f':
			b.WriteByte('\f')
			i++
			continue
		case 'r':
			b.WriteByte('\r')
			i++
			continue
		case 'x':
			if i+3 < len(s) {
				if v, ok := parseHexByte(s[i+2 : i+4]); ok {
					b.WriteByte(v)
					i += 3
					continue
				}
			}
		case 'u':
			if i+5 < len(s) {
				if r, ok := parseHexRune(s[i+2 : i+6]); ok {
					b.WriteRune(r)
					i += 5
					continue
				}
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func parseHexByte(s string) (byte, bool) {
	var v byte
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v = v<<4 + c - '0'
		case c >= 'a' && c <= 'f':
			v = v<<4 + 10 + c - 'a'
		case c >= 'A' && c <= 'F':
			v = v<<4 + 10 + c - 'A'
		default:
			return 0, false
		}
	}
	return v, true
}

func parseHexRune(s string) (rune, bool) {
	var v rune
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v = v<<4 + rune(c-'0')
		case c >= 'a' && c <= 'f':
			v = v<<4 + rune(10+c-'a')
		case c >= 'A' && c <= 'F':
			v = v<<4 + rune(10+c-'A')
		default:
			return 0, false
		}
	}
	return v, true
}

// ParseTokenizerFile parses a .test file containing tokenizer tests (JSON format).
func ParseTokenizerFile(path string) (*TokenizerTestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var testFile TokenizerTestFile
	if err := json.Unmarshal(data, &testFile); err != nil {
		return nil, err
	}

	return &testFile, nil
}

// ParseSerializerFile parses a .test file containing serializer tests (JSON format).
func ParseSerializerFile(path string) (*SerializerTestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var testFile SerializerTestFile
	if err := json.Unmarshal(data, &testFile); err != nil {
		return nil, err
	}

	return &testFile, nil
}

// ParseEncodingFile parses a .dat file containing encoding tests.
func ParseEncodingFile(path string) ([]EncodingTest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tests []EncodingTest
	var currentData []byte
	var currentEncoding string
	mode := ""

	flush := func() {
		if currentData != nil && currentEncoding != "" {
			tests = append(tests, EncodingTest{
				Data:             currentData,
				ExpectedEncoding: currentEncoding,
			})
		}
		currentData = nil
		currentEncoding = ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")

		if trimmed == "#"+sectionData {
			flush()
			mode = sectionData
			continue
		}
		if trimmed == "#"+sectionEncoding {
			mode = sectionEncoding
			continue
		}

		switch mode {
		case sectionData:
			currentData = append(currentData, []byte(line+"\n")...)
		case sectionEncoding:
			if currentEncoding == "" && strings.TrimSpace(trimmed) != "" {
				currentEncoding = strings.TrimSpace(trimmed)
			}
		}
	}

	flush()

	return tests, nil
}

// CollectTestFiles returns all test files matching the given pattern in directory.
func CollectTestFiles(dir, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return err
		}
		if matched {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return naturalLess(filepath.Base(files[i]), filepath.Base(files[j]))
	})

	return files, nil
}

// naturalLess compares strings with natural number ordering.
func naturalLess(a, b string) bool {
	partsA := naturalNumberRe.Split(a, -1)
	numsA := naturalNumberRe.FindAllString(a, -1)
	partsB := naturalNumberRe.Split(b, -1)
	numsB := naturalNumberRe.FindAllString(b, -1)

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := range maxLen {
		var pa, pb string
		if i < len(partsA) {
			pa = partsA[i]
		}
		if i < len(partsB) {
			pb = partsB[i]
		}

		if pa != pb {
			return pa < pb
		}

		var na, nb string
		if i < len(numsA) {
			na = numsA[i]
		}
		if i < len(numsB) {
			nb = numsB[i]
		}

		if na != nb {
			numA, _ := strconv.Atoi(na)
			numB, _ := strconv.Atoi(nb)
			return numA < numB
		}
	}

	return a < b
}

// UnescapeUnicode unescapes \uXXXX sequences in test data.
//
// html5lib "doubleEscaped" inputs encode Unicode using JSON-style \uXXXX
// sequences, including surrogate pairs for astral code points. Go's `encoding/json`
// rejects surrogate code points when represented directly as UTF-8, so we must
// merge surrogate pairs into a single rune.
func UnescapeUnicode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+5 >= len(s) || s[i+1] != 'u' {
			b.WriteByte(s[i])
			continue
		}

		hex := s[i+2 : i+6]
		v, err := strconv.ParseUint(hex, 16, 16)
		if err != nil {
			b.WriteByte(s[i])
			continue
		}
		r := rune(v)

		// Surrogate pair: \uD800-\uDBFF followed by \uDC00-\uDFFF.
		if r >= 0xD800 && r <= 0xDBFF && i+11 < len(s) && s[i+6] == '\\' && s[i+7] == 'u' {
			hex2 := s[i+8 : i+12]
			v2, err2 := strconv.ParseUint(hex2, 16, 16)
			if err2 == nil {
				lo := rune(v2)
				if lo >= 0xDC00 && lo <= 0xDFFF {
					codePoint := 0x10000 + ((r - 0xD800) << 10) + (lo - 0xDC00)
					b.WriteRune(codePoint)
					i += 11
					continue
				}
			}
		}

		b.WriteRune(r)
		i += 5
	}

	return b.String()
}

// FormatTestTreeOutput converts a parsed DOM tree to the html5lib-tests format.
// This is used to compare actual output against expected output.
func FormatTestTreeOutput(lines []string) string {
	return strings.Join(lines, "\n")
}

// TestResult holds the result of running a single test.
type TestResult struct {
	Passed         bool
	TestName       string
	Input          string
	Expected       string
	Actual         string
	ExpectedErrors []string
	ActualErrors   []string
	ErrorMessage   string
}

// TestSummary holds aggregate results for a test file.
type TestSummary struct {
	FileName string
	Passed   int
	Failed   int
	Skipped  int
	Total    int
	Results  []TestResult
}

// FormatSummary returns a formatted summary string.
func (s *TestSummary) FormatSummary() string {
	runnable := s.Passed + s.Failed
	if runnable == 0 {
		return fmt.Sprintf("%s: 0/0 (N/A)", s.FileName)
	}
	pct := float64(s.Passed) * 100 / float64(runnable)
	result := fmt.Sprintf("%s: %d/%d (%.0f%%)", s.FileName, s.Passed, runnable, pct)
	if s.Skipped > 0 {
		result += fmt.Sprintf(" (%d skipped)", s.Skipped)
	}
	return result
}
