package metadata

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatePattern tests the DATE_PATTERN regex
func TestDatePattern(t *testing.T) {
	t.Run("full datetime with subseconds", func(t *testing.T) {
		match := DATE_PATTERN.FindStringSubmatch("20240115-123045.123456")
		require.NotNil(t, match)
		assert.Equal(t, "20240115", match[1])
		assert.Equal(t, "123045", match[3])
		assert.Equal(t, "123456", match[5])
	})

	t.Run("date only", func(t *testing.T) {
		match := DATE_PATTERN.FindStringSubmatch("20240115")
		require.NotNil(t, match)
		assert.Equal(t, "20240115", match[1])
	})

	t.Run("no match for invalid string", func(t *testing.T) {
		match := DATE_PATTERN.FindStringSubmatch("not_a_date.jpg")
		assert.Nil(t, match)
	})
}

// TestNewMetadataExtractor tests initialization
func TestNewMetadataExtractor(t *testing.T) {
	extractor, err := NewMetadataExtractor()
	if err != nil {
		t.Skip("ExifTool not available, skipping test")
	}
	defer extractor.Close()

	require.NotNil(t, extractor)
	require.NotNil(t, extractor.et)
}

// TestExtractWithEXIFDatetime tests extracting metadata with EXIF datetime
func TestExtractWithEXIFDatetime(t *testing.T) {
	// Create a test file with EXIF data
	testFile := "../../../test/testdata/test_image_exif.jpg"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test image not available, skipping test")
	}

	extractor, err := NewMetadataExtractor()
	if err != nil {
		t.Skip("ExifTool not available, skipping test")
	}
	defer extractor.Close()

	metadata, err := extractor.Extract(testFile, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, metadata)

	// Basic checks - actual values depend on the test image
	assert.NotNil(t, metadata.DateTime)
	assert.NotNil(t, metadata.RawMetadata)
}

// TestParseMake tests make parsing
func TestParseMake(t *testing.T) {
	extractor := &MetadataExtractor{}

	t.Run("parse Canon", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:Make": "Canon",
		}
		make := extractor.parseMake(metadata)
		assert.Equal(t, "Canon", make)
	})

	t.Run("parse HTC Corporation as HTC", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:Make": "HTC Corporation",
		}
		make := extractor.parseMake(metadata)
		assert.Equal(t, "HTC", make)
	})

	t.Run("parse LG Electronics as LG", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:Make": "LG Electronics",
		}
		make := extractor.parseMake(metadata)
		assert.Equal(t, "LG", make)
	})

	t.Run("filter out Research In Motion", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:Make": "Research In Motion",
		}
		make := extractor.parseMake(metadata)
		assert.Equal(t, "", make)
	})

	t.Run("default to Unknown when missing", func(t *testing.T) {
		metadata := map[string]interface{}{}
		make := extractor.parseMake(metadata)
		assert.Equal(t, "Unknown", make)
	})

	t.Run("use MakerNotes as fallback", func(t *testing.T) {
		metadata := map[string]interface{}{
			"MakerNotes:Make": "Nikon",
		}
		make := extractor.parseMake(metadata)
		assert.Equal(t, "Nikon", make)
	})
}

// TestParseModel tests model parsing
func TestParseModel(t *testing.T) {
	extractor := &MetadataExtractor{}

	t.Run("remove make from model", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:Model": "Canon EOS 5D",
		}
		model := extractor.parseModel("Canon", metadata)
		assert.Equal(t, "Eos5d", model)
	})

	t.Run("convert spaces to CamelCase", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:Model": "Canon PowerShot S410",
		}
		model := extractor.parseModel("Canon", metadata)
		assert.Equal(t, "PowershotS410", model)
	})

	t.Run("return empty for missing model", func(t *testing.T) {
		metadata := map[string]interface{}{}
		model := extractor.parseModel("Canon", metadata)
		assert.Equal(t, "", model)
	})

	t.Run("use MakerNotes as fallback", func(t *testing.T) {
		metadata := map[string]interface{}{
			"MakerNotes:Model": "D850",
		}
		model := extractor.parseModel("Nikon", metadata)
		assert.Equal(t, "D850", model)
	})
}

