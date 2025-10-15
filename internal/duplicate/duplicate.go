// Package duplicate provides SHA256-based duplicate detection and filename collision resolution.
package duplicate

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Detector detects duplicate files and resolves filename collisions.
//
// Uses SHA256 hashing to determine if files are identical.
// Resolves collisions by appending _N suffix to filenames.
type Detector struct{}

// New creates a new duplicate detector.
func New() *Detector {
	return &Detector{}
}

// CalculateSHA256 calculates the SHA256 hash of a file.
//
// If an _original backup exists (from exiftool), use that to get the
// pre-modification hash for accurate duplicate detection.
func (d *Detector) CalculateSHA256(filePath string) (string, error) {
	// Check if _original backup exists (from exiftool EXIF writing)
	originalPath := filePath + "_original"
	hashPath := filePath

	if _, err := os.Stat(originalPath); err == nil {
		hashPath = originalPath
	}

	file, err := os.Open(hashPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	buffer := make([]byte, 4096)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}
		hash.Write(buffer[:n])
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// IsDuplicate checks if source and destination files are identical.
//
// Returns true if files have the same SHA256 hash, false otherwise.
func (d *Detector) IsDuplicate(source, destination string) (bool, error) {
	// If destination doesn't exist, it's not a duplicate
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		return false, nil
	}

	sourceHash, err := d.CalculateSHA256(source)
	if err != nil {
		return false, fmt.Errorf("failed to hash source: %w", err)
	}

	destHash, err := d.CalculateSHA256(destination)
	if err != nil {
		return false, fmt.Errorf("failed to hash destination: %w", err)
	}

	return sourceHash == destHash, nil
}

// ResolveCollision resolves filename collision by finding a unique path.
//
// If initialPath exists:
//   - If files are identical (same hash), return initialPath with source hash
//   - If files differ, append _N suffix until unique filename found
//
// Returns the resolved path and the source hash (nil if no collision occurred).
func (d *Detector) ResolveCollision(source, initialPath string) (string, *string, error) {
	// No collision - file doesn't exist
	if _, err := os.Stat(initialPath); os.IsNotExist(err) {
		return initialPath, nil, nil
	}

	// Calculate source hash once
	sourceHash, err := d.CalculateSHA256(source)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash source: %w", err)
	}

	// Check if files are identical
	destHash, err := d.CalculateSHA256(initialPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash initial destination: %w", err)
	}

	if sourceHash == destHash {
		// Files are identical - this is a duplicate
		return initialPath, &sourceHash, nil
	}

	// Files differ - find unique filename with increment
	increment := 1

	for {
		// Generate new path with increment
		currentPath := addIncrement(initialPath, increment)

		if _, err := os.Stat(currentPath); os.IsNotExist(err) {
			// Found unique path
			return currentPath, &sourceHash, nil
		}

		// Check if this existing file matches source
		destHash, err := d.CalculateSHA256(currentPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to hash collision path: %w", err)
		}

		if sourceHash == destHash {
			// Found matching file at this increment
			return currentPath, &sourceHash, nil
		}

		// Try next increment
		increment++

		// Safety limit
		if increment > 1000 {
			return "", nil, fmt.Errorf("too many collisions for %s", initialPath)
		}
	}
}

// CheckAndResolve checks for collisions and resolves them.
//
// Returns the final destination path and whether the file is a duplicate.
// is_duplicate is true if the file already exists with the same hash.
func (d *Detector) CheckAndResolve(source, initialDestination string) (string, bool, error) {
	finalPath, sourceHash, err := d.ResolveCollision(source, initialDestination)
	if err != nil {
		return "", false, err
	}

	// If source_hash is not nil, we calculated it (collision occurred)
	// If final_path exists and matches, it's a duplicate
	isDuplicate := false
	if sourceHash != nil {
		if _, err := os.Stat(finalPath); err == nil {
			destHash, err := d.CalculateSHA256(finalPath)
			if err != nil {
				return "", false, fmt.Errorf("failed to verify duplicate: %w", err)
			}
			isDuplicate = *sourceHash == destHash
		}
	}

	return finalPath, isDuplicate, nil
}

// addIncrement adds an increment suffix to a filename before the extension.
//
// Example: addIncrement("/path/file.jpg", 1) -> "/path/file_1.jpg"
func addIncrement(path string, increment int) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	newStem := fmt.Sprintf("%s_%d", stem, increment)
	return filepath.Join(dir, newStem+ext)
}
