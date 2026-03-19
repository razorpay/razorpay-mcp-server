# Go Concurrency Patterns

## Goroutines

### Basic Goroutine Usage

```go
// ✅ Good - simple goroutine
go func() {
    doWork()
}()

// ✅ Good - passing parameters by value
for _, item := range items {
    go func(i Item) {
        process(i)
    }(item)
}

// ❌ Bad - capturing loop variable
for _, item := range items {
    go func() {
        process(item) // Race condition! item changes
    }()
}
```

### Goroutine Lifecycle Management

```go
// ❌ Bad - no way to stop goroutine
func startWorker() {
    go func() {
        for {
            doWork()
            time.Sleep(time.Second)
        }
    }()
}

// ✅ Good - controlled shutdown with context
func startWorker(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                log.Println("worker stopping")
                return
            case <-ticker.C:
                doWork()
            }
        }
    }()
}

// Usage
ctx, cancel := context.WithCancel(context.Background())
startWorker(ctx)
// Later...
cancel() // Stops the worker
```

### Wait Groups for Coordination

```go
// ✅ Good - using WaitGroup
func processItems(items []Item) error {
    var (
        wg     sync.WaitGroup
        mu     sync.Mutex
        errors []error
    )
    
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            
            if err := process(i); err != nil {
                mu.Lock()
                errors = append(errors, err)
                mu.Unlock()
            }
        }(item)
    }
    
    wg.Wait()
    
    if len(errors) > 0 {
        return fmt.Errorf("processing failed with %d errors", len(errors))
    }
    
    return nil
}
```

### Goroutine Leaks

```go
// ❌ Bad - goroutine leak
func search(query string) Result {
    ch := make(chan Result)
    go func() {
        result := doExpensiveSearch(query)
        ch <- result // Blocks forever if no one reads
    }()
    
    select {
    case result := <-ch:
        return result
    case <-time.After(time.Second):
        return Result{} // Goroutine still running!
    }
}

// ✅ Good - no leak
func search(ctx context.Context, query string) (Result, error) {
    ch := make(chan Result, 1) // Buffered to prevent blocking
    
    go func() {
        result := doExpensiveSearch(query)
        select {
        case ch <- result:
        case <-ctx.Done():
            // Context cancelled, don't block
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

## Channels

### Channel Direction

```go
// ✅ Good - specify channel direction
func producer(ch chan<- int) {
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)
}

func consumer(ch <-chan int) {
    for v := range ch {
        process(v)
    }
}

// Usage
ch := make(chan int)
go producer(ch)
consumer(ch)
```

### Channel Sizing

```go
// ✅ Good - unbuffered for synchronization
ch := make(chan int)

// ✅ Good - buffer of 1 for single async send
ch := make(chan int, 1)

// ⚠️ Be careful - larger buffers need justification
const workerPoolSize = 10 // Document why this size
ch := make(chan Task, workerPoolSize)

// ❌ Bad - magic number
ch := make(chan int, 100)
```

### Channel Patterns

#### Fan-Out, Fan-In

```go
// ✅ Good - fan-out pattern
func fanOut(input <-chan int, workers int) []<-chan int {
    outputs := make([]<-chan int, workers)
    
    for i := 0; i < workers; i++ {
        outputs[i] = worker(input)
    }
    
    return outputs
}

func worker(input <-chan int) <-chan int {
    output := make(chan int)
    
    go func() {
        defer close(output)
        for v := range input {
            output <- process(v)
        }
    }()
    
    return output
}

// Fan-in pattern
func fanIn(channels ...<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                output <- v
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(output)
    }()
    
    return output
}
```

#### Pipeline Pattern

```go
// ✅ Good - pipeline pattern
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

func filter(in <-chan int, predicate func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            if predicate(n) {
                out <- n
            }
        }
    }()
    return out
}

// Usage
nums := generate(1, 2, 3, 4, 5)
squared := square(nums)
filtered := filter(squared, func(n int) bool { return n > 10 })

