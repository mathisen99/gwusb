package copy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// ChunkSize for copying large files (1MB)
	ChunkSize = 1024 * 1024
	// LargeFileThreshold files larger than this will be copied in chunks (5MB)
	LargeFileThreshold = 5 * 1024 * 1024
)

// ProgressFunc is called during file copying to report progress
type ProgressFunc func(bytesCopied, totalBytes int64, currentFile string)

// CopyStats holds statistics about the copy operation
type CopyStats struct {
	TotalFiles  int
	TotalBytes  int64
	CopiedFiles int
	CopiedBytes int64
	CurrentFile string
	Failed      []string
}

// CopyWithProgress copies all files from srcMount to dstMount with progress reporting
func CopyWithProgress(srcMount, dstMount string, progressFn ProgressFunc) error {
	// First pass: calculate total size and file count
	stats, err := calculateTotalSize(srcMount)
	if err != nil {
		return fmt.Errorf("failed to calculate total size: %v", err)
	}

	// Second pass: copy files with progress
	return copyFiles(srcMount, dstMount, stats, progressFn)
}

// calculateTotalSize walks the source directory and calculates total bytes and file count
func calculateTotalSize(srcMount string) (*CopyStats, error) {
	stats := &CopyStats{}

	err := filepath.Walk(srcMount, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.Mode().IsRegular() {
			stats.TotalFiles++
			stats.TotalBytes += info.Size()
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// copyFiles performs the actual file copying with progress reporting
func copyFiles(srcMount, dstMount string, stats *CopyStats, progressFn ProgressFunc) error {
	return filepath.Walk(srcMount, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			// Log failed file but continue
			relPath, _ := filepath.Rel(srcMount, srcPath)
			stats.Failed = append(stats.Failed, relPath)
			return nil
		}

		// Calculate destination path
		relPath, err := filepath.Rel(srcMount, srcPath)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstMount, relPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Handle regular files
		if info.Mode().IsRegular() {
			stats.CurrentFile = relPath
			if progressFn != nil {
				progressFn(stats.CopiedBytes, stats.TotalBytes, stats.CurrentFile)
			}

			if err := copyFile(srcPath, dstPath, info.Size(), stats, progressFn); err != nil {
				stats.Failed = append(stats.Failed, relPath)
				return nil // Continue with other files
			}

			stats.CopiedFiles++
		}

		return nil
	})
}

// copyFile copies a single file with progress reporting for large files
func copyFile(srcPath, dstPath string, fileSize int64, stats *CopyStats, progressFn ProgressFunc) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	// For small files, copy directly
	if fileSize < LargeFileThreshold {
		_, err := io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		stats.CopiedBytes += fileSize
		if progressFn != nil {
			progressFn(stats.CopiedBytes, stats.TotalBytes, stats.CurrentFile)
		}
		return nil
	}

	// For large files, copy in chunks with progress updates
	buffer := make([]byte, ChunkSize)
	var totalCopied int64

	for {
		n, err := srcFile.Read(buffer)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return err
		}

		_, writeErr := dstFile.Write(buffer[:n])
		if writeErr != nil {
			return writeErr
		}

		totalCopied += int64(n)
		stats.CopiedBytes += int64(n)

		// Report progress for large files
		if progressFn != nil {
			progressFn(stats.CopiedBytes, stats.TotalBytes, stats.CurrentFile)
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// PrintProgress prints progress information to stderr
func PrintProgress(bytesCopied, totalBytes int64, currentFile string) {
	percentage := float64(bytesCopied) / float64(totalBytes) * 100
	fmt.Fprintf(os.Stderr, "\rCopying: %.1f%% (%s) - %s",
		percentage, formatBytes(bytesCopied), currentFile)
}

// formatBytes formats byte count into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// CopyDirectory is a convenience function that copies a directory with default progress printing
func CopyDirectory(srcDir, dstDir string) error {
	return CopyWithProgress(srcDir, dstDir, PrintProgress)
}

// CopyDirectoryQuiet copies a directory without progress output
func CopyDirectoryQuiet(srcDir, dstDir string) error {
	return CopyWithProgress(srcDir, dstDir, nil)
}

// ValidateCopy verifies that the copy operation was successful
func ValidateCopy(srcMount, dstMount string) error {
	srcStats, err := calculateTotalSize(srcMount)
	if err != nil {
		return fmt.Errorf("failed to calculate source size: %v", err)
	}

	dstStats, err := calculateTotalSize(dstMount)
	if err != nil {
		return fmt.Errorf("failed to calculate destination size: %v", err)
	}

	if srcStats.TotalFiles != dstStats.TotalFiles {
		return fmt.Errorf("file count mismatch: source=%d, destination=%d",
			srcStats.TotalFiles, dstStats.TotalFiles)
	}

	if srcStats.TotalBytes != dstStats.TotalBytes {
		return fmt.Errorf("size mismatch: source=%d bytes, destination=%d bytes",
			srcStats.TotalBytes, dstStats.TotalBytes)
	}

	return nil
}

// FAT32 max file size (4GB - 1 byte)
const FAT32MaxFileSize = 4*1024*1024*1024 - 1

// SplitWIMMaxSize is the max size for split WIM parts (3.8GB to be safe)
const SplitWIMMaxSize = 3800

// LargeFile represents a file that exceeds FAT32 limits
type LargeFile struct {
	RelPath string
	Size    int64
}

// FindLargeFiles finds all files > 4GB in the source directory
func FindLargeFiles(srcMount string) ([]LargeFile, error) {
	var largeFiles []LargeFile

	err := filepath.Walk(srcMount, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Mode().IsRegular() && info.Size() > FAT32MaxFileSize {
			relPath, _ := filepath.Rel(srcMount, path)
			largeFiles = append(largeFiles, LargeFile{
				RelPath: relPath,
				Size:    info.Size(),
			})
		}
		return nil
	})

	return largeFiles, err
}

// IsWIMFile checks if a file is a WIM file
func IsWIMFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".wim")
}

