# Benchmark Summary

This document provides a quick reference for all available benchmarks in the JustGoHTML project.

## Quick Start

```bash
# Run all benchmarks
go test -bench=. -benchmem -benchtime=3s ./...

# Run specific categories
go test -bench=Parse -benchmem ./...              # Parsing benchmarks
go test -bench=Query -benchmem ./...              # Query benchmarks
go test -bench=ToHTML -benchmem ./serialize       # Serialization benchmarks
go test -bench=Stream -benchmem ./stream          # Streaming benchmarks
```

## Benchmark Categories

### 1. Parsing Benchmarks (`benchmark_comparison_test.go`)

**Purpose:** Compare JustGoHTML parsing performance against `golang.org/x/net/html` and `goquery`

**Benchmarks:**

- `BenchmarkJustGoHTML_Parse_Simple` - Parse simple HTML (~300 bytes)
- `BenchmarkJustGoHTML_Parse_Medium` - Parse medium HTML (~3KB blog post)
- `BenchmarkJustGoHTML_Parse_Complex` - Parse complex HTML (~5KB full page)
- `BenchmarkJustGoHTML_ParseBytes_*` - Same as above, using []byte input
- `BenchmarkNetHTML_Parse_*` - Comparison with golang.org/x/net/html
- `BenchmarkGoquery_Parse_*` - Comparison with goquery
- `BenchmarkJustGoHTML_Parse_Parallel` - Parallel parsing performance

**Key Metrics:**

- Time per operation (ns/op)
- Memory allocated (B/op)
- Number of allocations (allocs/op)

### 2. Query Benchmarks (`benchmark_comparison_test.go`)

**Purpose:** Measure CSS selector matching performance

**Benchmarks:**

- `BenchmarkJustGoHTML_Query_Simple` - Simple selector: `div.feature`
- `BenchmarkJustGoHTML_Query_Complex` - Complex selector with combinators
- `BenchmarkGoquery_Query_*` - Comparison with goquery

**Key Findings:**

- JustGoHTML query performance is within 2-3x of goquery
- Selector complexity has moderate impact on performance

### 3. Serialization Benchmarks (`serialize/benchmark_test.go`)

**Purpose:** Measure DOM-to-HTML serialization performance

**Benchmarks:**

- `BenchmarkToHTML_Simple` - Serialize simple document
- `BenchmarkToHTML_Medium` - Serialize medium complexity document
- `BenchmarkToHTML_Complex` - Serialize complex document
- `BenchmarkToHTML_Pretty` - Pretty-printed output with indentation
- `BenchmarkToHTML_Element` - Serialize single element subtree
- `BenchmarkToHTML_DeepNesting` - Deep nesting (20 levels)
- `BenchmarkToHTML_ManyAttributes` - Elements with 50+ attributes
- `BenchmarkToHTML_LargeText` - Large text nodes (10KB)
- `BenchmarkToHTML_SpecialChars` - HTML entity escaping
- `BenchmarkToHTML_ManyChildren` - Elements with 100+ children
- `BenchmarkToHTML_Script` - Script element serialization
- `BenchmarkToHTML_Style` - Style element serialization
- `BenchmarkToHTML_Parallel` - Parallel serialization

**Key Findings:**

- Very fast: 1.4 µs for simple HTML, 19 µs for complex pages
- Pretty printing adds minimal overhead
- Scales linearly with content size

### 4. Streaming Benchmarks (`stream/stream_test.go`)

**Purpose:** Measure event-based parsing throughput

**Benchmarks:**

- `BenchmarkStream` - Simple HTML streaming
- `BenchmarkStreamBytes` - Streaming from []byte
- `BenchmarkStream_Medium` - Medium complexity streaming
- `BenchmarkStream_Complex` - Complex HTML streaming
- `BenchmarkStream_Parallel` - Parallel streaming (4x improvement)
- `BenchmarkStream_FilterEvents` - Event filtering overhead

**Key Findings:**

- Ideal for large documents with low memory footprint
- Excellent parallel performance
- Event filtering adds minimal overhead

## Performance Comparison Summary

### Parsing Speed (Complex HTML)

| Parser                | Time/op       | Mem/op   | Allocs/op | Relative Speed  |
| --------------------- | ------------- | -------- | --------- | --------------- |
| **JustGoHTML**        | 163,500 ns/op | 63,346 B | 1,287     | 1.0x (baseline) |
| golang.org/x/net/html | 74,920 ns/op  | 38,048 B | 504       | 2.2x faster     |
| goquery               | 74,690 ns/op  | 38,128 B | 507       | 2.2x faster     |

