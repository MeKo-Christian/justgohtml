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
  - Implementation: `tokenizer/tokenizer.go:35-66` (pool setup), `tokenizer/tokenizer.go:246` (Next returns \*Token), all emit functions updated
  - Tests: `tokenizer/pool_test.go` (TestTokenPoolReuse, TestTokenPoolReset)

- [x] **3.2.2 ASCII fast path for tokenization** ✅
  - Detect ASCII-only input upfront in `reset()` function
  - Use byte-based operations for ASCII content (avoids rune conversion overhead)
  - Fall back to rune-based for non-ASCII (graceful degradation)
  - **Actual results: ~3% speedup for ASCII HTML** (28,457 ns vs 29,345 ns)
  - **Implementation:**
    - Added `isASCIIOnly` and `inputBytes` fields to Tokenizer struct
    - ASCII detection in `reset()` with single-pass scan
    - `getCharASCII()` - byte-indexed character reading
    - `appendTextRune()` - uses `WriteByte()` for ASCII
    - `consumeCaseInsensitiveASCII()` - simple arithmetic for case folding (no `unicode.ToLower()`)
    - ASCII helper functions: `isASCIIWhitespace()`, `isASCIIAlpha()`, `toASCIILower()`
  - **Files modified:**
    - `tokenizer/tokenizer.go` - Core implementation
    - `tokenizer/ascii_test.go` - ASCII detection tests (NEW)
    - `tokenizer/ascii_bench_test.go` - Performance benchmarks (NEW)
  - **Note:** Actual speedup is lower than expected 20-30% because:
    - Most hot paths already had ASCII optimizations (e.g., tag name lowercasing)
    - Rune-based operations are already quite fast in Go
    - The main benefit is in reduced allocations for future optimizations
    - This lays groundwork for task 3.3.1 (byte-based tokenization)

- [ ] **3.2.3 State machine dispatch table**
  - Replace large switch statement with function pointer array
  - Use direct indexing for state handler dispatch
  - Expected: 5-10% speedup in tokenizer hot loop
  - Location: `tokenizer/tokenizer.go:200-331`

### 3.3 Major Refactors (1-2 weeks each)

- [x] **3.3.1 Byte-based tokenization (string indexing instead of []rune)** ✅
  - Completely eliminated `buf []rune` field from Tokenizer struct
  - Use `utf8.DecodeRuneInString()` for UTF-8 character decoding on-the-fly
  - Direct byte indexing for ASCII content (via existing ASCII fast path)
  - **Actual results: Performance-neutral, no measurable speed or memory improvement**
    - Benchstat comparison shows -0.56% geomean speed change (within noise)
    - Memory usage identical: 100.3Ki for complex HTML (both branches)
    - Allocations identical: 1,596 for complex HTML (both branches)
  - **Implementation:**
    - Removed `[]rune` conversion entirely from `reset()`
    - Updated `getCharRune()` to use `utf8.DecodeRuneInString()` and `utf8.DecodeLastRuneInString()`
    - Updated `peek()` with ASCII fast path and UTF-8 fallback
    - Updated `consumeIf()` to use byte-based comparison (ASCII literals only)
    - Updated `consumeCaseInsensitiveRune()` to decode UTF-8 on-the-fly
    - BOM handling now uses UTF-8 byte sequence (0xEF 0xBB 0xBF) instead of rune check
  - **Files modified:**
    - `tokenizer/tokenizer.go` - Complete byte-based refactor
  - **Benefits:**
    - Cleaner architecture (no redundant string-to-rune conversion)
    - Maintains 100% HTML5 spec compliance (all 9,000+ tests pass)
    - No performance regression - safe to merge
  - **Why no improvement:**
    - Go compiler likely optimized `[]rune` conversion already
    - Real memory bottleneck is DOM tree structure, not tokenizer buffer
    - UTF-8 decoding overhead offsets theoretical savings

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
