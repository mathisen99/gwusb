# Design Document: WoeUSB-go v0.2.0 GUI

## Overview

WoeUSB-go v0.2.0 extends the existing CLI tool with a graphical user interface built using the Fyne toolkit. The GUI provides a beginner-friendly way to create bootable Windows USB drives while maintaining the safety feature of only showing USB devices (not internal drives). The design also includes distro-aware dependency checking that provides exact install commands for missing packages.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        cmd/woeusb/main.go                       │
│                    (CLI + GUI entry point)                      │
│                         --gui flag                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      internal/gui/                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   app.go    │  │  window.go  │  │     components/         │ │
│  │ (Fyne app)  │  │ (main win)  │  │  - device_selector.go   │ │
│  └─────────────┘  └─────────────┘  │  - file_browser.go      │ │
│                                     │  - progress_bar.go      │ │
│                                     │  - dependency_dialog.go │ │
│                                     └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   internal/distro/                              │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐  │
│  │   detect.go     │  │         packages.go                 │  │
│  │ (OS detection)  │  │ (distro-specific package names)     │  │
│  └─────────────────┘  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Existing internal/ packages                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │  deps    │ │  device  │ │  copy    │ │ partition│ ...      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. GUI Application (internal/gui/app.go)

```go
// App represents the main GUI application
type App struct {
    fyneApp    fyne.App
    mainWindow fyne.Window
    session    *session.Session
}

// NewApp creates a new GUI application instance
func NewApp() *App

// Run starts the GUI application
func (a *App) Run() error

// CheckDependencies verifies all required tools are installed
func (a *App) CheckDependencies() ([]MissingDep, error)
```

### 2. Main Window (internal/gui/window.go)

```go
// MainWindow represents the primary application window
type MainWindow struct {
    window         fyne.Window
    deviceSelector *DeviceSelector
    fileBrowser    *FileBrowser
    progressBar    *ProgressBar
    startButton    *widget.Button
    statusLabel    *widget.Label
    
    selectedDevice string
    selectedISO    string
}

// NewMainWindow creates the main application window
func NewMainWindow(app fyne.App) *MainWindow

// Show displays the main window
func (w *MainWindow) Show()

// UpdateState updates UI state based on selections
func (w *MainWindow) UpdateState()
```

### 3. Device Selector (internal/gui/components/device_selector.go)

```go
// DeviceSelector provides USB device selection
type DeviceSelector struct {
    widget.BaseWidget
    devices    []USBDevice
    selected   string
    onSelect   func(device string)
    list       *widget.List
    refreshBtn *widget.Button
}

// USBDevice represents a USB storage device
type USBDevice struct {
    Path       string  // e.g., /dev/sdb
    Name       string  // e.g., "SanDisk Cruzer"
    Size       int64   // Size in bytes
    SizeHuman  string  // e.g., "16 GB"
    Removable  bool    // Must be true for USB
}

// NewDeviceSelector creates a new device selector widget
func NewDeviceSelector(onSelect func(device string)) *DeviceSelector

// Refresh rescans for USB devices
func (d *DeviceSelector) Refresh() error

// GetUSBDevices returns only removable USB devices
func GetUSBDevices() ([]USBDevice, error)

// IsUSBDevice checks if a device is a removable USB device
func IsUSBDevice(devicePath string) (bool, error)
```

### 4. File Browser (internal/gui/components/file_browser.go)

```go
// FileBrowser provides ISO file selection
type FileBrowser struct {
    widget.BaseWidget
    selectedPath string
    pathLabel    *widget.Label
    browseBtn    *widget.Button
    onSelect     func(path string)
}

// NewFileBrowser creates a new file browser widget
func NewFileBrowser(onSelect func(path string)) *FileBrowser

// OpenDialog opens the file selection dialog
func (f *FileBrowser) OpenDialog(parent fyne.Window)

// ValidateISO checks if the selected file is a valid ISO
func ValidateISO(path string) error
```

### 5. Progress Bar (internal/gui/components/progress_bar.go)

```go
// ProgressBar displays operation progress
type ProgressBar struct {
    widget.BaseWidget
    bar         *widget.ProgressBar
    statusLabel *widget.Label
    percentage  float64
    status      string
}

// NewProgressBar creates a new progress bar widget
func NewProgressBar() *ProgressBar

// SetProgress updates the progress percentage (0.0 to 1.0)
func (p *ProgressBar) SetProgress(value float64)

// SetStatus updates the status text
func (p *ProgressBar) SetStatus(status string)

// Reset resets the progress bar to initial state
func (p *ProgressBar) Reset()
```