for result := range filtered {
    fmt.Println(result)
}
```

#### Worker Pool Pattern

```go
// ✅ Good - worker pool
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
    wg      sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, workers*2),
        results: make(chan Result, workers*2),
    }
}

func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx)
    }
}

func (p *WorkerPool) worker(ctx context.Context) {
    defer p.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-p.jobs:
            if !ok {
                return
            }
            result := job.Process()
            select {
            case p.results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (p *WorkerPool) Submit(job Job) {
    p.jobs <- job
}

func (p *WorkerPool) Stop() {
    close(p.jobs)
    p.wg.Wait()
    close(p.results)
}

// Usage
pool := NewWorkerPool(10)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

pool.Start(ctx)

// Submit jobs
for _, job := range jobs {
    pool.Submit(job)
}

// Collect results
go func() {
    for result := range pool.results {
        handleResult(result)
    }
}()

pool.Stop()
```

### Channel Best Practices

```go
// ✅ Good - close channels from sender
func producer(ch chan<- int) {
    defer close(ch) // Sender closes
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

// ❌ Bad - receiver closing
func consumer(ch <-chan int) {
    for v := range ch {
        process(v)
    }
    close(ch) // Don't close receive-only channels!
}

// ✅ Good - range over channel
for v := range ch {
    process(v)
}

// ❌ Bad - manual receive loop
for {
    v, ok := <-ch
    if !ok {
        break
    }
    process(v)
}
```

## Synchronization Primitives

### Mutex

```go
// ✅ Good - protecting shared state
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// ❌ Bad - not protecting all accesses
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    c.value++ // What if we forget defer?
    c.mu.Unlock()
}

func (c *Counter) Value() int {
    return c.value // Not protected!
}
```

### RWMutex

```go
// ✅ Good - read-heavy workload
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    return item, ok
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = item
}
```

### Once

```go
// ✅ Good - one-time initialization
type Service struct {
    once   sync.Once
    client *Client
}

func (s *Service) getClient() *Client {
    s.once.Do(func() {
        s.client = NewClient()
    })
    return s.client
}

// ❌ Bad - manual initialization check
type Service struct {
    mu       sync.Mutex
    client   *Client
    initDone bool
}

func (s *Service) getClient() *Client {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if !s.initDone {
        s.client = NewClient()
        s.initDone = true
    }
    return s.client
}
```

### Atomic Operations

```go
// ✅ Good - atomic counter (simpler than mutex)
type Counter struct {
    value atomic.Int64
}

func (c *Counter) Increment() {
    c.value.Add(1)
}

func (c *Counter) Value() int64 {
    return c.value.Load()
}

// ❌ Bad - using mutex for simple counter
type Counter struct {
    mu    sync.Mutex
    value int64
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

## Context

### Context Propagation

```go
// ✅ Good - context as first parameter
func ProcessPayment(ctx context.Context, payment *Payment) error {
    // Pass context to all downstream calls
    if err := ValidatePayment(ctx, payment); err != nil {
        return err
    }
    
    return SavePayment(ctx, payment)
}

// ❌ Bad - no context
func ProcessPayment(payment *Payment) error {
    // No way to handle cancellation or timeouts
    ValidatePayment(payment)
    SavePayment(payment)
    return nil
}
```

### Context Values

```go
// ✅ Good - type-safe context keys
type contextKey string

const (
    requestIDKey contextKey = "request_id"
    userIDKey    contextKey = "user_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
    requestID, ok := ctx.Value(requestIDKey).(string)
    return requestID, ok
}

// ❌ Bad - using string directly
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, "request_id", requestID)
}
```

### Context Cancellation

```go
// ✅ Good - respecting context cancellation
func ProcessBatch(ctx context.Context, items []Item) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        if err := ProcessItem(ctx, item); err != nil {
            return err
        }
    }
    return nil
}

// ✅ Good - timeout context
func CallExternalAPI(ctx context.Context, request Request) (Response, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return doAPICall(ctx, request)
}
```

## Race Conditions

### Common Race Conditions

```go
// ❌ Bad - race condition
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++ // Race!
}

// ✅ Good - protected
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

// ❌ Bad - map race
var cache = make(map[string]string)

func Get(key string) string {
    return cache[key] // Race!
}

func Set(key, value string) {
    cache[key] = value // Race!
}

// ✅ Good - sync.Map
var cache sync.Map

func Get(key string) (string, bool) {
    value, ok := cache.Load(key)
    if !ok {
        return "", false
    }
    return value.(string), true
}

func Set(key, value string) {
    cache.Store(key, value)
}
```

### Testing for Races

```go
// Run tests with race detector
// go test -race

func TestConcurrentAccess(t *testing.T) {
    counter := NewCounter()
    var wg sync.WaitGroup
    
    // Start 100 goroutines
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }
    
    wg.Wait()
    
    if counter.Value() != 100 {
        t.Errorf("expected 100, got %d", counter.Value())
    }
}
```

## Deadlocks

### Common Deadlock Patterns

```go
// ❌ Bad - deadlock potential
type Account struct {
    mu      sync.Mutex
    balance int
}

func Transfer(from, to *Account, amount int) {
    from.mu.Lock()
    defer from.mu.Unlock()
    
    to.mu.Lock() // Deadlock if another goroutine locks in opposite order!
    defer to.mu.Unlock()
    
    from.balance -= amount
    to.balance += amount
}

// ✅ Good - consistent lock ordering
func Transfer(from, to *Account, amount int) {
    // Always lock in consistent order (e.g., by ID)
    if from.ID < to.ID {
        from.mu.Lock()
        defer from.mu.Unlock()
        to.mu.Lock()
        defer to.mu.Unlock()
    } else {
        to.mu.Lock()
        defer to.mu.Unlock()
        from.mu.Lock()
        defer from.mu.Unlock()
    }
    
    from.balance -= amount
    to.balance += amount
}

// ❌ Bad - channel deadlock
func deadlock() {
    ch := make(chan int)
    ch <- 1 // Blocks forever - no one receiving!
}

// ✅ Good - buffered or separate goroutine
func fixed() {
    ch := make(chan int, 1)
    ch <- 1 // Doesn't block
}
```

## Anti-Patterns

### Don't Start Goroutines in Init

```go
// ❌ Bad
func init() {
    go backgroundWorker() // Can't control lifecycle
}

// ✅ Good
func Start() {
    go backgroundWorker() // Caller controls when to start
}
```

### Don't Use Goroutines for Everything

```go
// ❌ Bad - unnecessary goroutine
go func() {
    result := simpleCalculation()
    ch <- result
}()

// ✅ Good - direct call
result := simpleCalculation()
```

### Don't Share Memory by Communicating

```go
// ❌ Bad - overusing channels
type Data struct {
    ch chan int
}

func (d *Data) Get() int {
    return <-d.ch
}

func (d *Data) Set(v int) {
    d.ch <- v
}

// ✅ Good - use mutex for simple shared state
type Data struct {
    mu    sync.Mutex
    value int
}

func (d *Data) Get() int {
    d.mu.Lock()
    defer d.mu.Unlock()
    return d.value
}

func (d *Data) Set(v int) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.value = v
}
```

## Review Checklist

- [ ] Loop variables are properly captured in goroutines
- [ ] Goroutines have clear exit conditions (context cancellation)
- [ ] WaitGroups are used correctly (Add before goroutine, Done with defer)
- [ ] Channels are closed by senders, not receivers
- [ ] Channel direction is specified in function parameters
- [ ] No goroutine leaks (blocked sends/receives)
- [ ] Shared state is protected with mutexes or atomics
- [ ] Lock ordering is consistent to prevent deadlocks
- [ ] Context is first parameter in functions
- [ ] Context cancellation is checked in loops
- [ ] No race conditions (test with `go test -race`)
- [ ] Atomic operations used for simple counters
- [ ] sync.Once used for one-time initialization
- [ ] Worker pools have bounded resources

## References

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [Context Package](https://pkg.go.dev/context)
- [sync Package](https://pkg.go.dev/sync)
- [Go Race Detector](https://go.dev/blog/race-detector)

