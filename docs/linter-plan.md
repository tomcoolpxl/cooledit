# Linter Integration Plan for cooledit

## Executive Summary

This document outlines the design for integrating external linters into cooledit, enabling developers to run syntax checkers and navigate through errors directly within the editor. This feature follows the established patterns from the formatter integration while introducing new capabilities for error display and navigation.

---

## 1. Goals and Requirements

### Primary Goals
1. Run configurable linter commands per language
2. Parse linter output (file:line:col:severity:message format and variations)
3. Display diagnostics in a non-intrusive way (gutter markers, status bar, error panel)
4. Navigate between errors with keyboard shortcuts (F8/Shift+F8)
5. Integrate with the existing theme system for consistent styling
6. Make the feature undoable/reversible (run lint again after editing)

### Non-Goals (Initial Version)
- Real-time linting as you type (too resource intensive for terminal editor)
- LSP (Language Server Protocol) integration (complex, out of scope for Phase 7)
- Multi-file project-wide linting (single-file focus to start)

---

## 2. Architecture Overview

### New Package: `internal/linter`

```
internal/linter/
    linter.go        # Core types (Config, Diagnostic, Severity)
    executor.go      # Run external linter commands (reuse formatter.Execute)
    parser.go        # Parse linter output formats
    builtins.go      # Built-in linter configurations (like formatter)
```

### Key Types

```go
// Severity represents the diagnostic severity level
type Severity int

const (
    SeverityError Severity = iota
    SeverityWarning
    SeverityInfo
    SeverityHint
)

// Diagnostic represents a single linter diagnostic
type Diagnostic struct {
    Line     int      // 1-based line number
    Col      int      // 1-based column number (optional, 0 = unknown)
    EndLine  int      // End line for range (optional, 0 = same as Line)
    EndCol   int      // End column for range (optional)
    Severity Severity
    Message  string
    Source   string   // Linter name (e.g., "golangci-lint", "pylint")
    Code     string   // Error code (e.g., "E501", "unused-variable")
}

// Config holds the configuration for a linter
type Config struct {
    Command      string   `toml:"command"`
    Args         []string `toml:"args"`
    OutputFormat string   `toml:"output_format"` // "default", "json", "regexp"
    Pattern      string   `toml:"pattern"`       // Custom regex for parsing (optional)
}

// LintResult holds the result of running a linter
type LintResult struct {
    Diagnostics []Diagnostic
    Stderr      string
    ExitCode    int
}
```

### Integration with Existing Packages

| Package | Changes |
|---------|---------|
| `internal/config` | Add `Linters map[string]LinterConfigSpec` to schema |
| `internal/ui` | Add `diagnostics []Diagnostic`, navigation state, rendering |
| `internal/theme` | Add diagnostic gutter/underline colors to ThemeSpec |
| `internal/syntax/languages.go` | Could optionally store default linter per language |

---

## 3. Output Format Parsing

Linters output diagnostics in various formats. We need parsers for common patterns:

### 3.1 Standard Format (file:line:col: message)
```
main.go:10:5: undefined: foo
```

### 3.2 GCC/Clang Format (file:line:col: severity: message)
```
main.c:10:5: error: expected ';' after expression
main.c:15:3: warning: unused variable 'x'
```

### 3.3 Pylint/Flake8 Format (file:line:col: code message)
```
main.py:10:5: E501 line too long (85 > 79 characters)
main.py:15:1: W0612 unused variable 'x'
```

### 3.4 JSON Format (many modern linters)
```json
[{"file": "main.go", "line": 10, "column": 5, "severity": "error", "message": "..."}]
```

### 3.5 Custom Regex
Allow users to specify custom patterns for less common linters.

---

## 4. Built-in Linter Configurations

Following the formatter pattern, we'll provide built-in defaults for **16 languages**:

### Programming Languages

