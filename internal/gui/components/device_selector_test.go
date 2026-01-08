package components

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// BlockDeviceTestData represents generated block device data for property testing
type BlockDeviceTestData struct {
	Devices []BlockDevice
}

// Generate implements quick.Generator for BlockDeviceTestData
func (BlockDeviceTestData) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate 1-10 random block devices
	numDevices := r.Intn(10) + 1
	devices := make([]BlockDevice, numDevices)

	deviceTypes := []string{"disk", "part", "loop", "rom"}
	transports := []string{"usb", "sata", "nvme", "ata", "scsi", ""}
	removableValues := []string{"0", "1", "true", "false", ""}
	models := []string{"SanDisk Cruzer", "Kingston DataTraveler", "Samsung T7", "WD Elements", ""}

	for i := 0; i < numDevices; i++ {
		devices[i] = BlockDevice{
			Name:  generateDeviceName(r),
			Size:  generateSize(r),
			Type:  deviceTypes[r.Intn(len(deviceTypes))],
			Rm:    removableValues[r.Intn(len(removableValues))],
			Tran:  transports[r.Intn(len(transports))],
			Model: models[r.Intn(len(models))],
		}
	}

	return reflect.ValueOf(BlockDeviceTestData{Devices: devices})
}

func generateDeviceName(r *rand.Rand) string {
	prefixes := []string{"sda", "sdb", "sdc", "nvme0n1", "loop0", "sr0"}
	return prefixes[r.Intn(len(prefixes))]
}

func generateSize(r *rand.Rand) string {
	sizes := []string{"16G", "32G", "64G", "128G", "256G", "500M", "1T", "2T"}
	return sizes[r.Intn(len(sizes))]
}

