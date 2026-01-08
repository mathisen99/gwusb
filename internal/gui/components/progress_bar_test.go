package components

import (
	"sync"
	"testing"
)

// TestProperty8_ProgressBarUpdates tests Property 8:
// For any progress value between 0.0 and 1.0, SetProgress SHALL update
// the progress bar to display that percentage.
func TestProperty8_ProgressBarUpdates(t *testing.T) {
	ps := NewProgressState()

	testValues := []float64{0.0, 0.1, 0.25, 0.5, 0.75, 0.9, 1.0}

	for _, value := range testValues {
		ps.SetProgress(value)
		got := ps.GetProgress()
		if got != value {
			t.Errorf("SetProgress(%v) = %v, want %v", value, got, value)
		}
	}
}

// TestProgressState_ClampValues tests that values are clamped to [0, 1]
func TestProgressState_ClampValues(t *testing.T) {
	ps := NewProgressState()

	testCases := []struct {
		input    float64
		expected float64
	}{
		{-0.5, 0.0},
		{-1.0, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
		{2.0, 1.0},
		{100.0, 1.0},
	}

	for _, tc := range testCases {
		ps.SetProgress(tc.input)
		got := ps.GetProgress()
		if got != tc.expected {
			t.Errorf("SetProgress(%v) = %v, want %v (clamped)", tc.input, got, tc.expected)
		}
	}
}

// TestProgressState_SetStatus tests status text updates
func TestProgressState_SetStatus(t *testing.T) {
	ps := NewProgressState()

	statuses := []string{
		"Ready",
		"Copying files...",
		"Installing bootloader...",
		"Formatting partition...",
		"Complete!",
		"", // Empty status
	}

	for _, status := range statuses {
		ps.SetStatus(status)
		got := ps.GetStatus()
		if got != status {
			t.Errorf("SetStatus(%q) = %q, want %q", status, got, status)
		}
	}
}

// TestProgressState_Reset tests the reset functionality
func TestProgressState_Reset(t *testing.T) {
	ps := NewProgressState()

	// Set some values
	ps.SetProgress(0.75)
	ps.SetStatus("In progress...")

	// Verify they were set
	if ps.GetProgress() != 0.75 {
		t.Errorf("Progress not set correctly before reset")
	}
	if ps.GetStatus() != "In progress..." {
		t.Errorf("Status not set correctly before reset")
	}

	// Reset
	ps.Reset()

	// Verify reset values
	if ps.GetProgress() != 0.0 {
		t.Errorf("Reset() progress = %v, want 0.0", ps.GetProgress())
	}
	if ps.GetStatus() != "Ready" {
		t.Errorf("Reset() status = %q, want %q", ps.GetStatus(), "Ready")
	}
}

// TestProgressState_SetProgressAndStatus tests atomic update
func TestProgressState_SetProgressAndStatus(t *testing.T) {
	ps := NewProgressState()

	ps.SetProgressAndStatus(0.5, "Halfway there")

	if ps.GetProgress() != 0.5 {
		t.Errorf("SetProgressAndStatus() progress = %v, want 0.5", ps.GetProgress())
	}
	if ps.GetStatus() != "Halfway there" {
		t.Errorf("SetProgressAndStatus() status = %q, want %q", ps.GetStatus(), "Halfway there")
	}
}

// TestProgressState_ConcurrentAccess tests thread safety
func TestProgressState_ConcurrentAccess(t *testing.T) {
	ps := NewProgressState()

	var wg sync.WaitGroup
	iterations := 100

	// Multiple goroutines updating progress
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				ps.SetProgress(float64(j) / float64(iterations))
				ps.SetStatus("Working...")
				_ = ps.GetProgress()
				_ = ps.GetStatus()
			}
		}(i)
	}

	wg.Wait()

	// Just verify no panics occurred and we can still read values
	_ = ps.GetProgress()
	_ = ps.GetStatus()
}

// TestFormatProgress tests progress formatting
func TestFormatProgress(t *testing.T) {
	testCases := []struct {
		value    float64
		expected string
	}{
		{0.0, "0%"},
		{0.1, "10%"},
		{0.25, "25%"},
		{0.5, "50%"},
		{0.75, "75%"},
		{1.0, "100%"},
		{0.333, "33%"},
		{0.666, "67%"},
	}

	for _, tc := range testCases {
		got := FormatProgress(tc.value)
		if got != tc.expected {
			t.Errorf("FormatProgress(%v) = %q, want %q", tc.value, got, tc.expected)
		}
	}
}

// TestProgressState_InitialState tests the initial state after creation
func TestProgressState_InitialState(t *testing.T) {
	ps := NewProgressState()

	if ps.GetProgress() != 0.0 {
		t.Errorf("Initial progress = %v, want 0.0", ps.GetProgress())
	}
	if ps.GetStatus() != "Ready" {
		t.Errorf("Initial status = %q, want %q", ps.GetStatus(), "Ready")
	}
}
