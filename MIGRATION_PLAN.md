# SortPics Migration Plan: Python → Go

## Executive Summary

This document outlines the phased approach to migrating sortpics from Python to Go. The migration follows a **bottom-up, component-by-component** strategy, leveraging the clean architecture already established in the Python codebase.

**Status**: ✅ Pre-migration setup complete
**Next Phase**: Phase 1 - Core Components

---

## Pre-Migration Setup ✅ COMPLETE

### Completed Tasks

- [x] **Research EXIF libraries** → Selected: `barasher/go-exiftool`
  - Full feature parity with Python ExifTool wrapper
  - Video metadata support (QuickTime:CreateDate)
  - Write capability for XMP tags
  - Performance optimized via stay_open

- [x] **Research concurrency patterns** → Selected: `alitto/pond`
  - Bounded queue with backpressure (matches Python's Queue)
  - Context support for cancellation
  - Task groups for error propagation
  - Alternative: stdlib errgroup (viable for pure stdlib approach)

- [x] **Initialize Go module** → `github.com/chris/sortpics-go`

- [x] **Create directory structure**
  ```
  cmd/sortpics/          # CLI entry point
  internal/              # Private packages
    ├── metadata/        # EXIF extraction
    ├── pathgen/         # Path generation
    ├── duplicate/       # Duplicate detection
    └── rename/          # File operations
  pkg/config/            # Shared types
  test/testdata/         # Test fixtures
  ```

- [x] **Set up Cobra CLI framework**
  - Main command with all flags
  - Verify subcommand
  - Help text and examples

- [x] **Create Makefile** with targets:
  - `make build` - Build binary
  - `make test` - Run tests
  - `make test-coverage` - Coverage report
  - `make lint` - Run linters
  - `make fmt` - Format code

- [x] **Set up testing infrastructure**
  - Installed testify for assertions
  - Created placeholder test files
  - Configured golangci-lint

- [x] **Documentation**
  - README.md - Project overview and usage
  - RESEARCH.md - Technology selection rationale
  - MIGRATION_PLAN.md - This file

### Verification

```bash
cd ../sortpics-go
make test          # ✅ All tests pass
make build         # ✅ Binary builds
./bin/sortpics --help  # ✅ CLI works
```

---

## Migration Phases

### Phase 1: Core Components (No I/O) 🎯 NEXT

**Goal**: Port pure logic components with 100% test coverage

#### 1A. Duplicate Detector
- **Python source**: `sortpics/duplicate_detector.py` (59 lines, 100% coverage)
- **Python tests**: `tests/test_duplicate_detector.py` (18 tests)
- **Go target**: `internal/duplicate/duplicate.go`
- **Estimated effort**: 2-4 hours

**Tasks**:
- [ ] Implement `DuplicateDetector` struct
- [ ] `CalculateSHA256(path string) (string, error)` - File hashing
- [ ] `IsDuplicate(src, dst string) (bool, error)` - Compare hashes
- [ ] `ResolveCollision(path string) (string, error)` - Add `_N` suffix
- [ ] `CheckAndResolve(src, dst string) (string, bool, error)` - Main entry
- [ ] Port all 18 tests
- [ ] Verify 100% coverage

**Dependencies**: None (uses stdlib: `crypto/sha256`, `io`, `os`)

#### 1B. Path Generator
- **Python source**: `sortpics/path_generator.py` (46 lines, 100% coverage)
- **Python tests**: `tests/test_path_generator.py` (18 tests)
- **Go target**: `internal/pathgen/pathgen.go`
- **Estimated effort**: 2-4 hours

**Tasks**:
- [ ] Define `ImageMetadata` struct in `pkg/config/types.go`
- [ ] Implement `PathGenerator` struct
- [ ] `GeneratePath(metadata, destBase, ext) string` - Main entry
- [ ] `generateFilename(metadata, ext) string` - Build filename
- [ ] `addIncrement(path string, n int) string` - Collision suffix
- [ ] Port all 18 tests (filename format, precision, old naming)
- [ ] Verify 100% coverage

**Dependencies**: None (uses stdlib: `path/filepath`, `time`, `fmt`)

**Validation**: Compare output against Python for identical inputs

---

### Phase 2: Metadata Extraction

**Goal**: Integrate ExifTool for metadata extraction

#### 2A. Setup ExifTool Integration
- **Python source**: `sortpics/metadata.py` (127 lines, 98% coverage)
- **Python tests**: `tests/test_metadata.py` (21 tests)
- **Go target**: `internal/metadata/metadata.go`
- **Estimated effort**: 4-6 hours

**Tasks**:
- [ ] Install go-exiftool: `go get github.com/barasher/go-exiftool`
- [ ] Implement `MetadataExtractor` struct with ExifTool instance
- [ ] `NewMetadataExtractor() (*MetadataExtractor, error)` - Initialize
- [ ] `Close()` - Cleanup ExifTool process
- [ ] `Extract(path string) (*ImageMetadata, error)` - Main entry
- [ ] `extractDateTime(metadata) time.Time` - Fallback hierarchy
  - EXIF:DateTimeOriginal + SubSecTimeOriginal
  - EXIF:ModifyDate + SubSecTime
  - QuickTime:CreateDate (videos)
  - Filename pattern matching (regex)
  - Filesystem ctime
- [ ] `extractMakeModel(metadata) (string, string)` - Parse camera info
- [ ] `normalizeMake(make string) string` - Capitalize, handle edge cases
- [ ] `normalizeModel(model, make string) string` - Remove make, capitalize
- [ ] Implement time/day adjustments
- [ ] Port all 21 tests with mocked ExifTool responses
- [ ] Test with real image files
- [ ] Verify coverage (target: 95%+)

**Dependencies**:
- `github.com/barasher/go-exiftool`
- Phase 1 complete (`ImageMetadata` struct)

**Validation**:
- Extract metadata from sample images
- Compare results with Python implementation
- Test video files (MOV/MP4)

---

### Phase 3: File Operations

**Goal**: Implement file processing coordinator

#### 3A. Rename Package
- **Python source**: `sortpics/rename.py` (131 lines, 73% coverage)
- **Python tests**: `tests/test_rename.py` (12 tests)
- **Go target**: `internal/rename/rename.go`
- **Estimated effort**: 4-6 hours

**Tasks**:
- [ ] Define `ProcessingConfig` struct in `pkg/config/config.go`
- [ ] Implement `ImageRename` struct (coordinator)
- [ ] `NewImageRename(config) *ImageRename` - Constructor
- [ ] `ParseMetadata(sourcePath) error` - Orchestrate extraction
  - Call MetadataExtractor
  - Call PathGenerator
  - Call DuplicateDetector
- [ ] `Perform() error` - Execute file operation
  - Create destination directories
  - Atomic copy or move
  - Write metadata tags (XMP:Album, keywords)
- [ ] `safeCopy(src, dst string) error` - Atomic copy with temp file
- [ ] `safeMove(src, dst string) error` - Rename or copy+delete
- [ ] `isValidExtension(ext string) bool` - Filter file types
- [ ] `isRAW(ext string) bool` - Detect RAW formats
- [ ] Port 12 integration tests
- [ ] Test cross-filesystem moves (EXDEV handling)

**Dependencies**:
- Phase 1 complete (duplicate, pathgen)
- Phase 2 complete (metadata)

**Key Implementation Details**:
- Use `os.CreateTemp()` for atomic operations
- Handle `syscall.EXDEV` for cross-filesystem moves
- Use `uuid` for temp file names
- Cleanup temp files on error

---

### Phase 4: Orchestration & CLI

**Goal**: Complete working tool with parallel processing

#### 4A. Worker Pool Implementation
- **Python source**: `sortpics/__main__.py` (130 lines, 68% coverage)
- **Python tests**: `tests/test_main.py` (30 tests)
- **Go target**: `cmd/sortpics/cmd/root.go` (expand)
- **Estimated effort**: 6-8 hours

**Tasks**:
- [ ] Install pond: `go get github.com/alitto/pond`
- [ ] Implement `run()` function in root.go
- [ ] Parse and validate CLI arguments
- [ ] Create `ProcessingConfig` from flags
- [ ] Implement directory walking (with `--recursive` support)
- [ ] Create bounded worker pool
  ```go
  pool := pond.New(
      numWorkers,
      numWorkers * 2,  // Queue size
      pond.Context(ctx),
  )
  ```
- [ ] Submit tasks to pool
- [ ] Implement Ctrl-C handling with context cancellation
- [ ] Implement progress reporting (consider github.com/vbauerster/mpb)
- [ ] Collect and report errors
- [ ] Implement `--clean` mode (remove empty directories)
- [ ] Handle `--dry-run` mode
- [ ] Port 30 tests
- [ ] End-to-end integration tests

**Dependencies**:
- `github.com/alitto/pond`
- Phase 3 complete (rename package)

**Key Implementation Details**:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Ctrl-C handling
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
go func() {
    <-sigChan
    cancel()
}()

// Worker pool
pool := pond.New(workers, queueSize, pond.Context(ctx))
defer pool.StopAndWait()

// Submit tasks
filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
    if err != nil || info.IsDir() {
        return err
    }

    pool.Submit(func() {
        processFile(ctx, path, config)
    })
    return nil
})
```

---

### Phase 5: Verification Tool

**Goal**: Implement `sortpics verify` subcommand

#### 5A. Verify Command
- **Python source**: `sortpics/verify.py` (99 lines, 100% coverage)
- **Python tests**: `tests/test_verify.py` (25 tests)
- **Go target**: `cmd/sortpics/cmd/verify.go` (expand)
- **Estimated effort**: 4-6 hours

**Tasks**:
- [ ] Implement `internal/verify/verify.go` package
- [ ] `Verifier` struct
- [ ] `VerifyArchive(dirs []string) (*Report, error)` - Main entry
- [ ] `verifyFile(path string) *FileResult` - Check single file
  - Extract EXIF metadata
  - Parse filename (reverse of PathGenerator)
  - Compare datetime, make, model
  - Report mismatches
- [ ] `findDuplicates(dirs []string) []DuplicateSet` - SHA256-based
- [ ] `fixFile(path string) error` - Rename to match EXIF (for --fix)
- [ ] `Report` struct with summary statistics
- [ ] Implement `runVerify()` in verify.go command
- [ ] Pretty-print results (table format)
- [ ] Port 25 tests

**Dependencies**:
- Phase 2 complete (metadata extraction)
- Phase 1 complete (duplicate detection)

---

### Phase 6: Finalization

**Goal**: Production-ready release

#### 6A. Integration & Polish
- **Estimated effort**: 4-8 hours

**Tasks**:
- [ ] End-to-end testing with real photo libraries
- [ ] Performance benchmarking vs Python
  - Measure startup time
  - Measure throughput (files/second)
  - Measure memory usage
- [ ] Cross-platform testing
  - [ ] macOS
  - [ ] Linux (Ubuntu)
  - [ ] Windows
- [ ] Documentation
  - [ ] Usage examples in README
  - [ ] Migration guide for Python users
  - [ ] Installation instructions
  - [ ] Troubleshooting guide
- [ ] Release automation
  - [ ] Makefile targets for cross-compilation
  - [ ] GitHub Actions CI/CD (optional)
  - [ ] Versioning strategy

#### 6B. Performance Validation

**Expected improvements**:
- Startup time: ~10x faster (no Python interpreter)
- Throughput: 2-5x faster (native concurrency, compiled code)
- Memory: Lower overhead (efficient goroutines vs multiprocessing)
- Distribution: Single binary (no dependencies)

**Benchmarks to run**:
```bash
# Small dataset (100 files)
time sortpics --copy --dry-run /source /dest

