package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	verifyFix bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify [flags] DIRECTORY...",
	Short: "Verify archive filenames match EXIF metadata",
	Long: `Verify that filenames in an organized archive match their EXIF metadata.

This command validates that:
  - Filenames match EXIF DateTimeOriginal
  - Camera make/model in filename matches EXIF
  - No duplicate files exist (same content, different names)

Optional --fix mode will rename files to match EXIF data.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().BoolVar(&verifyFix, "fix", false, "automatically fix mismatches")
}

func runVerify(cmd *cobra.Command, args []string) error {
	dirs := args

	fmt.Printf("Verifying directories: %v\n", dirs)
	if verifyFix {
		fmt.Println("Fix mode: enabled")
	}

	return fmt.Errorf("not implemented yet - coming soon!")
}
