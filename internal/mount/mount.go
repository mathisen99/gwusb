package mount

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// MountInfo represents information about a mounted filesystem
type MountInfo struct {
	Device     string
	Mountpoint string
	Filesystem string
	Options    string
}

// GetMountInfo returns all currently mounted filesystems by parsing /proc/mounts
func GetMountInfo() ([]MountInfo, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/mounts: %v", err)
	}
	defer func() { _ = file.Close() }()

	var mounts []MountInfo
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			mounts = append(mounts, MountInfo{
				Device:     fields[0],
				Mountpoint: fields[1],
				Filesystem: fields[2],
				Options:    fields[3],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/mounts: %v", err)
	}

	return mounts, nil
}

// Mount mounts a filesystem using syscall or shell command
func Mount(source, mountpoint, fstype string, opts []string) error {
	// Ensure mountpoint exists
	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		return fmt.Errorf("failed to create mountpoint %s: %v", mountpoint, err)
	}

	// Try syscall first for better performance
	var flags uintptr
	var data string

	// Parse common mount options
	for _, opt := range opts {
		switch opt {
		case "ro", "readonly":
			flags |= syscall.MS_RDONLY
		case "noexec":
			flags |= syscall.MS_NOEXEC
		case "nosuid":
			flags |= syscall.MS_NOSUID
		case "nodev":
			flags |= syscall.MS_NODEV
		default:
			if data != "" {
				data += ","
			}
			data += opt
		}
	}

	// Attempt syscall mount
	err := syscall.Mount(source, mountpoint, fstype, flags, data)
	if err == nil {
		return nil
	}

	// Fallback to shell command
	args := []string{"-t", fstype}
	if len(opts) > 0 {
		args = append(args, "-o", strings.Join(opts, ","))
	}
	args = append(args, source, mountpoint)

	cmd := exec.Command("mount", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount %s at %s: %v", source, mountpoint, err)
	}

	return nil
}

// Unmount attempts to unmount a filesystem at the given mountpoint
func Unmount(mountpoint string) error {
	// Try syscall first
	err := syscall.Unmount(mountpoint, 0)
	if err == nil {
		return nil
	}

	// Fallback to shell command
	cmd := exec.Command("umount", mountpoint)
	if err := cmd.Run(); err != nil {
		// Try lazy unmount as fallback
		cmd = exec.Command("umount", "-l", mountpoint)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to unmount %s: %v", mountpoint, err)
		}
	}
	return nil
}

// CreateTempMountpoint creates a temporary directory for mounting
func CreateTempMountpoint(prefix string) (string, error) {
	tmpDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp mountpoint: %v", err)
	}
	return tmpDir, nil
}

// CleanupMountpoint unmounts and removes a temporary mountpoint
func CleanupMountpoint(mountpoint string) error {
	// Check if it's mounted first
	mounted, _, err := IsMounted(mountpoint)
	if err != nil {
		return fmt.Errorf("failed to check mount status: %v", err)
	}

	if mounted {
		if err := Unmount(mountpoint); err != nil {
			return fmt.Errorf("failed to unmount %s: %v", mountpoint, err)
		}
	}

	// Remove the directory
	if err := os.RemoveAll(mountpoint); err != nil {
		return fmt.Errorf("failed to remove mountpoint %s: %v", mountpoint, err)
	}

	return nil
}

// CheckNotBusy checks if a device is mounted and attempts to unmount it
func CheckNotBusy(devicePath string) error {
	mounts, err := GetMountInfo()
	if err != nil {
		return fmt.Errorf("failed to get mount info: %v", err)
	}

	// Find all mount points for this device or its partitions
	var mountedPaths []string
	for _, mount := range mounts {
		if mount.Device == devicePath || strings.HasPrefix(mount.Device, devicePath) {
			mountedPaths = append(mountedPaths, mount.Mountpoint)
		}
	}

	if len(mountedPaths) == 0 {
		return nil // Device is not mounted
	}

	// Attempt to unmount all mount points
	for _, mountpoint := range mountedPaths {
		if err := Unmount(mountpoint); err != nil {
			return fmt.Errorf("device %s is busy (mounted at %s) and cannot be unmounted: %v",
				devicePath, mountpoint, err)
		}
	}

	return nil
}

// IsMounted checks if a specific device or mountpoint is currently mounted
func IsMounted(path string) (bool, []string, error) {
	mounts, err := GetMountInfo()
	if err != nil {
		return false, nil, err
	}

	var mountpoints []string
	for _, mount := range mounts {
		if mount.Device == path || mount.Mountpoint == path || strings.HasPrefix(mount.Device, path) {
			mountpoints = append(mountpoints, mount.Mountpoint)
		}
	}

	return len(mountpoints) > 0, mountpoints, nil
}

// MountISO mounts an ISO file to a temporary mountpoint
func MountISO(isoPath string) (string, error) {
	mountpoint, err := CreateTempMountpoint("woeusb-iso-")
	if err != nil {
		return "", err
	}

	// Try UDF first (Windows 10/11 ISOs), then fall back to iso9660
	// Using "auto" lets the kernel detect the correct filesystem
	if err := Mount(isoPath, mountpoint, "udf", []string{"ro", "loop"}); err != nil {
		// Fallback to iso9660 for older ISOs
		if err := Mount(isoPath, mountpoint, "iso9660", []string{"ro", "loop"}); err != nil {
			_ = os.RemoveAll(mountpoint)
			return "", fmt.Errorf("failed to mount ISO %s: %v", isoPath, err)
		}
	}

	return mountpoint, nil
}

// MountDevice mounts a block device to a temporary mountpoint
func MountDevice(devicePath, fstype string) (string, error) {
	mountpoint, err := CreateTempMountpoint("woeusb-dev-")
	if err != nil {
		return "", err
	}

	// Normalize filesystem type
	switch strings.ToLower(fstype) {
	case "fat", "fat32", "vfat":
		fstype = "vfat"
	case "ntfs", "ntfs-3g":
		fstype = "ntfs3" // Use kernel ntfs3 driver (faster than ntfs-3g FUSE)
	}

	opts := []string{}

	if err := Mount(devicePath, mountpoint, fstype, opts); err != nil {
		_ = os.RemoveAll(mountpoint)
		return "", fmt.Errorf("failed to mount device %s: %v", devicePath, err)
	}

	return mountpoint, nil
}
