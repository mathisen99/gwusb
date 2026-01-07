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
- [ ] Implement `ValidateSource(path)` — check exists, is file or block device
- [ ] Implement `ValidateTarget(path, mode)` — check block device, device vs partition based on mode
- [ ] Handle both `/dev/sdX` and `/dev/nvme0n1` naming patterns
- **Relates to:** AC-3

## Task 6: Busy Check
- [ ] Implement `CheckNotBusy(path)` — parse `mount` output or `/proc/mounts`
- [ ] Unmount if mounted, or error if cannot unmount
- **Relates to:** AC-3

## Task 7: Mount/Unmount Operations
- [ ] Implement `Mount(source, mountpoint, fstype, opts)` using syscall or shell
- [ ] Implement `Unmount(mountpoint)` using syscall or shell
- [ ] Create temp mountpoints using `os.MkdirTemp`
- **Relates to:** AC-7, AC-11

## Task 8: FAT32 Limit Check
- [ ] Implement `CheckFAT32Limit(mountpoint)` — walk files, return true if any >4GB
- [ ] Auto-switch filesystem to NTFS and print warning
- **Relates to:** AC-4

## Task 9: Partition Operations (Device Mode)
- [ ] Implement `Wipe(device)` — run `wipefs --all`, verify no partitions remain
- [ ] Implement `CreateMBRTable(device)` — run `parted mklabel msdos`
- [ ] Implement `CreatePartition(device, fstype)` — run `parted mkpart` with correct start/end
- [ ] Implement `RereadPartitionTable(device)` — run `blockdev --rereadpt`, sleep 3s
- **Relates to:** AC-5, AC-6

## Task 10: Format Operations
- [ ] Implement `FormatFAT32(partition)` — run `mkdosfs -F 32`
- [ ] Implement `FormatNTFS(partition, label)` — run `mkntfs --quick --label`
- **Relates to:** AC-6

## Task 11: File Copying with Progress
- [ ] Implement `CopyWithProgress(srcMount, dstMount, progressFn)`
- [ ] Walk source directory, create dirs, copy files
- [ ] Large files (>5MB) copied in chunks for progress updates
- [ ] Call `progressFn` with bytes copied, total, current file
- [ ] Print progress to stderr
- **Relates to:** AC-7

## Task 12: GRUB Bootloader
- [ ] Implement `InstallGRUB(mountpoint, device, grubCmd)` — run grub-install
- [ ] Implement `WriteGRUBConfig(mountpoint, grubPrefix)` — write grub.cfg
- [ ] Detect grub vs grub2 prefix from command name
- **Relates to:** AC-8

## Task 13: Windows 7 UEFI Workaround
- [ ] Implement `IsWindows7(srcMount)` — check cversion.ini for MinServer=7xxx
- [ ] Implement `ExtractBootloader(srcMount, dstMount)` — run `7z e -so` to extract bootmgfw.efi
- [ ] Place at `/efi/boot/bootx64.efi`
- **Relates to:** AC-9

## Task 14: UEFI:NTFS Partition (NTFS Mode)
- [ ] Implement `CreateUEFINTFSPartition(device)` — create 512KB partition at end
- [ ] Implement `InstallUEFINTFS(partition, tempDir)` — download uefi-ntfs.img, write to partition
- [ ] Handle download failure gracefully (warning, not error)
- **Relates to:** AC-6

## Task 15: Boot Flag Workaround
- [ ] Implement `SetBootFlag(device, partNum)` — run `parted set N boot on`
- **Relates to:** AC-10

## Task 16: Main Orchestration
- [ ] Wire all components together in `main.go`
- [ ] Implement full device mode flow
- [ ] Implement full partition mode flow
- [ ] Ensure cleanup runs on all exit paths
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
