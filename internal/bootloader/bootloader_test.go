package bootloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectGRUBPrefix(t *testing.T) {
	tests := []struct {
		grubCmd  string
		expected string
	}{
		{"grub-install", "grub"},
		{"grub2-install", "grub2"},
		{"/usr/bin/grub-install", "grub"},
		{"/usr/bin/grub2-install", "grub2"},
		{"/usr/sbin/grub2-install", "grub2"},
	}

	for _, test := range tests {
		result := DetectGRUBPrefix(test.grubCmd)
		if result != test.expected {
			t.Errorf("DetectGRUBPrefix(%s) = %s, expected %s", test.grubCmd, result, test.expected)
		}
	}
}

func TestWriteGRUBConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "grub_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with grub prefix
	err = WriteGRUBConfig(tmpDir, "grub")
	if err != nil {
		t.Fatalf("WriteGRUBConfig failed for grub: %v", err)
	}

	// Check that grub.cfg was created
	grubCfgPath := filepath.Join(tmpDir, "boot", "grub", "grub.cfg")
	if _, err := os.Stat(grubCfgPath); os.IsNotExist(err) {
		t.Error("grub.cfg was not created for grub prefix")
	}

	// Read and verify content
	content, err := os.ReadFile(grubCfgPath)
	if err != nil {
		t.Fatalf("Failed to read grub.cfg: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "menuentry \"Windows\"") {
		t.Error("grub.cfg does not contain expected Windows menu entry")
	}

	// Test with grub2 prefix
	tmpDir2, err := os.MkdirTemp("", "grub2_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir for grub2: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir2) }()

	err = WriteGRUBConfig(tmpDir2, "grub2")
	if err != nil {
		t.Fatalf("WriteGRUBConfig failed for grub2: %v", err)
	}

	// Check that grub.cfg was created in grub2 directory
	grub2CfgPath := filepath.Join(tmpDir2, "boot", "grub2", "grub.cfg")
	if _, err := os.Stat(grub2CfgPath); os.IsNotExist(err) {
		t.Error("grub.cfg was not created for grub2 prefix")
	}
}

func TestCheckGRUBInstallation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "grub_check_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with missing installation (should fail)
	err = CheckGRUBInstallation(tmpDir, "grub")
	if err == nil {
		t.Error("CheckGRUBInstallation should have failed for missing installation")
	}

	// Create proper GRUB installation
	err = WriteGRUBConfig(tmpDir, "grub")
	if err != nil {
		t.Fatalf("Failed to write GRUB config: %v", err)
	}

	// Test with proper installation (should pass)
	err = CheckGRUBInstallation(tmpDir, "grub")
	if err != nil {
		t.Errorf("CheckGRUBInstallation failed for valid installation: %v", err)
	}
}

func TestInstallGRUB(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "grub_install_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with non-existent grub command (should fail gracefully)
	err = InstallGRUB(tmpDir, "/dev/nonexistent", "nonexistent-grub-install")
	if err == nil {
		t.Error("InstallGRUB should have failed with non-existent command")
	}

	// Note: We can't test actual GRUB installation without root privileges
	// and without potentially affecting the system
}

func TestGetGRUBVersion(t *testing.T) {
	// Test with non-existent command (should fail gracefully)
	_, err := GetGRUBVersion("nonexistent-grub-install")
	if err == nil {
		t.Error("GetGRUBVersion should have failed with non-existent command")
	}
}

func TestInstallGRUBWithConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "grub_full_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with non-existent grub command (should fail gracefully)
	err = InstallGRUBWithConfig(tmpDir, "/dev/nonexistent", "nonexistent-grub-install")
	if err == nil {
		t.Error("InstallGRUBWithConfig should have failed with non-existent command")
	}
}

