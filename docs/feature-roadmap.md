# cooledit Feature Roadmap

Prioritized list of nano-inspired features to implement.

---

## Implementation Status

| Feature | Status | Notes |
|---------|--------|-------|
| Smart Home Key | ✅ Complete | Phase 1 |
| Indent/Unindent Block | ✅ Complete | Phase 1 |
| Trim Trailing Whitespace | ✅ Complete | Phase 2 |
| Comment/Uncomment | ✅ Complete | Phase 2 |
| Position Log | ✅ Complete | Phase 3 |
| Scrollbar/Indicator | ✅ Complete | Phase 3 |
| Verbatim Character Input | ✅ Complete | Phase 4 |
| File Browser | ⏳ Pending | Phase 5 |
| Formatter Integration | ⏳ Pending | Phase 6 |
| Linter Integration | ⏳ Pending | Phase 7 |

---

## Priority 1: High Impact, Low Complexity

### 1. Smart Home Key
**Complexity:** Low  
**Impact:** High (quality of life improvement)

Home key jumps to first non-whitespace character, then to column 0 on second press.
- First press: Jump to first non-whitespace on line
- Second press: Jump to column 0
- Cycle between the two positions

**Implementation:** Modify `handleKey` for Home key, track last position.

---

### 2. Indent/Unindent Block
**Complexity:** Low  
**Impact:** High (essential for code editing)

Indent or unindent selected lines as a block.
- Tab with selection: Indent all selected lines
- Shift+Tab with selection: Unindent all selected lines
- Works on single line if no selection

**Implementation:** Modify Tab handling when selection is active.

---

### 3. Trim Trailing Whitespace
**Complexity:** Low  
**Impact:** Medium (code hygiene)

Options:
- Toggle in View menu
- Auto-trim on save (configurable)
- Visual indicator for trailing whitespace (already have whitespace display)

**Implementation:** Add config option, hook into save operation.

---

### 4. Comment/Uncomment
**Complexity:** Medium  
**Impact:** High (essential for code editing)

Single-keystroke line commenting per syntax.
- Ctrl+/ or similar to toggle comment on current line or selection
- Comment style defined per language (e.g., `//` for Go, `#` for Python)
- Uncomment if already commented

**Implementation:** 
- Add comment patterns to syntax definitions
- Create toggle logic that detects existing comments
- Support both line comments (`//`) and block comments (`/* */`)

---

## Priority 2: Medium Impact, Medium Complexity

### 5. Position Log
**Complexity:** Medium  
**Impact:** Medium (convenience feature)

Remember cursor position in recently edited files, restore on reopen.
- Store in JSON/TOML file alongside config
- Track last 50-100 files
- Store: filepath, line, column, timestamp
- On open: check if position exists, restore if file unchanged

**Implementation:**
- New `positionlog` package or extend autosave
- Hook into file open and close operations
- Add config toggle

---

### 6. Scrollbar/Indicator
**Complexity:** Medium  
**Impact:** Low-Medium (visual feedback)

Visual scrollbar on right edge showing viewport position.
- Single character column on far right
- Shows viewport position relative to file length
- Different character for viewport area vs rest

**Implementation:**
- Reserve rightmost column in layout
- Calculate viewport percentage
- Render in `drawViewport`

---

### 7. Verbatim Character Input
**Complexity:** Medium  
**Impact:** Low (niche use case)

Insert special characters by code point.
- Ctrl+Shift+U  for Unicode hex entry
- Ctrl+Shift+D for decimal entry
- Show input mode in status bar

**Implementation:**
- New input mode (ModeVerbatim)
- Accumulate hex/decimal digits
- Convert to rune and insert

---

## Priority 3: High Complexity

### 8. File Browser
**Complexity:** High  
**Impact:** High (discoverability)

Graphical file browser for open/save operations.
- Invoke from Ctrl+O (open) and Save As
- Navigate directories with arrow keys
- Show files and folders with icons/indicators
- Filter by extension (optional)
- Preview file info

**Implementation:**
- New UI mode (ModeFileBrowser)
- Directory listing with scrolling
- Parent directory navigation
- File selection and return

---

### 9. Formatter Integration
**Complexity:** High  
**Impact:** Medium (developer productivity)

Run full-buffer formatters defined per syntax.
- Configurable formatter command per language
- Ctrl+Shift+F or menu item to format
- Replace buffer with formatted output
- Undoable as single operation

**Implementation:**
- Add formatter config to syntax definitions
- Execute external command with buffer as stdin
- Replace buffer content, preserve cursor if possible

---

### 10. Linter Integration
**Complexity:** High  
**Impact:** Medium (developer productivity)

Run syntax checkers, navigate through errors.
- Configurable linter command per language
- Parse output (file:line:col: message format)
- Show errors in status bar or dedicated panel
- Navigate between errors with F8/Shift+F8

**Implementation:**
- Add linter config to syntax definitions
- Execute and parse linter output
- Store error list, highlight error lines
- Navigation commands

---

## Implementation Order Recommendation

| Phase | Features | Estimated Effort |
|-------|----------|------------------|
| Phase 1 | Smart Home, Indent/Unindent Block | 1-2 days |
| Phase 2 | Comment/Uncomment, Trim Whitespace | 2-3 days |
| Phase 3 | Position Log, Scrollbar | 2-3 days |
| Phase 4 | Verbatim Input | 1-2 days |
| Phase 5 | File Browser | 3-5 days |
| Phase 6 | Formatter Integration | 2-3 days |
| Phase 7 | Linter Integration | 3-5 days |

---

## Notes

- All features should be toggleable via config and/or menu
- All features should have sensible defaults
- Consider keyboard shortcut conflicts with existing bindings
- Test on Windows, Linux, and macOS