| Language | Linter | Command | Format |
|----------|--------|---------|--------|
| Go | golangci-lint | `golangci-lint run --out-format=line-number` | line-number |
| Python | ruff | `ruff check --output-format=concise` | default |
| JavaScript | eslint | `eslint --format=unix` | unix |
| TypeScript | eslint | `eslint --format=unix` | unix |
| Rust | cargo | `cargo check --message-format=short` | cargo |
| Java | checkstyle | `checkstyle -c /google_checks.xml` | checkstyle |
| C# | dotnet | `dotnet format --verify-no-changes --verbosity diagnostic` | dotnet |
| C | clang-tidy | `clang-tidy` | clang |
| C++ | clang-tidy | `clang-tidy` | clang |
| Bash | shellcheck | `shellcheck -f gcc` | gcc |

### DevOps / Infrastructure

| Language | Linter | Command | Format |
|----------|--------|---------|--------|
| Terraform | tflint | `tflint --format=compact` | compact |
| Terraform | terraform | `terraform validate -json` | json |
| Ansible | ansible-lint | `ansible-lint -p` | parseable |
| Dockerfile | hadolint | `hadolint --format=tty` | hadolint |
| Docker Compose | docker compose | `docker compose config -q` | stderr |
| Kubernetes | kubeval | `kubeval --output=tap` | tap |
| Kubernetes | kubectl | `kubectl apply --dry-run=client -f` | kubectl |

### Markup / Config

| Language | Linter | Command | Format |
|----------|--------|---------|--------|
| YAML | yamllint | `yamllint -f parsable` | parsable |
| JSON | jsonlint | `jsonlint --quiet` | default |
| Markdown | markdownlint | `markdownlint --json` | json |

---

## 5. User Interface Design

### 5.1 Gutter Markers (Line Number Column)

When line numbers are enabled, show diagnostic markers in the gutter:

```
  1 │ package main
  2 │ 
⚠ 3 │ import "fmt"
  4 │ 
✗ 5 │ func main() {
  6 │     foo := bar  // error here
  7 │ }
```

- `✗` (or `E`) for errors - red background
- `⚠` (or `W`) for warnings - yellow background
- `ℹ` (or `I`) for info/hints - blue background

If line numbers are off, use a narrow gutter (1-2 chars) just for markers.

### 5.2 Current Line Diagnostic Display

When cursor is on a line with diagnostics, show in status bar area:

```
main.go:5:3: error: undefined: foo [E001]
```

Or a dedicated diagnostic line above the status bar (similar to search mode).

### 5.3 Diagnostic Panel (Optional Future Enhancement)

A toggleable panel showing all diagnostics:

```
┌─ Diagnostics (3 errors, 2 warnings) ─────────────────────────────────────┐
│ ✗ main.go:5:3  error: undefined: foo                                     │
│ ✗ main.go:12:1 error: missing return statement                           │  
│ ⚠ main.go:8:5  warning: unused variable 'x'                              │
│ ⚠ main.go:15:3 warning: shadowed variable 'err'                          │
│ ℹ main.go:20:1 info: consider using switch statement                     │
└──────────────────────────────────────────────────────────────────────────┘
```

### 5.4 In-Line Indicators (Text Highlighting)

Underline or highlight the actual error span when known:

```go
    foo := bar
           ~~~  // underlined in red
```

---

## 6. Keyboard Shortcuts

| Shortcut | Action | Notes |
|----------|--------|-------|
| `Ctrl+Shift+L` | Run linter | Primary trigger |
| `F8` | Go to next diagnostic | Wraps around |
| `Shift+F8` | Go to previous diagnostic | Wraps around |
| `Ctrl+.` | Quick fix (future) | Show available fixes |
| `Esc` | Clear diagnostics display | When in diagnostic navigation |

### Menu Integration

**Edit Menu:**
- Run Linter (Ctrl+Shift+L) ← new item

**View Menu:**
- Show Diagnostics (checkmark toggle) ← new item

---

## 7. Configuration Schema

### TOML Configuration