// Property 1: USB Device Filtering
// For any set of block devices returned by lsblk, the GetUSBDevices function SHALL
// return only devices where removable=true AND tran=usb, excluding all devices
// with tran of "sata", "nvme", or "ata".
// **Validates: Requirements 2.1, 2.2**
func TestProperty1_USBDeviceFiltering(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	property := func(data BlockDeviceTestData) bool {
		result := FilterUSBDevices(data.Devices)

		// Build a map of device name to all matching block devices
		// (handles duplicate names in generated data)
		devicesByPath := make(map[string][]BlockDevice)
		for _, dev := range data.Devices {
			path := "/dev/" + dev.Name
			devicesByPath[path] = append(devicesByPath[path], dev)
		}

		// Verify all returned devices meet the USB criteria
		// At least one device with that path must be a valid USB device
		for _, usbDev := range result {
			devs, exists := devicesByPath[usbDev.Path]
			if !exists {
				t.Logf("Returned device %s not found in input", usbDev.Path)
				return false
			}

			// Check that at least one device with this path is a valid USB device
			foundValid := false
			for _, dev := range devs {
				if dev.Type == "disk" && dev.IsRemovable() && dev.Tran == "usb" && !excludedTransports[dev.Tran] {
					foundValid = true
					break
				}
			}

			if !foundValid {
				t.Logf("Device %s returned but no valid USB device found with that path", usbDev.Path)
				return false
			}
		}

		// Verify no valid USB devices were missed
		// Count expected USB devices (unique paths)
		expectedPaths := make(map[string]bool)
		for _, dev := range data.Devices {
			if dev.Type == "disk" && dev.IsRemovable() && dev.Tran == "usb" && !excludedTransports[dev.Tran] {
				expectedPaths["/dev/"+dev.Name] = true
			}
		}

		// Check all expected paths are in result
		resultPaths := make(map[string]bool)
		for _, usbDev := range result {
			resultPaths[usbDev.Path] = true
		}

		for path := range expectedPaths {
			if !resultPaths[path] {
				t.Logf("Valid USB device %s was not included in result", path)
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 failed: %v", err)
	}
}

// TestFilterUSBDevices_ExcludesNonUSB tests that non-USB devices are excluded
func TestFilterUSBDevices_ExcludesNonUSB(t *testing.T) {
	devices := []BlockDevice{
		{Name: "sda", Size: "500G", Type: "disk", Rm: "0", Tran: "sata", Model: "Internal HDD"},
		{Name: "nvme0n1", Size: "1T", Type: "disk", Rm: "0", Tran: "nvme", Model: "NVMe SSD"},
		{Name: "sdb", Size: "16G", Type: "disk", Rm: "1", Tran: "usb", Model: "USB Flash"},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 1 {
		t.Errorf("Expected 1 USB device, got %d", len(result))
	}

	if len(result) > 0 && result[0].Path != "/dev/sdb" {
		t.Errorf("Expected /dev/sdb, got %s", result[0].Path)
	}
}

// TestFilterUSBDevices_ExcludesPartitions tests that partitions are excluded
func TestFilterUSBDevices_ExcludesPartitions(t *testing.T) {
	devices := []BlockDevice{
		{Name: "sdb", Size: "16G", Type: "disk", Rm: "1", Tran: "usb", Model: "USB Flash"},
		{Name: "sdb1", Size: "16G", Type: "part", Rm: "1", Tran: "usb", Model: ""},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 1 {
		t.Errorf("Expected 1 USB device (disk only), got %d", len(result))
	}

	if len(result) > 0 && result[0].Path != "/dev/sdb" {
		t.Errorf("Expected /dev/sdb, got %s", result[0].Path)
	}
}

// TestFilterUSBDevices_ExcludesNonRemovable tests that non-removable devices are excluded
func TestFilterUSBDevices_ExcludesNonRemovable(t *testing.T) {
	devices := []BlockDevice{
		{Name: "sda", Size: "500G", Type: "disk", Rm: "0", Tran: "usb", Model: "USB HDD"},
		{Name: "sdb", Size: "16G", Type: "disk", Rm: "1", Tran: "usb", Model: "USB Flash"},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 1 {
		t.Errorf("Expected 1 USB device (removable only), got %d", len(result))
	}

	if len(result) > 0 && result[0].Path != "/dev/sdb" {
		t.Errorf("Expected /dev/sdb, got %s", result[0].Path)
	}
}

// TestFilterUSBDevices_ExcludesSATA tests that SATA devices are excluded
func TestFilterUSBDevices_ExcludesSATA(t *testing.T) {
	devices := []BlockDevice{
		{Name: "sda", Size: "500G", Type: "disk", Rm: "1", Tran: "sata", Model: "SATA Drive"},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 0 {
		t.Errorf("Expected 0 USB devices (SATA excluded), got %d", len(result))
	}
}

// TestFilterUSBDevices_ExcludesNVMe tests that NVMe devices are excluded
func TestFilterUSBDevices_ExcludesNVMe(t *testing.T) {
	devices := []BlockDevice{
		{Name: "nvme0n1", Size: "1T", Type: "disk", Rm: "1", Tran: "nvme", Model: "NVMe SSD"},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 0 {
		t.Errorf("Expected 0 USB devices (NVMe excluded), got %d", len(result))
	}
}

// TestFilterUSBDevices_ExcludesATA tests that ATA devices are excluded
func TestFilterUSBDevices_ExcludesATA(t *testing.T) {
	devices := []BlockDevice{
		{Name: "sda", Size: "500G", Type: "disk", Rm: "1", Tran: "ata", Model: "ATA Drive"},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 0 {
		t.Errorf("Expected 0 USB devices (ATA excluded), got %d", len(result))
	}
}

// TestFilterUSBDevices_EmptyInput tests handling of empty input
func TestFilterUSBDevices_EmptyInput(t *testing.T) {
	result := FilterUSBDevices([]BlockDevice{})

	if len(result) != 0 {
		t.Errorf("Expected 0 USB devices for empty input, got %d", len(result))
	}
}

// TestFilterUSBDevices_MultipleUSBDevices tests multiple USB devices
func TestFilterUSBDevices_MultipleUSBDevices(t *testing.T) {
	devices := []BlockDevice{
		{Name: "sdb", Size: "16G", Type: "disk", Rm: "1", Tran: "usb", Model: "USB Flash 1"},
		{Name: "sdc", Size: "32G", Type: "disk", Rm: "1", Tran: "usb", Model: "USB Flash 2"},
		{Name: "sdd", Size: "64G", Type: "disk", Rm: "1", Tran: "usb", Model: "USB Flash 3"},
	}

	result := FilterUSBDevices(devices)

	if len(result) != 3 {
		t.Errorf("Expected 3 USB devices, got %d", len(result))
	}
}

// TestParseLsblkOutput tests JSON parsing
func TestParseLsblkOutput(t *testing.T) {
	jsonData := []byte(`{
		"blockdevices": [
			{"name": "sda", "size": "500G", "type": "disk", "rm": "0", "tran": "sata", "model": "Internal HDD"},
			{"name": "sdb", "size": "16G", "type": "disk", "rm": "1", "tran": "usb", "model": "USB Flash"}
		]
	}`)

	result, err := ParseLsblkOutput(jsonData)
	if err != nil {
		t.Fatalf("ParseLsblkOutput failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 USB device, got %d", len(result))
	}

	if len(result) > 0 {
		if result[0].Path != "/dev/sdb" {
			t.Errorf("Expected /dev/sdb, got %s", result[0].Path)
		}
		if result[0].Name != "USB Flash" {
			t.Errorf("Expected 'USB Flash', got %s", result[0].Name)
		}
	}
}

// TestParseLsblkOutput_InvalidJSON tests handling of invalid JSON
func TestParseLsblkOutput_InvalidJSON(t *testing.T) {
	_, err := ParseLsblkOutput([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestParseSizeToBytes tests size parsing
func TestParseSizeToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"16G", 16 * 1024 * 1024 * 1024},
		{"500M", 500 * 1024 * 1024},
		{"1T", 1024 * 1024 * 1024 * 1024},
		{"1024K", 1024 * 1024},
		{"512B", 512},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseSizeToBytes(tt.input)
			if result != tt.expected {
				t.Errorf("parseSizeToBytes(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsRemovable tests removable field parsing
func TestIsRemovable(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"0", false},
		{"true", true},
		{"false", false},
		{"TRUE", true},
		{"FALSE", false},
		{"", false},
		{" 1 ", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isRemovable(tt.input)
			if result != tt.expected {
				t.Errorf("isRemovable(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// USBDeviceTestData represents generated USB device data for property testing
type USBDeviceTestData struct {
	Path      string
	Name      string
	Size      int64
	SizeHuman string
}

// Generate implements quick.Generator for USBDeviceTestData
func (USBDeviceTestData) Generate(r *rand.Rand, size int) reflect.Value {
	paths := []string{"/dev/sda", "/dev/sdb", "/dev/sdc", "/dev/sdd", "/dev/sde"}
	names := []string{"SanDisk Cruzer", "Kingston DataTraveler", "Samsung T7", "WD Elements", "Lexar JumpDrive", ""}
	sizes := []string{"16G", "32G", "64G", "128G", "256G", "500M", "1T"}

	sizeHuman := sizes[r.Intn(len(sizes))]

	return reflect.ValueOf(USBDeviceTestData{
		Path:      paths[r.Intn(len(paths))],
		Name:      names[r.Intn(len(names))],
		Size:      parseSizeToBytes(sizeHuman),
		SizeHuman: sizeHuman,
	})
}

// Property 9: Device Display Information
// For any USB device, the rendered display string SHALL contain the device path,
// human-readable size, and device name/model.
// **Validates: Requirements 2.3**
func TestProperty9_DeviceDisplayInformation(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	property := func(data USBDeviceTestData) bool {
		dev := USBDevice{
			Path:      data.Path,
			Name:      data.Name,
			Size:      data.Size,
			SizeHuman: data.SizeHuman,
			Removable: true,
			Transport: "usb",
		}

		result := FormatDeviceDisplay(dev)

		// Check: result must contain the device path
		if !containsString(result, data.Path) {
			t.Logf("Display string %q does not contain path %q", result, data.Path)
			return false
		}

		// Check: result must contain the human-readable size
		if !containsString(result, data.SizeHuman) {
			t.Logf("Display string %q does not contain size %q", result, data.SizeHuman)
			return false
		}

		// Check: result must contain the device name/model (or "Unknown Device" if empty)
		expectedName := data.Name
		if expectedName == "" {
			expectedName = "Unknown Device"
		}
		if !containsString(result, expectedName) {
			t.Logf("Display string %q does not contain name %q", result, expectedName)
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 failed: %v", err)
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestFormatDeviceDisplay_WithModel tests formatting with a model name
func TestFormatDeviceDisplay_WithModel(t *testing.T) {
	dev := USBDevice{
		Path:      "/dev/sdb",
		Name:      "SanDisk Cruzer",
		Size:      16 * 1024 * 1024 * 1024,
		SizeHuman: "16G",
		Removable: true,
		Transport: "usb",
	}

	result := FormatDeviceDisplay(dev)
	expected := "/dev/sdb - 16G (SanDisk Cruzer)"

	if result != expected {
		t.Errorf("FormatDeviceDisplay() = %q, want %q", result, expected)
	}
}

// TestFormatDeviceDisplay_WithoutModel tests formatting without a model name
func TestFormatDeviceDisplay_WithoutModel(t *testing.T) {
	dev := USBDevice{
		Path:      "/dev/sdc",
		Name:      "",
		Size:      32 * 1024 * 1024 * 1024,
		SizeHuman: "32G",
		Removable: true,
		Transport: "usb",
	}

	result := FormatDeviceDisplay(dev)
	expected := "/dev/sdc - 32G (Unknown Device)"

	if result != expected {
		t.Errorf("FormatDeviceDisplay() = %q, want %q", result, expected)
	}
}

// TestFormatDeviceDisplay_LargeSize tests formatting with large sizes
func TestFormatDeviceDisplay_LargeSize(t *testing.T) {
	dev := USBDevice{
		Path:      "/dev/sdd",
		Name:      "WD Elements",
		Size:      1024 * 1024 * 1024 * 1024,
		SizeHuman: "1T",
		Removable: true,
		Transport: "usb",
	}

	result := FormatDeviceDisplay(dev)
	expected := "/dev/sdd - 1T (WD Elements)"

	if result != expected {
		t.Errorf("FormatDeviceDisplay() = %q, want %q", result, expected)
	}
}
