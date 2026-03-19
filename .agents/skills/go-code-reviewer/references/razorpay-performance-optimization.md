# Go Performance Optimization

## Memory Allocation

### Avoid Unnecessary Allocations

```go
// ❌ Bad - allocates on every call
func FormatUser(user User) string {
    return fmt.Sprintf("%s <%s>", user.Name, user.Email)
}

// ✅ Good - reuse builder
type UserFormatter struct {
    builder strings.Builder
}

func (f *UserFormatter) Format(user User) string {
    f.builder.Reset()
    f.builder.WriteString(user.Name)
    f.builder.WriteString(" <")
    f.builder.WriteString(user.Email)
    f.builder.WriteString(">")
    return f.builder.String()
}
```

### Pre-allocate Slices

```go
// ❌ Bad - grows multiple times
func processItems(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// ✅ Good - single allocation
func processItems(items []Item) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// ✅ Even better - no append overhead
func processItems(items []Item) []Result {
    results := make([]Result, len(items))
    for i, item := range items {
        results[i] = process(item)
    }
    return results
}
```

### Map Capacity

```go
// ❌ Bad - grows during insertion
func buildIndex(items []Item) map[string]Item {
    index := make(map[string]Item)
    for _, item := range items {
        index[item.ID] = item
    }
    return index
}

// ✅ Good - single allocation
func buildIndex(items []Item) map[string]Item {
    index := make(map[string]Item, len(items))
    for _, item := range items {
        index[item.ID] = item
    }
    return index
}
```

### Pointer vs Value

```go
// ⚠️ Careful - large struct passed by value (copies)
type LargeStruct struct {
    Data [1000]int
    // ... many fields
}

func ProcessData(data LargeStruct) { // Copies entire struct!
    // ...
}

// ✅ Good - use pointer for large structs
func ProcessData(data *LargeStruct) {
    // ...
}

// ✅ Good - small struct can be value
type Point struct {
    X, Y int
}

func Distance(p Point) float64 { // Value is fine
    return math.Sqrt(float64(p.X*p.X + p.Y*p.Y))
}
```

## String Operations

### Use strings.Builder

```go
// ❌ Bad - allocates on every concatenation
func buildMessage(parts []string) string {
    message := ""
    for _, part := range parts {
        message += part + " "
    }
    return message
}

// ✅ Good - single allocation
func buildMessage(parts []string) string {
    var b strings.Builder
    b.Grow(len(parts) * 10) // Estimate total size
    for _, part := range parts {
        b.WriteString(part)
        b.WriteString(" ")
    }
    return b.String()
}
```

### Avoid Unnecessary String Conversions

```go
// ❌ Bad - converts to string just to compare
func isAdmin(role []byte) bool {
    return string(role) == "admin"
}

// ✅ Good - compare bytes directly
func isAdmin(role []byte) bool {
    return bytes.Equal(role, []byte("admin"))
}

// ✅ Or use constant
var adminRole = []byte("admin")

func isAdmin(role []byte) bool {
    return bytes.Equal(role, adminRole)
}
```

### Use strconv over fmt

```go
// ❌ Bad - slower
s := fmt.Sprint(42)
f := fmt.Sprintf("%f", 3.14)

// ✅ Good - faster
s := strconv.Itoa(42)
f := strconv.FormatFloat(3.14, 'f', -1, 64)
```

## Loops and Iterations

### Range Over Index

```go
// ❌ Bad - unnecessary index
for i := 0; i < len(items); i++ {
    process(items[i])
}

// ✅ Good - range
for _, item := range items {
    process(item)
}

// ✅ Good - if you need index
for i, item := range items {
    log.Printf("%d: %v", i, item)
}
```

### Avoid Repeated Lookups

```go
// ❌ Bad - looks up config on every iteration
for _, item := range items {
    if item.Score > config.GetThreshold() {
        process(item)
    }
}

// ✅ Good - cache lookup
threshold := config.GetThreshold()
for _, item := range items {
    if item.Score > threshold {
        process(item)
    }
}
```

### Loop Variable in Closures

```go
// ❌ Bad - captures loop variable
for _, item := range items {
    go func() {
        process(item) // All goroutines see last value!
    }()
}

// ✅ Good - pass as parameter
for _, item := range items {
    go func(i Item) {
        process(i)
    }(item)
}

// ✅ Good - shadow variable (Go 1.22+)
for _, item := range items {
    item := item // Creates new variable
    go func() {
        process(item)
    }()
}
```

## Concurrency

### Use Worker Pools

```go
// ❌ Bad - unbounded goroutines
func processAll(items []Item) {
    for _, item := range items {
        go process(item) // Could create millions of goroutines!
    }
}

// ✅ Good - bounded worker pool
func processAll(items []Item) {
    const workers = 10
    sem := make(chan struct{}, workers)
    
    for _, item := range items {
        sem <- struct{}{} // Acquire
        go func(i Item) {
            defer func() { <-sem }() // Release
            process(i)
        }(item)
    }
    
    // Wait for all to finish
    for i := 0; i < workers; i++ {
        sem <- struct{}{}
    }
}

// ✅ Even better - use errgroup
import "golang.org/x/sync/errgroup"

func processAll(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // Max 10 concurrent
    
    for _, item := range items {
        item := item
        g.Go(func() error {
            return process(ctx, item)
        })
    }
    
    return g.Wait()
}
```

