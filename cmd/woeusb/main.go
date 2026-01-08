package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/mathisen/woeusb-go/internal/bootloader"
	filecopy "github.com/mathisen/woeusb-go/internal/copy"
	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/filesystem"
	"github.com/mathisen/woeusb-go/internal/gui"
	"github.com/mathisen/woeusb-go/internal/mount"
	"github.com/mathisen/woeusb-go/internal/output"
	"github.com/mathisen/woeusb-go/internal/partition"
	"github.com/mathisen/woeusb-go/internal/session"
	"github.com/mathisen/woeusb-go/internal/validation"
)

const version = "1.0.0"

type config struct {
	device       bool
	partition    bool
	filesystem   string
	label        string
	biosBootFlag bool
	skipGrub     bool
	verbose      bool
	noColor      bool
	guiMode      bool
	source       string
	target       string
}

func main() {
	cfg := parseArgs()
	if cfg == nil {
		return
	}

	// Setup output options
	output.SetNoColor(cfg.noColor)
	output.SetVerbose(cfg.verbose)

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

	// Print header
	output.Step("WoeUSB-go v%s", version)
	output.Verbose("Source: %s", cfg.source)
	output.Verbose("Target: %s", cfg.target)
	output.Verbose("Filesystem: %s, Label: %s", cfg.filesystem, cfg.label)

	// Check dependencies
	output.Step("Checking dependencies...")
	if err := checkDependencies(); err != nil {
		output.Error("Dependency check failed: %v", err)
		os.Exit(1)
	}
	output.Info("All dependencies found")

	// Validate source and target
	output.Step("Validating source and target...")
	if err := validateInputs(cfg); err != nil {
		output.Error("Validation failed: %v", err)
		os.Exit(1)
	}
	output.Info("Validation passed")

	// Execute the appropriate mode
	var err error
	if cfg.device {
		err = executeDeviceMode(cfg, sess)
	} else {
		err = executePartitionMode(cfg, sess)
	}

	if err != nil {
		output.Error("%v", err)
		os.Exit(1)
	}

	output.Success("WoeUSB operation completed successfully!")
	output.Info("You may now safely remove the USB device")
}

