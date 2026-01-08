# Implementation Plan: WoeUSB-go v0.2.0 GUI

## Overview

This implementation plan adds a GUI to WoeUSB-go using the Fyne toolkit, along with distro-aware dependency checking. The implementation builds incrementally, starting with distro detection, then GUI components, and finally wiring everything together.

## Tasks

- [x] 1. Add Fyne dependency and update project structure
  - Add `fyne.io/fyne/v2` to go.mod
  - Create `internal/gui/` directory structure
  - Create `internal/distro/` directory structure
  - _Requirements: 7.1, 7.3_

- [x] 2. Implement distro detection
  - [x] 2.1 Create internal/distro/detect.go with Info struct and Detect function
    - Parse /etc/os-release file
    - Extract ID, ID_LIKE, NAME, VERSION fields
    - Determine package manager from distro ID
    - _Requirements: 6.1_
  
  - [x] 2.2 Write property test for distro detection
    - **Property 4: Distro Detection from os-release**
    - **Validates: Requirements 6.1**
  
  - [x] 2.3 Create internal/distro/packages.go with package mappings
    - Define RequiredBinaries and OptionalBinaries lists
    - Create packageMappings map for all supported distros
    - Create installCommands map for package managers
    - Implement GetPackageName and GetInstallCommand functions
    - _Requirements: 6.2, 6.3, 6.4_
  
  - [x] 2.4 Write property tests for package mapping
    - **Property 5: Package Name Mapping**
    - **Property 6: Install Command Generation**
    - **Validates: Requirements 6.2, 6.3**

- [x] 3. Checkpoint - Ensure distro detection tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Implement USB device detection
  - [x] 4.1 Create internal/gui/components/device_selector.go
    - Define USBDevice struct
    - Implement GetUSBDevices using lsblk JSON output
    - Filter for removable=true AND tran=usb
    - Exclude sata, nvme, ata transport types
    - _Requirements: 2.1, 2.2_
  
  - [x] 4.2 Write property test for USB device filtering
    - **Property 1: USB Device Filtering**
    - **Validates: Requirements 2.1, 2.2**
  
  - [x] 4.3 Implement device display formatting
    - Format device path, size, and model for display
    - _Requirements: 2.3_
  
  - [x] 4.4 Write property test for device display
    - **Property 9: Device Display Information**
    - **Validates: Requirements 2.3**

- [x] 5. Implement dependency checking integration
  - [x] 5.1 Update internal/deps/deps.go to use distro package
    - Integrate distro detection
    - Return MissingDep structs with distro-specific package names
    - _Requirements: 1.3, 6.2_
  
  - [x] 5.2 Write property test for dependency detection
    - **Property 3: Dependency Binary Detection**
    - **Validates: Requirements 1.3**

- [ ] 6. Checkpoint - Ensure device and dependency tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 7. Implement GUI components
  - [ ] 7.1 Create internal/gui/app.go
    - Initialize Fyne application
    - Implement root privilege check
    - Implement CheckDependencies method
    - _Requirements: 1.1, 1.2, 1.3_
  
  - [ ] 7.2 Write property test for root privilege detection
    - **Property 2: Root Privilege Detection**
    - **Validates: Requirements 1.2**
  
  - [ ] 7.3 Create internal/gui/components/file_browser.go
    - Implement file selection dialog with .iso filter
    - Implement ValidateISO function
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  
  - [ ] 7.4 Write property test for ISO validation
    - **Property 10: ISO File Validation**
    - **Validates: Requirements 3.3**
  
  - [ ] 7.5 Create internal/gui/components/progress_bar.go
    - Implement SetProgress and SetStatus methods
    - Implement Reset method
    - _Requirements: 5.1, 5.2_
  
  - [ ] 7.6 Write property test for progress bar updates
    - **Property 8: Progress Bar Updates**
    - **Validates: Requirements 5.1**
  
  - [ ] 7.7 Create internal/gui/components/dependency_dialog.go
    - Display missing dependencies with package names
    - Show complete install command for distro
    - Handle unknown distro fallback
    - _Requirements: 1.4, 6.2, 6.3, 6.5_

- [ ] 8. Checkpoint - Ensure GUI component tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 9. Implement main window and state management
  - [ ] 9.1 Create internal/gui/window.go
    - Layout device selector, file browser, progress bar, start button
    - Implement UpdateState for button enable/disable logic
    - _Requirements: 4.1, 4.5_
  
  - [ ] 9.2 Write property test for start button state
    - **Property 7: Start Button State**
    - **Validates: Requirements 4.1**
  
  - [ ] 9.3 Write property test for UI controls during operation
    - **Property 11: UI Controls Disabled During Operation**
    - **Validates: Requirements 4.5**
  
  - [ ] 9.4 Implement confirmation dialog and write operation
    - Show data loss warning before write
    - Connect to existing CLI write logic
    - Update progress bar during operation
    - _Requirements: 4.2, 4.3, 4.4, 4.6, 4.7_
  
  - [ ] 9.5 Implement window close handler during operation
    - Warn user about interrupting process
    - _Requirements: 5.4_

- [ ] 10. Wire GUI into main.go
  - [ ] 10.1 Add --gui flag to cmd/woeusb/main.go
    - Parse --gui flag
    - Launch GUI application when flag is set
    - _Requirements: 1.1_
  
  - [ ] 10.2 Implement GUI startup flow
    - Check dependencies on startup
    - Show dependency dialog if missing
    - Show main window if all dependencies satisfied
    - _Requirements: 1.3, 1.4, 6.6_

- [ ] 11. Final checkpoint - Full integration test
  - Ensure all tests pass, ask the user if questions arise.
  - Manual test: Launch GUI, verify USB detection, verify dependency dialog

## Notes

- All property-based tests are required
- Fyne requires OpenGL support on the system
- The GUI reuses existing internal packages for actual USB operations
- Property tests should run minimum 100 iterations each
