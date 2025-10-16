package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/cacack/sortpics-go/pkg/config"
)

// DATE_PATTERN matches filenames like: YYYYMMDD-HHMMSS.subsec
// Example: 20240115-123045.123456
var DATE_PATTERN = regexp.MustCompile(`([0-9]{8})(.)?([0-9]{6})?(.)?([0-9]+)?`)

// ExifNotFoundError is returned when exiftool is not available
type ExifNotFoundError struct {
	Err error
}

func (e *ExifNotFoundError) Error() string {
	return fmt.Sprintf("exiftool not found: %v", e.Err)
}

// MetadataExtractor extracts and parses metadata from image files.
//
// Uses a fallback hierarchy for datetime extraction:
// 1. EXIF:DateTimeOriginal or EXIF:ModifyDate (with SubSecTimeOriginal)
// 2. QuickTime:CreateDate (for MOV files)
// 3. Datetime pattern in filename (YYYYMMDD-HHMMSS.subsec)
// 4. File's ctime from filesystem
type MetadataExtractor struct {
	et *exiftool.Exiftool
}

// NewMetadataExtractor creates a new MetadataExtractor with an ExifTool instance.
// The caller is responsible for calling Close() when done.
func NewMetadataExtractor() (*MetadataExtractor, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, &ExifNotFoundError{Err: err}
	}
	return &MetadataExtractor{et: et}, nil
}

// Close closes the ExifTool process.
func (m *MetadataExtractor) Close() error {
	if m.et != nil {
		return m.et.Close()
	}
	return nil
}

// Extract extracts metadata from a file.
//
// Args:
//   - filePath: Path to the image file
//   - timeAdjust: Optional duration for time adjustment
//   - dayAdjust: Optional duration for day adjustment
//
// Returns ImageMetadata with extracted values or an error.
func (m *MetadataExtractor) Extract(filePath string, timeAdjust, dayAdjust *time.Duration) (*config.ImageMetadata, error) {
	// Get file stats (needed for ctime fallback)
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	// Extract raw metadata using exiftool
	rawMetadata, err := m.getMetadata(filePath)
	if err != nil {
		return nil, err
	}

	// Parse datetime with fallback hierarchy
	dt := m.parseDatetime(filePath, rawMetadata, fileStat)

	// Apply time/day adjustments if provided
	if timeAdjust != nil && dt != nil {
		adjusted := dt.Add(*timeAdjust)
		dt = &adjusted
	}
	if dayAdjust != nil && dt != nil {
		adjusted := dt.Add(*dayAdjust)
		dt = &adjusted
	}

	// Parse make and model
	make := m.parseMake(rawMetadata)
	model := m.parseModel(make, rawMetadata)

	return &config.ImageMetadata{
		DateTime:    dt,
		Make:        make,
		Model:       model,
		RawMetadata: rawMetadata,
	}, nil
}

// getMetadata gets raw metadata from file using exiftool
func (m *MetadataExtractor) getMetadata(filePath string) (map[string]interface{}, error) {
	fileInfos := m.et.ExtractMetadata(filePath)
	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("no metadata returned for file: %s", filePath)
	}

	fileInfo := fileInfos[0]
	if fileInfo.Err != nil {
		return nil, fmt.Errorf("exiftool error: %w", fileInfo.Err)
	}

	return fileInfo.Fields, nil
}

