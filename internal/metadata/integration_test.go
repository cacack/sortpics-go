package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationBasicFixtures tests metadata extraction from real image files
func TestIntegrationBasicFixtures(t *testing.T) {
	fixturesDir := "../../test/testdata"
	manifestPath := filepath.Join(fixturesDir, "manifest.json")

	// Skip if fixtures don't exist
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skip("Test fixtures not available")
	}

	// Load manifest
	manifestData, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var manifest map[string]struct {
		DateTime         string  `json:"datetime"`
		Make             string  `json:"make"`
		Model            string  `json:"model"`
		ExpectedFilename string  `json:"expected_filename"`
		ExpectedPath     string  `json:"expected_path"`
		ExpectedAction   string  `json:"expected_action"`
		Note             string  `json:"note"`
	}
	err = json.Unmarshal(manifestData, &manifest)
	require.NoError(t, err)

	// Create extractor
	extractor, err := NewMetadataExtractor()
	if err != nil {
		t.Skip("ExifTool not available, skipping integration test")
	}
	defer extractor.Close()

	// Test each basic fixture
	basicFixtures := []string{
		"basic/test_001.jpg",
		"basic/test_002.jpg",
		"basic/test_003.jpg",
		"basic/test_004.jpg",
		"basic/test_005.jpg",
	}

	for _, fixturePath := range basicFixtures {
		t.Run(fixturePath, func(t *testing.T) {
			fullPath := filepath.Join(fixturesDir, fixturePath)
			expected := manifest[fixturePath]

			// Extract metadata
			metadata, err := extractor.Extract(fullPath, nil, nil)
			require.NoError(t, err)
			require.NotNil(t, metadata)

			// Parse expected datetime
			expectedDT, err := time.Parse("2006-01-02T15:04:05.999999", expected.DateTime)
			require.NoError(t, err)

			// Verify datetime (allowing small precision differences)
			require.NotNil(t, metadata.DateTime)
			assert.Equal(t, expectedDT.Year(), metadata.DateTime.Year())
			assert.Equal(t, expectedDT.Month(), metadata.DateTime.Month())
			assert.Equal(t, expectedDT.Day(), metadata.DateTime.Day())
			assert.Equal(t, expectedDT.Hour(), metadata.DateTime.Hour())
			assert.Equal(t, expectedDT.Minute(), metadata.DateTime.Minute())
			assert.Equal(t, expectedDT.Second(), metadata.DateTime.Second())

			// Verify make and model
			assert.Equal(t, expected.Make, metadata.Make, "Make should match for %s", fixturePath)
			assert.Equal(t, expected.Model, metadata.Model, "Model should match for %s", fixturePath)
		})
	}
}

// TestIntegrationSpecialMakes tests special manufacturer name handling
func TestIntegrationSpecialMakes(t *testing.T) {
	fixturesDir := "../../test/testdata"
	manifestPath := filepath.Join(fixturesDir, "manifest.json")

	// Skip if fixtures don't exist
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Skip("Test fixtures not available")
	}

	// Load manifest
	manifestData, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var manifest map[string]struct {
		Make  *string `json:"make"`
		Model string  `json:"model"`
		Note  string  `json:"note"`
	}
	err = json.Unmarshal(manifestData, &manifest)
	require.NoError(t, err)

	// Create extractor
	extractor, err := NewMetadataExtractor()
	if err != nil {
		t.Skip("ExifTool not available, skipping integration test")
	}
	defer extractor.Close()

	testCases := []struct {
		fixture      string
		expectedMake string
		note         string
	}{
		{
			fixture:      "special_makes/htc_001.jpg",
			expectedMake: "HTC",
			note:         "HTC Corporation should normalize to HTC",
		},
		{
			fixture:      "special_makes/lg_001.jpg",
			expectedMake: "LG",
			note:         "LG Electronics should normalize to LG",
		},
		{
			fixture:      "special_makes/rim_001.jpg",
			expectedMake: "",
			note:         "Research In Motion should be filtered out",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fullPath := filepath.Join(fixturesDir, tc.fixture)

			metadata, err := extractor.Extract(fullPath, nil, nil)
			require.NoError(t, err)
			require.NotNil(t, metadata)

			assert.Equal(t, tc.expectedMake, metadata.Make, tc.note)
		})
	}
}

