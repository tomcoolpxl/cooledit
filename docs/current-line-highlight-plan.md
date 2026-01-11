# Current Line Highlight Feature Plan

## Feature Overview

**Feature Name:** Current Line Highlight (also known as "Active Line Background" or "Current Line Background")

**Description:** When enabled, the editor will display a full-width background highlight on the line where the cursor is currently positioned. The highlight extends across the entire width of the editor, including the line numbers gutter (if visible), creating a visual indicator of the active editing line.

**Status:** Off by default (user opt-in)

---

## User Experience Goals

1. **Visual Clarity**: Help users quickly identify which line they're currently editing
2. **Theme Integration**: Automatically derive appropriate colors for each theme to ensure readability
3. **Non-intrusive**: Subtle enough not to distract but visible enough to be useful
4. **Consistent**: Works seamlessly with all existing features (line numbers, selection, search highlighting, syntax highlighting, word wrap)

---

## Configuration

### Config File Settings

Add to `internal/config/schema.go`:

```go
type Editor struct {
    LineNumbers        bool `toml:"line_numbers"`
    SoftWrap           bool `toml:"soft_wrap"`
    TabWidth           int  `toml:"tab_width"`
    SyntaxHighlighting bool `toml:"syntax_highlighting"`
    ShowWhitespace     bool `toml:"show_whitespace"`
    CurrentLineHighlight bool `toml:"current_line_highlight"` // NEW: default false
}
```

### Example Config (config.toml)

```toml
[editor]
line_numbers = false
soft_wrap = false
tab_width = 4
syntax_highlighting = true
show_whitespace = false
current_line_highlight = false  # Off by default
```

---

## Menu Integration

Add new menu item in the **View** menu (in `internal/ui/menubar.go`):

Position: After "Toggle Line Numbers", before "Toggle Word Wrap"

```go
{
    Label: "Current Line Highlight", 
    IsCheckable: true, 
    IsChecked: func(u *UI) bool {
        return u.currentLineHighlight
    }, 
    Action: func(u *UI) {
        u.currentLineHighlight = !u.currentLineHighlight
        u.saveConfig()
    }
}
```

---

## Theme Color Strategy

### Design Principle: Automatic Color Derivation

To ensure the feature works well with all 14+ themes without manual configuration per theme, use **automatic color calculation** based on existing theme colors:

1. **Calculate background color** by blending the editor background with a subtle overlay:
   - Take the editor's base background color
   - Apply a subtle lightening (for dark themes) or darkening (for light themes)
   - Aim for ~5-10% luminosity change to keep it subtle

2. **Algorithm approach**:
   ```
   IF theme is dark (bg luminosity < 0.5):
       current_line_bg = lighten(editor.bg, 8-12%)
   ELSE (theme is light):
       current_line_bg = darken(editor.bg, 8-12%)
   ```

3. **Preserve text colors**: Foreground colors (text, syntax highlighting) remain unchanged

### Implementation Location

Add to `internal/theme/theme.go`:

```go
type EditorColors struct {
    Fg               term.Color // Normal text foreground
    Bg               term.Color // Normal text background
    SelectionFg      term.Color // Selected text foreground
    SelectionBg      term.Color // Selected text background
    LineNumbersFg    term.Color // Line numbers foreground
    LineNumbersBg    term.Color // Line numbers background
    CursorColor      term.Color // Cursor color
    BracketMatchBg   term.Color // Background for matched bracket pair
    BracketUnmatchBg term.Color // Background for unmatched bracket
    CurrentLineBg    term.Color // NEW: Background for current line highlight
}
```

### Color Calculation Helper Function

Create utility function in `internal/theme/theme.go`:

```go
// DeriveCurrentLineBg automatically calculates an appropriate current line
// background color based on the editor's base background color.
// For dark themes, it lightens; for light themes, it darkens.
func DeriveCurrentLineBg(editorBg term.Color) term.Color {
    r, g, b := editorBg.RGB()
    
    // Calculate relative luminance (perceived brightness)
    // Formula: Y = 0.299*R + 0.587*G + 0.114*B
    luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
    isDark := luminance < 128 // Threshold at mid-gray
    
    var factor float64
    if isDark {
        factor = 1.10 // Lighten by 10% for dark themes
    } else {
        factor = 0.90 // Darken by 10% for light themes
    }
    
    // Apply adjustment with clamping
    newR := clamp(int(float64(r) * factor), 0, 255)
    newG := clamp(int(float64(g) * factor), 0, 255)
    newB := clamp(int(float64(b) * factor), 0, 255)
    
    return term.NewRGBColor(newR, newG, newB)
}

func clamp(val, min, max int) int {
    if val < min {
        return min
    }
    if val > max {
        return max
    }
    return val
}
```

