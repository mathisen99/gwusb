package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mathisen/woeusb-go/internal/bootloader"
	"github.com/mathisen/woeusb-go/internal/copy"
	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/filesystem"
	"github.com/mathisen/woeusb-go/internal/mount"
	"github.com/mathisen/woeusb-go/internal/partition"
	"github.com/mathisen/woeusb-go/internal/session"
	"github.com/mathisen/woeusb-go/internal/validation"
)

const version = "0.1.0"

type config struct {
	device       bool
	partition    bool
	filesystem   string
	label        string
	biosBootFlag bool
	skipGrub     bool
	verbose      bool
	noColor      bool
	source       string
	target       string
}

func main() {
	cfg := parseArgs()
	if cfg == nil {
		return
	}

	// Setup session for cleanup
	sess := &session.Session{
		Source:      cfg.source,
		Target:      cfg.target,
		Mode:        getMode(cfg),
		Filesystem:  cfg.filesystem,
		Label:       cfg.label,
		SkipGRUB:    cfg.skipGrub,
		SetBootFlag: cfg.biosBootFlag,
		Verbose:     cfg.verbose,
		NoColor:     cfg.noColor,
	}

	// Setup signal handler for cleanup
	sess.SetupSignalHandler()
	defer func() { _ = sess.Cleanup() }()

	// Check dependencies
	if err := checkDependencies(); err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	// Validate source and target
	if err := validateInputs(cfg); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Execute the appropriate mode
	if cfg.device {
		if err := executeDeviceMode(cfg, sess); err != nil {
			log.Fatalf("Device mode failed: %v", err)
		}
	} else {
		if err := executePartitionMode(cfg, sess); err != nil {
			log.Fatalf("Partition mode failed: %v", err)
		}
	}

	fmt.Println("✓ WoeUSB operation completed successfully!")
}

func parseArgs() *config {
	var cfg config
	var showVersion bool

	flag.BoolVar(&cfg.device, "device", false, "Wipe entire device and create bootable USB")
	flag.BoolVar(&cfg.device, "d", false, "Wipe entire device (shorthand)")
	flag.BoolVar(&cfg.partition, "partition", false, "Use existing partition")
	flag.BoolVar(&cfg.partition, "p", false, "Use existing partition (shorthand)")
	flag.StringVar(&cfg.filesystem, "target-filesystem", "FAT", "Target filesystem: FAT or NTFS")
	flag.StringVar(&cfg.label, "label", "Windows USB", "Filesystem label")
	flag.StringVar(&cfg.label, "l", "Windows USB", "Filesystem label (shorthand)")
	flag.BoolVar(&cfg.biosBootFlag, "workaround-bios-boot-flag", false, "Set boot flag for buggy BIOSes")
	flag.BoolVar(&cfg.skipGrub, "workaround-skip-grub", false, "Skip GRUB installation")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&cfg.verbose, "v", false, "Verbose output (shorthand)")
	flag.BoolVar(&cfg.noColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&showVersion, "version", false, "Print version")
	flag.BoolVar(&showVersion, "V", false, "Print version (shorthand)")

	flag.Usage = usage
	flag.Parse()

	if showVersion {
		fmt.Printf("woeusb-go %s\n", version)
		return nil
	}

	if !cfg.device && !cfg.partition {
		fmt.Fprintln(os.Stderr, "Error: You must specify --device or --partition")
		usage()
		os.Exit(1)
	}

	if cfg.device && cfg.partition {
		fmt.Fprintln(os.Stderr, "Error: --device and --partition are mutually exclusive")
		usage()
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Error: source and target are required")
		usage()
		os.Exit(1)
	}

	cfg.source = args[0]
	cfg.target = args[1]

	return &cfg
}

