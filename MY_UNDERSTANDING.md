# Project Understanding (Python Source Quarantined)

## Executive Summary

**WoeUSB-ng** is a Linux utility that creates bootable Windows USB installation drives from ISO images or physical DVDs. It's a Python rewrite of the original bash-based WoeUSB project.

The tool handles the complete workflow: wiping the target USB device, creating an MBR partition table, formatting with FAT32 or NTFS, copying Windows installation files, and installing a GRUB bootloader for legacy BIOS boot support. It also includes workarounds for Windows 7 UEFI booting and buggy motherboard BIOSes.

**Target users:** Linux users who need to create Windows installation media without access to a Windows machine.

## How to Run the Original (Python) Project

### Installation
```bash
pip install WoeUSB-ng
# Or from source:
sudo python setup.py install
```

### CLI Usage (requires root)
```bash
# Device mode - wipes entire USB drive
sudo woeusb --device /path/to/windows.iso /dev/sdX

# Partition mode - writes to existing partition
sudo woeusb --partition /path/to/windows.iso /dev/sdX1

# With options
sudo woeusb --device --target-filesystem NTFS --label "WIN10" /path/to/windows.iso /dev/sdX
```

### GUI Usage
```bash
woeusbgui  # Uses pkexec for privilege escalation
```

### CLI Arguments
| Argument | Description |
|----------|-------------|
| `--device`, `-d` | Wipe entire device and create fresh partition |
| `--partition`, `-p` | Write to existing partition |
| `--target-filesystem` | `FAT` (default) or `NTFS` |
| `--label`, `-l` | Filesystem label (default: "Windows USB") |
| `--workaround-bios-boot-flag` | Set boot flag for buggy BIOSes |
| `--workaround-skip-grub` | Skip GRUB installation (UEFI-only boot) |
| `--verbose`, `-v` | Verbose output |
| `--no-color` | Disable colored output |

### Required System Dependencies
- `wipefs`, `parted`, `lsblk`, `blockdev`, `df`, `mount`, `umount`
- `mkdosfs` / `mkfs.vfat` / `mkfs.fat` (from dosfstools)
- `mkntfs` (from ntfs-3g)
- `grub-install` or `grub2-install`
- `7z` (from p7zip)

## Architecture Overview

### Module Structure

```
WoeUSB/
├── woeusb           # CLI entry point → core.run()
├── woeusbgui        # GUI entry point → gui.run() (with pkexec)
├── core.py          # Main orchestration, disk ops, file copying
├── utils.py         # Validation, dependency checks, helpers
├── gui.py           # wxPython GUI application
├── list_devices.py  # USB/DVD enumeration via lsblk/sysfs
├── workaround.py    # Win7 UEFI fix, partition table refresh
├── miscellaneous.py # Version string, i18n setup
├── data/            # Icons (icon.ico, logos)
└── locale/          # Translations (de, fr, pl, pt_BR, sv, tr, zh)
```

### Execution Flow (Device Mode)

```
1. Parse arguments (core.init)
2. Check dependencies (utils.check_runtime_dependencies)
   - Verify: wipefs, parted, lsblk, blockdev, df, 7z
   - Find: mkdosfs/mkfs.vfat, mkntfs, grub-install
3. Validate source/target (utils.check_runtime_parameters)
4. Check source/target not mounted (utils.check_source_and_target_not_busy)
5. Mount source ISO/DVD (core.mount_source_filesystem)
6. Check FAT32 4GB limit (utils.check_fat32_filesize_limitation)
   - Auto-switch to NTFS if files exceed 4GB
7. Wipe target device (core.wipe_existing_partition_table_and_filesystem_signatures)
   - wipefs --all /dev/sdX
8. Create MBR partition table (core.create_target_partition_table)
   - parted --script /dev/sdX mklabel msdos
9. Create partition (core.create_target_partition)
   - parted mkpart primary fat32/ntfs 4MiB 100%/-2049s
   - mkdosfs -F 32 OR mkntfs --quick
10. [NTFS only] Create UEFI:NTFS support partition
    - 512KB partition at end of disk
    - Download uefi-ntfs.img from GitHub, dd to partition
11. Mount target partition (core.mount_target_filesystem)
12. Check free space (utils.check_target_filesystem_free_space)
13. Copy files (core.copy_filesystem_files)
    - Walk source, create dirs, copy files
    - Large files (>5MB) copied in chunks for cancellation support
    - Progress reported via ReportCopyProgress thread
14. Windows 7 UEFI workaround (workaround.support_windows_7_uefi_boot)
    - Extract bootmgfw.efi from install.wim using 7z
    - Place at /efi/boot/bootx64.efi
15. Install GRUB (core.install_legacy_pc_bootloader_grub)
    - grub-install --target=i386-pc --boot-directory=<mount> --force /dev/sdX
16. Write GRUB config (core.install_legacy_pc_bootloader_grub_config)
    - Creates grub.cfg with: ntldr /bootmgr; boot
17. [Optional] Set boot flag (workaround.buggy_motherboards_that_ignore_disks_without_boot_flag_toggled)
    - parted set 1 boot on
18. Cleanup (core.cleanup)
    - Unmount source and target
    - Remove temp directories
```

