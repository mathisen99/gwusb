# Integration Testing Guide

This document describes the manual integration testing procedure for WoeUSB-ng Go rewrite.

## Prerequisites

### System Requirements
- Linux system with root privileges
- USB device (will be completely wiped)
- Windows ISO file for testing

### Required Dependencies
Ensure all dependencies are installed:
```bash
# Check dependencies
./woeusb-go --help

# Install missing dependencies (Ubuntu/Debian)
sudo apt install parted wipefs util-linux dosfstools ntfs-3g p7zip-full grub2-common

# Install missing dependencies (Fedora/RHEL)
sudo dnf install parted util-linux dosfstools ntfs-3g p7zip grub2-tools
```

## Test Scenarios

### Test 1: Device Mode with FAT32
**Purpose:** Test complete device wipe and FAT32 formatting

**Prerequisites:**
- Windows ISO with no files >4GB
- USB device ≥8GB

**Steps:**
1. Build the application:
   ```bash
   go build -o woeusb-go cmd/woeusb/main.go
   ```

2. Identify target USB device:
   ```bash
   lsblk
   # Note the device path (e.g., /dev/sdb)
   ```

3. Run device mode:
   ```bash
   sudo ./woeusb-go --device /path/to/windows.iso /dev/sdX
   ```

**Expected Results:**
- Device is completely wiped
- MBR partition table created
- Single FAT32 partition created and formatted
- All files copied successfully
- GRUB bootloader installed
- Boot flag set on partition
- USB boots successfully

### Test 2: Device Mode with NTFS
**Purpose:** Test NTFS formatting with large files

**Prerequisites:**
- Windows ISO with files >4GB (e.g., install.wim)
- USB device ≥8GB

**Steps:**
1. Run device mode with NTFS:
   ```bash
   sudo ./woeusb-go --device --target-filesystem NTFS /path/to/windows.iso /dev/sdX
   ```

**Expected Results:**
- Device wiped and partitioned
- NTFS partition created
- UEFI:NTFS partition created (512KB at end)
- All files copied including large ones
- GRUB and UEFI:NTFS bootloaders installed
- USB boots in both BIOS and UEFI modes

### Test 3: Partition Mode
**Purpose:** Test existing partition formatting

**Prerequisites:**
- USB device with existing partition
- Windows ISO file

**Steps:**
1. Create a partition manually:
   ```bash
   sudo parted /dev/sdX mklabel msdos
   sudo parted /dev/sdX mkpart primary 1MiB 100%
   ```

2. Run partition mode:
   ```bash
   sudo ./woeusb-go --partition /path/to/windows.iso /dev/sdX1
   ```

**Expected Results:**
- Existing partition formatted (FAT32 or NTFS as appropriate)
- Files copied successfully
- GRUB installed to partition
- Partition bootable

### Test 4: Windows 7 UEFI Workaround
**Purpose:** Test Windows 7 UEFI compatibility

**Prerequisites:**
- Windows 7 ISO file
- UEFI-capable system for testing

**Steps:**
1. Run with Windows 7 ISO:
   ```bash
   sudo ./woeusb-go --device /path/to/windows7.iso /dev/sdX
   ```

2. Check for UEFI workaround application:
   ```bash
   # Mount the created USB and check for bootx64.efi
   sudo mount /dev/sdX1 /mnt
   ls -la /mnt/efi/boot/
   sudo umount /mnt
   ```

**Expected Results:**
- Windows 7 detected automatically
- bootmgfw.efi extracted and placed as bootx64.efi
- USB boots on UEFI systems

### Test 5: Error Handling
**Purpose:** Test graceful error handling

**Test Cases:**

#### 5.1: Missing Dependencies
```bash
# Temporarily rename a required tool
sudo mv /usr/bin/parted /usr/bin/parted.bak
./woeusb-go --device test.iso /dev/sdX
sudo mv /usr/bin/parted.bak /usr/bin/parted
```
**Expected:** Clear error message listing missing dependencies

#### 5.2: Invalid Source
```bash
./woeusb-go --device /nonexistent.iso /dev/sdX
```
**Expected:** Error about source file not found

#### 5.3: Invalid Target
```bash
sudo ./woeusb-go --device test.iso /dev/nonexistent
```
**Expected:** Error about target device not found

#### 5.4: Busy Target
```bash
# Mount the target device first
sudo mount /dev/sdX1 /mnt
sudo ./woeusb-go --device test.iso /dev/sdX
sudo umount /mnt
```
**Expected:** Error about device being busy, or automatic unmount

#### 5.5: Insufficient Space
```bash
# Use a very small USB device with large ISO
sudo ./woeusb-go --device large-windows.iso /dev/small-usb
```
**Expected:** Error about insufficient space

### Test 6: Signal Handling
**Purpose:** Test cleanup on interruption

**Steps:**
1. Start a long-running operation:
   ```bash
   sudo ./woeusb-go --device large-windows.iso /dev/sdX
   ```

2. Interrupt with Ctrl+C after partitioning starts

**Expected Results:**
- Graceful shutdown message
- Temporary mountpoints unmounted
- Temporary directories cleaned up
- No orphaned processes

## Validation Checklist

After each test, verify:

### Functional Validation
- [ ] USB device boots successfully
- [ ] Windows installer starts correctly
- [ ] All expected files present on USB
- [ ] File permissions preserved
- [ ] Large files (>4GB) handled correctly
- [ ] UEFI and BIOS boot modes work (where applicable)

### System Validation
- [ ] No temporary files left behind
- [ ] No orphaned mount points
- [ ] System logs show no errors
- [ ] USB device properly unmounted

### Performance Validation
- [ ] Progress reporting works correctly
- [ ] Copy speed reasonable (>10MB/s on USB 3.0)
- [ ] Memory usage stable during large file copies
- [ ] CPU usage reasonable

## Troubleshooting

### Common Issues

#### Permission Denied
- Ensure running with sudo
- Check device permissions: `ls -la /dev/sdX`

#### Device Busy
- Check mounted filesystems: `mount | grep sdX`
- Kill processes using device: `lsof /dev/sdX`

#### Boot Failure
- Verify BIOS/UEFI boot mode matches USB preparation
- Check boot order in BIOS/UEFI settings
- Verify Windows ISO integrity

#### Slow Performance
- Check USB device speed (USB 2.0 vs 3.0)
- Verify sufficient system RAM
- Check for disk errors: `dmesg | tail`

## Test Environment Setup

### Virtual Machine Testing
For safer testing, use VMs:

```bash
# Create VM disk image
qemu-img create -f qcow2 test-usb.qcow2 8G

# Boot VM with USB image
qemu-system-x86_64 -m 2048 -boot d -cdrom test-usb.qcow2
```

### Automated Testing Script
```bash
#!/bin/bash
# Basic automated test runner
set -e

ISO_FILE="$1"
USB_DEVICE="$2"

if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root"
   exit 1
fi

echo "Testing WoeUSB-ng with $ISO_FILE on $USB_DEVICE"

# Test 1: Device mode
echo "Test 1: Device mode..."
./woeusb-go --device "$ISO_FILE" "$USB_DEVICE"

# Verify result
mount "${USB_DEVICE}1" /mnt
ls -la /mnt/
umount /mnt

echo "All tests passed!"
```

## Reporting Issues

When reporting test failures, include:
- Complete command used
- Full error output
- System information (`uname -a`, `lsb_release -a`)
- USB device information (`lsblk`, `fdisk -l`)
- ISO file information (`file`, `ls -la`)
- Relevant log entries (`dmesg`, `/var/log/syslog`)
