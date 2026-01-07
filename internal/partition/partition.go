package partition

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CreateUEFINTFSPartition creates a 512KB partition at the end of the device for UEFI:NTFS
func CreateUEFINTFSPartition(device string) (string, error) {
	// Create a small partition at the end of the device
	cmd := exec.Command("parted", "-s", device, "mkpart", "primary", "fat32", "-512KiB", "100%")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create UEFI:NTFS partition on %s: %v", device, err)
	}

	// Re-read partition table
	if err := RereadPartitionTable(device); err != nil {
		return "", fmt.Errorf("failed to re-read partition table: %v", err)
	}

	// Return the partition path (should be partition 2 for UEFI:NTFS)
	var partitionPath string
	if strings.Contains(device, "nvme") || strings.Contains(device, "mmcblk") {
		partitionPath = device + "p2"
	} else {
		partitionPath = device + "2"
	}

	return partitionPath, nil
}

// InstallUEFINTFS downloads uefi-ntfs.img and writes it to the partition
func InstallUEFINTFS(partition, tempDir string) error {
	// UEFI:NTFS image URL (official release)
	imageURL := "https://github.com/pbatard/uefi-ntfs/releases/download/v1.4/uefi-ntfs.img"

	// Download the image to temp directory
	imagePath := filepath.Join(tempDir, "uefi-ntfs.img")
	if err := downloadFile(imageURL, imagePath); err != nil {
		// Handle download failure gracefully (warning, not error)
		fmt.Fprintf(os.Stderr, "Warning: Failed to download UEFI:NTFS image: %v\n", err)
		fmt.Fprintf(os.Stderr, "UEFI booting may not work properly for NTFS partitions\n")
		return nil // Return nil to continue without failing
	}

	// Write the image to the partition
	if err := writeImageToPartition(imagePath, partition); err != nil {
		return fmt.Errorf("failed to write UEFI:NTFS image to partition %s: %v", partition, err)
	}

	// Clean up downloaded image
	_ = os.Remove(imagePath)

	return nil
}

// downloadFile downloads a file from URL to the specified path
func downloadFile(url, filepath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filepath, err)
	}
	defer func() { _ = out.Close() }()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write downloaded data: %v", err)
	}

	return nil
}

// writeImageToPartition writes an image file to a partition using dd
func writeImageToPartition(imagePath, partition string) error {
	cmd := exec.Command("dd", "if="+imagePath, "of="+partition, "bs=1M", "status=progress")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write image with dd: %v", err)
	}
	return nil
}

// CreateNTFSWithUEFI creates an NTFS partition setup with UEFI:NTFS support
func CreateNTFSWithUEFI(device, tempDir string) (string, string, error) {
	// Wipe the device first
	if err := Wipe(device); err != nil {
		return "", "", fmt.Errorf("failed to wipe device: %v", err)
	}

	// Create MBR partition table
	if err := CreateMBRTable(device); err != nil {
		return "", "", fmt.Errorf("failed to create MBR table: %v", err)
	}

	// Create the main NTFS partition (leaving space for UEFI:NTFS)
	if err := CreatePartition(device, "NTFS"); err != nil {
		return "", "", fmt.Errorf("failed to create main partition: %v", err)
	}

	// Create UEFI:NTFS partition
	uefiPartition, err := CreateUEFINTFSPartition(device)
	if err != nil {
		return "", "", fmt.Errorf("failed to create UEFI:NTFS partition: %v", err)
	}

	// Install UEFI:NTFS
	if err := InstallUEFINTFS(uefiPartition, tempDir); err != nil {
		return "", "", fmt.Errorf("failed to install UEFI:NTFS: %v", err)
	}

	// Return main partition path
	mainPartition := GetPartitionPath(device)
	return mainPartition, uefiPartition, nil
}

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
