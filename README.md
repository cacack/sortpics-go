# sortpics-go

Go port of [sortpics](../sortpics) - a photo and video organization tool that uses EXIF metadata to create a chronologically-organized archive.

## Project Status

✅ **Production Ready** - v0.1.0

This is a complete rewrite of sortpics in Go, providing:
- **Better performance**: Native concurrency with goroutines and worker pools
- **Single binary**: No Python runtime dependency
- **Cross-platform**: Easy distribution for Linux, macOS, Windows
- **Improved maintainability**: Strong typing and better tooling
- **90.6% test coverage**: Comprehensive test suite with integration tests

## Features

- **Smart metadata extraction**: EXIF → QuickTime → filename → filesystem fallback hierarchy
- **Organized structure**: `YYYY/MM/YYYY-MM-DD/YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
- **Duplicate detection**: SHA256-based with collision resolution
- **Atomic operations**: Safe copy/move with temporary files
- **Parallel processing**: Concurrent file processing with bounded queue
- **RAW file support**: Optional segregation to separate directory tree
- **Album tagging**: Set XMP:Album metadata
- **Timestamp adjustment**: Bulk time/day adjustments
- **Verification mode**: Validate existing archives

## Architecture

### Directory Structure

```
sortpics-go/
├── cmd/
│   └── sortpics/           # CLI entry point
│       ├── main.go         # Main function
│       └── cmd/            # Cobra command definitions
│           ├── root.go     # Root command + flags
│           └── verify.go   # Verify subcommand
├── internal/
│   ├── metadata/           # EXIF extraction (MetadataExtractor)
│   ├── pathgen/            # Path/filename generation (PathGenerator)
│   ├── duplicate/          # SHA256 duplicate detection (DuplicateDetector)
│   └── rename/             # File operations coordinator (ImageRename)
├── pkg/
│   └── config/             # Shared configuration types
├── test/
│   └── testdata/           # Test fixtures
├── Makefile                # Build automation
└── go.mod                  # Go module definition
```

### Component Design

Based on the Python architecture (95% test coverage, clean separation of concerns):

1. **metadata** - Extracts EXIF/QuickTime metadata via ExifTool
   - Datetime extraction with fallback hierarchy
   - Make/model parsing and normalization
   - Time/day adjustments

2. **pathgen** - Generates destination paths and filenames
   - Format: `YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
   - Directory: `YYYY/MM/YYYY-MM-DD/`
   - Configurable subsecond precision

3. **duplicate** - Detects duplicates and resolves collisions
   - SHA256 content hashing
   - Collision resolution with `_N` suffix
   - Skip identical files

4. **rename** - Coordinates file processing
   - Orchestrates metadata → path → duplicate flow
   - Atomic copy/move operations
   - Metadata tag writing

5. **CLI** - Command-line interface
   - Cobra-based flag parsing
   - Worker pool orchestration
   - Progress tracking

## Technology Stack

### Dependencies

