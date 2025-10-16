package config

import "time"

// ImageMetadata represents metadata extracted from an image file.
// This struct is used for generating destination paths and filenames.
type ImageMetadata struct {
	// DateTime is the creation date/time of the image, extracted from EXIF data,
	// video metadata, filename pattern, or filesystem ctime (in order of priority).
	DateTime *time.Time

	// Make is the camera manufacturer (e.g., "Canon", "Nikon").
	// Normalized to be capitalized.
	Make string

	// Model is the camera model (e.g., "EOS5D", "D850").
	// Normalized with make prefix removed and capitalized.
	Model string

	// RawMetadata contains the raw EXIF data as returned by ExifTool.
	// This is kept for potential future use or debugging.
	RawMetadata map[string]interface{}
}
