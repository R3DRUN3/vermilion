package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

func CreateArchive(outputDir string, paths []string) ([]string, error) {
	// Ensure the output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	var archivesList []string
	ctx := context.TODO()

	// Create a single archive for all files and directories
	archiveName := "all_sensitive_data.zip"
	archivePath := filepath.Join(outputDir, archiveName)

	// Prepare a map of files to include in the archive
	fileMap := make(map[string]string)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Skipping inaccessible file or directory %s: %v\n", path, err)
			continue
		}

		if info.IsDir() {
			err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Printf("Skipping inaccessible file %s: %v\n", p, err)
					return nil // Continue walking the directory
				}
				if !info.IsDir() && canAccessFile(p) {
					// Replace path separators with underscores to create unique names
					relativePath, _ := filepath.Rel("/", p) // Use absolute root as base
					uniqueName := strings.ReplaceAll(relativePath, string(filepath.Separator), "_")
					fileMap[p] = uniqueName
				}
				return nil
			})
			if err != nil {
				fmt.Printf("Error walking directory %s: %v\n", path, err)
			}
		} else if canAccessFile(path) {
			// Replace path separators with underscores for individual files
			relativePath, _ := filepath.Rel("/", path) // Use absolute root as base
			uniqueName := strings.ReplaceAll(relativePath, string(filepath.Separator), "_")
			fileMap[path] = uniqueName
		}
	}

	if len(fileMap) == 0 {
		fmt.Println("No files to archive.")
		return nil, nil
	}

	// Include system_info.json if it exists in the output directory
	systemInfoPath := filepath.Join(outputDir, "system_info.json")
	if _, err := os.Stat(systemInfoPath); err == nil && canAccessFile(systemInfoPath) {
		fileMap[systemInfoPath] = "system_info.json"
	}

	// Create the output archive file
	out, err := os.Create(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive file %s: %w", archivePath, err)
	}
	defer out.Close()

	// Create a compressed tarball format
	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}

	// Add files to the archive
	files, err := archives.FilesFromDisk(ctx, nil, fileMap)
	if err != nil {
		return nil, fmt.Errorf("failed to map files from disk: %w", err)
	}

	// Write the archive
	err = format.Archive(ctx, out, files)
	if err != nil {
		return nil, fmt.Errorf("failed to archive files: %w", err)
	}

	fmt.Printf("Created archive at %s\n", archivePath)
	archivesList = append(archivesList, archivePath)

	return archivesList, nil
}

// Helper function to check if a file is accessible
func canAccessFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Skipping file %s: %v\n", path, err)
		return false
	}
	file.Close()
	return true
}
