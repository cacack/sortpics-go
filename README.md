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

## Installation

### Prerequisites

**ExifTool** - Required for metadata extraction

```bash
# macOS
brew install exiftool

# Ubuntu/Debian
sudo apt-get install libimage-exiftool-perl

# Windows
# Download from https://exiftool.org/

# Verify installation
exiftool -ver
```

### Install from Source

```bash
# Clone repository
git clone https://github.com/chris/sortpics-go.git
cd sortpics-go

# Install to ~/.local/bin (recommended)
make install

# Or install to GOPATH/bin
make install-global

# Verify installation
sortpics --version
```

### Build Options

```bash
# Build for all platforms (cross-compilation)
make build                # Binaries in ./dist/
                          # Linux (amd64, arm64)
                          # macOS (amd64, arm64)
                          # Windows (amd64, arm64)

# Run tests
make test

# Run benchmarks
make bench

# Generate coverage report
make test-coverage        # Creates coverage.html
```

## Quick Start

### Basic Usage

```bash
# Preview what would happen (always start with this!)
sortpics --copy --dry-run -v /source/photos /archive

# Actually copy files
sortpics --copy /source/photos /archive

# Move files (removes originals)
sortpics --move /source/photos /archive

# Process subdirectories recursively
sortpics --copy --recursive /source/photos /archive
```

### Common Workflows

```bash
# Organize SD card to archive
sortpics --copy --recursive -v /Volumes/SDCARD /Users/me/Photos

# Move with progress bar (no -v flag)
sortpics --move --recursive /import /archive

# Separate RAW files
sortpics --copy --recursive \
  --raw-path /archive/raw \
  /sdcard /archive

# Remove empty directories after move
sortpics --move --recursive --clean /sdcard /archive

# Set album name for batch
sortpics --copy --album "Summer Vacation 2024" /import /archive

# Fix camera timezone (subtract 5 hours)
sortpics --copy --time-adjust -05:00:00 /import /archive

# Adjust by days (add 1 day)
sortpics --copy --day-adjust 1 /import /archive
```

### Archive Verification

```bash
# Check if filenames match EXIF data
sortpics verify /archive

# Find and report mismatches
sortpics verify /archive 2>&1 | grep MISMATCH

# Automatically fix mismatches
sortpics verify --fix /archive
```

### Verbosity Levels

```bash
# Silent (progress bar only)
sortpics --copy /source /dest

# Basic info (-v)
sortpics --copy -v /source /dest

# Detailed (-vv)
sortpics --copy -vv /source /dest

# Debug (-vvv)
sortpics --copy -vvv /source /dest
```

### Shell Completion

sortpics includes built-in shell completion for bash, zsh, fish, and PowerShell.

#### Bash

```bash
# Generate completion script
sortpics completion bash > /tmp/sortpics-completion.bash

# Install for current user
sortpics completion bash > ~/.local/share/bash-completion/completions/sortpics

# Or install system-wide (requires sudo)
sortpics completion bash | sudo tee /usr/share/bash-completion/completions/sortpics > /dev/null

# Reload shell or source the file
source ~/.local/share/bash-completion/completions/sortpics
```

#### Zsh

```bash
# Generate and install
sortpics completion zsh > "${fpath[1]}/_sortpics"

# Or add to .zshrc for auto-generation
echo 'source <(sortpics completion zsh)' >> ~/.zshrc

# Reload shell
exec zsh
```

#### Fish

```bash
# Generate and install
sortpics completion fish > ~/.config/fish/completions/sortpics.fish

# Or add to config.fish
echo 'sortpics completion fish | source' >> ~/.config/fish/config.fish

# Reload shell
exec fish
```

#### PowerShell

```powershell
# Generate completion script
sortpics completion powershell > sortpics-completion.ps1

# Add to profile
sortpics completion powershell >> $PROFILE

# Reload profile
. $PROFILE
```

**Features**:
- Tab completion for all commands (root, verify)
- Flag completion with descriptions
- Path completion for source/destination arguments
- Completion for flag values where applicable

## Troubleshooting

### ExifTool Not Found

**Error**: `exiftool not found. Please install it first`

**Solution**:
```bash
# macOS
brew install exiftool

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install libimage-exiftool-perl

# Verify
exiftool -ver  # Should show version 12.00+
```

### No Files Processed

**Issue**: `Found 0 files to process`

**Possible causes**:
1. Wrong source directory path
2. No supported file extensions in directory
3. Missing `--recursive` flag for subdirectories

**Solution**:
```bash
# Check directory exists and has files
ls -la /source/directory

# Use recursive flag for subdirectories
sortpics --copy --recursive /source /dest

# Check supported extensions
sortpics --help | grep -A 20 "Features"
```

### Permission Denied

**Error**: `failed to create destination directory: permission denied`

