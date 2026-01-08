package components

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

// TestProperty10_ISOFileValidation tests Property 10:
// For any file path, ValidateISO SHALL return nil if the file exists
// and is readable, and an error otherwise.
func TestProperty10_ISOFileValidation(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a valid ISO file
	validISO := filepath.Join(tmpDir, "test.iso")
	if err := os.WriteFile(validISO, []byte("ISO content"), 0644); err != nil {
		t.Fatalf("Failed to create test ISO: %v", err)
	}

	// Create a non-ISO file
	nonISO := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(nonISO, []byte("text content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a directory
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testCases := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid ISO file",
			path:        validISO,
			expectError: false,
		},
		{
			name:        "Empty path",
			path:        "",
			expectError: true,
			errorMsg:    "no file path provided",
		},
		{
			name:        "Non-existent file",
			path:        filepath.Join(tmpDir, "nonexistent.iso"),
			expectError: true,
			errorMsg:    "file does not exist",
		},
		{
			name:        "Non-ISO extension",
			path:        nonISO,
			expectError: true,
			errorMsg:    "file is not an ISO image",
		},
		{
			name:        "Directory instead of file",
			path:        testDir,
			expectError: true,
			errorMsg:    "path is a directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateISO(tc.path)
			if tc.expectError {
				if err == nil {
					t.Errorf("ValidateISO(%q) expected error containing %q, got nil", tc.path, tc.errorMsg)
				} else if tc.errorMsg != "" && !containsString(err.Error(), tc.errorMsg) {
					t.Errorf("ValidateISO(%q) error = %q, want error containing %q", tc.path, err.Error(), tc.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateISO(%q) unexpected error: %v", tc.path, err)
				}
			}
		})
	}
}

// TestValidateISO_CaseInsensitiveExtension tests that .ISO and .iso are both accepted
func TestValidateISO_CaseInsensitiveExtension(t *testing.T) {
	tmpDir := t.TempDir()

	extensions := []string{".iso", ".ISO", ".Iso", ".iSo"}
	for _, ext := range extensions {
		path := filepath.Join(tmpDir, "test"+ext)
		if err := os.WriteFile(path, []byte("ISO content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := ValidateISO(path)
		if err != nil {
			t.Errorf("ValidateISO(%q) should accept %s extension, got error: %v", path, ext, err)
		}
	}
}

// TestValidateISO_UnreadableFile tests that unreadable files are rejected
func TestValidateISO_UnreadableFile(t *testing.T) {
	// Skip if running as root (root can read anything)
	if os.Getuid() == 0 {
		t.Skip("Skipping unreadable file test when running as root")
	}

	tmpDir := t.TempDir()
	unreadable := filepath.Join(tmpDir, "unreadable.iso")

	// Create file with no read permissions
	if err := os.WriteFile(unreadable, []byte("ISO content"), 0000); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() { _ = os.Chmod(unreadable, 0644) }() // Restore permissions for cleanup

	err := ValidateISO(unreadable)
	if err == nil {
		t.Error("ValidateISO() should reject unreadable files")
	}
}

// TestValidateISOWithStatFunc_PropertyBased tests the validation with mock functions
func TestValidateISOWithStatFunc_PropertyBased(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		statFunc    func(string) (os.FileInfo, error)
		openFunc    func(string) (*os.File, error)
		expectError bool
	}{
		{
			name: "Valid ISO file",
			path: "/path/to/valid.iso",
			statFunc: func(path string) (os.FileInfo, error) {
				return mockFileInfo{name: "valid.iso", size: 1024, isDir: false}, nil
			},
			openFunc:    nil, // Skip open check
			expectError: false,
		},
		{
			name: "File is a directory",
			path: "/path/to/dir.iso",
			statFunc: func(path string) (os.FileInfo, error) {
				return mockFileInfo{name: "dir.iso", size: 0, isDir: true}, nil
			},
			openFunc:    nil,
			expectError: true,
		},
		{
			name: "File does not exist",
			path: "/path/to/missing.iso",
			statFunc: func(path string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			openFunc:    nil,
			expectError: true,
		},
		{
			name: "Wrong extension",
			path: "/path/to/file.txt",
			statFunc: func(path string) (os.FileInfo, error) {
				return mockFileInfo{name: "file.txt", size: 1024, isDir: false}, nil
			},
			openFunc:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateISOWithStatFunc(tc.path, tc.statFunc, tc.openFunc)
			if tc.expectError && err == nil {
				t.Errorf("ValidateISOWithStatFunc() expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("ValidateISOWithStatFunc() unexpected error: %v", err)
			}
		})
	}
}

// Note: containsString and findSubstring are defined in device_selector_test.go
