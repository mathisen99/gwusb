package mount

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// Unmount attempts to unmount a filesystem at the given mountpoint
func Unmount(mountpoint string) error {
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
