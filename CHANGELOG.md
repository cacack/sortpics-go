# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Shell completion support for bash, zsh, fish, and PowerShell with installation instructions
- Man page for Unix systems (docs/sortpics.1)
- Performance comparison documentation vs Python original (2.4x faster throughput)
- Comprehensive troubleshooting guide in README

## [0.1.0] - 2025-10-16

### Added
- Complete CLI implementation with all core features
- Worker pool with bounded concurrency (default: CPU count)
- Progress bar with real-time updates (auto-hides in verbose mode)
- Cross-platform builds (Linux, macOS, Windows × amd64/arm64)
- ExifTool detection with helpful installation instructions
- Graceful shutdown on Ctrl-C/SIGTERM with context cancellation
- Comprehensive error handling with platform-specific messages
- Archive verification with `verify` subcommand
- Automatic fix mode (`verify --fix`) for mismatched filenames
- Empty directory cleanup (`--clean` flag with `--move`)
- Performance benchmarking suite
- Integration tests for all major features
- Comprehensive documentation with troubleshooting guide

### Features
- **Copy/Move Operations**: Atomic file handling with SafeCopy/SafeMove
- **Metadata Extraction**: EXIF → QuickTime → filename → filesystem fallback
- **Path Generation**: `YYYY/MM/YYYY-MM-DD/YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
- **Duplicate Detection**: SHA256-based content hashing with collision resolution
- **RAW File Separation**: Optional dedicated path for RAW files
- **Album Tagging**: Set XMP:Album metadata on processed files
- **Timestamp Adjustment**: Bulk time/day adjustments for timezone fixes
- **Dry-Run Mode**: Preview operations without modifying files
- **Recursive Processing**: Process directory trees with `--recursive`
- **Multi-Source Support**: Process multiple source directories in one run
- **Verbose Logging**: Three levels (-v, -vv, -vvv) for debugging

### Performance
- ~1.04s per operation for full copy with I/O (5 files)
- ~256ms per operation for dry-run processing (metadata only)
- ~455µs for directory walking and file collection
- Optimal performance with worker count matching CPU cores
- Low memory overhead (~292KB per operation)

### Test Coverage
- Overall: 90.6%
- CLI: 72.6%
- Duplicate: 86.8%
- PathGen: 100.0%
- Metadata: 94.4%
- Rename: 81.1%

### Supported Formats
- **Images**: JPEG, PNG, TIFF
- **RAW**: CR2, CRW, NEF, NRW, DNG, ARW, SRF, SR2, SRW, ORF, PEF, PTX, MRW, RW2, RWL, X3F
- **Video**: MP4, MOV (QuickTime)

### Dependencies
- github.com/barasher/go-exiftool v1.10.0
- github.com/spf13/cobra v1.10.1
- github.com/stretchr/testify v1.11.1
- github.com/alitto/pond v1.9.2
- github.com/schollz/progressbar/v3 v3.18.0

### System Requirements
- Go 1.21+ (for building from source)
- ExifTool (for metadata extraction)

## [0.0.1] - 2025-10-14

### Added
- Initial project setup
- Core component architecture
- Phase 1: Duplicate detection and path generation
- Phase 2: Metadata extraction with ExifTool
- Phase 3: File operations (copy/move/rename)
- Test fixtures and integration test framework
- Makefile with build automation
- Project documentation (README, STATUS, MIGRATION_PLAN)

### Technical Details
- Bottom-up migration from Python to Go
- Component-based architecture (duplicate, pathgen, metadata, rename)
- Comprehensive test coverage from day one
- ExifTool wrapper for broad format support

---

## Migration from Python

This is a complete rewrite of the original Python `sortpics` tool in Go, providing:
- **Better performance**: Native concurrency with goroutines (no GIL)
- **Single binary**: No Python runtime dependency
- **Cross-platform**: Easy distribution for Linux, macOS, Windows
- **Strong typing**: Better maintainability and tooling
- **Lower memory**: Efficient compiled binary

[Unreleased]: https://github.com/chris/sortpics-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/chris/sortpics-go/releases/tag/v0.1.0
[0.0.1]: https://github.com/chris/sortpics-go/releases/tag/v0.0.1