### 6. Dependency Dialog (internal/gui/components/dependency_dialog.go)

```go
// DependencyDialog shows missing dependencies with install instructions
type DependencyDialog struct {
    dialog      dialog.Dialog
    missingDeps []MissingDep
    distro      *distro.Info
}

// MissingDep represents a missing dependency
type MissingDep struct {
    Binary      string // e.g., "wimlib-imagex"
    PackageName string // distro-specific package name
    Required    bool   // true if required, false if optional
}

// NewDependencyDialog creates a dependency dialog
func NewDependencyDialog(parent fyne.Window, missing []MissingDep, distro *distro.Info) *DependencyDialog

// Show displays the dialog
func (d *DependencyDialog) Show()

// GetInstallCommand returns the full install command for the distro
func (d *DependencyDialog) GetInstallCommand() string
```

### 7. Distro Detection (internal/distro/detect.go)

```go
// Info contains detected distribution information
type Info struct {
    ID          string // e.g., "ubuntu", "fedora", "arch"
    IDLike      string // e.g., "debian" for Ubuntu
    Name        string // e.g., "Ubuntu 25.10"
    Version     string // e.g., "25.10"
    PackageManager string // e.g., "apt", "dnf", "pacman", "zypper"
}

// Detect reads /etc/os-release and returns distro info
func Detect() (*Info, error)

// GetPackageManager returns the package manager for the distro
func (i *Info) GetPackageManager() string
```

### 8. Package Mapping (internal/distro/packages.go)

```go
// PackageMap maps binary names to distro-specific package names
type PackageMap struct {
    Binary   string            // e.g., "wimlib-imagex"
    Packages map[string]string // distro ID -> package name
}

// GetPackageName returns the package name for a binary on a distro
func GetPackageName(binary string, distroID string) string

// GetInstallCommand returns the full install command
func GetInstallCommand(distroID string, packages []string) string

// RequiredBinaries lists all required binary dependencies
var RequiredBinaries = []string{
    "wipefs",
    "parted", 
    "lsblk",
    "blockdev",
    "mount",
    "umount",
    "7z",
    "mkdosfs",
    "wimlib-imagex",
}

// OptionalBinaries lists optional binary dependencies
var OptionalBinaries = []string{
    "grub-install",
    "mkntfs",
}
```

## Data Models

### USB Device Detection

The device selector uses `lsblk` to enumerate block devices and filters for USB devices:

```go
// lsblk output parsing
type LsblkOutput struct {
    Blockdevices []BlockDevice `json:"blockdevices"`
}

type BlockDevice struct {
    Name       string        `json:"name"`
    Size       string        `json:"size"`
    Type       string        `json:"type"`      // "disk" or "part"
    Removable  string        `json:"rm"`        // "1" for removable
    Tran       string        `json:"tran"`      // "usb" for USB devices
    Model      string        `json:"model"`
    Children   []BlockDevice `json:"children"`
}
```

USB detection criteria:
1. `type` must be "disk" (not partition)
2. `rm` must be "1" (removable)
3. `tran` must be "usb" (USB transport)
4. Exclude devices with `tran` of "sata", "nvme", "ata"

### Distro Package Mapping

