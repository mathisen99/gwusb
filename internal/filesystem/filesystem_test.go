package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckFAT32Limit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "fat32_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a small file (within FAT32 limits)
	smallFile := filepath.Join(tmpDir, "small.txt")
	if err := os.WriteFile(smallFile, []byte("small content"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// Test with only small files
	hasOversized, oversizedFiles, err := CheckFAT32Limit(tmpDir)
	if err != nil {
		t.Fatalf("CheckFAT32Limit failed: %v", err)
	}

	if hasOversized {
		t.Error("Expected no oversized files for small content")
	}

	if len(oversizedFiles) != 0 {
		t.Errorf("Expected no oversized files, got: %v", oversizedFiles)
	}

	// Test with non-existent directory
	_, _, err = CheckFAT32Limit("/nonexistent/path")
	if err != nil {
		t.Logf("Got expected error for non-existent directory: %v", err)
	}
}

func TestGetLargestFileSize(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "largest_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files of different sizes
	smallFile := filepath.Join(tmpDir, "small.txt")
	if err := os.WriteFile(smallFile, []byte("small"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := make([]byte, 1000)
	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	maxSize, maxFile, err := GetLargestFileSize(tmpDir)
	if err != nil {
		t.Fatalf("GetLargestFileSize failed: %v", err)
	}

	if maxSize != 1000 {
		t.Errorf("Expected largest file size 1000, got: %d", maxSize)
	}

	if maxFile != "large.txt" {
		t.Errorf("Expected largest file 'large.txt', got: %s", maxFile)
	}
}

func TestFormatSizeHuman(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{FAT32MaxFileSize, "4.0 GB"},
	}

	for _, test := range tests {
		result := FormatSizeHuman(test.bytes)
		if result != test.expected {
			t.Errorf("FormatSizeHuman(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestSuggestFilesystem(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "suggest_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a small file
	smallFile := filepath.Join(tmpDir, "small.txt")
	if err := os.WriteFile(smallFile, []byte("small content"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// Test with small files (should suggest FAT32)
	fs, reason, err := SuggestFilesystem(tmpDir)
	if err != nil {
		t.Fatalf("SuggestFilesystem failed: %v", err)
	}

	if fs != "FAT32" {
		t.Errorf("Expected FAT32 for small files, got: %s", fs)
	}

	if reason != "All files are within FAT32 limits" {
		t.Errorf("Unexpected reason: %s", reason)
	}
}

func TestValidateFilesystemChoice(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "validate_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a small file
	smallFile := filepath.Join(tmpDir, "small.txt")
	if err := os.WriteFile(smallFile, []byte("small content"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// Test FAT32 with small files (should pass)
	err = ValidateFilesystemChoice(tmpDir, "FAT32")
	if err != nil {
		t.Errorf("Expected no error for FAT32 with small files, got: %v", err)
	}

	// Test NTFS (should always pass)
	err = ValidateFilesystemChoice(tmpDir, "NTFS")
	if err != nil {
		t.Errorf("Expected no error for NTFS, got: %v", err)
	}
}

func TestFormatFAT32(t *testing.T) {
	// Test with non-existent partition (should fail gracefully)
	err := FormatFAT32("/dev/nonexistent")
	if err == nil {
		t.Error("Expected error when formatting non-existent partition")
	}
}

func TestFormatNTFS(t *testing.T) {
	// Test with non-existent partition (should fail gracefully)
	err := FormatNTFS("/dev/nonexistent", "TestLabel")
	if err == nil {
		t.Error("Expected error when formatting non-existent partition")
	}

	// Test without label
	err = FormatNTFS("/dev/nonexistent", "")
	if err == nil {
		t.Error("Expected error when formatting non-existent partition")
	}
}

func TestFormatPartition(t *testing.T) {
	// Test with non-existent partition (should fail gracefully)
	err := FormatPartition("/dev/nonexistent", "FAT32", "TestLabel")
	if err == nil {
		t.Error("Expected error when formatting non-existent partition")
	}

	// Test with unsupported filesystem
	err = FormatPartition("/dev/nonexistent", "UNSUPPORTED", "TestLabel")
	if err == nil {
		t.Error("Expected error for unsupported filesystem type")
	}
}

func TestSetFAT32Label(t *testing.T) {
	// Test with non-existent partition (should fail gracefully)
	err := SetFAT32Label("/dev/nonexistent", "TestLabel")
	if err == nil {
		t.Error("Expected error when setting label on non-existent partition")
	}
}
