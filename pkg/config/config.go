package config

// ProcessingConfig holds all configuration options for image processing operations.
type ProcessingConfig struct {
	// OldNaming uses legacy filename format without make/model
	OldNaming bool

	// RawPath is an optional separate destination directory for RAW files
	RawPath string

	// Move determines whether to move (true) or copy (false) files
	Move bool

	// Precision is the number of subsecond digits to include in filenames (default: 6)
	Precision int

	// DryRun previews operations without executing them
	DryRun bool

	// TimeAdjust is a time adjustment string in "HH:MM:SS" format (can be negative)
	TimeAdjust string

	// DayAdjust is a day adjustment string as an integer (can be negative)
	DayAdjust string

	// Tags are keywords to add to image metadata
	Tags []string

	// Album is the album name to write to XMP:Album metadata
	Album string

	// AlbumFromDir extracts the album name from the parent directory
	AlbumFromDir bool
}
