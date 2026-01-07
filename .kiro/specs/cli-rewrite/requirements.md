# Requirements: CLI Rewrite (Go)

## Introduction

This spec covers a complete Go rewrite of the WoeUSB-ng CLI tool. The tool creates bootable Windows USB installation drives from ISO images or DVDs on Linux.

- Single static binary with no Python/wxPython runtime dependency
- Two installation modes: device (wipe entire USB) and partition (use existing partition)
- Support FAT32 and NTFS target filesystems
- Auto-switch to NTFS when source contains files >4GB
- Install GRUB bootloader for legacy BIOS boot support
- Include workarounds for Windows 7 UEFI boot and buggy BIOSes
- Progress reporting during file copy operations
- Proper cleanup on success, failure, or interruption

## User Stories

1. **US-1:** As a Linux user, I want to create a bootable Windows USB by wiping an entire USB device, so I can perform a clean Windows installation.

2. **US-2:** As a Linux user, I want to create a bootable Windows USB on an existing partition, so I can preserve other partitions on the device.

3. **US-3:** As a user, I want the tool to automatically switch to NTFS when my ISO contains files larger than 4GB, so I don't encounter FAT32 limitations.

4. **US-4:** As a user, I want to see progress during file copying, so I know the operation is proceeding.

5. **US-5:** As a user, I want the tool to check for required system dependencies before starting, so I get clear error messages if something is missing.

6. **US-6:** As a user, I want the tool to clean up (unmount, remove temp dirs) even if the operation fails or I interrupt it.

7. **US-7:** As a user with a buggy BIOS, I want to set the boot flag on the partition, so my system includes the USB in the boot menu.

8. **US-8:** As a user installing Windows 7, I want the tool to apply UEFI boot workarounds automatically.

## Acceptance Criteria

### AC-1: Argument Parsing
WHEN the user runs `woeusb --device <source> <target>`
THE SYSTEM SHALL wipe the target device and create a bootable Windows USB.

WHEN the user runs `woeusb --partition <source> <target>`
THE SYSTEM SHALL copy Windows files to the existing partition without wiping the device.

WHEN neither `--device` nor `--partition` is specified
THE SYSTEM SHALL print an error and exit with non-zero status.

### AC-2: Dependency Checking
WHEN the tool starts
THE SYSTEM SHALL verify that wipefs, parted, lsblk, blockdev, mount, umount, 7z, mkdosfs (or variant), mkntfs, and grub-install (or grub2-install) are available.

WHEN any required dependency is missing
THE SYSTEM SHALL print which dependency is missing and exit with non-zero status.

### AC-3: Source/Target Validation
WHEN the source is not a regular file or block device
THE SYSTEM SHALL print an error and exit.

WHEN the target is not a block device
THE SYSTEM SHALL print an error and exit.

WHEN `--device` mode is used and target ends with a digit (partition)
THE SYSTEM SHALL print an error indicating a whole device is required.

WHEN `--partition` mode is used and target does not end with a digit
THE SYSTEM SHALL print an error indicating a partition is required.

### AC-4: FAT32 Auto-Switch
WHEN the source filesystem contains any file larger than 4GB and `--target-filesystem` is FAT
THE SYSTEM SHALL automatically switch to NTFS and print a warning.

### AC-5: Device Wipe (Device Mode)
WHEN in device mode
THE SYSTEM SHALL run `wipefs --all` on the target device.

WHEN the device still shows partitions after wipe
THE SYSTEM SHALL print an error about a possibly locked/readonly drive and exit.

### AC-6: Partition Creation (Device Mode)
WHEN in device mode with FAT filesystem
THE SYSTEM SHALL create an MBR partition table and a single FAT32 partition starting at 4MiB.

WHEN in device mode with NTFS filesystem
THE SYSTEM SHALL create an MBR partition table, an NTFS partition, and a 512KB UEFI:NTFS support partition.

### AC-7: File Copying
WHEN copying files from source to target
THE SYSTEM SHALL preserve directory structure and copy all files.

WHEN copying files
THE SYSTEM SHALL print progress (bytes copied / total bytes).

### AC-8: Bootloader Installation
WHEN `--workaround-skip-grub` is NOT set
THE SYSTEM SHALL install GRUB with `--target=i386-pc` and create a grub.cfg containing `ntldr /bootmgr` and `boot`.

WHEN `--workaround-skip-grub` IS set
THE SYSTEM SHALL skip GRUB installation.

### AC-9: Windows 7 UEFI Workaround
WHEN the source appears to be Windows 7 (MinServer=7xxx in cversion.ini) and no EFI bootloader exists
THE SYSTEM SHALL extract bootmgfw.efi from install.wim and place it at /efi/boot/bootx64.efi.

### AC-10: Boot Flag Workaround
WHEN `--workaround-bios-boot-flag` is set
THE SYSTEM SHALL run `parted set 1 boot on` on the target device.

### AC-11: Cleanup
WHEN the operation completes (success or failure)
THE SYSTEM SHALL unmount any mounted filesystems and remove temporary directories.

WHEN interrupted by SIGINT/SIGTERM
THE SYSTEM SHALL perform cleanup before exiting.

### AC-12: Exit Codes
WHEN the operation succeeds
THE SYSTEM SHALL exit with code 0.

WHEN the operation fails
THE SYSTEM SHALL exit with a non-zero code.