func getMode(cfg *config) string {
	if cfg.device {
		return "device"
	}
	return "partition"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: woeusb-go [--device | --partition] [options] <source> <target>\n\n")
	fmt.Fprintf(os.Stderr, "Create a bootable Windows USB drive from an ISO or DVD.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func checkDependencies() error {
	_, err := deps.CheckDependencies()
	return err
}

func validateInputs(cfg *config) error {
	// Validate source
	if err := validation.ValidateSource(cfg.source); err != nil {
		return fmt.Errorf("source validation failed: %v", err)
	}

	// Validate target
	if err := validation.ValidateTarget(cfg.target, getMode(cfg)); err != nil {
		return fmt.Errorf("target validation failed: %v", err)
	}

	// Check if target is busy
	if err := mount.CheckNotBusy(cfg.target); err != nil {
		return fmt.Errorf("target busy check failed: %v", err)
	}

	return nil
}

func executeDeviceMode(cfg *config, sess *session.Session) error {
	fmt.Printf("Starting device mode: %s -> %s\n", cfg.source, cfg.target)

	// Mount source
	srcMount, err := mountSource(cfg.source)
	if err != nil {
		return fmt.Errorf("failed to mount source: %v", err)
	}
	defer func() { _ = mount.CleanupMountpoint(srcMount) }()

	// Always use FAT32 for maximum UEFI compatibility
	// Large WIM files will be split automatically
	cfg.filesystem = "FAT"

	// Create bootable FAT32 partition
	if err := partition.CreateBootablePartition(cfg.target, cfg.filesystem); err != nil {
		return fmt.Errorf("failed to create bootable partition: %v", err)
	}
	mainPartition := partition.GetPartitionPath(cfg.target)

	// Format the partition as FAT32
	if err := filesystem.FormatPartition(mainPartition, cfg.filesystem, cfg.label); err != nil {
		return fmt.Errorf("failed to format partition: %v", err)
	}

	// Mount target partition
	dstMount, err := mount.MountDevice(mainPartition, "vfat")
	if err != nil {
		return fmt.Errorf("failed to mount target partition: %v", err)
	}
	defer func() { _ = mount.CleanupMountpoint(dstMount) }()

	// Copy files with automatic WIM splitting for files > 4GB
	fmt.Println("Copying Windows files...")
	if err := copy.CopyWindowsISOWithWIMSplit(srcMount, dstMount, copy.PrintProgress); err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}

	// Set boot flag if requested (helps with some buggy BIOSes)
	if cfg.biosBootFlag {
		if err := partition.SetBootFlag(cfg.target, 1); err != nil {
			return fmt.Errorf("failed to set boot flag: %v", err)
		}
	}

	// Install GRUB for legacy BIOS support if not skipped
	if !cfg.skipGrub {
		dependencies, _ := deps.CheckDependencies()
		if dependencies.GrubCmd != "" {
			if err := bootloader.InstallGRUBWithConfig(dstMount, cfg.target, dependencies.GrubCmd); err != nil {
				// GRUB failure is non-fatal for UEFI-only systems
				fmt.Printf("Warning: GRUB installation failed (UEFI boot will still work): %v\n", err)
			}
		}
	}

	return nil
}

func executePartitionMode(cfg *config, sess *session.Session) error {
	fmt.Printf("Starting partition mode: %s -> %s\n", cfg.source, cfg.target)

	// Mount source
	srcMount, err := mountSource(cfg.source)
	if err != nil {
		return fmt.Errorf("failed to mount source: %v", err)
	}
	defer func() { _ = mount.CleanupMountpoint(srcMount) }()

	// Always use FAT32 for maximum compatibility
	cfg.filesystem = "FAT"

	// Format the partition as FAT32
	if err := filesystem.FormatPartition(cfg.target, cfg.filesystem, cfg.label); err != nil {
		return fmt.Errorf("failed to format partition: %v", err)
	}

	// Mount target partition
	dstMount, err := mount.MountDevice(cfg.target, "vfat")
	if err != nil {
		return fmt.Errorf("failed to mount target partition: %v", err)
	}
	defer func() { _ = mount.CleanupMountpoint(dstMount) }()

	// Copy files with automatic WIM splitting
	fmt.Println("Copying Windows files...")
	if err := copy.CopyWindowsISOWithWIMSplit(srcMount, dstMount, copy.PrintProgress); err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}

	return nil
}

func mountSource(source string) (string, error) {
	// Check if source is an ISO file or block device
	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}

	if info.Mode().IsRegular() {
		// ISO file
		return mount.MountISO(source)
	} else {
		// Block device - detect filesystem
		return mount.MountDevice(source, "auto")
	}
}

func init() {
	// Setup signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived interrupt signal, cleaning up...")
		os.Exit(1)
	}()
}
