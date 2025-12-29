# sortpics-go

A fast, reliable photo and video organization tool that creates chronologically-organized archives using EXIF metadata.

## Features

- **Smart metadata extraction** - EXIF, QuickTime, filename, filesystem fallback chain
- **Organized output** - `YYYY/MM/YYYY-MM-DD/YYYYMMDD-HHMMSS.subsec_Make-Model.ext`
- **Duplicate detection** - SHA256-based content hashing
- **RAW & video support** - CR2, NEF, ARW, DNG, MOV, MP4, and more
- **Album tagging** - Set XMP:Album metadata
- **Parallel processing** - Worker pool with bounded queue
- **Archive verification** - Validate and fix existing archives

## Getting Started

### Prerequisites

Install [ExifTool](https://exiftool.org/):

```bash
# macOS
brew install exiftool

# Ubuntu/Debian
sudo apt-get install libimage-exiftool-perl

# Verify
exiftool -ver
```

### Installation

```bash
git clone https://github.com/cacack/sortpics-go.git
cd sortpics-go
make install    # Installs to ~/.local/bin
```

### Basic Usage

```bash
# Preview first (always recommended)
sortpics --copy --dry-run -v /source /destination

# Copy files
sortpics --copy --recursive /source /destination

# Move files (removes originals)
sortpics --move --recursive /source /destination

# Separate RAW files
sortpics --copy --recursive --raw-path /archive/raw /source /archive

# Verify archive integrity
sortpics verify /archive
```

See [USAGE.md](USAGE.md) for detailed examples and workflows.

## Example Output

```
/archive/
  2024/
    03/
      2024-03-15/
        20240315-143052.123456_Canon-EOS5D.jpg
        20240315-143052.123456_Canon-EOS5D.CR2
    12/
      2024-12-25/
        20241225-091530.000000_Nikon-D850.jpg
```

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions.

**Quick fixes:**
- ExifTool not found: Install per prerequisites above
- No files processed: Add `--recursive` flag
- Permission denied: Use `make install` (no sudo needed)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
