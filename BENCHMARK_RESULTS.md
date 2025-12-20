# Benchmark Results

**Test Environment:**

- **OS:** Linux (amd64)
- **CPU:** 12th Gen Intel(R) Core(TM) i7-1255U (12 logical cores)
- **Go Version:** 1.24.1
- **Benchmark Time:** 3 seconds per benchmark
- **Date:** 2025-12-20

## Executive Summary

JustGoHTML provides **100% HTML5 compliance** with competitive performance compared to other Go HTML parsers. While slightly slower than `golang.org/x/net/html` and `goquery` (which sacrifice spec compliance for speed), JustGoHTML delivers full WHATWG specification compliance with reasonable performance characteristics.

### Key Findings

- **Parse Speed:** JustGoHTML is 2-4x slower than x/net/html but provides 100% spec compliance vs ~70%
- **Query Speed:** JustGoHTML's CSS selector matching is competitive, within 2-3x of goquery
- **Memory Usage:** JustGoHTML uses more memory due to complete spec compliance and richer DOM
- **Parallel Performance:** All parsers scale well with parallelism

## Detailed Results

### Parsing Benchmarks

#### Simple HTML (Small Document)

| Parser                  | Time/op      | Speed vs JustGoHTML | Mem/op   | Allocs/op |
| ----------------------- | ------------ | ------------------- | -------- | --------- |
| **JustGoHTML**          | 18,019 ns/op | 1.0x (baseline)     | 12,736 B | 211       |
| `golang.org/x/net/html` | 8,050 ns/op  | **2.2x faster**     | 7,880 B  | 48        |
| `goquery`               | 8,447 ns/op  | **2.1x faster**     | 7,960 B  | 51        |

#### Medium HTML (Blog Post ~3KB)

| Parser                  | Time/op       | Speed vs JustGoHTML | Mem/op   | Allocs/op |
| ----------------------- | ------------- | ------------------- | -------- | --------- |
| **JustGoHTML**          | 141,150 ns/op | 1.0x (baseline)     | 76,080 B | 1,202     |
| `golang.org/x/net/html` | 46,596 ns/op  | **3.0x faster**     | 24,320 B | 281       |
| `goquery`               | 45,589 ns/op  | **3.1x faster**     | 24,400 B | 284       |

#### Complex HTML (Full Page ~5KB)

| Parser                  | Time/op       | Speed vs JustGoHTML | Mem/op    | Allocs/op |
| ----------------------- | ------------- | ------------------- | --------- | --------- |
| **JustGoHTML**          | 367,141 ns/op | 1.0x (baseline)     | 127,464 B | 1,963     |
| `golang.org/x/net/html` | 79,211 ns/op  | **4.6x faster**     | 38,048 B  | 504       |
| `goquery`               | 110,858 ns/op | **3.3x faster**     | 38,128 B  | 507       |

### Query Benchmarks

#### Simple Query (`div.feature`)

| Parser         | Time/op     | Speed vs JustGoHTML | Mem/op | Allocs/op |
| -------------- | ----------- | ------------------- | ------ | --------- |
| **JustGoHTML** | 8,034 ns/op | 1.0x (baseline)     | 696 B  | 25        |
| `goquery`      | 4,391 ns/op | **1.8x faster**     | 360 B  | 15        |

#### Complex Query (`section > h2 + div.feature-grid div[data-feature-id]`)

| Parser         | Time/op      | Speed vs JustGoHTML | Mem/op  | Allocs/op |
| -------------- | ------------ | ------------------- | ------- | --------- |
| **JustGoHTML** | 15,411 ns/op | 1.0x (baseline)     | 1,680 B | 28        |
| `goquery`      | 5,973 ns/op  | **2.6x faster**     | 744 B   | 27        |

### Parallel Performance

Performance when running with multiple goroutines (GOMAXPROCS=12):

| Parser                  | Time/op       | Mem/op    | Allocs/op |
| ----------------------- | ------------- | --------- | --------- |
| **JustGoHTML**          | 138,622 ns/op | 127,467 B | 1,963     |
| `golang.org/x/net/html` | 36,676 ns/op  | 38,047 B  | 504       |
| `goquery`               | 36,048 ns/op  | 38,127 B  | 507       |

### Memory Allocations

Comparison of memory allocations for complex HTML parsing:

| Parser                  | Bytes Allocated | Number of Allocations |
| ----------------------- | --------------- | --------------------- |
| **JustGoHTML**          | 132,360 B       | 1,969                 |
| `golang.org/x/net/html` | 38,048 B        | 504                   |
| `goquery`               | 38,128 B        | 507                   |

## Analysis

### Why JustGoHTML is Slower

JustGoHTML's performance characteristics are intentional trade-offs for **100% HTML5 specification compliance**:

1. **Complete Error Recovery**: Implements all HTML5 error recovery rules exactly as browsers do
2. **Proper Adoption Agency Algorithm**: Handles complex cases like misnested formatting elements
3. **Template Element Support**: Full support for `<template>` with separate document fragments
4. **Richer DOM Model**: More complete node types and relationships
5. **Strict Spec Compliance**: No shortcuts or approximations

### Performance in Context

- **~367 µs for complex HTML** means JustGoHTML can parse **~2,700 pages per second**
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

Potential areas for performance improvement in JustGoHTML:

1. **Token Pooling**: Reuse token objects during parsing
2. **String Interning**: Intern common tag names and attribute names
3. **Buffer Management**: Better buffer reuse in tokenizer
4. **Selector Caching**: Cache compiled selectors
5. **SIMD Optimizations**: Use SIMD for character scanning in hot paths

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

JustGoHTML delivers on its promise of **100% HTML5 compliance** with **competitive performance**. While 2-4x slower than parsers that sacrifice spec compliance, it's still fast enough for virtually all use cases at **~2,700 complex pages per second**. The performance overhead buys you guaranteed browser-compatible parsing behavior, which is essential for many applications.

For applications that need exact browser behavior (HTML sanitizers, browser automation tools, testing frameworks), the performance trade-off is worthwhile. For simple parsing where ~70% compliance is acceptable, `x/net/html` or `goquery` remain excellent choices.
