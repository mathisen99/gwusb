package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/filesystem"
	"github.com/mathisen/woeusb-go/internal/mount"
	"github.com/mathisen/woeusb-go/internal/validation"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("woeusb-go v0.1.0")
		return
	}

	// Test dependency checker
	dependencies, err := deps.CheckDependencies()
	if err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	fmt.Println("All dependencies found:")
	fmt.Printf("  wipefs: %s\n", dependencies.Wipefs)
	fmt.Printf("  parted: %s\n", dependencies.Parted)
	fmt.Printf("  lsblk: %s\n", dependencies.Lsblk)
	fmt.Printf("  blockdev: %s\n", dependencies.Blockdev)
	fmt.Printf("  mount: %s\n", dependencies.Mount)
	fmt.Printf("  umount: %s\n", dependencies.Umount)
	fmt.Printf("  7z: %s\n", dependencies.SevenZip)
	fmt.Printf("  mkfat: %s\n", dependencies.MkFat)
	fmt.Printf("  mkntfs: %s\n", dependencies.MkNTFS)
	fmt.Printf("  grub: %s\n", dependencies.GrubCmd)

	// Test validation functions
	fmt.Println("\nTesting validation functions:")

	// Test source validation with current file
	err = validation.ValidateSource("go.mod")
	if err != nil {
		fmt.Printf("Source validation failed: %v\n", err)
	} else {
		fmt.Println("✓ Source validation passed for go.mod")
	}

	// Test device naming patterns
	testPaths := []string{"/dev/sda", "/dev/sda1", "/dev/nvme0n1", "/dev/nvme0n1p1"}
	for _, path := range testPaths {
		info, _ := validation.GetDeviceInfo(path)
		if info != nil {
			fmt.Printf("Device info for %s: is_device=%v\n", path, info["is_device"])
		}
	}

	// Test mount functionality
	fmt.Println("\nTesting mount functions:")

	// Check if root is mounted
	mounted, mountpoints, err := mount.IsMounted("/")
	if err != nil {
		fmt.Printf("Mount check failed: %v\n", err)
	} else if mounted {
		fmt.Printf("✓ Root filesystem is mounted at: %v\n", mountpoints)
	}

	// Test busy check on non-existent device
	err = mount.CheckNotBusy("/dev/nonexistent")
	if err != nil {
		fmt.Printf("Busy check failed: %v\n", err)
	} else {
		fmt.Println("✓ Non-existent device is not busy")
	}

	// Test temp mountpoint creation
	tempMount, err := mount.CreateTempMountpoint("woeusb-test-")
	if err != nil {
		fmt.Printf("Temp mountpoint creation failed: %v\n", err)
	} else {
		fmt.Printf("✓ Created temp mountpoint: %s\n", tempMount)
		// Clean up
		_ = mount.CleanupMountpoint(tempMount)
		fmt.Println("✓ Cleaned up temp mountpoint")
	}

	// Test filesystem functionality
	fmt.Println("\nTesting filesystem functions:")

	// Test FAT32 limit check on current directory
	hasOversized, oversizedFiles, err := filesystem.CheckFAT32Limit(".")
	if err != nil {
		fmt.Printf("FAT32 limit check failed: %v\n", err)
	} else if hasOversized {
		fmt.Printf("⚠ Found %d files exceeding FAT32 4GB limit: %v\n", len(oversizedFiles), oversizedFiles)
	} else {
		fmt.Println("✓ All files in current directory are within FAT32 limits")
	}

	// Test filesystem suggestion
	suggestedFS, reason, err := filesystem.SuggestFilesystem(".")
	if err != nil {
		fmt.Printf("Filesystem suggestion failed: %v\n", err)
	} else {
		fmt.Printf("✓ Suggested filesystem: %s (%s)\n", suggestedFS, reason)
	}

	// Test size formatting
	testSizes := []int64{1024, 1024 * 1024, filesystem.FAT32MaxFileSize}
	for _, size := range testSizes {
		fmt.Printf("Size %d bytes = %s\n", size, filesystem.FormatSizeHuman(size))
	}
}
