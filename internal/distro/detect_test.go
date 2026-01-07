package distro

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// OSReleaseData represents generated os-release content for property testing
type OSReleaseData struct {
	ID      string
	IDLike  string
	Name    string
	Version string
}

// Generate implements quick.Generator for OSReleaseData
func (OSReleaseData) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate valid distro IDs
	distroIDs := []string{"ubuntu", "debian", "fedora", "arch", "opensuse", "linuxmint", "manjaro", "pop", "centos", "rocky"}
	idLikes := []string{"", "debian", "ubuntu", "fedora", "rhel", "arch", "suse"}

	id := distroIDs[r.Intn(len(distroIDs))]
	idLike := idLikes[r.Intn(len(idLikes))]

	// Generate realistic names and versions
	names := []string{"Ubuntu", "Debian GNU/Linux", "Fedora Linux", "Arch Linux", "openSUSE Tumbleweed", "Linux Mint", "Manjaro Linux"}
	versions := []string{"25.10", "13", "42", "rolling", "16.0", "22", "24.0"}

	name := names[r.Intn(len(names))]
	version := versions[r.Intn(len(versions))]

	return reflect.ValueOf(OSReleaseData{
		ID:      id,
		IDLike:  idLike,
		Name:    name,
		Version: version,
	})
}

// ToOSReleaseContent converts OSReleaseData to os-release file content
func (d OSReleaseData) ToOSReleaseContent() string {
	var lines []string

	if d.ID != "" {
		lines = append(lines, fmt.Sprintf("ID=%s", d.ID))
	}
	if d.IDLike != "" {
		lines = append(lines, fmt.Sprintf("ID_LIKE=\"%s\"", d.IDLike))
	}
	if d.Name != "" {
		lines = append(lines, fmt.Sprintf("NAME=\"%s\"", d.Name))
	}
	if d.Version != "" {
		lines = append(lines, fmt.Sprintf("VERSION=\"%s\"", d.Version))
	}

	return strings.Join(lines, "\n")
}

