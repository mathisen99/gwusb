// Package gui provides the graphical user interface for WoeUSB-go
// using the Fyne toolkit for cross-platform rendering.
package gui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"

	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/distro"
)

// App represents the main GUI application
type App struct {
	fyneApp    fyne.App
	mainWindow *MainWindow
	distroInfo *distro.Info
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	return &App{
		fyneApp: app.New(),
	}
}

// Run starts the GUI application
func (a *App) Run() error {
	// Detect distro for dependency checking
	a.distroInfo, _ = distro.Detect() // Ignore error, will use fallback

	// Check root privileges
	if !IsRoot() {
		a.showRootWarning()
	}

	// Check dependencies
	missing := a.CheckDependencies()
	if len(missing) > 0 {
		a.showDependencyDialog(missing)
		return nil // User needs to install dependencies first
	}

	// Create and show main window
	a.mainWindow = NewMainWindow(a.fyneApp, a.distroInfo)
	a.mainWindow.Show()

	// Run the application
	a.fyneApp.Run()
	return nil
}

// CheckDependencies verifies all required tools are installed
// Returns a list of missing dependencies with distro-specific package names
func (a *App) CheckDependencies() []deps.MissingDep {
	result := deps.CheckDependenciesWithDistro()
	a.distroInfo = result.DistroInfo
	return result.Missing
}

// IsRoot checks if the application is running with root privileges
func IsRoot() bool {
	return os.Getuid() == 0
}

// IsRootWithGetter checks root using a custom UID getter (for testing)
func IsRootWithGetter(getUID func() int) bool {
	return getUID() == 0
}

// showRootWarning displays a warning dialog about missing root privileges
func (a *App) showRootWarning() {
	win := a.fyneApp.NewWindow("WoeUSB-go")
	win.Resize(fyne.NewSize(400, 200))

	dialog.ShowInformation(
		"Root Privileges Required",
		"WoeUSB-go requires root privileges to write to USB devices.\n\n"+
			"Please restart the application with sudo:\n"+
			"  sudo woeusb-go --gui",
		win,
	)
	win.Show()
}

// showDependencyDialog displays missing dependencies with install instructions
func (a *App) showDependencyDialog(missing []deps.MissingDep) {
	win := a.fyneApp.NewWindow("WoeUSB-go - Missing Dependencies")
	win.Resize(fyne.NewSize(600, 400))

	// Build the message
	msg := "The following dependencies are missing:\n\n"
	for _, dep := range missing {
		if dep.Required {
			msg += fmt.Sprintf("• %s (package: %s) [REQUIRED]\n", dep.Binary, dep.PackageName)
		} else {
			msg += fmt.Sprintf("• %s (package: %s) [optional]\n", dep.Binary, dep.PackageName)
		}
	}

	// Get install command
	installCmd := deps.GetInstallCommand(missing, a.distroInfo)
	if installCmd != "" {
		msg += fmt.Sprintf("\nInstall command:\n%s", installCmd)
	}

	dialog.ShowInformation("Missing Dependencies", msg, win)
	win.Show()
}
