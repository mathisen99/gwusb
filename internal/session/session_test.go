package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionCleanup(t *testing.T) {
	session := &Session{}

	// Test cleanup with no resources
	err := session.Cleanup()
	if err != nil {
		t.Errorf("Expected no error for empty cleanup, got: %v", err)
	}
}

func TestSessionCleanupWithTempDir(t *testing.T) {
	session := &Session{}

	// Create a real temp directory
	tempDir, err := os.MkdirTemp("", "session-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a file in the temp directory
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	session.TempDir = tempDir

	// Verify directory exists before cleanup
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatal("Temp directory should exist before cleanup")
	}

	// Cleanup should remove the directory
	err = session.Cleanup()
	if err != nil {
		t.Errorf("Unexpected error during cleanup: %v", err)
	}

	// Verify directory is removed after cleanup
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Temp directory should be removed after cleanup")
	}

	// Verify session state is cleared
	if session.TempDir != "" {
		t.Error("Session TempDir should be cleared after cleanup")
	}
}

func TestSessionCleanupWithMountpoints(t *testing.T) {
	session := &Session{}

	// Set non-existent mountpoints (cleanup should handle gracefully)
	session.SourceMount = "/tmp/nonexistent-source"
	session.TargetMount = "/tmp/nonexistent-target"

	// Cleanup should handle non-existent mountpoints gracefully
	err := session.Cleanup()
	// We expect errors for non-existent mountpoints, but it should not panic
	if err == nil {
		t.Log("Cleanup handled non-existent mountpoints gracefully")
	} else {
		t.Logf("Got expected errors for non-existent mountpoints: %v", err)
	}
}

func TestSessionSetupSignalHandler(t *testing.T) {
	session := &Session{}

	// This should not panic
	session.SetupSignalHandler()

	// We can't easily test the signal handling without actually sending signals,
	// but we can verify the function doesn't crash
}

func TestMultipleCleanups(t *testing.T) {
	session := &Session{}

	// Multiple cleanups should not panic
	err1 := session.Cleanup()
	err2 := session.Cleanup()
	err3 := session.Cleanup()

	if err1 != nil {
		t.Errorf("First cleanup failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second cleanup failed: %v", err2)
	}
	if err3 != nil {
		t.Errorf("Third cleanup failed: %v", err3)
	}
}

func TestSessionFields(t *testing.T) {
	session := &Session{
		Source:          "/path/to/source.iso",
		Target:          "/dev/sdb",
		TargetDevice:    "/dev/sdb",
		TargetPartition: "/dev/sdb1",
		Mode:            "device",
		Filesystem:      "FAT32",
		Label:           "WINDOWS",
		SourceMount:     "/tmp/source",
		TargetMount:     "/tmp/target",
		TempDir:         "/tmp/temp",
		SkipGRUB:        false,
		SetBootFlag:     true,
		Verbose:         true,
		NoColor:         false,
	}

	// Verify fields are set correctly
	if session.Source != "/path/to/source.iso" {
		t.Errorf("Expected Source '/path/to/source.iso', got '%s'", session.Source)
	}
	if session.Target != "/dev/sdb" {
		t.Errorf("Expected Target '/dev/sdb', got '%s'", session.Target)
	}
	if session.TargetDevice != "/dev/sdb" {
		t.Errorf("Expected TargetDevice '/dev/sdb', got '%s'", session.TargetDevice)
	}
	if session.TargetPartition != "/dev/sdb1" {
		t.Errorf("Expected TargetPartition '/dev/sdb1', got '%s'", session.TargetPartition)
	}
	if session.Mode != "device" {
		t.Errorf("Expected Mode 'device', got '%s'", session.Mode)
	}
	if session.Label != "WINDOWS" {
		t.Errorf("Expected Label 'WINDOWS', got '%s'", session.Label)
	}
	if session.SourceMount != "/tmp/source" {
		t.Errorf("Expected SourceMount '/tmp/source', got '%s'", session.SourceMount)
	}
	if session.TargetMount != "/tmp/target" {
		t.Errorf("Expected TargetMount '/tmp/target', got '%s'", session.TargetMount)
	}
	if session.TempDir != "/tmp/temp" {
		t.Errorf("Expected TempDir '/tmp/temp', got '%s'", session.TempDir)
	}
	if session.SkipGRUB != false {
		t.Errorf("Expected SkipGRUB false, got %v", session.SkipGRUB)
	}
	if session.NoColor != false {
		t.Errorf("Expected NoColor false, got %v", session.NoColor)
	}

	if session.Filesystem != "FAT32" {
		t.Errorf("Expected Filesystem 'FAT32', got '%s'", session.Filesystem)
	}

	if !session.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if !session.SetBootFlag {
		t.Error("Expected SetBootFlag to be true")
	}
}
