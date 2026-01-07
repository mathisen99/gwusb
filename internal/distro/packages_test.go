package distro

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// SupportedDistros lists all distros with package mappings
var SupportedDistros = []string{
	"ubuntu", "debian", "linuxmint", "fedora", "arch", "manjaro",
	"opensuse", "opensuse-tumbleweed", "opensuse-leap",
}

// PackageTestInput represents input for package mapping property tests
type PackageTestInput struct {
	Binary   string
	DistroID string
}

// Generate implements quick.Generator for PackageTestInput
func (PackageTestInput) Generate(r *rand.Rand, size int) reflect.Value {
	// Combine required and optional binaries
	allBinaries := append(RequiredBinaries, OptionalBinaries...)
	binary := allBinaries[r.Intn(len(allBinaries))]
	distroID := SupportedDistros[r.Intn(len(SupportedDistros))]

	return reflect.ValueOf(PackageTestInput{
		Binary:   binary,
		DistroID: distroID,
	})
}

// InstallCommandInput represents input for install command property tests
type InstallCommandInput struct {
	DistroID string
	Packages []string
}

// Generate implements quick.Generator for InstallCommandInput
func (InstallCommandInput) Generate(r *rand.Rand, size int) reflect.Value {
	distroID := SupportedDistros[r.Intn(len(SupportedDistros))]

	// Generate 1-5 random packages
	numPackages := r.Intn(5) + 1
	packages := make([]string, numPackages)
	samplePackages := []string{"wimtools", "p7zip-full", "dosfstools", "parted", "util-linux", "grub-pc", "ntfs-3g"}
	for i := 0; i < numPackages; i++ {
		packages[i] = samplePackages[r.Intn(len(samplePackages))]
	}

	return reflect.ValueOf(InstallCommandInput{
		DistroID: distroID,
		Packages: packages,
	})
}