**Trade-off:** JustGoHTML is ~2.2x slower but provides **100% HTML5 compliance** vs ~70%

### Query Speed

| Parser         | Simple Query | Complex Query | Relative Speed   |
| -------------- | ------------ | ------------- | ---------------- |
| **JustGoHTML** | 2,999 ns/op  | 4,168 ns/op   | 1.0x (baseline)  |
| goquery        | 3,955 ns/op  | 6,120 ns/op   | 0.8x (slower) ⚡ |

**Note:** JustGoHTML CSS selectors are now **faster than goquery** for many queries!

### Serialization Speed

| Document Type | Time/op      | Throughput        |
| ------------- | ------------ | ----------------- |
| Simple HTML   | 1,401 ns/op  | ~714,000 docs/sec |
| Medium HTML   | 12,233 ns/op | ~81,700 docs/sec  |
| Complex HTML  | 19,004 ns/op | ~52,600 docs/sec  |

### Streaming Speed

| Document Type | Time/op       | Throughput       |
| ------------- | ------------- | ---------------- |
| Simple HTML   | 29,280 ns/op  | ~34,150 docs/sec |
| Medium HTML   | 44,059 ns/op  | ~22,700 docs/sec |
| Complex HTML  | 104,046 ns/op | ~9,600 docs/sec  |

## Real-World Performance

### Throughput Examples

**Parsing:**

- ~6,100 complex pages per second (single core)
- ~13,150 pages per second (12 cores, parallel)

**Round-trip (Parse + Serialize):**

- Simple: ~500,000 round-trips per second
- Complex: ~5,500 round-trips per second

### Latency Examples

| Operation              | Latency  |
| ---------------------- | -------- |
| Parse simple HTML      | ~17 µs   |
| Parse complex HTML     | ~164 µs  |
| Serialize complex HTML | ~19 µs   |
| Query complex selector | ~4.2 µs  |
| Stream complex HTML    | ~104 µs  |

## Running Specific Benchmarks

### Parse Comparison

```bash
go test -bench='(JustGoHTML|NetHTML|Goquery)_Parse' -benchmem
```

### Memory Profiling

```bash
go test -bench=BenchmarkJustGoHTML_Parse_Complex \
  -benchmem -memprofile=mem.prof -benchtime=10s

go tool pprof mem.prof
```

### CPU Profiling

```bash
go test -bench=BenchmarkJustGoHTML_Parse_Complex \
  -cpuprofile=cpu.prof -benchtime=10s

go tool pprof cpu.prof
```

### Benchstat Comparison

```bash
# Before changes
go test -bench=. -benchmem -count=10 > old.txt

# After changes
go test -bench=. -benchmem -count=10 > new.txt

# Compare
benchstat old.txt new.txt
```

## Interpreting Results

### What Each Metric Means

- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

### Performance Guidelines

**When JustGoHTML performance is acceptable:**

- Parsing < 10,000 docs/sec: Use JustGoHTML for 100% compliance
- Need exact browser behavior: Always use JustGoHTML
- Building security tools: JustGoHTML's compliance is essential

**When to consider alternatives:**

- Parsing > 100,000 docs/sec: Consider x/net/html if compliance isn't critical
- Simple, well-formed HTML: x/net/html may be sufficient
- jQuery-like API preferred: goquery provides familiar interface

## Benchmark Maintenance

### Adding New Benchmarks

1. Follow naming convention: `Benchmark<Package>_<Operation>_<Variant>`
2. Use `b.ReportAllocs()` to track memory
3. Use `b.ResetTimer()` after setup
4. Include variety of input sizes (simple, medium, complex)

### Example Benchmark Template

```go
func BenchmarkMyFeature_Simple(b *testing.B) {
    // Setup
    input := prepareInput()

    b.ReportAllocs()
    b.ResetTimer()

    for range b.N {
        result := MyFeature(input)
        _ = result // Prevent optimization
    }
}
```

## See Also

- [BENCHMARK_RESULTS.md](../BENCHMARK_RESULTS.md) - Detailed benchmark results and analysis
- [docs/benchmarks.md](benchmarks.md) - Comprehensive benchmarking guide
- [README.md](../README.md) - Performance comparison table
