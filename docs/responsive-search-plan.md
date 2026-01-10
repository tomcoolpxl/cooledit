# Responsive Search Implementation Plan

**Created:** January 10, 2026  
**Author:** Tom Cool  
**Status:** ✅ **IMPLEMENTED** (January 10, 2026)

> **Implementation Note:** This plan has been fully implemented with the unified search mode (ModeSearch).
> Key changes from this plan:
> - Single unified search mode implemented (no separate incremental/find-replace modes)
> - Keyboard shortcuts changed: F3/Shift+F3 for navigation, Ctrl+R for replace, Ctrl+H for replace-all
> - All letters now type into search box (no single-letter commands)
> - See SEARCH_UX_FIX.md for final implementation details

## Executive Summary

This document outlines the plan to implement robust responsive (incremental) search functionality in CoolEdit. The current implementation has issues where keystrokes can leak through to the editor buffer, and the search interface lacks real-time feedback and proper state management.

**UX Update (Jan 10, 2026):** Based on comprehensive UX review, this plan now uses a **single unified search mode** instead of separate incremental and find/replace modes, matches industry-standard keyboard shortcuts (Alt+C, Alt+W), and adds critical safety features like replace-all confirmation and search history.

## Current State Analysis

### Existing Search Implementation

**Files Involved:**
- `internal/core/search.go` - Core search logic
- `internal/core/editor.go` - Editor command handling (CmdFind, CmdFindNext, CmdFindPrev)
- `internal/ui/prompt.go` - Prompt handling including PromptFind
- `internal/ui/ui.go` - UI event handling and mode management
- `internal/ui/render.go` - Status bar and UI rendering

**Current Search Flow:**
1. User presses `Ctrl+F` → enters `ModePrompt` with `PromptFind`
2. User types search term in prompt
3. User presses Enter → executes `CmdFind`, switches to `ModeFindReplace` if found
4. In `ModeFindReplace`: keys like `n`, `p`, `r`, `a`, `q` navigate/act on results

### Identified Issues

#### 1. **Key Leakage Problem** (Critical)
**Symptom:** Letters like `p`, `n`, `a`, `q` sometimes insert into the editor buffer instead of being handled as search commands.

**Root Cause Analysis:**
- The `handleFindReplaceKey()` function in `ui.go` (line 637) doesn't have complete key coverage
- Some key events may fall through to normal editor input handling
- Race conditions or mode transition issues may allow keys to leak through
- The function returns `true` at the end, claiming all keys are handled, but edge cases exist

**Evidence:**
- Test file `findreplace_test.go` specifically tests for this: `TestFindReplaceModePreventsCharacterInsertion`
- The test verifies that `n`, `p`, `q` don't insert characters

#### 2. **No Real-Time Search Feedback**
Currently, search happens only after pressing Enter. Modern editors show results as you type:
- No incremental highlighting
- No match count display
- No immediate feedback if no matches exist

#### 3. **Status Bar Issues**
When in `ModeFindReplace`:
- Shows command shortcuts but not dynamic search state
- No display of:
  - Number of matches found (e.g., "3 of 15 matches")
  - Current match position
  - Case sensitivity status
  - Search term being used
- Shortcuts displayed but no indication of case sensitivity toggle key

#### 4. **Case Sensitivity Management**
- No case sensitivity toggle functionality
- No persistence of case sensitivity preference across session
- No visual indication of current case sensitivity mode

#### 5. **Error Handling**
- "Not found" messages exit search mode immediately
- No way to correct typo without restarting search
- Poor user experience when no matches

#### 6. **Search State Management**
- Search and replace strings remembered in `lastFindTerm` and `lastReplaceTerm`
- But no structured search state object to track:
  - Case sensitivity preference
  - Match positions/count
  - Current match index
  - Search options (whole word, regex, etc. - future)

## Design Goals

### Core Requirements

1. **Responsive/Incremental Search (Unified Mode)**
   - Search executes on every keystroke while typing
   - Results highlight immediately
   - Editor jumps to first match automatically
   - User can navigate matches while still typing (no mode switch)
   - Enter moves to next match (or exits), Esc exits to normal mode

2. **Robust Key Handling**
   - Zero key leakage to editor buffer in search modes
   - Clear state machine for mode transitions
   - Comprehensive event handling in search modes

