package deps

import (
	"testing"
)

func TestCheckDependencies(t *testing.T) {
	deps, err := CheckDependencies()

	// On most systems, some dependencies might be missing
	// This test just ensures the function doesn't panic
	if err != nil {
		t.Logf("Missing dependencies (expected on some systems): %v", err)
		return
	}

	// If no error, verify we got valid paths
	if deps.Wipefs == "" {
		t.Error("Expected wipefs path to be set")
	}
	if deps.Parted == "" {
		t.Error("Expected parted path to be set")
	}
	if deps.MkFat == "" {
		t.Error("Expected mkfat path to be set")
	}
}
