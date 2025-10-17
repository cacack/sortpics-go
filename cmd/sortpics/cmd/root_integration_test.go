package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory for test output
	tmpDir, err := os.MkdirTemp("", "sortpics-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Get path to test data
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")

	t.Run("copy mode", func(t *testing.T) {
		destDir := filepath.Join(tmpDir, "copy-test")

		// Reset flags
		copyMode = true
		moveMode = false
		dryRun = false
		recursive = false
		verbose = 1
		numWorkers = 2
		precision = 6
		oldNaming = false
		rawPath = ""
		album = ""
		albumFromDir = false
		tags = []string{}
		timeAdjust = ""
		dayAdjust = 0

		// Run command
		err := run(nil, []string{testDataDir, destDir})
		require.NoError(t, err)

		// Verify files were copied
		expectedFiles := []string{
			"2024/01/2024-01-15/20240115-123045.123456_Canon-Eos5d.jpg",
			"2024/01/2024-01-15/20240115-144530.654321_Nikon-D850.jpg",
			"2024/02/2024-02-20/20240220-091522.111111_Sony-A7Iii.jpg",
			"2024/03/2024-03-10/20240310-182010.999999_Fujifilm-X-T4.jpg",
			"2024/12/2024-12-31/20241231-235959.000000_Olympus-Om-DE-M1.jpg",
		}

		for _, expectedFile := range expectedFiles {
			fullPath := filepath.Join(destDir, expectedFile)
			assert.FileExists(t, fullPath, "expected file should exist: %s", expectedFile)
		}
	})

	t.Run("dry run mode", func(t *testing.T) {
		destDir := filepath.Join(tmpDir, "dry-run-test")

		// Reset flags
		copyMode = true
		moveMode = false
		dryRun = true
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

		// Run command
		err := run(nil, []string{testDataDir, destDir})
		require.NoError(t, err)

		// Verify no files were created
		_, err = os.Stat(destDir)
		assert.True(t, os.IsNotExist(err), "destination directory should not exist in dry run")
	})

	t.Run("recursive mode", func(t *testing.T) {
		destDir := filepath.Join(tmpDir, "recursive-test")
		testMultiDir := filepath.Join("..", "..", "..", "test", "testdata")

		// Reset flags
		copyMode = true
		moveMode = false
		dryRun = false
		recursive = true
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

		// Run command with testdata directory (should find files in subdirs)
		err := run(nil, []string{testMultiDir, destDir})
		require.NoError(t, err)

		// Just verify some files were processed
		entries, err := os.ReadDir(destDir)
		require.NoError(t, err)
		assert.NotEmpty(t, entries, "destination should have files after recursive copy")
	})

	t.Run("RAW file separation", func(t *testing.T) {
		destDir := filepath.Join(tmpDir, "raw-test")
		rawDestDir := filepath.Join(tmpDir, "raw-test-raw")
		rawTestDir := filepath.Join("..", "..", "..", "test", "testdata", "raw")

		// Reset flags
		copyMode = true
		moveMode = false
		dryRun = false
		recursive = false
		verbose = 0
		numWorkers = 2
		precision = 6
		oldNaming = false
		rawPath = rawDestDir
		album = ""
		albumFromDir = false
		tags = []string{}
		timeAdjust = ""
		dayAdjust = 0

		// Run command
		err := run(nil, []string{rawTestDir, destDir})
		require.NoError(t, err)

		// Verify RAW files went to separate directory
		entries, err := os.ReadDir(rawDestDir)
		require.NoError(t, err)
		assert.NotEmpty(t, entries, "RAW destination should have files")

		// Verify main destination is empty or doesn't exist
		_, err = os.Stat(destDir)
		if err == nil {
			entries, err := os.ReadDir(destDir)
			require.NoError(t, err)
			assert.Empty(t, entries, "main destination should be empty when all files are RAW")
		}
	})
}