### Key Data Structures

**Global State (in core.py):**
- `current_state`: String tracking execution phase for cleanup logic
- `gui`: Reference to GUI handler for progress updates
- `CopyFiles_handle`: Thread reference for file copy progress

**Mountpoints:** Generated with timestamp + PID for uniqueness:
- `/media/woeusb_source_<timestamp>_<pid>`
- `/media/woeusb_target_<timestamp>_<pid>`

### Threading Model

1. **CLI mode:** Single-threaded except for `ReportCopyProgress` thread that polls target directory size
2. **GUI mode:** `WoeUSB_handler` thread runs core logic; main thread handles UI and progress dialog
3. **Cancellation:** `check_kill_signal()` polled throughout; raises `SystemExit` if `gui.kill` is True

## External Integrations

### System Commands (via subprocess)

| Command | Usage |
|---------|-------|
| `mount` | Mount ISO/DVD and target partition |
| `umount` | Unmount filesystems |
| `wipefs --all` | Clear partition signatures |
| `lsblk` | List block devices, get size/model/fstype |
| `blockdev --rereadpt` | Force kernel to re-read partition table |
| `parted` | Create partition table and partitions |
| `mkdosfs -F 32` | Format FAT32 |
| `mkntfs --quick` | Format NTFS |
| `grub-install` | Install GRUB bootloader |
| `7z e -so` | Extract files from WIM to stdout |
| `df` | Check free space |
| `find` | Locate EFI directories (case-insensitive) |
| `grep` | Check Windows version in cversion.ini |

### Network

- **UEFI:NTFS image download:** `https://github.com/pbatard/rufus/raw/master/res/uefi/uefi-ntfs.img`
  - Only needed for NTFS mode
  - Failure is non-fatal (warning printed)

### Filesystem Contracts

- **Source:** ISO 9660 or UDF filesystem (Windows installation media)
- **Target:** FAT32 or NTFS on MBR partition table
- **Polkit policy:** `/usr/share/polkit-1/actions/com.github.woeusb.woeusb-ng.policy`
- **Desktop entry:** `/usr/share/applications/WoeUSB-ng.desktop`
- **Icon:** `/usr/share/icons/WoeUSB-ng/icon.ico`

## Dependency Map

### Python Packages

| Package | Purpose | Required | Go Alternative |
|---------|---------|----------|----------------|
| `termcolor` | Colored terminal output | Optional | ANSI codes or `fatih/color` |
| `wxPython` | GUI toolkit | GUI only | Skip initially, or `fyne-io/fyne` |

### System Tools

| Tool | Package | Purpose | Go Approach |
|------|---------|---------|-------------|
| `wipefs` | util-linux | Clear signatures | Shell out |
| `parted` | parted | Partitioning | Shell out or `diskfs/go-diskfs` |
| `lsblk` | util-linux | Device info | Parse `/sys/block/` or `lsblk --json` |
| `blockdev` | util-linux | Reread partition table | `ioctl(BLKRRPART)` |
| `mount`/`umount` | util-linux | Mount filesystems | `syscall.Mount/Unmount` |
| `mkdosfs` | dosfstools | Format FAT32 | Shell out |
| `mkntfs` | ntfs-3g | Format NTFS | Shell out |
| `grub-install` | grub2 | Install bootloader | Shell out |
| `7z` | p7zip | Extract WIM files | Shell out or `mholt/archiver` |
| `df` | coreutils | Free space | `syscall.Statfs` |

