package output

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
)

var noColor = false

// SetNoColor disables color output
func SetNoColor(disabled bool) {
	noColor = disabled
}

func colorize(color, text string) string {
	if noColor {
		return text
	}
	return color + text + Reset
}

// Step prints a step header in cyan
func Step(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Cyan+Bold, "▶ "+msg))
}

// Info prints an info message in green
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Green, "  ✓ "+msg))
}

// Warning prints a warning message in yellow
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Yellow, "  ⚠ "+msg))
}

// Error prints an error message in red
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Red, "  ✗ "+msg))
}

// Notice prints a notice in magenta (for long operations)
func Notice(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Magenta, "  ℹ "+msg))
}

// Success prints a success message in bold green
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorize(Green+Bold, "✓ "+msg))
}

// Progress prints progress info (overwrites line)
func Progress(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Fprintf(os.Stderr, "\r  %s", msg)
	} else {
		fmt.Fprintf(os.Stderr, "\r  %s%s%s", Blue, msg, Reset)
	}
}

// ProgressDone finishes progress line
func ProgressDone() {
	fmt.Fprintln(os.Stderr)
}

// Verbose prints only if verbose mode is enabled
var verboseMode = false

func SetVerbose(enabled bool) {
	verboseMode = enabled
}

func Verbose(format string, args ...interface{}) {
	if verboseMode {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintln(os.Stderr, colorize(Cyan, "  [verbose] "+msg))
	}
}
