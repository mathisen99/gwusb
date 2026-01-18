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
// Supported distros: Ubuntu, Debian, Linux Mint, Pop!_OS, Elementary, Zorin,
// Fedora, RHEL, CentOS, Rocky, AlmaLinux, Arch, Manjaro, EndeavourOS,
// openSUSE (Tumbleweed, Leap), Void, Gentoo
var packageMappings = map[string]map[string]string{
	"wimlib-imagex": {
		// Debian-based
		"ubuntu":     "wimtools",
		"debian":     "wimtools",
		"linuxmint":  "wimtools",
		"pop":        "wimtools",
		"elementary": "wimtools",
		"zorin":      "wimtools",
		// RHEL-based
		"fedora":    "wimlib-utils",
		"rhel":      "wimlib-utils",
		"centos":    "wimlib-utils",
		"rocky":     "wimlib-utils",
		"almalinux": "wimlib-utils",
		// Arch-based
		"arch":        "wimlib",
		"manjaro":     "wimlib",
		"endeavouros": "wimlib",
		// SUSE-based
		"opensuse":            "wimtools",
		"opensuse-tumbleweed": "wimtools",
		"opensuse-leap":       "wimtools",
		"suse":                "wimtools",
		// Other
		"void":   "wimlib",
		"gentoo": "app-arch/wimlib",
	},
	"7z": {
		// Debian-based
		"ubuntu":     "p7zip-full",
		"debian":     "p7zip-full",
		"linuxmint":  "p7zip-full",
		"pop":        "p7zip-full",
		"elementary": "p7zip-full",
		"zorin":      "p7zip-full",
		// RHEL-based
		"fedora":    "p7zip-plugins",
		"rhel":      "p7zip-plugins",
		"centos":    "p7zip-plugins",
		"rocky":     "p7zip-plugins",
		"almalinux": "p7zip-plugins",
		// Arch-based
		"arch":        "p7zip",
		"manjaro":     "p7zip",
		"endeavouros": "p7zip",
		// SUSE-based
		"opensuse":            "p7zip",
		"opensuse-tumbleweed": "p7zip",
		"opensuse-leap":       "p7zip",
		"suse":                "p7zip",
		// Other
		"void":   "p7zip",
		"gentoo": "app-arch/p7zip",
	},
	"mkdosfs": {
		// Debian-based
		"ubuntu":     "dosfstools",
		"debian":     "dosfstools",
		"linuxmint":  "dosfstools",
		"pop":        "dosfstools",
		"elementary": "dosfstools",
		"zorin":      "dosfstools",
		// RHEL-based
		"fedora":    "dosfstools",
		"rhel":      "dosfstools",
		"centos":    "dosfstools",
		"rocky":     "dosfstools",
		"almalinux": "dosfstools",
		// Arch-based
		"arch":        "dosfstools",
		"manjaro":     "dosfstools",
		"endeavouros": "dosfstools",
		// SUSE-based
		"opensuse":            "dosfstools",
		"opensuse-tumbleweed": "dosfstools",
		"opensuse-leap":       "dosfstools",
		"suse":                "dosfstools",
		// Other
		"void":   "dosfstools",
		"gentoo": "sys-fs/dosfstools",
	},
	"parted": {
		// Debian-based
		"ubuntu":     "parted",
		"debian":     "parted",
		"linuxmint":  "parted",
		"pop":        "parted",
		"elementary": "parted",
		"zorin":      "parted",
		// RHEL-based
		"fedora":    "parted",
		"rhel":      "parted",
		"centos":    "parted",
		"rocky":     "parted",
		"almalinux": "parted",
		// Arch-based
		"arch":        "parted",
		"manjaro":     "parted",
		"endeavouros": "parted",
		// SUSE-based
		"opensuse":            "parted",
		"opensuse-tumbleweed": "parted",
		"opensuse-leap":       "parted",
		"suse":                "parted",
		// Other
		"void":   "parted",
		"gentoo": "sys-block/parted",
	},
	"wipefs": {
		// Debian-based
		"ubuntu":     "util-linux",
		"debian":     "util-linux",
		"linuxmint":  "util-linux",
		"pop":        "util-linux",
		"elementary": "util-linux",
		"zorin":      "util-linux",
		// RHEL-based
		"fedora":    "util-linux",
		"rhel":      "util-linux",
		"centos":    "util-linux",
		"rocky":     "util-linux",
		"almalinux": "util-linux",
		// Arch-based
		"arch":        "util-linux",
		"manjaro":     "util-linux",
		"endeavouros": "util-linux",
		// SUSE-based
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
		"suse":                "util-linux",
		// Other
		"void":   "util-linux",
		"gentoo": "sys-apps/util-linux",
	},
	"lsblk": {
		// Debian-based
		"ubuntu":     "util-linux",
		"debian":     "util-linux",
		"linuxmint":  "util-linux",
		"pop":        "util-linux",
		"elementary": "util-linux",
		"zorin":      "util-linux",
		// RHEL-based
		"fedora":    "util-linux",
		"rhel":      "util-linux",
		"centos":    "util-linux",
		"rocky":     "util-linux",
		"almalinux": "util-linux",
		// Arch-based
		"arch":        "util-linux",
		"manjaro":     "util-linux",
		"endeavouros": "util-linux",
		// SUSE-based
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
		"suse":                "util-linux",
		// Other
		"void":   "util-linux",
		"gentoo": "sys-apps/util-linux",
	},
	"blockdev": {
		// Debian-based
		"ubuntu":     "util-linux",
		"debian":     "util-linux",
		"linuxmint":  "util-linux",
		"pop":        "util-linux",
		"elementary": "util-linux",
		"zorin":      "util-linux",
		// RHEL-based
		"fedora":    "util-linux",
		"rhel":      "util-linux",
		"centos":    "util-linux",
		"rocky":     "util-linux",
		"almalinux": "util-linux",
		// Arch-based
		"arch":        "util-linux",
		"manjaro":     "util-linux",
		"endeavouros": "util-linux",
		// SUSE-based
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
		"suse":                "util-linux",
		// Other
		"void":   "util-linux",
		"gentoo": "sys-apps/util-linux",
	},
	"mount": {
		// Debian-based
		"ubuntu":     "util-linux",
		"debian":     "util-linux",
		"linuxmint":  "util-linux",
		"pop":        "util-linux",
		"elementary": "util-linux",
		"zorin":      "util-linux",
		// RHEL-based
		"fedora":    "util-linux",
		"rhel":      "util-linux",
		"centos":    "util-linux",
		"rocky":     "util-linux",
		"almalinux": "util-linux",
		// Arch-based
		"arch":        "util-linux",
		"manjaro":     "util-linux",
		"endeavouros": "util-linux",
		// SUSE-based
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
		"suse":                "util-linux",
		// Other
		"void":   "util-linux",
		"gentoo": "sys-apps/util-linux",
	},
	"umount": {
		// Debian-based
		"ubuntu":     "util-linux",
		"debian":     "util-linux",
		"linuxmint":  "util-linux",
		"pop":        "util-linux",
		"elementary": "util-linux",
		"zorin":      "util-linux",
		// RHEL-based
		"fedora":    "util-linux",
		"rhel":      "util-linux",
		"centos":    "util-linux",
		"rocky":     "util-linux",
		"almalinux": "util-linux",
		// Arch-based
		"arch":        "util-linux",
		"manjaro":     "util-linux",
		"endeavouros": "util-linux",
		// SUSE-based
		"opensuse":            "util-linux",
		"opensuse-tumbleweed": "util-linux",
		"opensuse-leap":       "util-linux",
		"suse":                "util-linux",
		// Other
		"void":   "util-linux",
		"gentoo": "sys-apps/util-linux",
	},
	"grub-install": {
		// Debian-based (grub-pc for BIOS, grub-efi-amd64 for UEFI)
		"ubuntu":     "grub-pc",
		"debian":     "grub-pc",
		"linuxmint":  "grub-pc",
		"pop":        "grub-pc",
		"elementary": "grub-pc",
		"zorin":      "grub-pc",
		// RHEL-based
		"fedora":    "grub2-tools",
		"rhel":      "grub2-tools",
		"centos":    "grub2-tools",
		"rocky":     "grub2-tools",
		"almalinux": "grub2-tools",
		// Arch-based
		"arch":        "grub",
		"manjaro":     "grub",
		"endeavouros": "grub",
		// SUSE-based
		"opensuse":            "grub2",
		"opensuse-tumbleweed": "grub2",
		"opensuse-leap":       "grub2",
		"suse":                "grub2",
		// Other
		"void":   "grub",
		"gentoo": "sys-boot/grub",
	},
	"mkntfs": {
		// Debian-based
		"ubuntu":     "ntfs-3g",
		"debian":     "ntfs-3g",
		"linuxmint":  "ntfs-3g",
		"pop":        "ntfs-3g",
		"elementary": "ntfs-3g",
		"zorin":      "ntfs-3g",
		// RHEL-based
		"fedora":    "ntfs-3g",
		"rhel":      "ntfs-3g",
		"centos":    "ntfs-3g",
		"rocky":     "ntfs-3g",
		"almalinux": "ntfs-3g",
		// Arch-based
		"arch":        "ntfs-3g",
		"manjaro":     "ntfs-3g",
		"endeavouros": "ntfs-3g",
		// SUSE-based
		"opensuse":            "ntfs-3g",
		"opensuse-tumbleweed": "ntfs-3g",
		"opensuse-leap":       "ntfs-3g",
		"suse":                "ntfs-3g",
		// Other
		"void":   "ntfs-3g",
		"gentoo": "sys-fs/ntfs3g",
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
	"void":                "sudo xbps-install -S",
	"gentoo":              "sudo emerge",
}

// idLikeToInstallCommand maps ID_LIKE values to install commands
var idLikeToInstallCommand = map[string]string{
	"debian": "sudo apt install",
	"ubuntu": "sudo apt install",
	"fedora": "sudo dnf install",
	"rhel":   "sudo dnf install",
	"arch":   "sudo pacman -S",
	"suse":   "sudo zypper install",
	"void":   "sudo xbps-install -S",
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