// Property 5: Package Name Mapping
// For any supported distro ID and binary name, GetPackageName SHALL return
// the correct distro-specific package name as defined in the package mapping.
// **Validates: Requirements 6.2**
func TestProperty5_PackageNameMapping(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	property := func(input PackageTestInput) bool {
		result := GetPackageName(input.Binary, input.DistroID)

		// Result should never be empty for supported distros and known binaries
		if result == "" {
			t.Logf("Empty result for binary=%q, distro=%q", input.Binary, input.DistroID)
			return false
		}

		// Verify the result matches what's in the mapping
		if mapping, ok := packageMappings[input.Binary]; ok {
			if expected, ok := mapping[input.DistroID]; ok {
				if result != expected {
					t.Logf("Mismatch: got %q, want %q for binary=%q, distro=%q",
						result, expected, input.Binary, input.DistroID)
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5 failed: %v", err)
	}
}

// Property 6: Install Command Generation
// For any supported distro ID and list of package names, GetInstallCommand SHALL
// return a valid install command using the correct package manager prefix.
// **Validates: Requirements 6.3**
func TestProperty6_InstallCommandGeneration(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	property := func(input InstallCommandInput) bool {
		result := GetInstallCommand(input.DistroID, input.Packages)

		// Result should never be empty
		if result == "" {
			t.Logf("Empty result for distro=%q, packages=%v", input.DistroID, input.Packages)
			return false
		}

		// For supported distros, result should start with the correct prefix
		expectedPrefix, ok := installCommands[input.DistroID]
		if ok {
			if !strings.HasPrefix(result, expectedPrefix) {
				t.Logf("Wrong prefix: got %q, want prefix %q for distro=%q",
					result, expectedPrefix, input.DistroID)
				return false
			}

			// All packages should be present in the command (deduplicated)
			seen := make(map[string]bool)
			for _, pkg := range input.Packages {
				if !seen[pkg] {
					seen[pkg] = true
					if !strings.Contains(result, pkg) {
						t.Logf("Package %q not found in command %q", pkg, result)
						return false
					}
				}
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6 failed: %v", err)
	}
}

// TestGetPackageName_AllMappings verifies all defined mappings work correctly
func TestGetPackageName_AllMappings(t *testing.T) {
	for binary, distroMap := range packageMappings {
		for distroID, expectedPkg := range distroMap {
			t.Run(binary+"_"+distroID, func(t *testing.T) {
				result := GetPackageName(binary, distroID)
				if result != expectedPkg {
					t.Errorf("GetPackageName(%q, %q) = %q, want %q",
						binary, distroID, result, expectedPkg)
				}
			})
		}
	}
}

// TestGetPackageName_UnknownDistro tests fallback behavior
func TestGetPackageName_UnknownDistro(t *testing.T) {
	result := GetPackageName("wimlib-imagex", "unknown-distro")
	// Should return the binary name as fallback
	if result != "wimlib-imagex" {
		t.Errorf("Expected fallback to binary name, got %q", result)
	}
}

// TestGetInstallCommand_AllDistros verifies install commands for all supported distros
func TestGetInstallCommand_AllDistros(t *testing.T) {
	packages := []string{"wimtools", "p7zip-full"}

	tests := []struct {
		distroID       string
		expectedPrefix string
	}{
		{"ubuntu", "sudo apt install"},
		{"debian", "sudo apt install"},
		{"linuxmint", "sudo apt install"},
		{"fedora", "sudo dnf install"},
		{"arch", "sudo pacman -S"},
		{"manjaro", "sudo pacman -S"},
		{"opensuse", "sudo zypper install"},
		{"opensuse-tumbleweed", "sudo zypper install"},
		{"opensuse-leap", "sudo zypper install"},
	}

	for _, tt := range tests {
		t.Run(tt.distroID, func(t *testing.T) {
			result := GetInstallCommand(tt.distroID, packages)
			if !strings.HasPrefix(result, tt.expectedPrefix) {
				t.Errorf("GetInstallCommand(%q, %v) = %q, want prefix %q",
					tt.distroID, packages, result, tt.expectedPrefix)
			}
		})
	}
}

// TestGetInstallCommand_UnknownDistro tests fallback for unknown distros
func TestGetInstallCommand_UnknownDistro(t *testing.T) {
	packages := []string{"pkg1", "pkg2"}
	result := GetInstallCommand("unknown-distro", packages)

	// Should return a comment with generic instructions
	if !strings.HasPrefix(result, "#") {
		t.Errorf("Expected comment for unknown distro, got %q", result)
	}
	if !strings.Contains(result, "pkg1") || !strings.Contains(result, "pkg2") {
		t.Errorf("Expected packages in fallback message, got %q", result)
	}
}

// TestGetInstallCommand_Deduplication tests that duplicate packages are removed
func TestGetInstallCommand_Deduplication(t *testing.T) {
	packages := []string{"wimtools", "wimtools", "p7zip-full", "wimtools"}
	result := GetInstallCommand("ubuntu", packages)

	// Count occurrences of "wimtools"
	count := strings.Count(result, "wimtools")
	if count != 1 {
		t.Errorf("Expected 1 occurrence of 'wimtools', got %d in %q", count, result)
	}
}

// TestGetPackageNameWithFallback tests ID_LIKE fallback
func TestGetPackageNameWithFallback(t *testing.T) {
	tests := []struct {
		name     string
		binary   string
		info     *Info
		expected string
	}{
		{
			name:     "Direct match",
			binary:   "wimlib-imagex",
			info:     &Info{ID: "ubuntu"},
			expected: "wimtools",
		},
		{
			name:     "ID_LIKE fallback to debian",
			binary:   "wimlib-imagex",
			info:     &Info{ID: "pop", IDLike: "ubuntu debian"},
			expected: "wimtools",
		},
		{
			name:     "Unknown distro fallback",
			binary:   "wimlib-imagex",
			info:     &Info{ID: "unknown"},
			expected: "wimlib-imagex",
		},
		{
			name:     "Nil info",
			binary:   "wimlib-imagex",
			info:     nil,
			expected: "wimlib-imagex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPackageNameWithFallback(tt.binary, tt.info)
			if result != tt.expected {
				t.Errorf("GetPackageNameWithFallback(%q, %+v) = %q, want %q",
					tt.binary, tt.info, result, tt.expected)
			}
		})
	}
}

// TestGetInstallCommandWithInfo tests install command with Info struct
func TestGetInstallCommandWithInfo(t *testing.T) {
	packages := []string{"wimtools", "p7zip-full"}

	tests := []struct {
		name           string
		info           *Info
		expectedPrefix string
	}{
		{
			name:           "Ubuntu direct",
			info:           &Info{ID: "ubuntu"},
			expectedPrefix: "sudo apt install",
		},
		{
			name:           "Pop OS with ID_LIKE",
			info:           &Info{ID: "pop", IDLike: "ubuntu debian"},
			expectedPrefix: "sudo apt install",
		},
		{
			name:           "Unknown distro",
			info:           &Info{ID: "unknown"},
			expectedPrefix: "#",
		},
		{
			name:           "Nil info",
			info:           nil,
			expectedPrefix: "#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInstallCommandWithInfo(tt.info, packages)
			if !strings.HasPrefix(result, tt.expectedPrefix) {
				t.Errorf("GetInstallCommandWithInfo(%+v, %v) = %q, want prefix %q",
					tt.info, packages, result, tt.expectedPrefix)
			}
		})
	}
}