### Unnecessary in Go Rewrite

- `awk` subprocess (used once, trivially replaceable)
- `grep` subprocess (used for simple string matching)
- `find` subprocess (Go's `filepath.Walk` or `filepath.Glob`)
- XML DOM parsing for polkit (use `encoding/xml`)

## Rewrite Plan (Go)

### Proposed Structure

```
woeusb-go/
├── cmd/
│   └── woeusb/
│       └── main.go          # CLI entry point
├── internal/
│   ├── device/
│   │   ├── list.go          # USB/DVD enumeration
│   │   └── partition.go     # Partition operations (parted wrapper)
│   ├── filesystem/
│   │   ├── mount.go         # Mount/unmount operations
│   │   ├── copy.go          # File copying with progress
│   │   └── format.go        # mkdosfs/mkntfs wrappers
│   ├── bootloader/
│   │   ├── grub.go          # GRUB installation
│   │   └── uefi.go          # UEFI:NTFS and Win7 workarounds
│   └── ui/
│       └── progress.go      # Progress reporting interface
├── go.mod
└── go.sum
```

### Key Packages to Create

1. **`device`** - Block device discovery and validation
   - Parse `/sys/block/` for removable devices
   - Wrap `lsblk --json` for device metadata
   - Validate source/target parameters

2. **`filesystem`** - Mount and copy operations
   - Use `syscall.Mount/Unmount` directly
   - Implement file copy with progress callback
   - Check FAT32 4GB limit

3. **`partition`** - Disk partitioning
   - Wrap `wipefs`, `parted`, `mkdosfs`, `mkntfs`
   - Handle UEFI:NTFS partition creation

4. **`bootloader`** - GRUB and UEFI support
   - Wrap `grub-install`
   - Implement Windows 7 EFI workaround (7z extraction)

### Replacement Strategy

| Python Component | Go Replacement |
|------------------|----------------|
| `argparse` | `spf13/cobra` or stdlib `flag` |
| `subprocess.run` | `os/exec` |
| `shutil.copy2` | `io.Copy` with metadata preservation |
| `os.walk` | `filepath.WalkDir` |
| `threading.Thread` | Goroutines + channels |
| `termcolor` | ANSI escape codes |
| `urllib.request` | `net/http` |
| `xml.dom.minidom` | `encoding/xml` |
| Global state | Struct with methods |

### Incremental Milestones

1. **M1: Device enumeration** - List USB drives, validate parameters
2. **M2: Partition operations** - Wipe, create table, create partition, format
3. **M3: File copying** - Mount source, copy files with progress
4. **M4: Bootloader** - GRUB installation and config
5. **M5: Workarounds** - Windows 7 UEFI, boot flag, UEFI:NTFS
6. **M6: Polish** - Error messages, i18n, cleanup handling

## Risks / Unknowns / Questions

### Confirmed Behaviors
- FAT32 auto-switch threshold: files > 4GB (`2^32 - 1` bytes)
- GRUB config is minimal: just `ntldr /bootmgr` + `boot`
- UEFI:NTFS partition is exactly 1024 sectors (512KB on 512-byte sector disks)
- Windows 7 detection: checks `MinServer=7xxx.x` in `sources/cversion.ini`

### Unknowns
1. **UEFI:NTFS image stability** - Downloaded from Rufus repo; URL may change
2. **grub-install variations** - Different distros use `grub-install` vs `grub2-install`
3. **Partition alignment** - 4MiB start is hardcoded; may not be optimal for all drives
4. **Large file copy interruption** - Current 5MB chunk size is arbitrary

### Questions for Testing
1. Does the tool work with Windows 11 ISOs? (likely yes, but untested in code)
2. What happens with GPT-partitioned USB drives? (code explicitly rejects GPT)
3. How does it handle NVMe devices? (`/dev/nvme0n1p1` naming differs)

### Risky Patterns to Fix in Go
1. **Global mutable state** → Use a `Session` struct passed through functions
2. **Integer return codes** → Use proper `error` types
3. **Thread communication via shared vars** → Use channels
4. **No cleanup on panic** → Use `defer` for all cleanup
5. **Hardcoded paths** (`/media/woeusb_*`) → Use `os.MkdirTemp`
