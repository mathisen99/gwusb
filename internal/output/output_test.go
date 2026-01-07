package output

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSetNoColor(t *testing.T) {
	// Test enabling no-color
	SetNoColor(true)
	if !noColor {
		t.Error("Expected noColor to be true")
	}

	// Test disabling no-color
	SetNoColor(false)
	if noColor {
		t.Error("Expected noColor to be false")
	}
}

func TestSetVerbose(t *testing.T) {
	// Test enabling verbose
	SetVerbose(true)
	if !verboseMode {
		t.Error("Expected verboseMode to be true")
	}

	// Test disabling verbose
	SetVerbose(false)
	if verboseMode {
		t.Error("Expected verboseMode to be false")
	}
}

func TestColorize(t *testing.T) {
	// Test with colors enabled
	SetNoColor(false)
	result := colorize(Green, "test")
	if !strings.Contains(result, Green) {
		t.Error("Expected color code in output when colors enabled")
	}
	if !strings.Contains(result, Reset) {
		t.Error("Expected reset code in output when colors enabled")
	}

	// Test with colors disabled
	SetNoColor(true)
	result = colorize(Green, "test")
	if strings.Contains(result, Green) {
		t.Error("Expected no color code when colors disabled")
	}
	if result != "test" {
		t.Errorf("Expected plain 'test', got '%s'", result)
	}

	// Reset for other tests
	SetNoColor(false)
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestStep(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Step("Test step %d", 1)
	})
	if !strings.Contains(output, "▶ Test step 1") {
		t.Errorf("Expected step output, got: %s", output)
	}
	SetNoColor(false)
}

func TestInfo(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Info("Test info")
	})
	if !strings.Contains(output, "✓ Test info") {
		t.Errorf("Expected info output, got: %s", output)
	}
	SetNoColor(false)
}

func TestWarning(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Warning("Test warning")
	})
	if !strings.Contains(output, "⚠ Test warning") {
		t.Errorf("Expected warning output, got: %s", output)
	}
	SetNoColor(false)
}

func TestError(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Error("Test error")
	})
	if !strings.Contains(output, "✗ Test error") {
		t.Errorf("Expected error output, got: %s", output)
	}
	SetNoColor(false)
}

func TestNotice(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Notice("Test notice")
	})
	if !strings.Contains(output, "ℹ Test notice") {
		t.Errorf("Expected notice output, got: %s", output)
	}
	SetNoColor(false)
}

func TestSuccess(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Success("Test success")
	})
	if !strings.Contains(output, "✓ Test success") {
		t.Errorf("Expected success output, got: %s", output)
	}
	SetNoColor(false)
}

func TestProgress(t *testing.T) {
	SetNoColor(true)
	output := captureStderr(func() {
		Progress("Copying %d%%", 50)
	})
	if !strings.Contains(output, "Copying 50%") {
		t.Errorf("Expected progress output, got: %s", output)
	}
	SetNoColor(false)
}

func TestProgressWithColor(t *testing.T) {
	SetNoColor(false)
	output := captureStderr(func() {
		Progress("Copying %d%%", 75)
	})
	if !strings.Contains(output, "Copying 75%") {
		t.Errorf("Expected progress output, got: %s", output)
	}
	if !strings.Contains(output, Blue) {
		t.Error("Expected blue color in progress output")
	}
}

func TestProgressDone(t *testing.T) {
	output := captureStderr(func() {
		ProgressDone()
	})
	if output != "\n" {
		t.Errorf("Expected newline, got: %q", output)
	}
}

func TestVerbose(t *testing.T) {
	SetNoColor(true)

	// Test with verbose disabled
	SetVerbose(false)
	output := captureStderr(func() {
		Verbose("Should not appear")
	})
	if output != "" {
		t.Errorf("Expected no output when verbose disabled, got: %s", output)
	}

	// Test with verbose enabled
	SetVerbose(true)
	output = captureStderr(func() {
		Verbose("Should appear")
	})
	if !strings.Contains(output, "[verbose] Should appear") {
		t.Errorf("Expected verbose output, got: %s", output)
	}

	SetVerbose(false)
	SetNoColor(false)
}

func TestColorConstants(t *testing.T) {
	// Verify color constants are defined
	if Reset == "" {
		t.Error("Reset constant should not be empty")
	}
	if Red == "" {
		t.Error("Red constant should not be empty")
	}
	if Green == "" {
		t.Error("Green constant should not be empty")
	}
	if Yellow == "" {
		t.Error("Yellow constant should not be empty")
	}
	if Blue == "" {
		t.Error("Blue constant should not be empty")
	}
	if Magenta == "" {
		t.Error("Magenta constant should not be empty")
	}
	if Cyan == "" {
		t.Error("Cyan constant should not be empty")
	}
	if Bold == "" {
		t.Error("Bold constant should not be empty")
	}
}
