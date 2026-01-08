package components

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ProgressState holds the progress bar state (testable without Fyne)
type ProgressState struct {
	percentage float64
	status     string
	mu         sync.RWMutex
}

// NewProgressState creates a new progress state
func NewProgressState() *ProgressState {
	return &ProgressState{
		percentage: 0.0,
		status:     "Ready",
	}
}

// SetProgress updates the progress percentage (0.0 to 1.0)
func (ps *ProgressState) SetProgress(value float64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Clamp value between 0 and 1
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	ps.percentage = value
}

// SetStatus updates the status text
func (ps *ProgressState) SetStatus(status string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.status = status
}

// SetProgressAndStatus updates both progress and status atomically
func (ps *ProgressState) SetProgressAndStatus(value float64, status string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Clamp value between 0 and 1
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	ps.percentage = value
	ps.status = status
}

// Reset resets the progress state to initial values
func (ps *ProgressState) Reset() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.percentage = 0.0
	ps.status = "Ready"
}

// GetProgress returns the current progress value
func (ps *ProgressState) GetProgress() float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.percentage
}

// GetStatus returns the current status text
func (ps *ProgressState) GetStatus() string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.status
}

// ProgressBar displays operation progress as a Fyne widget
type ProgressBar struct {
	widget.BaseWidget
	state       *ProgressState
	bar         *widget.ProgressBar
	statusLabel *widget.Label
	container   *fyne.Container
}

// NewProgressBar creates a new progress bar widget
func NewProgressBar() *ProgressBar {
	pb := &ProgressBar{
		state: NewProgressState(),
	}

	pb.bar = widget.NewProgressBar()
	pb.bar.Min = 0
	pb.bar.Max = 1

	pb.statusLabel = widget.NewLabel("Ready")
	pb.statusLabel.Alignment = fyne.TextAlignCenter

	pb.container = container.NewVBox(
		pb.bar,
		pb.statusLabel,
	)

	pb.ExtendBaseWidget(pb)
	return pb
}

// CreateRenderer implements fyne.Widget
func (pb *ProgressBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(pb.container)
}

// SetProgress updates the progress percentage (0.0 to 1.0)
func (pb *ProgressBar) SetProgress(value float64) {
	pb.state.SetProgress(value)
	// Update UI on main thread
	fyne.Do(func() {
		pb.bar.SetValue(pb.state.GetProgress())
	})
}

// SetStatus updates the status text
func (pb *ProgressBar) SetStatus(status string) {
	pb.state.SetStatus(status)
	// Update UI on main thread
	fyne.Do(func() {
		pb.statusLabel.SetText(status)
	})
}

// SetProgressAndStatus updates both progress and status atomically
func (pb *ProgressBar) SetProgressAndStatus(value float64, status string) {
	pb.state.SetProgressAndStatus(value, status)
	// Update UI on main thread
	fyne.Do(func() {
		pb.bar.SetValue(pb.state.GetProgress())
		pb.statusLabel.SetText(status)
	})
}

// Reset resets the progress bar to initial state
func (pb *ProgressBar) Reset() {
	pb.state.Reset()
	// Update UI on main thread
	fyne.Do(func() {
		pb.bar.SetValue(0)
		pb.statusLabel.SetText("Ready")
	})
}

// GetProgress returns the current progress value
func (pb *ProgressBar) GetProgress() float64 {
	return pb.state.GetProgress()
}

// GetStatus returns the current status text
func (pb *ProgressBar) GetStatus() string {
	return pb.state.GetStatus()
}

// FormatProgress returns a formatted progress string (e.g., "45%")
func FormatProgress(value float64) string {
	return fmt.Sprintf("%.0f%%", value*100)
}
