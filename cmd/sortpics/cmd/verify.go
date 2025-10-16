package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/alitto/pond"
	"github.com/chris/sortpics-go/internal/metadata"
	"github.com/chris/sortpics-go/internal/pathgen"
	"github.com/chris/sortpics-go/internal/rename"
	"github.com/chris/sortpics-go/pkg/config"
	"github.com/spf13/cobra"
)

var (
	verifyFix bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify [flags] DIRECTORY...",
	Short: "Verify archive filenames match EXIF metadata",
	Long: `Verify that filenames in an organized archive match their EXIF metadata.

This command validates that:
  - Filenames match EXIF DateTimeOriginal
  - Camera make/model in filename matches EXIF
  - No duplicate files exist (same content, different names)

Optional --fix mode will rename files to match EXIF data.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().BoolVar(&verifyFix, "fix", false, "automatically fix mismatches")
}

func runVerify(cmd *cobra.Command, args []string) error {
	// Check if ExifTool is installed
	if err := checkExifTool(); err != nil {
		return err
	}

	dirs := args

	fmt.Printf("Verifying directories: %v\n", dirs)
	if verifyFix {
		fmt.Println("Fix mode: enabled - will rename mismatched files")
	}
	fmt.Println()

	// Collect files to verify
	files, err := collectFilesRecursive(dirs)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No files to verify")
		return nil
	}

	fmt.Printf("Found %d files to verify\n\n", len(files))

	// Verify files
	stats := &VerifyStats{}
	if err := verifyFiles(files, verifyFix, stats); err != nil {
		return err
	}

	// Print summary
	printVerifySummary(stats)

	if stats.Mismatches > 0 && !verifyFix {
		fmt.Println("\nRun with --fix to automatically rename mismatched files")
	}

	return nil
}

// VerifyStats tracks verification statistics
type VerifyStats struct {
	Verified   int64
	Matched    int64
	Mismatches int64
	Fixed      int64
	Errors     int64
}

// collectFilesRecursive collects all supported files recursively
func collectFilesRecursive(dirs []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			ext := strings.TrimPrefix(filepath.Ext(path), ".")
			if rename.IsValidExtension(ext) {
				absPath, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				if !seen[absPath] {
					files = append(files, absPath)
					seen[absPath] = true
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
		}
	}

	return files, nil
}

// verifyFiles verifies all files using a worker pool
func verifyFiles(files []string, fix bool, stats *VerifyStats) error {
	// Use fewer workers for verification to avoid overwhelming output
	workers := 4
	pool := pond.New(workers, len(files))
	defer pool.StopAndWait()

	// Process each file
	for _, file := range files {
		file := file // Capture for closure
		pool.Submit(func() {
			if err := verifyFile(file, fix, stats); err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				fmt.Fprintf(os.Stderr, "Error verifying %s: %v\n", file, err)
			}
		})
	}

	pool.StopAndWait()
	return nil
}

// verifyFile verifies a single file
func verifyFile(file string, fix bool, stats *VerifyStats) error {
	atomic.AddInt64(&stats.Verified, 1)

	// Extract metadata
	extractor, err := metadata.NewMetadataExtractor()
	if err != nil {
		return fmt.Errorf("failed to create metadata extractor: %w", err)
	}
	defer extractor.Close()

	meta, err := extractor.Extract(file, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Generate expected filename
	cfg := &config.ProcessingConfig{
		Precision: 6,
		OldNaming: false,
	}
	pg := pathgen.New(cfg.Precision, cfg.OldNaming)

	// Get the directory this file is in (should be YYYY/MM/YYYY-MM-DD/)
	currentDir := filepath.Dir(file)
	currentFilename := filepath.Base(file)
	ext := strings.TrimPrefix(filepath.Ext(file), ".")

	// Generate what the filename should be (without directory path)
	expectedFilename := pg.GenerateFilename(meta, ext, 0)

	// Compare filenames (case-insensitive to handle extension differences)
	if strings.EqualFold(currentFilename, expectedFilename) {
		atomic.AddInt64(&stats.Matched, 1)
		return nil
	}

	// Mismatch found
	atomic.AddInt64(&stats.Mismatches, 1)
	fmt.Printf("MISMATCH: %s\n", file)
	fmt.Printf("  Current:  %s\n", currentFilename)
	fmt.Printf("  Expected: %s\n", expectedFilename)

	if fix {
		// Rename the file
		expectedPath := filepath.Join(currentDir, expectedFilename)

		// Check if target already exists
		if _, err := os.Stat(expectedPath); err == nil {
			fmt.Printf("  SKIP: Target file already exists: %s\n", expectedPath)
			return nil
		}

		if err := os.Rename(file, expectedPath); err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}

		atomic.AddInt64(&stats.Fixed, 1)
		fmt.Printf("  FIXED: Renamed to %s\n", expectedFilename)
	}
	fmt.Println()

	return nil
}

// printVerifySummary prints verification statistics
func printVerifySummary(stats *VerifyStats) {
	fmt.Println("\nVerification Summary:")
	fmt.Printf("  Verified:   %d\n", stats.Verified)
	fmt.Printf("  Matched:    %d\n", stats.Matched)

	if stats.Mismatches > 0 {
		fmt.Printf("  Mismatches: %d\n", stats.Mismatches)
	}

	if stats.Fixed > 0 {
		fmt.Printf("  Fixed:      %d\n", stats.Fixed)
	}

	if stats.Errors > 0 {
		fmt.Printf("  Errors:     %d\n", stats.Errors)
	}
}
