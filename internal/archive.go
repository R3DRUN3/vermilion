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

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Failed to stat path %s: %v\n", path, err)
			continue
		}

		// Generate archive name based on the path
		baseName := strings.ReplaceAll(path, string(filepath.Separator), "_")
		baseNameWithoutDots := strings.ReplaceAll(baseName, ".", "")
		archivePath := filepath.Join(outputDir, baseNameWithoutDots+".zip")

		// Map the directory (or file) to its desired path in the archive
		mapPath := baseName
		if info.IsDir() {
			// For directories, archive them entirely as a folder
			mapPath = baseName
		}
		files, err := archives.FilesFromDisk(ctx, nil, map[string]string{
			path: mapPath, // Include the directory or file in the archive
		})
		if err != nil {
			fmt.Printf("Failed to map files from disk for %s: %v\n", path, err)
			continue
		}

		// Create the output archive file
		out, err := os.Create(archivePath)
		if err != nil {
			fmt.Printf("Failed to create archive file %s: %v\n", archivePath, err)
			continue
		}
		defer out.Close()

		// Create a compressed tarball format
		format := archives.CompressedArchive{
			Compression: archives.Gz{},
			Archival:    archives.Tar{},
		}

		// Create the archive
		err = format.Archive(ctx, out, files)
		if err != nil {
			fmt.Printf("Failed to archive %s: %v\n", path, err)
			continue
		}

		fmt.Printf("Archived %s to %s\n", path, archivePath)
		archivesList = append(archivesList, archivePath)
	}

	return archivesList, nil
}