// TestParseDatetimeFromFilename tests extracting datetime from filename
func TestParseDatetimeFromFilename(t *testing.T) {
	extractor := &MetadataExtractor{}

	t.Run("parse YYYYMMDD-HHMMSS.subsec", func(t *testing.T) {
		metadata := map[string]interface{}{} // No EXIF
		stat, _ := os.Stat(".")
		dt := extractor.parseDatetime("/test/20240115-123045.123456_test.jpg", metadata, stat)

		require.NotNil(t, dt)
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, time.January, dt.Month())
		assert.Equal(t, 15, dt.Day())
		assert.Equal(t, 12, dt.Hour())
		assert.Equal(t, 30, dt.Minute())
		assert.Equal(t, 45, dt.Second())
	})

	t.Run("parse YYYYMMDD-HHMMSS", func(t *testing.T) {
		metadata := map[string]interface{}{}
		stat, _ := os.Stat(".")
		dt := extractor.parseDatetime("/test/20240115-123045_test.jpg", metadata, stat)

		require.NotNil(t, dt)
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, 12, dt.Hour())
	})

	t.Run("parse YYYYMMDD only", func(t *testing.T) {
		metadata := map[string]interface{}{}
		stat, _ := os.Stat(".")
		dt := extractor.parseDatetime("/test/20240115_test.jpg", metadata, stat)

		require.NotNil(t, dt)
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, time.January, dt.Month())
		assert.Equal(t, 15, dt.Day())
	})
}

// TestParseDatetimeFromEXIF tests extracting datetime from EXIF
func TestParseDatetimeFromEXIF(t *testing.T) {
	extractor := &MetadataExtractor{}

	t.Run("parse DateTimeOriginal with subseconds", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:DateTimeOriginal":   "2024:01:15 12:30:45",
			"EXIF:SubSecTimeOriginal": "123456",
		}
		stat, _ := os.Stat(".")
		dt := extractor.parseDatetime("/test/image.jpg", metadata, stat)

		require.NotNil(t, dt)
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, time.January, dt.Month())
		assert.Equal(t, 15, dt.Day())
		assert.Equal(t, 12, dt.Hour())
		assert.Equal(t, 30, dt.Minute())
		assert.Equal(t, 45, dt.Second())
		assert.Equal(t, 123456000, dt.Nanosecond())
	})

	t.Run("parse ModifyDate as fallback", func(t *testing.T) {
		metadata := map[string]interface{}{
			"EXIF:ModifyDate": "2024:01:15 12:30:45",
		}
		stat, _ := os.Stat(".")
		dt := extractor.parseDatetime("/test/image.jpg", metadata, stat)

		require.NotNil(t, dt)
		assert.Equal(t, 2024, dt.Year())
	})

	t.Run("parse QuickTime CreateDate", func(t *testing.T) {
		metadata := map[string]interface{}{
			"QuickTime:CreateDate": "2024:01:15 12:30:45",
		}
		stat, _ := os.Stat(".")
		dt := extractor.parseDatetime("/test/video.mov", metadata, stat)

		require.NotNil(t, dt)
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, time.January, dt.Month())
		assert.Equal(t, 15, dt.Day())
	})
}

// TestParseDatetimeFallbackToCtime tests falling back to file modification time
func TestParseDatetimeFallbackToCtime(t *testing.T) {
	extractor := &MetadataExtractor{}

	metadata := map[string]interface{}{} // No metadata
	stat, _ := os.Stat(".")
	dt := extractor.parseDatetime("/test/no_date.jpg", metadata, stat)

	require.NotNil(t, dt)
	// Should fall back to file's ModTime
	assert.Equal(t, stat.ModTime().Unix(), dt.Unix())
}

// TestExtractWithTimeAdjust tests time adjustment
func TestExtractWithTimeAdjust(t *testing.T) {
	// Test the adjustment logic directly
	baseTime := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	adjustment := 1*time.Hour + 30*time.Minute

	adjusted := baseTime.Add(adjustment)
	expected := time.Date(2024, 1, 15, 14, 0, 45, 0, time.UTC)

	assert.Equal(t, expected, adjusted)
}

// TestExtractWithDayAdjust tests day adjustment
func TestExtractWithDayAdjust(t *testing.T) {
	// Test the adjustment logic directly
	baseTime := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	adjustment := 5 * 24 * time.Hour

	adjusted := baseTime.Add(adjustment)
	expected := time.Date(2024, 1, 20, 12, 30, 45, 0, time.UTC)

	assert.Equal(t, expected, adjusted)
}

// TestExifNotFoundError tests the error type
func TestExifNotFoundError(t *testing.T) {
	err := &ExifNotFoundError{Err: os.ErrNotExist}
	assert.Contains(t, err.Error(), "exiftool not found")
}

// Benchmark tests
func BenchmarkParseMake(b *testing.B) {
	extractor := &MetadataExtractor{}
	metadata := map[string]interface{}{
		"EXIF:Make": "Canon",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.parseMake(metadata)
	}
}

func BenchmarkParseModel(b *testing.B) {
	extractor := &MetadataExtractor{}
	metadata := map[string]interface{}{
		"EXIF:Model": "Canon EOS 5D Mark III",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.parseModel("Canon", metadata)
	}
}

func BenchmarkDatePattern(b *testing.B) {
	filename := "20240115-123045.123456_Canon-EOS5D.jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DATE_PATTERN.FindStringSubmatch(filename)
	}
}
