package distro

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Info contains detected distribution information
type Info struct {
	ID             string // e.g., "ubuntu", "fedora", "arch"
	IDLike         string // e.g., "debian" for Ubuntu
	Name           string // e.g., "Ubuntu 25.10"
	Version        string // e.g., "25.10"
	PackageManager string // e.g., "apt", "dnf", "pacman", "zypper"
}

// packageManagers maps distro IDs to their package managers
var packageManagers = map[string]string{
	"ubuntu":              "apt",
	"debian":              "apt",
	"linuxmint":           "apt",
	"pop":                 "apt",
	"elementary":          "apt",
	"zorin":               "apt",
	"fedora":              "dnf",
	"rhel":                "dnf",
	"centos":              "dnf",
	"rocky":               "dnf",
	"almalinux":           "dnf",
	"arch":                "pacman",
	"manjaro":             "pacman",
	"endeavouros":         "pacman",
	"opensuse":            "zypper",
	"opensuse-tumbleweed": "zypper",
	"opensuse-leap":       "zypper",
	"suse":                "zypper",
	"gentoo":              "emerge",
	"void":                "xbps",
}

// idLikeToPackageManager maps ID_LIKE values to package managers
var idLikeToPackageManager = map[string]string{
	"debian": "apt",
	"ubuntu": "apt",
	"fedora": "dnf",
	"rhel":   "dnf",
	"arch":   "pacman",
	"suse":   "zypper",
}

// Detect reads /etc/os-release and returns distro info
func Detect() (*Info, error) {
	return DetectFromFile("/etc/os-release")
}

// DetectFromFile reads the specified os-release file and returns distro info
// This is useful for testing with custom os-release content
func DetectFromFile(path string) (*Info, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	return ParseOSRelease(file)
}

// ParseOSRelease parses os-release content from a reader
func ParseOSRelease(r *os.File) (*Info, error) {
	info := &Info{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE or KEY="VALUE"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := strings.Trim(parts[1], `"'`)

		switch key {
		case "ID":
			info.ID = value
		case "ID_LIKE":
			info.IDLike = value
		case "NAME":
			info.Name = value
		case "VERSION":
			info.Version = value
		case "VERSION_ID":
			// Use VERSION_ID as fallback if VERSION is not set
			if info.Version == "" {
				info.Version = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading os-release: %w", err)
	}

	// Determine package manager
	info.PackageManager = info.GetPackageManager()

	return info, nil
}

// GetPackageManager returns the package manager for the distro
func (i *Info) GetPackageManager() string {
	// First try direct ID match
	if pm, ok := packageManagers[i.ID]; ok {
		return pm
	}

	// Try ID_LIKE (may contain multiple space-separated values)
	if i.IDLike != "" {
		for _, like := range strings.Fields(i.IDLike) {
			if pm, ok := idLikeToPackageManager[like]; ok {
				return pm
			}
		}
	}

	// Unknown distro
	return ""
}