```go
var packageMappings = map[string]map[string]string{
    "wimlib-imagex": {
        "ubuntu":   "wimtools",
        "debian":   "wimtools",
        "linuxmint": "wimtools",
        "fedora":   "wimlib-utils",
        "arch":     "wimlib",
        "opensuse": "wimlib",
        "opensuse-tumbleweed": "wimlib",
        "opensuse-leap": "wimlib",
    },
    "7z": {
        "ubuntu":   "p7zip-full",
        "debian":   "p7zip-full",
        "linuxmint": "p7zip-full",
        "fedora":   "p7zip-plugins",
        "arch":     "p7zip",
        "opensuse": "p7zip-full",
    },
    "mkdosfs": {
        "ubuntu":   "dosfstools",
        "debian":   "dosfstools",
        "linuxmint": "dosfstools",
        "fedora":   "dosfstools",
        "arch":     "dosfstools",
        "opensuse": "dosfstools",
    },
    "grub-install": {
        "ubuntu":   "grub-pc",
        "debian":   "grub-pc",
        "linuxmint": "grub-pc",
        "fedora":   "grub2-pc",
        "arch":     "grub",
        "opensuse": "grub2",
    },
    "mkntfs": {
        "ubuntu":   "ntfs-3g",
        "debian":   "ntfs-3g",
        "linuxmint": "ntfs-3g",
        "fedora":   "ntfs-3g",
        "arch":     "ntfs-3g",
        "opensuse": "ntfs-3g",
    },
}

var installCommands = map[string]string{
    "ubuntu":   "sudo apt install",
    "debian":   "sudo apt install",
    "linuxmint": "sudo apt install",
    "fedora":   "sudo dnf install",
    "arch":     "sudo pacman -S",
    "opensuse": "sudo zypper install",
    "opensuse-tumbleweed": "sudo zypper install",
    "opensuse-leap": "sudo zypper install",
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: USB Device Filtering

*For any* set of block devices returned by lsblk, the GetUSBDevices function SHALL return only devices where `removable=true` AND `tran=usb`, excluding all devices with `tran` of "sata", "nvme", or "ata".

**Validates: Requirements 2.1, 2.2**

### Property 2: Root Privilege Detection

*For any* effective user ID, the root check function SHALL return true if and only if the UID equals 0.

**Validates: Requirements 1.2**

### Property 3: Dependency Binary Detection

*For any* binary name in the required list, the dependency checker SHALL correctly identify whether the binary exists in PATH.

**Validates: Requirements 1.3**

### Property 4: Distro Detection from os-release

*For any* valid /etc/os-release file content, the Detect function SHALL correctly parse and return the ID, ID_LIKE, NAME, and VERSION fields.

**Validates: Requirements 6.1**

### Property 5: Package Name Mapping

*For any* supported distro ID and binary name, GetPackageName SHALL return the correct distro-specific package name as defined in the package mapping.

**Validates: Requirements 6.2**

### Property 6: Install Command Generation

*For any* supported distro ID and list of package names, GetInstallCommand SHALL return a valid install command using the correct package manager prefix.

**Validates: Requirements 6.3**

### Property 7: Start Button State

*For any* combination of device selection (empty or set) and ISO selection (empty or set), the Start button SHALL be enabled if and only if both selections are non-empty.

**Validates: Requirements 4.1**

### Property 8: Progress Bar Updates

*For any* progress value between 0.0 and 1.0, SetProgress SHALL update the progress bar to display that percentage.

**Validates: Requirements 5.1**

### Property 9: Device Display Information

*For any* USB device, the rendered display string SHALL contain the device path, human-readable size, and device name/model.

**Validates: Requirements 2.3**

### Property 10: ISO File Validation

*For any* file path, ValidateISO SHALL return nil if the file exists and is readable, and an error otherwise.

**Validates: Requirements 3.3**

### Property 11: UI Controls Disabled During Operation

*For any* operation state (in_progress or idle), the Start button and selection controls SHALL be disabled if and only if the state is in_progress.

**Validates: Requirements 4.5**

## Error Handling

### Dependency Errors
- Missing required dependencies: Show DependencyDialog with install instructions
- Unknown distro: Show generic package names with fallback message
- Cannot read /etc/os-release: Fall back to generic instructions

### Device Errors
- No USB devices found: Display "No USB devices detected" message
- Device becomes unavailable during operation: Abort with error message
- Permission denied: Prompt user to run as root

### File Errors
- ISO file not found: Display error in file browser
- ISO file not readable: Display permission error
- Invalid ISO format: Display validation error

### Operation Errors
- Write failure: Display error with details, offer retry
- Interrupted by user: Clean up and display cancellation message
- Disk full: Display space error with required size

## Testing Strategy

### Unit Tests
- Distro detection parsing (various os-release formats)
- Package name mapping for all supported distros
- Install command generation
- USB device filtering logic
- Progress bar state management
- Button enable/disable logic

### Property-Based Tests
- USB device filtering (Property 1)
- Distro detection (Property 4)
- Package name mapping (Property 5)
- Install command generation (Property 6)
- Start button state logic (Property 7)

### Integration Tests
- Full dependency check flow
- Device enumeration on real system
- GUI launch and basic interaction

### Manual Tests
- X11 and Wayland compatibility
- Real USB device detection
- Full write operation end-to-end
