package bootloader

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsWindows7 checks if the source contains Windows 7 by examining cversion.ini
func IsWindows7(srcMount string) (bool, error) {
	cversionPath := filepath.Join(srcMount, "sources", "cversion.ini")

	file, err := os.Open(cversionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File doesn't exist, not Windows 7
		}
		return false, fmt.Errorf("failed to open cversion.ini: %v", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "MinServer=") {
			version := strings.TrimPrefix(line, "MinServer=")
			// Windows 7 versions start with 7
			if strings.HasPrefix(version, "7") {
				return true, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading cversion.ini: %v", err)
	}

	return false, nil
}

// ExtractBootloader extracts bootmgfw.efi from Windows 7 sources using 7z
func ExtractBootloader(srcMount, dstMount string) error {
	// Look for install.wim or install.esd in sources directory
	sourcesDir := filepath.Join(srcMount, "sources")
	var installFile string

	// Check for install.wim first, then install.esd
	wimPath := filepath.Join(sourcesDir, "install.wim")
	esdPath := filepath.Join(sourcesDir, "install.esd")

	if _, err := os.Stat(wimPath); err == nil {
		installFile = wimPath
	} else if _, err := os.Stat(esdPath); err == nil {
		installFile = esdPath
	} else {
		return fmt.Errorf("neither install.wim nor install.esd found in sources directory")
	}

	// Create EFI boot directory
	efiBootDir := filepath.Join(dstMount, "efi", "boot")
	if err := os.MkdirAll(efiBootDir, 0755); err != nil {
		return fmt.Errorf("failed to create EFI boot directory: %v", err)
	}

	// Extract bootmgfw.efi using 7z
	bootloaderPath := filepath.Join(efiBootDir, "bootx64.efi")

	// Use 7z to extract bootmgfw.efi from the install file
	// The path in the WIM/ESD is typically: 1/Windows/Boot/EFI/bootmgfw.efi
	cmd := exec.Command("7z", "e", "-so", installFile, "1/Windows/Boot/EFI/bootmgfw.efi")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to extract bootmgfw.efi with 7z: %v", err)
	}

	// Write the extracted bootloader to bootx64.efi
	if err := os.WriteFile(bootloaderPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write bootx64.efi: %v", err)
	}

	return nil
}

// ApplyWindows7UEFIWorkaround applies the complete Windows 7 UEFI workaround
func ApplyWindows7UEFIWorkaround(srcMount, dstMount string) error {
	// First check if this is Windows 7
	isWin7, err := IsWindows7(srcMount)
	if err != nil {
		return fmt.Errorf("failed to check Windows version: %v", err)
	}

	if !isWin7 {
		return nil // Not Windows 7, no workaround needed
	}

	// Extract and place the bootloader
	if err := ExtractBootloader(srcMount, dstMount); err != nil {
		return fmt.Errorf("failed to extract bootloader: %v", err)
	}

	return nil
}

// CheckUEFIBootloader verifies that the UEFI bootloader is properly installed
func CheckUEFIBootloader(dstMount string) error {
	bootloaderPath := filepath.Join(dstMount, "efi", "boot", "bootx64.efi")

	info, err := os.Stat(bootloaderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("UEFI bootloader not found at %s", bootloaderPath)
		}
		return fmt.Errorf("failed to check UEFI bootloader: %v", err)
	}

	// Check that the file is not empty
	if info.Size() == 0 {
		return fmt.Errorf("UEFI bootloader file is empty: %s", bootloaderPath)
	}

	return nil
}

// InstallGRUB installs GRUB bootloader to the specified device
func InstallGRUB(mountpoint, device, grubCmd string) error {
	// Prepare grub-install arguments
	args := []string{
		"--target=i386-pc",
		"--boot-directory=" + filepath.Join(mountpoint, "boot"),
		"--force",
		device,
	}

	cmd := exec.Command(grubCmd, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install GRUB with %s: %v", grubCmd, err)
	}

	return nil
}

