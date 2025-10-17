# Troubleshooting Guide

Common issues and solutions when using sortpics-go.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Processing Issues](#processing-issues)
- [Performance Issues](#performance-issues)
- [File Issues](#file-issues)
- [Output Issues](#output-issues)

## Installation Issues

### ExifTool Not Found

**Error:**
```
exiftool not found. Please install it first
```

**Cause:** ExifTool binary is not installed or not in PATH.

**Solution:**

```bash
# macOS
brew install exiftool

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install libimage-exiftool-perl

# Arch Linux
sudo pacman -S perl-image-exiftool

# Windows
# Download from https://exiftool.org/
# Extract exiftool(-k).exe to a directory in your PATH

# Verify installation
exiftool -ver
```

**Expected output:** Version 12.00 or higher

**Still not working?**

Check if ExifTool is in your PATH:
```bash
which exiftool   # macOS/Linux
where exiftool   # Windows
```

If not found, add the installation directory to your PATH or specify the full path when running sortpics.

### Build Fails with "Go Version Too Old"

**Error:**
```
go: module requires go >= 1.21
```

**Solution:**

Update Go to version 1.21 or higher:
```bash
# macOS
brew upgrade go

# Ubuntu (using snap)
sudo snap refresh go --classic

# Or download from https://go.dev/dl/

# Verify version
go version
```

### Permission Denied During Installation

**Error:**
```
permission denied: /usr/local/bin/sortpics
```

**Solution:**

Install to user directory instead:
```bash
# Install to ~/.local/bin (no sudo needed)
make install

# Ensure ~/.local/bin is in PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Or use `sudo` for system-wide installation:
```bash
sudo make install-global
```

## Processing Issues

### No Files Found

**Error:**
```
Found 0 files to process
```

**Possible causes:**

1. **Wrong source path**
   ```bash
   # Check directory exists
   ls -la /path/to/source
   ```

2. **No supported file extensions**
   ```bash
   # Check what files are there
   ls /path/to/source

   # Supported extensions include:
   # Images: .jpg, .jpeg, .png, .tif, .tiff, .heic
   # RAW: .cr2, .nef, .arw, .dng, .raf, .orf
   # Video: .mov, .mp4, .avi, .m4v
   ```

3. **Files in subdirectories (need --recursive flag)**
   ```bash
   # Add --recursive to process subdirectories
   sortpics --copy --recursive /source /dest
   ```

### Permission Denied on Source Files

**Error:**
```
failed to read file: permission denied
```

**Solution:**

```bash
# Check file permissions
ls -l /path/to/file

# Fix permissions if you own the files
chmod 644 /path/to/file

# Or run with sudo (not recommended)
sudo sortpics --copy /source /dest
```

### Permission Denied on Destination

**Error:**
```
failed to create destination directory: permission denied
```

**Solution:**

```bash
# Check destination permissions
ls -ld /destination/path

# Create destination first with correct permissions
mkdir -p /destination/path

# Or use a destination you have write access to
sortpics --copy /source ~/Photos/Archive

# As last resort, use sudo (not recommended)
sudo sortpics --copy /source /dest
```

### Files Skipped with "Already Exists"

**Message:**
```
SKIP: /dest/2024/03/2024-03-15/20240315-143052_Canon-EOS5D.jpg (already exists)
```

**This is normal behavior** - sortpics skips files that already exist at the destination with the same content (SHA256 hash). This prevents duplicate imports.

**If you want to re-import:**
```bash
# Remove existing files first
rm -rf /dest/2024/03/2024-03-15/

# Or use a different destination
sortpics --copy /source /dest-new
```

### Operation Cancelled

**Message:**
```
Context cancelled: processing stopped
```

**Cause:** User pressed Ctrl-C or context timeout.

**This is normal** - sortpics supports graceful cancellation. Files already copied remain in place.

**To resume:**
```bash
# Run the command again - already-copied files will be skipped
sortpics --copy /source /dest
```

## Performance Issues

### Processing Is Very Slow

**Possible causes:**

1. **Too many workers for your system**
   ```bash
   # Reduce worker count
   sortpics --copy --workers 4 /source /dest
   ```

2. **Slow disk I/O (USB 2.0, network drive, etc.)**
   ```bash
   # Check disk speed with a test copy
   time cp /source/large-file.jpg /dest/

   # Consider copying to local disk first
   sortpics --copy /usb-drive ~/temp-import
   sortpics --copy ~/temp-import /final-destination
   ```

3. **Many small files**
   ```bash
   # ExifTool overhead is higher for small files
   # This is expected behavior
   ```

4. **Verbose logging enabled**
   ```bash
   # Disable verbose logging for better performance
   sortpics --copy /source /dest
   # (no -v flag)
   ```

### High Memory Usage

**Cause:** Processing many files with high worker count.

**Solution:**

```bash
# Reduce worker count
sortpics --copy --workers 2 /source /dest

# Process in smaller batches
sortpics --copy /source/batch1 /dest
sortpics --copy /source/batch2 /dest
```

### Progress Bar Not Updating

**Cause:** Progress bar only updates when files complete processing.

**If stuck for a long time:**
```bash
# Use verbose mode to see what's happening
sortpics --copy -v /source /dest

# Check if ExifTool is running
ps aux | grep exiftool
```

## File Issues

### Files Have Wrong Timestamps

**Possible causes:**

1. **Camera timezone was wrong**
   ```bash
   # Adjust time by offset
   sortpics --copy --time-adjust -05:00:00 /source /dest
   ```

2. **Camera date was wrong**
   ```bash
   # Adjust by days
   sortpics --copy --day-adjust -7 /source /dest
   ```

3. **No EXIF data - using filesystem time**
   ```bash
   # Check if file has EXIF
   exiftool /path/to/file.jpg | grep DateTime

   # Expected files without EXIF get filesystem time
   # This is normal for screenshots, edited images, etc.
   ```

### Duplicates Not Detected

**Issue:** Files you think are duplicates are not being skipped.

**Explanation:** Duplicate detection is **content-based** (SHA256 hash), not filename-based.

**Files are only duplicates if:**
- They have **identical content** (byte-for-byte)
- They hash to the same SHA256

**To verify:**
```bash
# Check if files are truly identical
sha256sum file1.jpg file2.jpg

# Different hashes = not duplicates (will both be copied)
# Same hash = duplicates (second will be skipped)
```

**Similar but edited files are NOT duplicates:**
- Different JPEG quality
- Different resolution
- Different edits/crops
- Different metadata only

### Filename Collisions

**Issue:** Files with same timestamp and make/model create collisions.

**This is expected** - collisions get a suffix:
```
20240315-143052_Canon-EOS5D.jpg      # Original
20240315-143052_Canon-EOS5D_2.jpg    # Collision
20240315-143052_Canon-EOS5D_3.jpg    # Another
```

**This happens when:**
- Burst mode photos have same timestamp (subsecond precision helps)
- Multiple cameras with same make/model
- Time was reset on camera

**Solutions:**
```bash
# Increase subsecond precision (default is 6 digits)
# This helps with burst mode photos
sortpics --copy --precision 6 /source /dest

# Collisions are normal and handled automatically
```

### Filename Too Long

**Error:**
```
file name too long
```

**Cause:** Generated filename exceeds filesystem limit (usually 255 characters).

**This can happen with:**
- Very long camera model names
- High subsecond precision
- Long file extensions

**Solutions:**

```bash
# Use old naming format (shorter)
sortpics --copy --old-naming /source /dest

# Reduce subsecond precision
sortpics --copy --precision 3 /source /dest

# Or no subseconds
sortpics --copy --precision 0 /source /dest
```

### RAW Files Not Recognized

**Issue:** RAW files not being processed.

**Check:**
```bash
# Verify ExifTool can read the RAW format
exiftool /path/to/file.CR2

# Should show EXIF data
```

**Supported RAW formats:**
- Canon: .CR2, .CR3
- Nikon: .NEF
- Sony: .ARW
- Adobe: .DNG
- Fuji: .RAF
- Olympus: .ORF
- Panasonic: .RW2
- Pentax: .PEF

**If not supported:**
```bash
# Update ExifTool to latest version
brew upgrade exiftool  # macOS
```

### Video Files Not Working

**Issue:** Video files not being processed.

**Check metadata:**
```bash
# Check if video has QuickTime metadata
exiftool /path/to/video.mov | grep -i create

# Should show QuickTime:CreateDate
```

**Supported video formats:**
- QuickTime: .mov
- MPEG-4: .mp4, .m4v
- AVI: .avi

**Videos without QuickTime:CreateDate** will fall back to filesystem time.

## Output Issues

### Progress Bar Interferes with Other Output

**Solution:** Progress bar auto-disables in verbose mode.

```bash
# Use verbose mode for logging instead of progress bar
sortpics --copy -v /source /dest
```

### No Output at All

**Cause:** Default mode shows only progress bar.

**Solution:**
```bash
# Use verbose mode
sortpics --copy -v /source /dest

# Or debug mode
sortpics --copy -vvv /source /dest
```

### Too Much Debug Output

**Solution:**
```bash
# Reduce verbosity
sortpics --copy -v /source /dest    # Basic info
sortpics --copy /source /dest       # Progress bar only
```

### Verification Shows Mismatches

**Output:**
```
✗ MISMATCH: file.jpg (filename: 2024-03-15, EXIF: 2024-03-16)
```

**This means:** Filename doesn't match actual EXIF data.

**Causes:**
- Files were renamed manually
- Files were copied from another system
- EXIF data was edited after import

**Solution:**
```bash
# Fix mismatches automatically
sortpics verify --fix /archive

# Or investigate specific files
exiftool /path/to/file.jpg | grep DateTime
```

## Platform-Specific Issues

### macOS: "sortpics" Cannot Be Opened

**Error:**
```
"sortpics" cannot be opened because the developer cannot be verified
```

**Solution:**
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /path/to/sortpics

# Or build from source
make install
```

### Windows: ExifTool Not Found

**Issue:** ExifTool installed but not found.

**Solution:**
```bash
# Add ExifTool to PATH
# Or rename exiftool(-k).exe to exiftool.exe

# Verify
exiftool -ver
```

### Linux: Permission Denied on Binary

**Solution:**
```bash
# Make binary executable
chmod +x /path/to/sortpics

# Or reinstall
make install
```

## Cross-Filesystem Issues

### "Invalid Cross-Device Link" Error

**Error:**
```
invalid cross-device link
```

**This should be handled automatically** by falling back to copy+delete.

**If you see this error**, please report it as a bug with:
```bash
# Source and destination filesystems
df -T /source
df -T /destination

# Operating system
uname -a
```

## Getting Help

### Gathering Debug Information

When reporting issues, include:

```bash
# Version
sortpics --version

# ExifTool version
exiftool -ver

# Operating system
uname -a  # macOS/Linux
systeminfo | findstr /B /C:"OS"  # Windows

# Go version (if building from source)
go version

# Debug output
sortpics --copy -vvv /source /dest 2>&1 | tee debug.log
```

### Reporting Bugs

Open an issue at: https://github.com/cacack/sortpics-go/issues

Include:
1. Command you ran
2. Error message or unexpected behavior
3. Debug information (see above)
4. Sample file (if applicable and not confidential)

### Common Mistakes

1. **Forgetting --recursive flag** - Files in subdirectories won't be processed
2. **Using --move without --dry-run first** - Always preview first
3. **Wrong source/destination order** - It's always: `sortpics [flags] SOURCE DEST`
4. **Expecting filename-based duplicate detection** - It's content-based (SHA256)
5. **Not checking ExifTool is installed** - Run `exiftool -ver` first
