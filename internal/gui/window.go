package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/mathisen/woeusb-go/internal/distro"
	"github.com/mathisen/woeusb-go/internal/gui/components"
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
	w.statusLabel.SetText("Starting...")

	// TODO: Connect to actual write logic from internal/session
	// For now, this is a placeholder
	go func() {
		// Simulate progress updates
		w.progressBar.SetProgressAndStatus(0.1, "Preparing device...")
		// The actual implementation will call the session package
	}()
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
