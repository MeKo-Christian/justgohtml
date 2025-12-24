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

- [x] **3.2.4.1 Pre-computed rune literals for consumeIf/consumeCaseInsensitive** ❌ REJECTED - No Measurable Impact
  - Pre-computed rune slices for literals: `"--"`, `"DOCTYPE"`, `"[CDATA["`, `"PUBLIC"`, `"SYSTEM"`
  - Changed function signatures to accept `[]rune` instead of `string`
  - **Actual results: NO PERFORMANCE IMPROVEMENT**
    - Parse_Simple: ~16.25µs vs ~16.28µs (p=0.684, not significant)
    - Parse_Medium: ~96.20µs vs ~78.73µs (p=0.218, not significant)
    - Parse_Complex: ~147.3µs vs ~162.2µs (p=0.190, not significant)
    - Geomean: +3.47% (within statistical noise)
  - **Why it failed:** Functions not in hot path - only called for rare doctypes/comments/CDATA
  - **Conclusion:** `[]rune()` overhead for 2-7 char strings is negligible; optimization adds complexity for zero gain
  - Reference: PR #4 (closed), branch `feat/precomputed-rune-literals` kept for reference

- [x] **3.2.4.2 Batch text node emission with strings.Builder capacity hints** ✅ SUCCESSFUL
  - Added `textBufferHint` field to track capacity for next text node
  - Pre-grow buffer on first character with `Grow(textBufferHint)` (line 423)
  - After `flushText()`, save text length as hint for next node (line 454)
  - Initialize with default 64-byte hint on reset (line 243)
  - **Actual results: MODERATE IMPROVEMENT**
    - **Speed: 1.17% faster** (geomean across all benchmarks)
    - **Complex documents: 3.21-4.19% faster** (where text processing matters most)
    - **Memory: 0.36% less** (small reduction in reallocations)
    - **Allocations: 7.61% fewer** (geomean, up to 10% reduction on complex HTML)
  - **Why it worked:** Pre-allocating based on previous text size eliminates most Builder reallocations
  - **Best impact on:** Text-heavy documents (Medium: -3.21%, Complex: -4.19%)
  - Implementation: `tokenizer/tokenizer.go:160` (field), `tokenizer/tokenizer.go:423` (pre-grow), `tokenizer/tokenizer.go:438-454` (hint update)

- [x] **3.2.4.3 Eliminate pendingTokens slice operations in hot path** ✅ SUCCESSFUL
  - Replaced `pendingTokens []Token` slice with fixed-size ring buffer `[4]Token`
  - Added `pendingHead`, `pendingTail`, `pendingCount` indices for O(1) operations
  - Used bitwise AND (`& 3`) for efficient modulo-4 wraparound
  - **Actual results: SIGNIFICANT IMPROVEMENT**
    - **Speed: 11-35% faster** (geomean -11.42%)
    - **Memory: 36-44% less** (geomean -33.24%)
    - **Allocations: 19-23% fewer** (geomean -16.83%)
  - **Why it worked:** Eliminated slice header updates on every token consumption
  - Implementation: `tokenizer/tokenizer.go:163-166` (struct), `tokenizer/tokenizer.go:298-314` (Next), `tokenizer/tokenizer.go:387-390` (emit)

- [x] **3.2.4.4 Inline hot path character classification** ❌ REJECTED - No Performance Benefit
  - Created lookup tables in `internal/constants/charclass.go`:
    - `isWhitespace[256]bool`, `isASCIIUpper[256]bool`, `isASCIILower[256]bool`
    - `isASCIIAlpha[256]bool`, `isASCIIAlphaNum[256]bool`
    - Helper functions: `IsWhitespace(c)`, `IsASCIIUpper(c)`, `IsASCIIAlpha(c)`, `ToLower(c)`
  - **Tested optimizations:**
    - Replaced `unicode.ToLower(c)` with `constants.ToLower(c)` (4 locations)
    - Replaced `(c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')` with `constants.IsASCIIAlpha(c)` (4 locations)
    - One whitespace switch-to-lookup replacement in `stateTagName()` (1 location)
    - **Extended test**: Converted all 15 whitespace switch cases to lookup checks
  - **Actual results: NO PERFORMANCE BENEFIT**
    - Initial partial optimization: ~0-5% (p>0.05, not significant)
    - **Full whitespace optimization (15 conversions)**: +1.1% slower (p=0.089, not significant)
      - Baseline: 1.126ms ± 4%
      - Optimized: 1.139ms ± 2%
    - Memory: No change
    - Allocations: No change
  - **Why lookup tables didn't help:**
    - Modern CPUs optimize switch statements excellently (branch prediction, jump tables)
    - Lookup table introduces memory indirection that offsets any theoretical gains
    - Switch statements on simple character ranges already compile to efficient code
    - Character classification is not the bottleneck in tokenization
  - **Conclusion:** Lookup table approach rejected - switch statements are already optimal
  - **Infrastructure retained:** Lookup tables remain in codebase for potential use in other contexts
  - Implementation: `internal/constants/charclass.go` (tables and tests)

  **Subtasks (completed but REJECTED due to no performance benefit):**

  - [x] **3.2.4.4.1 Convert remaining whitespace switch cases to lookup checks** ✅ COMPLETED → ❌ REVERTED
    - Converted all 15 whitespace switch cases to `constants.IsWhitespace(c)` lookups
    - All tests passed with no regressions
    - **Result**: Changes reverted - no performance benefit achieved

  - [x] **3.2.4.4.2 Optimize character classification in attribute parsing states** ✅ COMPLETED (covered by 3.2.4.4.1) → ❌ REVERTED
    - Covered by the 15 conversions in 3.2.4.4.1
    - **Result**: Changes reverted along with 3.2.4.4.1

  - [x] **3.2.4.4.3 Benchmark and validate complete implementation** ✅ COMPLETED
    - Ran comprehensive benchmarks: `-benchtime=10s -count=10` (10 samples)
    - **Results**: NO PERFORMANCE IMPROVEMENT
      - Baseline: 1.126ms ± 4% (n=10)
      - Optimized: 1.139ms ± 2% (n=10)
      - Change: +1.1% **SLOWER** (p=0.089, NOT statistically significant)
    - **Decision**: Optimization rejected and reverted

  - [x] **3.2.4.4.4 Consider additional lookup table optimizations** ❌ SKIPPED
    - Not pursued since 3.2.4.4.1-3 showed lookup tables don't improve performance

- [ ] **3.2.4.5 Reduce attribute map operations**
  - Currently: Every attribute does `t.currentTagAttrIndex[name] = struct{}{}` (line 447)
  - Fix: Only track duplicates for tags with >1 attribute (common case: 0-3 attrs)
  - Use simple slice scan for small attribute counts, map only when >4 attrs
  - **Why this won't fail:** Reduces map overhead for common case
  - Expected: 5-10% speedup for attribute-heavy documents

**Priority order (highest impact first):** 3.2.4.5 (remaining)

**Completed (in order of impact):** 3.2.4.3 (11% faster), 3.2.4.2 (4% faster on complex), 3.2.4.4 (partial - <1% measured), 3.2.4.1 (rejected - no impact)

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
