# Technology Stack Decision: Pure Go vs ExifTool Wrapper

## TL;DR - The Choice

**Two viable paths identified:**

### Option A: Pure Go Stack ⭐ For Single Binary
- **Images**: `github.com/dsoprea/go-exif` (538 stars, comprehensive EXIF)
- **Videos**: `github.com/Eyevinn/mp4ff` (564 stars, clean API, auto epoch conversion)
- **Effort**: 15-20 hours implementation + ongoing maintenance
- **Benefit**: Single static binary, no external dependencies

### Option B: ExifTool Wrapper ⭐ For Simplicity & Reliability
- **Everything**: `github.com/barasher/go-exiftool` (266 stars, wraps ExifTool binary)
- **Effort**: 1-2 hours implementation
- **Benefit**: 500+ formats, battle-tested, write support

---

## Pure Go Video Libraries: Research Summary

### ✅ Winner: github.com/Eyevinn/mp4ff

**Why it's the best pure Go video option:**

```go
// Clean, simple API
file, _ := os.Open("video.mp4")
mp4File, _ := mp4.DecodeFile(file)

// Get creation time (already converted to Unix seconds!)
unixSeconds := mp4File.Moov.Mvhd.CreationTimeS()
creationTime := time.Unix(int64(unixSeconds), 0)
```

**Key Advantages:**
- Automatic Apple epoch → Unix conversion (saves manual math)
- High-level API (not low-level box parsing)
- Actively maintained (2025 updates)
- Production-proven by Eyevinn Technology
- Clean, idiomatic Go code

**Stats:**
- 564 stars, 106 forks
- Active development
- Comprehensive codec support
- MIT license

### Runner-up: github.com/abema/go-mp4

Good alternative but requires more manual work (epoch conversion, box traversal).

---

## Pure Go Solution: Implementation Analysis

### Architecture

```go
package metadata

type Extractor struct {
    // No external process needed
}

func (e *Extractor) Extract(filename string) (*Metadata, error) {
    ext := strings.ToLower(filepath.Ext(filename))

    // Route by format
    switch ext {
    case ".jpg", ".jpeg", ".tiff":
        return e.extractImage(filename)  // dsoprea/go-exif
    case ".mp4", ".mov", ".m4v":
        return e.extractVideo(filename)  // Eyevinn/mp4ff
    default:
        return nil, fmt.Errorf("unsupported: %s", ext)
    }
}

func (e *Extractor) extractImage(path string) (*Metadata, error) {
    // Use dsoprea/go-exif for JPEG/TIFF
    // Extract DateTimeOriginal, Make, Model, SubSec
    // ~50-100 lines
}

func (e *Extractor) extractVideo(path string) (*Metadata, error) {
    // Use Eyevinn/mp4ff for MP4/MOV
    // Extract mvhd.CreationTimeS()
    // ~30-50 lines
}
```

### Effort Breakdown

| Task | Hours | Notes |
|------|-------|-------|
| Image metadata (go-exif) | 4-6 | Complex API, EXIF structure |
| Video metadata (mp4ff) | 2-3 | Clean API, straightforward |
| Fallback logic | 2-3 | Filename → ctime chain |
| Make/model normalization | 1-2 | String processing |
| Testing | 8-12 | Edge cases, various formats |
| **Total** | **17-26h** | **2-3 days focused work** |

### What You Get

**Pros:**
- ✅ **Single binary** - No ExifTool dependency
- ✅ **Cross-compile** - `GOOS=linux GOARCH=amd64 go build`
- ✅ **Fast startup** - No external process spawn
- ✅ **Pure Go** - Simpler deployment story
- ✅ **Full control** - Understand every line

**Cons:**
- ❌ **Limited formats** - Only JPEG, TIFF, MP4, MOV (no RAW out of box)
- ❌ **No write support** for videos - Can't set XMP:Album on MP4/MOV
- ❌ **Maintenance burden** - You own the edge cases
- ❌ **Less comprehensive** - 5-10 formats vs 500+
- ❌ **Reinventing wheel** - ExifTool already solved this

### RAW File Problem

Your Python version supports:
- Canon: CR2, CRW
- Nikon: NEF, NRW
- Sony: ARW, SRF, SR2
- DNG, ORF, PEF, RW2, etc.

**Pure Go options for RAW:**
- `dsoprea/go-exif` can extract EXIF from some RAW formats
- But requires format-specific parsers (JPEG structure vs TIFF-based RAW)
- Each RAW format needs separate implementation/testing
- **Significant additional effort**: +10-15 hours

