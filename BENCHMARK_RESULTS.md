# Benchmark Results

**Test Environment:**

- **OS:** Linux (amd64)
- **CPU:** 12th Gen Intel(R) Core(TM) i7-1255U (12 logical cores)
- **Go Version:** 1.24.1
- **Benchmark Time:** 5 seconds per benchmark
- **Date:** 2025-12-22
- **Optimizations Applied:** String interning (3.1.1), Attribute map pooling (3.1.2), Selector sibling iteration (3.1.3), State machine dispatch table (3.2.3)

## Executive Summary

JustGoHTML provides **100% HTML5 compliance** with **competitive performance** compared to other Go HTML parsers. After completing Phase 3.1 and 3.2.3 optimizations (string interning, attribute map pooling, selector sibling iteration, and state machine dispatch table), JustGoHTML has significantly closed the performance gap with `golang.org/x/net/html` and `goquery` while maintaining full WHATWG specification compliance.

### Key Findings

- **Parse Speed:** JustGoHTML is now 2.2-2.4x slower than x/net/html while providing 100% spec compliance vs ~70%
- **Phase 3.2.3 Dispatch Table:** Additional 5-17% speedup on top of Phase 3.1 improvements
- **Query Speed:** JustGoHTML's CSS selector matching is now highly competitive - faster than goquery for many queries
- **Memory Usage:** JustGoHTML uses more memory due to complete spec compliance and richer DOM, but reduced allocations through pooling
- **Parallel Performance:** All parsers scale well with parallelism

## Detailed Results

### Parsing Benchmarks

#### Simple HTML (Small Document)

| Parser                  | Time/op      | Speed vs JustGoHTML | Mem/op   | Allocs/op | Improvement            |
| ----------------------- | ------------ | ------------------- | -------- | --------- | ---------------------- |
| **JustGoHTML**          | 14,020 ns/op | 1.0x (baseline)     | 11,446 B | 173       | ⚡ **+6% from 3.2.3**  |
| `golang.org/x/net/html` | 7,088 ns/op  | **2.0x faster**     | 7,880 B  | 48        | -                      |
| `goquery`               | 6,414 ns/op  | **2.2x faster**     | 7,960 B  | 51        | -                      |

#### Medium HTML (Blog Post ~3KB)

| Parser                  | Time/op      | Speed vs JustGoHTML | Mem/op   | Allocs/op | Improvement            |
| ----------------------- | ------------ | ------------------- | -------- | --------- | ---------------------- |
| **JustGoHTML**          | 89,190 ns/op | 1.0x (baseline)     | 63,172 B | 968       | ⚡ **+17% from 3.2.3** |
| `golang.org/x/net/html` | 40,730 ns/op | **2.2x faster**     | 24,320 B | 281       | -                      |
| `goquery`               | 37,370 ns/op | **2.4x faster**     | 24,400 B | 284       | -                      |

#### Complex HTML (Full Page ~5KB)

| Parser                  | Time/op       | Speed vs JustGoHTML | Mem/op    | Allocs/op | Improvement            |
| ----------------------- | ------------- | ------------------- | --------- | --------- | ---------------------- |
| **JustGoHTML**          | 151,800 ns/op | 1.0x (baseline)     | 103,444 B | 1,597     | ⚡ **+5% from 3.2.3**  |
| `golang.org/x/net/html` | 64,710 ns/op  | **2.3x faster**     | 38,048 B  | 504       | -                      |
| `goquery`               | 61,680 ns/op  | **2.5x faster**     | 38,128 B  | 507       | -                      |

### Query Benchmarks

#### Simple Query (`div.feature`)

| Parser         | Time/op     | Speed vs JustGoHTML | Mem/op | Allocs/op | Improvement            |
| -------------- | ----------- | ------------------- | ------ | --------- | ---------------------- |
| **JustGoHTML** | 2,338 ns/op | 1.0x (baseline)     | 696 B  | 25        | ⚡ **+18% from 3.2.3** |
| `goquery`      | 3,097 ns/op | 0.8x (slower)       | 360 B  | 15        | -                      |

#### Complex Query (`section > h2 + div.feature-grid div[data-feature-id]`)

| Parser         | Time/op     | Speed vs JustGoHTML | Mem/op  | Allocs/op | Improvement      |
| -------------- | ----------- | ------------------- | ------- | --------- | ---------------- |
| **JustGoHTML** | 3,417 ns/op | 1.0x (baseline)     | 1,680 B | 28        | ⚡ **+8% total** |
| `goquery`      | 4,804 ns/op | 0.7x (slower)       | 744 B   | 27        | -                |

### Parallel Performance

Performance when running with multiple goroutines (GOMAXPROCS=12):