// SplitWIM splits a WIM file into smaller SWM files using wimlib-imagex
func SplitWIM(wimPath, outputDir string, maxSizeMB int) error {
	// Output will be install.swm, install2.swm, etc.
	baseName := strings.TrimSuffix(filepath.Base(wimPath), filepath.Ext(wimPath))
	outputPattern := filepath.Join(outputDir, baseName+".swm")

	cmd := exec.Command("wimlib-imagex", "split", wimPath, outputPattern, fmt.Sprintf("%d", maxSizeMB))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to split WIM file: %v", err)
	}

	return nil
}

// CopyWindowsISOWithWIMSplit copies Windows ISO contents to FAT32, splitting large WIM files
func CopyWindowsISOWithWIMSplit(srcMount, dstMount string, progressFn ProgressFunc) error {
	// Find large files
	largeFiles, err := FindLargeFiles(srcMount)
	if err != nil {
		return fmt.Errorf("failed to scan for large files: %v", err)
	}

	// Check if any large files are NOT WIM files (can't handle those on FAT32)
	for _, lf := range largeFiles {
		if !IsWIMFile(lf.RelPath) {
			return fmt.Errorf("file '%s' (%.1f GB) exceeds FAT32 4GB limit and is not a WIM file - cannot proceed with FAT32",
				lf.RelPath, float64(lf.Size)/(1024*1024*1024))
		}
	}

	// Build exclusion list for large WIM files
	var excludeFiles []string
	for _, lf := range largeFiles {
		excludeFiles = append(excludeFiles, lf.RelPath)
		fmt.Printf("Will split: %s (%.1f GB)\n", lf.RelPath, float64(lf.Size)/(1024*1024*1024))
	}

	// First pass: copy all files except large WIMs
	stats, err := calculateTotalSizeExcluding(srcMount, excludeFiles)
	if err != nil {
		return fmt.Errorf("failed to calculate total size: %v", err)
	}

	fmt.Println("Copying files (excluding large WIM files)...")
	if err := copyFilesExcluding(srcMount, dstMount, excludeFiles, stats, progressFn); err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}
	fmt.Println()

	// Second pass: split and copy large WIM files
	for _, lf := range largeFiles {
		fmt.Printf("Splitting %s...\n", lf.RelPath)

		srcWIM := filepath.Join(srcMount, lf.RelPath)
		dstDir := filepath.Join(dstMount, filepath.Dir(lf.RelPath))

		// Ensure destination directory exists
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dstDir, err)
		}

		// Split WIM directly to destination
		if err := SplitWIM(srcWIM, dstDir, SplitWIMMaxSize); err != nil {
			return fmt.Errorf("failed to split %s: %v", lf.RelPath, err)
		}

		fmt.Printf("✓ Split %s into SWM files\n", lf.RelPath)
	}

	return nil
}

// calculateTotalSizeExcluding calculates total size excluding specified files
func calculateTotalSizeExcluding(srcMount string, excludeFiles []string) (*CopyStats, error) {
	stats := &CopyStats{}
	excludeMap := make(map[string]bool)
	for _, f := range excludeFiles {
		excludeMap[f] = true
	}

	err := filepath.Walk(srcMount, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(srcMount, path)
		if excludeMap[relPath] {
			return nil
		}

		if info.Mode().IsRegular() {
			stats.TotalFiles++
			stats.TotalBytes += info.Size()
		}

		return nil
	})

	return stats, err
}

// copyFilesExcluding copies files excluding specified paths
func copyFilesExcluding(srcMount, dstMount string, excludeFiles []string, stats *CopyStats, progressFn ProgressFunc) error {
	excludeMap := make(map[string]bool)
	for _, f := range excludeFiles {
		excludeMap[f] = true
	}

	return filepath.Walk(srcMount, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			relPath, _ := filepath.Rel(srcMount, srcPath)
			stats.Failed = append(stats.Failed, relPath)
			return nil
		}

		relPath, err := filepath.Rel(srcMount, srcPath)
		if err != nil {
			return err
		}

		// Skip excluded files
		if excludeMap[relPath] {
			return nil
		}

		dstPath := filepath.Join(dstMount, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if info.Mode().IsRegular() {
			stats.CurrentFile = relPath
			if progressFn != nil {
				progressFn(stats.CopiedBytes, stats.TotalBytes, stats.CurrentFile)
			}

			if err := copyFile(srcPath, dstPath, info.Size(), stats, progressFn); err != nil {
				stats.Failed = append(stats.Failed, relPath)
				return nil
			}

			stats.CopiedFiles++
		}

		return nil
	})
}