- **CLI Framework**: [spf13/cobra](https://github.com/spf13/cobra) - Industry-standard CLI framework
- **EXIF Library**: [barasher/go-exiftool](https://github.com/barasher/go-exiftool) - ExifTool wrapper with stay_open optimization
- **Worker Pool**: [alitto/pond](https://github.com/alitto/pond) - Bounded queue, context support, error propagation
- **Testing**: [stretchr/testify](https://github.com/stretchr/testify) - Assertions and test utilities

### Technology Rationale

**Why go-exiftool?**
- Full feature parity with Python version (already uses ExifTool)
- Comprehensive format support: 500+ file types including RAW (CR2, NEF, DNG, ARW)
- Complete video metadata support (QuickTime:CreateDate for MOV/MP4)
- Write capability for XMP:Album tags (including videos)
- Mature, battle-tested (20+ years of ExifTool development)
- Performance optimized via stay_open feature

**Pure Go alternatives evaluated:**
- Image EXIF: `dsoprea/go-exif` (pure Go, excellent but EXIF-only)
- Video metadata: `Eyevinn/mp4ff` (pure Go, clean API for MP4/MOV)
- **Decision**: ExifTool wrapper chosen for comprehensive RAW support, write capability, and reduced development time (5h vs 20+h)
- See [DECISION.md](DECISION.md) for detailed analysis

**Why pond?**
- Bounded queue with backpressure (matches Python's `Queue(maxsize=...)`)
- Native context.Context support for cancellation
- Task groups for error propagation
- Clean API, zero dependencies

**Alternatives considered:**
- Pure stdlib: Manual worker pool + channels (viable, more boilerplate)
- `errgroup`: Clean but no bounded queue
- `ants`: High performance but less convenient API

## Development

### Prerequisites

```bash
# Go 1.21+ required
go version

# ExifTool required for metadata extraction
# macOS:
brew install exiftool

# Ubuntu/Debian:
sudo apt-get install libimage-exiftool-perl

# Verify installation:
exiftool -ver
```

### Setup

```bash
# Clone repository
git clone https://github.com/chris/sortpics-go.git
cd sortpics-go

# Download dependencies
make deps

# Run tests
make test

# Build binary
make build
```

### Common Tasks

```bash
# Build and install
make install

# Run tests with coverage
make test-coverage

# Run linters (requires golangci-lint)
make lint

# Format code
make fmt

# Run directly (development)
make run-dev ARGS="--help"

# Build and run
make run ARGS="--dry-run --copy /source /dest"
```

## Migration Status

### ✅ All Phases Complete!

- ✅ **Phase 1**: Core Components (duplicate, pathgen) - 86.8% / 100% coverage
- ✅ **Phase 2**: Metadata Extraction - 94.4% coverage
- ✅ **Phase 3**: File Operations (rename) - 81.1% coverage
- ✅ **Phase 4**: Orchestration & CLI - 72.6% coverage
- ✅ **Verify Command**: Archive validation with --fix mode
- ✅ **Integration Tests**: All major features tested

**Overall Coverage**: 90.6% ✅

See [STATUS.md](STATUS.md) and [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for details.

## Usage

```bash
# Copy files (preview)
sortpics --copy --dry-run --recursive /sdcard /photos

# Move files with verbose output
sortpics --move --recursive -vvv /sdcard /photos

# Separate RAW files
sortpics --copy --recursive --raw-path /photos/raw /sdcard /photos

# Set album metadata
sortpics --copy --album "Vacation 2024" /sdcard /photos

# Adjust timestamps (camera timezone was wrong)
sortpics --copy --time-adjust -05:00:00 /sdcard /photos

# Verify archive integrity
sortpics verify /photos

# Verify and fix mismatches
sortpics verify --fix /photos
```

## Testing Strategy

Following Python's comprehensive testing approach (95% coverage):

- **Unit tests**: Each component in isolation with mocks
- **Integration tests**: Component interactions
- **End-to-end tests**: Real files through entire pipeline
- **Benchmark tests**: Performance validation

Target: 90%+ coverage maintained throughout migration

## Performance Goals

Expected improvements over Python version:

- **Startup time**: ~10x faster (no interpreter overhead)
- **Throughput**: 2-5x faster (native concurrency, no GIL)
- **Memory**: Lower overhead (compiled binary, efficient goroutines)
- **Distribution**: Single binary (no Python installation required)

## Contributing

This is a personal migration project, but suggestions and feedback are welcome via issues.

## License

Same as original sortpics project (to be determined)

## References

### Python Original
- Repository: `../sortpics`
- Documentation: `../sortpics/docs/`
- Test suite: `../sortpics/tests/` (125 tests, 95% coverage)

### Research Documents
- EXIF library evaluation (see research notes)
- Concurrency patterns analysis (see research notes)
- Architecture specification: `../sortpics/docs/architecture.md`

## Acknowledgments

- Original sortpics Python implementation
- ExifTool by Phil Harvey
- Go community and library authors
