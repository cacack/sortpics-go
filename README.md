# sortpics-go

A fast, reliable photo and video organization tool that creates chronologically-organized archives using EXIF metadata.

**Status:** ✅ Production Ready (v0.1.0)

Complete Go rewrite of [sortpics](../sortpics) providing better performance, single-binary distribution, and 90.6% test coverage.

## Why sortpics-go?

- **2-3x faster** than Python original with native concurrency
- **Single binary** - no runtime dependencies
- **Cross-platform** - Linux, macOS, Windows (AMD64 & ARM64)
- **Battle-tested** - comprehensive test suite with real-world fixtures
- **Safe** - atomic operations with duplicate detection

## Features

- **Smart metadata extraction** - EXIF → QuickTime → filename → filesystem fallback
- **Organized output** - `YYYY/MM/YYYY-MM-DD/YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
- **Duplicate detection** - SHA256-based content hashing
- **RAW support** - All major formats (CR2, NEF, ARW, DNG, etc.)
- **Video support** - MOV, MP4 with QuickTime metadata
- **Album tagging** - Set XMP:Album metadata
- **Parallel processing** - Worker pool with bounded queue
- **Verification mode** - Validate existing archives

## Quick Start

### Installation

**Prerequisites:** [ExifTool](https://exiftool.org/) (install first)

```bash
# macOS
brew install exiftool

# Ubuntu/Debian
sudo apt-get install libimage-exiftool-perl

# Verify
exiftool -ver
```

**Install sortpics-go:**

```bash
# From source
git clone https://github.com/cacack/sortpics-go.git
cd sortpics-go
make install    # Installs to ~/.local/bin

# Verify
sortpics --version
```

### Basic Usage

```bash
# Preview what will happen (always start here!)
sortpics --copy --dry-run -v /source /destination

# Actually copy files
sortpics --copy --recursive /source /destination

# Move files (removes originals)
sortpics --move --recursive /source /destination

# Separate RAW files
sortpics --copy --recursive --raw-path /archive/raw /source /archive

# Verify archive integrity
sortpics verify /archive
```

## Example Output

**Input:** Messy collection of photos and videos

**Output:** Organized by date with descriptive filenames

```
/archive/
  2024/
    03/
      2024-03-15/
        20240315-143052.123456_Canon-EOS5D.jpg
        20240315-143052.123456_Canon-EOS5D.CR2
        20240315-180430.000000_Apple-iPhone14.mov
    12/
      2024-12-25/
        20241225-091530.000000_Nikon-D850.jpg
```

## Performance

**2-3x faster** than Python original with native concurrency and better memory efficiency.

See [CONTRIBUTING.md](CONTRIBUTING.md#benchmark-results) for detailed benchmarks.

## How It Works

sortpics-go uses a clean pipeline architecture:

```
File → Metadata Extraction → Path Generation → Duplicate Detection → Atomic Copy/Move
```

Built with [ExifTool](https://exiftool.org/) for comprehensive format support (500+ file types) and robust concurrency using worker pools. See [CONTRIBUTING.md](CONTRIBUTING.md#project-architecture) for details.

## Documentation

- **[USAGE.md](USAGE.md)** - Detailed usage guide with workflows
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development setup, testing, and benchmarks
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues and solutions
- **[DECISION.md](DECISION.md)** - Technology decisions (ExifTool vs Pure Go)
- **[CHANGELOG.md](CHANGELOG.md)** - Release history

## Common Use Cases

```bash
# Import from camera
sortpics --copy --recursive /Volumes/SDCARD ~/Photos

# Organize existing collection
sortpics --copy --recursive --dry-run /messy/photos /organized

# Professional workflow with RAW separation
sortpics --move --recursive \
  --raw-path /archive/raw \
  --album "Client X - Product Shoot" \
  /import /archive
```

See [USAGE.md](USAGE.md) for detailed examples and workflows.

## Troubleshooting

Common issues and solutions are covered in [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

**Quick fixes:**
- ExifTool not found: `brew install exiftool` (macOS) or `sudo apt-get install libimage-exiftool-perl` (Ubuntu)
- No files processed: Add `--recursive` flag for subdirectories
- Permission denied: Use `make install` (installs to ~/.local/bin, no sudo needed)

## Contributing

This is a personal project, but suggestions and bug reports are welcome via [issues](https://github.com/cacack/sortpics-go/issues).

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Same as original sortpics project (to be determined).

## Acknowledgments

- Original [sortpics](../sortpics) Python implementation
- [ExifTool](https://exiftool.org/) by Phil Harvey
- Go community and library authors
