package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chris/sortpics-go/pkg/config"
)

// BenchmarkCopyMode benchmarks copy operation with worker pool
func BenchmarkCopyMode(b *testing.B) {
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")

	// Reset flags for benchmark
	copyMode = true
	moveMode = false
	dryRun = false
	recursive = false
	verbose = 0
	numWorkers = 8
	precision = 6
	oldNaming = false
	rawPath = ""
	album = ""
	albumFromDir = false
	tags = []string{}
	timeAdjust = ""
	dayAdjust = 0
	clean = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create temp directory for each iteration
		tmpDir, err := os.MkdirTemp("", "sortpics-bench-*")
		if err != nil {
			b.Fatal(err)
		}

		err = run(nil, []string{testDataDir, tmpDir})
		if err != nil {
			b.Fatal(err)
		}

		os.RemoveAll(tmpDir)
	}
}

// BenchmarkProcessFiles benchmarks the file processing function
func BenchmarkProcessFiles(b *testing.B) {
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")
	files, err := collectFiles([]string{testDataDir}, false, 0)
	if err != nil {
		b.Fatal(err)
	}

	cfg := &config.ProcessingConfig{
		Precision: 6,
		OldNaming: false,
		DryRun:    true, // Use dry-run to avoid actual file I/O in benchmark
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpDir, err := os.MkdirTemp("", "sortpics-bench-*")
		if err != nil {
			b.Fatal(err)
		}

		_, err = processFiles(ctx, files, tmpDir, cfg, 8, 0)
		if err != nil {
			b.Fatal(err)
		}

		os.RemoveAll(tmpDir)
	}
}

// BenchmarkCollectFiles benchmarks directory walking
func BenchmarkCollectFiles(b *testing.B) {
	testDataRoot := filepath.Join("..", "..", "..", "test", "testdata")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := collectFiles([]string{testDataRoot}, true, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProcessFilesParallel benchmarks with different worker counts
func BenchmarkProcessFilesParallel(b *testing.B) {
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata", "basic")
	files, err := collectFiles([]string{testDataDir}, false, 0)
	if err != nil {
		b.Fatal(err)
	}

	cfg := &config.ProcessingConfig{
		Precision: 6,
		OldNaming: false,
		DryRun:    true,
	}

	ctx := context.Background()

	for _, workers := range []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tmpDir, err := os.MkdirTemp("", "sortpics-bench-*")
				if err != nil {
					b.Fatal(err)
				}

				_, err = processFiles(ctx, files, tmpDir, cfg, workers, 0)
				if err != nil {
					b.Fatal(err)
				}

				os.RemoveAll(tmpDir)
			}
		})
	}
}