### Initialization

When loading/building themes (in `internal/theme/builtin.go`):

```go
// For each built-in theme, after defining Editor.Bg:
theme.Editor.CurrentLineBg = DeriveCurrentLineBg(theme.Editor.Bg)
```

For custom user themes (in `internal/config/theme.go`):

```go
func ConvertThemeSpec(name string, spec ThemeSpec) *theme.Theme {
    editorBg := parseColorField(spec.Editor.Bg)
    
    // Derive current line bg if not explicitly provided
    currentLineBg := parseColorField(spec.Editor.CurrentLineBg)
    if currentLineBg == term.ColorDefault { // User didn't specify
        currentLineBg = theme.DeriveCurrentLineBg(editorBg)
    }
    
    return &theme.Theme{
        // ... existing fields
        Editor: theme.EditorColors{
            // ... existing fields
            CurrentLineBg: currentLineBg,
        },
    }
}
```

### Allow User Override (Optional)

Users can optionally override the auto-calculated color in their config:

```toml
[themes.myCustomTheme.editor]
fg = "#ffffff"
bg = "#1e1e1e"
current_line_bg = "#2a2a2a"  # Optional: override auto-calculated color
```

---

## Rendering Implementation

### Changes to `internal/ui/render.go`

Modify the viewport rendering functions to apply the current line background:

#### In `drawViewportNoWrap`:

```go
func (u *UI) drawViewportNoWrap(...) {
    cursorLine := u.editor.Cursor().Line
    
    for sy := 0; sy < vpRect.H; sy++ {
        docY := vp.TopLine + sy
        isCurrentLine := (docY == cursorLine) && u.currentLineHighlight
        
        // Draw Gutter (line numbers)
        gutterStyle := u.getLineNumberStyle()
        if isCurrentLine {
            // Apply current line background to gutter
            gutterStyle.Background = u.theme.Editor.CurrentLineBg
        }
        
        if u.showLineNumbers {
            // ... existing gutter rendering code ...
            // Use gutterStyle which now has CurrentLineBg if isCurrentLine
        }
        
        // Draw line content
        editorStyle := u.getEditorStyle()
        if isCurrentLine {
            // Override background for current line
            editorStyle.Background = u.theme.Editor.CurrentLineBg
        }
        
        // ... rest of line rendering ...
        // When setting cell styles, use editorStyle which now has CurrentLineBg
        
        // IMPORTANT: Fill remaining space to end of viewport width
        if isCurrentLine {
            // Fill to the right edge even if line is shorter
            for sx := lastDrawnCol; sx < availW; sx++ {
                u.screen.SetCell(drawX+sx, vpRect.Y+sy, ' ', editorStyle)
            }
        }
    }
}
```

#### In `drawViewportWrapped`:

Apply similar logic for wrapped line rendering:

```go
func (u *UI) drawViewportWrapped(...) {
    cursorLine := u.editor.Cursor().Line
    
    // When rendering wrapped lines, check if the source line (docY) 
    // matches the cursor line
    for each wrapped segment {
        isCurrentLine := (docY == cursorLine) && u.currentLineHighlight
        
        if isCurrentLine {
            baseStyle.Background = u.theme.Editor.CurrentLineBg
            gutterStyle.Background = u.theme.Editor.CurrentLineBg
        }
        
        // ... rendering logic ...
    }
}
```

### Rendering Priority (Z-order)

The current line highlight should be **below** other highlighting:

1. **Base layer**: Current line background (if enabled)
2. **Middle layer**: Search matches, bracket matching
3. **Top layer**: Selection highlighting

This means:
- If text is selected, selection colors take precedence
- If text matches a search, search highlight takes precedence
- Current line background is the lowest priority visual indicator