# Medium dataset (1000 files)
time sortpics --copy /source /dest

# Large dataset (10000+ files)
time sortpics --copy --workers 16 /source /dest
```

Compare with Python:
```bash
time poetry run sortpics --copy /source /dest
```

---

## Testing Strategy

### Unit Tests
- Follow Python's test structure (95% coverage)
- Use testify for assertions: `assert.Equal(t, expected, actual)`
- Mock external dependencies (ExifTool responses)
- Test edge cases: empty files, corrupted EXIF, etc.

### Integration Tests
- Use `test/testdata/` for sample files
- Create test fixtures with known metadata
- Test complete pipeline: file in → organized file out

### Benchmark Tests
- Add `*_test.go` files with benchmark functions
- Focus on hot paths: SHA256, metadata parsing
- Track performance across versions

### Coverage Goals
- Overall: 90%+ (Python has 95%)
- Core packages: 95%+ (duplicate, pathgen, metadata)
- CLI: 70%+ (harder to test, focused on integration)

Run coverage:
```bash
make test-coverage
open coverage.html
```

---

## Migration Guidelines

### Code Style
- Follow Go conventions: `gofmt`, `golangci-lint`
- Use descriptive names: `calculateSHA256` not `calc_sha`
- Error handling: return errors, don't panic
- Comments: document exported functions

### Python → Go Mapping

| Python | Go |
|--------|---|
| `class MetadataExtractor:` | `type MetadataExtractor struct {}` |
| `def extract(self, path):` | `func (m *MetadataExtractor) Extract(path string) error` |
| `@dataclass` | `type Config struct {}` |
| `typing.Optional[str]` | `*string` or empty string |
| `List[str]` | `[]string` |
| `Dict[str, Any]` | `map[string]interface{}` |
| `raise ValueError()` | `return fmt.Errorf()` |
| `with lock:` | `mu.Lock(); defer mu.Unlock()` |

### Error Handling Pattern
```go
// Python
if not path.exists():
    raise FileNotFoundError(f"File not found: {path}")

