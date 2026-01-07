package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateTotalSize(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "copy_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	if err := os.WriteFile(file2, []byte("world!"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	stats, err := calculateTotalSize(tmpDir)
	if err != nil {
		t.Fatalf("calculateTotalSize failed: %v", err)
	}

	if stats.TotalFiles != 2 {
		t.Errorf("Expected 2 files, got %d", stats.TotalFiles)
	}

	if stats.TotalBytes != 11 { // "hello" (5) + "world!" (6)
		t.Errorf("Expected 11 bytes, got %d", stats.TotalBytes)
	}
}

func TestCopyWithProgress(t *testing.T) {
	// Create source directory
	srcDir, err := os.MkdirTemp("", "copy_src")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(srcDir) }()

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "copy_dst")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dstDir) }()

	// Create test files in source
	testFile := filepath.Join(srcDir, "test.txt")
	testContent := []byte("test content for copying")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create subdirectory with file
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("sub content"), 0644); err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	// Track progress calls
	var progressCalls int
	progressFn := func(bytesCopied, totalBytes int64, currentFile string) {
		progressCalls++
		if bytesCopied > totalBytes {
			t.Errorf("Bytes copied (%d) exceeds total (%d)", bytesCopied, totalBytes)
		}
	}

	// Copy with progress
	err = CopyWithProgress(srcDir, dstDir, progressFn)
	if err != nil {
		t.Fatalf("CopyWithProgress failed: %v", err)
	}

	// Verify files were copied
	dstTestFile := filepath.Join(dstDir, "test.txt")
	if _, err := os.Stat(dstTestFile); os.IsNotExist(err) {
		t.Error("Test file was not copied")
	}

	dstSubFile := filepath.Join(dstDir, "subdir", "subfile.txt")
	if _, err := os.Stat(dstSubFile); os.IsNotExist(err) {
		t.Error("Sub file was not copied")
	}

	// Verify content
	copiedContent, err := os.ReadFile(dstTestFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copiedContent) != string(testContent) {
		t.Errorf("Content mismatch: expected %s, got %s", testContent, copiedContent)
	}

	// Verify progress was called
	if progressCalls == 0 {
		t.Error("Progress function was never called")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestValidateCopy(t *testing.T) {
	// Create source directory
	srcDir, err := os.MkdirTemp("", "validate_src")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(srcDir) }()

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "validate_dst")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dstDir) }()

	// Create identical content in both directories
	testContent := []byte("identical content")

	srcFile := filepath.Join(srcDir, "test.txt")
	dstFile := filepath.Join(dstDir, "test.txt")

	if err := os.WriteFile(srcFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := os.WriteFile(dstFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}

	// Validation should pass
	err = ValidateCopy(srcDir, dstDir)
	if err != nil {
		t.Errorf("ValidateCopy failed for identical directories: %v", err)
	}

	// Remove destination file to test mismatch
	_ = os.Remove(dstFile)

	err = ValidateCopy(srcDir, dstDir)
	if err == nil {
		t.Error("ValidateCopy should have failed for mismatched directories")
	}
}

func TestCopyDirectoryQuiet(t *testing.T) {
	// Create source directory
	srcDir, err := os.MkdirTemp("", "quiet_src")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(srcDir) }()

	// Create destination directory
	dstDir, err := os.MkdirTemp("", "quiet_dst")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dstDir) }()

	// Create test file
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Copy quietly (no progress function)
	err = CopyDirectoryQuiet(srcDir, dstDir)
	if err != nil {
		t.Fatalf("CopyDirectoryQuiet failed: %v", err)
	}

	// Verify file was copied
	dstTestFile := filepath.Join(dstDir, "test.txt")
	if _, err := os.Stat(dstTestFile); os.IsNotExist(err) {
		t.Error("Test file was not copied")
	}
}
