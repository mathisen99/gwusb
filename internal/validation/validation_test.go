package validation

import (
	"os"
	"testing"
)

func TestValidateSource(t *testing.T) {
	// Test with non-existent file
	err := ValidateSource("/nonexistent/file.iso")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with current directory (should fail - not a file or block device)
	err = ValidateSource(".")
	if err == nil {
		t.Error("Expected error for directory")
	}

	// Test with a regular file (create temp file)
	tmpFile, err := os.CreateTemp("", "test*.iso")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	err = ValidateSource(tmpFile.Name())
	if err != nil {
		t.Errorf("Expected no error for regular file, got: %v", err)
	}
}

func TestValidateTarget(t *testing.T) {
	// Test with non-existent device
	err := ValidateTarget("/dev/nonexistent", "device")
	if err == nil {
		t.Error("Expected error for non-existent device")
	}

	// Test with invalid mode
	err = ValidateTarget("/dev/null", "invalid")
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
}

func TestIsWholeDevice(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/dev/sda", true},
		{"/dev/sdb", true},
		{"/dev/sda1", false},
		{"/dev/sdb2", false},
		{"/dev/nvme0n1", true},
		{"/dev/nvme1n1", true},
		{"/dev/nvme0n1p1", false},
		{"/dev/nvme1n1p2", false},
		{"/dev/mmcblk0", true},
		{"/dev/mmcblk1", true},
		{"/dev/mmcblk0p1", false},
		{"/dev/mmcblk1p2", false},
	}

	for _, test := range tests {
		result := isWholeDevice(test.path)
		if result != test.expected {
			t.Errorf("isWholeDevice(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}
