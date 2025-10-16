# Test Fixtures

This directory contains test images with known EXIF metadata for integration testing.

## Generated Files

**Total**: 15 files (14 JPEG, 1 PNG)
**Size**: ~115KB

### Basic Test Fixtures (5 files)

Standard test images with complete EXIF data:

- `basic/test_001.jpg` - Canon EOS 5D, 2024-01-15 12:30:45.123456
- `basic/test_002.jpg` - Nikon D850, 2024-01-15 14:45:30.654321
- `basic/test_003.jpg` - Sony A7 III, 2024-02-20 09:15:22.111111
- `basic/test_004.jpg` - Fujifilm X-T4, 2024-03-10 18:20:10.999999
- `basic/test_005.jpg` - Olympus OM-D E-M1, 2024-12-31 23:59:59.000000

### Mixed Format Fixtures (2 files)

Different file types:

- `mixed/photo_001.jpg` - Canon EOS R5 (JPEG with EXIF)
- `mixed/screenshot_001.png` - PNG without EXIF

### No EXIF Fixtures (2 files)

Files without EXIF data to test fallback mechanisms:

- `no_exif/no_metadata.jpg` - No EXIF, will fall back to ctime
- `no_exif/20240615-143022.123456_test.jpg` - Datetime in filename pattern

### Special Make Cases (3 files)

Test special manufacturer name handling:

- `special_makes/htc_001.jpg` - "HTC Corporation" → "HTC"
- `special_makes/lg_001.jpg` - "LG Electronics" → "LG"
- `special_makes/rim_001.jpg` - "Research In Motion" → filtered out

### Collision Test Fixtures (3 files)

Test duplicate detection and collision resolution:

- `collision/photo_a.jpg` - Red image, Canon EOS 5D
- `collision/photo_b.jpg` - Blue image, Canon EOS 5D (same datetime, different content)
- `collision/photo_a_duplicate.jpg` - Exact copy of photo_a.jpg

**SHA256 Hashes**:
- `photo_a.jpg`: b8bf7255f4bd6438e1e72ff2375eeaab5167cad357f1d8e6ac920294a990775b
- `photo_a_duplicate.jpg`: b8bf7255f4bd6438e1e72ff2375eeaab5167cad357f1d8e6ac920294a990775b (identical)
- `photo_b.jpg`: 324be392d0a64e4f9f9159a614d4ed090d61c55befbab8b33c0c1fd63b614bf9 (different)

## Manifest

See `manifest.json` for complete metadata and expected output paths for each fixture.

## Regenerating Fixtures

If you need to regenerate the fixtures:

```bash
cd tests/integration/fixtures
rm -rf basic mixed no_exif special_makes collision
poetry run python generate_fixtures.py
```

Requirements:
- Pillow (installed as dev dependency)
- exiftool (system binary at /usr/local/bin/exiftool)

## Verifying EXIF Data

Check EXIF data for any fixture:

```bash
exiftool -json basic/test_001.jpg
exiftool basic/test_001.jpg | grep -E "(Make|Model|Date)"
```

## Using in Tests

### Python Tests

```python
from pathlib import Path

FIXTURES_DIR = Path(__file__).parent / "integration" / "fixtures"

def test_basic_processing():
    source = FIXTURES_DIR / "basic"
    # Run sortpics on fixtures
    # Verify output
```

### Integration Tests

The `test_scenarios.sh` script automatically uses these fixtures:

```bash
cd tests/integration
./test_scenarios.sh python
```

## Expected Behavior

When processing these fixtures with sortpics:

1. **Basic fixtures** → Organized into YYYY/MM/YYYY-MM-DD directories
2. **Mixed fixtures** → PNG and JPEG both processed
3. **No EXIF fixtures** → Fall back to filename pattern or ctime
4. **Special makes** → Manufacturer names normalized correctly
5. **Collision fixtures**:
   - photo_a.jpg → `20240601-100000.123456_Canon-Eos5d.jpg`
   - photo_b.jpg → `20240601-100000.123456_Canon-Eos5d_1.jpg` (incremented)
   - photo_a_duplicate.jpg → Skipped (duplicate detected)

## Notes

- All images are 640x480 solid color rectangles (very small for fast testing)
- EXIF subsecond precision is set correctly (6 digits)
- Special characters are tested (spaces in model names)
- Collision scenario includes both duplicate (same hash) and collision (different hash)