3. **Rich Visual Feedback**
   - Match count display (e.g., "Match 1 of 5")
   - All matches highlighted in viewport
   - Different highlight colors for current match vs other matches
   - Case sensitivity indicator
   - Error state when no matches

4. **Status Bar Enhancements**
   - Search mode: Show search term, match count, current match index
   - Clear case sensitivity indicator (text-based, not cryptic symbols)
   - Display toggle keys (e.g., "Alt+C: Case")
   - Keep right-side status info (line, col, encoding)
   - Responsive design: degrade gracefully on narrow terminals
   - Show available actions (N:Next, R:Replace, etc.)

5. **Case Sensitivity**
   - Toggle with `Alt+C` (matches VS Code, avoids terminal conflicts)
   - Visual indicator: "Match Case" or "Ignore Case" (clear text labels)
   - Session persistence (remembered while editor runs)
   - Default: case-insensitive (or smart-case based on query)
   - Optional: Smart case auto-detection (lowercase query = insensitive)

6. **Error Handling**
   - Show "No matches" message without exiting search mode
   - Allow user to continue editing search term
   - Visual indication of error state (red background in status bar?)

7. **State Consistency**
   - Search and replace strings persist across searches in session
   - Case sensitivity persists
   - Clear state transitions between modes
   - Guardrails prevent invalid state combinations

8. **Safety Features**
   - Replace-all requires confirmation
   - Replace-all is undoable as single operation
   - Clear escape hatches (Esc always exits)
   - Performance feedback for slow searches

9. **Discoverability**
   - F1 help overlay shows all search shortcuts
   - Status bar displays available actions
   - Pre-fill search from current selection
   - Search history with up/down arrows

## Proposed Architecture

### New Data Structures

#### SearchSession (in `internal/core/search.go`)
```go
type SearchSession struct {
    Query          string      // Current search term
    CaseSensitive  bool        // Case sensitivity toggle
    Matches        []Match     // All match positions in current buffer
    CurrentIndex   int         // Index of currently selected match (-1 if none)
    LastReplaceStr string      // Last replacement string
}

type Match struct {
    Line   int
    Col    int
    Length int
}
```

#### Enhanced SearchState (in `internal/core/editor.go`)
```go
type SearchState struct {
    LastQuery      string
    CaseSensitive  bool         // Session-level preference
    Session        *SearchSession // Active search session (nil when not searching)
}
```

### Mode Transitions

```
                 Ctrl+F (or selection + Ctrl+F)
ModeNormal ──────────────────────────────────────→ ModeSearch (unified)
                                                          │
    ↑                                                     │
    │                                                     │ Type: search in real-time
    │                                                     │ N/P/F3: navigate matches
    │                                                     │ R: prompt for replace
    │                                                     │ A: prompt for replace all
    │ Esc/Q                                               │ Alt+C: toggle case
    │                                                     │ Alt+W: toggle whole word
    │                                                     │ Up/Down: search history
    └─────────────────────────────────────────────────────┘
```

#### Unified ModeSearch
- Displays search prompt at bottom with live query
- Real-time search as user types
- All navigation works while typing (no need to "commit")
- Status bar shows: "Find: <query> | Match 1 of 5 | Match Case | Alt+C Alt+W | F3 Esc"
- Enter → moves to next match (same as N or F3)
- Esc/Q → exits, returns to ModeNormal
- R → prompts for replacement text, returns to search
- A → prompts for replace-all with confirmation

### Search Algorithm Enhancement

**Current:** Simple `strings.Index()` forward/backward search

**Enhanced:**
```go
// New function in search.go
func FindAllMatches(lines [][]rune, query string, caseSensitive bool) []Match

// Enhanced search with case sensitivity
func SearchWithOptions(lines [][]rune, query string, startLine, startCol int, 
                       dir Direction, caseSensitive bool) (int, int, bool)
```

### UI Changes

#### Status Bar Rendering (in `render.go:drawStatusBar()`)

**ModeSearch (Responsive Design):**

**Wide terminal (>100 cols):**
```
Find: query_text | Match 3 of 15 | Match Case | Alt+C:Toggle Alt+W:Word | F3:Next R:Replace Esc:Exit | Ln 42, Col 8 UTF-8 LF
```

**Medium terminal (80 cols):**
```
Find: query_text | 3 of 15 | Match Case | Alt+C Alt+W | F3 R Esc | Ln 42, Col 8
```

