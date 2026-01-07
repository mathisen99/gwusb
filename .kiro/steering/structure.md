---
inclusion: always
---

# Structure: WoeUSB-ng Go Rewrite

## Repository Layout

```
woeusb-ng/
├── .kiro/
│   ├── specs/cli-rewrite/     # Spec files (requirements, design, tasks)
│   └── steering/              # This file and siblings
├── Project_we_are_replicating/ # Quarantined Python source (reference only, gitignored)
├── cmd/
│   └── woeusb/
│       └── main.go            # CLI entry point
├── internal/
│   ├── device/
│   │   ├── list.go            # USB/DVD enumeration
│   │   └── validate.go        # Source/target validation
│   ├── partition/
│   │   ├── wipe.go            # wipefs wrapper
│   │   ├── table.go           # parted mklabel/mkpart
│   │   └── format.go          # mkdosfs/mkntfs wrappers
│   ├── filesystem/
│   │   ├── mount.go           # mount/umount operations
│   │   └── copy.go            # File copying with progress
│   ├── bootloader/
│   │   ├── grub.go            # GRUB installation
│   │   └── uefi.go            # Win7 workaround, UEFI:NTFS
│   └── session/
│       └── session.go         # Session struct, cleanup logic
├── go.mod
├── go.sum
├── README.md
├── MY_UNDERSTANDING.md        # Technical analysis of original project
└── .gitignore
```

## Package Boundaries

| Package | Responsibility |
|---------|----------------|
| `cmd/woeusb` | CLI parsing, orchestration, signal handling |
| `internal/session` | Session state, cleanup logic |
| `internal/device` | Block device discovery and validation |
| `internal/partition` | Wipe, partition table, partition creation |
| `internal/filesystem` | Mount/unmount, file copy with progress |
| `internal/bootloader` | GRUB install, Win7 UEFI workaround, UEFI:NTFS |

## Naming Conventions

- Package names: lowercase, single word (`device`, `partition`, `filesystem`)
- Files: lowercase with underscores if needed (`validate.go`, `copy.go`)
- Exported functions: PascalCase (`ValidateSource`, `CopyWithProgress`)
- Internal functions: camelCase (`runCommand`, `parseOutput`)
- Errors: wrap with context using `fmt.Errorf("action failed: %w", err)`

## Adding New Features

1. Identify which package owns the feature
2. Add implementation in that package
3. Wire into `cmd/woeusb/main.go` orchestration
4. Add/update tests
5. Update README if user-facing

## Testing Layout

```
internal/
├── device/
│   ├── validate.go
│   └── validate_test.go       # Unit tests alongside source
├── filesystem/
│   ├── copy.go
│   └── copy_test.go
...
```

- Unit tests: `*_test.go` in same package
- Integration tests: require real USB device, documented in README (manual)

## Logging / Output Conventions

- Status messages: print to stderr with color (green=info, yellow=warn, red=error)
- Progress: print to stderr, overwrite line for progress bar
- `--verbose`: additional debug output
- `--no-color`: disable ANSI codes
- Errors: print to stderr, exit non-zero

## Quarantined Source Reference

The original Python project is preserved at:

```
Project_we_are_replicating/
```

This folder is:
- Gitignored (not committed to new repo)
- Reference-only for understanding original behavior
- Not to be modified or executed

## Open Questions

- Should we add a `pkg/` directory for reusable libraries? (Probably not needed for CLI-only)
- Makefile vs just `go build`?