// Go
if _, err := os.Stat(path); os.IsNotExist(err) {
    return fmt.Errorf("file not found: %s", path)
}
```

### Testing Pattern
```go
// Python (pytest)
def test_calculate_sha256():
    result = detector.calculate_sha256("test.jpg")
    assert result == "abc123..."

// Go (testify)
func TestCalculateSHA256(t *testing.T) {
    result, err := detector.CalculateSHA256("test.jpg")
    require.NoError(t, err)
    assert.Equal(t, "abc123...", result)
}
```

---

## Risk Mitigation

### Known Challenges

1. **ExifTool dependency**
   - Risk: Binary not installed on user systems
   - Mitigation: Clear error message, installation docs
   - Alternative: Consider bundling ExifTool (licensing permitting)

2. **Cross-platform compatibility**
   - Risk: File path separators, atomic operations
   - Mitigation: Use `filepath` package, test on all platforms
   - Test cross-filesystem moves (symlinks, network drives)

3. **Performance expectations**
   - Risk: Users expect significant speedup
   - Mitigation: Benchmark early, optimize hot paths
   - Document realistic performance improvements

4. **Feature parity**
   - Risk: Missing edge cases from Python version
   - Mitigation: Comprehensive test porting
   - Side-by-side validation with sample datasets

---

## Success Criteria

Phase complete when:
- [ ] All tests passing (including ported Python tests)
- [ ] Coverage ≥90%
- [ ] Linters passing (golangci-lint)
- [ ] Documentation complete
- [ ] Manual testing successful with real files

Final release ready when:
- [ ] All phases complete
- [ ] Performance benchmarks meet expectations
- [ ] Cross-platform testing complete
- [ ] No known critical bugs
- [ ] User documentation complete

---

## Next Steps

**Immediate**: Start Phase 1A - Duplicate Detector

```bash
cd internal/duplicate
# Create duplicate.go
# Port from sortpics/duplicate_detector.py
# Port tests from tests/test_duplicate_detector.py
make test
```

**Commands to get started**:
```bash
# Review Python implementation
cd ../sortpics
cat sortpics/duplicate_detector.py
cat tests/test_duplicate_detector.py

# Start Go implementation
cd ../sortpics-go/internal/duplicate
# Create duplicate.go and implement
```

**Recommended workflow**:
1. Read Python source + tests
2. Write Go types and function signatures
3. Implement one function at a time
4. Port corresponding test
5. Run test, iterate until passing
6. Move to next function
7. Verify coverage when package complete

---

## Resources

### Documentation
- Python source: `../sortpics/sortpics/`
- Python tests: `../sortpics/tests/`
- Architecture: `../sortpics/docs/architecture.md`
- Specification: `../sortpics/docs/specification.md`

### Go Libraries
- ExifTool: https://github.com/barasher/go-exiftool
- Worker pool: https://github.com/alitto/pond
- CLI: https://github.com/spf13/cobra
- Testing: https://github.com/stretchr/testify

### Learning Resources
- Go by Example: https://gobyexample.com
- Effective Go: https://go.dev/doc/effective_go
- Go Testing: https://go.dev/doc/tutorial/add-a-test

---

**Last Updated**: October 14, 2025
**Status**: Pre-migration complete, ready for Phase 1
