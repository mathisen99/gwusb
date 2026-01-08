// Package gui provides the graphical user interface for WoeUSB-go
// using the Fyne toolkit for cross-platform rendering.
package gui

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"github.com/mathisen/woeusb-go/internal/deps"
	"github.com/mathisen/woeusb-go/internal/distro"
)

// darkTheme implements a custom dark theme for WoeUSB-go
type darkTheme struct{}

func (d *darkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (d *darkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (d *darkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (d *darkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// App represents the main GUI application
type App struct {
	fyneApp    fyne.App
	mainWindow *MainWindow
	distroInfo *distro.Info
}

// NewApp creates a new GUI application instance
func NewApp() *App {
	a := app.NewWithID("io.github.woeusb-go")
	a.Settings().SetTheme(&darkTheme{})
	return &App{
		fyneApp: a,
	}
}

// Run starts the GUI application
func (a *App) Run() error {
	// Detect distro for dependency checking
	a.distroInfo, _ = distro.Detect() // Ignore error, will use fallback

	// Check dependencies (but don't block on root - we'll use pkexec)
	missing := a.CheckDependencies()
	if len(missing) > 0 {
		// Only show dialog for required dependencies
		hasRequired := false
		for _, dep := range missing {
			if dep.Required {
				hasRequired = true
				break
			}
		}
		if hasRequired {
			a.showDependencyDialog(missing)
			return nil // User needs to install dependencies first
		}
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

// showDependencyDialog displays missing dependencies with install instructions
func (a *App) showDependencyDialog(missing []deps.MissingDep) {
	win := a.fyneApp.NewWindow("WoeUSB-go - Missing Dependencies")
	win.Resize(fyne.NewSize(600, 400))

	// Build the message using strings.Builder for efficiency
	var sb strings.Builder
	sb.WriteString("The following dependencies are missing:\n\n")
	for _, dep := range missing {
		if dep.Required {
			sb.WriteString(fmt.Sprintf("• %s (package: %s) [REQUIRED]\n", dep.Binary, dep.PackageName))
		} else {
			sb.WriteString(fmt.Sprintf("• %s (package: %s) [optional]\n", dep.Binary, dep.PackageName))
		}
	}

	// Get install command
	installCmd := deps.GetInstallCommand(missing, a.distroInfo)
	if installCmd != "" {
		sb.WriteString(fmt.Sprintf("\nInstall command:\n%s", installCmd))
	}

	dialog.ShowInformation("Missing Dependencies", sb.String(), win)
	win.Show()
}
