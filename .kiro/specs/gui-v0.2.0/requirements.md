# Requirements Document

## Introduction

WoeUSB-go v0.2.0 adds a graphical user interface (GUI) to the existing CLI tool, making it accessible to users who prefer visual interaction. The GUI will be simple, beginner-friendly, and work across all major Linux distributions. It will also include distro-aware dependency checking that provides exact install commands for missing packages.

## Glossary

- **GUI**: Graphical User Interface - the visual application window
- **USB_Device**: A removable USB storage device (flash drive, external HDD via USB)
- **ISO_File**: A Windows 10 or Windows 11 installation image file (.iso extension)
- **Distro**: Linux distribution (Ubuntu, Fedora, Arch, openSUSE, etc.)
- **Package_Manager**: System tool for installing software (apt, dnf, pacman, zypper)
- **Fyne**: Cross-platform Go GUI toolkit used for building the interface
- **Device_Selector**: GUI component for choosing target USB device
- **File_Browser**: GUI component for selecting ISO files

## Requirements

### Requirement 1: GUI Application Launch

**User Story:** As a user, I want to launch a graphical application, so that I can create bootable USB drives without using the command line.

#### Acceptance Criteria

1. WHEN the user runs `woeusb-go --gui` THEN the GUI_Application SHALL display a window with the main interface
2. WHEN the GUI_Application starts THEN it SHALL check for root privileges and display a warning if not running as root
3. WHEN the GUI_Application starts THEN it SHALL verify all required dependencies are installed before showing the main interface
4. IF required dependencies are missing THEN the GUI_Application SHALL display the Dependency_Dialog with distro-specific install instructions

### Requirement 2: USB Device Selection

**User Story:** As a user, I want to select only USB devices from a list, so that I cannot accidentally wipe my internal hard drives.

#### Acceptance Criteria

1. WHEN the Device_Selector is displayed THEN the GUI_Application SHALL list only removable USB storage devices
2. WHEN listing devices THEN the GUI_Application SHALL exclude internal hard drives, SSDs, and NVMe drives
3. WHEN a USB device is listed THEN the GUI_Application SHALL display the device name, size, and mount path
4. WHEN the user selects a USB device THEN the GUI_Application SHALL store the selection for the write operation
5. WHEN no USB devices are detected THEN the GUI_Application SHALL display a message indicating no USB devices found
6. WHEN the user clicks a refresh button THEN the Device_Selector SHALL rescan for USB devices

### Requirement 3: ISO File Selection

**User Story:** As a user, I want to browse and select a Windows ISO file, so that I can choose the installation image to write.

#### Acceptance Criteria

1. WHEN the user clicks the browse button THEN the File_Browser SHALL open a file selection dialog
2. WHEN the File_Browser opens THEN it SHALL filter to show only .iso files by default
3. WHEN the user selects an ISO file THEN the GUI_Application SHALL validate it is a readable file
4. WHEN a valid ISO is selected THEN the GUI_Application SHALL display the filename in the interface
5. IF the selected file is not a valid ISO THEN the GUI_Application SHALL display an error message

### Requirement 4: Write Operation

**User Story:** As a user, I want to start the USB creation process with a single button click, so that I can easily create bootable media.

#### Acceptance Criteria

1. WHEN both USB device and ISO file are selected THEN the Start_Button SHALL become enabled
2. WHEN the user clicks the Start_Button THEN the GUI_Application SHALL display a confirmation dialog warning about data loss
3. WHEN the user confirms the operation THEN the GUI_Application SHALL begin the USB creation process
4. WHILE the write operation is in progress THEN the GUI_Application SHALL display a progress indicator
5. WHILE the write operation is in progress THEN the GUI_Application SHALL disable the Start_Button and selection controls
6. WHEN the write operation completes successfully THEN the GUI_Application SHALL display a success message
7. IF the write operation fails THEN the GUI_Application SHALL display an error message with details

### Requirement 5: Progress Reporting

**User Story:** As a user, I want to see the progress of the USB creation, so that I know how long to wait and that the process is working.

#### Acceptance Criteria

1. WHILE copying files THEN the GUI_Application SHALL display a progress bar showing percentage complete
2. WHILE copying files THEN the GUI_Application SHALL display the current operation status (e.g., "Copying files...", "Installing bootloader...")
3. WHEN the operation is in progress THEN the GUI_Application SHALL remain responsive to user interaction
4. WHEN the user attempts to close the window during operation THEN the GUI_Application SHALL warn about interrupting the process

### Requirement 6: Distro-Aware Dependency Checking

**User Story:** As a user, I want to see exactly what commands to run to install missing dependencies for my specific Linux distribution, so that I can quickly resolve any missing requirements.

#### Acceptance Criteria

1. WHEN checking dependencies THEN the Dependency_Checker SHALL detect the Linux distribution from /etc/os-release
2. WHEN a dependency is missing THEN the Dependency_Dialog SHALL display the exact package name for the detected distro
3. WHEN displaying install instructions THEN the Dependency_Dialog SHALL show the complete command (e.g., `sudo apt install wimtools`)
4. THE Dependency_Checker SHALL support Ubuntu/Debian (apt), Fedora (dnf), Arch Linux (pacman), openSUSE (zypper), and Linux Mint (apt)
5. WHEN the distro is not recognized THEN the Dependency_Dialog SHALL display generic package names with a note to check distro documentation
6. WHEN all dependencies are satisfied THEN the GUI_Application SHALL proceed to the main interface without showing the Dependency_Dialog

### Requirement 7: Cross-Distribution Compatibility

