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

- [ ] **3.1.2 Attribute map pooling**
  - Use `sync.Pool` for `currentTagAttrIndex` map allocations
  - Reset and reuse maps instead of creating new ones per tag
  - Expected: 15-20% reduction in allocations
  - Location: `tokenizer/tokenizer.go:32`, `tokenizer/tokenizer.go:109`

- [ ] **3.1.3 Selector sibling iteration optimization**
  - Avoid allocating sibling slices in `getElementSiblings()` and `getSiblingsOfSameType()`
  - Use direct iteration for first/last child checks
  - Cache sibling lists when needed multiple times
  - Expected: 15-20% speedup for selector matching
  - Location: `selector/matcher.go:278-291`, `selector/matcher.go:322-338`

### 3.2 Medium Effort (3-5 days each)

- [ ] **3.2.1 Token pooling**
  - Implement `sync.Pool` for token objects
  - Pool tokens during parsing and return to pool after tree builder consumes them
  - Expected: 20-30% allocation reduction
  - Location: `tokenizer/tokenizer.go:391`

- [ ] **3.2.2 ASCII fast path for tokenization**
  - Detect ASCII-only input upfront
  - Use byte-based operations for ASCII content (avoids rune conversion overhead)
  - Fall back to rune-based for non-ASCII
  - Expected: 20-30% speedup for ASCII-heavy HTML
  - Location: `tokenizer/tokenizer.go:88-97`

- [ ] **3.2.3 State machine dispatch table**
  - Replace large switch statement with function pointer array
  - Use direct indexing for state handler dispatch
  - Expected: 5-10% speedup in tokenizer hot loop
  - Location: `tokenizer/tokenizer.go:200-331`

### 3.3 Major Refactors (1-2 weeks each)

- [ ] **3.3.1 Byte-based tokenization (string indexing instead of []rune)**
  - Replace `buf []rune` with direct string indexing
  - Use `utf8.DecodeRuneInString()` for character-by-character parsing
  - Eliminates 3x memory overhead of rune slice conversion
  - Expected: 30-40% speedup, 50% memory reduction
  - Location: `tokenizer/tokenizer.go:16`, `tokenizer/tokenizer.go:96`
  - **Note:** This is the single biggest optimization opportunity

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
