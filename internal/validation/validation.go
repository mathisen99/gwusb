package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
)

// ValidateSource checks if the source path exists and is either a file or block device
func ValidateSource(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source does not exist: %s", path)
		}
		return fmt.Errorf("cannot access source: %v", err)
	}

	mode := info.Mode()
	if mode.IsRegular() {
		return nil // Regular file (ISO)
	}

	// Check if it's a block device
	if mode&os.ModeDevice != 0 && mode&os.ModeCharDevice == 0 {
		return nil // Block device
	}

	return fmt.Errorf("source must be a regular file or block device: %s", path)
}

// ValidateTarget checks if the target is a valid block device based on the mode
func ValidateTarget(path, mode string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target does not exist: %s", path)
		}
		return fmt.Errorf("cannot access target: %v", err)
	}

	// Must be a block device
	fileMode := info.Mode()
	if fileMode&os.ModeDevice == 0 || fileMode&os.ModeCharDevice != 0 {
		return fmt.Errorf("target must be a block device: %s", path)
	}

	// Validate device vs partition based on mode
	switch mode {
	case "device":
		if !isWholeDevice(path) {
			return fmt.Errorf("device mode requires whole device (e.g., /dev/sdb), not partition: %s", path)
		}
	case "partition":
		if isWholeDevice(path) {
			return fmt.Errorf("partition mode requires partition (e.g., /dev/sdb1), not whole device: %s", path)
		}
	default:
		return fmt.Errorf("invalid mode: %s (must be 'device' or 'partition')", mode)
	}

	return nil
}

// isWholeDevice determines if the path refers to a whole device or a partition
// Handles both /dev/sdX and /dev/nvme0n1 naming patterns
func isWholeDevice(path string) bool {
	base := filepath.Base(path)

	// Standard SCSI/SATA devices: /dev/sda, /dev/sdb, etc.
	if matched, _ := regexp.MatchString(`^sd[a-z]$`, base); matched {
		return true
	}

	// Standard SCSI/SATA partitions: /dev/sda1, /dev/sdb2, etc.
	if matched, _ := regexp.MatchString(`^sd[a-z][0-9]+$`, base); matched {
		return false
	}

	// NVMe devices: /dev/nvme0n1, /dev/nvme1n1, etc.
	if matched, _ := regexp.MatchString(`^nvme[0-9]+n[0-9]+$`, base); matched {
		return true
	}

	// NVMe partitions: /dev/nvme0n1p1, /dev/nvme1n1p2, etc.
	if matched, _ := regexp.MatchString(`^nvme[0-9]+n[0-9]+p[0-9]+$`, base); matched {
		return false
	}

	// MMC devices: /dev/mmcblk0, /dev/mmcblk1, etc.
	if matched, _ := regexp.MatchString(`^mmcblk[0-9]+$`, base); matched {
		return true
	}

	// MMC partitions: /dev/mmcblk0p1, /dev/mmcblk1p2, etc.
	if matched, _ := regexp.MatchString(`^mmcblk[0-9]+p[0-9]+$`, base); matched {
		return false
	}

	// Default: assume it's a device if no number suffix
	return !regexp.MustCompile(`[0-9]+$`).MatchString(base)
}

// GetDeviceInfo returns basic information about a block device
func GetDeviceInfo(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("cannot get device info for %s", path)
	}

	return map[string]interface{}{
		"path":      path,
		"size":      info.Size(),
		"major":     int(stat.Rdev >> 8),
		"minor":     int(stat.Rdev & 0xff),
		"is_device": isWholeDevice(path),
	}, nil
}