**User Story:** As a user on any major Linux distribution, I want the GUI to work without additional configuration, so that I can use the tool regardless of my distro choice.

#### Acceptance Criteria

1. THE GUI_Application SHALL use the Fyne toolkit for cross-platform rendering
2. THE GUI_Application SHALL not require GTK, Qt, or other desktop-specific libraries at runtime
3. THE GUI_Application SHALL compile to a single static binary including GUI components
4. THE GUI_Application SHALL function on X11 and Wayland display servers

## Verified Package Names by Distribution

The following package names have been verified as of January 2026:

### Ubuntu 25.10 "Questing" / 26.04 "Resolute" (apt)
**Required:**
- `wimtools` - provides wimlib-imagex (version 1.14.4 in universe repo)
- `p7zip-full` - provides 7z command
- `dosfstools` - provides mkdosfs/mkfs.vfat
- `parted` - provides parted
- `util-linux` - provides wipefs, lsblk, blockdev, mount, umount (usually pre-installed)

**Optional (for additional features):**
- `grub-pc` - provides grub-install (for legacy BIOS boot support)
- `ntfs-3g` - provides mkntfs (for NTFS filesystem support)

Install all required: `sudo apt install wimtools p7zip-full dosfstools parted`
Install with optional: `sudo apt install wimtools p7zip-full dosfstools parted grub-pc ntfs-3g`

### Debian 13 "Trixie" (apt)
**Required:**
- `wimtools` - provides wimlib-imagex (version 1.14.4-1.1+b3)
- `p7zip-full` - provides 7z command
- `dosfstools` - provides mkdosfs/mkfs.vfat
- `parted` - provides parted
- `util-linux` - provides wipefs, lsblk, blockdev, mount, umount (usually pre-installed)

**Optional (for additional features):**
- `grub-pc` - provides grub-install (for legacy BIOS boot support)
- `ntfs-3g` - provides mkntfs (for NTFS filesystem support)

Install all required: `sudo apt install wimtools p7zip-full dosfstools parted`
Install with optional: `sudo apt install wimtools p7zip-full dosfstools parted grub-pc ntfs-3g`

### Linux Mint 22.x "Wilma/Xia/Zara" (apt - based on Ubuntu Noble)
**Required:**
- `wimtools` - provides wimlib-imagex (same as Ubuntu)
- `p7zip-full` - provides 7z command
- `dosfstools` - provides mkdosfs/mkfs.vfat
- `parted` - provides parted
- `util-linux` - provides wipefs, lsblk, blockdev, mount, umount (usually pre-installed)

**Optional (for additional features):**
- `grub-pc` - provides grub-install (for legacy BIOS boot support)
- `ntfs-3g` - provides mkntfs (for NTFS filesystem support)

Install all required: `sudo apt install wimtools p7zip-full dosfstools parted`
Install with optional: `sudo apt install wimtools p7zip-full dosfstools parted grub-pc ntfs-3g`

### Fedora 42/43 (dnf5)
**Required:**
- `wimlib-utils` - provides wimlib-imagex
- `p7zip-plugins` - provides 7z command
- `dosfstools` - provides mkdosfs/mkfs.vfat
- `parted` - provides parted
- `util-linux` - provides wipefs, lsblk, blockdev, mount, umount (usually pre-installed)

**Optional (for additional features):**
- `grub2-pc` - provides grub2-install (for legacy BIOS boot support)
- `ntfs-3g` - provides mkntfs (for NTFS filesystem support)

Install all required: `sudo dnf install wimlib-utils p7zip-plugins dosfstools parted`
Install with optional: `sudo dnf install wimlib-utils p7zip-plugins dosfstools parted grub2-pc ntfs-3g`

### Arch Linux (pacman)
**Required:**
- `wimlib` - provides wimlib-imagex (version 1.14.4-2 in extra repo, updated March 2025)
- `p7zip` - provides 7z command
- `dosfstools` - provides mkdosfs/mkfs.vfat
- `parted` - provides parted
- `util-linux` - provides wipefs, lsblk, blockdev, mount, umount (usually pre-installed)

**Optional (for additional features):**
- `grub` - provides grub-install (for legacy BIOS boot support)
- `ntfs-3g` - provides mkntfs (for NTFS filesystem support)

Install all required: `sudo pacman -S wimlib p7zip dosfstools parted`
Install with optional: `sudo pacman -S wimlib p7zip dosfstools parted grub ntfs-3g`

### openSUSE Leap 16.0 / Tumbleweed (zypper)
**Required:**
- `wimlib` - provides wimlib-imagex (version 1.14.4 in official repo for Tumbleweed, devel:libraries:c_c++ for Leap)
- `p7zip-full` - provides 7z command
- `dosfstools` - provides mkdosfs/mkfs.vfat
- `parted` - provides parted
- `util-linux` - provides wipefs, lsblk, blockdev, mount, umount (usually pre-installed)

**Optional (for additional features):**
- `grub2` - provides grub2-install (for legacy BIOS boot support)
- `ntfs-3g` - provides mkntfs (for NTFS filesystem support)

Install all required (Tumbleweed): `sudo zypper install wimlib p7zip-full dosfstools parted`
Install with optional (Tumbleweed): `sudo zypper install wimlib p7zip-full dosfstools parted grub2 ntfs-3g`

Install all required (Leap 16): 
```
sudo zypper addrepo https://download.opensuse.org/repositories/devel:/libraries:/c_c++/openSUSE_Leap_16.0/ devel-libs
sudo zypper refresh
sudo zypper install wimlib p7zip-full dosfstools parted
```

Note: On openSUSE Leap 16.0, wimlib requires adding the devel:libraries:c_c++ repository first.