func TestIsWindows7(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "win7_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with no cversion.ini (should return false)
	isWin7, err := IsWindows7(tmpDir)
	if err != nil {
		t.Fatalf("IsWindows7 failed: %v", err)
	}
	if isWin7 {
		t.Error("Expected false for directory without cversion.ini")
	}

	// Create sources directory and cversion.ini for Windows 7
	sourcesDir := filepath.Join(tmpDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create sources dir: %v", err)
	}

	cversionPath := filepath.Join(sourcesDir, "cversion.ini")
	cversionContent := `[Version]
MinServer=7.1.7601
`
	if err := os.WriteFile(cversionPath, []byte(cversionContent), 0644); err != nil {
		t.Fatalf("Failed to create cversion.ini: %v", err)
	}

	// Test with Windows 7 cversion.ini (should return true)
	isWin7, err = IsWindows7(tmpDir)
	if err != nil {
		t.Fatalf("IsWindows7 failed: %v", err)
	}
	if !isWin7 {
		t.Error("Expected true for Windows 7 cversion.ini")
	}

	// Test with Windows 10 cversion.ini (should return false)
	cversionContent10 := `[Version]
MinServer=10.0.19041
`
	if err := os.WriteFile(cversionPath, []byte(cversionContent10), 0644); err != nil {
		t.Fatalf("Failed to update cversion.ini: %v", err)
	}

	isWin7, err = IsWindows7(tmpDir)
	if err != nil {
		t.Fatalf("IsWindows7 failed: %v", err)
	}
	if isWin7 {
		t.Error("Expected false for Windows 10 cversion.ini")
	}
}

func TestExtractBootloader(t *testing.T) {
	// Create temporary directories for testing
	srcDir, err := os.MkdirTemp("", "extract_src")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(srcDir) }()

	dstDir, err := os.MkdirTemp("", "extract_dst")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dstDir) }()

	// Test with missing install.wim/install.esd (should fail)
	err = ExtractBootloader(srcDir, dstDir)
	if err == nil {
		t.Error("ExtractBootloader should have failed with missing install files")
	}

	// Note: We can't easily test actual 7z extraction without creating
	// a proper WIM/ESD file, which would be complex for a unit test
}

func TestCheckUEFIBootloader(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "uefi_check_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test with missing bootloader (should fail)
	err = CheckUEFIBootloader(tmpDir)
	if err == nil {
		t.Error("CheckUEFIBootloader should have failed with missing bootloader")
	}

	// Create EFI boot directory and bootloader file
	efiBootDir := filepath.Join(tmpDir, "efi", "boot")
	if err := os.MkdirAll(efiBootDir, 0755); err != nil {
		t.Fatalf("Failed to create EFI boot dir: %v", err)
	}

	bootloaderPath := filepath.Join(efiBootDir, "bootx64.efi")
	if err := os.WriteFile(bootloaderPath, []byte("fake bootloader"), 0644); err != nil {
		t.Fatalf("Failed to create bootloader file: %v", err)
	}

	// Test with valid bootloader (should pass)
	err = CheckUEFIBootloader(tmpDir)
	if err != nil {
		t.Errorf("CheckUEFIBootloader failed for valid bootloader: %v", err)
	}

	// Test with empty bootloader file (should fail)
	if err := os.WriteFile(bootloaderPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty bootloader file: %v", err)
	}

	err = CheckUEFIBootloader(tmpDir)
	if err == nil {
		t.Error("CheckUEFIBootloader should have failed with empty bootloader")
	}
}

func TestApplyWindows7UEFIWorkaround(t *testing.T) {
	// Create temporary directories for testing
	srcDir, err := os.MkdirTemp("", "workaround_src")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(srcDir) }()

	dstDir, err := os.MkdirTemp("", "workaround_dst")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dstDir) }()

	// Test with non-Windows 7 (should do nothing)
	err = ApplyWindows7UEFIWorkaround(srcDir, dstDir)
	if err != nil {
		t.Errorf("ApplyWindows7UEFIWorkaround failed for non-Windows 7: %v", err)
	}

	// Note: Testing with actual Windows 7 would require creating proper
	// cversion.ini and install.wim files, which is complex for unit tests
}
