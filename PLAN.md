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

This phase is now documented as a post-mortem: what did not work, why, and the lessons we keep. The optimizations that worked are assumed and not repeated here.

### 3.1 What Did Not Work (and why)

- **Token pooling (sync.Pool, pointer tokens)**: 20–30% slower, more memory, more allocs. Pointer indirection and pool overhead outweighed any reuse benefits; cache locality got worse.
- **ASCII fast path + byte tokenization**: 12% slower on complex HTML. UTF-8 decoding per rune and expensive `peek()` killed performance; Go already optimizes `[]rune(string)` well.
- **Precomputed rune literals for small matches**: No measurable change. These code paths are not hot; added complexity without measurable win.
- **Lookup tables vs switch for whitespace**: ~1% slower. Switches already compile to efficient jump tables; lookup adds memory indirection.
- **DOM node pooling (arena allocator)**: Fewer allocs (-26%) but **slower** (+6.2%) and **more memory** (+21.5%). Chunked arenas increased working set; allocations were not the limiting factor.

### 3.2 Lessons Learned (hard constraints)

- **Measure before and after**: No change without benchmarks and `benchstat`.
- **Hot path only**: Optimizing cold code is wasted time and risk.
- **Avoid pointer churn**: Value structs and contiguous memory win frequently in Go.
- **Be skeptical of pools**: `sync.Pool` and arenas can reduce allocs but often increase CPU and RSS.
- **Large refactors need early proof**: If a prototype regresses, stop and document.

### 3.3 Next-Step Candidates (profile-driven, not evaluated yet)

Profiler snapshot (BenchmarkJustGoHTML_Parse_Complex, 5s):
- tokenizer: `getChar`, `stateData`, `stateTagName`, `stateAttributeValueDoubleQuoted`
- string/rune conversion: `stringtoslicerune`, `slicerunetostring`, `encoderune`
- allocation churn: `mallocgc*`, `growslice`
- tree builder: `processInBody`, `hasElementInScopeInternal`
- DOM: `Element.AppendChild`

Candidates (hypotheses only; must benchmark):
- **Bulk text scanning in Data/RCDATA/RAWTEXT**: scan to next `<` or NUL and append spans in one go to reduce `getChar`/`WriteRune` churn and `encoderune`.
- **Attribute/value fast path**: reduce per-rune work in attribute states with span-based parsing (still honoring spec edge cases).
- **Tree builder scope checks**: `hasElementInScopeInternal` shows up; cache scope boundaries or track nearest markers to avoid repeated scans.
- **AppendChild growth behavior**: reduce `growslice` by pre-sizing children for common cases (e.g., when inserting known batches).
- **String/hash hot spots**: targeted fast path for the most common tag/attr names to reduce `mapaccess1_faststr`/`aeshashbody` activity.

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
