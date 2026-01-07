---
inclusion: always
---

# Tech: WoeUSB-ng Go Rewrite

## Primary Language

Go (latest stable)

## Dependencies Policy

Prefer stdlib. Minimize external dependencies.

**Allowed:**
- Standard library (`flag`, `os/exec`, `syscall`, `filepath`, `net/http`, `io`, `fmt`)
- No external packages required for core functionality

**Avoid unless necessary:**
- Heavy frameworks (cobra, viper) — use stdlib `flag` instead
- GUI toolkits (deferred to future phase)

## Tooling Conventions

**Build:**
```bash
go build -o woeusb ./cmd/woeusb
```

**Test:**
```bash
go test ./...
```

**Lint:**
TODO: Decide on linter (golangci-lint recommended)

**Format:**
```bash
go fmt ./...
```

## Runtime / Deployment

- Single static binary
- Runs on Linux only
- Requires root privileges (`os.Getuid() == 0`)
- No daemon mode; runs to completion and exits

## External System Tools (Required)

The binary shells out to these Linux tools:

| Tool | Package | Purpose |
|------|---------|---------|
| `wipefs` | util-linux | Clear partition signatures |
| `parted` | parted | Create partition table and partitions |
| `lsblk` | util-linux | List block devices |
| `blockdev` | util-linux | Re-read partition table |
| `mount` | util-linux | Mount filesystems |
| `umount` | util-linux | Unmount filesystems |
| `mkdosfs` | dosfstools | Format FAT32 (or `mkfs.vfat`, `mkfs.fat`) |
| `mkntfs` | ntfs-3g | Format NTFS |
| `grub-install` | grub2 | Install GRUB bootloader (or `grub2-install`) |
| `7z` | p7zip | Extract files from WIM archives |

## Network Access

- UEFI:NTFS image download: `https://github.com/pbatard/rufus/raw/master/res/uefi/uefi-ntfs.img`
- Only used in NTFS mode
- Failure is non-fatal (prints warning, continues)

## Security / Secret Handling

- No secrets or credentials involved
- Tool operates on local block devices only
- Must validate target is a block device before destructive operations
- Warn (don't fail) if not running as root

## Error Handling Conventions

- All functions return `error`
- Wrap errors with context: `fmt.Errorf("failed to wipe device: %w", err)`
- Use `defer` for cleanup
- Handle SIGINT/SIGTERM for graceful cleanup

## Open Questions

- Linter configuration (golangci-lint settings)
- CI/CD pipeline (GitHub Actions?)
- Release binary distribution method
