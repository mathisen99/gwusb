package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// FileBrowser provides ISO file selection
type FileBrowser struct {
	widget.BaseWidget
	selectedPath string
	pathLabel    *widget.Label
	browseBtn    *widget.Button
	onSelect     func(path string)
	container    *fyne.Container
}

// NewFileBrowser creates a new file browser widget
func NewFileBrowser(onSelect func(path string)) *FileBrowser {
	fb := &FileBrowser{
		onSelect: onSelect,
	}

	fb.pathLabel = widget.NewLabel("No ISO file selected")
	fb.pathLabel.Wrapping = fyne.TextWrapWord

	fb.browseBtn = widget.NewButton("Browse...", func() {
		// This will be called when button is clicked
		// The actual dialog opening requires a parent window
	})

	fb.container = container.NewBorder(
		nil, nil, nil,
		fb.browseBtn,
		fb.pathLabel,
	)

	fb.ExtendBaseWidget(fb)
	return fb
}

// CreateRenderer implements fyne.Widget
func (fb *FileBrowser) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fb.container)
}

// OpenDialog opens the file selection dialog
func (fb *FileBrowser) OpenDialog(parent fyne.Window) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer func() { _ = reader.Close() }()

		path := reader.URI().Path()

		// Validate the ISO file
		if err := ValidateISO(path); err != nil {
			dialog.ShowError(err, parent)
			return
		}

		fb.selectedPath = path
		fb.pathLabel.SetText(filepath.Base(path))

		if fb.onSelect != nil {
			fb.onSelect(path)
		}
	}, parent)

	// Filter for ISO files
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".iso", ".ISO"}))

	// Start in home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		listableURI, err := storage.ListerForURI(storage.NewFileURI(homeDir))
		if err == nil {
			fd.SetLocation(listableURI)
		}
	}

	fd.Show()
}

// SetBrowseAction sets the browse button action with a parent window
func (fb *FileBrowser) SetBrowseAction(parent fyne.Window) {
	fb.browseBtn.OnTapped = func() {
		fb.OpenDialog(parent)
	}
}

// GetSelectedPath returns the currently selected ISO path
func (fb *FileBrowser) GetSelectedPath() string {
	return fb.selectedPath
}

// SetSelectedPath sets the selected path (for testing or programmatic use)
func (fb *FileBrowser) SetSelectedPath(path string) error {
	if err := ValidateISO(path); err != nil {
		return err
	}
	fb.selectedPath = path
	fb.pathLabel.SetText(filepath.Base(path))
	if fb.onSelect != nil {
		fb.onSelect(path)
	}
	return nil
}

// ValidateISO checks if the selected file is a valid ISO
// Returns nil if the file exists and is readable, error otherwise
func ValidateISO(path string) error {
	if path == "" {
		return fmt.Errorf("no file path provided")
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a regular file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".iso" {
		return fmt.Errorf("file is not an ISO image (has %s extension): %s", ext, path)
	}

	// Check if file is readable by attempting to open it
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	_ = file.Close()

	return nil
}

// ValidateISOWithStatFunc validates ISO using a custom stat function (for testing)
func ValidateISOWithStatFunc(path string, statFunc func(string) (os.FileInfo, error), openFunc func(string) (*os.File, error)) error {
	if path == "" {
		return fmt.Errorf("no file path provided")
	}

	// Check if file exists
	info, err := statFunc(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a regular file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".iso" {
		return fmt.Errorf("file is not an ISO image (has %s extension): %s", ext, path)
	}

	// Check if file is readable
	if openFunc != nil {
		file, err := openFunc(path)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}
		_ = file.Close()
	}

	return nil
}