**Narrow terminal (60 cols):**
```
Find: query | 3/15 | Case:Y | F3 Esc | Ln 42 Col 8
```

**Error state (no matches - red/yellow background):**
```
Find: notfound | No matches | Alt+C:Case | Esc:Exit | Ln 42, Col 8
```

**Searching state (debouncing):**
```
Find: long_query_being_typed | Searching... | Esc:Cancel
```

**Right side:** Always preserve `Ln X, Col Y  UTF-8 LF` (priority 1)

#### New Rendering Functions

```go
// in render.go
func (u *UI) drawIncrementalSearchStatus()
func (u *UI) drawFindReplaceStatus()
func (u *UI) highlightSearchMatches(matches []Match, currentMatch int)
```

#### Color Scheme Additions (in `theme/theme.go`)
```go
type EditorColors struct {
    // ... existing fields ...
    SearchMatchBg       Color  // Background for non-current matches
    SearchMatchFg       Color  // Foreground for non-current matches
    SearchCurrentBg     Color  // Background for current match
    SearchCurrentFg     Color  // Foreground for current match
    SearchErrorBg       Color  // Status bar background when no matches
}
```

### Event Handling

#### Handler: `handleSearchKey()` (in `ui.go`) - Unified Mode

```go
func (u *UI) handleSearchKey(e term.KeyEvent) bool {
    switch e.Key {
    case term.KeyEscape:
        // Exit search, return to normal mode
        u.exitSearch()
        return true
        
    case term.KeyEnter:
        // Move to next match (same as 'n' or F3)
        u.nextSearchMatch()
        return true
        
    case term.KeyBackspace:
        // Remove character from search query
        if len(u.searchQuery) > 0 {
            u.searchQuery = u.searchQuery[:len(u.searchQuery)-1]
            u.performSearch()
        } else {
            // Backspace on empty search = exit (intuitive)
            u.exitSearch()
        }
        return true
        
    case term.KeyUp, term.KeyDown:
        // Navigate search HISTORY, not matches
        if e.Key == term.KeyUp {
            u.searchHistoryPrev()
        } else {
            u.searchHistoryNext()
        }
        return true
        
    case term.KeyF3:
        // Next/previous match
        if e.Modifiers == term.ModShift {
            u.prevSearchMatch()
        } else {
            u.nextSearchMatch()
        }
        return true
        
    case term.KeyF1:
        // Show search help overlay
        u.showSearchHelp()
        return true
        
    case term.KeyRune:
        if e.Modifiers == term.ModAlt {
            switch e.Rune {
            case 'c', 'C':
                // Toggle case sensitivity (Alt+C matches VS Code)
                u.editor.ToggleCaseSensitivity()
                u.performSearch()
                return true
            case 'w', 'W':
                // Toggle whole word matching
                u.editor.ToggleWholeWord()
                u.performSearch()
                return true
            }
        } else if e.Modifiers == term.ModCtrl {
            switch e.Rune {
            case 'v', 'V':
                // Paste into search (helpful for long patterns)
                if text, err := u.editor.clipboard.Get(); err == nil {
                    u.searchQuery = append(u.searchQuery, []rune(text)...)
                    u.performSearch()
                }
                return true
            }
        } else {
            // Handle regular runes
            switch e.Rune {
            case 'n', 'N':
                // Next match
                u.nextSearchMatch()
                return true
            case 'p', 'P':
                // Previous match
                u.prevSearchMatch()
                return true
            case 'r', 'R':
                // Replace current match
                u.enterReplacePrompt()
                return true
            case 'a', 'A':
                // Replace all with confirmation
                u.enterReplaceAllPrompt()
                return true
            case 'q', 'Q':
                // Exit search
                u.exitSearch()
                return true
            default:
                // Add character to search query
                u.searchQuery = append(u.searchQuery, e.Rune)
                u.performSearch()
                return true
            }
        }
    }
    
    // Critical: Return true for ALL keys to prevent leakage
    return true
}
```

#### Additional Search Features

**Search History Management:**
```go
type SearchHistory struct {
    queries []string
    index   int
    maxSize int
}

func (u *UI) searchHistoryPrev() {
    // Move backwards in search history
    if u.searchHistory.index > 0 {
        u.searchHistory.index--
        u.searchQuery = []rune(u.searchHistory.queries[u.searchHistory.index])
        u.performSearch()
    }
}
```

