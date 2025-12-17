# CLAUDE.md

This file provides guidance to AI agents when working with code in this repository.

## Project Overview

JustGoHTML is a pure Go HTML5 parser implementing the WHATWG HTML5 specification. This is a rewrite of the Python JustGoHTML library (located in `reference/JustGoHTML-python/`) in Go with the following goals:

- **100% HTML5 Compliance**: Must pass all 9,000+ tests from the official html5lib-tests suite
- **Zero Dependencies**: Pure Go using only the standard library (no CGO)
- **Idiomatic Go**: Follow Go conventions and best practices
- **High Performance**: Target 5-10x improvement over the Python version
- **Simple API**: Clean, discoverable public interface

The project is currently in active development. See `PLAN.md` for the complete architecture and phased implementation plan.

## Essential Commands

### Development

```bash
just setup-deps    # Install all required formatters and tools (gofumpt, gci, golangci-lint, prettier, taplo, treefmt)
just setup         # Verify development environment is configured
just fmt           # Format all code using treefmt
just lint          # Run golangci-lint with strict settings
just lint-fix      # Auto-fix linting issues where possible
```

### Testing

```bash
just test          # Run all tests
just test-v        # Run tests with verbose output
just test-race     # Run tests with race detector
just test-coverage # Generate coverage report (target: 100%)
just test-spec     # Run html5lib-tests (when available)
```

### Building

```bash
just build         # Build the CLI binary
just build-all     # Build for all platforms (Linux, macOS, Windows)
just install       # Install CLI locally
just run [ARGS]    # Run CLI during development
```

### Validation

```bash
just check         # Run all checks (formatting, tests, linting)
just check-ci      # Comprehensive CI check with coverage
```

### Development Tools

```bash
just watch         # Watch for changes and run tests (requires fswatch or inotifywait)
just bench         # Run benchmarks
just tidy          # Tidy go modules
just clean         # Clean build artifacts
```

## Architecture

### Package Structure

```
JustGoHTML/
├── cmd/JustGoHTML/           # CLI entry point
├── internal/
│   ├── constants/          # HTML5 elements, entities, scopes
│   └── testutil/           # Test utilities
├── tokenizer/              # HTML5 tokenizer (state machine, ~49 states)
├── treebuilder/            # Tree construction (insertion modes, adoption agency)
├── dom/                    # DOM node types (Document, Element, Text, Comment)
├── selector/               # CSS selector parser and matcher
├── encoding/               # Encoding detection and decoding
├── serialize/              # HTML/text/markdown serialization
├── stream/                 # Streaming API
├── errors/                 # Parse errors and error codes
├── JustGoHTML.go             # Main public API (Parse, ParseBytes, ParseFragment)
└── options.go              # Configuration options
```

### Core Architecture Flow

1. **Input Processing**: `encoding` package detects encoding from BOM/meta charset/hints
2. **Tokenization**: `tokenizer` package implements HTML5 state machine (§8.2.4 of spec)
3. **Tree Construction**: `treebuilder` package constructs DOM tree using insertion modes (§8.2.5 of spec)
4. **DOM**: `dom` package provides the node tree with standard traversal methods
5. **Querying**: `selector` package parses CSS selectors and matches against DOM
6. **Output**: `serialize` package converts DOM back to HTML/text/markdown

### Key Design Principles

1. **Spec Compliance First**: Every algorithm follows WHATWG HTML5 spec exactly. Cite spec sections in comments (e.g., "Per §13.2.5.72")
2. **No Exceptions in Hot Paths**: The tokenizer and tree builder must never panic during parsing (HTML5 defines error recovery for everything)
3. **Zero Dependencies**: Only use Go standard library
4. **Minimal Allocations**: Pool tokens, reuse buffers where safe
5. **Deterministic Structures**: Avoid reflection, hasattr-style code; all structures are statically known

### Reference Implementation

The Python reference implementation is in `reference/JustGoHTML-python/`. When implementing features:

- Port the algorithm structure directly, but make it idiomatic Go
- Maintain the same test coverage (100%)
- Keep the same public API shape (adapted for Go conventions)
- Use the Python code as a guide for edge cases and error handling

## Testing Requirements

### Coverage

- **100% coverage required** for all packages
- Use `just test-coverage` to verify
- All new code must include comprehensive tests

### html5lib-tests Compliance

The project must pass the official html5lib-tests suite:

- Tokenizer tests (~2,000 tests)
- Tree construction tests (1,743 tests)
- Serializer tests (~100 tests)
- Encoding tests (~50 tests)

Test files with `html5lib_test.go` suffix indicate html5lib integration tests.

### Testing Behavior

Tests are configured to fail immediately on mismatch or error.
- All 9,000+ tests run by default.
- Failures are reported via `t.Errorf` or `t.Fatalf`.
- Use standard Go test flags (e.g., `-run`) to focus on specific tests or files.
- To ignore known failures, use `t.Skip` in the test code if necessary, but the goal is to fix them.

