package gui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/mathisen/woeusb-go/internal/bootloader"
	filecopy "github.com/mathisen/woeusb-go/internal/copy"
	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/distro"
	"github.com/mathisen/woeusb-go/internal/filesystem"
	"github.com/mathisen/woeusb-go/internal/gui/components"
	"github.com/mathisen/woeusb-go/internal/mount"
	"github.com/mathisen/woeusb-go/internal/partition"
)

// OperationState represents the current state of the write operation
type OperationState int

const (
	StateIdle OperationState = iota
	StateInProgress
	StateComplete
	StateError
)

// MainWindow represents the primary application window
type MainWindow struct {
	window         fyne.Window
	deviceSelector *components.DeviceSelector
	fileBrowser    *components.FileBrowser
	progressBar    *components.ProgressBar
	startButton    *widget.Button
	refreshButton  *widget.Button
	statusLabel    *widget.Label

	selectedDevice string
	selectedISO    string
	state          OperationState
	distroInfo     *distro.Info
}

// NewMainWindow creates the main application window
func NewMainWindow(app fyne.App, distroInfo *distro.Info) *MainWindow {
	w := &MainWindow{
		window:     app.NewWindow("WoeUSB-go"),
		state:      StateIdle,
		distroInfo: distroInfo,
	}

	w.buildUI()
	w.window.Resize(fyne.NewSize(500, 400))
	w.window.SetMaster()

	return w
}

// buildUI constructs the main window UI
func (w *MainWindow) buildUI() {
	// Device selector section
	deviceLabel := widget.NewLabel("Target USB Device:")
	deviceLabel.TextStyle = fyne.TextStyle{Bold: true}

	w.deviceSelector = components.NewDeviceSelector(func(device string) {
		w.selectedDevice = device
		w.UpdateState()
	})

	w.refreshButton = widget.NewButton("Refresh", func() {
		_ = w.deviceSelector.RefreshDevices()
	})

	deviceSection := container.NewVBox(
		deviceLabel,
		w.deviceSelector,
		w.refreshButton,
	)

	// File browser section
	isoLabel := widget.NewLabel("Windows ISO File:")
	isoLabel.TextStyle = fyne.TextStyle{Bold: true}

	w.fileBrowser = components.NewFileBrowser(func(path string) {
		w.selectedISO = path
		w.UpdateState()
	})
	w.fileBrowser.SetBrowseAction(w.window)

	isoSection := container.NewVBox(
		isoLabel,
		w.fileBrowser,
	)

	// Progress section
	w.progressBar = components.NewProgressBar()

	// Status label
	w.statusLabel = widget.NewLabel("")
	w.statusLabel.Alignment = fyne.TextAlignCenter

	// Start button
	w.startButton = widget.NewButton("Create Bootable USB", w.onStartClicked)
	w.startButton.Importance = widget.HighImportance
	w.startButton.Disable() // Disabled until selections are made

	// Layout
	content := container.NewVBox(
		deviceSection,
		widget.NewSeparator(),
		isoSection,
		widget.NewSeparator(),
		w.progressBar,
		w.statusLabel,
		widget.NewSeparator(),
		w.startButton,
	)

	w.window.SetContent(container.NewPadded(content))

	// Handle window close during operation
	w.window.SetCloseIntercept(w.onCloseRequested)
}

// Show displays the main window
func (w *MainWindow) Show() {
	// Initial device scan
	_ = w.deviceSelector.RefreshDevices()
	w.window.Show()
}

// UpdateState updates UI state based on selections and operation state
func (w *MainWindow) UpdateState() {
	// Start button is enabled only when both device and ISO are selected
	// and no operation is in progress
	canStart := w.selectedDevice != "" && w.selectedISO != "" && w.state == StateIdle

	if canStart {
		w.startButton.Enable()
	} else {
		w.startButton.Disable()
	}

	// Disable controls during operation
	if w.state == StateInProgress {
		w.refreshButton.Disable()
	} else {
		w.refreshButton.Enable()
	}
}

