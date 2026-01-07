package partition

import (
	"testing"
)

func TestGetPartitionPath(t *testing.T) {
	tests := []struct {
		device   string
		expected string
	}{
		{"/dev/sda", "/dev/sda1"},
		{"/dev/sdb", "/dev/sdb1"},
		{"/dev/nvme0n1", "/dev/nvme0n1p1"},
		{"/dev/nvme1n1", "/dev/nvme1n1p1"},
		{"/dev/mmcblk0", "/dev/mmcblk0p1"},
		{"/dev/mmcblk1", "/dev/mmcblk1p1"},
	}

	for _, test := range tests {
		result := GetPartitionPath(test.device)
		if result != test.expected {
			t.Errorf("GetPartitionPath(%s) = %s, expected %s", test.device, result, test.expected)
		}
	}
}

func TestWipe(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	err := Wipe("/dev/nonexistent")
	if err == nil {
		t.Error("Expected error when wiping non-existent device")
	}

	// Note: We can't test actual device wiping without root privileges
	// and without potentially destroying data
}

func TestCreateMBRTable(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	err := CreateMBRTable("/dev/nonexistent")
	if err == nil {
		t.Error("Expected error when creating MBR table on non-existent device")
	}
}

func TestCreatePartition(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	err := CreatePartition("/dev/nonexistent", "FAT32")
	if err == nil {
		t.Error("Expected error when creating partition on non-existent device")
	}

	// Test with unsupported filesystem
	err = CreatePartition("/dev/nonexistent", "UNSUPPORTED")
	if err == nil {
		t.Error("Expected error for unsupported filesystem type")
	}
}

func TestRereadPartitionTable(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	err := RereadPartitionTable("/dev/nonexistent")
	if err == nil {
		t.Error("Expected error when re-reading partition table of non-existent device")
	}
}

func TestSetBootFlag(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	err := SetBootFlag("/dev/nonexistent", 1)
	if err == nil {
		t.Error("Expected error when setting boot flag on non-existent device")
	}
}

func TestGetDeviceSize(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	_, err := GetDeviceSize("/dev/nonexistent")
	if err == nil {
		t.Error("Expected error when getting size of non-existent device")
	}
}

func TestCreateBootablePartition(t *testing.T) {
	// Test with non-existent device (should fail gracefully)
	err := CreateBootablePartition("/dev/nonexistent", "FAT32")
	if err == nil {
		t.Error("Expected error when creating bootable partition on non-existent device")
	}

	// Note: This is a comprehensive test that would require actual hardware
	// and root privileges to test properly
}
