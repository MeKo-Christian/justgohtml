# JustGoHTML: Python to Go Rewrite Plan

This document outlines a comprehensive plan for rewriting JustGoHTML from Python to Go while maintaining 100% HTML5 compliance and the zero-dependency philosophy.

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture Mapping](#2-architecture-mapping)
3. [Phase 1: Foundation](#3-phase-1-foundation)
4. [Phase 2: Tokenizer](#4-phase-2-tokenizer)
5. [Phase 3: Tree Builder](#5-phase-3-tree-builder)
6. [Phase 4: DOM & Selectors](#6-phase-4-dom--selectors)
7. [Phase 5: Public API & CLI](#7-phase-5-public-api--cli)
8. [Phase 6: Testing & Validation](#8-phase-6-testing--validation)
9. [Phase 7: Documentation & Release](#9-phase-7-documentation--release)
10. [Phase 8: WebAssembly Playground](#10-phase-8-webassembly-playground)
11. [Technical Decisions](#11-technical-decisions)
12. [Risk Assessment](#12-risk-assessment)
13. [Success Criteria](#13-success-criteria)

---

## Project Overview

### Current State (Python)

| Metric              | Value              |
| ------------------- | ------------------ |
| Total Lines of Code | ~9,858             |
| Modules             | 18                 |
| Test Coverage       | 100%               |
| HTML5 Compliance    | 100% (9k+ tests)   |
| Dependencies        | Zero (stdlib only) |

### Goals for Go Rewrite

1. **Maintain 100% HTML5 Compliance** - Pass all html5lib-tests
2. **Zero Dependencies** - Use Go stdlib only
3. **Idiomatic Go** - Follow Go conventions and best practices
4. **Performance** - Target 5-10x improvement over Python version
5. **API Simplicity** - Clean, discoverable public API
6. **Cross-Platform** - Single binary distribution

### Non-Goals

- Adding features not in the Python version
- Supporting older Go versions (target Go 1.21+)
- CGO dependencies

---

## Architecture Mapping

### Python → Go Module Mapping

```
Python Module              →  Go Package
─────────────────────────────────────────────────────────
src/JustGoHTML/
├── __init__.py            →  JustGoHTML.go (main exports)
├── __main__.py            →  cmd/JustGoHTML/main.go (CLI)
├── parser.go              →  parser.go
├── tokenizer.py (2,647)   →  tokenizer/tokenizer.go
├── treebuilder.py (1,279) →  treebuilder/builder.go
├── treebuilder_modes.py   →  treebuilder/modes.go
├── treebuilder_utils.py   →  treebuilder/utils.go
├── node.py (632)          →  dom/node.go
├── selector.py (965)      →  selector/selector.go
├── serialize.py (258)     →  serialize/serialize.go
├── stream.py (107)        →  stream/stream.go
├── encoding.py (405)      →  encoding/encoding.go
├── tokens.py (223)        →  tokenizer/tokens.go
├── entities.py (344)      →  entities/entities.go
├── constants.py (445)     →  internal/constants/constants.go
├── errors.py (140)        →  errors/errors.go
└── context.py (12)        →  context.go
```

### Proposed Go Project Structure

```
JustGoHTML/
├── cmd/
│   └── JustGoHTML/
│       └── main.go              # CLI entry point
├── internal/
│   ├── constants/
│   │   ├── elements.go          # Element classifications
│   │   ├── entities.go          # HTML5 named entities
│   │   └── scopes.go            # Scope terminators
│   └── testutil/
│       └── helpers.go           # Test utilities
├── tokenizer/
│   ├── tokenizer.go             # Main tokenizer
│   ├── tokens.go                # Token types
│   ├── states.go                # State machine constants
│   └── tokenizer_test.go
├── treebuilder/
│   ├── builder.go               # TreeBuilder core
│   ├── modes.go                 # Insertion mode handlers
│   ├── adoption.go              # Adoption agency algorithm
│   ├── foreign.go               # SVG/MathML handling
│   └── builder_test.go
├── dom/
│   ├── node.go                  # Node interface & types
│   ├── element.go               # ElementNode
│   ├── text.go                  # TextNode
│   ├── document.go              # Document node
│   └── node_test.go
├── selector/
│   ├── parser.go                # Selector parsing
│   ├── matcher.go               # DOM matching
│   ├── ast.go                   # Selector AST types
│   └── selector_test.go
├── encoding/
│   ├── detect.go                # Encoding detection
│   ├── decode.go                # Decoding logic
│   └── encoding_test.go
├── serialize/
│   ├── html.go                  # HTML serialization
│   ├── text.go                  # Text extraction
│   ├── markdown.go              # Markdown conversion
│   └── serialize_test.go
├── stream/
│   ├── stream.go                # Streaming API
│   └── stream_test.go
├── errors/
│   ├── parse_error.go           # Parse errors
│   ├── selector_error.go        # Selector errors
│   └── messages.go              # Error messages
├── parser.go                    # JustGoHTML main entry
├── options.go                   # Configuration options
├── JustGoHTML_test.go             # Integration tests
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Key Type Mappings

| Python                   | Go                                          |
| ------------------------ | ------------------------------------------- |
| `str`                    | `string`                                    |
| `bytes`                  | `[]byte`                                    |
| `dict[str, str \| None]` | `map[string]*string` or custom `Attrs` type |
| `list[T]`                | `[]T`                                       |
| `Optional[T]`            | `*T` or custom optional type                |
| `enum.Enum`              | `const` block with `iota`                   |
| `Generator[T, ...]`      | `chan T` or iterator pattern                |
| `Callable[...]`          | `func(...)`                                 |

---

## Phase 1: Foundation

### 1.1 Project Setup

- [x] Initialize Go module: `go mod init github.com/MeKo-Christian/JustGoHTML`
- [x] Create directory structure per Section 2.2
- [x] Set up justfile with common commands (fmt, lint, test, build)
- [x] Configure golangci-lint with strict settings
- [x] Set up GitHub Actions CI pipeline
- [x] Configure treefmt for code formatting
- [x] Update .gitignore for Go artifacts

### 1.2 Core Types & Constants

**File: `internal/constants/elements.go`**

- [x] Port `SPECIAL_ELEMENTS` set
- [x] Port `FORMATTING_ELEMENTS` set
- [x] Port `VOID_ELEMENTS` set
- [x] Port scope terminator maps (button, table, list item, etc.)
- [x] Port SVG/MathML tag adjustments
- [x] Port foreign attribute mappings

**File: `internal/constants/entities.go`**

- [x] Port 2,231 HTML5 named entities (2,125 unique entity names)
- [x] Implement entity lookup (using `map[string]string`)
- [x] Handle legacy entities (without semicolon) via `LegacyEntities` map
- [x] Implement numeric character reference replacements (28 replacements)

**File: `errors/parse_error.go`**

- [x] Define `ParseError` struct with code, line, column, message
- [x] Port 40+ error codes from Python
- [x] Implement `Error()` method for Go error interface

**File: `errors/messages.go`**

- [x] Port human-readable error message templates

### 1.3 Encoding Detection

**File: `encoding/encoding.go`**

- [x] Implement BOM detection (UTF-8)
- [x] Implement meta charset scanning
- [x] Implement XML encoding declaration parsing
- [x] Handle encoding label normalization
- [x] Implement fallback chain

**File: `encoding/decode.go`**

- [x] Implement `Decode([]byte, hints) (string, Encoding, error)`
- [x] Handle encoding errors gracefully
- [x] Support basic encoding labels (UTF-8, windows-1252, ISO-8859-1)
- [x] Support full 20+ encoding labels (43 distinct labels supported)

### 1.4 Deliverables

- [x] All tests passing for constants and encoding
- [x] Over 90% test coverage for Phase 1 code (actually 96.2% coverage; encoding: 95.9%, errors: 100%, constants: 100%)
- [x] Benchmarks for entity lookup (8 benchmarks covering common/uncommon/missing lookups)

---

## Phase 2: Tokenizer

### 2.1 Token Types

**File: `tokenizer/tokens.go`**

- [x] Define `TokenKind` enum (StartTag, EndTag, Character, Comment, DOCTYPE, EOF)
- [x] Define `Tag` struct with name, attrs, self-closing flag
- [x] Define `CharacterToken` struct
- [x] Define `CommentToken` struct
- [x] Define `DoctypeToken` struct with name, public ID, system ID, force-quirks
- [x] Define `Token` interface or sum type

### 2.2 State Machine

**File: `tokenizer/states.go`**

- [x] Define tokenizer state constants
- [x] Document state transitions

**File: `tokenizer/tokenizer.go`**

Core structure:

```go
type Tokenizer struct {
    input         string
    pos           int
    state         State
    returnState   State
    currentTag    *Tag
    currentAttr   *Attribute
    tempBuffer    strings.Builder
    charRefCode   int
    errors        []ParseError
    // ... additional fields
}
```

Implement state handlers:

- [x] DATA state
- [x] RCDATA state
- [x] RAWTEXT state
- [x] SCRIPT_DATA state (and escaped variants)
- [x] PLAINTEXT state
- [x] TAG_OPEN state
- [x] END_TAG_OPEN state
- [x] TAG_NAME state
- [x] RCDATA/RAWTEXT/SCRIPT_DATA less-than sign states
- [x] BEFORE_ATTRIBUTE_NAME state
- [x] ATTRIBUTE_NAME state
- [x] AFTER_ATTRIBUTE_NAME state
- [x] BEFORE_ATTRIBUTE_VALUE state
- [x] ATTRIBUTE*VALUE*\* states (double-quoted, single-quoted, unquoted)
- [x] AFTER_ATTRIBUTE_VALUE_QUOTED state
- [x] SELF_CLOSING_START_TAG state
- [x] BOGUS_COMMENT state
- [x] MARKUP_DECLARATION_OPEN state
- [x] COMMENT\_\* states (7 states)
- [x] DOCTYPE\_\* states (16 states)
- [x] CDATA*SECTION*\* states
- [x] Character reference handling (via entity decoder)

Additional methods:

- [x] `consumeCharacterReference()` - entity decoding (see `tokenizer/entities.go`)
- [x] `flushCodePoints()` - emit pending characters
- [x] `emitCurrentTag()` - emit tag token
- [x] `isAppropriateEndTag()` - check end tag validity

### 2.3 Regex Optimization

Go doesn't have Python's regex performance, so consider:

- [ ] Use `strings.IndexAny()` for character class scanning
- [ ] Use `strings.IndexByte()` for single character scanning
- [x] Pre-compile any necessary regexes at package init

Notes:

- The tokenizer currently reads runes one-by-one to mirror the spec (line/column tracking, reconsume handling). Switching to `IndexAny` would complicate correctness and doesn’t match the existing per-rune state machine paths.
- There are no single-byte scanning loops to swap to `IndexByte`; converting between runes/bytes would add overhead and risk correctness. Revisit if we add safe skip-ahead string scanning (e.g., raw text until `<`) with benchmarks.

### 2.4 Testing

- [x] Port tokenizer tests from html5lib-tests
- [x] Add tokenizer unit tests
- [x] Cover character reference resolution (html5lib + focused unit tests)
- [x] Cover error conditions (html5lib + JustGoHTML tests)
- [x] Benchmark: tokens per second

---

## Phase 3: Tree Builder

### 3.1 Core Types

**File: `treebuilder/modes.go`** (scaffolding created)

- [x] Define `InsertionMode` enum with all 21 modes
- [x] Implement `String()` method for debugging

**File: `treebuilder/context.go`** (scaffolding created)

- [x] Define `FragmentContext` struct

**File: `treebuilder/builder.go`**

```go
type InsertionMode int

const (
    Initial InsertionMode = iota
    BeforeHTML
    BeforeHead
    InHead
    InHeadNoscript
    AfterHead
    InBody
    Text
    InTable
    InTableText
    InCaption
    InColumnGroup
    InTableBody
    InRow
    InCell
    InSelect
    InSelectInTable
    InTemplate
    AfterBody
    InFrameset
    AfterFrameset
    AfterAfterBody
    AfterAfterFrameset
)

type TreeBuilder struct {
    document         *dom.Document
    openElements     []*dom.Element
    activeFormatting []formattingEntry
    templateModes    []InsertionMode
    pendingTableText []string

    mode             InsertionMode
    originalMode     InsertionMode

    headElement      *dom.Element
    formElement      *dom.Element

    framesetOK       bool
    quirksMode       QuirksMode

    fragmentContext  *FragmentContext

    errors           []ParseError
    strict           bool
}
```

### 3.2 Insertion Mode Handlers

**File: `treebuilder/modes.go`**

Each mode handler signature:

```go
func (tb *TreeBuilder) processInMode(token Token) error
```

Implement all 21 mode handlers:

- [x] `processInitial`
- [x] `processBeforeHTML`
- [x] `processBeforeHead`
- [x] `processInHead`
- [x] `processInHeadNoscript`
- [x] `processAfterHead`
- [x] `processInBody` (largest, most complex)
- [x] `processText`
- [x] `processInTable`
- [x] `processInTableText`
- [x] `processInCaption`
- [x] `processInColumnGroup`
- [x] `processInTableBody`
- [x] `processInRow`
- [x] `processInCell`
- [x] `processInSelect`
- [x] `processInSelectInTable`
- [x] `processInTemplate`
- [x] `processAfterBody`
- [x] `processInFrameset`
- [x] `processAfterFrameset`
- [x] `processAfterAfterBody`
- [x] `processAfterAfterFrameset`

### 3.3 Supporting Algorithms

**File: `treebuilder/adoption.go`**

- [x] Implement adoption agency algorithm
- [x] Handle formatting element reconstruction

**File: `treebuilder/foreign.go`**

- [x] SVG namespace handling
- [x] MathML namespace handling
- [x] Integration point detection
- [x] Tag name case adjustment
- [x] Attribute name adjustment

**File: `treebuilder/utils.go`**

- [x] `hasElementInScope(tagName, scope)`
- [x] `hasElementInButtonScope(tagName)`
- [x] `hasElementInTableScope(tagName)`
- [x] `hasElementInListItemScope(tagName)`
- [x] `generateImpliedEndTags(except)`
- [x] `resetInsertionModeAppropriately()`
- [x] `reconstructActiveFormattingElements()`
- [x] `clearActiveFormattingElements()`
- [x] `pushActiveFormattingMarker()`
- [x] Quirks mode detection logic

### 3.4 Testing

- [x] Port tree-construction tests from html5lib-tests (4,543 tests total)
  - Test harness complete: runs all 56 .dat files
  - Current results: 2,773 passed (61%), 1,028 failed (23%), 742 skipped (16%)
  - No infinite loops or timeouts
  - Fixed: void elements (basefont, bgsound, link, meta, base) now handled correctly in InBody mode
- [~] Test all insertion modes (partially covered via html5lib tests)
- [~] Test quirks mode detection (partially covered via html5lib tests)
- [~] Test fragment parsing (partially covered via html5lib tests)
- [~] Test foreign content handling (partially covered via html5lib tests; fixed infinite loop bug)
- [x] Benchmark: parse time for large documents
  - BenchmarkTreeBuilder: ~5.7ms for 1000 complex divs (~70KB HTML)
  - BenchmarkTreeBuilderSmall: ~4.8µs for minimal document
  - BenchmarkTreeBuilderTables: ~305µs for 100-row table
  - BenchmarkTreeBuilderNested: ~64µs for 100-level deep nesting
  - BenchmarkTreeBuilderForeign: ~355µs for 100 SVG elements
  - Implemented xml-coercion/iframe-srcdoc options in parser and test harness; tree-construction tests now decode escaped input sequences

---

## 4. Phase 4: DOM & Selectors

### 4.1 DOM Node Types

**File: `dom/node.go`** ✅

- [x] Define `NodeType` enum
- [x] Define `Node` interface with common operations
- [x] Implement `baseNode` with shared functionality

**File: `dom/element.go`** ✅

- [x] Define `Element` struct with TagName, Namespace, Attributes, TemplateContent
- [x] Implement `AppendChild`, `InsertBefore`, `RemoveChild`
- [x] Implement `ReplaceChild`
- [x] Implement `Clone(deep bool)`
- [x] Implement `Query`, `QueryFirst` (stub - delegates to selector package)
- [x] Implement `Text()` for text extraction
- [x] Implement `Attr`, `HasAttr`, `SetAttr`, `RemoveAttr`
- [x] Implement `ID()`, `Classes()`, `HasClass()`

**File: `dom/text.go`** ✅

- [x] Define `Text` struct
- [x] Define `Comment` struct
- [x] Implement `Node` interface for both

**File: `dom/document.go`** ✅

- [x] Define `Document` struct with Doctype, QuirksMode
- [x] Define `DocumentType` struct
- [x] Define `DocumentFragment` struct
- [x] Implement `DocumentElement()`, `Head()`, `Body()`, `Title()`
- [x] Implement `Query`, `QueryFirst`

### 4.2 Attribute Handling

**File: `dom/attributes.go`** ✅

- [x] Define `Attributes` struct with ordered storage
- [x] Define `Attribute` struct with Namespace, Name, Value
- [x] Implement `Get`, `GetNS` (case-insensitive for HTML)
- [x] Implement `Set`, `SetNS`
- [x] Implement `Has`, `HasNS`
- [x] Implement `Remove`, `RemoveNS`
- [x] Implement `All`, `Len`, `Clone`

### 4.3 CSS Selector Engine ✅

**File: `selector/ast.go`** ✅

- [x] Define `SelectorKind` enum (Tag, Universal, ID, Class, Attr, Pseudo)
- [x] Define `AttrOperator` enum (Exists, Equals, Includes, DashPrefix, PrefixMatch, SuffixMatch, Substring)
- [x] Define `Combinator` enum (None, Descendant, Child, Adjacent, General)
- [x] Define `SimpleSelector`, `CompoundSelector`, `ComplexPart`, `ComplexSelector`, `SelectorList` types

**File: `selector/parser.go`** ✅

- [x] Tokenize selector string (character-by-character state machine)
- [x] Parse selector groups (comma-separated)
- [x] Parse complex selectors (with combinators)
- [x] Parse compound selectors
- [x] Parse pseudo-classes with arguments
- [x] Handle escape sequences in strings

**File: `selector/matcher.go`** ✅

Implement matching for:

- [x] Type selector (`div`, `*`)
- [x] ID selector (`#id`)
- [x] Class selector (`.class`)
- [x] Attribute selectors (`[attr]`, `[attr="val"]`, `[attr~="val"]`, `[attr^="val"]`, `[attr$="val"]`, `[attr*="val"]`, `[attr|="val"]`)
- [x] `:first-child`, `:last-child`, `:only-child`
- [x] `:nth-child(an+b)`, `:nth-last-child(an+b)`
- [x] `:first-of-type`, `:last-of-type`, `:only-of-type`
- [x] `:nth-of-type(an+b)`, `:nth-last-of-type(an+b)`
- [x] `:not(selector)`
- [x] `:empty`
- [x] `:root`
- [x] Combinators (descendant, child, adjacent sibling, general sibling)
- [x] Right-to-left matching algorithm for efficiency

### 4.4 Testing

- [x] Node tree manipulation tests (99.1% coverage in dom package)
- [x] Selector parsing tests
- [x] Selector matching tests
- [x] Query integration tests
- [x] Benchmark: selectors per second (BenchmarkParse, BenchmarkMatch)

---

## Phase 5: Public API & CLI

### 5.1 Main Entry Point ✅

**File: `justhtml.go`**

- [x] Define `Parse(html string, opts ...Option) (*Document, error)`
- [x] Define `ParseBytes(html []byte, opts ...Option) (*Document, error)`
- [x] Define `ParseFragment(html string, context string, opts ...Option) ([]*Element, error)`
- [x] Implement tokenizer -> tree builder pipeline

**File: `options.go`**

- [x] Define `Option` function type
- [x] Implement `WithEncoding(enc string) Option`
- [x] Implement `WithFragment(tagName string) Option`
- [x] Implement `WithFragmentNS(tagName, namespace string) Option`
- [x] Implement `WithIframeSrcdoc() Option`
- [x] Implement `WithStrictMode() Option`
- [x] Implement `WithCollectErrors() Option`
- [x] Implement `WithXMLCoercion() Option`

### 5.2 Streaming API ✅

**File: `stream/stream.go`**

```go
type EventType int

const (
    StartTagEvent EventType = iota
    EndTagEvent
    TextEvent
    CommentEvent
    DoctypeEvent
)

type Event struct {
    Type    EventType
    Name    string           // tag name for tags
    Attrs   map[string]string // for start tags
    Data    string           // for text/comment
}

// Stream returns a channel of parsing events.
func Stream(html string, opts ...Option) <-chan Event

// StreamBytes returns a channel of parsing events from bytes.
func StreamBytes(html []byte, opts ...Option) <-chan Event
```

### 5.3 Serialization

**File: `serialize/html.go`**

```go
type SerializeOption func(*serializeConfig)

func WithPrettyPrint() SerializeOption
func WithIndent(size int) SerializeOption

func ToHTML(node dom.Node, opts ...SerializeOption) string
```

**File: `serialize/text.go`**

```go
type TextOption func(*textConfig)

func WithSeparator(sep string) TextOption
func WithStripWhitespace(strip bool) TextOption

func ToText(node dom.Node, opts ...TextOption) string
```

**File: `serialize/markdown.go`**

```go
func ToMarkdown(node dom.Node) string
```

### 5.4 CLI Tool ✅

**File: `cmd/justgohtml/main.go`**

```go
Usage: justgohtml [options] <file>

Arguments:
  file                 HTML file path or '-' for stdin

Options:
  --selector, -s       CSS selector to filter output
  --format, -f         Output format: html, text, markdown (default: html)
  --first              Output only the first match
  --separator          Separator for text output (default: " ")
  --strip              Strip whitespace from text (default: true)
  --pretty             Pretty-print HTML output (default: true)
  --indent             Indentation size for pretty-print (default: 2)
  --version, -v        Show version
  --help, -h           Show help
```

Implementation:

- [x] Argument parsing with `flag` package
- [x] Stdin reading
- [x] File reading
- [x] Selector filtering
- [x] Format output (HTML, text, markdown)
- [x] Error handling and exit codes

### 5.5 Testing

- [ ] API usage tests
- [ ] CLI argument parsing tests
- [ ] CLI output format tests
- [ ] Stdin handling tests

---

## Phase 6: Testing & Validation

### 6.1 Test Infrastructure

- [x] Set up test helpers in `internal/testutil/`
- [x] Create test fixtures directory structure
- [x] Download html5lib-tests submodule or vendored copy

### 6.2 html5lib-tests Integration

**Tests to pass:**

| Test Category     | Count | Priority | Status                     |
| ----------------- | ----- | -------- | -------------------------- |
| Tree Construction | 1,843 | Critical | 70% (1,282 pass, 549 fail) |
| Tokenizer         | 6,826 | Critical | 100% (all pass)            |
| Serializer        | 236   | High     | 89% (210 pass, 26 fail)    |
| Encoding          | 87    | High     | 100% (all pass)            |

- [x] Create test harness for tree construction tests
- [x] Create test harness for tokenizer tests
- [x] Create test harness for serializer tests
- [x] Create test harness for encoding tests
- [ ] All 9,000+ tests passing (currently ~8,100 passing)

### 6.3 Unit Test Coverage

Target: 100% coverage (matching Python version)

- [ ] tokenizer package: 100%
- [ ] treebuilder package: 100%
- [ ] dom package: 100%
- [ ] selector package: 100%
- [ ] encoding package: 100%
- [ ] serialize package: 100%
- [ ] stream package: 100%

### 6.4 Fuzz Testing

- [ ] Set up Go fuzzing for tokenizer
- [ ] Set up Go fuzzing for tree builder
- [ ] Set up Go fuzzing for selector parser
- [ ] Run fuzzer for extended period (match Python's 6M documents)

### 6.5 Benchmark Suite

**File: `benchmarks/benchmark_test.go`**

- [ ] Parse speed benchmark (Wikipedia homepage)
- [ ] Memory allocation benchmark
- [ ] Selector matching benchmark
- [ ] Serialization benchmark
- [ ] Streaming benchmark
- [ ] Comparison with other Go HTML parsers (golang.org/x/net/html)

---

## Phase 7: Documentation & Release

### 7.1 Documentation

- [ ] README.md with quickstart
- [ ] GoDoc comments on all public types and functions
- [ ] docs/quickstart.md
- [ ] docs/api.md
- [ ] docs/cli.md
- [ ] docs/selectors.md
- [ ] docs/streaming.md
- [ ] docs/encoding.md
- [ ] docs/errors.md
- [ ] docs/fragments.md

### 7.2 Release Preparation

- [ ] Semantic versioning: start at v0.1.0
- [ ] CHANGELOG.md
- [ ] LICENSE file (MIT)
- [ ] CONTRIBUTING.md
- [ ] GitHub release automation
- [ ] Binary builds for Linux, macOS, Windows

### 7.3 CI/CD

- [ ] GitHub Actions workflow for tests
- [ ] GitHub Actions workflow for linting
- [ ] GitHub Actions workflow for releases
- [ ] Dependabot configuration
- [ ] Badge setup in README

---

## Technical Decisions

### 8.1 String vs []byte

**Decision:** Use `string` for DOM storage, `[]byte` for input processing.

**Rationale:**

- Go strings are immutable, matching HTML spec behavior
- DOM queries and manipulation work naturally with strings
- Input can be bytes; convert after encoding detection

### 8.2 Error Handling

**Decision:** Return errors, don't panic. Offer strict mode option.

```go
// Normal mode: collect errors, continue parsing
doc, err := JustGoHTML.Parse(html)
if err != nil {
    // err contains all parse errors
    parseErrors := err.(JustGoHTML.ParseErrors)
}

// Strict mode: fail on first error
doc, err := JustGoHTML.Parse(html, JustGoHTML.WithStrictMode())
```

### 8.3 Concurrency

**Decision:** Single-threaded parsing, concurrent-safe read-only DOM access.

**Rationale:**

- HTML parsing is inherently sequential
- DOM queries can be parallelized by users if needed
- Streaming API enables concurrent processing

### 8.4 Memory Management

**Decision:** Pool tokens during parsing, copy when needed.

```go
var tokenPool = sync.Pool{
    New: func() interface{} {
        return &Tag{}
    },
}
```

### 8.5 Entity Storage

**Decision:** Use `map[string]string` for entities with potential trie optimization.

**Rationale:**

- 2,231 entities is manageable for a map
- Profile first, optimize if necessary
- Trie would help with prefix matching for legacy entities

---

## 9. Risk Assessment

### 9.1 High Risk Items

| Risk                       | Mitigation                                         |
| -------------------------- | -------------------------------------------------- |
| Spec compliance gaps       | Run html5lib-tests continuously during development |
| Adoption agency complexity | Port directly from Python, test exhaustively       |
| Performance regressions    | Benchmark suite, profile before optimizing         |
| Edge cases in encoding     | Comprehensive encoding test suite                  |

### 9.2 Medium Risk Items

| Risk                         | Mitigation                               |
| ---------------------------- | ---------------------------------------- |
| Selector parser completeness | Test against real-world selectors        |
| Foreign content (SVG/MathML) | Specific test cases from html5lib        |
| Memory usage for large docs  | Streaming API for memory-constrained use |

### 9.3 Low Risk Items

| Risk                       | Mitigation                      |
| -------------------------- | ------------------------------- |
| CLI feature parity         | Simple, well-defined scope      |
| Serialization correctness  | Direct port, serializer tests   |
| Documentation completeness | Copy structure from Python docs |

---

## 10. Success Criteria

### 10.1 Must Have (MVP)

- [ ] Pass all 1,743 tree-construction tests
- [ ] Pass all tokenizer tests
- [ ] Working CLI with same features as Python
- [ ] `Parse()`, `ParseBytes()`, `ParseFragment()` APIs
- [ ] CSS selector support (`Query()`, `QueryFirst()`)
- [ ] HTML, text, and Markdown output
- [ ] 100% test coverage

### 10.2 Should Have (v1.0)

- [ ] Pass all 9,000+ html5lib-tests
- [ ] Streaming API
- [ ] Encoding detection
- [ ] Complete documentation
- [ ] Benchmarks showing performance improvement
- [ ] Fuzz tested

### 10.3 Nice to Have (Future)

- [x] WASM build for browser usage (see Phase 8)
- [ ] HTML sanitization helpers
- [ ] Additional output formats (JSON DOM)
- [ ] Parallel selector matching

---

## Phase 8: WebAssembly Playground (Future)

### 8.1 WebAssembly Compilation

**Status**: Complete

- [x] Add `GOOS=js GOARCH=wasm` build configuration
- [x] Create JavaScript bindings for WASM module
- [x] Implement streaming tokenizer output for real-time display
- [x] Optimize WASM bundle size (consider UPX compression)
- [x] Add browser-compatible module export

**Files created:**

- `cmd/wasm/main.go` - WASM entry point with JS bindings
- `justfile` - Added `build-wasm`, `build-wasm-tiny`, `serve-playground` commands

### 8.2 Interactive Playground UI

**Status**: Complete

**File: `playground/index.html`**

- [x] Split-pane interface (HTML input | output)
- [x] Real-time parsing with tokenizer event stream
- [x] Multi-format output tabs (HTML, Markdown, Text, Tree view)
- [ ] CSS selector query box with live results highlighting
- [x] Parse error display with line/column information
- [x] Copy-to-clipboard buttons for each output format
- [x] GitHub Pages deployment (via `.github/workflows/deploy-playground.yaml`)

**Design Reference**: The JavaScript port's [`playground.html`](https://github.com/simonw/justjshtml/blob/main/playground.html) provides a model for a simple, single-file implementation.

### 8.3 Performance Considerations

- [ ] Profile WASM bundle size (target: < 2MB)
- [ ] Implement incremental parsing for large documents
- [ ] Add debouncing for real-time input handling
- [ ] Consider worker thread for non-blocking parsing

### 8.4 Testing

- [ ] Add integration tests for WASM bindings
- [ ] Test playground in modern browsers (Chrome, Firefox, Safari, Edge)
- [ ] Add E2E tests using Playwright or similar

### 8.5 Documentation

- [ ] Add WASM build instructions to README
- [ ] Document JavaScript API for WASM module
- [ ] Add browser compatibility matrix
- [ ] Create playground user guide

### 8.6 Deliverables

- [ ] Live playground at `https://github.com/EmilStenstrom/JustGoHTML/releases` (via GitHub Pages)
- [ ] WASM module documented and published
- [ ] Browser compatibility verified
- [ ] Performance benchmarks (WASM vs native Go)

---

## Appendix A: Estimated Effort by Phase

| Phase               | Estimated Go LOC | Complexity |
| ------------------- | ---------------- | ---------- |
| 1. Foundation       | ~1,500           | Medium     |
| 2. Tokenizer        | ~3,000           | High       |
| 3. Tree Builder     | ~4,000           | Very High  |
| 4. DOM & Selectors  | ~2,000           | Medium     |
| 5. Public API & CLI | ~500             | Low        |
| 6. Testing          | ~3,000           | Medium     |
| 7. Documentation    | N/A              | Low        |
| 8. WASM Playground  | ~1,500           | High       |
| **Total**           | **~15,500**      |            |

---

## Appendix B: Reference Materials

- [WHATWG HTML Living Standard - Parsing](https://html.spec.whatwg.org/multipage/parsing.html)
- [WHATWG HTML Living Standard - Tokenization](https://html.spec.whatwg.org/multipage/parsing.html#tokenization)
- [html5lib-tests Repository](https://github.com/html5lib/html5lib-tests)
- [html5ever (Rust reference)](https://github.com/servo/html5ever)
- [Go html package](https://pkg.go.dev/golang.org/x/net/html)
- [CSS Selectors Level 4](https://www.w3.org/TR/selectors-4/)

---

## Appendix C: Quick Reference - State Machine States

### Tokenizer States (49)

```
DATA, RCDATA, RAWTEXT, SCRIPT_DATA, PLAINTEXT,
TAG_OPEN, END_TAG_OPEN, TAG_NAME,
RCDATA_LESS_THAN_SIGN, RCDATA_END_TAG_OPEN, RCDATA_END_TAG_NAME,
RAWTEXT_LESS_THAN_SIGN, RAWTEXT_END_TAG_OPEN, RAWTEXT_END_TAG_NAME,
SCRIPT_DATA_LESS_THAN_SIGN, SCRIPT_DATA_END_TAG_OPEN, SCRIPT_DATA_END_TAG_NAME,
SCRIPT_DATA_ESCAPE_START, SCRIPT_DATA_ESCAPE_START_DASH,
SCRIPT_DATA_ESCAPED, SCRIPT_DATA_ESCAPED_DASH, SCRIPT_DATA_ESCAPED_DASH_DASH,
SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN, SCRIPT_DATA_ESCAPED_END_TAG_OPEN,
SCRIPT_DATA_ESCAPED_END_TAG_NAME, SCRIPT_DATA_DOUBLE_ESCAPE_START,
SCRIPT_DATA_DOUBLE_ESCAPED, SCRIPT_DATA_DOUBLE_ESCAPED_DASH,
SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH, SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN,
SCRIPT_DATA_DOUBLE_ESCAPE_END,
BEFORE_ATTRIBUTE_NAME, ATTRIBUTE_NAME, AFTER_ATTRIBUTE_NAME,
BEFORE_ATTRIBUTE_VALUE, ATTRIBUTE_VALUE_DOUBLE_QUOTED, ATTRIBUTE_VALUE_SINGLE_QUOTED,
ATTRIBUTE_VALUE_UNQUOTED, AFTER_ATTRIBUTE_VALUE_QUOTED,
SELF_CLOSING_START_TAG, BOGUS_COMMENT,
MARKUP_DECLARATION_OPEN, COMMENT_START, COMMENT_START_DASH, COMMENT,
COMMENT_LESS_THAN_SIGN, COMMENT_LESS_THAN_SIGN_BANG, COMMENT_LESS_THAN_SIGN_BANG_DASH,
COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH, COMMENT_END_DASH, COMMENT_END, COMMENT_END_BANG,
DOCTYPE, BEFORE_DOCTYPE_NAME, DOCTYPE_NAME, AFTER_DOCTYPE_NAME,
AFTER_DOCTYPE_PUBLIC_KEYWORD, BEFORE_DOCTYPE_PUBLIC_IDENTIFIER,
DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED, DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED,
AFTER_DOCTYPE_PUBLIC_IDENTIFIER, BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS,
AFTER_DOCTYPE_SYSTEM_KEYWORD, BEFORE_DOCTYPE_SYSTEM_IDENTIFIER,
DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED, DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED,
AFTER_DOCTYPE_SYSTEM_IDENTIFIER, BOGUS_DOCTYPE,
CDATA_SECTION, CDATA_SECTION_BRACKET, CDATA_SECTION_END,
CHARACTER_REFERENCE, NAMED_CHARACTER_REFERENCE, AMBIGUOUS_AMPERSAND,
NUMERIC_CHARACTER_REFERENCE, HEXADECIMAL_CHARACTER_REFERENCE_START,
DECIMAL_CHARACTER_REFERENCE_START, HEXADECIMAL_CHARACTER_REFERENCE,
DECIMAL_CHARACTER_REFERENCE, NUMERIC_CHARACTER_REFERENCE_END
```

### Tree Builder Insertion Modes (21)

```
INITIAL, BEFORE_HTML, BEFORE_HEAD, IN_HEAD, IN_HEAD_NOSCRIPT,
AFTER_HEAD, IN_BODY, TEXT, IN_TABLE, IN_TABLE_TEXT,
IN_CAPTION, IN_COLUMN_GROUP, IN_TABLE_BODY, IN_ROW, IN_CELL,
IN_SELECT, IN_SELECT_IN_TABLE, IN_TEMPLATE,
AFTER_BODY, IN_FRAMESET, AFTER_FRAMESET,
AFTER_AFTER_BODY, AFTER_AFTER_FRAMESET
```
