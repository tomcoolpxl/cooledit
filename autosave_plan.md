# Autosave Implementation Plan for cooledit

## Overview

This document provides a complete implementation plan for autosave functionality in cooledit. The design follows cooledit's principles: single-file editor, explicit save semantics, cross-platform safety, and minimal complexity.

## Goals

1. **Prevent data loss** from crashes, power failures, or accidental closures
2. **Non-intrusive** - autosave happens silently without user interaction
3. **Explicit save semantics preserved** - autosave files are backups, not replacements
4. **Cross-platform** - works on Windows, Linux, and macOS
5. **Recovery workflow** - clear, simple recovery on startup

---

## 1. Autosave Strategy and Triggers

### 1.1 When to Autosave

Autosave triggers on an **idle timer with debouncing**:

- **Idle threshold**: 2 seconds after last edit (configurable)
- **Minimum interval**: 30 seconds between autosaves (configurable)
- **Only when modified**: Skip autosave if `editor.Modified()` is false

### 1.2 What Triggers Reset of Idle Timer

- Any buffer modification (insert, delete, paste, undo, redo)
- Timer resets on each keystroke that modifies the buffer

### 1.3 Implementation Approach

```go
// internal/autosave/autosave.go

type AutosaveManager struct {
    enabled       bool
    idleTimeout   time.Duration  // default: 2s
    minInterval   time.Duration  // default: 30s
    lastAutosave  time.Time
    idleTimer     *time.Timer
    mu            sync.Mutex

    // Callback to get current buffer state
    getState      func() AutosaveState
    // Callback when autosave completes
    onAutosave    func(err error)
}

type AutosaveState struct {
    Lines    [][]rune
    Path     string      // empty for unnamed buffers
    EOL      string
    Encoding string
    Modified bool
}
```

### 1.4 Timer Logic

```go
func (a *AutosaveManager) NotifyEdit() {
    a.mu.Lock()
    defer a.mu.Unlock()

    if !a.enabled {
        return
    }

    // Reset idle timer
    if a.idleTimer != nil {
        a.idleTimer.Stop()
    }

    a.idleTimer = time.AfterFunc(a.idleTimeout, func() {
        a.tryAutosave()
    })
}

func (a *AutosaveManager) tryAutosave() {
    a.mu.Lock()
    defer a.mu.Unlock()

    // Check minimum interval
    if time.Since(a.lastAutosave) < a.minInterval {
        return
    }

    state := a.getState()
    if !state.Modified {
        return
    }

    err := a.writeAutosaveFile(state)
    if err == nil {
        a.lastAutosave = time.Now()
    }

    if a.onAutosave != nil {
        a.onAutosave(err)
    }
}
```

---

## 2. Storage Locations and File Format

### 2.1 Autosave Directory

All autosave files are stored in a dedicated directory:

| Platform | Location |
|----------|----------|
| Windows | `%APPDATA%\cooledit\autosave\` |
| Linux | `~/.local/share/cooledit/autosave/` |
| macOS | `~/Library/Application Support/cooledit/autosave/` |

### 2.2 File Naming Scheme

**For named files:**
```
{hash-of-original-path}.autosave
```

**For unnamed buffers:**
```
unnamed-{timestamp}.autosave
```

**Metadata file (alongside each autosave):**
```
{hash-of-original-path}.meta
```

### 2.3 Hash Function

Use a stable hash for path → filename mapping:

```go
func pathToFilename(path string) string {
    // Normalize path separators and case (Windows)
    normalized := filepath.Clean(path)
    if runtime.GOOS == "windows" {
        normalized = strings.ToLower(normalized)
    }

    // FNV-1a hash for speed and simplicity
    h := fnv.New64a()
    h.Write([]byte(normalized))
    return fmt.Sprintf("%016x.autosave", h.Sum64())
}
```

### 2.4 Autosave File Format

The autosave file is **plain text** matching the original file's encoding and EOL format. This ensures:
- Easy manual recovery if needed
- Same format as final save
- No special parsing required

### 2.5 Metadata File Format

TOML format for consistency with config:

```toml
# .meta file
original_path = "C:\\Users\\tom\\project\\main.go"
encoding = "UTF-8"
eol = "\n"
timestamp = 2026-01-11T14:30:00Z
cooledit_version = "0.2.0"
```

### 2.6 Implementation

```go
// internal/autosave/storage.go

func AutosaveDir() (string, error) {
    var base string
    switch runtime.GOOS {
    case "windows":
        base = os.Getenv("APPDATA")
    case "darwin":
        home, _ := os.UserHomeDir()
        base = filepath.Join(home, "Library", "Application Support")
    default:
        home, _ := os.UserHomeDir()
        base = filepath.Join(home, ".local", "share")
    }

    dir := filepath.Join(base, "cooledit", "autosave")
    return dir, os.MkdirAll(dir, 0700)
}

