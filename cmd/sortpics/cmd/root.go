package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/alitto/pond"
	"github.com/chris/sortpics-go/internal/rename"
	"github.com/chris/sortpics-go/pkg/config"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	// Operation mode flags
	copyMode  bool
	moveMode  bool
	dryRun    bool
	recursive bool
	clean     bool
	verbose   int

	// Path flags
	rawPath string

	// Naming flags
	precision int
	oldNaming bool

	// Time adjustment flags
	timeAdjust string
	dayAdjust  int

	// Metadata flags
	album        string
	albumFromDir bool
	tags         []string

	// Performance flags
	numWorkers int
)

var rootCmd = &cobra.Command{
	Use:   "sortpics [flags] SOURCE... DESTINATION",
	Short: "Organize photos and videos by EXIF metadata",
	Long: `sortpics organizes photos and videos into a chronological directory structure.

Files are renamed based on EXIF metadata:
  Format: YYYYMMDD-HHMMSS.microsec_Make-Model.ext
  Directory: YYYY/MM/YYYY-MM-DD/

Features:
  - EXIF metadata extraction with fallback hierarchy
  - SHA256-based duplicate detection
  - Atomic file operations
  - Parallel processing
  - RAW file segregation
  - Album and keyword tagging`,
	Version: version,
	Args:    cobra.MinimumNArgs(2),
	RunE:    run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Operation mode flags
	rootCmd.Flags().BoolVarP(&copyMode, "copy", "c", false, "copy files (leave originals)")
	rootCmd.Flags().BoolVarP(&moveMode, "move", "m", false, "move files (remove originals)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview operations without executing")
	rootCmd.Flags().BoolVar(&dryRun, "pretend", false, "alias for --dry-run")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "process subdirectories recursively")
	rootCmd.Flags().BoolVarP(&clean, "clean", "C", false, "remove empty directories after move")
	rootCmd.Flags().CountVarP(&verbose, "verbose", "v", "increase verbosity (-v, -vv, -vvv)")

	// Path flags
	rootCmd.Flags().StringVar(&rawPath, "raw-path", "", "separate path for RAW files")

	// Naming flags
	rootCmd.Flags().IntVarP(&precision, "precision", "p", 6, "subsecond precision (digits)")
	rootCmd.Flags().BoolVar(&oldNaming, "old-naming", false, "use old naming format (no separator)")

	// Time adjustment flags
	rootCmd.Flags().StringVar(&timeAdjust, "time-adjust", "", "adjust time (HH:MM:SS or -HH:MM:SS)")
	rootCmd.Flags().IntVar(&dayAdjust, "day-adjust", 0, "adjust days (positive or negative)")

	// Metadata flags
	rootCmd.Flags().StringVar(&album, "album", "", "set album metadata")
	rootCmd.Flags().BoolVar(&albumFromDir, "album-from-directory", false, "use parent directory as album")
	rootCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "add keyword tags (can be repeated)")

	// Performance flags
	rootCmd.Flags().IntVarP(&numWorkers, "workers", "w", runtime.NumCPU(), "number of worker goroutines")

	// Mark mutually exclusive flags
	rootCmd.MarkFlagsMutuallyExclusive("copy", "move")
	rootCmd.MarkFlagsMutuallyExclusive("album", "album-from-directory")
}