| Parser                  | Time/op       | Mem/op    | Allocs/op |
| ----------------------- | ------------- | --------- | --------- |
| **JustGoHTML**          | 97,580 ns/op  | 103,527 B | 1,597     |
| `golang.org/x/net/html` | 129,640 ns/op | 38,047 B  | 504       |
| `goquery`               | 127,700 ns/op | 38,127 B  | 507       |

### Memory Allocations

Comparison of memory allocations for complex HTML parsing:

| Parser                  | Bytes Allocated | Number of Allocations |
| ----------------------- | --------------- | --------------------- |
| **JustGoHTML**          | 108,344 B       | 1,603                 |
| `golang.org/x/net/html` | 38,048 B        | 504                   |
| `goquery`               | 38,128 B        | 507                   |

## Analysis

### Performance Optimizations Applied

All Phase 3.1 "Quick Wins" optimizations have been completed:

#### 1. String Interning for Tag/Attribute Names (Task 3.1.1)

Implemented string interning to reduce memory allocations for common HTML tag and attribute names:

- **90+ pre-allocated common tag names** (div, span, p, a, etc.)
- **60+ pre-allocated common attribute names** (class, id, href, src, etc.)
- **Zero allocation overhead**: Map lookups take ~6ns with 0 allocations
- **Results**: 17-40% speedup across parsing benchmarks
- **Implementation**: [internal/constants/intern.go](internal/constants/intern.go)

#### 2. Attribute Map Pooling (Task 3.1.2)

Implemented `sync.Pool` for attribute map allocations to eliminate repeated allocations during tokenization:

- **Pooled attribute index maps**: Reuse maps instead of allocating for each tag
- **Smart cleanup**: Maps are cleared before reuse to prevent data leakage
- **Pre-allocated capacity**: Pool maintains maps with capacity 8 for typical attribute count
- **Results**: Reduced allocations and improved memory efficiency during parsing
- **Implementation**: [tokenizer/tokenizer.go:11-33](tokenizer/tokenizer.go#L11-L33)

#### 3. Selector Sibling Iteration Optimization (Task 3.1.3)

Optimized CSS selector matching to avoid allocating sibling slices:

- **Zero-allocation position checks**: `:first-child`, `:last-child`, `:only-child` no longer build full sibling lists
- **Inline counting**: nth-child and nth-of-type selectors count during iteration
- **Early exit logic**: Functions return immediately when answer is determined
- **Results**: 28-39% faster selector matching, up to 76% total improvement for complex queries
- **Implementation**: [selector/matcher.go:340-557](selector/matcher.go#L340-L557)

### Why JustGoHTML is Still Slower

JustGoHTML's remaining performance gap is due to intentional trade-offs for **100% HTML5 specification compliance**:

1. **Complete Error Recovery**: Implements all HTML5 error recovery rules exactly as browsers do
2. **Proper Adoption Agency Algorithm**: Handles complex cases like misnested formatting elements
3. **Template Element Support**: Full support for `<template>` with separate document fragments
4. **Richer DOM Model**: More complete node types and relationships
5. **Strict Spec Compliance**: No shortcuts or approximations

### Performance in Context

- **~152 µs for complex HTML** means JustGoHTML can now parse **~6,600 pages per second**
- **5-17% additional speedup** from Phase 3.2.3 dispatch table optimization
- **CSS selector matching faster than goquery** for many queries
- For typical web scraping or HTML processing, this performance is more than adequate
- The 100% spec compliance means you get the **same result as a browser would**

### When to Use Each Parser

**Use JustGoHTML when:**

- You need 100% HTML5 specification compliance
- You're processing browser-rendered HTML and need exact browser behavior
- You need CSS selector support built-in
- You want zero dependencies (stdlib only)
- You're parsing malformed or unusual HTML

**Use `golang.org/x/net/html` when:**

- You need maximum speed and ~70% compliance is acceptable
- You're parsing simple, well-formed HTML
- You don't need CSS selectors
- You want the absolute minimal footprint

**Use `goquery` when:**

- You need CSS selectors with good performance
- You're okay with ~70% HTML5 compliance
- You want jQuery-like syntax
- Performance is critical and spec compliance is not

## Additional Benchmarks

### Serialization Performance

Serialization benchmarks measure how fast JustGoHTML can convert DOM trees back to HTML strings:

| Benchmark         | Time/op       | Mem/op   | Allocs/op |
| ----------------- | ------------- | -------- | --------- |
| Simple HTML       | 1,401 ns/op   | 744 B    | 25        |
| Medium HTML       | 12,233 ns/op  | 5,880 B  | 138       |
| Complex HTML      | 19,004 ns/op  | 13,656 B | 249       |
| Pretty Printing   | 15,503 ns/op  | 10,992 B | 179       |
| Large Text (10KB) | 104,574 ns/op | 56,832 B | 18        |

**Key Findings:**

- Serialization is very fast: ~1.4 µs for simple HTML, ~19 µs for complex pages
- Pretty printing adds minimal overhead (~3-4 µs)
- Scales linearly with content size
- Excellent performance for round-trip parsing + serialization

Run serialization benchmarks:

```bash
go test -bench=BenchmarkToHTML -benchmem ./serialize
```

### Streaming Performance

Streaming API benchmarks measure event-based parsing throughput:

| Benchmark       | Time/op       | Mem/op   | Allocs/op |
| --------------- | ------------- | -------- | --------- |
| Simple (String) | 29,280 ns/op  | 10,864 B | 159       |
| Simple (Bytes)  | 33,990 ns/op  | 11,168 B | 171       |
| Medium HTML     | 44,059 ns/op  | 18,352 B | 267       |
| Complex HTML    | 104,046 ns/op | 36,720 B | 502       |
| Parallel        | 7,180 ns/op   | 7,408 B  | 107       |
| Event Filtering | 20,668 ns/op  | 10,272 B | 142       |

**Key Findings:**

- Streaming is ideal for processing large documents with low memory footprint
- Excellent parallel performance (4x improvement)
- Event filtering adds minimal overhead
- Good for incremental processing and memory-constrained environments

Run streaming benchmarks:

```bash
go test -bench=BenchmarkStream -benchmem ./stream
```

## Future Optimizations

See [PLAN.md Phase 4](PLAN.md#3-performance-optimization-phase-4) for the complete optimization roadmap.

**Completed optimizations (Phase 3.1 + 3.2.3):**

- ✅ **String Interning** (Task 3.1.1): Intern common tag names and attribute names → 17-40% speedup achieved
- ✅ **Attribute Map Pooling** (Task 3.1.2): Use `sync.Pool` for attribute map allocations → Reduced allocations and improved memory efficiency
- ✅ **Selector Sibling Iteration** (Task 3.1.3): Optimize sibling iteration to avoid allocations → 70-76% selector speedup achieved
- ✅ **State Machine Dispatch Table** (Task 3.2.3): Function pointer array for state dispatch → 5-17% additional speedup achieved

**Rejected optimizations (proven counterproductive):**

- ❌ **Token Pooling** (Task 3.2.1): Caused 20-30% slowdown due to pointer indirection overhead
- ❌ **ASCII Fast Path** (Task 3.2.2): Added complexity without measurable benefit
- ❌ **Byte-based Tokenization** (Task 3.3.1): 12% slower due to UTF-8 decoding overhead

**Remaining optimization opportunities:**

1. **DOM Element Pooling** (Task 3.3.2): Pool DOM node allocations (10-15% allocation reduction expected)
2. **Buffer Management**: Better buffer reuse in tokenizer
3. **Selector Caching**: Cache compiled selectors
4. **SIMD Optimizations**: Use SIMD for character scanning in hot paths

## Running These Benchmarks

To reproduce these benchmarks:

```bash
# Run all benchmarks
go test -bench=. -benchmem -benchtime=3s

# Run specific benchmark groups
go test -bench='Parse_Simple' -benchmem -benchtime=3s
go test -bench='Query' -benchmem -benchtime=3s
go test -bench='Parallel' -benchmem -benchtime=3s

# Compare with longer benchmark time
go test -bench=. -benchmem -benchtime=10s

# Generate CPU profile
go test -bench=BenchmarkJustGoHTML_Parse_Complex -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## Conclusion

JustGoHTML delivers on its promise of **100% HTML5 compliance** with **competitive performance**. After completing Phase 3.1 and 3.2.3 optimizations, JustGoHTML is now **2.2-2.5x slower** for parsing than parsers that sacrifice spec compliance, while parsing at **~6,600 complex pages per second**. The performance overhead buys you guaranteed browser-compatible parsing behavior, which is essential for many applications.

**Phase 3.1 + 3.2.3 achievements:**

- ✅ **5-17% additional speedup** from dispatch table optimization (on top of Phase 3.1 gains)
- ✅ **CSS selector matching faster than goquery** - simple queries in ~2.3 µs, complex in ~3.4 µs
- ✅ **Reduced memory allocations** through pooling and inline counting
- ✅ **Zero allocation overhead** for tag/attribute name lookups and simple position checks
- ✅ Significantly narrowed the performance gap with x/net/html while maintaining 100% compliance

**Optimization learnings:**

Through rigorous benchmarking, we discovered that several "obvious" optimizations actually hurt performance:

- Token pooling adds pointer indirection overhead
- Byte-based tokenization is slower than Go's optimized `[]rune` conversions
- ASCII fast paths add complexity without benefit

The successful optimizations (string interning, map pooling, sibling iteration, dispatch table) share a common pattern: they reduce work without adding indirection.

For applications that need exact browser behavior (HTML sanitizers, browser automation tools, testing frameworks), JustGoHTML's combination of performance and compliance is ideal. For simple parsing where ~70% compliance is acceptable, `x/net/html` or `goquery` remain excellent choices.
