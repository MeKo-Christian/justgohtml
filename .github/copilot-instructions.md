## JustGoHTML – Agent Instructions

# Decision & Clarification Policy

- Propose **and execute** the best alternative by default; ask only for destructive/irreversible choices
- Keep preambles to a single declarative sentence ("Scanning repo and drafting minimal fix") — no approval requests

## Architecture

### Core Components

- **tokenizer** (`tokenizer/`): HTML5 spec state machine (~60 states). Handles RCDATA, RAWTEXT, CDATA, script escaping, comments, DOCTYPE
- **treebuilder** (`treebuilder/`): Consumes tokens, constructs DOM tree following HTML5 insertion mode rules (23 modes)
- **dom** (`dom/`): Document object model - Element, Text, Document, DocumentType nodes
- **selector** (`selector/`): CSS selector parsing and matching (not yet implemented)
- **encoding** (`encoding/`): HTML5 encoding detection per spec (prescan, BOM, meta tags)
- **serialize** (`serialize/`): Converts DOM back to HTML string

### Data Flow

```
Input bytes → encoding.Decode() → Tokenizer → TreeBuilder → DOM
DOM → selector.Match() → Query results
DOM → serialize.ToHTML() → HTML string
```

### Key Files

- `JustGoHTML.go`: Main API (`Parse`, `ParseBytes`, `ParseFragment`)
- `internal/testutil/html5lib.go`: Test harness for html5lib-tests suite
- `internal/constants/`: HTML5 element categories, scoping rules

## Development Workflow

### Build & Run

```bash
just setup         # Install dev dependencies (golangci-lint, treefmt, etc.)
just build         # Build CLI binary
just run -- input  # Run CLI during development
```

### Testing

```bash
just test          # Run all tests
just test-v        # Verbose output
just test-race     # With race detector
just test-coverage # Generate coverage report (required: 100%)

# Run specific test file
go test ./tokenizer -v

# Run html5lib-tests (9k+ browser vendor tests)
go test ./tokenizer -run HTML5Lib
go test ./treebuilder -run HTML5Lib
go test ./serialize -run HTML5Lib
go test ./encoding -run HTML5Lib
```

### Code Quality

```bash
just fmt           # Format all code (treefmt)
just lint          # Run golangci-lint
just lint-fix      # Auto-fix linting issues
just check         # Format + test + lint (pre-commit)
```

### html5lib-tests Location

Tests in `testdata/html5lib-tests/` (git submodule):

- `tokenizer/*.test` (JSON format)
- `tree-construction/*.dat` (text format)
- `serializer/*.test` (JSON format)
- `encoding/*.dat` (text format)

Test harness in `internal/testutil/html5lib.go` parses these formats

## Go-Specific Conventions

### Code Style

1. **Idiomatic Go**: Follow [Effective Go](https://go.dev/doc/effective_go), use `gofmt`/`gofumpt`
2. **Zero dependencies**: Standard library only (no CGO, no external deps)
3. **Explicit error handling**: Return errors, don't panic except for truly unrecoverable states
4. **Interface design**: Small, focused interfaces (e.g., `Node` interface in dom/)
5. **Package naming**: Short, lowercase, single-word when possible

### Performance

- **Hot path awareness**: Tokenizer processes every character - minimize allocations
- **Buffer reuse**: Use `strings.Builder` for accumulation, reuse where possible
- **Avoid string slicing**: Use indices instead of creating substrings in tokenizer
- **Benchmark changes**: Run `just bench` after tokenizer/parser modifications

### Testing Requirements

1. **100% coverage mandatory** - every line and branch must be tested
2. **Table-driven tests** for multiple similar cases
3. **Parallel tests** where possible (`t.Parallel()`)
4. **html5lib-tests must pass** - authoritative source of HTML5 spec compliance

## Implementation Status & TODOs

**Completed:**

- Package structure, build system, CI
- html5lib-tests integration framework
- DOM node types, basic structure
- Encoding detection (partial)

**In Progress (check TODOs in code):**

- Tokenizer state machine (`tokenizer/tokenizer.go:99`)
- Tree builder insertion modes (`treebuilder/modes.go`)
- CSS selector parsing (`selector/selector.go:20`)
- Serializer (`serialize/serialize.go`)

**Reference Implementation:**
See `reference/JustGoHTML-python/` for Python version implementing complete spec

## Common Patterns

### Error Handling

```go
// Parse errors (non-fatal, recovered per spec)
type ParseError struct {
    Code    string
    Message string
    Line    int
    Column  int
}

// API errors (fatal, user-facing)
return nil, &errors.SelectorError{
    Selector: selector,
    Position: pos,
    Message:  "invalid selector",
}
```

### DOM Manipulation

```go
// Always use Node interface methods
element.AppendChild(child)      // Not: element.children = append(...)
element.InsertBefore(child, ref)
element.RemoveChild(child)
```

### Tokenizer Pattern

```go
// State machine returns next token or nil (keep processing)
func (t *Tokenizer) consumeNext() *Token {
    switch t.state {
    case DataState:
        // ... state logic
        return &Token{Type: CharacterToken, Data: "x"}
    }
    return nil  // No token yet, keep processing
}
```

## Spec Compliance

1. **Follow WHATWG HTML spec exactly** - no heuristics or shortcuts
2. **Cite spec sections** in comments when implementing complex logic (e.g., "Per §13.2.5.72")
3. **html5lib-tests are canonical** - if tests fail, implementation is wrong (not tests)
4. **Error recovery per spec** - malformed HTML must parse exactly like browsers

## Resources

- Spec: https://html.spec.whatwg.org/multipage/parsing.html
- Tests: https://github.com/html5lib/html5lib-tests
- Python reference: `reference/JustGoHTML-python/src/JustGoHTML/`
- Plan: `PLAN.md` - comprehensive rewrite roadmap
