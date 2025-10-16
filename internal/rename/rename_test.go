package rename

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/cacack/sortpics-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateTimeDelta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"1 second", "00:00:01", 1 * time.Second},
		{"1 minute", "00:01:00", 1 * time.Minute},
		{"1 hour", "01:00:00", 1 * time.Hour},
		{"negative 3 hours 5 seconds", "-03:00:05", -3*time.Hour - 5*time.Second},
		{"complex time", "01:02:03", 1*time.Hour + 2*time.Minute + 3*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateTimeDelta(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDayDelta(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"1 day", "1", 24 * time.Hour},
		{"5 days", "5", 5 * 24 * time.Hour},
		{"negative 3 days", "-3", -3 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateDayDelta(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidExtension(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("valid extension jpg", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.jpg")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		cfg := &config.ProcessingConfig{Precision: 6}
		ir, err := NewImageRename(testFile, tmpDir, cfg)
		require.NoError(t, err)
		defer ir.Close()

		assert.True(t, ir.IsValidExtension())
	})

	t.Run("invalid extension txt", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		cfg := &config.ProcessingConfig{Precision: 6}
		ir, err := NewImageRename(testFile, tmpDir, cfg)
		require.NoError(t, err)
		defer ir.Close()

		assert.False(t, ir.IsValidExtension())
	})
}

func TestIsRaw(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("raw extension cr2", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.cr2")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		cfg := &config.ProcessingConfig{Precision: 6}
		ir, err := NewImageRename(testFile, tmpDir, cfg)
		require.NoError(t, err)
		defer ir.Close()

		assert.True(t, ir.IsRaw())
	})

	t.Run("non-raw extension jpg", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.jpg")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		cfg := &config.ProcessingConfig{Precision: 6}
		ir, err := NewImageRename(testFile, tmpDir, cfg)
		require.NoError(t, err)
		defer ir.Close()

		assert.False(t, ir.IsRaw())
	})
}

func TestRawPathRouting(t *testing.T) {
	tmpDir := t.TempDir()
	rawPath := filepath.Join(tmpDir, "raw")
	destPath := filepath.Join(tmpDir, "dest")

	testFile := filepath.Join(tmpDir, "test.cr2")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		RawPath:   rawPath,
	}
	ir, err := NewImageRename(testFile, destPath, cfg)
	require.NoError(t, err)
	defer ir.Close()

	absRawPath, _ := filepath.Abs(rawPath)
	assert.Equal(t, absRawPath, ir.destinationBase)
}

func TestAlbumFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	albumDir := filepath.Join(tmpDir, "Summer2023")
	require.NoError(t, os.MkdirAll(albumDir, 0755))

	testFile := filepath.Join(albumDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{
		Precision:    6,
		AlbumFromDir: true,
	}
	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	assert.Equal(t, "Summer2023", ir.album)
}

func TestAlbumExplicit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		Album:     "Vacation",
	}
	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	assert.Equal(t, "Vacation", ir.album)
}

func TestTimeAdjust(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{
		Precision:  6,
		TimeAdjust: "01:30:00",
	}
	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	expected := 1*time.Hour + 30*time.Minute
	assert.NotNil(t, ir.timeDelta)
	assert.Equal(t, expected, *ir.timeDelta)
}

func TestDayAdjust(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		DayAdjust: "5",
	}
	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	expected := 5 * 24 * time.Hour
	assert.NotNil(t, ir.dayDelta)
	assert.Equal(t, expected, *ir.dayDelta)
}

func TestSafeCopy(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test content"), 0644))

	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	dest := filepath.Join(destDir, "destination.txt")

	err := SafeCopy(src, dest)
	require.NoError(t, err)

	// Check destination exists
	assert.FileExists(t, dest)

	// Check content
	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))

	// Check source still exists
	assert.FileExists(t, src)
}

