// Package components provides reusable GUI widgets for the WoeUSB-go application.
package components

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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
	Rm       interface{}   `json:"rm"`   // Can be bool or string depending on lsblk version
	Tran     string        `json:"tran"` // "usb" for USB devices
	Model    string        `json:"model"`
	Children []BlockDevice `json:"children,omitempty"`
}

// IsRemovable returns true if the device is marked as removable
func (bd BlockDevice) IsRemovable() bool {
	return isRemovableValue(bd.Rm)
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
// - rm (removable) is true/"1"/"true"
// - tran (transport) is "usb"
// - tran is NOT in excluded transports (sata, nvme, ata)
func IsUSBBlockDevice(dev BlockDevice) bool {
	// Must be a disk (not a partition)
	if dev.Type != "disk" {
		return false
	}

	// Must be removable
	if !dev.IsRemovable() {
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

// isRemovableValue checks if the removable field indicates a removable device
// Handles both string ("1", "true") and bool (true) values
func isRemovableValue(rm interface{}) bool {
	if rm == nil {
		return false
	}
	switch v := rm.(type) {
	case bool:
		return v
	case string:
		v = strings.TrimSpace(v)
		return v == "1" || strings.ToLower(v) == "true"
	default:
		return false
	}
}

// isRemovable checks if the removable field indicates a removable device (string version for tests)
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
		Removable: dev.IsRemovable(),
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

// DeviceSelector provides USB device selection as a Fyne widget
type DeviceSelector struct {
	widget.BaseWidget
	devices   []USBDevice
	selected  string
	onSelect  func(device string)
	list      *widget.Select
	container *fyne.Container
	noDevices *widget.Label
}

// NewDeviceSelector creates a new device selector widget
func NewDeviceSelector(onSelect func(device string)) *DeviceSelector {
	ds := &DeviceSelector{
		onSelect: onSelect,
	}

	ds.noDevices = widget.NewLabel("No USB devices detected")
	ds.noDevices.Hide()

	ds.list = widget.NewSelect([]string{}, func(selected string) {
		// Extract device path from display string
		if selected != "" {
			// Format is "/dev/sdX - SIZE (NAME)"
			parts := strings.Split(selected, " - ")
			if len(parts) > 0 {
				ds.selected = parts[0]
				if ds.onSelect != nil {
					ds.onSelect(ds.selected)
				}
			}
		}
	})
	ds.list.PlaceHolder = "Select a USB device..."

	ds.container = container.NewStack(ds.list, ds.noDevices)

	ds.ExtendBaseWidget(ds)
	return ds
}

// CreateRenderer implements fyne.Widget
func (ds *DeviceSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ds.container)
}

// Refresh implements fyne.Widget interface (refreshes the widget display)
func (ds *DeviceSelector) Refresh() {
	ds.BaseWidget.Refresh()
}

// RefreshDevices rescans for USB devices
func (ds *DeviceSelector) RefreshDevices() error {
	devices, err := GetUSBDevices()
	if err != nil {
		return fmt.Errorf("failed to get USB devices: %w", err)
	}

	ds.devices = devices
	ds.updateList()
	return nil
}

// RefreshDevicesWithRunner rescans using a custom command runner (for testing)
func (ds *DeviceSelector) RefreshDevicesWithRunner(runner CommandRunner) error {
	devices, err := GetUSBDevicesWithRunner(runner)
	if err != nil {
		return fmt.Errorf("failed to get USB devices: %w", err)
	}

	ds.devices = devices
	ds.updateList()
	return nil
}

// updateList updates the select widget with current devices
func (ds *DeviceSelector) updateList() {
	if len(ds.devices) == 0 {
		ds.list.Hide()
		ds.noDevices.Show()
		ds.selected = ""
		if ds.onSelect != nil {
			ds.onSelect("")
		}
		return
	}

	ds.noDevices.Hide()
	ds.list.Show()

	options := make([]string, len(ds.devices))
	for i, dev := range ds.devices {
		options[i] = FormatDeviceDisplay(dev)
	}
	ds.list.Options = options
	ds.list.Refresh()
}

// GetSelected returns the currently selected device path
func (ds *DeviceSelector) GetSelected() string {
	return ds.selected
}

// GetDevices returns the list of detected USB devices
func (ds *DeviceSelector) GetDevices() []USBDevice {
	return ds.devices
}

// SetSelected sets the selected device programmatically
func (ds *DeviceSelector) SetSelected(devicePath string) {
	ds.selected = devicePath
	// Find and select the matching option
	for _, dev := range ds.devices {
		if dev.Path == devicePath {
			ds.list.SetSelected(FormatDeviceDisplay(dev))
			break
		}
	}
}