**Pre-fill from Selection:**
```go
func (u *UI) enterSearch() {
    // If text is selected, use it as initial search query
    if u.editor.HasSelection() {
        sl, sc, el, ec := u.editor.GetSelectionRange()
        selectedText := u.editor.RangeText(sl, sc, el, ec)
        u.searchQuery = []rune(selectedText)
    } else if u.lastSearchQuery != "" {
        // Otherwise use last search
        u.searchQuery = []rune(u.lastSearchQuery)
    }
    u.mode = ModeSearch
    u.performSearch()
}
```

**Performance Throttling:**
```go
func (u *UI) performSearch() {
    // Debounce: wait 150ms after last keystroke
    if u.searchDebounceTimer != nil {
        u.searchDebounceTimer.Stop()
    }
    
    u.searchDebounceTimer = time.AfterFunc(150*time.Millisecond, func() {
        u.doSearchWithFeedback()
    })
    
    // Show "Searching..." immediately if query is long
    if len(u.searchQuery) > 20 {
        u.showSearchingIndicator()
    }
}
```

## Implementation Plan

### Phase 1: Infrastructure & Bug Fixes (Week 1)

**Priority: Critical Bug Fixes**

#### Task 1.1: Fix Key Leakage
- [ ] Implement unified `handleSearchKey()` with comprehensive coverage
- [ ] Ensure function returns `true` for ALL possible key events
- [ ] Add default case to consume unexpected keys
- [ ] Run existing test: `TestFindReplaceModePreventsCharacterInsertion`
- [ ] Add additional test cases for edge cases
- [ ] Remove old two-mode approach (ModeIncrementalSearch + ModeFindReplace)

#### Task 1.2: Add Case-Sensitive Search
- [ ] Extend `Search()` function signature with `caseSensitive bool` parameter
- [ ] Implement case-sensitive and case-insensitive string matching
- [ ] Update `SearchForward()` and `SearchBackward()`
- [ ] Add `SearchState.CaseSensitive` field
- [ ] Write unit tests in `search_test.go`

#### Task 1.3: Implement FindAllMatches
- [ ] Create `FindAllMatches()` function in `search.go`
- [ ] Returns all match positions in buffer
- [ ] Respects case sensitivity setting
- [ ] Optimize for large files (limit matches? highlight viewport only?)
- [ ] Write unit tests

### Phase 2: Unified Search Mode & Core Features (Week 1-2)

#### Task 2.1: Create SearchSession
- [ ] Add `SearchSession` struct to `search.go`
- [ ] Add session management methods
- [ ] Integrate into `Editor` struct
- [ ] Add lifecycle methods (start, update, end session)
- [ ] Add whole word matching support

#### Task 2.2: Implement Unified ModeSearch
- [ ] Define single `ModeSearch` constant (replaces two-mode approach)
- [ ] Add UI state fields for search (query, history, options)
- [ ] Implement `enterSearch()` with selection pre-fill
- [ ] Implement `exitSearch()` with cleanup
- [ ] Implement `performSearch()` with debouncing

#### Task 2.3: Search History
- [ ] Add `SearchHistory` struct
- [ ] Store last 20 searches in memory
- [ ] Implement up/down arrow navigation
- [ ] Add to search session on exit
- [ ] Optional: Persist to config file

#### Task 2.4: Implement Real-Time Search
- [ ] Wire up unified `handleSearchKey()`
- [ ] On each keystroke, debounce and call `performSearch()`
- [ ] Update editor selection to current match
- [ ] Ensure cursor follows current match
- [ ] Add "Searching..." indicator for slow searches

### Phase 3: Visual Feedback (Week 2)

#### Task 3.1: Theme Support for Search Highlighting
- [ ] Add search highlight colors to `theme.go`
- [ ] Update all built-in themes with sensible defaults
- [ ] Add color fields: `SearchMatchBg`, `SearchMatchFg`, `SearchCurrentBg`, `SearchCurrentFg`, `SearchErrorBg`

#### Task 3.2: Implement Match Highlighting
- [ ] Create `highlightSearchMatches()` in `render.go`
- [ ] Modify `drawViewportNoWrap()` to check for search matches
- [ ] Highlight current match differently from other matches
- [ ] Modify `drawViewportWrapped()` similarly
- [ ] Ensure selection highlighting doesn't conflict with search highlighting