// TestIntegrationNoEXIF tests fallback mechanisms for files without EXIF
func TestIntegrationNoEXIF(t *testing.T) {
	fixturesDir := "../../test/testdata"

	// Create extractor
	extractor, err := NewMetadataExtractor()
	if err != nil {
		t.Skip("ExifTool not available, skipping integration test")
	}
	defer extractor.Close()

	t.Run("datetime from filename pattern", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "no_exif/20240615-143022.123456_test.jpg")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skip("Test fixture not available")
		}

		metadata, err := extractor.Extract(fixturePath, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		// Should extract datetime from filename
		require.NotNil(t, metadata.DateTime)
		assert.Equal(t, 2024, metadata.DateTime.Year())
		assert.Equal(t, time.June, metadata.DateTime.Month())
		assert.Equal(t, 15, metadata.DateTime.Day())
		assert.Equal(t, 14, metadata.DateTime.Hour())
		assert.Equal(t, 30, metadata.DateTime.Minute())
		assert.Equal(t, 22, metadata.DateTime.Second())

		// Should default to Unknown for make
		assert.Equal(t, "Unknown", metadata.Make)
	})

	t.Run("fallback to file mtime", func(t *testing.T) {
		fixturePath := filepath.Join(fixturesDir, "no_exif/no_metadata.jpg")
		if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
			t.Skip("Test fixture not available")
		}

		metadata, err := extractor.Extract(fixturePath, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		// Should fall back to file modification time
		require.NotNil(t, metadata.DateTime)

		// Get file stat to compare
		stat, err := os.Stat(fixturePath)
		require.NoError(t, err)
		assert.Equal(t, stat.ModTime().Unix(), metadata.DateTime.Unix())
	})
}

// TestIntegrationTimeAdjustments tests time and day adjustments
func TestIntegrationTimeAdjustments(t *testing.T) {
	fixturesDir := "../../test/testdata"
	fixturePath := filepath.Join(fixturesDir, "basic/test_001.jpg")

	// Skip if fixture doesn't exist
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Test fixture not available")
	}

	extractor, err := NewMetadataExtractor()
	if err != nil {
		t.Skip("ExifTool not available, skipping integration test")
	}
	defer extractor.Close()

	t.Run("time adjustment", func(t *testing.T) {
		adjustment := 2*time.Hour + 30*time.Minute
		metadata, err := extractor.Extract(fixturePath, &adjustment, nil)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.NotNil(t, metadata.DateTime)

		// Original time is 12:30:45, adjusted should be 15:00:45
		assert.Equal(t, 15, metadata.DateTime.Hour())
		assert.Equal(t, 0, metadata.DateTime.Minute())
	})

	t.Run("day adjustment", func(t *testing.T) {
		adjustment := 5 * 24 * time.Hour
		metadata, err := extractor.Extract(fixturePath, nil, &adjustment)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.NotNil(t, metadata.DateTime)

		// Original date is 2024-01-15, adjusted should be 2024-01-20
		assert.Equal(t, 2024, metadata.DateTime.Year())
		assert.Equal(t, time.January, metadata.DateTime.Month())
		assert.Equal(t, 20, metadata.DateTime.Day())
	})
}

// BenchmarkIntegrationExtract benchmarks metadata extraction from real files
func BenchmarkIntegrationExtract(b *testing.B) {
	fixturesDir := "../../test/testdata"
	fixturePath := filepath.Join(fixturesDir, "basic/test_001.jpg")

	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		b.Skip("Test fixture not available")
	}

	extractor, err := NewMetadataExtractor()
	if err != nil {
		b.Skip("ExifTool not available, skipping benchmark")
	}
	defer extractor.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.Extract(fixturePath, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
