package gui

import (
	"testing"
)

// TestProperty7_StartButtonState tests Property 7:
// For any combination of device selection (empty or set) and ISO selection
// (empty or set), the Start button SHALL be enabled if and only if both
// selections are non-empty.
func TestProperty7_StartButtonState(t *testing.T) {
	testCases := []struct {
		name           string
		deviceSelected bool
		isoSelected    bool
		state          OperationState
		expected       bool
	}{
		{"No device, no ISO, idle", false, false, StateIdle, false},
		{"Device only, idle", true, false, StateIdle, false},
		{"ISO only, idle", false, true, StateIdle, false},
		{"Both selected, idle", true, true, StateIdle, true},
		{"Both selected, in progress", true, true, StateInProgress, false},
		{"Both selected, complete", true, true, StateComplete, false},
		{"Both selected, error", true, true, StateError, false},
		{"No device, no ISO, in progress", false, false, StateInProgress, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CanStart(tc.deviceSelected, tc.isoSelected, tc.state)
			if result != tc.expected {
				t.Errorf("CanStart(%v, %v, %v) = %v, want %v",
					tc.deviceSelected, tc.isoSelected, tc.state, result, tc.expected)
			}
		})
	}
}

// TestProperty11_UIControlsDisabledDuringOperation tests Property 11:
// For any operation state (in_progress or idle), the Start button and
// selection controls SHALL be disabled if and only if the state is in_progress.
func TestProperty11_UIControlsDisabledDuringOperation(t *testing.T) {
	testCases := []struct {
		name     string
		state    OperationState
		expected bool
	}{
		{"Idle state - controls enabled", StateIdle, false},
		{"In progress - controls disabled", StateInProgress, true},
		{"Complete state - controls enabled", StateComplete, false},
		{"Error state - controls enabled", StateError, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldDisableControls(tc.state)
			if result != tc.expected {
				t.Errorf("ShouldDisableControls(%v) = %v, want %v",
					tc.state, result, tc.expected)
			}
		})
	}
}

// TestOperationState_Values tests that operation states have expected values
func TestOperationState_Values(t *testing.T) {
	// Verify states are distinct
	states := []OperationState{StateIdle, StateInProgress, StateComplete, StateError}
	seen := make(map[OperationState]bool)

	for _, state := range states {
		if seen[state] {
			t.Errorf("Duplicate state value: %v", state)
		}
		seen[state] = true
	}
}

// TestCanStart_AllCombinations tests all possible combinations
func TestCanStart_AllCombinations(t *testing.T) {
	states := []OperationState{StateIdle, StateInProgress, StateComplete, StateError}
	bools := []bool{true, false}

	for _, device := range bools {
		for _, iso := range bools {
			for _, state := range states {
				result := CanStart(device, iso, state)

				// Only true when both selected AND state is idle
				expected := device && iso && state == StateIdle

				if result != expected {
					t.Errorf("CanStart(%v, %v, %v) = %v, want %v",
						device, iso, state, result, expected)
				}
			}
		}
	}
}