#### Task 3.3: Enhanced Status Bar
- [ ] Implement `drawIncrementalSearchStatus()`
- [ ] Implement `drawFindReplaceStatus()`
- [ ] Show match count and current match index
- [ ] Show case sensitivity indicator
- [ ] Show relevant shortcuts
- [ ] Preserve right-side info (line, col, encoding, EOL)

### Phase 4: Search Options Toggles (Week 2)

#### Task 4.1: Case Sensitivity Toggle
- [ ] Add `Alt+C` key binding in search mode (matches VS Code)
- [ ] Add `ToggleCaseSensitivity()` to `Editor`
- [ ] Re-run search when toggled
- [ ] Session persistence (within editor lifetime)
- [ ] Optional: Smart case detection (lowercase = insensitive)

#### Task 4.2: Whole Word Toggle
- [ ] Add `Alt+W` key binding in search mode
- [ ] Implement whole word matching in search algorithm
- [ ] Add `ToggleWholeWord()` to `Editor`
- [ ] Re-run search when toggled
- [ ] Session persistence

#### Task 4.3: Visual Indicators
- [ ] Design clear text indicators: "Match Case" / "Ignore Case"
- [ ] Add "Whole Word" indicator when active
- [ ] Display in status bar (responsive to terminal width)
- [ ] Update immediately on toggle
- [ ] Use clear priority: current state > shortcuts

### Phase 4.5: Safety Features (Week 2)

#### Task 4.5.1: Replace-All Confirmation
- [ ] Implement confirmation dialog before replace-all
- [ ] Show match count in confirmation
- [ ] Add "Y/N/Preview" options
- [ ] Preview shows first 3-5 replacements

#### Task 4.5.2: Replace Undo Support
- [ ] Make replace-all a single undo operation
- [ ] Show "Replaced N matches - Ctrl+Z to undo" message
- [ ] Test undo/redo with replacements

#### Task 4.5.3: Performance Safeguards
- [ ] Implement search debouncing (150ms)
- [ ] Add "Searching..." indicator
- [ ] Limit max matches to 1000
- [ ] Show "1000+ matches (stopped)" when limit hit
- [ ] Add timeout for very slow searches (2 seconds)

### Phase 5: Error Handling & Polish (Week 3) ✅ **COMPLETED**

#### Task 5.1: Error State Management ✅
- [x] "No matches" shows error message but keeps search mode active
- [x] Use different status bar color for error state (`theme.Search.ErrorBg`)
- [x] Allow user to continue editing search term
- [x] Clear error when search becomes valid

**Implementation Details:**
- Modified `drawSearchStatus()` in `render.go` to detect error state and apply error background color
- Error state is active when: query is non-empty, not currently searching, and no matches found
- Search mode remains active allowing query editing
- Error clears automatically when matches are found

#### Task 5.2: Mode Transition Guardrails ✅
- [x] Document state machine in comments
- [x] Add assertions/guards for invalid transitions
- [x] Ensure clean state on mode exit
- [x] Handle edge cases:
  - Empty search string
  - Search while already in search mode
  - File changes during search

**Implementation Details:**
- Added comprehensive state machine diagram in `ui.go` showing all mode transitions
- Enhanced documentation for `enterSearch()`, `exitSearch()`, `performSearch()`, and `doSearch()`
- Added `EndSearchSession()` calls in `LoadFile()` and `SetNewFile()` to clear state on file changes
- Documented key handling, edge cases, thread safety, and side effects

#### Task 5.3: Session State Persistence ✅
- [x] Ensure search term persists correctly
- [x] Ensure replace term persists correctly
- [x] Ensure case sensitivity persists
- [x] Clear state on file switch

**Implementation Details:**
- `lastSearchQuery` persists and pre-fills search mode
- `lastReplaceTerm` persists and pre-fills replacement prompts
- `SearchState.CaseSensitive` and `SearchState.WholeWord` persist at editor level
- Search history maintains up to 20 recent queries
- Enhanced `SearchState` documentation explaining persistence behavior

**Tests Added:**
- `TestSearchTermPersistence` - Verifies search term persistence
- `TestReplaceTermPersistence` - Verifies replace term persistence
- `TestCaseSensitivityPersistence` - Verifies case sensitivity toggle persistence
- `TestWholeWordPersistence` - Verifies whole word toggle persistence
- `TestSearchHistoryPersistence` - Verifies search history navigation

