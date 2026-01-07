package partition

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Wipe removes all filesystem signatures and partition table from a device
func Wipe(device string) error {
	// Run wipefs --all to remove all signatures
	cmd := exec.Command("wipefs", "--all", device)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to wipe device %s: %v", device, err)
	}

	// Verify no partitions remain by checking if lsblk shows any children
	if err := verifyNoPartitions(device); err != nil {
		return fmt.Errorf("verification failed after wiping %s: %v", device, err)
	}

	return nil
}

// CreateMBRTable creates a new MBR (msdos) partition table on the device
func CreateMBRTable(device string) error {
	cmd := exec.Command("parted", "-s", device, "mklabel", "msdos")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create MBR table on %s: %v", device, err)
	}
	return nil
}

// CreatePartition creates a partition on the device with the specified filesystem type
func CreatePartition(device, fstype string) error {
	var partType string
	var start, end string

	// Determine partition type and layout based on filesystem
	switch strings.ToUpper(fstype) {
	case "FAT32", "FAT":
		partType = "primary"
		start = "1MiB"
		end = "100%"
	case "NTFS":
		// For NTFS, we might want to leave space for UEFI:NTFS partition
		partType = "primary"
		start = "1MiB"
		end = "-512KiB" // Leave 512KB at the end
	default:
		return fmt.Errorf("unsupported filesystem type: %s", fstype)
	}

	// Create the partition
	cmd := exec.Command("parted", "-s", device, "mkpart", partType, start, end)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create partition on %s: %v", device, err)
	}

	return nil
}

// RereadPartitionTable forces the kernel to re-read the partition table
func RereadPartitionTable(device string) error {
	// Run blockdev --rereadpt
	cmd := exec.Command("blockdev", "--rereadpt", device)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to re-read partition table for %s: %v", device, err)
	}

	// Sleep for 3 seconds to allow the kernel to process the changes
	time.Sleep(3 * time.Second)

	return nil
}

// GetPartitionPath returns the path to the first partition of a device
func GetPartitionPath(device string) string {
	// Handle different device naming conventions
	if strings.Contains(device, "nvme") || strings.Contains(device, "mmcblk") {
		return device + "p1"
	}
	return device + "1"
}

// verifyNoPartitions checks that no partitions exist on the device
func verifyNoPartitions(device string) error {
	cmd := exec.Command("lsblk", "-n", "-o", "TYPE", device)
	output, err := cmd.Output()
	if err != nil {
		// If lsblk fails, the device might not exist or be accessible
		// This could be expected after wiping, so we don't treat it as an error
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "part" {
			return fmt.Errorf("partitions still exist on device %s", device)
		}
	}

	return nil
}

// CreateBootablePartition creates a bootable partition suitable for Windows USB
func CreateBootablePartition(device, fstype string) error {
	// Wipe the device first
	if err := Wipe(device); err != nil {
		return fmt.Errorf("failed to wipe device: %v", err)
	}

	// Create MBR partition table
	if err := CreateMBRTable(device); err != nil {
		return fmt.Errorf("failed to create MBR table: %v", err)
	}

	// Create the main partition
	if err := CreatePartition(device, fstype); err != nil {
		return fmt.Errorf("failed to create partition: %v", err)
	}

	// Re-read partition table
	if err := RereadPartitionTable(device); err != nil {
		return fmt.Errorf("failed to re-read partition table: %v", err)
	}

	return nil
}

// SetBootFlag sets the boot flag on the specified partition
func SetBootFlag(device string, partNum int) error {
	cmd := exec.Command("parted", "-s", device, "set", fmt.Sprintf("%d", partNum), "boot", "on")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set boot flag on %s partition %d: %v", device, partNum, err)
	}
	return nil
}

// GetDeviceSize returns the size of the device in bytes
func GetDeviceSize(device string) (int64, error) {
	cmd := exec.Command("blockdev", "--getsize64", device)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get device size for %s: %v", device, err)
	}

	var size int64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &size); err != nil {
		return 0, fmt.Errorf("failed to parse device size: %v", err)
	}

	return size, nil
}