---

## ExifTool Wrapper: Implementation Analysis

### Architecture

```go
package metadata

import "github.com/barasher/go-exiftool"

type Extractor struct {
    et *exiftool.Exiftool
}

func NewExtractor() (*Extractor, error) {
    et, err := exiftool.NewExiftool()
    if err != nil {
        return nil, fmt.Errorf("exiftool not found: %w", err)
    }
    return &Extractor{et: et}, nil
}

func (e *Extractor) Extract(filename string) (*Metadata, error) {
    fileInfos := e.et.ExtractMetadata(filename)

    fileInfo := fileInfos[0]
    if fileInfo.Err != nil {
        return nil, fileInfo.Err
    }

    // Extract with fallback logic
    datetime := e.extractDateTime(fileInfo.Fields)
    make, model := e.extractMakeModel(fileInfo.Fields)

    return &Metadata{
        DateTime: datetime,
        Make:     make,
        Model:    model,
    }, nil
}
```

### Effort Breakdown

| Task | Hours | Notes |
|------|-------|-------|
| ExifTool integration | 0.5 | Install + basic usage |
| Datetime extraction | 1 | Fallback logic |
| Make/model extraction | 0.5 | String processing |
| Write support | 0.5 | XMP:Album tags |
| Testing | 2-3 | Various formats |
| **Total** | **4-5h** | **Half day work** |

### What You Get

**Pros:**
- ✅ **500+ formats** - JPEG, RAW (CR2, NEF, etc), MP4, MOV, PDF, etc.
- ✅ **Battle-tested** - 20+ years of ExifTool development
- ✅ **Edge cases handled** - Timezones, corrupted EXIF, camera quirks
- ✅ **Write support** - Can write XMP:Album to any format
- ✅ **Fast development** - ~5 hours vs ~20+ hours
- ✅ **Less code** - ~100 lines vs ~500+ lines
- ✅ **Less maintenance** - ExifTool updates handle new formats

**Cons:**
- ❌ **External dependency** - Requires ExifTool binary installed
- ❌ **Deployment complexity** - Users must `brew install exiftool`
- ❌ **Startup overhead** - ~50-100ms to spawn process (mitigated by stay_open)
- ❌ **Not pure Go** - Can't claim "single binary"

---

## Decision Matrix

### By Priority

| Priority | Pure Go | ExifTool Wrapper | Winner |
|----------|---------|------------------|--------|
| **Single binary deployment** | ✅ Yes | ❌ No | Pure Go |
| **Format coverage** | ⚠️ Limited (5-10) | ✅ Comprehensive (500+) | Wrapper |
| **Development speed** | ❌ Slow (20h+) | ✅ Fast (5h) | Wrapper |
| **Maintenance burden** | ❌ High | ✅ Low | Wrapper |
| **Write metadata** | ⚠️ Images only | ✅ All formats | Wrapper |
| **RAW file support** | ⚠️ Limited | ✅ Full | Wrapper |
| **Edge case handling** | ❌ You handle | ✅ ExifTool handles | Wrapper |
| **Cross-platform** | ✅ Easy | ⚠️ Requires ExifTool | Pure Go |

### By Use Case

| Use Case | Recommendation | Why |
|----------|----------------|-----|
| **Photo organizer tool (sortpics)** | ExifTool Wrapper | Need RAW support, write capability, comprehensive formats |
| **Embedded system** | Pure Go | No package manager, single binary critical |
| **Cloud service** | Pure Go | Container size matters, predictable dependencies |
| **Professional photographer** | ExifTool Wrapper | Need all RAW formats, reliable metadata |
| **Learning project** | Pure Go | Educational value in implementing parsers |
| **Quick prototype** | ExifTool Wrapper | Get working fast |

---

## Specific to SortPics

### Your Requirements (from Python version)

1. ✅ **EXIF:DateTimeOriginal** - Both support
2. ✅ **QuickTime:CreateDate** - Both support
3. ✅ **Make/Model extraction** - Both support
4. ⚠️ **RAW file support** - Wrapper wins (CR2, NEF, DNG, etc.)
5. ⚠️ **Write XMP:Album** - Wrapper wins (can write to videos)
6. ✅ **Fallback hierarchy** - Both can implement
7. ✅ **Subsecond precision** - Both support

### Your Current Python Implementation

