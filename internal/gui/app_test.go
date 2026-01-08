package gui

import (
	"testing"
)

// TestProperty2_RootPrivilegeDetection tests Property 2:
// For any effective user ID, the root check function SHALL return true
// if and only if the UID equals 0.
func TestProperty2_RootPrivilegeDetection(t *testing.T) {
	testCases := []struct {
		name     string
		uid      int
		expected bool
	}{
		{"UID 0 is root", 0, true},
		{"UID 1 is not root", 1, false},
		{"UID 1000 is not root", 1000, false},
		{"UID 65534 (nobody) is not root", 65534, false},
		{"Negative UID is not root", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRootWithGetter(func() int { return tc.uid })
			if result != tc.expected {
				t.Errorf("IsRootWithGetter() for UID %d = %v, want %v", tc.uid, result, tc.expected)
			}
		})
	}
}

// TestIsRoot_PropertyBased runs property-based tests for root detection
func TestIsRoot_PropertyBased(t *testing.T) {
	// Property: Only UID 0 should return true
	for uid := -10; uid <= 100; uid++ {
		result := IsRootWithGetter(func() int { return uid })
		expected := uid == 0
		if result != expected {
			t.Errorf("IsRootWithGetter() for UID %d = %v, want %v", uid, result, expected)
		}
	}

	// Test some large UIDs
	largeUIDs := []int{500, 1000, 65534, 65535, 100000}
	for _, uid := range largeUIDs {
		result := IsRootWithGetter(func() int { return uid })
		if result {
			t.Errorf("IsRootWithGetter() for UID %d should be false", uid)
		}
	}
}
