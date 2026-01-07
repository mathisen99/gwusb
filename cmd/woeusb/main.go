package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mathisen/woeusb-go/internal/bootloader"
	"github.com/mathisen/woeusb-go/internal/copy"
	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/filesystem"
	"github.com/mathisen/woeusb-go/internal/mount"
	"github.com/mathisen/woeusb-go/internal/partition"
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

	// Test partition functionality
	fmt.Println("\nTesting partition functions:")

	// Test partition path generation
	testDevices := []string{"/dev/sda", "/dev/nvme0n1", "/dev/mmcblk0"}
	for _, device := range testDevices {
		partPath := partition.GetPartitionPath(device)
		fmt.Printf("Partition path for %s: %s\n", device, partPath)
	}

	// Test device size (will fail for non-existent devices, which is expected)
	size, err := partition.GetDeviceSize("/dev/nonexistent")
	if err != nil {
		fmt.Printf("✓ Expected error for non-existent device: %v\n", err)
	} else {
		fmt.Printf("Device size: %d bytes\n", size)
	}

	// Test formatting functionality
	fmt.Println("\nTesting format functions:")

	// Test formatting operations (will fail for non-existent partitions, which is expected)
	err = filesystem.FormatFAT32("/dev/nonexistent")
	if err != nil {
		fmt.Printf("✓ Expected error for FAT32 format: %v\n", err)
	}

	err = filesystem.FormatNTFS("/dev/nonexistent", "Windows USB")
	if err != nil {
		fmt.Printf("✓ Expected error for NTFS format: %v\n", err)
	}

	// Test comprehensive formatting
	err = filesystem.FormatPartition("/dev/nonexistent", "FAT32", "USB Drive")
	if err != nil {
		fmt.Printf("✓ Expected error for partition format: %v\n", err)
	}

	// Test copy functionality
	fmt.Println("\nTesting copy functions:")

	// Create temporary directories for copy testing
	srcDir, err := os.MkdirTemp("", "woeusb-copy-src-")
	if err != nil {
		fmt.Printf("Failed to create temp source dir: %v\n", err)
	} else {
		defer func() { _ = os.RemoveAll(srcDir) }()

		dstDir, err := os.MkdirTemp("", "woeusb-copy-dst-")
		if err != nil {
			fmt.Printf("Failed to create temp destination dir: %v\n", err)
		} else {
			defer func() { _ = os.RemoveAll(dstDir) }()

			// Create test files
			testFile := srcDir + "/test.txt"
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				fmt.Printf("Failed to create test file: %v\n", err)
			} else {
				// Test copy with progress
				fmt.Printf("Copying from %s to %s\n", srcDir, dstDir)
				err = copy.CopyDirectoryQuiet(srcDir, dstDir)
				if err != nil {
					fmt.Printf("Copy failed: %v\n", err)
				} else {
					fmt.Println("✓ Copy completed successfully")

					// Validate copy
					err = copy.ValidateCopy(srcDir, dstDir)
					if err != nil {
						fmt.Printf("Copy validation failed: %v\n", err)
					} else {
						fmt.Println("✓ Copy validation passed")
					}
				}
			}
		}
	}

	// Test bootloader functionality
	fmt.Println("\nTesting bootloader functions:")

	// Test GRUB prefix detection
	testCommands := []string{"grub-install", "grub2-install", "/usr/bin/grub-install"}
	for _, cmd := range testCommands {
		prefix := bootloader.DetectGRUBPrefix(cmd)
		fmt.Printf("GRUB prefix for %s: %s\n", cmd, prefix)
	}

	// Test GRUB configuration writing
	tmpBootDir, err := os.MkdirTemp("", "woeusb-boot-")
	if err != nil {
		fmt.Printf("Failed to create temp boot dir: %v\n", err)
	} else {
		defer func() { _ = os.RemoveAll(tmpBootDir) }()

		err = bootloader.WriteGRUBConfig(tmpBootDir, "grub")
		if err != nil {
			fmt.Printf("Failed to write GRUB config: %v\n", err)
		} else {
			fmt.Println("✓ GRUB configuration written successfully")

			// Verify installation
			err = bootloader.CheckGRUBInstallation(tmpBootDir, "grub")
			if err != nil {
				fmt.Printf("GRUB installation check failed: %v\n", err)
			} else {
				fmt.Println("✓ GRUB installation verified")
			}

			// Test Windows 7 detection
			fmt.Println("\nTesting Windows 7 UEFI workaround:")

			// Test with non-Windows 7 directory
			isWin7, err := bootloader.IsWindows7(tmpBootDir)
			if err != nil {
				fmt.Printf("Windows 7 check failed: %v\n", err)
			} else {
				fmt.Printf("Is Windows 7: %v\n", isWin7)
			}

			// Test UEFI bootloader check
			err = bootloader.CheckUEFIBootloader(tmpBootDir)
			if err != nil {
				fmt.Printf("✓ Expected error for missing UEFI bootloader: %v\n", err)
			}

			// Test UEFI:NTFS functionality
			fmt.Println("\nTesting UEFI:NTFS functions:")

			// Test UEFI:NTFS partition creation (will fail for non-existent device)
			_, err = partition.CreateUEFINTFSPartition("/dev/nonexistent")
			if err != nil {
				fmt.Printf("✓ Expected error for UEFI:NTFS partition creation: %v\n", err)
			}

			// Test UEFI:NTFS installation
			err = partition.InstallUEFINTFS("/dev/nonexistent", tmpBootDir)
			if err != nil {
				fmt.Printf("UEFI:NTFS installation failed: %v\n", err)
			} else {
				fmt.Println("✓ UEFI:NTFS installation handled gracefully")
			}

			// Test boot flag functionality
			fmt.Println("\nTesting boot flag workaround:")

			// Test setting boot flag (will fail for non-existent device)
			err = partition.SetBootFlag("/dev/nonexistent", 1)
			if err != nil {
				fmt.Printf("✓ Expected error for boot flag setting: %v\n", err)
			}
		}
	}
}
