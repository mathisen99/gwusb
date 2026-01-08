package components

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/distro"
)

// DependencyDialog shows missing dependencies with install instructions
type DependencyDialog struct {
	parent      fyne.Window
	app         fyne.App
	missingDeps []deps.MissingDep
	distroInfo  *distro.Info
}

// NewDependencyDialog creates a dependency dialog
func NewDependencyDialog(parent fyne.Window, missing []deps.MissingDep, info *distro.Info) *DependencyDialog {
	return &DependencyDialog{
		parent:      parent,
		app:         fyne.CurrentApp(),
		missingDeps: missing,
		distroInfo:  info,
	}
}

// Show displays the dialog
func (d *DependencyDialog) Show() {
	content := d.buildContent()

	customDialog := dialog.NewCustom(
		"Missing Dependencies",
		"Close",
		content,
		d.parent,
	)
	customDialog.Resize(fyne.NewSize(600, 400))
	customDialog.Show()
}

// buildContent creates the dialog content
func (d *DependencyDialog) buildContent() fyne.CanvasObject {
	// Header
	header := widget.NewLabel("The following dependencies are required but not installed:")
	header.TextStyle = fyne.TextStyle{Bold: true}

	// Missing dependencies list
	var depItems []fyne.CanvasObject
	for _, dep := range d.missingDeps {
		label := d.formatDependency(dep)
		depItems = append(depItems, widget.NewLabel(label))
	}
	depList := container.NewVBox(depItems...)

	// Install command
	installCmd := d.GetInstallCommand()
	cmdLabel := widget.NewLabel("Install command:")
	cmdLabel.TextStyle = fyne.TextStyle{Bold: true}

	cmdEntry := widget.NewEntry()
	cmdEntry.SetText(installCmd)
	cmdEntry.Disable() // Read-only

	// Copy button
	copyBtn := widget.NewButton("Copy Command", func() {
		if d.app != nil {
			d.app.Clipboard().SetContent(installCmd)
		}
	})

	// Distro info
	distroLabel := widget.NewLabel(d.getDistroDescription())
	distroLabel.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		depList,
		widget.NewSeparator(),
		cmdLabel,
		cmdEntry,
		copyBtn,
		widget.NewSeparator(),
		distroLabel,
	)
}

// formatDependency formats a single dependency for display
func (d *DependencyDialog) formatDependency(dep deps.MissingDep) string {
	reqStr := "[optional]"
	if dep.Required {
		reqStr = "[REQUIRED]"
	}
	return fmt.Sprintf("• %s (package: %s) %s", dep.Binary, dep.PackageName, reqStr)
}

// GetInstallCommand returns the full install command for the distro
func (d *DependencyDialog) GetInstallCommand() string {
	return deps.GetInstallCommand(d.missingDeps, d.distroInfo)
}

// getDistroDescription returns a description of the detected distro
func (d *DependencyDialog) getDistroDescription() string {
	if d.distroInfo == nil {
		return "Distribution: Unknown (using generic package names)"
	}

	name := d.distroInfo.Name
	if name == "" {
		name = d.distroInfo.ID
	}
	if name == "" {
		return "Distribution: Unknown (using generic package names)"
	}

	pm := d.distroInfo.PackageManager
	if pm == "" {
		pm = "unknown"
	}

	return fmt.Sprintf("Detected: %s (package manager: %s)", name, pm)
}

// FormatMissingDeps formats a list of missing dependencies for display
func FormatMissingDeps(missing []deps.MissingDep) string {
	if len(missing) == 0 {
		return "All dependencies are installed."
	}

	var lines []string
	for _, dep := range missing {
		reqStr := "optional"
		if dep.Required {
			reqStr = "REQUIRED"
		}
		lines = append(lines, fmt.Sprintf("• %s (package: %s) [%s]", dep.Binary, dep.PackageName, reqStr))
	}
	return strings.Join(lines, "\n")
}

// HasRequiredMissing checks if any required dependencies are missing
func HasRequiredMissing(missing []deps.MissingDep) bool {
	for _, dep := range missing {
		if dep.Required {
			return true
		}
	}
	return false
}