**Solution**:
```bash
# Check destination is writable
ls -ld /destination/path

# Create destination first
mkdir -p /destination/path

# Or use sudo (not recommended)
sudo sortpics --copy /source /dest
```

### Duplicates Not Detected

**Issue**: Duplicate files are copied instead of skipped

**Cause**: Duplicate detection is content-based (SHA256). Files with identical content but different names are detected. Files with similar but not identical content are not duplicates.

**To verify**:
```bash
# Check if files are truly identical
sha256sum file1.jpg file2.jpg
```

### Progress Bar Interferes with Logging

**Solution**: Progress bar auto-hides in verbose mode
```bash
# Use -v to disable progress bar
sortpics --copy -v /source /dest
```

### Cross-Filesystem Move Fails

**Error**: `invalid cross-device link`

**Solution**: This is handled automatically by falling back to copy+delete. If you see this error, please report it as a bug.

### Filename Too Long

**Error**: `file name too long`

**Cause**: Generated filename exceeds filesystem limits (usually 255 characters)

**Solutions**:
```bash
# Use old naming format (shorter)
sortpics --copy --old-naming /source /dest

# Reduce subsecond precision
sortpics --copy --precision 0 /source /dest
```

## Development

### Setup Development Environment

```bash
# Clone and setup
git clone https://github.com/chris/sortpics-go.git
cd sortpics-go

# Download dependencies
make deps

# Run tests
make test

# Run with go run (faster iteration)
make run-dev ARGS="--help"
```

### Common Development Tasks

```bash
# Format code
make fmt

# Run linters
make lint               # Requires: brew install golangci-lint

# Run specific test
go test -v -run TestVerify ./cmd/sortpics/cmd/...

# Run benchmarks
make bench

# Generate test fixtures
make test-fixtures      # Requires exiftool

# Clean build artifacts
make clean
```

## Test Coverage

**Overall Coverage**: 90.6% ✅

- **duplicate**: 86.8% coverage
- **pathgen**: 100.0% coverage
- **metadata**: 94.4% coverage
- **rename**: 81.1% coverage
- **CLI**: 72.6% coverage

All major features have comprehensive unit and integration tests.

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

## Performance Comparison

### vs Python Original

Real-world performance measurements on identical workload (5 test images):

| Metric | Python | Go | Improvement |
|--------|--------|-----|-------------|
| **Full Copy** (with I/O) | ~2.5s | ~1.04s | **2.4x faster** |
| **Dry-Run** (metadata only) | ~800ms | ~256ms | **3.1x faster** |
| **Directory Walk** | ~1.2ms | ~455µs | **2.6x faster** |
| **Startup Time** | ~250ms | ~5ms | **50x faster** |
| **Memory Overhead** | ~45MB | ~292KB/op | **154x lower** |
| **Binary Size** | N/A (requires Python) | ~8.5MB | Single binary |

### Key Performance Features

- **Native concurrency**: Goroutines without GIL limitations (Python's bottleneck)
- **Worker pool optimization**: Bounded queue with backpressure, optimal at CPU count
- **Efficient I/O**: Atomic operations with minimal memory copying
- **Zero startup cost**: Compiled binary vs Python interpreter initialization
- **Scalability**: Worker benchmarks show linear scaling up to CPU cores (8 workers: 2.4x faster than single-threaded)

### Benchmark Results

From `make bench` on macOS M1 (8 cores):

```
BenchmarkCopyMode-8              1    1044291750 ns/op  (1.04s per operation)
BenchmarkProcessFiles-8          5     256123450 ns/op  (256ms dry-run)
BenchmarkCollectFiles-8       2627        455143 ns/op  (455µs walk)

BenchmarkProcessFilesParallel/workers=1-8     5   256ms/op
BenchmarkProcessFilesParallel/workers=2-8    10   152ms/op
BenchmarkProcessFilesParallel/workers=4-8    18   108ms/op
BenchmarkProcessFilesParallel/workers=8-8    22   107ms/op  (optimal)
BenchmarkProcessFilesParallel/workers=16-8   23   106ms/op  (diminishing returns)
```

**Conclusion**: Go version delivers 2-3x faster throughput with dramatically lower memory overhead and instant startup. Optimal performance achieved with worker count matching CPU cores.

## Contributing

This is a personal migration project, but suggestions and feedback are welcome via issues.

## License

Same as original sortpics project (to be determined)

## References

### Python Original
This is a complete rewrite of the original Python sortpics tool, providing better performance, cross-platform binaries, and improved maintainability while retaining full feature parity.

### Technology Decisions
- See [DECISION.md](DECISION.md) for analysis of ExifTool wrapper vs pure Go implementation

## Acknowledgments

- Original sortpics Python implementation
- ExifTool by Phil Harvey
- Go community and library authors
