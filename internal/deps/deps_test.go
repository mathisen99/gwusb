package deps

import (
	"testing"
)

func TestCheckDependencies(t *testing.T) {
	deps, err := CheckDependencies()

	// On most systems, some dependencies might be missing
	// This test just ensures the function doesn't panic
	if err != nil {
		t.Logf("Missing dependencies (expected on some systems): %v", err)
		return
	}

	// If no error, verify we got valid paths
	if deps.Wipefs == "" {
		t.Error("Expected wipefs path to be set")
	}
	if deps.Parted == "" {
		t.Error("Expected parted path to be set")
	}
	if deps.MkFat == "" {
		t.Error("Expected mkfat path to be set")
	}
}

func TestCheckDependenciesWithDistro(t *testing.T) {
	result := CheckDependenciesWithDistro()

	// Result should never be nil
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Deps should never be nil
	if result.Deps == nil {
		t.Fatal("Expected non-nil Deps")
	}

	// Missing should be initialized (may be empty)
	if result.Missing == nil {
		t.Error("Expected Missing to be initialized")
	}

	// DistroInfo may be nil if /etc/os-release doesn't exist
	// This is acceptable behavior
	t.Logf("Distro info: %+v", result.DistroInfo)
	t.Logf("Missing dependencies: %d", len(result.Missing))

	// Verify missing deps have proper structure
	for _, m := range result.Missing {
		if m.Binary == "" {
			t.Error("Missing dep should have Binary set")
		}
		if m.PackageName == "" {
			t.Error("Missing dep should have PackageName set")
		}
	}
}

func TestBinaryExists(t *testing.T) {
	// Test with a binary that should exist on all Linux systems
	if !BinaryExists("ls") {
		t.Error("Expected 'ls' to exist")
	}

	// Test with a binary that should not exist
	if BinaryExists("nonexistent-binary-12345") {
		t.Error("Expected nonexistent binary to not exist")
	}
}

func TestGetInstallCommand(t *testing.T) {
	// Test with empty missing list
	cmd := GetInstallCommand([]MissingDep{}, nil)
	if cmd != "" {
		t.Errorf("Expected empty string for no missing deps, got: %s", cmd)
	}

	// Test with missing deps and nil distro info
	missing := []MissingDep{
		{Binary: "wimlib-imagex", PackageName: "wimtools", Required: true},
		{Binary: "7z", PackageName: "p7zip-full", Required: true},
	}
	cmd = GetInstallCommand(missing, nil)
	if cmd == "" {
		t.Error("Expected non-empty install command")
	}
	t.Logf("Install command (nil distro): %s", cmd)
}

func TestGetRequiredMissing(t *testing.T) {
	missing := []MissingDep{
		{Binary: "wipefs", PackageName: "util-linux", Required: true},
		{Binary: "mkntfs", PackageName: "ntfs-3g", Required: false},
		{Binary: "7z", PackageName: "p7zip-full", Required: true},
	}

	required := GetRequiredMissing(missing)
	if len(required) != 2 {
		t.Errorf("Expected 2 required deps, got %d", len(required))
	}

	for _, m := range required {
		if !m.Required {
			t.Errorf("Expected all returned deps to be required, got: %+v", m)
		}
	}
}

func TestGetOptionalMissing(t *testing.T) {
	missing := []MissingDep{
		{Binary: "wipefs", PackageName: "util-linux", Required: true},
		{Binary: "mkntfs", PackageName: "ntfs-3g", Required: false},
		{Binary: "grub-install", PackageName: "grub-pc", Required: false},
	}

	optional := GetOptionalMissing(missing)
	if len(optional) != 2 {
		t.Errorf("Expected 2 optional deps, got %d", len(optional))
	}

	for _, m := range optional {
		if m.Required {
			t.Errorf("Expected all returned deps to be optional, got: %+v", m)
		}
	}
}

// Property 3: Dependency Binary Detection
// For any binary name in the required list, the dependency checker SHALL correctly
// identify whether the binary exists in PATH.
// **Validates: Requirements 1.3**
func TestProperty3_DependencyBinaryDetection(t *testing.T) {
	// This property test verifies that BinaryExists correctly identifies
	// whether a binary exists in PATH for any given binary name.

	// We test with a mix of:
	// 1. Binaries that should exist on most Linux systems
	// 2. Binaries that should not exist (random strings)
	// 3. Required binaries from the distro package

	// Test 1: Common system binaries that should exist
	existingBinaries := []string{"ls", "cat", "echo", "sh", "test"}
	for _, binary := range existingBinaries {
		exists := BinaryExists(binary)
		if !exists {
			t.Errorf("Expected binary %q to exist in PATH", binary)
		}
	}

	// Test 2: Random non-existent binaries
	nonExistentBinaries := []string{
		"nonexistent-binary-xyz-12345",
		"fake-tool-abc-67890",
		"imaginary-command-qwerty",
		"this-binary-does-not-exist-anywhere",
		"random-gibberish-tool-99999",
	}
	for _, binary := range nonExistentBinaries {
		exists := BinaryExists(binary)
		if exists {
			t.Errorf("Expected binary %q to NOT exist in PATH", binary)
		}
	}

	// Test 3: Verify consistency - calling BinaryExists multiple times
	// should return the same result (idempotent)
	testBinary := "ls"
	result1 := BinaryExists(testBinary)
	result2 := BinaryExists(testBinary)
	result3 := BinaryExists(testBinary)
	if result1 != result2 || result2 != result3 {
		t.Errorf("BinaryExists should be idempotent: got %v, %v, %v", result1, result2, result3)
	}

	// Test 4: Verify that CheckDependenciesWithDistro correctly identifies
	// missing vs found binaries
	result := CheckDependenciesWithDistro()

	// For each found binary, BinaryExists should return true
	if result.Deps.Wipefs != "" && !BinaryExists("wipefs") {
		t.Error("Wipefs path set but BinaryExists returns false")
	}
	if result.Deps.Parted != "" && !BinaryExists("parted") {
		t.Error("Parted path set but BinaryExists returns false")
	}
	if result.Deps.Lsblk != "" && !BinaryExists("lsblk") {
		t.Error("Lsblk path set but BinaryExists returns false")
	}
	if result.Deps.SevenZip != "" && !BinaryExists("7z") {
		t.Error("7z path set but BinaryExists returns false")
	}
	if result.Deps.WimlibSplit != "" && !BinaryExists("wimlib-imagex") {
		t.Error("wimlib-imagex path set but BinaryExists returns false")
	}

	// For each missing binary, BinaryExists should return false
	for _, missing := range result.Missing {
		// Note: Some binaries have aliases (mkdosfs/mkfs.vfat/mkfs.fat, grub-install/grub2-install)
		// so we only check the primary binary name
		switch missing.Binary {
		case "mkdosfs":
			// mkdosfs might be available as mkfs.vfat or mkfs.fat
			if BinaryExists("mkdosfs") || BinaryExists("mkfs.vfat") || BinaryExists("mkfs.fat") {
				t.Errorf("mkdosfs reported as missing but one of its aliases exists")
			}
		case "grub-install":
			// grub-install might be available as grub2-install
			if BinaryExists("grub-install") || BinaryExists("grub2-install") {
				t.Errorf("grub-install reported as missing but one of its aliases exists")
			}
		default:
			if BinaryExists(missing.Binary) {
				t.Errorf("Binary %q reported as missing but BinaryExists returns true", missing.Binary)
			}
		}
	}
}
