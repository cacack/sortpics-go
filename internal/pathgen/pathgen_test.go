package pathgen

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/chris/sortpics-go/pkg/config"
	"github.com/stretchr/testify/assert"
)

// TestNew tests PathGenerator initialization
func TestNew(t *testing.T) {
	generator := New(6, false)
	assert.Equal(t, 6, generator.Precision)
	assert.False(t, generator.OldNaming)
}

// TestNewWithOldNaming tests PathGenerator initialization with old naming
func TestNewWithOldNaming(t *testing.T) {
	generator := New(2, true)
	assert.Equal(t, 2, generator.Precision)
	assert.True(t, generator.OldNaming)
}

// TestGenerateDirectory tests directory generation with datetime
func TestGenerateDirectory(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)
	baseDir := "/archive"

	directory := generator.GenerateDirectory(metadata, baseDir)

	expected := filepath.Join("/archive", "2024", "01", "2024-01-15")
	assert.Equal(t, expected, directory)
}

// TestGenerateDirectoryNoDatetime tests directory generation without datetime
func TestGenerateDirectoryNoDatetime(t *testing.T) {
	metadata := &config.ImageMetadata{
		DateTime: nil,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)
	baseDir := "/archive"

	directory := generator.GenerateDirectory(metadata, baseDir)

	expected := filepath.Join("/archive", "unknown")
	assert.Equal(t, expected, directory)
}

// TestGenerateFilenameWithMakeAndModel tests filename generation with both make and model
func TestGenerateFilenameWithMakeAndModel(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.123456_Canon-EOS5d.jpg", filename)
}

// TestGenerateFilenameOldNaming tests filename generation with old naming convention
func TestGenerateFilenameOldNaming(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(2, true)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.12_CanonEOS5d.jpg", filename)
}

// TestGenerateFilenameWithIncrement tests filename generation with collision increment
func TestGenerateFilenameWithIncrement(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 1)

	assert.Equal(t, "20240115-123045.123456_Canon-EOS5d_1.jpg", filename)
}

// TestGenerateFilenamePrecision2 tests filename generation with 2-digit precision
func TestGenerateFilenamePrecision2(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(2, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.12_Canon-EOS5d.jpg", filename)
}

// TestGenerateFilenameNoSubsec tests filename generation with no microseconds
func TestGenerateFilenameNoSubsec(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.000000_Canon-EOS5d.jpg", filename)
}

// TestGenerateFilenameOnlyMake tests filename generation with only make
func TestGenerateFilenameOnlyMake(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.123456_Canon.jpg", filename)
}

// TestGenerateFilenameOnlyModel tests filename generation with only model
func TestGenerateFilenameOnlyModel(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "",
		Model:    "EOS5d",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.123456_EOS5d.jpg", filename)
}

// TestGenerateFilenameNoMakeOrModel tests filename generation with neither make nor model
func TestGenerateFilenameNoMakeOrModel(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "",
		Model:    "",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "20240115-123045.123456_Unknown.jpg", filename)
}

// TestGenerateFilenameNoDatetime tests filename generation without datetime
func TestGenerateFilenameNoDatetime(t *testing.T) {
	metadata := &config.ImageMetadata{
		DateTime: nil,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	assert.Equal(t, "unknown_Canon-EOS5d.jpg", filename)
}

// TestGenerateFilenameLowercaseExtension tests that extension is converted to lowercase
func TestGenerateFilenameLowercaseExtension(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)

	filename := generator.GenerateFilename(metadata, "JPG", 0)

	assert.True(t, filepath.Ext(filename) == ".jpg")
}

// TestGeneratePathFull tests full path generation
func TestGeneratePathFull(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)
	baseDir := "/archive"

	path := generator.GeneratePath(metadata, baseDir, "jpg", 0)

	expected := filepath.Join("/archive", "2024", "01", "2024-01-15", "20240115-123045.123456_Canon-EOS5d.jpg")
	assert.Equal(t, expected, path)
}

// TestGeneratePathWithIncrement tests full path generation with increment
func TestGeneratePathWithIncrement(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)
	baseDir := "/archive"

	path := generator.GeneratePath(metadata, baseDir, "jpg", 2)

	expected := filepath.Join("/archive", "2024", "01", "2024-01-15", "20240115-123045.123456_Canon-EOS5d_2.jpg")
	assert.Equal(t, expected, path)
}

// TestGeneratePathUnknownDirectory tests path generation when datetime is missing
func TestGeneratePathUnknownDirectory(t *testing.T) {
	metadata := &config.ImageMetadata{
		DateTime: nil,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(6, false)
	baseDir := "/archive"

	path := generator.GeneratePath(metadata, baseDir, "jpg", 0)

	expected := filepath.Join("/archive", "unknown", "unknown_Canon-EOS5d.jpg")
	assert.Equal(t, expected, path)
}

// TestGeneratePathOldNaming tests full path generation with old naming convention
func TestGeneratePathOldNaming(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 12000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	generator := New(2, true)
	baseDir := "/archive"

	path := generator.GeneratePath(metadata, baseDir, "jpg", 0)

	expected := filepath.Join("/archive", "2024", "01", "2024-01-15", "20240115-123045.00_CanonEOS5d.jpg")
	assert.Equal(t, expected, path)
}

// TestGenerateFilenamePrecisionGreaterThan6 tests filename generation with precision > 6
func TestGenerateFilenamePrecisionGreaterThan6(t *testing.T) {
	dt := time.Date(2024, 1, 15, 12, 30, 45, 123456000, time.UTC)
	metadata := &config.ImageMetadata{
		DateTime: &dt,
		Make:     "Canon",
		Model:    "EOS5d",
	}
	// Use precision of 8 (greater than max 6)
	generator := New(8, false)

	filename := generator.GenerateFilename(metadata, "jpg", 0)

	// Should return full 6-digit subsecond precision (maximum available)
	assert.Equal(t, "20240115-123045.123456_Canon-EOS5d.jpg", filename)
}
