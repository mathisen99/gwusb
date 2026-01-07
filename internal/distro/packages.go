package distro

import (
	"strings"
)

// RequiredBinaries lists all required binary dependencies
var RequiredBinaries = []string{
	"wipefs",
	"parted",
	"lsblk",
	"blockdev",
	"mount",
	"umount",
	"7z",
	"mkdosfs",
	"wimlib-imagex",
}

// OptionalBinaries lists optional binary dependencies
var OptionalBinaries = []string{
	"grub-install",
	"mkntfs",
}

// packageMappings maps binary names to distro-specific package names
var packageMappings = map[string]map[string]string{
	"wimlib-imagex": {
		"ubuntu":              "wimtools",
		"debian":              "wimtools",
		"linuxmint":           "wimtools",
		"fedora":              "wimlib-utils",
		"arch":                "wimlib",
		"manjaro":             "wimlib",
		"opensuse":            "wimlib",
		"opensuse-tumbleweed": "wimlib",
		"opensuse-leap":       "wimlib",
	},
	"7z": {
		"ubuntu":              "p7zip-full",
		"debian":              "p7zip-full",
		"linuxmint":           "p7zip-full",
		"fedora":              "p7zip-plugins",
		"arch":                "p7zip",
		"manjaro":             "p7zip",
		"opensuse":            "p7zip-full",
		"opensuse-tumbleweed": "p7zip-full",
		"opensuse-leap":       "p7zip-full",
	},
	"mkdosfs": {
		"ubuntu":              "dosfstools",
		"debian":              "dosfstools",
		"linuxmint":           "dosfstools",
		"fedora":              "dosfstools",
		"arch":                "dosfstools",
		"manjaro":             "dosfstools",
		"opensuse":            "dosfstools",
		"opensuse-tumbleweed": "dosfstools",
		"opensuse-leap":       "dosfstools",
	},
	"parted": {
		"ubuntu":              "parted",
		"debian":              "parted",
		"linuxmint":           "parted",
		"fedora":              "parted",
		"arch":                "parted",
		"manjaro":             "parted",
		"opensuse":            "parted",
		"opensuse-tumbleweed": "parted",
		"opensuse-leap":       "parted",
	},
	"wipefs": {
		"ubuntu":              "util-linux",
		"debian":              "util-linux",
		"linuxmint":           "util-linux",
		"fedora":              "util-linux",
		"arch":                "util-linux",
		"manjaro":             "util-linux",
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
	},
	"lsblk": {
		"ubuntu":              "util-linux",
		"debian":              "util-linux",
		"linuxmint":           "util-linux",
		"fedora":              "util-linux",
		"arch":                "util-linux",
		"manjaro":             "util-linux",
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
	},
	"blockdev": {
		"ubuntu":              "util-linux",
		"debian":              "util-linux",
		"linuxmint":           "util-linux",
		"fedora":              "util-linux",
		"arch":                "util-linux",
		"manjaro":             "util-linux",
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
	},
	"mount": {
		"ubuntu":              "util-linux",
		"debian":              "util-linux",
		"linuxmint":           "util-linux",
		"fedora":              "util-linux",
		"arch":                "util-linux",
		"manjaro":             "util-linux",
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
	},
	"umount": {
		"ubuntu":              "util-linux",
		"debian":              "util-linux",
		"linuxmint":           "util-linux",
		"fedora":              "util-linux",
		"arch":                "util-linux",
		"manjaro":             "util-linux",
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
	},
	"grub-install": {
		"ubuntu":              "grub-pc",
		"debian":              "grub-pc",
		"linuxmint":           "grub-pc",
		"fedora":              "grub2-pc",
		"arch":                "grub",
		"manjaro":             "grub",
		"opensuse":            "grub2",
		"opensuse-tumbleweed": "grub2",
		"opensuse-leap":       "grub2",
	},
	"mkntfs": {
		"ubuntu":              "ntfs-3g",
		"debian":              "ntfs-3g",
		"linuxmint":           "ntfs-3g",
		"fedora":              "ntfs-3g",
		"arch":                "ntfs-3g",
		"manjaro":             "ntfs-3g",
		"opensuse":            "ntfs-3g",
		"opensuse-tumbleweed": "ntfs-3g",
		"opensuse-leap":       "ntfs-3g",
	},
}