func parseArgs() *config {
	var cfg config
	var showVersion bool
	var checkDepsOnly bool

	flag.BoolVar(&cfg.device, "device", false, "Wipe entire device and create bootable USB")
	flag.BoolVar(&cfg.device, "d", false, "Wipe entire device (shorthand)")
	flag.BoolVar(&cfg.partition, "partition", false, "Use existing partition")
	flag.BoolVar(&cfg.partition, "p", false, "Use existing partition (shorthand)")
	flag.BoolVar(&checkDepsOnly, "check-deps", false, "Check if all required dependencies are installed and exit")
	flag.BoolVar(&cfg.guiMode, "gui", false, "Launch graphical user interface")
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

	// Handle --check-deps flag
	if checkDepsOnly {
		runDependencyCheck()
		return nil
	}

	// Handle --gui flag
	if cfg.guiMode {
		runGUI()
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

// runGUI launches the graphical user interface
func runGUI() {
	app := gui.NewApp()
	if err := app.Run(); err != nil {
		output.Error("GUI error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// runDependencyCheck checks all dependencies and prints detailed status
func runDependencyCheck() {
	output.Step("Checking system dependencies...")

	allFound := true

	// Required tools
	requiredTools := []struct {
		name string
		pkg  string
		cmds []string
	}{
		{"wipefs", "util-linux", []string{"wipefs"}},
		{"parted", "parted", []string{"parted"}},
		{"lsblk", "util-linux", []string{"lsblk"}},
		{"blockdev", "util-linux", []string{"blockdev"}},
		{"mount", "util-linux", []string{"mount"}},
		{"umount", "util-linux", []string{"umount"}},
		{"7z", "p7zip-full / p7zip", []string{"7z"}},
		{"mkdosfs", "dosfstools", []string{"mkdosfs", "mkfs.vfat", "mkfs.fat"}},
		{"wimlib-imagex", "wimlib / wimtools", []string{"wimlib-imagex"}},
	}

	for _, tool := range requiredTools {
		found := false
		var foundPath string
		for _, cmd := range tool.cmds {
			if path, err := exec.LookPath(cmd); err == nil {
				found = true
				foundPath = path
				break
			}
		}
		if found {
			output.Info("%s: found at %s", tool.name, foundPath)
		} else {
			output.Error("%s: NOT FOUND (install package: %s)", tool.name, tool.pkg)
			allFound = false
		}
	}

	// Optional tools
	optionalTools := []struct {
		name    string
		pkg     string
		cmds    []string
		purpose string
	}{
		{"grub-install", "grub2 / grub-pc", []string{"grub-install", "grub2-install"}, "legacy BIOS boot"},
		{"mkntfs", "ntfs-3g / ntfsprogs", []string{"mkntfs"}, "NTFS filesystem support"},
	}

	output.Step("Checking optional dependencies...")
	for _, tool := range optionalTools {
		found := false
		var foundPath string
		for _, cmd := range tool.cmds {
			if path, err := exec.LookPath(cmd); err == nil {
				found = true
				foundPath = path
				break
			}
		}
		if found {
			output.Info("%s: found at %s", tool.name, foundPath)
		} else {
			output.Warning("%s: not found (needed for %s, install: %s)", tool.name, tool.purpose, tool.pkg)
		}
	}

	fmt.Println()
	if allFound {
		output.Success("All required dependencies are installed!")
		os.Exit(0)
	} else {
		output.Error("Some required dependencies are missing. Please install them before using woeusb-go.")
		os.Exit(1)
	}
}

func getMode(cfg *config) string {
	if cfg.device {
		return "device"
	}
	return "partition"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: woeusb-go [--device | --partition] [options] <source> <target>\n")
	fmt.Fprintf(os.Stderr, "       woeusb-go --gui\n\n")
	fmt.Fprintf(os.Stderr, "Create a bootable Windows USB drive from an ISO or DVD.\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  woeusb-go --device /path/to/windows.iso /dev/sdX\n")
	fmt.Fprintf(os.Stderr, "  woeusb-go --partition /path/to/windows.iso /dev/sdX1\n")
	fmt.Fprintf(os.Stderr, "  woeusb-go --gui\n")
	fmt.Fprintf(os.Stderr, "  woeusb-go --check-deps\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func checkDependencies() error {
	_, err := deps.CheckDependencies()
	return err
}

func validateInputs(cfg *config) error {
	if err := validation.ValidateSource(cfg.source); err != nil {
		return fmt.Errorf("source validation failed: %v", err)
	}

	if err := validation.ValidateTarget(cfg.target, getMode(cfg)); err != nil {
		return fmt.Errorf("target validation failed: %v", err)
	}

	if err := mount.CheckNotBusy(cfg.target); err != nil {
		return fmt.Errorf("target busy check failed: %v", err)
	}

	return nil
}

func executeDeviceMode(cfg *config, sess *session.Session) error {
	output.Step("Mounting source ISO...")
	srcMount, err := mountSource(cfg.source)
	if err != nil {
		return fmt.Errorf("failed to mount source: %v", err)
	}
	sess.SourceMount = srcMount
	output.Info("Source mounted at %s", srcMount)

	// Default to FAT if not specified
	if cfg.filesystem == "" {
		cfg.filesystem = "FAT"
	}

	output.Step("Wiping device %s...", cfg.target)
	output.Notice("This will destroy ALL data on the device!")
	if err := partition.CreateBootablePartition(cfg.target, cfg.filesystem); err != nil {
		return fmt.Errorf("failed to create bootable partition: %v", err)
	}
	output.Info("Partition table created")

	mainPartition := partition.GetPartitionPath(cfg.target)
	output.Verbose("Main partition: %s", mainPartition)

	output.Step("Formatting partition as %s...", cfg.filesystem)
	if err := filesystem.FormatPartition(mainPartition, cfg.filesystem, cfg.label); err != nil {
		return fmt.Errorf("failed to format partition: %v", err)
	}
	output.Info("Partition formatted with label '%s'", cfg.label)

	output.Step("Mounting target partition...")
	fsType := "vfat"
	if cfg.filesystem == "NTFS" {
		fsType = "ntfs-3g"
	}
	dstMount, err := mount.MountDevice(mainPartition, fsType)
	if err != nil {
		return fmt.Errorf("failed to mount target partition: %v", err)
	}
	sess.TargetMount = dstMount
	output.Info("Target mounted at %s", dstMount)

	output.Step("Copying Windows files...")
	output.Notice("This may take a while depending on USB speed. Do not interrupt!")
	if err := filecopy.CopyWindowsISOWithWIMSplit(srcMount, dstMount, filecopy.PrintProgress); err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}
	output.Info("All files copied successfully")

	if cfg.biosBootFlag {
		output.Step("Setting boot flag for BIOS compatibility...")
		if err := partition.SetBootFlag(cfg.target, 1); err != nil {
			return fmt.Errorf("failed to set boot flag: %v", err)
		}
		output.Info("Boot flag set")
	}

	if !cfg.skipGrub {
		output.Step("Installing GRUB bootloader for legacy BIOS support...")
		dependencies, _ := deps.CheckDependencies()
		if dependencies.GrubCmd != "" {
			if err := bootloader.InstallGRUBWithConfig(dstMount, cfg.target, dependencies.GrubCmd); err != nil {
				output.Warning("GRUB installation failed (UEFI boot will still work): %v", err)
			} else {
				output.Info("GRUB installed successfully")
			}
		} else {
			output.Warning("GRUB not found, skipping legacy BIOS boot support")
		}
	} else {
		output.Verbose("Skipping GRUB installation as requested")
	}

	output.Step("Cleaning up...")
	if err := mount.CleanupMountpoint(dstMount); err != nil {
		output.Warning("Failed to unmount target: %v", err)
	}
	if err := mount.CleanupMountpoint(srcMount); err != nil {
		output.Warning("Failed to unmount source: %v", err)
	}
	sess.SourceMount = ""
	sess.TargetMount = ""
	output.Info("Cleanup complete")

	return nil
}

func executePartitionMode(cfg *config, sess *session.Session) error {
	output.Step("Mounting source ISO...")
	srcMount, err := mountSource(cfg.source)
	if err != nil {
		return fmt.Errorf("failed to mount source: %v", err)
	}
	sess.SourceMount = srcMount
	output.Info("Source mounted at %s", srcMount)

	// Default to FAT if not specified
	if cfg.filesystem == "" {
		cfg.filesystem = "FAT"
	}

	output.Step("Formatting partition %s as %s...", cfg.target, cfg.filesystem)
	output.Notice("This will destroy all data on the partition!")
	if err := filesystem.FormatPartition(cfg.target, cfg.filesystem, cfg.label); err != nil {
		return fmt.Errorf("failed to format partition: %v", err)
	}
	output.Info("Partition formatted with label '%s'", cfg.label)

	output.Step("Mounting target partition...")
	fsType := "vfat"
	if cfg.filesystem == "NTFS" {
		fsType = "ntfs-3g"
	}
	dstMount, err := mount.MountDevice(cfg.target, fsType)
	if err != nil {
		return fmt.Errorf("failed to mount target partition: %v", err)
	}
	sess.TargetMount = dstMount
	output.Info("Target mounted at %s", dstMount)

	output.Step("Copying Windows files...")
	output.Notice("This may take a while depending on USB speed. Do not interrupt!")
	if err := filecopy.CopyWindowsISOWithWIMSplit(srcMount, dstMount, filecopy.PrintProgress); err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}
	output.Info("All files copied successfully")

	output.Step("Cleaning up...")
	if err := mount.CleanupMountpoint(dstMount); err != nil {
		output.Warning("Failed to unmount target: %v", err)
	}
	if err := mount.CleanupMountpoint(srcMount); err != nil {
		output.Warning("Failed to unmount source: %v", err)
	}
	sess.SourceMount = ""
	sess.TargetMount = ""
	output.Info("Cleanup complete")

	return nil
}

func mountSource(source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}

	if info.Mode().IsRegular() {
		return mount.MountISO(source)
	}
	return mount.MountDevice(source, "auto")
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		output.Warning("Received interrupt signal, cleaning up...")
		os.Exit(1)
	}()
}