### Phase 6: Testing (Week 3)

#### Task 6.1: Non-UI Unit Tests

**File: `internal/core/search_test.go`**
- [ ] `TestFindAllMatches` - various patterns, case sensitivity
- [ ] `TestSearchCaseSensitive` - verify case-sensitive matching
- [ ] `TestSearchCaseInsensitive` - verify case-insensitive matching
- [ ] `TestSearchWithEmptyQuery` - edge case
- [ ] `TestSearchMultipleLines` - cross-line scenarios
- [ ] `TestSearchSpecialCharacters` - special chars in pattern
- [ ] `TestSearchSessionLifecycle` - create, update, destroy

**File: `internal/core/editor_test.go`**
- [ ] `TestEditorIncrementalSearch` - integration with editor
- [ ] `TestToggleCaseSensitivity` - toggle and persistence
- [ ] `TestSearchSessionPersistence` - across multiple searches

#### Task 6.2: UI Automated Tests

**File: `internal/ui/search_test.go` (new) - Unified Mode Tests**
- [ ] `TestUnifiedSearchBasic` - type, see results in real-time
- [ ] `TestUnifiedSearchRealTime` - verify search on each keystroke
- [ ] `TestUnifiedSearchNoLeakage` - comprehensive key leakage test for ALL keys
- [ ] `TestSearchCaseToggle` - Alt+C toggles case sensitivity
- [ ] `TestSearchWholeWordToggle` - Alt+W toggles whole word
- [ ] `TestSearchHistory` - up/down arrows navigate history
- [ ] `TestSearchNavigationWhileTyping` - N/P work while typing
- [ ] `TestSearchFromSelection` - Ctrl+F pre-fills from selection
- [ ] `TestSearchEscape` - Esc exits properly
- [ ] `TestSearchErrorState` - no matches found behavior
- [ ] `TestSearchEmptyQuery` - handle empty search
- [ ] `TestSearchBackspaceOnEmpty` - backspace exits search
- [ ] `TestSearchF1Help` - F1 shows help overlay

**File: `internal/ui/replace_test.go` (new) - Replace Safety Tests**
- [ ] `TestReplaceCurrent` - R key prompts and replaces single match
- [ ] `TestReplaceAllConfirmation` - A key shows confirmation dialog
- [ ] `TestReplaceAllUndo` - Ctrl+Z undoes entire replace-all
- [ ] `TestReplaceAllPreview` - Preview option shows sample replacements
- [ ] `TestReplaceWithEmptyString` - edge case handling
- [ ] `TestReplaceAllCancel` - can cancel from confirmation
- [ ] `TestReplaceStatusMessage` - shows count and undo hint

**File: `internal/ui/render_test.go` (enhanced)**
- [ ] `TestSearchHighlighting` - verify matches are highlighted
- [ ] `TestSearchCurrentMatchHighlighting` - current match different color
- [ ] `TestStatusBarIncrementalSearch` - status bar rendering
- [ ] `TestStatusBarFindReplace` - status bar rendering

#### Task 6.3: Integration Tests
- [ ] Test full workflow: Ctrl+F → type → see highlights → Enter → navigate → replace
- [ ] Test multiple searches in same session
- [ ] Test case sensitivity toggle workflow
- [ ] Test error recovery workflow

### Phase 7: Documentation (Week 3)

#### Task 7.1: Code Documentation
- [ ] Add comprehensive godoc comments to all new functions
- [ ] Document state machine in `ui.go`
- [ ] Document search session lifecycle
- [ ] Update existing comments for modified functions

#### Task 7.2: User Documentation
- [ ] Update `docs/requirements.md` with search features
- [ ] Update help screen (`drawHelp()`) with case sensitivity toggle
- [ ] Update README if needed

#### Task 7.3: Architecture Documentation
- [ ] Update `docs/architecture.md` with search mode state machine
- [ ] Document SearchSession design
- [ ] Document incremental search algorithm

## Test Strategy

### Test Coverage Goals
- **Core Search Logic:** 100% coverage
- **UI Event Handlers:** 100% coverage for search modes
- **Integration:** All user workflows tested

### Testing Approach

#### 1. Non-UI Tests (internal/core)
These test pure logic without UI dependencies:
- Search algorithm correctness
- Case sensitivity logic
- Match finding and counting
- SearchSession state management