### Test Patterns

- Use table-driven tests for state machine behavior
- Test error conditions explicitly (HTML5 parsing never fails, but may emit parse errors)
- Include benchmarks for performance-critical paths (tokenizer, selector matching)

## Code Style

### Formatting

- Use `just fmt` (runs treefmt) before committing
- golangci-lint enforces strict rules (see `.golangci.toml`)
- Line length managed by formatter
- Pre-commit hooks verify formatting and run tests

### Naming Conventions

- Short names acceptable in limited scopes: `sb` (strings.Builder), `dt` (DOCTYPE token)
- State constants use SCREAMING_SNAKE_CASE to match HTML5 spec (e.g., `TAG_OPEN`, `BEFORE_ATTRIBUTE_NAME`)
- Insertion modes use PascalCase (e.g., `InBody`, `AfterHead`)

### Comments

- Explain **why**, not **what**
- Reference HTML5 spec sections when implementing spec algorithms
- Document deviations from Python implementation when necessary
- Public APIs require GoDoc comments

### Error Handling

- Parse errors are collected, not fatal (HTML5 error recovery)
- Return `error` from public APIs, don't panic
- Support strict mode via `WithStrictMode()` option (fails on first parse error)
- Internal errors (e.g., impossible states) should panic with clear messages

## Implementation Notes

### Tokenizer State Machine

- 49 states defined in `tokenizer/states.go`
- Main loop in `tokenizer.go` processes input character by character
- Character references (entities) handled specially
- Last start tag tracked for appropriate end tag matching

### Tree Builder Insertion Modes

- 21 insertion modes implementing HTML5 tree construction
- Stack of open elements tracked in `openElements`
- Active formatting elements list for adoption agency algorithm
- Template insertion modes stack for `<template>` handling
- Fragment parsing context for innerHTML-style parsing

### DOM Implementation

- `Node` interface with concrete types: Element, Text, Comment, Document
- Attributes stored in order-preserving structure (`dom/attributes.go`)
- Case-insensitive attribute lookup for HTML namespace
- Template elements have special `TemplateContent` field (DocumentFragment)

### CSS Selectors

- Full CSS selector parsing and matching
- Supports: type, ID, class, attribute selectors, pseudo-classes, combinators
- Pseudo-classes: `:first-child`, `:nth-child()`, `:not()`, `:empty`, etc.
- `Query()` returns all matches, `QueryFirst()` returns first match

## Common Development Tasks

### Running a Single Test

```bash
go test ./tokenizer -run TestSpecificFunction -v
```

### Adding a New Parse Error

1. Add error code to `errors/codes.go`
2. Add message template to `errors/messages.go`
3. Emit error in tokenizer/tree builder with line/column info

### Implementing a New Tokenizer State

1. Add state constant to `tokenizer/states.go`
2. Implement state handler function following spec algorithm
3. Add state transition logic
4. Add tests covering all code paths

### Adding CSS Selector Support

1. Extend AST types in `selector/ast.go`
2. Add parsing logic in `selector/parser.go`
3. Implement matching in `selector/matcher.go`
4. Add comprehensive tests

## Current Status

The project is in Phase 2-4 of the implementation plan:

- ✅ Phase 1: Foundation (encoding, errors, constants)
- 🚧 Phase 2: Tokenizer (state machine in progress)
- 🚧 Phase 3: Tree Builder (scaffolding complete, modes in progress)
- 🚧 Phase 4: DOM & Selectors (scaffolding complete)
- ⏳ Phase 5: Public API & CLI (scaffolding exists)
- ⏳ Phase 6: Testing & Validation
- ⏳ Phase 7: Documentation & Release

Check `PLAN.md` for detailed task breakdown and progress tracking.

## Performance Considerations

### Optimization Guidelines

1. **Profile before optimizing**: Use `just bench` to identify bottlenecks
2. **Tokenizer hot path**: Minimize allocations, use string scanning (IndexByte, IndexAny) instead of regex
3. **Token pooling**: Consider `sync.Pool` for token reuse during parsing
4. **String interning**: May benefit from interning common tag/attribute names
5. **Selector caching**: Pre-compiled selectors can be reused

### Benchmarking

- Add benchmarks for new features in `*_test.go` files
- Target: Parse Wikipedia homepage in milliseconds
- Aim for 5-10x performance improvement over Python version

## Contribution Guidelines

See `CONTRIBUTING.md` for detailed contribution guidelines. Key points:

- All changes require tests with 100% coverage
- Pre-commit hooks enforce formatting and tests
- Follow WHATWG HTML5 spec exactly
- Reference spec sections in complex algorithms
- Keep public API simple and discoverable
