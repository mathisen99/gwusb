package mount

import (
	"os"
	"testing"
)

func TestGetMountInfo(t *testing.T) {
	mounts, err := GetMountInfo()
	if err != nil {
		t.Fatalf("GetMountInfo failed: %v", err)
	}

	if len(mounts) == 0 {
		t.Error("Expected at least some mounted filesystems")
	}

	// Check that root filesystem is mounted
	foundRoot := false
	for _, mount := range mounts {
		if mount.Mountpoint == "/" {
			foundRoot = true
			break
		}
	}

	if !foundRoot {
		t.Error("Expected to find root filesystem mounted at /")
	}
}

func TestIsMounted(t *testing.T) {
	// Test with root filesystem (should always be mounted)
	mounted, mountpoints, err := IsMounted("/")
	if err != nil {
		t.Fatalf("IsMounted failed: %v", err)
	}

	if !mounted {
		t.Error("Expected root filesystem to be mounted")
	}

	if len(mountpoints) == 0 {
		t.Error("Expected at least one mountpoint for root filesystem")
	}

	// Test with non-existent path
	mounted, mountpoints, err = IsMounted("/nonexistent/path")
	if err != nil {
		t.Fatalf("IsMounted failed for non-existent path: %v", err)
	}

	if mounted {
		t.Error("Expected non-existent path to not be mounted")
	}

	if len(mountpoints) > 0 {
		t.Error("Expected no mountpoints for non-existent path")
	}
}

func TestCheckNotBusy(t *testing.T) {
	// Test with non-existent device (should not be busy)
	err := CheckNotBusy("/dev/nonexistent")
	if err != nil {
		t.Errorf("Expected no error for non-existent device, got: %v", err)
	}

	// Note: We can't easily test actual unmounting without root privileges
	// and without potentially disrupting the test system
}

func TestUnmount(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mount_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test unmounting non-existent mountpoint (should fail gracefully)
	err = Unmount(tmpDir)
	if err == nil {
		t.Error("Expected error when unmounting non-mounted directory")
	}
}
