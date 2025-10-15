package cmd

import (
	"fmt"
	"os"
	"runtime"

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

	if dryRun {
		fmt.Println("DRY RUN - no files will be modified")
	}

	// TODO: Implement actual processing logic
	fmt.Printf("Sources: %v\n", sourceDirs)
	fmt.Printf("Destination: %s\n", destDir)
	fmt.Printf("Operation: ")
	if copyMode {
		fmt.Println("copy")
	} else {
		fmt.Println("move")
	}
	fmt.Printf("Workers: %d\n", numWorkers)

	return fmt.Errorf("not implemented yet - coming soon!")
}