// Property 4: Distro Detection from os-release
// For any valid /etc/os-release file content, the Detect function SHALL correctly
// parse and return the ID, ID_LIKE, NAME, and VERSION fields.
// **Validates: Requirements 6.1**
func TestProperty4_DistroDetectionFromOSRelease(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	property := func(data OSReleaseData) bool {
		// Create a temporary file with the os-release content
		tmpFile, err := os.CreateTemp("", "os-release-test-*")
		if err != nil {
			t.Logf("Failed to create temp file: %v", err)
			return false
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		content := data.ToOSReleaseContent()
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Logf("Failed to write temp file: %v", err)
			_ = tmpFile.Close()
			return false
		}
		_ = tmpFile.Close()

		// Parse the file
		info, err := DetectFromFile(tmpFile.Name())
		if err != nil {
			t.Logf("Failed to parse os-release: %v", err)
			return false
		}

		// Verify all fields are correctly parsed
		if info.ID != data.ID {
			t.Logf("ID mismatch: got %q, want %q", info.ID, data.ID)
			return false
		}
		if info.IDLike != data.IDLike {
			t.Logf("IDLike mismatch: got %q, want %q", info.IDLike, data.IDLike)
			return false
		}
		if info.Name != data.Name {
			t.Logf("Name mismatch: got %q, want %q", info.Name, data.Name)
			return false
		}
		if info.Version != data.Version {
			t.Logf("Version mismatch: got %q, want %q", info.Version, data.Version)
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 4 failed: %v", err)
	}
}

// TestDetectFromFile_RealFormats tests parsing of real-world os-release formats
func TestDetectFromFile_RealFormats(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Info
	}{
		{
			name: "Ubuntu",
			content: `NAME="Ubuntu"
VERSION="25.10 (Questing)"
ID=ubuntu
ID_LIKE=debian
VERSION_ID="25.10"`,
			expected: Info{
				ID:             "ubuntu",
				IDLike:         "debian",
				Name:           "Ubuntu",
				Version:        "25.10 (Questing)",
				PackageManager: "apt",
			},
		},
		{
			name: "Fedora",
			content: `NAME="Fedora Linux"
VERSION="42 (Workstation Edition)"
ID=fedora
VERSION_ID=42`,
			expected: Info{
				ID:             "fedora",
				IDLike:         "",
				Name:           "Fedora Linux",
				Version:        "42 (Workstation Edition)",
				PackageManager: "dnf",
			},
		},
		{
			name: "Arch",
			content: `NAME="Arch Linux"
ID=arch
BUILD_ID=rolling`,
			expected: Info{
				ID:             "arch",
				IDLike:         "",
				Name:           "Arch Linux",
				Version:        "",
				PackageManager: "pacman",
			},
		},
		{
			name: "openSUSE Tumbleweed",
			content: `NAME="openSUSE Tumbleweed"
ID="opensuse-tumbleweed"
ID_LIKE="opensuse suse"
VERSION_ID="20260101"`,
			expected: Info{
				ID:             "opensuse-tumbleweed",
				IDLike:         "opensuse suse",
				Name:           "openSUSE Tumbleweed",
				Version:        "20260101",
				PackageManager: "zypper",
			},
		},
		{
			name: "Linux Mint",
			content: `NAME="Linux Mint"
VERSION="22 (Wilma)"
ID=linuxmint
ID_LIKE="ubuntu debian"
VERSION_ID="22"`,
			expected: Info{
				ID:             "linuxmint",
				IDLike:         "ubuntu debian",
				Name:           "Linux Mint",
				Version:        "22 (Wilma)",
				PackageManager: "apt",
			},
		},
		{
			name: "Debian",
			content: `NAME="Debian GNU/Linux"
VERSION="13 (trixie)"
ID=debian
VERSION_ID="13"`,
			expected: Info{
				ID:             "debian",
				IDLike:         "",
				Name:           "Debian GNU/Linux",
				Version:        "13 (trixie)",
				PackageManager: "apt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "os-release-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			_ = tmpFile.Close()

			info, err := DetectFromFile(tmpFile.Name())
			if err != nil {
				t.Fatalf("DetectFromFile failed: %v", err)
			}

			if info.ID != tt.expected.ID {
				t.Errorf("ID: got %q, want %q", info.ID, tt.expected.ID)
			}
			if info.IDLike != tt.expected.IDLike {
				t.Errorf("IDLike: got %q, want %q", info.IDLike, tt.expected.IDLike)
			}
			if info.Name != tt.expected.Name {
				t.Errorf("Name: got %q, want %q", info.Name, tt.expected.Name)
			}
			if info.Version != tt.expected.Version {
				t.Errorf("Version: got %q, want %q", info.Version, tt.expected.Version)
			}
			if info.PackageManager != tt.expected.PackageManager {
				t.Errorf("PackageManager: got %q, want %q", info.PackageManager, tt.expected.PackageManager)
			}
		})
	}
}

// TestGetPackageManager tests package manager detection
func TestGetPackageManager(t *testing.T) {
	tests := []struct {
		name     string
		info     Info
		expected string
	}{
		{"Ubuntu direct", Info{ID: "ubuntu"}, "apt"},
		{"Fedora direct", Info{ID: "fedora"}, "dnf"},
		{"Arch direct", Info{ID: "arch"}, "pacman"},
		{"openSUSE direct", Info{ID: "opensuse"}, "zypper"},
		{"Unknown with debian-like", Info{ID: "unknown", IDLike: "debian"}, "apt"},
		{"Unknown with ubuntu-like", Info{ID: "unknown", IDLike: "ubuntu"}, "apt"},
		{"Unknown with fedora-like", Info{ID: "unknown", IDLike: "fedora"}, "dnf"},
		{"Unknown with multiple ID_LIKE", Info{ID: "unknown", IDLike: "ubuntu debian"}, "apt"},
		{"Completely unknown", Info{ID: "unknown"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.GetPackageManager()
			if got != tt.expected {
				t.Errorf("GetPackageManager() = %q, want %q", got, tt.expected)
			}
		})
	}
}
