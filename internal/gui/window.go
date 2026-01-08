package gui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	// Check if we're running as root
	if IsRoot() {
		// Already root, proceed directly
		w.SetState(StateInProgress)
		w.progressBar.Reset()
		go w.runWriteOperation("")
	} else {
		// Need to elevate - show password dialog
		components.ShowPasswordDialogWithInfo(
			w.window,
			"WoeUSB-go needs administrator privileges to write to the USB device.",
			func(result components.PasswordResult) {
				if result.Cancelled {
					// User cancelled, don't start operation
					return
				}
				// Validate password first by running a simple sudo command
				w.SetState(StateInProgress)
				w.progressBar.Reset()
				w.updateProgress(0.01, "Validating credentials...")

				go func() {
					// Test sudo credentials
					if err := w.validateSudoPassword(result.Password); err != nil {
						w.SetState(StateError)
						w.updateStatus("Authentication failed")
						w.showError("Incorrect password. Please try again.")
						return
					}
					// Run the write operation with sudo
					w.runWriteOperation(result.Password)
				}()
			},
		)
	}
}

// validateSudoPassword tests if the password is correct
func (w *MainWindow) validateSudoPassword(password string) error {
	cmd := exec.Command("sudo", "-S", "-v")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	_, _ = stdin.Write([]byte(password + "\n"))
	_ = stdin.Close()

	return cmd.Wait()
}

// runWriteOperation executes the write operation (with or without sudo)
func (w *MainWindow) runWriteOperation(password string) {
	var err error

	if password != "" {
		// Cache sudo credentials for subsequent commands
		w.updateProgress(0.02, "Authenticating...")
		// Run with sudo using the provided password
		err = w.executeWithSudo(password)
	} else {
		// Already root, run directly
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
}

// executeWithSudo runs the CLI tool with elevated privileges via sudo -S
func (w *MainWindow) executeWithSudo(password string) error {
	w.updateProgress(0.02, "Authenticating...")

	// Get the path to our own executable
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Build the command: sudo -S /path/to/woeusb-go --device <iso> <device>
	// Use -n after authentication to prevent further password prompts
	cmd := exec.Command("sudo", "-S", executable, "--device", w.selectedISO, w.selectedDevice)

	// Create pipe for stdin to send password
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

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
		return fmt.Errorf("failed to start sudo: %v", err)
	}

	// Send password to sudo via stdin, then close
	_, err = stdin.Write([]byte(password + "\n"))
	if err != nil {
		return fmt.Errorf("failed to send password: %v", err)
	}
	_ = stdin.Close()

	// Read output in goroutines to update progress
	// Use a WaitGroup to ensure we read all output before Wait() returns
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		w.readOutputWithCR(stdout)
	}()
	go func() {
		defer wg.Done()
		w.readOutputWithCR(stderr)
	}()

	// Wait for output readers to finish
	wg.Wait()

	// Wait for command completion
	if err := cmd.Wait(); err != nil {
		// Check if it's an authentication failure
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return fmt.Errorf("authentication failed - incorrect password")
			}
		}
		return fmt.Errorf("write operation failed: %v", err)
	}

	return nil
}

// readOutputWithCR reads from a pipe handling both \n and \r as line separators
func (w *MainWindow) readOutputWithCR(r io.Reader) {
	buf := make([]byte, 4096)
	var line strings.Builder

	for {
		n, err := r.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				ch := buf[i]
				if ch == '\n' || ch == '\r' {
					if line.Len() > 0 {
						w.parseProgressLine(line.String())
						line.Reset()
					}
				} else {
					line.WriteByte(ch)
				}
			}
		}
		if err != nil {
			// Process any remaining content
			if line.Len() > 0 {
				w.parseProgressLine(line.String())
			}
			break
		}
	}
}

// parseProgressLine extracts progress info from CLI output
func (w *MainWindow) parseProgressLine(line string) {
	// Skip empty lines
	if strings.TrimSpace(line) == "" {
		return
	}

	// Try to parse percentage from "Copying: XX.X%" format
	if strings.Contains(line, "Copying:") && strings.Contains(line, "%") {
		// Extract percentage from line like "Copying: 45.2% (1.2 GB) - sources/install.wim"
		var pct float64
		if _, err := fmt.Sscanf(line, "Copying: %f%%", &pct); err == nil {
			// Scale copy progress from 0.25 to 0.85
			progress := 0.25 + (pct/100.0)*0.60
			w.updateProgress(progress, line)
			return
		}
	}

	// Try to parse wimlib-imagex split progress
	if strings.Contains(line, "Writing") && strings.Contains(line, "MiB") {
		w.updateProgress(0.85, "Splitting WIM file: "+line)
		return
	}

	// Map CLI output to progress updates
	switch {
	case strings.Contains(line, "Mounting source") || strings.Contains(line, "Mounting ISO"):
		w.updateProgress(0.05, "Mounting ISO file...")
	case strings.Contains(line, "Wiping") || strings.Contains(line, "partition table"):
		w.updateProgress(0.10, "Creating partition table...")
	case strings.Contains(line, "Formatting"):
		w.updateProgress(0.15, "Formatting partition...")
	case strings.Contains(line, "Mounting target"):
		w.updateProgress(0.20, "Mounting target partition...")
	case strings.Contains(line, "Will split"):
		w.updateProgress(0.22, line)
	case strings.Contains(line, "Copying files"):
		w.updateProgress(0.25, "Copying files...")
	case strings.Contains(line, "Splitting"):
		w.updateProgress(0.85, line)
	case strings.Contains(line, "Split") && strings.Contains(line, "SWM"):
		w.updateProgress(0.88, line)
	case strings.Contains(line, "Installing GRUB") || strings.Contains(line, "GRUB"):
		w.updateProgress(0.90, "Installing bootloader...")
	case strings.Contains(line, "Cleaning up"):
		w.updateProgress(0.95, "Cleaning up...")
	case strings.Contains(line, "completed successfully"):
		w.updateProgress(1.0, "Complete!")
	default:
		// Show any other meaningful output
		if len(line) > 5 && !strings.HasPrefix(line, "[sudo]") {
			w.updateProgress(-1, line) // -1 means don't update progress bar, just status
		}
	}
}

// updateProgress safely updates progress from any goroutine
// If value is -1, only updates status text without changing progress bar
func (w *MainWindow) updateProgress(value float64, status string) {
	if value >= 0 {
		w.progressBar.SetProgressAndStatus(value, status)
	} else {
		w.progressBar.SetStatus(status)
	}
}

// updateStatus safely updates status label from any goroutine
func (w *MainWindow) updateStatus(status string) {
	fyne.Do(func() {
		w.statusLabel.SetText(status)
	})
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