// installCommands maps distro IDs to their install command prefixes
var installCommands = map[string]string{
	"ubuntu":              "sudo apt install",
	"debian":              "sudo apt install",
	"linuxmint":           "sudo apt install",
	"pop":                 "sudo apt install",
	"elementary":          "sudo apt install",
	"zorin":               "sudo apt install",
	"fedora":              "sudo dnf install",
	"rhel":                "sudo dnf install",
	"centos":              "sudo dnf install",
	"rocky":               "sudo dnf install",
	"almalinux":           "sudo dnf install",
	"arch":                "sudo pacman -S",
	"manjaro":             "sudo pacman -S",
	"endeavouros":         "sudo pacman -S",
	"opensuse":            "sudo zypper install",
	"opensuse-tumbleweed": "sudo zypper install",
	"opensuse-leap":       "sudo zypper install",
	"suse":                "sudo zypper install",
}

// idLikeToInstallCommand maps ID_LIKE values to install commands
var idLikeToInstallCommand = map[string]string{
	"debian": "sudo apt install",
	"ubuntu": "sudo apt install",
	"fedora": "sudo dnf install",
	"rhel":   "sudo dnf install",
	"arch":   "sudo pacman -S",
	"suse":   "sudo zypper install",
}

// GetPackageName returns the package name for a binary on a distro
// If the distro is not found, it tries ID_LIKE fallback, then returns the binary name
func GetPackageName(binary string, distroID string) string {
	if mapping, ok := packageMappings[binary]; ok {
		if pkg, ok := mapping[distroID]; ok {
			return pkg
		}
	}
	// Return binary name as fallback (generic)
	return binary
}

// GetPackageNameWithFallback returns the package name for a binary, using ID_LIKE as fallback
func GetPackageNameWithFallback(binary string, info *Info) string {
	if info == nil {
		return binary
	}

	// Try direct ID match first
	if mapping, ok := packageMappings[binary]; ok {
		if pkg, ok := mapping[info.ID]; ok {
			return pkg
		}

		// Try ID_LIKE fallback (may contain multiple space-separated values)
		if info.IDLike != "" {
			for _, like := range strings.Fields(info.IDLike) {
				if pkg, ok := mapping[like]; ok {
					return pkg
				}
			}
		}
	}

	// Return binary name as fallback
	return binary
}

// GetInstallCommand returns the full install command for a distro
func GetInstallCommand(distroID string, packages []string) string {
	prefix := getInstallPrefix(distroID)
	if prefix == "" {
		// Unknown distro, return generic message
		return "# Install packages using your package manager: " + strings.Join(packages, " ")
	}

	// Deduplicate packages
	seen := make(map[string]bool)
	var uniquePackages []string
	for _, pkg := range packages {
		if !seen[pkg] {
			seen[pkg] = true
			uniquePackages = append(uniquePackages, pkg)
		}
	}

	return prefix + " " + strings.Join(uniquePackages, " ")
}

// GetInstallCommandWithInfo returns the full install command using distro Info
func GetInstallCommandWithInfo(info *Info, packages []string) string {
	if info == nil {
		return "# Install packages using your package manager: " + strings.Join(packages, " ")
	}

	prefix := getInstallPrefixWithInfo(info)
	if prefix == "" {
		return "# Install packages using your package manager: " + strings.Join(packages, " ")
	}

	// Deduplicate packages
	seen := make(map[string]bool)
	var uniquePackages []string
	for _, pkg := range packages {
		if !seen[pkg] {
			seen[pkg] = true
			uniquePackages = append(uniquePackages, pkg)
		}
	}

	return prefix + " " + strings.Join(uniquePackages, " ")
}

// getInstallPrefix returns the install command prefix for a distro ID
func getInstallPrefix(distroID string) string {
	if cmd, ok := installCommands[distroID]; ok {
		return cmd
	}
	return ""
}

// getInstallPrefixWithInfo returns the install command prefix using distro Info
func getInstallPrefixWithInfo(info *Info) string {
	// Try direct ID match
	if cmd, ok := installCommands[info.ID]; ok {
		return cmd
	}

	// Try ID_LIKE fallback
	if info.IDLike != "" {
		for _, like := range strings.Fields(info.IDLike) {
			if cmd, ok := idLikeToInstallCommand[like]; ok {
				return cmd
			}
		}
	}

	return ""
}