```toml
[editor]
# ... existing settings ...
show_diagnostics = true          # Show diagnostic markers (default: true)

[linters.go]
command = "golangci-lint"
args = ["run", "--out-format=line-number"]

[linters.python]
command = "ruff"                 # Override default (pylint → ruff)
args = ["check", "--output-format=concise"]

[linters.custom]
command = "my-linter"
args = ["--some-option"]
output_format = "regexp"
pattern = "^(?P<file>[^:]+):(?P<line>\\d+):(?P<message>.+)$"
```

---

## 8. Implementation Phases

### Phase 7.1: Core Infrastructure (2-3 days)
- [ ] Create `internal/linter` package structure
- [ ] Define types: `Severity`, `Diagnostic`, `Config`, `LintResult`
- [ ] Implement `Execute()` (reuse formatter executor pattern)
- [ ] Implement output parsers:
  - [ ] Default (file:line:col: message)
  - [ ] GCC format (file:line:col: severity: message)
  - [ ] JSON format (generic)
  - [ ] Hadolint format
  - [ ] Checkstyle format
- [ ] Add built-in linter configurations for 16 languages
- [ ] Unit tests for parsers

### Phase 7.2: UI Integration - Basic (2-3 days)
- [ ] Add `diagnostics []linter.Diagnostic` to UI state
- [ ] Add `currentDiagnosticIndex int` for navigation
- [ ] Add `Linters` config to schema.go
- [ ] Wire up `Ctrl+Shift+L` keybinding to run linter
- [ ] Implement temp file handling for unsaved buffers
- [ ] Show lint result count in status bar ("3 errors, 2 warnings")
- [ ] Show current line diagnostic in status bar when cursor moves
- [ ] Add theme colors for diagnostics (error, warning, info)

### Phase 7.3: Gutter Markers (1-2 days)
- [ ] Modify gutter rendering to show diagnostic markers
- [ ] Use symbols: `✗` (error), `⚠` (warning), `ℹ` (info)
- [ ] Fallback ASCII: `E`, `W`, `I` for limited terminals
- [ ] Respect theme colors for marker foreground
- [ ] Handle multiple diagnostics per line (show highest severity)

### Phase 7.4: Navigation (1 day)
- [ ] Implement F8 for next diagnostic
- [ ] Implement Shift+F8 for previous diagnostic
- [ ] Add wrap-around behavior
- [ ] Jump cursor to diagnostic location (line, col)
- [ ] Update status bar to show navigated diagnostic

### Phase 7.5: Clear on Edit (1 day)
- [ ] Hook into editor modification events
- [ ] Clear all diagnostics when buffer is modified
- [ ] Clear diagnostic count from status bar
- [ ] Reset navigation index

### Phase 7.6: Polish & Edge Cases (1-2 days)
- [ ] Handle empty output (no errors) - show "✓ No issues"
- [ ] Handle linter not found - show helpful message
- [ ] Handle linter timeout (10s default)
- [ ] Menu integration (Edit → Run Linter)
- [ ] Update help screen with F8/Shift+F8
- [ ] Update CLAUDE.md documentation

---

## 9. Theme Integration

Add to `ThemeSpec` in config/schema.go:

```go
type DiagnosticThemeSpec struct {
    ErrorFg       string `toml:"error_fg"`
    ErrorBg       string `toml:"error_bg"`
    WarningFg     string `toml:"warning_fg"`
    WarningBg     string `toml:"warning_bg"`
    InfoFg        string `toml:"info_fg"`
    InfoBg        string `toml:"info_bg"`
    HintFg        string `toml:"hint_fg"`
    HintBg        string `toml:"hint_bg"`
    GutterErrorBg string `toml:"gutter_error_bg"`
    GutterWarnBg  string `toml:"gutter_warn_bg"`
}
```