type AutosaveMeta struct {
    OriginalPath    string    `toml:"original_path"`
    Encoding        string    `toml:"encoding"`
    EOL             string    `toml:"eol"`
    Timestamp       time.Time `toml:"timestamp"`
    CoolEditVersion string    `toml:"cooledit_version"`
}
```

---

## 3. Recovery Detection and Startup Flow

### 3.1 Startup Check

On application startup, **before** loading any file:

```go
// internal/autosave/recovery.go

type RecoveryCandidate struct {
    AutosavePath string
    Meta         AutosaveMeta
    OriginalExists bool
    OriginalNewer  bool  // Original file was modified after autosave
}

func FindRecoveryCandidate(targetPath string) (*RecoveryCandidate, error) {
    dir, err := AutosaveDir()
    if err != nil {
        return nil, nil // No autosave dir = no recovery
    }

    if targetPath == "" {
        // Check for unnamed buffer autosaves
        return findUnnamedRecovery(dir)
    }

    // Check for specific file
    filename := pathToFilename(targetPath)
    autosavePath := filepath.Join(dir, filename)
    metaPath := filepath.Join(dir, strings.TrimSuffix(filename, ".autosave") + ".meta")

    if _, err := os.Stat(autosavePath); os.IsNotExist(err) {
        return nil, nil // No autosave exists
    }

    // Read metadata
    var meta AutosaveMeta
    if _, err := toml.DecodeFile(metaPath, &meta); err != nil {
        // Metadata missing/corrupt - still offer recovery
        meta.OriginalPath = targetPath
    }

    candidate := &RecoveryCandidate{
        AutosavePath: autosavePath,
        Meta:         meta,
    }

    // Check if original exists and compare timestamps
    if info, err := os.Stat(targetPath); err == nil {
        candidate.OriginalExists = true
        candidate.OriginalNewer = info.ModTime().After(meta.Timestamp)
    }

    return candidate, nil
}
```

### 3.2 Startup Flow Diagram

```
                    ┌─────────────────┐
                    │  App Startup    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │ Check for       │
                    │ recovery file   │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
      No recovery                  Recovery found
              │                             │
              ▼                             ▼
     ┌────────────────┐            ┌────────────────┐
     │ Normal startup │            │ Show recovery  │
     │ (load file)    │            │ prompt         │
     └────────────────┘            └───────┬────────┘
                                           │
                          ┌────────────────┼────────────────┐
                          │                │                │
                   [R]ecover         [O]riginal       [D]iscard
                          │                │                │
                          ▼                ▼                ▼
                   ┌──────────┐     ┌──────────┐     ┌──────────┐
                   │ Load     │     │ Load     │     │ Delete   │
                   │ autosave │     │ original │     │ autosave │
                   │ (marked  │     │ file     │     │ Load     │
                   │ modified)│     │          │     │ original │
                   └──────────┘     └──────────┘     └──────────┘
```

---

## 4. Recovery UX and Lifecycle Rules

### 4.1 Recovery Prompt

When a recovery candidate is found, show a modal prompt:

```
┌─────────────────────────────────────────────────────────────┐
│  Recover unsaved changes?                                   │
│                                                             │
│  An autosave backup was found for:                          │
│  C:\Users\tom\project\main.go                               │
│                                                             │
│  Autosave from: 2026-01-11 14:30:00                         │
│  Original file: 2026-01-11 12:00:00                         │
│                                                             │
│  [R]ecover backup  [O]pen original  [D]iscard backup        │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Recovery Actions

| Action | Behavior |
|--------|----------|
| **Recover (R)** | Load autosave content, mark as modified, keep autosave file |
| **Open Original (O)** | Load original file, keep autosave for future |
| **Discard (D)** | Delete autosave+meta files, load original file |

### 4.3 Lifecycle Rules

#### When to Create Autosave
- After idle timeout when buffer is modified
- Never for files that haven't been modified

#### When to Update Autosave
- On idle timeout if buffer modified since last autosave
- After minimum interval has passed

#### When to Delete Autosave
- On explicit save (Ctrl+S) after successful write
- On "Discard" during recovery prompt
- On clean exit with no unsaved changes

#### When to Keep Autosave
- On crash (obviously)
- On "Recover" action (until user saves)
- On "Open Original" action (user might want it later)
- On exit with unsaved changes + "Don't Save" choice

### 4.4 Exit Flow Integration

```
                    ┌─────────────────┐
                    │  User quits     │
                    │  (Ctrl+Q)       │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
      Not modified                   Modified
              │                             │
              ▼                             ▼
     ┌────────────────┐            ┌────────────────┐
     │ Delete autosave│            │ "Save changes?"│
     │ Clean exit     │            │ Y/N/Cancel     │
     └────────────────┘            └───────┬────────┘
                                           │
                    ┌──────────────────────┼───────────────────┐
                    │                      │                   │
                 [Y]es                  [N]o              [C]ancel
                    │                      │                   │
                    ▼                      ▼                   ▼
             ┌──────────┐           ┌──────────┐        ┌──────────┐
             │ Save     │           │ Keep     │        │ Return   │
             │ Delete   │           │ autosave │        │ to       │
             │ autosave │           │ Exit     │        │ editing  │
             │ Exit     │           └──────────┘        └──────────┘
             └──────────┘
```