// SetState sets the operation state and updates UI accordingly
func (w *MainWindow) SetState(state OperationState) {
	w.state = state
	w.UpdateState()
}

// GetState returns the current operation state
func (w *MainWindow) GetState() OperationState {
	return w.state
}

// onStartClicked handles the start button click
func (w *MainWindow) onStartClicked() {
	// Show confirmation dialog
	dialog.ShowConfirm(
		"Confirm Write Operation",
		"WARNING: All data on "+w.selectedDevice+" will be permanently erased!\n\n"+
			"Are you sure you want to continue?",
		func(confirmed bool) {
			if confirmed {
				w.startWriteOperation()
			}
		},
		w.window,
	)
}

// startWriteOperation begins the USB creation process
func (w *MainWindow) startWriteOperation() {
	w.SetState(StateInProgress)
	w.progressBar.Reset()

	go func() {
		var err error

		// Check if we're running as root
		if !IsRoot() {
			// Re-launch the CLI with pkexec for the actual write
			err = w.executeWithPkexec()
		} else {
			err = w.executeDeviceMode()
		}

		// Update UI on completion (schedule on main thread)
		time.Sleep(100 * time.Millisecond) // Small delay to ensure UI updates

		if err != nil {
			w.SetState(StateError)
			w.updateStatus(fmt.Sprintf("Error: %v", err))
			w.showError(err.Error())
		} else {
			w.SetState(StateComplete)
			w.updateProgress(1.0, "Complete!")
			w.showSuccess()
		}
	}()
}

// executeWithPkexec runs the CLI tool with elevated privileges via pkexec
func (w *MainWindow) executeWithPkexec() error {
	w.updateProgress(0.02, "Requesting administrator privileges...")

	// Get the path to our own executable
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Build the command: pkexec /path/to/woeusb-go --device <iso> <device>
	cmd := exec.Command("pkexec", executable, "--device", w.selectedISO, w.selectedDevice)

	// Create pipes for stdout/stderr to capture progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pkexec: %v", err)
	}

	// Read output in goroutines to update progress
	go w.readOutput(stdout)
	go w.readOutput(stderr)

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("write operation failed: %v", err)
	}

	return nil
}

// readOutput reads from a pipe and updates progress based on output
func (w *MainWindow) readOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// Parse progress from CLI output and update GUI
		w.parseProgressLine(line)
	}
}

// parseProgressLine extracts progress info from CLI output
func (w *MainWindow) parseProgressLine(line string) {
	// Map CLI output to progress updates
	switch {
	case strings.Contains(line, "Mounting source"):
		w.updateProgress(0.05, "Mounting ISO file...")
	case strings.Contains(line, "Wiping device"):
		w.updateProgress(0.10, "Creating partition table...")
	case strings.Contains(line, "Formatting partition"):
		w.updateProgress(0.15, "Formatting partition...")
	case strings.Contains(line, "Mounting target"):
		w.updateProgress(0.20, "Mounting target partition...")
	case strings.Contains(line, "Copying"):
		w.updateProgress(0.50, "Copying Windows files...")
	case strings.Contains(line, "Installing GRUB"):
		w.updateProgress(0.90, "Installing bootloader...")
	case strings.Contains(line, "Cleaning up"):
		w.updateProgress(0.95, "Cleaning up...")
	case strings.Contains(line, "completed successfully"):
		w.updateProgress(1.0, "Complete!")
	}
}

// updateProgress safely updates progress from any goroutine
func (w *MainWindow) updateProgress(value float64, status string) {
	w.progressBar.SetProgressAndStatus(value, status)
}

// updateStatus safely updates status label from any goroutine
func (w *MainWindow) updateStatus(status string) {
	w.statusLabel.SetText(status)
}

// showError displays an error dialog
func (w *MainWindow) showError(message string) {
	dialog.ShowError(fmt.Errorf("%s", message), w.window)
}

// showSuccess displays a success dialog
func (w *MainWindow) showSuccess() {
	dialog.ShowInformation("Success",
		"Bootable USB created successfully!\n\nYou may now safely remove the USB device.",
		w.window)
}