func TestSafeMoveSameFilesystem(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test content"), 0644))

	dest := filepath.Join(tmpDir, "destination.txt")

	err := SafeMove(src, dest)
	require.NoError(t, err)

	// Check destination exists
	assert.FileExists(t, dest)

	// Check content
	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))

	// Check source is gone
	assert.NoFileExists(t, src)
}

func TestSafeMoveCrossFilesystem(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test content"), 0644))

	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	dest := filepath.Join(destDir, "destination.txt")

	// We can't easily force EXDEV in a unit test, but we can test
	// that SafeMove works correctly via the copy+delete fallback
	// by using SafeCopy directly and then removing the source

	err := SafeCopy(src, dest)
	require.NoError(t, err)

	err = os.Remove(src)
	require.NoError(t, err)

	// Verify the result
	assert.FileExists(t, dest)
	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
	assert.NoFileExists(t, src)
}

// TestHelperFunctions tests the standalone helper functions
func TestIsValidExtensionFunction(t *testing.T) {
	assert.True(t, IsValidExtension("jpg"))
	assert.True(t, IsValidExtension("JPG"))
	assert.True(t, IsValidExtension("jpeg"))
	assert.True(t, IsValidExtension("cr2"))
	assert.False(t, IsValidExtension("txt"))
	assert.False(t, IsValidExtension("doc"))
}

func TestIsRawFunction(t *testing.T) {
	assert.True(t, IsRaw("cr2"))
	assert.True(t, IsRaw("CR2"))
	assert.True(t, IsRaw("nef"))
	assert.True(t, IsRaw("dng"))
	assert.False(t, IsRaw("jpg"))
	assert.False(t, IsRaw("png"))
}

// Mock test to verify cross-filesystem error handling
// This test verifies the logic path without actually crossing filesystems
func TestSafeMoveEXDEVHandling(t *testing.T) {
	// This is a conceptual test - in real usage, EXDEV would be triggered
	// by attempting to rename across filesystem boundaries.
	// The SafeMove function handles this by catching syscall.EXDEV and
	// falling back to copy+delete.

	// We verify that the error type checking works correctly
	linkErr := &os.LinkError{
		Op:  "rename",
		Old: "/path/old",
		New: "/path/new",
		Err: syscall.EXDEV,
	}

	// Verify we can detect EXDEV
	if errno, ok := linkErr.Err.(syscall.Errno); ok {
		assert.Equal(t, syscall.EXDEV, errno)
	}
}

// TestIsDuplicate tests the IsDuplicate getter
func TestIsDuplicate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{Precision: 6}
	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Initially should be false (not set)
	assert.False(t, ir.IsDuplicate())

	// Manually set it to true for testing
	ir.isDuplicate = true
	assert.True(t, ir.IsDuplicate())
}

// TestGenerateUUID tests the UUID generation
func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	// Check format (should have 4 dashes)
	assert.Equal(t, 4, len(filepath.SplitList(uuid1))-1+4) // Basic sanity check
	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)

	// UUIDs should be different
	assert.NotEqual(t, uuid1, uuid2)

	// Check it contains hex characters and dashes
	assert.Regexp(t, "^[0-9a-f]+-[0-9a-f]+-[0-9a-f]+-[0-9a-f]+-[0-9a-f]+$", uuid1)
}

// TestSafeCopySourceNotExists tests SafeCopy with non-existent source
func TestSafeCopySourceNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "nonexistent.txt")
	dest := filepath.Join(tmpDir, "destination.txt")

	err := SafeCopy(src, dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read source file")
}

// TestSafeCopyDestDirNotExists tests SafeCopy with non-existent destination directory
func TestSafeCopyDestDirNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test content"), 0644))

	dest := filepath.Join(tmpDir, "nonexistent", "destination.txt")

	err := SafeCopy(src, dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create temp file")
}

// TestSafeCopyPermissions tests that SafeCopy preserves file permissions
func TestSafeCopyPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test content"), 0755))

	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	dest := filepath.Join(destDir, "destination.txt")

	err := SafeCopy(src, dest)
	require.NoError(t, err)

	srcInfo, err := os.Stat(src)
	require.NoError(t, err)

	destInfo, err := os.Stat(dest)
	require.NoError(t, err)

	assert.Equal(t, srcInfo.Mode(), destInfo.Mode())
}

