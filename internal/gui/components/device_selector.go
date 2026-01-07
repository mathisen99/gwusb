// Package components provides reusable GUI widgets for the WoeUSB-go application.
package components

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// USBDevice represents a USB storage device
type USBDevice struct {
	Path      string // e.g., /dev/sdb
	Name      string // e.g., "SanDisk Cruzer"
	Size      int64  // Size in bytes
	SizeHuman string // e.g., "16 GB"
	Removable bool   // Must be true for USB
	Transport string // Transport type (usb, sata, nvme, etc.)
}

// LsblkOutput represents the JSON output from lsblk command
type LsblkOutput struct {
	Blockdevices []BlockDevice `json:"blockdevices"`
}

// BlockDevice represents a block device from lsblk output
type BlockDevice struct {
	Name     string        `json:"name"`
	Size     string        `json:"size"`
	Type     string        `json:"type"` // "disk" or "part"
	Rm       string        `json:"rm"`   // "1" for removable, "0" for non-removable
	Tran     string        `json:"tran"` // "usb" for USB devices
	Model    string        `json:"model"`
	Children []BlockDevice `json:"children,omitempty"`
}

// excludedTransports lists transport types that should be excluded
var excludedTransports = map[string]bool{
	"sata": true,
	"nvme": true,
	"ata":  true,
}

// GetUSBDevices returns only removable USB devices by parsing lsblk JSON output
func GetUSBDevices() ([]USBDevice, error) {
	return GetUSBDevicesWithRunner(defaultCommandRunner{})
}

// CommandRunner interface for executing commands (allows testing)
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// defaultCommandRunner implements CommandRunner using os/exec
type defaultCommandRunner struct{}

func (d defaultCommandRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// GetUSBDevicesWithRunner returns USB devices using a custom command runner
func GetUSBDevicesWithRunner(runner CommandRunner) ([]USBDevice, error) {
	output, err := runner.Run("lsblk", "-J", "-o", "NAME,SIZE,TYPE,RM,TRAN,MODEL")
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk: %w", err)
	}

	return ParseLsblkOutput(output)
}

// ParseLsblkOutput parses lsblk JSON output and filters for USB devices
func ParseLsblkOutput(jsonData []byte) ([]USBDevice, error) {
	var lsblkOut LsblkOutput
	if err := json.Unmarshal(jsonData, &lsblkOut); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	return FilterUSBDevices(lsblkOut.Blockdevices), nil
}

// FilterUSBDevices filters block devices to return only USB devices
// Criteria: type=disk, removable=true, tran=usb, not in excluded transports
func FilterUSBDevices(devices []BlockDevice) []USBDevice {
	var usbDevices []USBDevice

	for _, dev := range devices {
		if IsUSBBlockDevice(dev) {
			usbDevices = append(usbDevices, BlockDeviceToUSBDevice(dev))
		}
	}

	return usbDevices
}

// IsUSBBlockDevice checks if a block device is a removable USB device
// Returns true if:
// - type is "disk"
// - rm (removable) is "1" or "true"
// - tran (transport) is "usb"
// - tran is NOT in excluded transports (sata, nvme, ata)
func IsUSBBlockDevice(dev BlockDevice) bool {
	// Must be a disk (not a partition)
	if dev.Type != "disk" {
		return false
	}

	// Must be removable
	if !isRemovable(dev.Rm) {
		return false
	}

	// Must be USB transport
	if strings.ToLower(dev.Tran) != "usb" {
		return false
	}

	// Must not be an excluded transport type
	if excludedTransports[strings.ToLower(dev.Tran)] {
		return false
	}

	return true
}

// isRemovable checks if the removable field indicates a removable device
func isRemovable(rm string) bool {
	rm = strings.TrimSpace(rm)
	return rm == "1" || strings.ToLower(rm) == "true"
}

// BlockDeviceToUSBDevice converts a BlockDevice to a USBDevice
func BlockDeviceToUSBDevice(dev BlockDevice) USBDevice {
	return USBDevice{
		Path:      "/dev/" + dev.Name,
		Name:      strings.TrimSpace(dev.Model),
		Size:      parseSizeToBytes(dev.Size),
		SizeHuman: dev.Size,
		Removable: isRemovable(dev.Rm),
		Transport: dev.Tran,
	}
}

// parseSizeToBytes converts human-readable size (e.g., "16G", "500M") to bytes
func parseSizeToBytes(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0
	}

	// Handle sizes like "14.5G", "500M", "1T"
	multipliers := map[byte]int64{
		'B': 1,
		'K': 1024,
		'M': 1024 * 1024,
		'G': 1024 * 1024 * 1024,
		'T': 1024 * 1024 * 1024 * 1024,
	}

	lastChar := sizeStr[len(sizeStr)-1]
	multiplier, hasMultiplier := multipliers[lastChar]
	if !hasMultiplier {
		// Try parsing as plain number
		val, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return 0
		}
		return val
	}

	numStr := sizeStr[:len(sizeStr)-1]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	return int64(val * float64(multiplier))
}

// FormatDeviceDisplay formats a USB device for display in the UI
// Returns a string containing device path, size, and model
func FormatDeviceDisplay(dev USBDevice) string {
	name := dev.Name
	if name == "" {
		name = "Unknown Device"
	}
	return fmt.Sprintf("%s - %s (%s)", dev.Path, dev.SizeHuman, name)
}