// WriteGRUBConfig writes a GRUB configuration file
func WriteGRUBConfig(mountpoint, grubPrefix string) error {
	// Determine the correct boot directory based on grub prefix
	var bootDir string
	if strings.Contains(grubPrefix, "grub2") {
		bootDir = filepath.Join(mountpoint, "boot", "grub2")
	} else {
		bootDir = filepath.Join(mountpoint, "boot", "grub")
	}

	// Ensure boot directory exists
	if err := os.MkdirAll(bootDir, 0755); err != nil {
		return fmt.Errorf("failed to create boot directory %s: %v", bootDir, err)
	}

	// Write grub.cfg
	grubCfgPath := filepath.Join(bootDir, "grub.cfg")
	grubConfig := generateGRUBConfig(grubPrefix)

	if err := os.WriteFile(grubCfgPath, []byte(grubConfig), 0644); err != nil {
		return fmt.Errorf("failed to write GRUB config to %s: %v", grubCfgPath, err)
	}

	return nil
}

// DetectGRUBPrefix detects whether to use grub or grub2 prefix from command name
func DetectGRUBPrefix(grubCmd string) string {
	if strings.Contains(grubCmd, "grub2") {
		return "grub2"
	}
	return "grub"
}

// generateGRUBConfig generates a basic GRUB configuration for Windows USB
func generateGRUBConfig(grubPrefix string) string {
	return `# GRUB configuration for Windows USB
# Generated by WoeUSB-ng

set timeout=10
set default=0

menuentry "Windows" {
    insmod part_msdos
    insmod ntfs
    insmod search_fs_uuid
    insmod chain
    search --fs-uuid --set=root --hint-bios=hd0,msdos1 --hint-efi=hd0,msdos1 --hint-baremetal=ahci0,msdos1
    chainloader +1
}

menuentry "Windows (fallback)" {
    insmod part_msdos
    insmod fat
    insmod search_fs_uuid
    insmod chain
    search --fs-uuid --set=root --hint-bios=hd0,msdos1 --hint-efi=hd0,msdos1 --hint-baremetal=ahci0,msdos1
    chainloader +1
}
`
}

// InstallGRUBWithConfig installs GRUB and writes configuration in one step
func InstallGRUBWithConfig(mountpoint, device, grubCmd string) error {
	// Install GRUB
	if err := InstallGRUB(mountpoint, device, grubCmd); err != nil {
		return fmt.Errorf("GRUB installation failed: %v", err)
	}

	// Detect prefix and write config
	grubPrefix := DetectGRUBPrefix(grubCmd)
	if err := WriteGRUBConfig(mountpoint, grubPrefix); err != nil {
		return fmt.Errorf("GRUB configuration failed: %v", err)
	}

	return nil
}

// CheckGRUBInstallation verifies that GRUB was installed correctly
func CheckGRUBInstallation(mountpoint, grubPrefix string) error {
	// Check for boot directory
	var bootDir string
	if grubPrefix == "grub2" {
		bootDir = filepath.Join(mountpoint, "boot", "grub2")
	} else {
		bootDir = filepath.Join(mountpoint, "boot", "grub")
	}

	if _, err := os.Stat(bootDir); os.IsNotExist(err) {
		return fmt.Errorf("GRUB boot directory not found: %s", bootDir)
	}

	// Check for grub.cfg
	grubCfgPath := filepath.Join(bootDir, "grub.cfg")
	if _, err := os.Stat(grubCfgPath); os.IsNotExist(err) {
		return fmt.Errorf("GRUB configuration file not found: %s", grubCfgPath)
	}

	return nil
}

// GetGRUBVersion attempts to get the version of the GRUB command
func GetGRUBVersion(grubCmd string) (string, error) {
	cmd := exec.Command(grubCmd, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GRUB version: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}