// parseDatetime parses datetime from metadata with fallback hierarchy
//
// Tries in order:
// 1. EXIF datetime fields (DateTimeOriginal or ModifyDate with SubSecTimeOriginal)
// 2. QuickTime datetime fields (CreateDate for videos)
// 3. Datetime pattern in filename
// 4. File ctime
func (m *MetadataExtractor) parseDatetime(filePath string, rawMetadata map[string]interface{}, fileStat os.FileInfo) *time.Time {
	// Try EXIF datetime fields (with and without EXIF: prefix)
	for _, key := range []string{"EXIF:DateTimeOriginal", "DateTimeOriginal", "EXIF:ModifyDate", "ModifyDate"} {
		if dateTimeRaw, ok := rawMetadata[key]; ok {
			if dateTimeStr, ok := dateTimeRaw.(string); ok {
				// Parse base datetime: "2024:01:15 12:30:45"
				dt, err := time.Parse("2006:01:02 15:04:05", dateTimeStr)
				if err != nil {
					continue
				}

				// Try to get subsecond precision and add it (with and without EXIF: prefix)
				for _, subsecKey := range []string{"EXIF:SubSecTimeOriginal", "SubSecTimeOriginal"} {
					if subsec, ok := rawMetadata[subsecKey]; ok {
						// Handle both string and numeric types
						var subsecStr string
						switch v := subsec.(type) {
						case string:
							subsecStr = v
						case int, int64, float64:
							subsecStr = fmt.Sprintf("%v", v)
						}
						if subsecStr != "" {
							microseconds := parseSubseconds(subsecStr)
							dt = dt.Add(time.Duration(microseconds) * time.Microsecond)
							break
						}
					}
				}

				return &dt
			}
		}
	}

	// Try QuickTime (MOV files) (with and without QuickTime: prefix)
	for _, key := range []string{"QuickTime:CreateDate", "CreateDate"} {
		if dateTimeRaw, ok := rawMetadata[key]; ok {
			if dateTimeStr, ok := dateTimeRaw.(string); ok {
				if dt, err := time.Parse("2006:01:02 15:04:05", dateTimeStr); err == nil {
					return &dt
				}
			}
		}
	}

	// Try to extract from filename
	if match := DATE_PATTERN.FindStringSubmatch(filepath.Base(filePath)); match != nil {
		timestamp := ""
		if match[1] != "" {
			timestamp = match[1]
		}
		if match[3] != "" {
			timestamp = fmt.Sprintf("%s-%s", timestamp, match[3])
		}
		if match[5] != "" {
			timestamp = fmt.Sprintf("%s.%s", timestamp, match[5])
		}

		if timestamp != "" {
			// Try parsing the extracted timestamp
			// Format: YYYYMMDD-HHMMSS.subsec
			for _, layout := range []string{
				"20060102-150405.999999",
				"20060102-150405",
				"20060102",
			} {
				if dt, err := time.Parse(layout, timestamp); err == nil {
					return &dt
				}
			}
		}
	}

	// Fall back to file ctime
	// Note: Go's FileInfo doesn't expose ctime directly, using ModTime as fallback
	dt := fileStat.ModTime()
	return &dt
}

// parseMake parses camera make from metadata
//
// Handles special cases like HTC, LG, and filters out "Research".
// Returns "Unknown" if make is not found.
func (m *MetadataExtractor) parseMake(rawMetadata map[string]interface{}) string {
	var make string

	// Try various make keys (with and without prefixes)
	for _, key := range []string{"EXIF:Make", "Make", "MakerNotes:Make"} {
		if makeRaw, ok := rawMetadata[key]; ok {
			if makeStr, ok := makeRaw.(string); ok {
				make = makeStr
				break
			}
		}
	}

	if make != "" {
		// Take first word and capitalize
		words := strings.Fields(make)
		if len(words) > 0 {
			make = strings.Title(strings.ToLower(words[0]))

			// Handle special cases
			switch make {
			case "Htc":
				return "HTC"
			case "Lg":
				return "LG"
			case "Research":
				return ""
			}

			return make
		}
	}

	return "Unknown"
}

// parseModel parses camera model from metadata
//
// Removes make from model name and normalizes formatting.
// Returns empty string if model is not found.
func (m *MetadataExtractor) parseModel(make string, rawMetadata map[string]interface{}) string {
	var model string

	// Try various model keys (with and without prefixes)
	for _, key := range []string{"EXIF:Model", "Model", "MakerNotes:Model"} {
		if modelRaw, ok := rawMetadata[key]; ok {
			if modelStr, ok := modelRaw.(string); ok {
				model = modelStr
				break
			}
		}
	}

	// Remove make from the model
	if make != "" && model != "" {
		model = strings.ReplaceAll(model, make, "")
		model = strings.ReplaceAll(model, strings.ToUpper(make), "")
		model = strings.TrimSpace(model)
	}

	// Normalize spaces to CamelCase
	if strings.Contains(model, " ") {
		words := strings.Fields(model)
		var camelCaseParts []string
		for _, word := range words {
			camelCaseParts = append(camelCaseParts, strings.Title(strings.ToLower(word)))
		}
		model = strings.Join(camelCaseParts, "")
	}

	if model == "" {
		return ""
	}

	return model
}

// parseSubseconds parses subsecond string to microseconds
func parseSubseconds(subsecStr string) int {
	// Pad or truncate to 6 digits for microseconds
	if len(subsecStr) > 6 {
		subsecStr = subsecStr[:6]
	} else if len(subsecStr) < 6 {
		subsecStr = subsecStr + strings.Repeat("0", 6-len(subsecStr))
	}

	microseconds, _ := strconv.Atoi(subsecStr)
	return microseconds
}