### Avoid Goroutine Leaks

```go
// ❌ Bad - goroutine leaks on timeout
func search(query string) Result {
    ch := make(chan Result)
    go func() {
        result := expensiveSearch(query)
        ch <- result // Blocks if no receiver!
    }()
    
    select {
    case result := <-ch:
        return result
    case <-time.After(time.Second):
        return Result{} // Goroutine still running!
    }
}

// ✅ Good - buffered channel prevents leak
func search(ctx context.Context, query string) (Result, error) {
    ch := make(chan Result, 1) // Buffer of 1
    
    go func() {
        result := expensiveSearch(query)
        select {
        case ch <- result:
        case <-ctx.Done():
            // Don't block on send
        }
    }()
    
    select {
    case result := <-ch:
        return result, nil
    case <-ctx.Done():
        return Result{}, ctx.Err()
    }
}
```

### Mutex Contention

```go
// ❌ Bad - single mutex for entire cache
type Cache struct {
    mu   sync.Mutex
    data map[string]Value
}

func (c *Cache) Get(key string) (Value, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    v, ok := c.data[key]
    return v, ok
}

// ✅ Good - use sync.Map for read-heavy workloads
type Cache struct {
    data sync.Map
}

func (c *Cache) Get(key string) (Value, bool) {
    v, ok := c.data.Load(key)
    if !ok {
        return Value{}, false
    }
    return v.(Value), true
}

// ✅ Good - RWMutex for read-heavy
type Cache struct {
    mu   sync.RWMutex
    data map[string]Value
}

func (c *Cache) Get(key string) (Value, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

// ✅ Good - sharded locks for reduced contention
type ShardedCache struct {
    shards [256]*CacheShard
}

type CacheShard struct {
    mu   sync.RWMutex
    data map[string]Value
}

func (c *ShardedCache) getShard(key string) *CacheShard {
    hash := fnv.New32a()
    hash.Write([]byte(key))
    return c.shards[hash.Sum32()%256]
}

func (c *ShardedCache) Get(key string) (Value, bool) {
    shard := c.getShard(key)
    shard.mu.RLock()
    defer shard.mu.RUnlock()
    v, ok := shard.data[key]
    return v, ok
}
```

## I/O Operations

### Buffered I/O

```go
// ❌ Bad - unbuffered writes
func writeData(w io.Writer, data [][]byte) error {
    for _, chunk := range data {
        if _, err := w.Write(chunk); err != nil {
            return err
        }
    }
    return nil
}

// ✅ Good - buffered writes
func writeData(w io.Writer, data [][]byte) error {
    bw := bufio.NewWriter(w)
    defer bw.Flush()
    
    for _, chunk := range data {
        if _, err := bw.Write(chunk); err != nil {
            return err
        }
    }
    return bw.Flush()
}
```

### Batch Database Operations

```go
// ❌ Bad - N queries
func saveUsers(users []User) error {
    for _, user := range users {
        if err := db.Save(user); err != nil {
            return err
        }
    }
    return nil
}

// ✅ Good - batch insert
func saveUsers(users []User) error {
    return db.CreateInBatches(users, 100) // Batch of 100
}

// ✅ Good - prepared statement
func saveUsers(users []User) error {
    stmt, err := db.Prepare("INSERT INTO users(name, email) VALUES(?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, user := range users {
        if _, err := stmt.Exec(user.Name, user.Email); err != nil {
            return err
        }
    }
    return nil
}
```

## Data Structures

### Use Appropriate Types

```go
// ❌ Bad - slice for frequent lookups
func contains(items []string, target string) bool {
    for _, item := range items {
        if item == target {
            return true
        }
    }
    return false
}

// ✅ Good - map for O(1) lookup
func contains(items map[string]bool, target string) bool {
    return items[target]
}

// ❌ Bad - map for ordered iteration
data := make(map[int]Value)
// Maps don't maintain order!

// ✅ Good - slice for ordered data
data := make([]Value, 0, 100)
```

### Sync.Pool for Temporary Objects

```go
// ✅ Good - reuse buffers
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(data []byte) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    // Use buffer
    buf.Write(data)
    // ... process ...
    
    return buf.Bytes(), nil
}
```

## Compilation and Inlining

### Small Functions for Inlining

```go
// ✅ Good - small functions are inlined
func add(a, b int) int {
    return a + b
}

func (p *Point) X() int {
    return p.x
}

// ⚠️ May not inline - too complex
func complexCalculation(a, b, c, d int) int {
    // Many operations, loops, etc.
    // ...
}
```

### Avoid Interface{} When Possible

```go
// ❌ Bad - type assertion overhead
func sum(values []interface{}) int {
    total := 0
    for _, v := range values {
        total += v.(int) // Type assertion on every iteration
    }
    return total
}

// ✅ Good - type-safe
func sum(values []int) int {
    total := 0
    for _, v := range values {
        total += v
    }
    return total
}

// ✅ Good - generics (Go 1.18+)
func sum[T constraints.Integer](values []T) T {
    var total T
    for _, v := range values {
        total += v
    }
    return total
}
```

