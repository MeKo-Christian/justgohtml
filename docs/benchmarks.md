# Benchmarking Guide

This guide explains how to run and interpret JustGoHTML benchmarks, particularly the comparison benchmarks with other Go HTML parsers.

## Quick Start

```bash
# Run all benchmarks
go test -bench=. -benchmem -benchtime=3s

# Run specific benchmark category
go test -bench=Parse -benchmem
go test -bench=Query -benchmem
go test -bench=Parallel -benchmem

# Compare JustGoHTML vs competitors for parsing
go test -bench='Parse_Simple|Parse_Medium|Parse_Complex' -benchmem

# Compare query performance
go test -bench='Query' -benchmem
```

## Understanding Benchmark Output

Example output:

```
BenchmarkJustGoHTML_Parse_Simple-12    	  204322	     18019 ns/op	   12736 B/op	     211 allocs/op
```

Breaking this down:

- `BenchmarkJustGoHTML_Parse_Simple-12`: Benchmark name with 12 parallel processes (GOMAXPROCS)
- `204322`: Number of iterations run
- `18019 ns/op`: Average time per operation (nanoseconds)
- `12736 B/op`: Average bytes allocated per operation
- `211 allocs/op`: Average number of allocations per operation

## Benchmark Categories

### 1. Parse Benchmarks

Compare parsing speed across three complexity levels:

- **Simple**: Minimal HTML with basic structure (~300 bytes)
- **Medium**: Blog post with metadata, navigation, sections (~3KB)
- **Complex**: Full page with forms, nested divs, attributes (~5KB)

```bash
# Run all parse benchmarks
go test -bench='Parse_(Simple|Medium|Complex)' -benchmem -benchtime=5s
```

**What to look for:**

- `ns/op`: Lower is better (faster parsing)
- `B/op`: Lower is better (less memory usage)
- `allocs/op`: Lower is better (fewer GC pressure)

### 2. Query Benchmarks

Compare CSS selector matching performance:

```bash
# Simple selector: div.feature
# Complex selector: section > h2 + div.feature-grid div[data-feature-id]
go test -bench='Query' -benchmem
```

**What to look for:**

- JustGoHTML vs goquery query performance
- Impact of selector complexity on performance

### 3. Memory Allocation Benchmarks

Focus on memory usage for the complex HTML case:

```bash
go test -bench='AllocsPerOp' -benchmem -benchtime=5s
```

**What to look for:**

- Total bytes allocated
- Number of allocations (affects GC overhead)

### 4. Parallel Benchmarks

Test performance under concurrent load:

```bash
go test -bench='Parallel' -benchmem -benchtime=5s
```

**What to look for:**

- How well each parser scales with parallelism
- Memory usage under concurrent load

## Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkJustGoHTML_Parse_Complex -cpuprofile=cpu.prof -benchtime=10s

# Analyze profile
go tool pprof cpu.prof
# In pprof shell:
# - top10: Show top 10 functions by CPU time
# - list FunctionName: Show annotated source for function
# - web: Generate visual graph (requires graphviz)
```

### Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkJustGoHTML_Parse_Complex -memprofile=mem.prof -benchtime=10s

# Analyze profile
go tool pprof mem.prof
# In pprof shell:
# - top10: Show top 10 functions by memory allocation
# - list FunctionName: Show annotated source
```

### Trace Analysis

```bash
# Generate execution trace
go test -bench=BenchmarkJustGoHTML_Parse_Complex -trace=trace.out -benchtime=3s

# View trace
go tool trace trace.out
```

## Comparing Results

### Using benchstat

Install benchstat:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Compare two benchmark runs:

```bash
# Run benchmarks and save results
go test -bench=. -benchmem -count=10 > old.txt

# After making changes
go test -bench=. -benchmem -count=10 > new.txt

# Compare
benchstat old.txt new.txt
```

Example output:

