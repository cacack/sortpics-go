package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory with organized files
	tmpDir, err := os.MkdirTemp("", "sortpics-verify-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// First, organize some files
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")
	destDir := filepath.Join(tmpDir, "archive")

	// Use the copy command to organize files
	copyMode = true
	moveMode = false
	dryRun = false
	recursive = false
	verbose = 0
	numWorkers = 2
	precision = 6
	oldNaming = false
	rawPath = ""
	album = ""
	albumFromDir = false
	tags = []string{}
	timeAdjust = ""
	dayAdjust = 0
	clean = false

	err = run(nil, []string{testDataDir, destDir})
	require.NoError(t, err)

	t.Run("verify matching archive", func(t *testing.T) {
		// Verify the archive - all files should match
		stats := &VerifyStats{}
		files, err := collectFilesRecursive([]string{destDir})
		require.NoError(t, err)
		assert.NotEmpty(t, files, "should have files to verify")

		err = verifyFiles(files, false, stats)
		require.NoError(t, err)

		assert.Equal(t, int64(5), stats.Verified, "should verify 5 files")
		assert.Equal(t, int64(5), stats.Matched, "all files should match")
		assert.Equal(t, int64(0), stats.Mismatches, "should have no mismatches")
		assert.Equal(t, int64(0), stats.Errors, "should have no errors")
	})

	t.Run("verify with mismatch", func(t *testing.T) {
		// Rename a file to cause a mismatch
		files, err := collectFilesRecursive([]string{destDir})
		require.NoError(t, err)
		require.NotEmpty(t, files)

		// Rename the first file
		originalFile := files[0]
		dir := filepath.Dir(originalFile)
		wrongFile := filepath.Join(dir, "wrong_name.jpg")
		err = os.Rename(originalFile, wrongFile)
		require.NoError(t, err)

		// Verify - should find the mismatch
		stats := &VerifyStats{}
		files, err = collectFilesRecursive([]string{destDir})
		require.NoError(t, err)

		err = verifyFiles(files, false, stats)
		require.NoError(t, err)

		assert.Equal(t, int64(5), stats.Verified)
		assert.Equal(t, int64(4), stats.Matched)
		assert.Equal(t, int64(1), stats.Mismatches)

		// Restore original name for next test
		err = os.Rename(wrongFile, originalFile)
		require.NoError(t, err)
	})

	t.Run("verify with fix mode", func(t *testing.T) {
		// Rename a file to cause a mismatch
		files, err := collectFilesRecursive([]string{destDir})
		require.NoError(t, err)
		require.NotEmpty(t, files)

		// Rename the first file
		originalFile := files[0]
		dir := filepath.Dir(originalFile)
		wrongFile := filepath.Join(dir, "mismatch_test.jpg")
		err = os.Rename(originalFile, wrongFile)
		require.NoError(t, err)

		// Verify with fix mode - should rename it back
		stats := &VerifyStats{}
		files, err = collectFilesRecursive([]string{destDir})
		require.NoError(t, err)

		err = verifyFiles(files, true, stats)
		require.NoError(t, err)

		assert.Equal(t, int64(1), stats.Mismatches)
		assert.Equal(t, int64(1), stats.Fixed)

		// Check that the file was renamed back to original name
		_, err = os.Stat(originalFile)
		assert.NoError(t, err, "file should be renamed back to original")

		_, err = os.Stat(wrongFile)
		assert.True(t, os.IsNotExist(err), "wrong file should no longer exist")
	})
}

func TestCollectFilesRecursive(t *testing.T) {
	testDataRoot := filepath.Join("..", "..", "..", "test", "testdata")

	t.Run("collect from basic directory", func(t *testing.T) {
		basicDir := filepath.Join(testDataRoot, "basic")
		files, err := collectFilesRecursive([]string{basicDir})
		require.NoError(t, err)
		assert.NotEmpty(t, files)

		// All files should be from basic directory
		for _, file := range files {
			assert.Contains(t, file, "basic")
		}
	})

	t.Run("collect from multiple directories", func(t *testing.T) {
		basicDir := filepath.Join(testDataRoot, "basic")
		rawDir := filepath.Join(testDataRoot, "raw")

		files, err := collectFilesRecursive([]string{basicDir, rawDir})
		require.NoError(t, err)
		assert.NotEmpty(t, files)

		// Should have files from both directories
		hasBasic := false
		hasRaw := false
		for _, file := range files {
			if filepath.Base(filepath.Dir(file)) == "basic" {
				hasBasic = true
			}
			if filepath.Base(filepath.Dir(file)) == "raw" {
				hasRaw = true
			}
		}
		assert.True(t, hasBasic, "should have files from basic")
		assert.True(t, hasRaw, "should have files from raw")
	})

	t.Run("invalid directory", func(t *testing.T) {
		_, err := collectFilesRecursive([]string{"/nonexistent/directory"})
		assert.Error(t, err)
	})
}

func TestVerifyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create a properly organized file
	tmpDir, err := os.MkdirTemp("", "sortpics-verifyfile-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Copy and organize a test file
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")
	destDir := filepath.Join(tmpDir, "archive")

	copyMode = true
	moveMode = false
	dryRun = false
	recursive = false
	verbose = 0
	numWorkers = 2
	precision = 6
	oldNaming = false
	rawPath = ""
	album = ""
	albumFromDir = false
	tags = []string{}
	timeAdjust = ""
	dayAdjust = 0
	clean = false

	err = run(nil, []string{testDataDir, destDir})
	require.NoError(t, err)

	files, err := collectFilesRecursive([]string{destDir})
	require.NoError(t, err)
	require.NotEmpty(t, files)

	t.Run("verify matching file", func(t *testing.T) {
		stats := &VerifyStats{}
		err := verifyFile(files[0], false, stats)
		require.NoError(t, err)

		assert.Equal(t, int64(1), stats.Verified)
		assert.Equal(t, int64(1), stats.Matched)
		assert.Equal(t, int64(0), stats.Mismatches)
	})

	t.Run("verify mismatched file", func(t *testing.T) {
		// Rename to cause mismatch
		wrongName := filepath.Join(filepath.Dir(files[0]), "wrong.jpg")
		err := os.Rename(files[0], wrongName)
		require.NoError(t, err)
		defer os.Rename(wrongName, files[0])

		stats := &VerifyStats{}
		err = verifyFile(wrongName, false, stats)
		require.NoError(t, err)

		assert.Equal(t, int64(1), stats.Verified)
		assert.Equal(t, int64(0), stats.Matched)
		assert.Equal(t, int64(1), stats.Mismatches)
	})
}
