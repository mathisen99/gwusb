---
inclusion: always
---

# Product: WoeUSB-ng Go Rewrite

## Purpose & Problem Statement

WoeUSB-ng creates bootable Windows USB installation drives from ISO images or DVDs on Linux. The original Python implementation requires Python 3, termcolor, and wxPython (for GUI), creating dependency sprawl and installation friction.

This Go rewrite produces a single static binary with no runtime dependencies beyond standard Linux system tools.

## Target Users

- Linux users who need to create Windows installation media without access to a Windows machine
- System administrators preparing deployment media
- Users dual-booting or reinstalling Windows

## Core Capabilities

- Create bootable Windows USB by wiping entire device (device mode)
- Create bootable Windows USB on existing partition (partition mode)
- Support FAT32 and NTFS target filesystems
- Auto-switch to NTFS when source contains files >4GB
- Install GRUB bootloader for legacy BIOS boot support
- Apply Windows 7 UEFI boot workaround automatically
- Apply boot flag workaround for buggy BIOSes
- Report progress during file copy operations
- Clean up on success, failure, or interruption (SIGINT/SIGTERM)

## Non-Goals / Out of Scope (This Phase)

- GUI application (deferred to future spec)
- Internationalization/translations (deferred)
- Polkit policy installation (CLI runs as root directly)
- Desktop entry installation

## Success Metrics

- Single binary with no Python runtime required
- Feature parity with Python CLI (device mode, partition mode, all workarounds)
- Proper cleanup on all exit paths
- Clear error messages for missing dependencies or invalid inputs

## Constraints

- CLI-only for this phase
- Must run on Linux (uses Linux-specific tools: parted, wipefs, grub-install, etc.)
- Requires root privileges for disk operations
- Depends on external system tools (cannot be fully self-contained)

## Open Questions

- Should UEFI:NTFS image be bundled or downloaded at runtime? (Currently: downloaded from Rufus GitHub repo)
- How to handle NVMe device naming (`/dev/nvme0n1p1`) vs SATA (`/dev/sda1`)?
