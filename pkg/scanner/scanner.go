package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectShell returns the user's default shell (bash, zsh, etc.)
func DetectShell() string {
	// Get the user's shell from the SHELL environment variable
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Default to bash if SHELL is not set (though this shouldn't happen often)
		shell = "/bin/bash"
	}
	return shell
}

// GetHistoryFile returns the path to the correct history file for the detected shell
func GetHistoryFile() (string, error) {
	shell := DetectShell()
	var historyFile string

	// Determine which shell is in use and return the appropriate history file path
	switch {
	case strings.Contains(shell, "bash"):
		historyFile = filepath.Join(os.Getenv("HOME"), ".bash_history")
	case strings.Contains(shell, "zsh"):
		historyFile = filepath.Join(os.Getenv("HOME"), ".zsh_history")
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}

	// Check if the history file exists
	if _, err := os.Stat(historyFile); err == nil {
		return historyFile, nil
	}
	return "", fmt.Errorf("history file not found for shell: %s", shell)
}

// ScanForSensitiveData scans directories and returns sensitive files
func ScanForSensitiveData() ([]string, error) {
	var sensitiveFiles []string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// List of common sensitive files and directories
	sensitivePaths := []string{
		homeDir + "/.ssh/",
		homeDir + "/.gnupg/",
		homeDir + "/.aws/",
		homeDir + "/.docker",
		homeDir + "/.kube",
		"/etc/passwd",
		"/etc/shadow",
	}

	// Add the correct history file based on the shell
	historyFile, err := GetHistoryFile()
	if err != nil {
		return nil, err
	}
	sensitivePaths = append(sensitivePaths, historyFile)

	// Check if each file exists and add it to the list
	for _, path := range sensitivePaths {
		if _, err := os.Stat(path); err == nil {
			sensitiveFiles = append(sensitiveFiles, path)
		}
	}

	return sensitiveFiles, nil
}
