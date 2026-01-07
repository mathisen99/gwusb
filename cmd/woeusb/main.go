package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mathisen/woeusb-go/internal/deps"
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
}
