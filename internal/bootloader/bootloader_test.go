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
