# Usage Guide

Complete guide to using sortpics-go for organizing your photo and video archives.

## Table of Contents

- [Basic Commands](#basic-commands)
- [Common Workflows](#common-workflows)
- [Advanced Features](#advanced-features)
- [Archive Verification](#archive-verification)
- [Output Options](#output-options)
- [Shell Completion](#shell-completion)

## Basic Commands

### Preview Before Processing

Always start with `--dry-run` to preview what will happen:

```bash
# Preview copy operation
sortpics --copy --dry-run -v /source/photos /archive

# Preview move operation
sortpics --move --dry-run -v /source/photos /archive
```

### Copy vs Move

```bash
# Copy files (leaves originals intact)
sortpics --copy /source/photos /archive

# Move files (removes originals after successful copy)
sortpics --move /source/photos /archive
```

### Recursive Processing

```bash
# Process subdirectories recursively
sortpics --copy --recursive /source/photos /archive

# Short form
sortpics --copy -r /source/photos /archive
```

## Common Workflows

### Importing from SD Card

```bash
# Basic import with progress bar
sortpics --copy --recursive /Volumes/SDCARD /Users/me/Photos

# With verbose output
sortpics --copy --recursive -v /Volumes/SDCARD /Users/me/Photos

# Preview first
sortpics --copy --recursive --dry-run /Volumes/SDCARD /Users/me/Photos
```

### Separating RAW Files

Keep RAW files in a separate directory tree:

```bash
sortpics --copy --recursive \
  --raw-path /archive/raw \
  /sdcard /archive
```

Result structure:
```
/archive/
  2024/03/2024-03-15/20240315-143052_Canon-EOS5D.jpg
/archive/raw/
  2024/03/2024-03-15/20240315-143052_Canon-EOS5D.CR2
```

### Setting Album Metadata

Tag all imported files with an album name:

```bash
sortpics --copy --album "Summer Vacation 2024" /import /archive
```

This writes `XMP:Album` metadata to each file.

### Timestamp Adjustments

#### Fix Camera Timezone

If your camera was set to the wrong timezone:

```bash
# Subtract 5 hours
sortpics --copy --time-adjust -05:00:00 /import /archive

# Add 3 hours and 30 minutes
sortpics --copy --time-adjust +03:30:00 /import /archive
```

#### Adjust by Days

Shift dates forward or backward:

```bash
# Add 1 day
sortpics --copy --day-adjust 1 /import /archive

# Subtract 7 days
sortpics --copy --day-adjust -7 /import /archive
```

### Cleanup Empty Directories

Remove empty source directories after moving files:

```bash
sortpics --move --recursive --clean /sdcard /archive
```

### Alternative Filename Format

Use the legacy filename format:

```bash
sortpics --copy --old-naming /source /dest
```

### Subsecond Precision

Control timestamp precision in filenames:

```bash
# No subseconds (YYYYMMDD-HHMMSS)
sortpics --copy --precision 0 /source /dest

# 3 digits (YYYYMMDD-HHMMSS.123)
sortpics --copy --precision 3 /source /dest

# 6 digits - default (YYYYMMDD-HHMMSS.123456)
sortpics --copy --precision 6 /source /dest
```

## Advanced Features

### Parallel Processing

Control the number of concurrent workers:

```bash
# Use 4 workers (default is CPU count)
sortpics --copy --workers 4 /source /dest

# Single-threaded processing
sortpics --copy --workers 1 /source /dest
```

### File Extension Filtering

Process only specific file types:

```bash
# Only JPEG files
sortpics --copy --extensions .jpg,.jpeg /source /dest

# Only RAW files
sortpics --copy --extensions .cr2,.nef,.arw /source /dest
```

## Archive Verification

### Check Archive Integrity

Verify that filenames match EXIF metadata:

```bash
# Check entire archive
sortpics verify /archive

# Check specific subdirectory
sortpics verify /archive/2024
```

Output shows:
- ✓ Files with matching metadata
- ✗ Files with mismatches (wrong timestamp or make/model)
- Files with no EXIF data

### Find Mismatches

Filter verification output to show only problems:

```bash
sortpics verify /archive 2>&1 | grep MISMATCH
```

### Automatically Fix Mismatches

Rename files to match their EXIF data:

```bash
sortpics verify --fix /archive
```

This will rename files in place to match their actual EXIF timestamps and make/model.

## Output Options

### Verbosity Levels

Control how much information is displayed:

```bash
# Silent mode (progress bar only)
sortpics --copy /source /dest

# Basic info (-v)
sortpics --copy -v /source /dest
# Shows: files processed, skipped, errors

# Detailed (-vv)
sortpics --copy -vv /source /dest
# Shows: each file operation, metadata extraction

# Debug (-vvv)
sortpics --copy -vvv /source /dest
# Shows: everything including internal operations
```

### Progress Bar

Progress bar displays automatically in non-verbose mode:

```
Processing files: 100% |████████████████| (1234/1234, 45 files/s)
```

Disabled when using `-v` or higher verbosity.

## Shell Completion

sortpics includes built-in shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Install for current user
sortpics completion bash > ~/.local/share/bash-completion/completions/sortpics

# Install system-wide (requires sudo)
sortpics completion bash | sudo tee /usr/share/bash-completion/completions/sortpics > /dev/null

# Reload shell
source ~/.local/share/bash-completion/completions/sortpics
```

### Zsh

```bash
# Generate and install
sortpics completion zsh > "${fpath[1]}/_sortpics"

# Or add to .zshrc for auto-generation
echo 'source <(sortpics completion zsh)' >> ~/.zshrc

# Reload shell
exec zsh
```

### Fish

```bash
# Generate and install
sortpics completion fish > ~/.config/fish/completions/sortpics.fish

# Or add to config.fish
echo 'sortpics completion fish | source' >> ~/.config/fish/config.fish

# Reload shell
exec fish
```

### PowerShell

```powershell
# Generate completion script
sortpics completion powershell > sortpics-completion.ps1

# Add to profile
sortpics completion powershell >> $PROFILE

# Reload profile
. $PROFILE
```

### Completion Features

- Tab completion for all commands (root, verify)
- Flag completion with descriptions
- Path completion for source/destination arguments
- Completion for flag values where applicable

## Output Format

### Filename Structure

```
YYYYMMDD-HHMMSS.subsec_Make-Model.ext
```

Examples:
- `20240315-143052.123456_Canon-EOS5D.jpg`
- `20231220-091530.000000_Apple-iPhone14.mov`
- `20240701-180000.000000_Unknown.jpg` (no EXIF)

### Directory Structure

```
YYYY/MM/YYYY-MM-DD/
```

Example:
```
/archive/
  2024/
    03/
      2024-03-15/
        20240315-143052.123456_Canon-EOS5D.jpg
        20240315-143052.123456_Canon-EOS5D.CR2
    12/
      2024-12-20/
        20241220-091530.000000_Apple-iPhone14.mov
```

### Duplicate Handling

Files with identical content (SHA256 hash) are skipped. Files with identical filenames but different content get a suffix:

```
20240315-143052_Canon-EOS5D.jpg     # Original
20240315-143052_Canon-EOS5D_2.jpg   # Collision
20240315-143052_Canon-EOS5D_3.jpg   # Another collision
```

### Make/Model Normalization

Camera makes and models are normalized for consistent filenames:

| Original | Normalized |
|----------|------------|
| `NIKON CORPORATION` | `Nikon` |
| `Canon EOS 5D Mark IV` | `Canon-EOS5DMarkIV` |
| `Apple iPhone 14 Pro` | `Apple-iPhone14Pro` |
| `SONY ILCE-7M3` | `Sony-ILCE7M3` |

Spaces are removed, redundant make prefixes are stripped, and capitalization is normalized.

## Examples by Use Case

### Professional Photographer Workflow

```bash
# 1. Import from multiple cards to staging
sortpics --copy --recursive /Volumes/CF_CARD /import/shoot-2024-03-15
sortpics --copy --recursive /Volumes/SD_CARD /import/shoot-2024-03-15

# 2. Review and cull in staging directory

# 3. Separate RAW and JPEG to archive with album tag
sortpics --move --recursive \
  --raw-path /archive/raw \
  --album "Client X - Product Shoot" \
  /import/shoot-2024-03-15 \
  /archive

# 4. Verify archive
sortpics verify /archive/2024/03
```

### Family Photo Collection

```bash
# Import phone photos
sortpics --copy --recursive \
  --album "Summer 2024" \
  /Users/me/Pictures/iPhone \
  /Users/me/Photos/Archive

# Import camera photos
sortpics --copy --recursive \
  /Volumes/SDCARD \
  /Users/me/Photos/Archive

# Verify everything
sortpics verify /Users/me/Photos/Archive
```

### Organizing Existing Collection

```bash
# Fix messy directory with dry-run first
sortpics --copy --recursive --dry-run /messy/photos /organized

# Actually organize
sortpics --copy --recursive /messy/photos /organized

# Verify
sortpics verify /organized

# If all looks good, remove originals
rm -rf /messy/photos
```
