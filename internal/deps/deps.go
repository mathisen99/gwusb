package deps

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mathisen/woeusb-go/internal/distro"
)

// MissingDep represents a missing dependency with distro-specific info
type MissingDep struct {
	Binary      string // e.g., "wimlib-imagex"
	PackageName string // distro-specific package name
	Required    bool   // true if required, false if optional
}

// Dependencies holds paths to all required external tools
type Dependencies struct {
	Wipefs      string
	Parted      string
	Lsblk       string
	Blockdev    string
	Mount       string
	Umount      string
	SevenZip    string
	MkFat       string
	MkNTFS      string
	GrubCmd     string
	WimlibSplit string // wimlib-imagex for splitting WIM files
}

// CheckResult contains the result of dependency checking
type CheckResult struct {
	Deps       *Dependencies
	Missing    []MissingDep
	DistroInfo *distro.Info
}

// CheckDependencies verifies all required tools are installed
// Returns Dependencies struct and error if required dependencies are missing
func CheckDependencies() (*Dependencies, error) {
	result := CheckDependenciesWithDistro()

	if len(result.Missing) > 0 {
		var requiredMissing []string
		for _, m := range result.Missing {
			if m.Required {
				requiredMissing = append(requiredMissing, m.Binary)
			}
		}
		if len(requiredMissing) > 0 {
			return nil, fmt.Errorf("missing required dependencies: %s", strings.Join(requiredMissing, ", "))
		}
	}

	return result.Deps, nil
}

// CheckDependenciesWithDistro verifies all required tools and returns detailed info
// including distro-specific package names for missing dependencies
func CheckDependenciesWithDistro() *CheckResult {
	result := &CheckResult{
		Deps:    &Dependencies{},
		Missing: []MissingDep{},
	}

	// Detect distro for package name mapping
	distroInfo, err := distro.Detect()
	if err != nil {
		// Continue without distro info - will use generic package names
		distroInfo = nil
	}
	result.DistroInfo = distroInfo

	// Check required tools
	requiredTools := []struct {
		binary string
		field  *string
	}{
		{"wipefs", &result.Deps.Wipefs},
		{"parted", &result.Deps.Parted},
		{"lsblk", &result.Deps.Lsblk},
		{"blockdev", &result.Deps.Blockdev},
		{"mount", &result.Deps.Mount},
		{"umount", &result.Deps.Umount},
		{"7z", &result.Deps.SevenZip},
	}

	for _, tool := range requiredTools {
		if path, err := exec.LookPath(tool.binary); err != nil {
			result.Missing = append(result.Missing, MissingDep{
				Binary:      tool.binary,
				PackageName: distro.GetPackageNameWithFallback(tool.binary, distroInfo),
				Required:    true,
			})
		} else {
			*tool.field = path
		}
	}

	// Find mkdosfs/mkfs.vfat/mkfs.fat (return first found)
	fatCmds := []string{"mkdosfs", "mkfs.vfat", "mkfs.fat"}
	fatFound := false
	for _, cmd := range fatCmds {
		if path, err := exec.LookPath(cmd); err == nil {
			result.Deps.MkFat = path
			fatFound = true
			break
		}
	}
	if !fatFound {
		result.Missing = append(result.Missing, MissingDep{
			Binary:      "mkdosfs",
			PackageName: distro.GetPackageNameWithFallback("mkdosfs", distroInfo),
			Required:    true,
		})
	}

	// Find wimlib-imagex (required for Win10/11)
	if path, err := exec.LookPath("wimlib-imagex"); err != nil {
		result.Missing = append(result.Missing, MissingDep{
			Binary:      "wimlib-imagex",
			PackageName: distro.GetPackageNameWithFallback("wimlib-imagex", distroInfo),
			Required:    true,
		})
	} else {
		result.Deps.WimlibSplit = path
	}

	// Find mkntfs (optional - only needed if user forces NTFS)
	if path, err := exec.LookPath("mkntfs"); err == nil {
		result.Deps.MkNTFS = path
	} else {
		result.Missing = append(result.Missing, MissingDep{
			Binary:      "mkntfs",
			PackageName: distro.GetPackageNameWithFallback("mkntfs", distroInfo),
			Required:    false,
		})
	}

	// Find grub-install or grub2-install (optional for UEFI-only systems)
	grubCmds := []string{"grub-install", "grub2-install"}
	grubFound := false
	for _, cmd := range grubCmds {
		if path, err := exec.LookPath(cmd); err == nil {
			result.Deps.GrubCmd = path
			grubFound = true
			break
		}
	}
	if !grubFound {
		result.Missing = append(result.Missing, MissingDep{
			Binary:      "grub-install",
			PackageName: distro.GetPackageNameWithFallback("grub-install", distroInfo),
			Required:    false,
		})
	}

	return result
}

// BinaryExists checks if a binary exists in PATH
func BinaryExists(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

// GetInstallCommand returns the full install command for missing dependencies
func GetInstallCommand(missing []MissingDep, distroInfo *distro.Info) string {
	if len(missing) == 0 {
		return ""
	}

	var packages []string
	for _, m := range missing {
		packages = append(packages, m.PackageName)
	}

	return distro.GetInstallCommandWithInfo(distroInfo, packages)
}

// GetRequiredMissing returns only the required missing dependencies
func GetRequiredMissing(missing []MissingDep) []MissingDep {
	var required []MissingDep
	for _, m := range missing {
		if m.Required {
			required = append(required, m)
		}
	}
	return required
}

// GetOptionalMissing returns only the optional missing dependencies
func GetOptionalMissing(missing []MissingDep) []MissingDep {
	var optional []MissingDep
	for _, m := range missing {
		if !m.Required {
			optional = append(optional, m)
		}
	}
	return optional
}
