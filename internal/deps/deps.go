package deps

import (
	"fmt"
	"os/exec"
	"strings"
)

type Dependencies struct {
	Wipefs      string
	Parted      string
	Lsblk       string
	Blockdev    string
	Mount       string
	Umount      string
	SevenZip    string
	MkFat       string
	MkNTFS      string
	GrubCmd     string
	WimlibSplit string // wimlib-imagex for splitting WIM files
}

func CheckDependencies() (*Dependencies, error) {
	deps := &Dependencies{}
	var missing []string

	// Required tools
	required := map[string]*string{
		"wipefs":   &deps.Wipefs,
		"parted":   &deps.Parted,
		"lsblk":    &deps.Lsblk,
		"blockdev": &deps.Blockdev,
		"mount":    &deps.Mount,
		"umount":   &deps.Umount,
		"7z":       &deps.SevenZip,
	}

	for cmd, field := range required {
		if path, err := exec.LookPath(cmd); err != nil {
			missing = append(missing, cmd)
		} else {
			*field = path
		}
	}

	// Find mkdosfs/mkfs.vfat/mkfs.fat (return first found)
	fatCmds := []string{"mkdosfs", "mkfs.vfat", "mkfs.fat"}
	for _, cmd := range fatCmds {
		if path, err := exec.LookPath(cmd); err == nil {
			deps.MkFat = path
			break
		}
	}
	if deps.MkFat == "" {
		missing = append(missing, "mkdosfs/mkfs.vfat/mkfs.fat")
	}

	// Find mkntfs (optional - only needed if user forces NTFS)
	if path, err := exec.LookPath("mkntfs"); err == nil {
		deps.MkNTFS = path
	}

	// Find grub-install or grub2-install (optional for UEFI-only systems)
	grubCmds := []string{"grub-install", "grub2-install"}
	for _, cmd := range grubCmds {
		if path, err := exec.LookPath(cmd); err == nil {
			deps.GrubCmd = path
			break
		}
	}

	// Find wimlib-imagex for splitting WIM files (required for Win10/11)
	if path, err := exec.LookPath("wimlib-imagex"); err != nil {
		missing = append(missing, "wimlib-imagex (install wimlib package)")
	} else {
		deps.WimlibSplit = path
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required dependencies: %s", strings.Join(missing, ", "))
	}

	return deps, nil
}
