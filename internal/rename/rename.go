package rename

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/cacack/sortpics-go/internal/duplicate"
	"github.com/cacack/sortpics-go/internal/metadata"
	"github.com/cacack/sortpics-go/internal/pathgen"
	"github.com/cacack/sortpics-go/pkg/config"
)

// ValidExtensions lists all supported image and video file extensions
var ValidExtensions = []string{
	// Standard images
	"jpg", "jpeg", "png", "tiff", "tif",
	// RAW formats
	"arw", "cr2", "crw", "dcr", "dng", "mrw", "nef", "nrw",
	"orf", "pef", "ptx", "raw", "rw2", "rwl", "srf", "sr2",
	"srw", "x3f",
	// Video formats
	"mov", "mp4", "m4v", "avi", "mpg", "mpeg",
}

// RawExtensions lists all RAW image file extensions
var RawExtensions = []string{
	"arw", // Sony
	"crw", // Canon
	"cr2", // Canon
	"dng", // Adobe, Leica
	"mrw", // Minolta
	"nef", // Nikon
	"nrw", // Nikon
	"orf", // Olympus
	"pef", // Pentax
	"ptx", // Pentax
	"raw", // Panasonic, Leica
	"rw2", // Panasonic
	"rwl", // Leica
	"srf", // Sony
	"sr2", // Sony
	"srw", // Samsung
	"x3f", // Sigma
}

// ImageRename orchestrates metadata extraction, path generation, and file operations
type ImageRename struct {
	config              *config.ProcessingConfig
	source              string
	destinationBase     string
	extension           string
	timeDelta           *time.Duration
	dayDelta            *time.Duration
	album               string
	tags                []string
	metadataExtractor   *metadata.MetadataExtractor
	pathGenerator       *pathgen.PathGenerator
	duplicateDetector   *duplicate.Detector

	// Results from ParseMetadata
	destination         string
	destinationDir      string
	isDuplicate         bool
	datetime            *time.Time
	make                string
	model               string
	rawMetadata         map[string]interface{}
}

// NewImageRename creates a new ImageRename instance
func NewImageRename(sourceFilename string, destinationBaseDir string, cfg *config.ProcessingConfig) (*ImageRename, error) {
	if cfg == nil {
		cfg = &config.ProcessingConfig{
			Precision: 6,
		}
	}

	// Resolve paths
	absSource, err := filepath.Abs(sourceFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source path: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(absSource), ".")

	// Calculate time and day deltas
	var timeDelta, dayDelta *time.Duration
	if cfg.TimeAdjust != "" {
		td, err := CalculateTimeDelta(cfg.TimeAdjust)
		if err != nil {
			return nil, fmt.Errorf("invalid time adjustment: %w", err)
		}
		timeDelta = &td
	}
	if cfg.DayAdjust != "" {
		dd, err := CalculateDayDelta(cfg.DayAdjust)
		if err != nil {
			return nil, fmt.Errorf("invalid day adjustment: %w", err)
		}
		dayDelta = &dd
	}

	// Determine destination base (RAW files may go to separate path)
	destBase := destinationBaseDir
	if IsRaw(extension) && cfg.RawPath != "" {
		destBase = cfg.RawPath
	}
	absDestBase, err := filepath.Abs(destBase)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Handle album from directory
	album := cfg.Album
	if cfg.AlbumFromDir {
		album = filepath.Base(filepath.Dir(absSource))
	}

	// Initialize metadata extractor
	metaExtractor, err := metadata.NewMetadataExtractor()
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata extractor: %w", err)
	}

	return &ImageRename{
		config:            cfg,
		source:            absSource,
		destinationBase:   absDestBase,
		extension:         extension,
		timeDelta:         timeDelta,
		dayDelta:          dayDelta,
		album:             album,
		tags:              cfg.Tags,
		metadataExtractor: metaExtractor,
		pathGenerator:     pathgen.New(cfg.Precision, cfg.OldNaming),
		duplicateDetector: duplicate.New(),
	}, nil
}

// Close cleans up resources (e.g., ExifTool process)
func (ir *ImageRename) Close() error {
	return ir.metadataExtractor.Close()
}

// IsValidExtension checks if the file extension is supported
func (ir *ImageRename) IsValidExtension() bool {
	return IsValidExtension(ir.extension)
}

// IsRaw checks if the file is a RAW image format
func (ir *ImageRename) IsRaw() bool {
	return IsRaw(ir.extension)
}

// ParseMetadata extracts metadata and generates destination path
func (ir *ImageRename) ParseMetadata() error {
	// Extract metadata
	meta, err := ir.metadataExtractor.Extract(ir.source, ir.timeDelta, ir.dayDelta)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Store extracted values
	ir.datetime = meta.DateTime
	ir.make = meta.Make
	ir.model = meta.Model
	ir.rawMetadata = meta.RawMetadata

	// Generate destination path (increment=0 for initial path)
	initialDestination := ir.pathGenerator.GeneratePath(meta, ir.destinationBase, ir.extension, 0)

	// Resolve collisions
	finalDestination, isDuplicate, err := ir.duplicateDetector.CheckAndResolve(ir.source, initialDestination)
	if err != nil {
		return fmt.Errorf("failed to check duplicates: %w", err)
	}

	ir.destination = finalDestination
	ir.destinationDir = filepath.Dir(finalDestination)
	ir.isDuplicate = isDuplicate

	return nil
}

