package filesystem

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// FAT32MaxFileSize is the maximum file size supported by FAT32 (4GB - 1 byte)
	FAT32MaxFileSize = 4*1024*1024*1024 - 1 // 4,294,967,295 bytes
)

// FormatFAT32 formats a partition with FAT32 filesystem
func FormatFAT32(partition string) error {
	cmd := exec.Command("mkdosfs", "-F", "32", partition)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format %s as FAT32: %v", partition, err)
	}
	return nil
}

// FormatNTFS formats a partition with NTFS filesystem and sets a label
func FormatNTFS(partition, label string) error {
	args := []string{"--quick"}
	if label != "" {
		args = append(args, "--label", label)
	}
	args = append(args, partition)

	cmd := exec.Command("mkntfs", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format %s as NTFS: %v", partition, err)
	}
	return nil
}

// FormatPartition formats a partition with the specified filesystem and label
func FormatPartition(partition, fstype, label string) error {
	switch strings.ToUpper(fstype) {
	case "FAT32", "FAT":
		if err := FormatFAT32(partition); err != nil {
			return err
		}
		// Set label after formatting if specified
		if label != "" {
			return SetFAT32Label(partition, label)
		}
		return nil
	case "NTFS":
		return FormatNTFS(partition, label)
	default:
		return fmt.Errorf("unsupported filesystem type: %s", fstype)
	}
}

// SetFAT32Label sets the label on a FAT32 partition
func SetFAT32Label(partition, label string) error {
	// Use fatlabel to set the label
	cmd := exec.Command("fatlabel", partition, label)
	if err := cmd.Run(); err != nil {
		// Fallback to dosfslabel if fatlabel is not available
		cmd = exec.Command("dosfslabel", partition, label)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set FAT32 label on %s: %v", partition, err)
		}
	}
	return nil
}

// CheckFAT32Limit walks through all files in the mountpoint and returns true if any file exceeds FAT32 limits
func CheckFAT32Limit(mountpoint string) (bool, []string, error) {
	var oversizedFiles []string

	err := filepath.Walk(mountpoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access rather than failing completely
			return nil
		}

		// Only check regular files
		if info.Mode().IsRegular() && info.Size() > FAT32MaxFileSize {
			relPath, _ := filepath.Rel(mountpoint, path)
			oversizedFiles = append(oversizedFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return false, nil, fmt.Errorf("failed to walk directory %s: %v", mountpoint, err)
	}

	return len(oversizedFiles) > 0, oversizedFiles, nil
}

// GetLargestFileSize returns the size of the largest file in the mountpoint
func GetLargestFileSize(mountpoint string) (int64, string, error) {
	var maxSize int64
	var maxFile string

	err := filepath.Walk(mountpoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.Mode().IsRegular() && info.Size() > maxSize {
			maxSize = info.Size()
			maxFile = path
		}

		return nil
	})

	if err != nil {
		return 0, "", fmt.Errorf("failed to walk directory %s: %v", mountpoint, err)
	}

	if maxFile != "" {
		relPath, _ := filepath.Rel(mountpoint, maxFile)
		maxFile = relPath
	}

	return maxSize, maxFile, nil
}

// FormatSizeHuman formats a byte size into human-readable format
func FormatSizeHuman(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// SuggestFilesystem suggests the appropriate filesystem based on content analysis
func SuggestFilesystem(mountpoint string) (string, string, error) {
	hasOversized, oversizedFiles, err := CheckFAT32Limit(mountpoint)
	if err != nil {
		return "", "", err
	}

	if hasOversized {
		maxSize, maxFile, err := GetLargestFileSize(mountpoint)
		if err != nil {
			return "NTFS", fmt.Sprintf("Files exceed FAT32 4GB limit (%d files)", len(oversizedFiles)), nil
		}

		reason := fmt.Sprintf("File '%s' (%s) exceeds FAT32 4GB limit", maxFile, FormatSizeHuman(maxSize))
		if len(oversizedFiles) > 1 {
			reason += fmt.Sprintf(" (and %d other files)", len(oversizedFiles)-1)
		}

		return "NTFS", reason, nil
	}

	return "FAT32", "All files are within FAT32 limits", nil
}

// ValidateFilesystemChoice validates if the chosen filesystem can handle the content
func ValidateFilesystemChoice(mountpoint, filesystem string) error {
	if filesystem == "FAT32" || filesystem == "FAT" {
		hasOversized, oversizedFiles, err := CheckFAT32Limit(mountpoint)
		if err != nil {
			return err
		}

		if hasOversized {
			return fmt.Errorf("cannot use FAT32: %d files exceed 4GB limit: %v",
				len(oversizedFiles), oversizedFiles)
		}
	}

	return nil
}