---

## 5. Configuration Options

### 5.1 Config Schema Addition

```go
// internal/config/schema.go

type Autosave struct {
    Enabled      bool `toml:"enabled"`       // default: true
    IdleTimeout  int  `toml:"idle_timeout"`  // seconds, default: 2
    MinInterval  int  `toml:"min_interval"`  // seconds, default: 30
}

// Add to Config struct:
type Config struct {
    // ... existing fields ...
    Autosave Autosave `toml:"autosave"`
}
```

### 5.2 Config File Example

```toml
[autosave]
enabled = true
idle_timeout = 2      # seconds of idle before autosave
min_interval = 30     # minimum seconds between autosaves
```

### 5.3 Menu Integration

Add to View menu:
- "Autosave" toggle (checkmark when enabled)

---

## 6. Implementation Phases

### Phase 1: Core Infrastructure
**Files to create:**
- `internal/autosave/autosave.go` - AutosaveManager
- `internal/autosave/storage.go` - File operations

**Tasks:**
1. Create autosave directory structure
2. Implement path-to-filename hashing
3. Implement autosave file writing (reuse `fileio.Save` logic)
4. Implement metadata file writing
5. Add autosave config to schema

### Phase 2: Timer Integration
**Files to modify:**
- `internal/ui/ui.go` - Hook into edit operations
- `internal/core/editor.go` - Add NotifyEdit hook

**Tasks:**
1. Add AutosaveManager to UI struct
2. Call NotifyEdit on buffer modifications
3. Implement idle timer with debouncing
4. Wire up autosave callback

### Phase 3: Recovery System
**Files to create:**
- `internal/autosave/recovery.go` - Recovery detection

**Files to modify:**
- `internal/app/app.go` - Startup flow
- `internal/ui/ui.go` - Recovery prompt UI

**Tasks:**
1. Implement FindRecoveryCandidate
2. Add recovery prompt mode
3. Handle recovery actions (R/O/D)
4. Load recovered content as modified buffer

### Phase 4: Lifecycle Management
**Files to modify:**
- `internal/ui/ui.go` - Save and quit handlers

**Tasks:**
1. Delete autosave on successful save
2. Handle autosave on quit flow
3. Clean exit removes autosave if not modified

### Phase 5: Polish and Testing
**Files to create:**
- `internal/autosave/autosave_test.go`
- `internal/autosave/recovery_test.go`

**Tasks:**
1. Unit tests for timer logic
2. Unit tests for file operations
3. Unit tests for recovery detection
4. Integration tests for full recovery flow
5. Add View menu toggle
6. Update F1 help screen

---

## 7. File Summary

### New Files
| File | Purpose |
|------|---------|
| `internal/autosave/autosave.go` | AutosaveManager, timer logic |
| `internal/autosave/storage.go` | Directory, file I/O, metadata |
| `internal/autosave/recovery.go` | Recovery detection and candidates |
| `internal/autosave/autosave_test.go` | Tests |
| `internal/autosave/recovery_test.go` | Recovery tests |

### Modified Files
| File | Changes |
|------|---------|
| `internal/config/schema.go` | Add Autosave config struct |
| `internal/config/config.go` | Add defaults for autosave |
| `internal/ui/ui.go` | Add AutosaveManager, recovery mode |
| `internal/app/app.go` | Startup recovery check |
| `internal/core/editor.go` | Add NotifyEdit callback hook |

---

## 8. Error Handling

### 8.1 Autosave Failures
- Log errors silently (don't interrupt user)
- Retry on next idle cycle
- Optional: show status bar indicator if autosave fails repeatedly

### 8.2 Recovery Failures
- If autosave file is corrupt, offer to discard
- If metadata missing, still offer recovery (use defaults)
- If directory inaccessible, skip recovery and continue

### 8.3 Atomic Writes
Reuse existing atomic write pattern from `fileio.Save`:
```go
// Write to .tmp, then rename
tmp := path + ".tmp"
os.WriteFile(tmp, data, 0600)
os.Rename(tmp, path)
```

---

## 9. Security Considerations

1. **Permissions**: Autosave files created with 0600 (user read/write only)
2. **Directory permissions**: Autosave directory created with 0700
3. **No sensitive data in filenames**: Use hash of path, not actual path
4. **Cleanup on discard**: Properly delete both autosave and metadata files

---

## 10. Future Enhancements (Not in Initial Scope)

- Multiple unnamed buffer recovery
- Autosave version history (keep last N autosaves)
- Cloud sync of autosave directory
- Configurable autosave location

---

## Summary

This plan provides a complete, production-ready autosave implementation that:

1. Uses an idle-based trigger with debouncing to minimize disk I/O
2. Stores autosave files in a dedicated platform-appropriate directory
3. Uses hash-based filenames for safe cross-platform paths
4. Provides a clear 3-option recovery prompt on startup
5. Integrates cleanly with existing save and quit workflows
6. Follows cooledit's design principles of simplicity and explicit user control