func TestIntegrationCleanAfterMove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp source and dest directories
	srcDir, err := os.MkdirTemp("", "sortpics-clean-src-*")
	require.NoError(t, err)
	defer os.RemoveAll(srcDir)

	destDir, err := os.MkdirTemp("", "sortpics-clean-dest-*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	// Create subdirectory with a test file and .DSC file
	subDir := filepath.Join(srcDir, "DCIM")
	require.NoError(t, os.Mkdir(subDir, 0755))

	// Copy a real test file to the source
	testFile := filepath.Join("..", "..", "..", "test", "testdata", "basic", "test_001.jpg")
	srcFile := filepath.Join(subDir, "test_001.jpg")
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(srcFile, data, 0644))

	// Add a .DSC file
	dscFile := filepath.Join(subDir, "NIKON001.DSC")
	require.NoError(t, os.WriteFile(dscFile, []byte{}, 0644))

	// Reset flags for move with clean
	copyMode = false
	moveMode = true
	dryRun = false
	recursive = true
	clean = true
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

	// Run command
	err = run(nil, []string{srcDir, destDir})
	require.NoError(t, err)

	// Verify .DSC file was removed
	assert.NoFileExists(t, dscFile, "DSC file should be removed during cleanup")

	// Verify subdirectory was removed (should be empty after move)
	assert.NoDirExists(t, subDir, "Empty subdirectory should be removed")

	// Verify file was moved to destination
	entries, err := os.ReadDir(destDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "Destination should have the moved file")
}

func TestCleanEmptyDirectoriesNonRecursive(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "sortpics-nonrecursive-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create subdirectory (should NOT be cleaned in non-recursive mode)
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	// Run non-recursive cleanup
	stats := cleanEmptyDirectories([]string{tmpDir}, false, 0)

	// Verify subdirectory still exists (non-recursive doesn't descend)
	assert.DirExists(t, subDir, "Subdirectory should still exist in non-recursive mode")

	// Verify no directories were removed (tmpDir is not empty, has subdir)
	assert.Equal(t, 0, stats.Removed, "Should not remove directories in non-recursive mode when not empty")
}

func TestCameraMetadataFileDetection(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"Nikon DSC file uppercase", "NIKON001.DSC", true},
		{"Nikon DSC file lowercase", "nikon001.dsc", true},
		{"Nikon DSC file mixed case", "Nikon001.Dsc", true},
		{"Regular file", "photo.jpg", false},
		{"Hidden file", ".hidden", false},
		{"DSC in filename but not extension", "DSC_0001.jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCameraMetadataFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanEmptyDirectoriesWithDSC(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "sortpics-clean-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create directory with .DSC file (realistic Nikon filename)
	miscDir := filepath.Join(tmpDir, "MISC")
	require.NoError(t, os.Mkdir(miscDir, 0755))

	dscFile := filepath.Join(miscDir, "NIKON001.DSC")
	require.NoError(t, os.WriteFile(dscFile, []byte{}, 0644))

	// Verify setup
	assert.FileExists(t, dscFile)

	// Run cleanup
	stats := cleanEmptyDirectories([]string{tmpDir}, true, 0)

	// Verify .DSC file was removed
	assert.Equal(t, 1, stats.FilesRemoved, "Should remove 1 camera metadata file")
	assert.NoFileExists(t, dscFile, "NIKON001.DSC file should be removed")

	// Verify directory was also removed since it's now empty
	assert.Equal(t, 2, stats.Removed, "Should remove 2 directories (MISC and tmpDir)")
	assert.NoDirExists(t, miscDir, "MISC directory should be removed")
}

func TestCollectFiles(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")

	t.Run("non-recursive", func(t *testing.T) {
		files, err := collectFiles([]string{testDataDir}, false, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, files)

		// All files should be from the basic directory
		for _, file := range files {
			assert.Contains(t, file, "basic", "file should be from basic directory")
		}
	})

	t.Run("recursive", func(t *testing.T) {
		testDataRoot := filepath.Join("..", "..", "..", "test", "testdata")
		files, err := collectFiles([]string{testDataRoot}, true, 0)
		require.NoError(t, err)
		assert.NotEmpty(t, files)

		// Should find files from multiple subdirectories
		hasBasic := false
		hasRaw := false
		for _, file := range files {
			if filepath.Dir(file) == filepath.Clean(testDataDir) {
				hasBasic = true
			}
			if filepath.Base(filepath.Dir(file)) == "raw" {
				hasRaw = true
			}
		}
		assert.True(t, hasBasic || hasRaw, "should find files from subdirectories")
	})

	t.Run("invalid directory", func(t *testing.T) {
		_, err := collectFiles([]string{"/nonexistent/directory"}, false, 0)
		assert.Error(t, err)
	})
}