// executeDeviceMode performs the actual USB creation
func (w *MainWindow) executeDeviceMode() error {
	var srcMount, dstMount string
	var err error

	// Cleanup function
	defer func() {
		if dstMount != "" {
			_ = mount.CleanupMountpoint(dstMount)
		}
		if srcMount != "" {
			_ = mount.CleanupMountpoint(srcMount)
		}
	}()

	// Step 1: Mount source ISO
	w.updateProgress(0.05, "Mounting ISO file...")
	srcMount, err = mount.MountISO(w.selectedISO)
	if err != nil {
		return fmt.Errorf("failed to mount ISO: %v", err)
	}

	// Step 2: Create partition table
	w.updateProgress(0.10, "Creating partition table...")
	if err := partition.CreateBootablePartition(w.selectedDevice, "FAT"); err != nil {
		return fmt.Errorf("failed to create partition: %v", err)
	}

	// Step 3: Get partition path and format
	mainPartition := partition.GetPartitionPath(w.selectedDevice)
	w.updateProgress(0.15, "Formatting partition as FAT32...")
	if err := filesystem.FormatPartition(mainPartition, "FAT", "YOURWINDOWS"); err != nil {
		return fmt.Errorf("failed to format partition: %v", err)
	}

	// Step 4: Mount target partition
	w.updateProgress(0.20, "Mounting target partition...")
	dstMount, err = mount.MountDevice(mainPartition, "vfat")
	if err != nil {
		return fmt.Errorf("failed to mount target: %v", err)
	}

	// Step 5: Copy files with progress callback
	w.updateProgress(0.25, "Copying Windows files (this may take a while)...")

	progressCallback := func(current, total int64, filename string) {
		if total > 0 {
			// Scale progress from 0.25 to 0.90 during copy
			copyProgress := float64(current) / float64(total)
			overallProgress := 0.25 + (copyProgress * 0.65)
			status := fmt.Sprintf("Copying: %s (%.1f%%)", filename, copyProgress*100)
			w.updateProgress(overallProgress, status)
		}
	}

	if err := filecopy.CopyWindowsISOWithWIMSplit(srcMount, dstMount, progressCallback); err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}

	// Step 6: Install GRUB bootloader
	w.updateProgress(0.92, "Installing GRUB bootloader...")
	dependencies, _ := deps.CheckDependencies()
	if dependencies != nil && dependencies.GrubCmd != "" {
		if err := bootloader.InstallGRUBWithConfig(dstMount, w.selectedDevice, dependencies.GrubCmd); err != nil {
			// GRUB failure is non-fatal, UEFI boot will still work
			w.updateProgress(0.95, "GRUB install failed (UEFI boot will work)")
		}
	}

	// Step 7: Cleanup
	w.updateProgress(0.98, "Cleaning up...")
	_ = mount.CleanupMountpoint(dstMount) // Non-fatal, ignore error
	dstMount = ""

	_ = mount.CleanupMountpoint(srcMount) // Non-fatal, ignore error
	srcMount = ""

	return nil
}

// onCloseRequested handles window close requests
func (w *MainWindow) onCloseRequested() {
	if w.state == StateInProgress {
		dialog.ShowConfirm(
			"Operation in Progress",
			"A write operation is currently in progress.\n\n"+
				"Closing now may leave the USB device in an unusable state.\n\n"+
				"Are you sure you want to close?",
			func(confirmed bool) {
				if confirmed {
					// TODO: Cancel operation and cleanup
					w.window.Close()
				}
			},
			w.window,
		)
	} else {
		w.window.Close()
	}
}

// CanStart returns true if the start button should be enabled
// This is exposed for testing Property 7
func CanStart(deviceSelected, isoSelected bool, state OperationState) bool {
	return deviceSelected && isoSelected && state == StateIdle
}

// ShouldDisableControls returns true if UI controls should be disabled
// This is exposed for testing Property 11
func ShouldDisableControls(state OperationState) bool {
	return state == StateInProgress
}