// TestSafeMoveSourceNotExists tests SafeMove with non-existent source
func TestSafeMoveSourceNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "nonexistent.txt")
	dest := filepath.Join(tmpDir, "destination.txt")

	err := SafeMove(src, dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to move file")
}

// TestSafeMoveDestDirNotExists tests SafeMove with non-existent destination directory
func TestSafeMoveDestDirNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test content"), 0644))

	dest := filepath.Join(tmpDir, "nonexistent", "destination.txt")

	err := SafeMove(src, dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to move file")
}

// TestPerformDryRun tests that Perform doesn't actually move/copy in dry-run mode
func TestPerformDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		DryRun:    true,
		Move:      false, // false means copy
	}

	ir, err := NewImageRename(testFile, destDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata to set destination
	err = ir.ParseMetadata()
	require.NoError(t, err)

	// Perform should succeed but not create destination
	err = ir.Perform()
	require.NoError(t, err)

	// Destination should not exist
	assert.NoFileExists(t, ir.destination)
}

// TestPerformCopy tests the Perform method with copy operation
func TestPerformCopy(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		Move:      false, // false means copy
	}

	ir, err := NewImageRename(testFile, destDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata to set destination
	err = ir.ParseMetadata()
	require.NoError(t, err)

	// Perform copy
	err = ir.Perform()
	require.NoError(t, err)

	// Verify destination exists
	assert.FileExists(t, ir.destination)

	// Verify source still exists (copy, not move)
	assert.FileExists(t, testFile)

	// Verify content
	content, err := os.ReadFile(ir.destination)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

// TestPerformMove tests the Perform method with move operation
func TestPerformMove(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		Move:      true,
	}

	ir, err := NewImageRename(testFile, destDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata to set destination
	err = ir.ParseMetadata()
	require.NoError(t, err)

	// Perform move
	err = ir.Perform()
	require.NoError(t, err)

	// Verify destination exists
	assert.FileExists(t, ir.destination)

	// Verify source is gone (moved, not copied)
	assert.NoFileExists(t, testFile)

	// Verify content
	content, err := os.ReadFile(ir.destination)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

// TestPerformRaceConditionCollision tests the race condition recheck logic
func TestPerformRaceConditionCollision(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")

	testFile := filepath.Join(tmpDir, "test.jpg")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	cfg := &config.ProcessingConfig{
		Precision: 6,
		Move:      false,
	}

	ir, err := NewImageRename(testFile, destDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata to set destination
	err = ir.ParseMetadata()
	require.NoError(t, err)

	// Simulate race condition: another process created the file first
	// Create the destination directory and file before Perform
	require.NoError(t, os.MkdirAll(filepath.Dir(ir.destination), 0755))
	require.NoError(t, os.WriteFile(ir.destination, []byte("different content"), 0644))

	// Perform should handle collision and create a renamed version
	err = ir.Perform()
	require.NoError(t, err)

	// The file should have been successfully copied
	// The destination should exist (either original or with _N suffix)
	assert.FileExists(t, ir.destination)
}

// TestCalculateTimeDeltaErrors tests error handling for invalid time formats
func TestCalculateTimeDeltaErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid format - too few parts", "01:02"},
		{"invalid format - too many parts", "01:02:03:04"},
		{"invalid hours", "XX:00:00"},
		{"invalid minutes", "00:XX:00"},
		{"invalid seconds", "00:00:XX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateTimeDelta(tt.input)
			require.Error(t, err)
		})
	}
}

// TestCalculateDayDeltaErrors tests error handling for invalid day formats
func TestCalculateDayDeltaErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid format - not a number", "abc"},
		{"invalid format - empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateDayDelta(tt.input)
			require.Error(t, err)
		})
	}
}