Default colors in built-in themes:
- Error: Red (#ff0000 or terminal red)
- Warning: Yellow (#ffff00 or terminal yellow)
- Info: Blue (#0000ff or terminal blue)
- Hint: Cyan (#00ffff or terminal cyan)

---

## 10. Error Handling & Edge Cases

| Scenario | Behavior |
|----------|----------|
| Linter not installed | Show message: "Linter 'X' not found. Install it or configure a different linter." |
| Linter timeout (>10s) | Cancel, show message: "Linter timed out" |
| Linter crashes | Show stderr in message: "Linter failed: {first line of stderr}" |
| No linter configured | Show message: "No linter configured for {language}" |
| Zero diagnostics | Show success message: "✓ No issues found" |
| Parse failure | Show raw output in message, log warning |
| File modified after lint | **Clear all diagnostics immediately** |
| Unsaved file | **Auto-save to temp file, lint that, delete temp** |
| New file (no path) | Create temp file with appropriate extension based on language |

---

## 11. Design Decisions (Finalized)

### UI/UX Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Trigger mechanism** | Manual only (`Ctrl+Shift+L`) | Keep it simple, no surprises |
| **Visual style** | Gutter markers only | Cleaner, less visual noise, nano-like |
| **Stale diagnostics** | Clear immediately on edit | Safe, no confusion with outdated markers |
| **Diagnostic display** | Status bar (replace center help) | Consistent with other modes |
| **Keybinding** | `Ctrl+Shift+L` | Memorable: **L**int |

### Technical Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| **Unsaved files** | Auto-save to temp file | Seamless UX, no interruption |
| **Multi-diagnostic/line** | Show first, cycle with F8 | Simple, discoverable |
| **Persistence** | No, always fresh | Simpler implementation |
| **Diagnostic panel** | Not in Phase 1 | Keep minimal: gutter + status bar |

---

## 12. Relationship to Existing Code

### Reuse from Formatter
- `formatter.Execute()` → can be refactored to shared `executor` package
- Config schema pattern (map[string]Config with command/args)
- Built-in defaults pattern with user override

### Similar to Search
- Navigation with F8/Shift+F8 (like F3/Shift+F3)
- State tracking (current diagnostic index)
- Mode for diagnostic messages (like ModeMessage)

### Extends Syntax
- Could add default linter to `Language` struct
- Uses same language detection for choosing linter

---

## 13. Success Criteria

The linter integration is complete when:

1. ✅ User can run `Ctrl+Shift+L` to lint current file
2. ✅ Diagnostics appear as gutter markers
3. ✅ F8/Shift+F8 navigates between diagnostics
4. ✅ Current line diagnostic shows in status bar
5. ✅ Built-in linters work for at least 5 languages
6. ✅ User can configure custom linters via TOML
7. ✅ Theme colors are applied consistently
8. ✅ Error cases are handled gracefully
9. ✅ Feature is documented in CLAUDE.md and help screen

---

## 14. Estimated Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| 7.1 Core | 2-3 days | linter package, parsers, 16 language builtins |
| 7.2 Basic UI | 2-3 days | keybinding, temp file, status bar display |
| 7.3 Gutter | 1-2 days | visual markers in line number column |
| 7.4 Navigation | 1 day | F8/Shift+F8 navigation |
| 7.5 Clear on Edit | 1 day | diagnostics cleared when buffer modified |
| 7.6 Polish | 1-2 days | edge cases, menu, docs |
| **Total** | **8-12 days** | Full linter integration |

---

## 15. Future Enhancements (Post-Phase 7)

These features are explicitly deferred to keep Phase 7 focused:

- [ ] Diagnostic panel (toggleable split view with all diagnostics)
- [ ] In-line underlines for error spans  
- [ ] Auto-lint on save (configurable)
- [ ] Auto-lint on file open (configurable)
- [ ] Custom regex patterns for linter output
- [ ] LSP integration
- [ ] Quick-fix suggestions (Ctrl+.)
- [ ] Project-wide linting (all files)

---

*Document Version: 1.1*  
*Author: GitHub Copilot*  
*Date: January 11, 2026*  
*Status: Design Finalized - Ready for Implementation*