func run(cmd *cobra.Command, args []string) error {
	// Check if ExifTool is installed
	if err := checkExifTool(); err != nil {
		return err
	}

	// Validate flags
	if !copyMode && !moveMode {
		return fmt.Errorf("must specify either --copy or --move")
	}

	if clean && !moveMode {
		return fmt.Errorf("--clean requires --move")
	}

	// Parse arguments
	sourceDirs := args[:len(args)-1]
	destDir := args[len(args)-1]

	// Validate paths
	for _, src := range sourceDirs {
		if _, err := os.Stat(src); os.IsNotExist(err) {
			return fmt.Errorf("source directory does not exist: %s", src)
		}
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal. Canceling...")
		cancel()
	}()

	// Convert day adjust to string if needed
	dayAdjustStr := ""
	if dayAdjust != 0 {
		dayAdjustStr = fmt.Sprintf("%d", dayAdjust)
	}

	// Build processing config
	cfg := &config.ProcessingConfig{
		OldNaming:    oldNaming,
		RawPath:      rawPath,
		Move:         moveMode,
		Precision:    precision,
		DryRun:       dryRun,
		TimeAdjust:   timeAdjust,
		DayAdjust:    dayAdjustStr,
		Tags:         tags,
		Album:        album,
		AlbumFromDir: albumFromDir,
	}

	if dryRun {
		fmt.Println("DRY RUN - no files will be modified")
	}

	// Print operation summary
	if verbose > 0 {
		fmt.Printf("Operation: ")
		if copyMode {
			fmt.Println("copy")
		} else {
			fmt.Println("move")
		}
		fmt.Printf("Workers: %d\n", numWorkers)
		fmt.Printf("Source(s): %v\n", sourceDirs)
		fmt.Printf("Destination: %s\n", destDir)
		if rawPath != "" {
			fmt.Printf("RAW path: %s\n", rawPath)
		}
	}

	// Collect files to process
	files, err := collectFiles(sourceDirs, recursive, verbose)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No files to process")
		return nil
	}

	fmt.Printf("Found %d files to process\n", len(files))

	// Process files
	stats, err := processFiles(ctx, files, destDir, cfg, numWorkers, verbose)
	if err != nil {
		return err
	}

	// Print summary
	printSummary(stats, verbose)

	// Clean empty directories if requested (only for move operations)
	if clean && moveMode && !dryRun {
		fmt.Println("\nCleaning empty directories...")
		cleanStats := cleanEmptyDirectories(sourceDirs, recursive, verbose)
		if cleanStats.Removed > 0 {
			fmt.Printf("Removed %d empty directories\n", cleanStats.Removed)
		}
	}

	return nil
}

// Stats tracks processing statistics
type Stats struct {
	Processed  int64
	Duplicates int64
	Skipped    int64
	Errors     int64
}

// collectFiles walks source directories and collects all supported image/video files
func collectFiles(sourceDirs []string, recursive bool, verbose int) ([]string, error) {
	var files []string
	seen := make(map[string]bool) // Deduplicate if multiple sources overlap

	for _, sourceDir := range sourceDirs {
		if recursive {
			err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				// Check if file has valid extension
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
				return nil, fmt.Errorf("failed to walk directory %s: %w", sourceDir, err)
			}
		} else {
			// Non-recursive: only process files directly in the directory
			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				return nil, fmt.Errorf("failed to read directory %s: %w", sourceDir, err)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				path := filepath.Join(sourceDir, entry.Name())
				ext := strings.TrimPrefix(filepath.Ext(path), ".")
				if rename.IsValidExtension(ext) {
					absPath, err := filepath.Abs(path)
					if err != nil {
						return nil, err
					}
					if !seen[absPath] {
						files = append(files, absPath)
						seen[absPath] = true
					}
				}
			}
		}
	}

	return files, nil
}

// checkExifTool verifies that exiftool is installed and available
func checkExifTool() error {
	_, err := exec.LookPath("exiftool")
	if err != nil {
		return fmt.Errorf(`exiftool not found. Please install it first:

macOS:    brew install exiftool
Ubuntu:   sudo apt-get install libimage-exiftool-perl
Windows:  Download from https://exiftool.org/

After installation, verify with: exiftool -ver`)
	}
	return nil
}

// processFiles processes all files using a worker pool
func processFiles(ctx context.Context, files []string, destDir string, cfg *config.ProcessingConfig, workers int, verbose int) (*Stats, error) {
	stats := &Stats{}

	// Create progress bar (only if not verbose)
	var bar *progressbar.ProgressBar
	if verbose == 0 {
		bar = progressbar.NewOptions(len(files),
			progressbar.OptionSetDescription("Processing"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(40),
			progressbar.OptionThrottle(65*1000000), // 65ms
			progressbar.OptionShowElapsedTimeOnFinish(),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
		)
	}

	// Create worker pool with bounded queue
	pool := pond.New(workers, len(files))

	// Monitor context cancellation
	go func() {
		<-ctx.Done()
		pool.StopAndWait()
	}()

	// Process each file
	for _, file := range files {
		file := file // Capture for closure

		// Check if context is canceled before submitting
		select {
		case <-ctx.Done():
			pool.StopAndWait()
			if bar != nil {
				bar.Finish()
			}
			return stats, fmt.Errorf("processing canceled by user")
		default:
		}

		pool.Submit(func() {
			// Check if context is canceled
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := processFile(file, destDir, cfg, stats, verbose); err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				if verbose > 0 {
					fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
				}
			}
			// Update progress bar
			if bar != nil {
				bar.Add(1)
			}
		})
	}

	// Wait for all tasks to complete (or cancellation)
	pool.StopAndWait()

	// Finish progress bar
	if bar != nil {
		bar.Finish()
	}

	// Check if we were canceled
	if ctx.Err() != nil {
		return stats, fmt.Errorf("processing canceled by user")
	}

	return stats, nil
}