Implementation: Apply current line background first, then check for selection/search and override if needed.

---

## UI Component Integration

### State Management (`internal/ui/ui.go`)

Add field to `UI` struct:

```go
type UI struct {
    // ... existing fields ...
    showLineNumbers      bool
    softWrap             bool
    showWhitespace       bool
    currentLineHighlight bool // NEW: toggled via menu
    // ... rest of fields ...
}
```

Initialize from config:

```go
func NewUI(editor *core.Editor, screen term.Screen, config *config.Config) *UI {
    // ... existing initialization ...
    
    ui.currentLineHighlight = config.Editor.CurrentLineHighlight
    
    // ... rest ...
}
```

### Config Save Function

Update `saveConfig()` in `internal/ui/ui.go`:

```go
func (u *UI) saveConfig() {
    // ... existing saves ...
    u.config.Editor.CurrentLineHighlight = u.currentLineHighlight
    // ... write to file ...
}
```

---

## Edge Cases & Considerations

### 1. **Word Wrap Mode**
- When word wrap is enabled, multiple screen lines represent one logical line
- **Decision**: Highlight all wrapped segments of the current line
- Implementation: Check `docY` (document line) instead of `screenY` when determining current line

### 2. **Selection Active**
- When text is selected, selection background takes priority
- Current line highlight visible only on non-selected portions
- Implementation: Already handled by existing rendering order (selection check after base style)

### 3. **Search Matches**
- Search highlights should remain visible on the current line
- Current line background should not obscure search matches
- Implementation: Apply search styles after current line background

### 4. **Syntax Highlighting**
- Foreground colors (syntax highlighting) remain unchanged
- Only background changes for current line
- Implementation: Preserve foreground colors, only modify background in style

### 5. **Line Numbers Gutter**
- Current line highlight extends into the gutter area
- Line number text remains readable (foreground unchanged)
- Implementation: Apply CurrentLineBg to gutterStyle.Background

### 6. **Horizontal Scrolling**
- Highlight extends to viewport edge, not just to text content
- Fill empty space on the right with the highlight background
- Implementation: Explicit fill loop after text rendering

### 7. **Empty Lines**
- Current line highlight visible even on empty lines
- Implementation: Fill entire viewport width with highlight color

### 8. **Multiple Windows/Splits** (if implemented in future)
- Each viewport should highlight its own current line
- Current implementation is single-viewport, so not applicable yet

---

## Testing Strategy

### Manual Testing Checklist

1. **Basic Functionality**
   - [ ] Toggle on/off via menu item
   - [ ] Setting persists across restarts
   - [ ] Highlight follows cursor movement (arrow keys)
   - [ ] Highlight updates when clicking with mouse (if supported)

2. **Theme Compatibility**
   - [ ] Test with all 14 built-in themes
   - [ ] Verify readability in light themes (default, light, solarized-light, gruvbox-light)
   - [ ] Verify readability in dark themes (dark, monokai, solarized-dark, gruvbox-dark, dracula, nord)
   - [ ] Verify readability in retro themes (dos, ibm-green, ibm-amber, cyberpunk)
   - [ ] Check contrast ratio is sufficient (WCAG guidelines suggest 4.5:1 for text)

3. **Feature Interactions**
   - [ ] Works with line numbers on
   - [ ] Works with line numbers off
   - [ ] Works with word wrap on (all wrapped segments highlighted)
   - [ ] Works with word wrap off
   - [ ] Works with syntax highlighting on
   - [ ] Works with syntax highlighting off
   - [ ] Works with whitespace visualization on
   - [ ] Selection highlighting takes precedence
   - [ ] Search highlighting remains visible
   - [ ] Bracket matching remains visible

4. **Performance**
   - [ ] No noticeable lag with large files (10,000+ lines)
   - [ ] Smooth scrolling with highlight enabled
   - [ ] No flickering during cursor movement

5. **Edge Cases**
   - [ ] Empty file (one empty line)
   - [ ] Single line file
   - [ ] Very long lines (horizontal scrolling)
   - [ ] End of file (cursor on last line)
   - [ ] Small terminal window

### Unit Testing

Create `internal/ui/current_line_test.go`:

```go
func TestCurrentLineHighlight(t *testing.T) {
    // Test toggle functionality
    // Test config persistence
    // Test that cursor line is correctly identified
}
```

### Integration Testing

Add to existing `internal/ui/render_test.go`:

```go
func TestRenderCurrentLineHighlight(t *testing.T) {
    // Verify correct background color applied
    // Verify full width rendering
    // Verify gutter inclusion
}
```

---

## Documentation Updates

### 1. README.md

Add to Features section:

```markdown
- Current line highlighting (toggle in View menu, off by default)
```

Add to Configuration section:

```toml
[editor]
current_line_highlight = false  # Highlight the line where cursor is located
```

### 2. Help Screen (F1)

Add to View shortcuts section:

```
View menu: Current Line Highlight - Toggle current line background
```

### 3. CHANGELOG.md

Add to next release:

```markdown
## [Unreleased]

### Added
- Current line highlighting: optional full-width background highlight for the active line
  - Off by default, toggle via View menu
  - Automatically adapts to all themes with appropriate contrast
  - Works with line numbers, word wrap, and all other features
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Estimated: 2-3 hours)
1. Add config field to schema
2. Add UI state field
3. Add theme color field
4. Implement color derivation algorithm
5. Initialize colors for all built-in themes

### Phase 2: Rendering (Estimated: 3-4 hours)
1. Modify `drawViewportNoWrap` to apply current line background
2. Modify `drawViewportWrapped` to apply current line background
3. Ensure full-width rendering (including gutter and trailing space)
4. Test rendering priority with selection and search

### Phase 3: UI Integration (Estimated: 1-2 hours)
1. Add menu item to View menu
2. Implement toggle action
3. Wire up config save/load
4. Test menu interaction

### Phase 4: Testing & Polish (Estimated: 2-3 hours)
1. Manual testing with all themes
2. Test all feature interactions
3. Adjust color derivation algorithm if needed
4. Performance testing with large files

### Phase 5: Documentation (Estimated: 1 hour)
1. Update README.md
2. Update help screen
3. Update CHANGELOG.md

**Total Estimated Time: 9-13 hours**

---

## Alternative Approaches Considered

### 1. **Fixed Color per Theme**
- **Pros**: Precise control, could match specific theme aesthetics
- **Cons**: Requires manual configuration for 14+ themes, hard to maintain, custom themes would need explicit color
- **Decision**: Rejected in favor of automatic derivation

### 2. **User-Configurable Only**
- **Pros**: Maximum flexibility
- **Cons**: Poor default experience, requires user to pick color for each theme
- **Decision**: Rejected; auto-derivation with optional override is better UX

### 3. **Highlight Cursor Line Number Only**
- **Pros**: Simpler implementation
- **Cons**: Much less visible, doesn't achieve the goal of clear line identification
- **Decision**: Rejected; full-width highlight provides better UX

### 4. **Highlight Text Content Only (not full width)**
- **Pros**: Less visually "heavy"
- **Cons**: Looks inconsistent, especially on short lines; doesn't include gutter
- **Decision**: Rejected; full-width is standard in modern editors (VS Code, Sublime, etc.)

---

## Success Criteria

1. ✅ Feature is off by default
2. ✅ Can be toggled via View menu
3. ✅ Setting persists across sessions
4. ✅ Works correctly with all 14 built-in themes
5. ✅ Highlight is readable but subtle (doesn't overwhelm)
6. ✅ Compatible with all existing features (no visual conflicts)
7. ✅ No performance degradation
8. ✅ Documented in README and help screen

---

## Future Enhancements (Out of Scope)

1. **Configurable highlight intensity**: Let users adjust the lightening/darkening percentage
2. **Different highlight styles**: Border instead of background, or combined
3. **Multiple cursor support**: Highlight multiple lines if multi-cursor is added
4. **Inactive pane dimming**: If split-pane feature is added, dim non-active panes

---

## References

- Similar features in other editors:
  - VS Code: `editor.renderLineHighlight` setting
  - Sublime Text: `highlight_line` setting
  - Vim: `cursorline` option
  - Emacs: `hl-line-mode`

All major text editors include this feature, confirming its value to users.

---

**Document Version:** 1.0  
**Date:** January 11, 2026  
**Author:** GitHub Copilot  
**Status:** Ready for Implementation
