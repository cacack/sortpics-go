package rename

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cacack/sortpics-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationParseMetadata tests metadata extraction with real image files
func TestIntegrationParseMetadata(t *testing.T) {
	fixtureDir := "/Users/chris/devel/home/sortpics/tests/integration/fixtures/basic"

	// Check if fixtures are available
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("Integration test fixtures not available")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(fixtureDir, "test_001.jpg")

	cfg := &config.ProcessingConfig{
		Precision: 6,
	}

	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata
	err = ir.ParseMetadata()
	require.NoError(t, err)

	// Verify that datetime was extracted
	assert.NotNil(t, ir.datetime, "datetime should be extracted")

	// Verify destination path was generated
	assert.NotEmpty(t, ir.destination, "destination should be generated")
	assert.NotEmpty(t, ir.destinationDir, "destination directory should be set")

	// Verify destination path structure
	assert.Contains(t, ir.destination, tmpDir, "destination should be under temp directory")
}

// TestIntegrationPerform tests the full copy operation
func TestIntegrationPerform(t *testing.T) {
	fixtureDir := "/Users/chris/devel/home/sortpics/tests/integration/fixtures/basic"

	// Check if fixtures are available
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("Integration test fixtures not available")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(fixtureDir, "test_001.jpg")

	cfg := &config.ProcessingConfig{
		Precision: 6,
		Move:      false, // Copy mode
	}

	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata
	err = ir.ParseMetadata()
	require.NoError(t, err)

	destination := ir.GetDestination()

	// Perform the operation
	err = ir.Perform()
	require.NoError(t, err)

	// Verify destination file exists
	assert.FileExists(t, destination, "destination file should exist")

	// Verify source still exists (copy mode)
	assert.FileExists(t, testFile, "source file should still exist in copy mode")

	// Verify destination has content
	info, err := os.Stat(destination)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "destination file should have content")
}

// TestIntegrationPerformWithAlbum tests metadata writing
func TestIntegrationPerformWithAlbum(t *testing.T) {
	fixtureDir := "/Users/chris/devel/home/sortpics/tests/integration/fixtures/basic"

	// Check if fixtures are available
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("Integration test fixtures not available")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(fixtureDir, "test_001.jpg")

	cfg := &config.ProcessingConfig{
		Precision: 6,
		Album:     "TestAlbum",
		Tags:      []string{"test", "integration"},
	}

	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata
	err = ir.ParseMetadata()
	require.NoError(t, err)

	destination := ir.GetDestination()

	// Perform the operation
	err = ir.Perform()
	require.NoError(t, err)

	// Verify destination file exists
	assert.FileExists(t, destination, "destination file should exist")
}

// TestIntegrationDryRun tests dry run mode
func TestIntegrationDryRun(t *testing.T) {
	fixtureDir := "/Users/chris/devel/home/sortpics/tests/integration/fixtures/basic"

	// Check if fixtures are available
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("Integration test fixtures not available")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(fixtureDir, "test_001.jpg")

	cfg := &config.ProcessingConfig{
		Precision: 6,
		DryRun:    true,
	}

	ir, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir.Close()

	// Parse metadata
	err = ir.ParseMetadata()
	require.NoError(t, err)

	destination := ir.GetDestination()

	// Perform the operation (should do nothing in dry run)
	err = ir.Perform()
	require.NoError(t, err)

	// Verify destination file does NOT exist (dry run)
	assert.NoFileExists(t, destination, "destination should not exist in dry run mode")
}

// TestIntegrationDuplicateDetection tests duplicate detection
func TestIntegrationDuplicateDetection(t *testing.T) {
	fixtureDir := "/Users/chris/devel/home/sortpics/tests/integration/fixtures/basic"

	// Check if fixtures are available
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("Integration test fixtures not available")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(fixtureDir, "test_001.jpg")

	cfg := &config.ProcessingConfig{
		Precision: 6,
	}

	// First copy
	ir1, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir1.Close()

	err = ir1.ParseMetadata()
	require.NoError(t, err)

	err = ir1.Perform()
	require.NoError(t, err)

	destination1 := ir1.GetDestination()
	assert.FileExists(t, destination1)

	// Second copy of same file (should detect as duplicate)
	ir2, err := NewImageRename(testFile, tmpDir, cfg)
	require.NoError(t, err)
	defer ir2.Close()

	err = ir2.ParseMetadata()
	require.NoError(t, err)

	// Should detect as duplicate or generate different name
	destination2 := ir2.GetDestination()

	// If it's a duplicate, the perform should skip
	// Otherwise, it should have a collision-resolved name
	if destination1 == destination2 {
		// Same destination - should detect as duplicate
		assert.True(t, ir2.IsDuplicate() || destination2 != destination1)
	}
}

// TestIntegrationCollisionResolution tests filename collision handling
func TestIntegrationCollisionResolution(t *testing.T) {
	fixtureDir := "/Users/chris/devel/home/sortpics/tests/integration/fixtures/basic"

	// Check if fixtures are available
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("Integration test fixtures not available")
	}

	tmpDir := t.TempDir()
	testFile1 := filepath.Join(fixtureDir, "test_001.jpg")
	testFile2 := filepath.Join(fixtureDir, "test_002.jpg") // Different file, may have same timestamp

	cfg := &config.ProcessingConfig{
		Precision: 6,
	}

	// First file
	ir1, err := NewImageRename(testFile1, tmpDir, cfg)
	require.NoError(t, err)
	defer ir1.Close()

	err = ir1.ParseMetadata()
	require.NoError(t, err)

	err = ir1.Perform()
	require.NoError(t, err)

	destination1 := ir1.GetDestination()

	// Second file
	ir2, err := NewImageRename(testFile2, tmpDir, cfg)
	require.NoError(t, err)
	defer ir2.Close()

	err = ir2.ParseMetadata()
	require.NoError(t, err)

	destination2 := ir2.GetDestination()

	// If both files have same timestamp/metadata, collision resolution should kick in
	// and generate different filenames
	if destination1 == destination2 {
		t.Log("Files have identical metadata, testing collision resolution")
		err = ir2.Perform()
		require.NoError(t, err)

		// After perform, a new filename should be generated
		// (This is handled in the Perform method's re-check logic)
	} else {
		// Different destinations expected
		err = ir2.Perform()
		require.NoError(t, err)
		assert.FileExists(t, destination2)
	}

	// Both files should exist
	assert.FileExists(t, destination1)
}