// processFile processes a single file
func processFile(file string, destDir string, cfg *config.ProcessingConfig, stats *Stats, verbose int) error {
	// Create ImageRename instance
	ir, err := rename.NewImageRename(file, destDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to create rename instance: %w", err)
	}
	defer ir.Close()

	// Check if valid extension
	if !ir.IsValidExtension() {
		atomic.AddInt64(&stats.Skipped, 1)
		if verbose > 1 {
			fmt.Printf("Skipping (unsupported): %s\n", file)
		}
		return nil
	}

	// Parse metadata
	if err := ir.ParseMetadata(); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Check if duplicate
	if ir.IsDuplicate() {
		atomic.AddInt64(&stats.Duplicates, 1)
		if verbose > 1 {
			fmt.Printf("Skipping (duplicate): %s\n", file)
		}
		return nil
	}

	// Show what we're doing
	if verbose > 0 {
		operation := "Copying"
		if cfg.Move {
			operation = "Moving"
		}
		if cfg.DryRun {
			operation = "[DRY RUN] " + operation
		}
		fmt.Printf("%s: %s -> %s\n", operation, file, ir.GetDestination())
	}

	// Perform the operation
	if err := ir.Perform(); err != nil {
		return fmt.Errorf("failed to perform operation: %w", err)
	}

	atomic.AddInt64(&stats.Processed, 1)
	return nil
}

// printSummary prints processing statistics
func printSummary(stats *Stats, verbose int) {
	fmt.Println("\nSummary:")
	fmt.Printf("  Processed:  %d\n", stats.Processed)
	if stats.Duplicates > 0 {
		fmt.Printf("  Duplicates: %d\n", stats.Duplicates)
	}
	if stats.Skipped > 0 {
		fmt.Printf("  Skipped:    %d\n", stats.Skipped)
	}
	if stats.Errors > 0 {
		fmt.Printf("  Errors:     %d\n", stats.Errors)
	}
}

// CleanStats tracks directory cleaning statistics
type CleanStats struct {
	Checked int
	Removed int
}

// cleanEmptyDirectories removes empty directories from source paths
func cleanEmptyDirectories(sourceDirs []string, recursive bool, verbose int) *CleanStats {
	stats := &CleanStats{}

	for _, sourceDir := range sourceDirs {
		if recursive {
			// Walk bottom-up to remove nested empty directories
			cleanEmptyDirsRecursive(sourceDir, stats, verbose)
		} else {
			// Only check the source directory itself
			if isEmpty, _ := isDirEmpty(sourceDir); isEmpty {
				if verbose > 0 {
					fmt.Printf("Removing empty directory: %s\n", sourceDir)
				}
				if err := os.Remove(sourceDir); err == nil {
					stats.Removed++
				}
				stats.Checked++
			}
		}
	}

	return stats
}

// cleanEmptyDirsRecursive recursively removes empty directories
func cleanEmptyDirsRecursive(dir string, stats *CleanStats, verbose int) {
	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// First, recursively clean subdirectories
	for _, entry := range entries {
		if entry.IsDir() {
			subdir := filepath.Join(dir, entry.Name())
			cleanEmptyDirsRecursive(subdir, stats, verbose)
		}
	}

	// Now check if this directory is empty and remove it
	stats.Checked++
	if isEmpty, _ := isDirEmpty(dir); isEmpty {
		if verbose > 0 {
			fmt.Printf("Removing empty directory: %s\n", dir)
		}
		if err := os.Remove(dir); err == nil {
			stats.Removed++
		}
	}
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
