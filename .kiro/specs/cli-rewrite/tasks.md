# Tasks: CLI Rewrite (Go)

## Task 1: Project Scaffold
- [x] Create `go.mod` with module name `github.com/user/woeusb-go`
- [x] Create directory structure: `cmd/woeusb/`, `internal/{device,partition,filesystem,bootloader,session}/`
- [x] Create stub `main.go` that prints version
- **Relates to:** All requirements (foundation)

## Task 2: CLI Argument Parsing
- [x] Implement flag parsing: `--device`, `--partition`, `--target-filesystem`, `--label`, `--workaround-bios-boot-flag`, `--workaround-skip-grub`, `--verbose`, `--no-color`, `--version`
- [x] Validate mutually exclusive `--device` / `--partition`
- [x] Print usage on `--help` or invalid args
- **Relates to:** AC-1

## Task 3: Session and Cleanup
- [x] Implement `Session` struct in `internal/session/`
- [x] Implement `Cleanup()` method (unmount, remove temp dirs)
- [x] Add signal handler for SIGINT/SIGTERM that calls cleanup
- **Relates to:** AC-11

## Task 4: Dependency Checker
- [x] Implement `CheckDependencies()` that verifies: wipefs, parted, lsblk, blockdev, mount, umount, 7z
- [x] Find mkdosfs/mkfs.vfat/mkfs.fat (return first found)
- [x] Find mkntfs
- [x] Find grub-install or grub2-install
- [x] Return clear error listing missing dependencies
- **Relates to:** AC-2

## Task 5: Source/Target Validation
- [x] Implement `ValidateSource(path)` — check exists, is file or block device
- [x] Implement `ValidateTarget(path, mode)` — check block device, device vs partition based on mode
- [x] Handle both `/dev/sdX` and `/dev/nvme0n1` naming patterns
- **Relates to:** AC-3

## Task 6: Busy Check
- [x] Implement `CheckNotBusy(path)` — parse `mount` output or `/proc/mounts`
- [x] Unmount if mounted, or error if cannot unmount
- **Relates to:** AC-3

## Task 7: Mount/Unmount Operations
- [x] Implement `Mount(source, mountpoint, fstype, opts)` using syscall or shell
- [x] Implement `Unmount(mountpoint)` using syscall or shell
- [x] Create temp mountpoints using `os.MkdirTemp`
- **Relates to:** AC-7, AC-11

## Task 8: FAT32 Limit Check
- [x] Implement `CheckFAT32Limit(mountpoint)` — walk files, return true if any >4GB
- [x] Auto-switch filesystem to NTFS and print warning
- **Relates to:** AC-4

## Task 9: Partition Operations (Device Mode)
- [x] Implement `Wipe(device)` — run `wipefs --all`, verify no partitions remain
- [x] Implement `CreateMBRTable(device)` — run `parted mklabel msdos`
- [x] Implement `CreatePartition(device, fstype)` — run `parted mkpart` with correct start/end
- [x] Implement `RereadPartitionTable(device)` — run `blockdev --rereadpt`, sleep 3s
- **Relates to:** AC-5, AC-6

## Task 10: Format Operations
- [x] Implement `FormatFAT32(partition)` — run `mkdosfs -F 32`
- [x] Implement `FormatNTFS(partition, label)` — run `mkntfs --quick --label`
- **Relates to:** AC-6

## Task 11: File Copying with Progress
- [x] Implement `CopyWithProgress(srcMount, dstMount, progressFn)`
- [x] Walk source directory, create dirs, copy files
- [x] Large files (>5MB) copied in chunks for progress updates
- [x] Call `progressFn` with bytes copied, total, current file
- [x] Print progress to stderr
- **Relates to:** AC-7

## Task 12: GRUB Bootloader
- [x] Implement `InstallGRUB(mountpoint, device, grubCmd)` — run grub-install
- [x] Implement `WriteGRUBConfig(mountpoint, grubPrefix)` — write grub.cfg
- [x] Detect grub vs grub2 prefix from command name
- **Relates to:** AC-8

## Task 13: Windows 7 UEFI Workaround
- [x] Implement `IsWindows7(srcMount)` — check cversion.ini for MinServer=7xxx
- [x] Implement `ExtractBootloader(srcMount, dstMount)` — run `7z e -so` to extract bootmgfw.efi
- [x] Place at `/efi/boot/bootx64.efi`
- **Relates to:** AC-9

## Task 14: UEFI:NTFS Partition (NTFS Mode)
- [x] Implement `CreateUEFINTFSPartition(device)` — create 512KB partition at end
- [x] Implement `InstallUEFINTFS(partition, tempDir)` — download uefi-ntfs.img, write to partition
- [x] Handle download failure gracefully (warning, not error)
- **Relates to:** AC-6

## Task 15: Boot Flag Workaround
- [x] Implement `SetBootFlag(device, partNum)` — run `parted set N boot on`
- **Relates to:** AC-10

## Task 16: Main Orchestration
- [x] Wire all components together in `main.go`
- [x] Implement full device mode flow
- [x] Implement full partition mode flow
- [x] Ensure cleanup runs on all exit paths
- **Relates to:** All ACs

## Task 17: Testing
- [ ] Write unit tests for validation functions
- [ ] Write unit tests for FAT32 limit check
- [ ] Document manual integration test procedure (requires real USB device)
- **Relates to:** AC-12

## Task 18: Documentation
- [ ] Update README.md with build and usage instructions
- [ ] Document required system dependencies
- [ ] Add examples for common use cases
- **Relates to:** All requirements