```
name                     old time/op    new time/op    delta
JustGoHTML_Parse_Simple    18.0µs ± 2%    15.5µs ± 1%  -13.89%  (p=0.000 n=10+10)

name                     old alloc/op   new alloc/op   delta
JustGoHTML_Parse_Simple    12.7kB ± 0%    11.2kB ± 0%  -11.81%  (p=0.000 n=10+10)
```

## Benchmark Testing Best Practices

### 1. Stable Environment

- Close other applications
- Disable CPU frequency scaling: `sudo cpupower frequency-set --governor performance`
- Run multiple iterations: `-count=10`
- Use longer benchmark time for stability: `-benchtime=10s`

### 2. Isolate What You're Testing

```go
func BenchmarkExample(b *testing.B) {
    // Setup outside the loop
    data := []byte(testHTML)

    b.ReportAllocs()
    b.ResetTimer() // Don't measure setup time

    for range b.N {
        // Only this is measured
        doc, _ := Parse(string(data))
        _ = doc
    }
}
```

### 3. Avoid Compiler Optimizations

```go
// BAD: Result might be optimized away
for range b.N {
    Parse(html)
}

// GOOD: Ensure result is used
var result *Document
for range b.N {
    result, _ = Parse(html)
}
_ = result
```

## Real-World Performance

### Throughput Calculation

From benchmark results:

```
BenchmarkJustGoHTML_Parse_Complex-12    10000    367141 ns/op
```

Throughput: `1,000,000,000 ns/s ÷ 367,141 ns/op ≈ 2,724 ops/s`

For a 5KB HTML page, this means:

- **~2,724 pages per second** on a single core
- **~13.6 MB/s** parsing throughput
- **~367 µs** latency per page

### Practical Scenarios

**Web Scraping:**

- Parsing 10,000 pages = ~3.7 seconds
- With 10 goroutines = ~0.37 seconds

**API Processing:**

- Processing 1,000 requests/s with HTML parsing
- Each parse takes 367 µs, well within typical API budget

**Batch Processing:**

- Processing 1 million pages = ~367 seconds (6 minutes)
- With parallelism (12 cores) = ~30 seconds

## Interpreting Comparison Results

### JustGoHTML vs golang.org/x/net/html

**Speed:** JustGoHTML is ~2.2x slower
**Memory:** JustGoHTML uses 66KB vs 38KB for complex HTML (optimized with ring buffer)
**Why:** Full HTML5 spec compliance vs ~70% compliance
**Trade-off:** You get exact browser behavior

**When to use JustGoHTML:**

- Need 100% spec compliance
- Processing complex/malformed HTML
- Need CSS selectors built-in
- Zero dependencies required

**When to use x/net/html:**

- Maximum speed critical
- Simple, well-formed HTML
- ~70% compliance acceptable

### JustGoHTML vs goquery

**Speed:** JustGoHTML parsing is ~2.2x slower, querying is competitive (often faster)
**Memory:** JustGoHTML uses ~66KB vs ~38KB for complex HTML
**Why:** Same as x/net/html (goquery wraps it)
**Trade-off:** Full spec compliance vs speed

**When to use JustGoHTML:**

- Need 100% spec compliance
- Want stdlib-only dependency
- Building security-critical tools

**When to use goquery:**

- Need jQuery-like API
- Speed > spec compliance
- Working with simple HTML

## Continuous Benchmarking

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run quick benchmarks before commit
go test -bench=Parse_Simple -benchtime=1s -run='^$' > /tmp/bench.txt
if [ $? -ne 0 ]; then
    echo "Benchmarks failed!"
    exit 1
fi
```

### CI Integration

In GitHub Actions:

```yaml
- name: Run Benchmarks
  run: |
    go test -bench=. -benchmem -benchtime=3s | tee benchmark.txt

- name: Upload Benchmark Results
  uses: actions/upload-artifact@v3
  with:
    name: benchmark-results
    path: benchmark.txt
```

## Further Resources

- [Go Benchmarking Guide](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [profiling Go programs](https://go.dev/blog/pprof)
- [benchstat documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [JustGoHTML Benchmark Results](../BENCHMARK_RESULTS.md)
