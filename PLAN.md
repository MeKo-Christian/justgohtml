# JustGoHTML: Remaining Tasks

This document tracks the remaining work for the JustGoHTML Go rewrite.

## 1. Implementation Gaps

### 1.1 Tokenizer Optimization

- [ ] Use \`strings.IndexAny()\` or \`strings.IndexByte()\` for character scanning where safe to improve performance without breaking spec compliance.

### 1.2 Tree Builder Compliance

Resolve remaining html5lib tree-construction gaps:

- Adoption agency edge cases (\`nobr\`, nested formatting in \`tests23.dat\`, \`tests26.dat\`)
- RAWTEXT/RCDATA/comment handling around \`<script>\`, \`<style>\`, \`<title>\` (\`tests1.dat\`, \`tests5.dat\`, \`tests4.dat\`)
- Select mode content model edge cases (\`webkit02.dat\`)
- Foreign content table-tag collisions (SVG/MathML \`<tr>\` in \`svg.dat\`, \`math.dat\`)
- Foster parenting/table text placement (\`webkit02.dat\`)

## 2. Testing & Validation

### 2.1 Unit Test Coverage

Target: 100% for all packages.

| Package     | Current | Target | Status      |
| ----------- | ------- | ------ | ----------- |
| tokenizer   | 92.4%   | 100%   | Near target |
| treebuilder | 92.4%   | 100%   | Near target |
| dom         | 99.1%   | 100%   | Near target |
| selector    | 96.3%   | 100%   | Near target |
| encoding    | 96.2%   | 100%   | Near target |
| serialize   | 97.2%   | 100%   | Near target |
| stream      | 89.2%   | 100%   | Near target |

### 2.2 Integration & CLI Testing

- [x] API usage tests
- [x] CLI argument parsing tests
- [x] CLI output format tests
- [x] Stdin handling tests

### 2.3 Fuzz Testing

- [ ] Set up Go fuzzing for tokenizer
- [ ] Set up Go fuzzing for tree builder
- [ ] Set up Go fuzzing for selector parser

### 2.4 Benchmark Suite

- [x] Parse speed benchmark (simple, medium, complex HTML)
- [x] Memory allocation benchmark
- [x] Selector matching benchmark
- [x] Serialization benchmark (see `serialize/benchmark_test.go`)
- [x] Streaming benchmark (see `stream/stream_test.go`)
- [x] Comparison with other Go HTML parsers (`golang.org/x/net/html`, `goquery`)
  - See [BENCHMARK_RESULTS.md](BENCHMARK_RESULTS.md) for detailed comparison results
  - Run `go test -bench=. -benchmem` to execute all benchmarks

## 3. Performance Optimization (Phase 4)

### 3.1 Quick Wins (1-2 days each)

- [x] **3.1.1 String interning for tag/attribute names** ✅
  - Pre-allocate common tag names ("div", "span", "p", "a", etc.)
  - Pre-allocate common attribute names ("class", "id", "href", "src", etc.)
  - Use interning in tokenizer when creating tag/attr name strings
  - **Actual results: 17-40% speedup, zero allocation overhead for interning lookups**
  - Implementation: `internal/constants/intern.go`, `tokenizer/tokenizer.go:470`, `tokenizer/tokenizer.go:444`

- [x] **3.1.2 Attribute map pooling** ✅
  - Use `sync.Pool` for `currentTagAttrIndex` map allocations
  - Reset and reuse maps instead of creating new ones per tag
  - **Actual results: Reduced allocations, improved memory efficiency**
  - Implementation: `tokenizer/tokenizer.go:11-33` (pool setup), multiple allocation sites replaced with pooled maps

- [x] **3.1.3 Selector sibling iteration optimization** ✅
  - Avoid allocating sibling slices in `getElementSiblings()` and `getSiblingsOfSameType()`
  - Use direct iteration for first/last child checks
  - Inline counting for nth-child and nth-of-type selectors
  - **Actual results: 28-29% speedup for selector matching, zero allocations for simple position checks**
  - Implementation: `selector/matcher.go:340-557` (optimized all sibling iteration functions)

### 3.2 Medium Effort (3-5 days each)

- [x] **3.2.1 Token pooling** ❌ REJECTED - Performance Regression
  - Attempted implementation with `sync.Pool` for token objects
  - Changed Token API to use pointers (`*Token`) throughout
  - Pool tokens during parsing with automatic lifecycle management
  - **Actual results: SIGNIFICANT PERFORMANCE REGRESSION**
    - **20-30% slower** execution time (13,817 ns → 17,646 ns for simple HTML)
    - **60-75% MORE memory** usage (10,710 B → 18,621 B for simple HTML)
    - **40-50% MORE allocations** (172 → 252 for simple HTML)
  - **Root causes:**
    - Pointer indirection overhead on every token access
    - Degraded cache locality (pointers scattered vs contiguous value structs)
    - `sync.Pool` Get/Put overhead outweighs allocation savings for small structs
    - Pre-allocated `Attrs` slices increased base memory per token
  - **Conclusion:** Token pooling is counterproductive for this use case
  - Reference: PR #1 (closed), branch `feature/token-pooling` kept for reference
  - Implementation: `tokenizer/tokenizer.go:35-66` (pool setup), `tokenizer/tokenizer.go:246` (Next returns *Token), all emit functions updated
  - Tests: `tokenizer/pool_test.go` (TestTokenPoolReuse, TestTokenPoolReset)

- [x] **3.2.2 ASCII fast path for tokenization** ❌ DEFERRED - Part of Failed Byte-Based Refactor
  - Attempted ASCII detection with byte-based operations for ASCII content
  - Implemented fallback to UTF-8 decoding for non-ASCII
  - **Actual results: PERFORMANCE REGRESSION (tested as part of 3.3.1)**
    - **12% slower** on complex HTML (151µs → 170µs)
    - Memory improved 18% but speed loss unacceptable
  - **Root causes:**
    - UTF-8 decoding overhead (`utf8.DecodeRuneInString`) on every character access
    - `peek()` function became expensive (walks UTF-8 runes)
    - ASCII detection loop in `reset()` adds overhead
    - Go compiler already optimizes `[]rune(string)` conversions well
  - **Conclusion:** ASCII fast path adds complexity without performance benefit
  - Reference: branch `feat/byte-based-tokenization` (not merged)

- [x] **3.2.3 State machine dispatch table** ✅ SUCCESSFUL
  - Replaced large switch statement (~55 cases) with function pointer array
  - Used direct array indexing for O(1) state handler dispatch
  - Dispatch table initialized once per tokenizer with all state handlers
  - **Actual results: SIGNIFICANT SPEEDUP**
    - **ParseBytes_Medium: 32% faster** (127.62µs → 86.68µs)
    - **ParseBytes_Complex: 9.2% faster** (167.1µs → 151.8µs)
    - **Parse_Parallel: 14% faster** (113.39µs → 97.58µs)
    - **Geometric mean: 12.9% faster** across all benchmarks
    - Memory overhead: ~2.3% more (dispatch table allocation per tokenizer)
    - Allocation overhead: ~0.2% more (1 extra alloc per parse)
  - **Why it works:** Direct array lookup eliminates switch comparison overhead
  - Implementation: `tokenizer/tokenizer.go:35-36` (type), `tokenizer/tokenizer.go:122-191` (init), `tokenizer/tokenizer.go:307-318` (step)

### 3.2.4 New Optimization Opportunities (avoiding past mistakes)

Based on lessons learned from failed optimizations (3.2.1, 3.2.2, 3.3.1), here are optimization opportunities that work WITH Go's strengths rather than against them:

- [ ] **3.2.4.1 Pre-computed rune literals for consumeIf/consumeCaseInsensitive**
  - Currently: `consumeIf("--")` calls `[]rune(lit)` on every invocation (lines 541-555)
  - Fix: Pre-compute rune slices for known literals ("--", "DOCTYPE", "[CDATA[", "PUBLIC", "SYSTEM")
  - Use package-level `var` with pre-converted rune slices
  - **Why this won't fail like 3.3.1:** No per-character overhead, just avoids repeated conversions
  - Expected: 2-5% speedup on documents with many comments/doctypes/CDATA

- [ ] **3.2.4.2 Batch text node emission with strings.Builder capacity hints**
  - Currently: `textBuffer` grows dynamically per WriteRune (line 82, 398-403)
  - Fix: Pre-size Builder with `Grow()` based on remaining input estimate
  - After `flushText()`, reuse capacity hint from previous text length
  - **Why this won't fail:** Reduces reallocation, not adding overhead
  - Expected: 3-7% speedup for text-heavy documents

- [ ] **3.2.4.3 Eliminate pendingTokens slice operations in hot path**
  - Currently: `Next()` does `t.pendingTokens[0]` then `t.pendingTokens[1:]` (lines 293-304)
  - Fix: Use ring buffer or index-based approach instead of slice reslicing
  - Most tokens emit one at a time; avoid slice header updates
  - **Why this won't fail:** Reduces GC pressure, no pointer indirection added
  - Expected: 5-10% speedup

- [ ] **3.2.4.4 Inline hot path character classification**
  - Currently: `switch c { case '\t', '\n', '\f', ' ': ... }` in multiple places
  - Fix: Create lookup table `var isWhitespace [256]bool` for ASCII range
  - Use `if c < 256 && isWhitespace[c]` for fast classification
  - Also: `isASCIIAlpha`, `isASCIIUpper` tables
  - **Why this won't fail:** Tables are read-only, excellent cache behavior
  - Expected: 3-8% speedup in tag parsing states

- [ ] **3.2.4.5 Reduce attribute map operations**
  - Currently: Every attribute does `t.currentTagAttrIndex[name] = struct{}{}` (line 447)
  - Fix: Only track duplicates for tags with >1 attribute (common case: 0-3 attrs)
  - Use simple slice scan for small attribute counts, map only when >4 attrs
  - **Why this won't fail:** Reduces map overhead for common case
  - Expected: 5-10% speedup for attribute-heavy documents

**Priority order (highest impact first):** 3.2.4.3, 3.2.4.4, 3.2.4.5, 3.2.4.2, 3.2.4.1

### 3.3 Major Refactors (1-2 weeks each)

- [x] **3.3.1 Byte-based tokenization (string indexing instead of []rune)** ❌ REJECTED - Performance Regression
  - Replaced `buf []rune` with direct string indexing and UTF-8 decoding
  - Implemented `utf8.DecodeRuneInString()` for character-by-character parsing
  - Added ASCII-only detection for fast path optimization
  - **Actual results: UNACCEPTABLE SPEED REGRESSION**
    - **12% slower** on complex HTML (151µs → 170µs) - opposite of expected 30-40% speedup
    - **18% memory reduction** (100.3Ki → 82.3Ki) - good but doesn't justify speed loss
    - **Throughput impact:** 6,350 pages/sec → 5,900 pages/sec
  - **Root causes:**
    - UTF-8 decoding overhead on every `getChar()` call outweighs memory savings
    - Go's compiler already optimizes `[]rune(string)` conversion efficiently
    - `peek()` became expensive (must walk UTF-8 sequences for lookahead)
    - ASCII detection scan in `reset()` adds overhead for every parse
  - **Conclusion:** Premature optimization - theory didn't match reality. Keep current `[]rune` approach.
  - Reference: branch `feat/byte-based-tokenization` (benchmarked, not merged)
  - **Note:** This was THOUGHT to be the biggest optimization opportunity - benchmarks proved otherwise

- [ ] **3.3.2 DOM element pooling**
  - Implement `sync.Pool` for DOM element allocations
  - Careful lifecycle management to avoid pool contamination
  - Pool `Element`, `Text`, `Comment` node types
  - Expected: 10-15% allocation reduction
  - Location: `dom/element.go:34-42`

### 3.4 Performance Validation

- [ ] Run full benchmark suite after each optimization
- [ ] Compare against baseline (current performance)
- [ ] Validate no regression in html5lib test compliance
- [ ] Profile with `pprof` to identify new bottlenecks
- [ ] Document performance improvements in BENCHMARK_RESULTS.md

**Expected Overall Impact (all optimizations):**

- Speed: 2-3x faster (competitive with x/net/html while maintaining 100% compliance)
- Memory: 50-60% less memory
- Allocations: 50-60% fewer allocations

## 4. Documentation & Release

### 4.1 Documentation

- [ ] README.md with quickstart
- [ ] GoDoc comments on all public types and functions
- [ ] Detailed guides: \`docs/quickstart.md\`, \`api.md\`, \`cli.md\`, \`selectors.md\`, \`streaming.md\`, \`encoding.md\`, \`errors.md\`, \`fragments.md\`

### 4.2 Release Preparation

- [ ] Semantic versioning (v0.1.0)
- [ ] CHANGELOG.md
- [ ] LICENSE file (MIT)
- [ ] CONTRIBUTING.md
- [ ] GitHub release automation
- [ ] Binary builds for Linux, macOS, Windows

### 4.3 CI/CD

- [ ] GitHub Actions workflow for tests, linting, and releases
- [ ] Dependabot configuration
- [ ] Badge setup in README