**Example:**
```go
func TestFindAllMatchesCaseInsensitive(t *testing.T) {
    lines := [][]rune{
        []rune("Hello World"),
        []rune("hello world"),
        []rune("HELLO WORLD"),
    }
    
    matches := FindAllMatches(lines, "hello", false)
    if len(matches) != 3 {
        t.Errorf("expected 3 matches, got %d", len(matches))
    }
    
    // Verify each match position...
}
```

#### 2. UI Tests with FakeScreen
Use the existing `FakeScreen` infrastructure in `fake_screen_test.go`:
- Simulate key presses
- Verify mode transitions
- Check no keys leak to editor buffer
- Verify status bar content
- Verify cursor positioning

**Example:**
```go
func TestIncrementalSearchNoLeakage(t *testing.T) {
    ui, screen := newTestUI(80, 24)
    
    // Start with "original text"
    insertText(ui, "original text")
    ui.editor.Apply(core.CmdMoveHome{}, 10)
    
    // Enter incremental search
    dispatch(ui, term.KeyEvent{Key: term.KeyCtrl, Rune: 'f'})
    if ui.mode != ModeIncrementalSearch {
        t.Fatal("should be in incremental search")
    }
    
    // Type search query
    for _, r := range "test" {
        dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
    }
    
    // Press random keys that could leak
    keys := []rune{'n', 'p', 'a', 'q', 'x', 'y', 'z'}
    for _, r := range keys {
        dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: r})
    }
    
    // Exit search
    dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
    
    // Verify text unchanged
    lines := ui.editor.Lines()
    text := string(lines[0])
    if text != "original text" {
        t.Errorf("text modified! Expected 'original text', got '%s'", text)
    }
}
```

#### 3. Integration Tests
Full workflow tests:
- Search → find → replace workflow
- Multiple searches in sequence
- Case sensitivity persistence
- Error recovery

## Risks and Mitigations

### Risk 1: Performance with Large Files
**Concern:** `FindAllMatches()` could be slow on files with thousands of lines

**Mitigation:**
- Only highlight matches in current viewport + buffer
- Implement match caching
- Set maximum match limit (e.g., 1000 matches)
- Profile and optimize

### Risk 2: Key Leakage Regression
**Concern:** Bug could reappear with future changes

**Mitigation:**
- Comprehensive test coverage
- Code review checklist for event handlers
- Defensive programming: always return true in mode-specific handlers
- Add assertions in development builds

### Risk 3: Mode Transition Bugs
**Concern:** Complex state machine could have edge cases

**Mitigation:**
- Document state machine thoroughly
- Add state transition guards/assertions
- Test all possible transitions
- Consider state machine library if complexity grows

### Risk 4: Visual Complexity
**Concern:** Multiple highlight types (selection, search, brackets) could conflict

**Mitigation:**
- Define clear precedence order: selection > current match > other matches > brackets
- Test with all highlighting types active
- Provide theme customization

## Success Criteria

### Must Have (MVP)
- [x] Key leakage bug is completely eliminated
- [ ] Unified search mode (no confusing mode transitions)
- [ ] Incremental search works: type → see results immediately
- [ ] Status bar shows match count and current position
- [ ] Case sensitivity toggle (Alt+C) works and persists
- [ ] Whole word toggle (Alt+W) works and persists
- [ ] Clear text indicators for options (not cryptic symbols)
- [ ] Search history with up/down arrows
- [ ] Pre-fill search from selection
- [ ] Replace-all confirmation dialog
- [ ] Replace-all is undoable
- [ ] F1 help overlay in search mode
- [ ] Escape always exits search
- [ ] All existing tests pass
- [ ] 100% test coverage for new code

### Should Have
- [ ] All matches highlighted in viewport
- [ ] Current match highlighted differently
- [ ] "No matches" error state without exiting search (red status bar)
- [ ] Search and replace terms persist across session
- [ ] Clean theme support for search colors
- [ ] Responsive status bar (degrades gracefully)
- [ ] Performance feedback ("Searching..." indicator)
- [ ] Search debouncing (150ms)
- [ ] Smart case detection

### Nice to Have (Future)
- [ ] Regex search support
- [ ] Whole word matching option
- [ ] Search wrap-around setting
- [ ] Search history (recent searches)
- [ ] Replace preview before committing
- [ ] Undo replace operation

## Future Enhancements

### Post-MVP Features