Uses ExifTool binary via subprocess:
```python
# sortpics already requires ExifTool
# Users must have it installed
```

**Implication:** You already have the ExifTool deployment "problem" in Python version. Go wrapper doesn't make it worse.

### Migration Consideration

**If you choose Pure Go:**
- Breaking change for users (features change)
- Need migration guide: "Go version supports fewer formats"
- Users with CR2/NEF/ARW files may be disappointed

**If you choose ExifTool Wrapper:**
- Drop-in replacement for Python version
- Same capabilities, better performance
- Users already have ExifTool installed

---

## Performance Comparison

### Startup Time

| Approach | First File | Batch (100 files) | Notes |
|----------|-----------|-------------------|-------|
| Pure Go | ~1-5ms | ~1-5ms each | No process spawn |
| ExifTool Wrapper | ~50-100ms | ~5-10ms each | stay_open reuses process |

**Real-world impact:**
- Organizing 1000 files: Pure Go saves ~5 seconds total
- But you're doing file I/O (copy/move) which takes minutes
- **Metadata extraction is <5% of total time**

### Memory Usage

| Approach | Overhead | Notes |
|----------|----------|-------|
| Pure Go | Low (~10MB) | Just Go binary |
| ExifTool Wrapper | Medium (~50MB) | ExifTool process + IPC |

**Real-world impact:**
- Both trivial on modern systems
- Not a deciding factor

---

## Recommendation

### For SortPics: Use ExifTool Wrapper ✅

**Rationale:**

1. **Feature parity**: Your Python version supports CR2, NEF, ARW, etc. Users expect this.
2. **Write support**: You write XMP:Album tags. Pure Go can't do this for videos.
3. **Development velocity**: 5 hours vs 20+ hours to first working version.
4. **Reliability**: 20 years of ExifTool bug fixes vs your implementation.
5. **Deployment**: Users already need ExifTool for Python version.

### When to Reconsider Pure Go

**Choose Pure Go if:**
- [ ] You're okay dropping RAW support (CR2, NEF, etc.)
- [ ] You're okay with read-only video metadata
- [ ] Single binary is critical to your use case
- [ ] You're targeting embedded systems
- [ ] You want the learning experience

For a photo organization tool aimed at photographers, **ExifTool wrapper is the pragmatic choice.**

---

## Implementation Plan

### If Choosing ExifTool Wrapper (Recommended)

```bash
# 1. Add dependency
cd /Users/chris/devel/home/sortpics-go
go get github.com/barasher/go-exiftool

# 2. Implement metadata package (~5 hours)
# internal/metadata/metadata.go

# 3. Port tests from Python (~2 hours)
# internal/metadata/metadata_test.go

# 4. Integration test with real files (~1 hour)
```

**Total effort:** ~8 hours to fully working metadata extraction

### If Choosing Pure Go

```bash
# 1. Add dependencies
go get github.com/dsoprea/go-exif/v3
go get github.com/Eyevinn/mp4ff

# 2. Implement image extractor (~6 hours)
# internal/metadata/image.go

# 3. Implement video extractor (~3 hours)
# internal/metadata/video.go

# 4. Implement fallback logic (~3 hours)
# internal/metadata/metadata.go

# 5. Port and expand tests (~12 hours)
# internal/metadata/metadata_test.go

# 6. Handle edge cases (~10 hours)
# Corrupted files, missing EXIF, timezone issues
```

**Total effort:** ~34 hours to production-ready metadata extraction

---

## Your Decision

**What matters most to you?**

### Choose Pure Go if:
- Single binary is non-negotiable
- You're deploying to environments without package managers
- You're okay with limited format support (JPEG, MP4, MOV only)
- You want maximum control over implementation
- Learning/educational value is important

### Choose ExifTool Wrapper if:
- You want feature parity with Python version
- RAW file support is important (CR2, NEF, DNG, etc.)
- Fast development is priority
- Reliability/battle-tested code matters
- You want to focus on features, not metadata parsing

---

## Next Steps

**Please decide:**

1. **Go with ExifTool wrapper** (recommended for sortpics)
   - I'll install `barasher/go-exiftool`
   - Implement metadata extraction (~5 hours)
   - Port tests from Python
   - Ready to proceed with migration

2. **Go with Pure Go** (for single binary benefits)
   - I'll install `dsoprea/go-exif` + `Eyevinn/mp4ff`
   - Implement dual extraction paths (~20+ hours)
   - Document format limitations
   - Ready for longer implementation

**What's your preference?**