## Benchmarking

### Write Benchmarks

```go
// ✅ Good - benchmark functions
func BenchmarkProcessItem(b *testing.B) {
    item := Item{ID: "123", Data: "test"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        processItem(item)
    }
}

// ✅ Good - benchmark with setup
func BenchmarkDatabaseQuery(b *testing.B) {
    db := setupTestDB()
    defer db.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        db.Query("SELECT * FROM users WHERE id = ?", i)
    }
}

// ✅ Good - table benchmarks
func BenchmarkSum(b *testing.B) {
    cases := []struct {
        name string
        size int
    }{
        {"small", 10},
        {"medium", 100},
        {"large", 1000},
    }
    
    for _, tc := range cases {
        b.Run(tc.name, func(b *testing.B) {
            data := make([]int, tc.size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                sum(data)
            }
        })
    }
}
```

### Benchmark Comparison

```bash
# Run benchmarks
go test -bench=. -benchmem

# Compare before/after
go test -bench=. -benchmem > old.txt
# Make changes
go test -bench=. -benchmem > new.txt
benchcmp old.txt new.txt

# Or use benchstat
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

## Profiling

### CPU Profiling

```go
import (
    "os"
    "runtime/pprof"
)

func main() {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal(err)
    }
    defer pprof.StopCPUProfile()
    
    // Your code here
}

// Analyze:
// go tool pprof cpu.prof
```

### Memory Profiling

```go
import (
    "os"
    "runtime"
    "runtime/pprof"
)

func main() {
    // Your code here
    
    f, err := os.Create("mem.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    runtime.GC() // Get up-to-date statistics
    if err := pprof.WriteHeapProfile(f); err != nil {
        log.Fatal(err)
    }
}

// Analyze:
// go tool pprof mem.prof
```

### HTTP Profiling

```go
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Your application code
}

// Access profiles at:
// http://localhost:6060/debug/pprof/
// go tool pprof http://localhost:6060/debug/pprof/heap
// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Common Performance Pitfalls

### Unnecessary Boxing

```go
// ❌ Bad - boxes integers into interfaces
func printNumbers(numbers []int) {
    for _, n := range numbers {
        fmt.Println(n) // Converts to interface{}
    }
}

// ✅ Better - format once
func printNumbers(numbers []int) {
    var b strings.Builder
    for _, n := range numbers {
        b.WriteString(strconv.Itoa(n))
        b.WriteString("\n")
    }
    fmt.Print(b.String())
}
```

### Excessive Logging

```go
// ❌ Bad - logs on every iteration
for i, item := range items {
    log.Printf("Processing item %d: %v", i, item)
    process(item)
}

// ✅ Good - log summary
log.Printf("Processing %d items", len(items))
for _, item := range items {
    process(item)
}
log.Printf("Completed processing")
```

### Defer in Tight Loops

```go
// ⚠️ Careful - defer has overhead in loops
func processFiles(filenames []string) error {
    for _, name := range filenames {
        f, err := os.Open(name)
        if err != nil {
            return err
        }
        defer f.Close() // Defers accumulate!
        
        process(f)
    }
    return nil
}

// ✅ Good - explicit close or use function
func processFiles(filenames []string) error {
    for _, name := range filenames {
        if err := processFile(name); err != nil {
            return err
        }
    }
    return nil
}

func processFile(name string) error {
    f, err := os.Open(name)
    if err != nil {
        return err
    }
    defer f.Close() // Defer in function scope
    
    return process(f)
}
```

### Regular Expression Compilation

```go
// ❌ Bad - compiles on every call
func validateEmail(email string) bool {
    regex := regexp.MustCompile(`^[a-z0-9]+@[a-z0-9]+\.[a-z]+$`)
    return regex.MatchString(email)
}

// ✅ Good - compile once
var emailRegex = regexp.MustCompile(`^[a-z0-9]+@[a-z0-9]+\.[a-z]+$`)

func validateEmail(email string) bool {
    return emailRegex.MatchString(email)
}
```

## Review Checklist

- [ ] Slices and maps pre-allocated with capacity
- [ ] `strings.Builder` used for string concatenation
- [ ] `strconv` used instead of `fmt` for conversions
- [ ] Large structs passed by pointer
- [ ] Worker pools used instead of unbounded goroutines
- [ ] No goroutine leaks (buffered channels or context)
- [ ] `sync.RWMutex` for read-heavy workloads
- [ ] Buffered I/O for file operations
- [ ] Batch database operations
- [ ] Appropriate data structures (map vs slice)
- [ ] `sync.Pool` for frequently allocated objects
- [ ] No `interface{}` when concrete type works
- [ ] Benchmarks written for critical paths
- [ ] Regular expressions compiled once
- [ ] Defer not in tight loops
- [ ] No excessive logging in hot paths

## References

- [Go Performance Wiki](https://github.com/golang/go/wiki/Performance)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)