1. **Regex Search**
   - Toggle between literal and regex mode (Alt+R)
   - Display regex errors inline
   - Syntax highlighting in search prompt
   - Help overlay with common regex patterns

2. **Search History Persistence**
   - Persist search history to config file across sessions (in-memory already in MVP)
   - Configurable history size
   - Clear history option

3. **Advanced Options**
   - Search in selection (scope indicator)
   - Multi-line search patterns (Ctrl+Enter for newline)
   - Search wrap-around toggle
   - Search direction indicator (↑↓)

4. **Replace Enhancements**
   - Preview replacements before applying
   - Undo/redo individual replacements
   - Capture groups in regex replace

5. **Performance**
   - Incremental match finding (don't rescan entire file each keystroke)
   - Background indexing for large files
   - Match caching and invalidation

## Timeline Summary

| Phase | Duration | Tasks | Status |
|-------|----------|-------|--------|
| Phase 1: Infrastructure & Bug Fixes | Week 1 | Critical bugs, case-sensitive search, FindAllMatches | ✅ COMPLETED |
| Phase 2: Incremental Search Core | Week 1-2 | SearchSession, ModeSearch, real-time search | ✅ COMPLETED |
| Phase 3: Visual Feedback | Week 2 | Theme support, match highlighting, status bar | ✅ COMPLETED |
| Phase 4: Search Options Toggles | Week 2 | Case sensitivity, whole word, visual indicators | ✅ COMPLETED |
| Phase 4.5: Safety Features | Week 2 | Replace-all confirmation, undo support, safeguards | ✅ COMPLETED |
| Phase 5: Error Handling & Polish | Week 3 | Error states, guardrails, state persistence | ✅ COMPLETED |
| Phase 6: Testing | Week 3 | Unit tests, UI tests, integration tests | ✅ COMPLETED |
| Phase 7: Documentation | Week 3 | Code docs, user docs, architecture docs | ✅ COMPLETED |

**Status:** Phase 7 completed on January 10, 2026
**Phases Completed:** 1, 2, 3, 4, 4.5, 5, 6, 7
**Project Status:** ✅ **COMPLETE - ALL PHASES FINISHED**

## Notes

- This plan assumes one developer working full-time
- Priority should be given to fixing the key leakage bug first
- Testing should be written alongside implementation, not after
- Consider pair programming for complex state machine code
- Regular testing on Windows, Linux, and macOS to ensure terminal compatibility

## Questions Resolved (from UX Review)

1. **Match Highlighting Precedence:** What's the visual priority when a search match overlaps with a selection?
   - ✅ **Decided:** Selection takes priority over search highlighting

2. **Maximum Matches:** Should we limit the number of matches displayed?
   - ✅ **Decided:** 1000 matches max, show "1000+ matches (stopped)" if more

3. **Search Wrap-Around:** Should search wrap from end to beginning?
   - ✅ **Decided:** No wrap in MVP, add as configurable option later

4. **Empty Search Behavior:** What happens with empty search string?
   - ✅ **Decided:** Backspace on empty query exits search mode (intuitive)
   - Show placeholder "Search..." when empty
   - Display last search: "Last: 'pattern'"

5. **Replace All Confirmation:** Should "Replace All" ask for confirmation?
   - ✅ **Decided:** YES - show confirmation with match count
   - Add "Y/N/Preview" options
   - Make entire replace-all undoable as single operation

6. **Two Modes vs One:** Should we have separate incremental and find/replace modes?
   - ✅ **Decided:** Single unified ModeSearch
   - User can type and navigate simultaneously
   - No confusing mode transitions

7. **Case Sensitivity Keybinding:** Ctrl+I or Alt+C?
   - ✅ **Decided:** Alt+C (matches VS Code, avoids terminal conflicts)

8. **Case Indicator:** [Aa] vs text label?
   - ✅ **Decided:** Clear text labels: "Match Case" / "Ignore Case"

9. **Arrow Keys:** Match navigation or search history?
   - ✅ **Decided:** Search history (more valuable)
   - Use N/P/F3 for match navigation

## References

- Current implementation: `internal/core/search.go`, `internal/ui/ui.go`, `internal/ui/prompt.go`
- Existing tests: `internal/ui/findreplace_test.go`
- Related: Bracket matching feature (similar highlighting needs)
- Terminal library: `internal/term/`

---

**End of Document**
