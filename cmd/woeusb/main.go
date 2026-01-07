package main

import (
	"flag"
	"fmt"
	"os"
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

	fmt.Printf("Mode: %s\n", mode(cfg))
	fmt.Printf("Source: %s\n", cfg.source)
	fmt.Printf("Target: %s\n", cfg.target)
	fmt.Printf("Filesystem: %s\n", cfg.filesystem)
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
		fmt.Println("woeusb", version)
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

func mode(cfg *config) string {
	if cfg.device {
		return "device"
	}
	return "partition"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: woeusb [--device | --partition] [options] <source> <target>\n\n")
	fmt.Fprintf(os.Stderr, "Create a bootable Windows USB drive from an ISO or DVD.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}
