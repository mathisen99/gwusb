package session

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Session struct {
	Source          string
	Target          string
	TargetDevice    string
	TargetPartition string
	Mode            string // "device" or "partition"
	Filesystem      string // "FAT" or "NTFS"
	Label           string
	SourceMount     string
	TargetMount     string
	TempDir         string
	SkipGRUB        bool
	SetBootFlag     bool
	Verbose         bool
	NoColor         bool
}

func (s *Session) Cleanup() error {
	var errs []error

	if s.SourceMount != "" {
		if err := syscall.Unmount(s.SourceMount, 0); err != nil {
			errs = append(errs, fmt.Errorf("unmount source: %w", err))
		} else {
			_ = os.Remove(s.SourceMount)
			s.SourceMount = ""
		}
	}

	if s.TargetMount != "" {
		if err := syscall.Unmount(s.TargetMount, 0); err != nil {
			errs = append(errs, fmt.Errorf("unmount target: %w", err))
		} else {
			_ = os.Remove(s.TargetMount)
			s.TargetMount = ""
		}
	}

	if s.TempDir != "" {
		if err := os.RemoveAll(s.TempDir); err != nil {
			errs = append(errs, fmt.Errorf("remove temp dir: %w", err))
		}
		s.TempDir = ""
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

func (s *Session) SetupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Fprintln(os.Stderr, "\nInterrupted, cleaning up...")
		_ = s.Cleanup()
		os.Exit(1)
	}()
}
