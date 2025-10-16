# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

sortpics-go is a Go port of a Python photo/video organization tool. It organizes media files into a chronological directory structure using EXIF metadata with a fallback hierarchy: EXIF → QuickTime → filename → filesystem.

**Status**: Phase 2 complete (duplicate: 86.8%, pathgen: 97.6%, metadata: 73.3%). Ready for Phase 3 (file operations).

## Build & Test Commands

```bash
# Build
make build              # Build binary to ./bin/sortpics
make install            # Install to $GOPATH/bin

# Testing
make test               # Run all tests with race detector
make test-verbose       # Verbose test output
make test-coverage      # Generate coverage.html report
make bench              # Run benchmarks

# Code Quality
make lint               # Run golangci-lint (requires: brew install golangci-lint)
make fmt                # Format all code
make vet                # Run go vet
make check              # Run fmt + vet + test

# Development
make run ARGS="--help"                      # Build and run
make run-dev ARGS="--dry-run /src /dest"   # Run with go run (faster iteration)
make deps               # Download dependencies
make tidy               # Tidy go modules
make clean              # Remove build artifacts
```

## Architecture

### Component Hierarchy (Bottom-up)

1. **internal/duplicate** - ✅ COMPLETE - SHA256 hashing, duplicate detection, collision resolution (`_N` suffix)
2. **internal/pathgen** - ✅ COMPLETE - Generates destination paths in format: `YYYY/MM/YYYY-MM-DD/YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
3. **internal/metadata** - ✅ COMPLETE - EXIF extraction via ExifTool wrapper, datetime fallback logic, make/model normalization
4. **internal/rename** - PENDING - Orchestrates metadata → pathgen → duplicate → file operations (atomic copy/move)
5. **cmd/sortpics/cmd** - PENDING - CLI framework (Cobra), worker pool orchestration, progress tracking

### Key Design Principles

- **ExifTool dependency**: Uses `github.com/barasher/go-exiftool` wrapper for 500+ format support (JPEG, RAW, MP4, MOV, etc.). Users must have ExifTool installed. See DECISION.md for rationale vs pure Go.
- **Atomic operations**: Uses temp files for safe copy/move with cleanup on error
- **Bounded concurrency**: Will use `github.com/alitto/pond` worker pool (Phase 4) with backpressure
- **High test coverage target**: 90%+ overall (Python original: 95%)

### Data Flow

```
File → MetadataExtractor → ImageMetadata → PathGenerator → destination path
                                         ↓
                                  DuplicateDetector → collision check → final path
                                         ↓
                                   ImageRename → atomic copy/move
```

## Migration Strategy

**Porting from Python**: This is a component-by-component migration from `../sortpics`. When implementing:

1. Reference Python source in `../sortpics/sortpics/`
2. Reference Python tests in `../sortpics/tests/`
3. Port test cases first (test-driven)
4. Match Python behavior for identical inputs
5. Maintain or exceed Python test coverage

**Current phase**: Phase 3 - File operations (Phases 1-2 complete: duplicate 86.8%, pathgen 97.6%, metadata 73.3%)

**Next phases**: See MIGRATION_PLAN.md for detailed 6-phase roadmap

## Testing Guidelines

- Use `github.com/stretchr/testify` for assertions: `assert.Equal(t, expected, actual)` and `require.NoError(t, err)`
- Test file naming: `*_test.go` in same package
- Use `test/testdata/` for fixtures
- Mock ExifTool responses for metadata tests
- Target: 95%+ for core packages (duplicate, pathgen, metadata), 90%+ overall

## Important Files

- **MIGRATION_PLAN.md** - Detailed 6-phase implementation plan with tasks and estimates
- **STATUS.md** - Current progress tracker, completed setup, next steps
- **DECISION.md** - Technology decisions (ExifTool wrapper vs pure Go analysis)
- **RESEARCH.md** - Library evaluation rationale
- **Makefile** - All build targets with help text (`make help`)

## Dependencies

**Installed**:
- `github.com/barasher/go-exiftool` v1.10.0 - ExifTool wrapper (requires `brew install exiftool`)
- `github.com/spf13/cobra` v1.10.1 - CLI framework
- `github.com/stretchr/testify` v1.11.1 - Testing assertions

**To be installed** (Phase 4):
- `github.com/alitto/pond` - Worker pool with bounded queue

**System requirements**:
- Go 1.21+ (using 1.24.3)
- ExifTool binary (`exiftool -ver` to verify)

## CLI Usage (Target)

```bash
# Copy with preview
sortpics --copy --dry-run --recursive /sdcard /photos

# Move files
sortpics --move --recursive -vvv /sdcard /photos

# Separate RAW files
sortpics --copy --raw-path /photos/raw /sdcard /photos

# Set album metadata
sortpics --copy --album "Vacation 2024" /sdcard /photos

# Verify archive
sortpics verify /photos
sortpics verify --fix /photos
```

## Output Format

**Filename**: `YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
- Example: `20240315-143052.123456_Canon-EOS5D.jpg`
- Subsecond precision configurable (default: 6 digits)
- Make/Model normalized (capitalized, make prefix removed from model)

**Directory**: `YYYY/MM/YYYY-MM-DD/`
- Example: `2024/03/2024-03-15/`

**Collision resolution**: Append `_N` suffix (e.g., `filename_2.jpg`)

## Development Notes

- Cross-filesystem moves: Handle `syscall.EXDEV` by falling back to copy+delete
- RAW formats: CR2, NEF, DNG, ARW, etc. (via ExifTool)
- Video support: MP4, MOV (QuickTime:CreateDate metadata)
- Write capability: Can set XMP:Album and keywords on any format
- Atomic operations: Use `os.CreateTemp()` with UUID names, cleanup on error
- Use conventional commit.
- Commit after completing a new phase or when it makes sense to ensure logical commits.