# Research Notes

This document contains the research conducted for technology selection during the Python to Go migration.

## Table of Contents
1. [EXIF Libraries Evaluation](#exif-libraries-evaluation)
2. [Concurrency Patterns Analysis](#concurrency-patterns-analysis)

---

## EXIF Libraries Evaluation

**Research Date**: October 14, 2025

### Requirements
- Extract EXIF:DateTimeOriginal, EXIF:ModifyDate, EXIF:SubSecTimeOriginal
- Handle QuickTime:CreateDate for videos (MOV/MP4)
- Extract make/model
- Write metadata back to files (XMP:Album)
- Good performance for batch processing

### Evaluated Libraries

#### 1. github.com/barasher/go-exiftool ⭐ **RECOMMENDED**

**Stats**: 278 stars, actively maintained (last release June 2023)

**Pros**:
- Complete feature parity with ExifTool CLI
- Full video support (QuickTime metadata)
- Write capability (v1.7.0+)
- Performance optimized via stay_open feature
- Battle-tested, inherits ExifTool's maturity
- Easy migration from Python (same tag names)

**Cons**:
- Requires ExifTool binary installation
- Not a single static binary
- External process overhead (mitigated by stay_open)

**Decision**: Selected for feature completeness and proven reliability

#### 2. github.com/dsoprea/go-exif

**Stats**: 565 stars, actively maintained

**Pros**:
- Pure Go implementation
- No external dependencies
- Write capability
- Well-tested (560K+ images)

**Cons**:
- EXIF-only (no native video support)
- No QuickTime/MP4 support
- Complex API
- Would need separate MP4 parser

**Decision**: Rejected - lack of video support is a blocker

#### 3. github.com/rwcarlsen/goexif

**Stats**: 664 stars, unmaintained (last update 2019)

**Decision**: Rejected - abandoned project, read-only, no video support

### Final Decision: barasher/go-exiftool ✅

**Selected for:**
- Full feature parity with Python version
- Comprehensive format support (500+ file types)
- Battle-tested reliability (20+ years ExifTool development)
- Write capability for XMP:Album tags
- Same tag names as Python implementation (easy migration)

**Video Metadata Alternative Evaluated:**
After additional research into pure Go video libraries (see DECISION.md), we evaluated:
- `Eyevinn/mp4ff` (564 stars) - Clean API, automatic epoch conversion
- `abema/go-mp4` (504 stars) - Low-level box parsing
- Hybrid approach: `dsoprea/go-exif` for images + `mp4ff` for videos

**Why We Chose ExifTool Over Pure Go:**
1. RAW file support (CR2, NEF, DNG, ARW, etc.) critical for photographers
2. Write capability for videos (XMP:Album tags)
3. Development velocity: 5 hours vs 20+ hours
4. Users already have ExifTool installed (Python version requirement)
5. Comprehensive edge case handling (corrupted files, timezones, etc.)

Provides smooth migration path with full feature parity.

---

## Concurrency Patterns Analysis

**Research Date**: October 14, 2025

### Current Python Architecture
- `multiprocessing.Pool` with custom worker function
- Bounded queue: `Queue(maxsize=thread_count * 2)`
- `multiprocessing.Lock()` for IO coordination
- Graceful shutdown via sentinels + pool.close()/join()
- Forced shutdown on Ctrl-C with pool.terminate()

### Go Requirements
- Bounded queue with backpressure
- Context-based cancellation
- Error handling per-task
- Coordinated logging
- Graceful shutdown

### Evaluated Approaches

#### Option 1: Manual Worker Pool (stdlib)

**Implementation**: Channels + goroutines + sync.WaitGroup

**Pros**:
- Full control
- No dependencies
- Clear mechanics
- Lightweight

**Cons**:
- ~100 lines of boilerplate
- Manual resource management
- Need to implement error collection

**Example**:
```go
pool := NewWorkerPool(numWorkers, queueSize)
pool.Start(ctx)
for _, file := range files {
    pool.Submit(Task{Path: file})
}
pool.Close()
pool.Wait()
```

#### Option 2: errgroup with SetLimit (stdlib)

**Implementation**: golang.org/x/sync/errgroup

**Pros**:
- Clean, idiomatic code
- Built-in error propagation
- Context cancellation automatic
- SetLimit provides bounded concurrency
- ~40 lines of code

**Cons**:
- No task queue (less visibility)
- All-or-nothing error behavior
- Less control over worker lifecycle

**Example**:
```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(8)
for _, file := range files {
    file := file
    g.Go(func() error {
        return processFile(ctx, file)
    })
}
return g.Wait()
```

#### Option 3: alitto/pond ⭐ **RECOMMENDED**

**Implementation**: Third-party worker pool

**Pros**:
- **Bounded queue with backpressure** (Submit() blocks when full)
- Native context.Context support
- Task groups for error propagation
- Can return errors and results
- Zero dependencies
- Clean API

**Cons**:
- Third-party dependency (but stable and well-maintained)

**Example**:
```go
pool := pond.New(8, 16, pond.Context(ctx))
for _, file := range files {
    file := file
    pool.SubmitErr(func() error {
        return processFile(file)
    })
}
pool.StopAndWait()
```

#### Option 4: gammazero/workerpool

**Pros**: Simple API, 600+ stars

**Cons**: Unlimited queue (no backpressure), no context support

**Decision**: Rejected - lacks bounded queue

#### Option 5: panjf2000/ants

**Pros**: Highest performance, 12K+ stars, goroutine pooling

**Cons**: No built-in error handling, manual WaitGroup, focused on throughput over convenience

**Decision**: Rejected - optimization not needed, convenience preferred

### Winner: alitto/pond

**Rationale**:
1. Bounded queue exactly matches Python's `Queue(maxsize=...)`
2. `Submit()` blocks when full (automatic backpressure)
3. Context support for clean Ctrl-C handling
4. Task groups for error propagation
5. Modern design, actively maintained

**Alternative**: errgroup for pure stdlib approach, but lacks bounded queue

### Migration Pattern

```go
// Create pool with bounded queue
pool := pond.New(
    numWorkers,      // 8 workers
    queueSize,       // 16 max queued tasks
    pond.Context(ctx), // Cancellation
)

// Submit tasks (blocks if queue full)
for _, file := range files {
    file := file
    pool.Submit(func() {
        processFile(file)
    })
}

// Graceful shutdown
pool.StopAndWait()
```

This closely mirrors the Python architecture while leveraging Go's native concurrency.

---

## Conclusions

**EXIF**: Use barasher/go-exiftool for feature completeness
**Concurrency**: Use alitto/pond for bounded queue + context support
**Testing**: Use stretchr/testify for familiar assert/require API
**CLI**: Use spf13/cobra (industry standard)

These choices prioritize:
1. Feature parity with Python version
2. Ease of migration
3. Maintainability
4. Active maintenance
