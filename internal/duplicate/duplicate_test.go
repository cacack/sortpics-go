package duplicate

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	detector := New()
	assert.NotNil(t, detector)
}

func TestCalculateSHA256(t *testing.T) {
	t.Run("consistent hash", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		detector := New()
		hash1, err := detector.CalculateSHA256(testFile)
		require.NoError(t, err)

		hash2, err := detector.CalculateSHA256(testFile)
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("hash format", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		detector := New()
		hash, err := detector.CalculateSHA256(testFile)
		require.NoError(t, err)

		// Should be 64 hex characters
		assert.Len(t, hash, 64)
		for _, c := range hash {
			assert.Contains(t, "0123456789abcdef", string(c))
		}
	})

	t.Run("different files have different hashes", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		err := os.WriteFile(file1, []byte("content 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("content 2"), 0644)
		require.NoError(t, err)

		detector := New()
		hash1, err := detector.CalculateSHA256(file1)
		require.NoError(t, err)
		hash2, err := detector.CalculateSHA256(file2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("same content has same hash", func(t *testing.T) {
		tmpDir := t.TempDir()
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		err := os.WriteFile(file1, []byte("same content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(file2, []byte("same content"), 0644)
		require.NoError(t, err)

		detector := New()
		hash1, err := detector.CalculateSHA256(file1)
		require.NoError(t, err)
		hash2, err := detector.CalculateSHA256(file2)
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("binary file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.bin")
		err := os.WriteFile(testFile, []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, 0644)
		require.NoError(t, err)

		detector := New()
		hash, err := detector.CalculateSHA256(testFile)
		require.NoError(t, err)

		// Should still work with binary data
		assert.Len(t, hash, 64)
		for _, c := range hash {
			assert.Contains(t, "0123456789abcdef", string(c))
		}
	})

	t.Run("uses _original file if exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.jpg")
		originalFile := testFile + "_original"

		// Write different content to main file and _original
		err := os.WriteFile(testFile, []byte("modified content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(originalFile, []byte("original content"), 0644)
		require.NoError(t, err)

		detector := New()
		hash, err := detector.CalculateSHA256(testFile)
		require.NoError(t, err)

		// Should hash the _original file, not the modified one
		// Calculate what the hash should be for "original content"
		expectedHash, err := detector.CalculateSHA256(originalFile)
		require.NoError(t, err)

		assert.Equal(t, expectedHash, hash)
	})
}

func TestIsDuplicate(t *testing.T) {
	t.Run("destination doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("content"), 0644)
		require.NoError(t, err)

		detector := New()
		result, err := detector.IsDuplicate(source, dest)
		require.NoError(t, err)

		assert.False(t, result)
	})

	t.Run("same content", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("same content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("same content"), 0644)
		require.NoError(t, err)

		detector := New()
		result, err := detector.IsDuplicate(source, dest)
		require.NoError(t, err)

		assert.True(t, result)
	})

	t.Run("different content", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("content 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("content 2"), 0644)
		require.NoError(t, err)

		detector := New()
		result, err := detector.IsDuplicate(source, dest)
		require.NoError(t, err)

		assert.False(t, result)
	})
}

func TestResolveCollision(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("content"), 0644)
		require.NoError(t, err)

		detector := New()
		resolved, sourceHash, err := detector.ResolveCollision(source, dest)
		require.NoError(t, err)

		assert.Equal(t, dest, resolved)
		assert.Nil(t, sourceHash)
	})

	t.Run("duplicate file", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("same content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("same content"), 0644)
		require.NoError(t, err)

		detector := New()
		resolved, sourceHash, err := detector.ResolveCollision(source, dest)
		require.NoError(t, err)

		assert.Equal(t, dest, resolved)
		assert.NotNil(t, sourceHash)
		assert.Len(t, *sourceHash, 64)
	})

	t.Run("different file - single increment", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("content 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("content 2"), 0644)
		require.NoError(t, err)

		detector := New()
		resolved, sourceHash, err := detector.ResolveCollision(source, dest)
		require.NoError(t, err)

		expected := filepath.Join(tmpDir, "dest_1.txt")
		assert.Equal(t, expected, resolved)
		assert.NotNil(t, sourceHash)
	})

	t.Run("multiple increments", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		dest1 := filepath.Join(tmpDir, "dest_1.txt")
		dest2 := filepath.Join(tmpDir, "dest_2.txt")

		err := os.WriteFile(source, []byte("content source"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("content 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest1, []byte("content 2"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest2, []byte("content 3"), 0644)
		require.NoError(t, err)

		detector := New()
		resolved, sourceHash, err := detector.ResolveCollision(source, dest)
		require.NoError(t, err)

		expected := filepath.Join(tmpDir, "dest_3.txt")
		assert.Equal(t, expected, resolved)
		assert.NotNil(t, sourceHash)
	})

	t.Run("finds duplicate at increment", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		dest1 := filepath.Join(tmpDir, "dest_1.txt")
		dest2 := filepath.Join(tmpDir, "dest_2.txt")

		content := "duplicate content"
		err := os.WriteFile(source, []byte(content), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("different 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest1, []byte("different 2"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest2, []byte(content), 0644) // This one matches source
		require.NoError(t, err)

		detector := New()
		resolved, sourceHash, err := detector.ResolveCollision(source, dest)
		require.NoError(t, err)

		assert.Equal(t, dest2, resolved)
		assert.NotNil(t, sourceHash)
	})

	t.Run("preserves extension", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.jpg")
		dest := filepath.Join(tmpDir, "dest.jpg")
		err := os.WriteFile(source, []byte("content 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("content 2"), 0644)
		require.NoError(t, err)

		detector := New()
		resolved, sourceHash, err := detector.ResolveCollision(source, dest)
		require.NoError(t, err)

		expected := filepath.Join(tmpDir, "dest_1.jpg")
		assert.Equal(t, expected, resolved)
		assert.Equal(t, ".jpg", filepath.Ext(resolved))
		assert.NotNil(t, sourceHash)
	})

	t.Run("safety limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		err := os.WriteFile(source, []byte("unique content"), 0644)
		require.NoError(t, err)

		// Create 1001 collision files
		dest := filepath.Join(tmpDir, "dest.txt")
		err = os.WriteFile(dest, []byte("collision 0"), 0644)
		require.NoError(t, err)

		for i := 1; i <= 1000; i++ {
			collisionFile := filepath.Join(tmpDir, "dest_"+string(rune('0'+i/100))+string(rune('0'+(i/10)%10))+string(rune('0'+i%10))+".txt")
			// Use proper formatting for the filename
			collisionFile = filepath.Join(tmpDir, "dest_"+fmt.Sprintf("%d", i)+".txt")
			err = os.WriteFile(collisionFile, []byte("collision"), 0644)
			require.NoError(t, err)
		}

		detector := New()
		_, _, err = detector.ResolveCollision(source, dest)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many collisions")
	})
}

func TestCheckAndResolve(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("content"), 0644)
		require.NoError(t, err)

		detector := New()
		finalPath, isDuplicate, err := detector.CheckAndResolve(source, dest)
		require.NoError(t, err)

		assert.Equal(t, dest, finalPath)
		assert.False(t, isDuplicate)
	})

	t.Run("duplicate file", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("same content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("same content"), 0644)
		require.NoError(t, err)

		detector := New()
		finalPath, isDuplicate, err := detector.CheckAndResolve(source, dest)
		require.NoError(t, err)

		assert.Equal(t, dest, finalPath)
		assert.True(t, isDuplicate)
	})

	t.Run("collision - different files", func(t *testing.T) {
		tmpDir := t.TempDir()
		source := filepath.Join(tmpDir, "source.txt")
		dest := filepath.Join(tmpDir, "dest.txt")
		err := os.WriteFile(source, []byte("content 1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(dest, []byte("content 2"), 0644)
		require.NoError(t, err)

		detector := New()
		finalPath, isDuplicate, err := detector.CheckAndResolve(source, dest)
		require.NoError(t, err)

		expected := filepath.Join(tmpDir, "dest_1.txt")
		assert.Equal(t, expected, finalPath)
		assert.False(t, isDuplicate)
	})
}

func TestAddIncrement(t *testing.T) {
	t.Run("basic increment", func(t *testing.T) {
		path := "/path/to/file.jpg"
		result := addIncrement(path, 1)
		assert.Equal(t, "/path/to/file_1.jpg", result)
	})

	t.Run("increment 2", func(t *testing.T) {
		path := "/path/to/file.jpg"
		result := addIncrement(path, 2)
		assert.Equal(t, "/path/to/file_2.jpg", result)
	})

	t.Run("no extension", func(t *testing.T) {
		path := "/path/to/file"
		result := addIncrement(path, 1)
		assert.Equal(t, "/path/to/file_1", result)
	})

	t.Run("multiple dots", func(t *testing.T) {
		path := "/path/to/file.backup.tar.gz"
		result := addIncrement(path, 1)
		assert.Equal(t, "/path/to/file.backup.tar_1.gz", result)
	})
}
