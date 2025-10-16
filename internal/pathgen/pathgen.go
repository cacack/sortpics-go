package pathgen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cacack/sortpics-go/pkg/config"
)

// PathGenerator generates destination paths and filenames for organized photo archives.
//
// Filename format: YYYYMMDD-HHMMSS.subsec_Make-Model.ext
// Directory structure: YYYY/MM/YYYY-MM-DD/
type PathGenerator struct {
	// Precision is the number of subsecond digits to include (0-6).
	// Default is 6 for full microsecond precision.
	Precision int

	// OldNaming uses the legacy naming convention with no hyphen between make and model.
	// Format: YYYYMMDD-HHMMSS.subsec_MakeModel.ext (no hyphen between make and model)
	OldNaming bool
}

// New creates a new PathGenerator with the specified precision and naming convention.
func New(precision int, oldNaming bool) *PathGenerator {
	return &PathGenerator{
		Precision: precision,
		OldNaming: oldNaming,
	}
}

// GeneratePath generates the full destination path including directory and filename.
//
// Args:
//   - metadata: Image metadata containing datetime, make, model
//   - baseDir: Base destination directory
//   - extension: File extension without dot (e.g., "jpg", "cr2")
//   - increment: Optional collision increment (0 for no increment, >0 for _N suffix)
//
// Returns the full path: baseDir/YYYY/MM/YYYY-MM-DD/YYYYMMDD-HHMMSS.subsec_Make-Model.ext
func (pg *PathGenerator) GeneratePath(metadata *config.ImageMetadata, baseDir, extension string, increment int) string {
	directory := pg.GenerateDirectory(metadata, baseDir)
	filename := pg.GenerateFilename(metadata, extension, increment)
	return filepath.Join(directory, filename)
}

// GenerateDirectory generates the directory structure: baseDir/YYYY/MM/YYYY-MM-DD/
//
// If metadata.DateTime is nil, returns: baseDir/unknown/
func (pg *PathGenerator) GenerateDirectory(metadata *config.ImageMetadata, baseDir string) string {
	if metadata.DateTime == nil {
		return filepath.Join(baseDir, "unknown")
	}

	dt := metadata.DateTime
	year := fmt.Sprintf("%04d", dt.Year())
	month := fmt.Sprintf("%02d", int(dt.Month()))
	day := fmt.Sprintf("%02d", dt.Day())

	// Format: YYYY/MM/YYYY-MM-DD
	yearMonth := filepath.Join(year, month)
	fullDate := fmt.Sprintf("%s-%s-%s", year, month, day)

	return filepath.Join(baseDir, yearMonth, fullDate)
}

// GenerateFilename generates the filename: YYYYMMDD-HHMMSS.subsec_Make-Model.ext
//
// If metadata.DateTime is nil, returns: unknown_Make-Model.ext
// If both make and model are empty, uses "Unknown" for the camera part.
// Extension is always converted to lowercase.
func (pg *PathGenerator) GenerateFilename(metadata *config.ImageMetadata, extension string, increment int) string {
	// Generate camera part
	camera := pg.generateCameraPart(metadata)

	// Generate increment suffix
	incrementStr := ""
	if increment > 0 {
		incrementStr = fmt.Sprintf("_%d", increment)
	}

	// Convert extension to lowercase
	ext := strings.ToLower(extension)

	// Generate filename based on whether datetime is available
	if metadata.DateTime == nil {
		return fmt.Sprintf("unknown_%s%s.%s", camera, incrementStr, ext)
	}

	// Generate datetime and subsecond parts
	dt := metadata.DateTime
	datePart := fmt.Sprintf("%04d%02d%02d-%02d%02d%02d",
		dt.Year(), int(dt.Month()), dt.Day(),
		dt.Hour(), dt.Minute(), dt.Second())

	subsec := pg.generateSubsecPart(metadata)

	return fmt.Sprintf("%s.%s_%s%s.%s", datePart, subsec, camera, incrementStr, ext)
}

// generateCameraPart creates the camera portion of the filename.
//
// Returns one of:
//   - "Make-Model" (default)
//   - "MakeModel" (old naming, no hyphen)
//   - "Make" (if only make is present)
//   - "Model" (if only model is present)
//   - "Unknown" (if both are empty)
func (pg *PathGenerator) generateCameraPart(metadata *config.ImageMetadata) string {
	hasMake := metadata.Make != ""
	hasModel := metadata.Model != ""

	if hasMake && hasModel {
		if pg.OldNaming {
			return fmt.Sprintf("%s%s", metadata.Make, metadata.Model)
		}
		return fmt.Sprintf("%s-%s", metadata.Make, metadata.Model)
	}

	if hasMake {
		return metadata.Make
	}

	if hasModel {
		return metadata.Model
	}

	return "Unknown"
}

// generateSubsecPart creates the subsecond portion of the filename.
//
// Returns a string of digits with length equal to pg.Precision.
// If DateTime is nil or has no microseconds, returns "000000" (or fewer zeros based on precision).
func (pg *PathGenerator) generateSubsecPart(metadata *config.ImageMetadata) string {
	if metadata.DateTime == nil || metadata.DateTime.Nanosecond() == 0 {
		return strings.Repeat("0", pg.Precision)
	}

	// Convert nanoseconds to microseconds (6 digits)
	microseconds := metadata.DateTime.Nanosecond() / 1000
	fullSubsec := fmt.Sprintf("%06d", microseconds)

	// Return only the requested precision
	if pg.Precision > 6 {
		return fullSubsec
	}
	return fullSubsec[:pg.Precision]
}
