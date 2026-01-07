package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mathisen/woeusb-go/internal/deps"
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
}