// Perform executes the file operation (copy or move)
func (ir *ImageRename) Perform() error {
	if ir.config.DryRun {
		// In dry run mode, just return without doing anything
		return nil
	}

	// Create destination directory
	if err := os.MkdirAll(ir.destinationDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Re-check for collisions (race condition in multiprocessing)
	if _, err := os.Stat(ir.destination); err == nil {
		finalDestination, isDuplicate, err := ir.duplicateDetector.CheckAndResolve(ir.source, ir.destination)
		if err != nil {
			return fmt.Errorf("failed to recheck duplicates: %w", err)
		}
		if isDuplicate {
			// Skip duplicate files
			return nil
		}
		ir.destination = finalDestination
		ir.destinationDir = filepath.Dir(finalDestination)
		if err := os.MkdirAll(ir.destinationDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	// Perform copy or move
	if ir.config.Move {
		if err := SafeMove(ir.source, ir.destination); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
	} else {
		if err := SafeCopy(ir.source, ir.destination); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	// Write metadata tags
	if err := ir.writeMetadata(); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// writeMetadata writes EXIF and XMP tags to the destination file
func (ir *ImageRename) writeMetadata() error {
	if ir.datetime == nil {
		return nil
	}

	et, err := exiftool.NewExiftool()
	if err != nil {
		return fmt.Errorf("failed to initialize exiftool: %w", err)
	}
	defer et.Close()

	// Format datetime for EXIF
	datetimeStr := ir.datetime.Format("2006:01:02 15:04:05")

	// Extract metadata first to get FileMetadata structure
	fmList := et.ExtractMetadata(ir.destination)
	if len(fmList) == 0 {
		return fmt.Errorf("failed to extract metadata for writing")
	}

	fm := fmList[0]
	if fm.Err != nil {
		return fmt.Errorf("failed to extract metadata: %w", fm.Err)
	}

	// Set datetime tags
	fm.SetString("EXIF:DateTimeOriginal", datetimeStr)
	fm.SetString("EXIF:CreateDate", datetimeStr)
	fm.SetString("EXIF:ModifyDate", datetimeStr)

	// Add album if specified
	if ir.album != "" {
		fm.SetString("XMP:Album", ir.album)
	}

	// Add keywords if specified
	if len(ir.tags) > 0 {
		fm.SetStrings("Keywords", ir.tags)
	}

	// Write metadata back
	et.WriteMetadata([]exiftool.FileMetadata{fm})

	return nil
}

// GetDestination returns the destination path after ParseMetadata
func (ir *ImageRename) GetDestination() string {
	return ir.destination
}

// IsDuplicate returns whether the file is a duplicate
func (ir *ImageRename) IsDuplicate() bool {
	return ir.isDuplicate
}

// IsValidExtension checks if the given extension is supported
func IsValidExtension(ext string) bool {
	extLower := strings.ToLower(ext)
	for _, validExt := range ValidExtensions {
		if extLower == validExt {
			return true
		}
	}
	return false
}

// IsRaw checks if the given extension is a RAW format
func IsRaw(ext string) bool {
	extLower := strings.ToLower(ext)
	for _, rawExt := range RawExtensions {
		if extLower == rawExt {
			return true
		}
	}
	return false
}

// CalculateTimeDelta parses a time adjustment string in "HH:MM:SS" format
func CalculateTimeDelta(timeDelta string) (time.Duration, error) {
	parts := strings.Split(timeDelta, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format, expected HH:MM:SS")
	}

	var hours, minutes, seconds int
	var negate bool

	// Check for negative sign on hours
	if strings.HasPrefix(parts[0], "-") {
		negate = true
	}

	// Parse hours (will be negative if prefixed with -)
	if _, err := fmt.Sscanf(parts[0], "%d", &hours); err != nil {
		return 0, fmt.Errorf("invalid hours: %w", err)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minutes); err != nil {
		return 0, fmt.Errorf("invalid minutes: %w", err)
	}
	if _, err := fmt.Sscanf(parts[2], "%d", &seconds); err != nil {
		return 0, fmt.Errorf("invalid seconds: %w", err)
	}

	// Apply negation to minutes and seconds if hours were negative
	if negate {
		minutes = -minutes
		seconds = -seconds
	}

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second

	return duration, nil
}

// CalculateDayDelta parses a day adjustment string
func CalculateDayDelta(dayDelta string) (time.Duration, error) {
	var days int
	if _, err := fmt.Sscanf(dayDelta, "%d", &days); err != nil {
		return 0, fmt.Errorf("invalid day delta: %w", err)
	}
	return time.Duration(days) * 24 * time.Hour, nil
}

// SafeCopy copies a file atomically using a temporary file
func SafeCopy(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create temp file in destination directory
	destDir := filepath.Dir(dst)
	tmpFile, err := os.CreateTemp(destDir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Write data to temp file
	if _, err = tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	if err = os.Chmod(tmpPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err = os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// SafeMove moves a file atomically, handling cross-filesystem moves
func SafeMove(src, dst string) error {
	// Try atomic rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Check if it's a cross-filesystem error
	if linkErr, ok := err.(*os.LinkError); ok {
		if errno, ok := linkErr.Err.(syscall.Errno); ok && errno == syscall.EXDEV {
			// Cross-filesystem move: copy then delete
			if err := SafeCopy(src, dst); err != nil {
				return err
			}
			if err := os.Remove(src); err != nil {
				return fmt.Errorf("failed to remove source after copy: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("failed to move file: %w", err)
}

// generateUUID creates a simple UUID for temporary files
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
