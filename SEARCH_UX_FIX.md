# Search UX Bug Fix: Letters Typed Instead of Commands

## Problem
When typing in the search field (Ctrl+F), certain letters like 'n', 'r', 'a', 'q' were triggering commands instead of being added to the search query. This made it impossible to search for words containing these letters.

### Example Bugs:
- Typing "n" triggered "next match" instead of adding 'n' to the search
- Typing "r" triggered "replace" instead of adding 'r' to the search  
- Typing "q" exited search mode instead of adding 'q' to the search
- This affected words like: "nanny", "runner", "question", etc.

## Root Cause
The search mode had Vim-like single-key commands that conflicted with normal text input. The implementation prioritized commands over text entry in a text input field - a fundamental UX violation.

## Solution

### 1. Fixed Key Handling (internal/ui/ui.go)
**Changed behavior:** ALL letters and printable characters are now added to the search query by default.

**Before:**
```go
case 'n', 'N':
    if e.Modifiers == term.ModShift {
        u.searchQuery = append(u.searchQuery, e.Rune)  // Only uppercase
    } else {
        u.nextSearchMatch()  // Lowercase navigates - BUG!
    }
```

**After:**
```go
// All regular runes are added to search query
u.searchQuery = append(u.searchQuery, e.Rune)
u.performSearch()
```

### 2. New Command Shortcuts
Since single letters now type text, commands use proper keyboard shortcuts:

| Old Shortcut | New Shortcut | Command |
|-------------|--------------|---------|
| `n` | `F3` or `Enter` | Next match |
| `p` | `Shift+F3` | Previous match |
| `r` | `Ctrl+R` | Replace current match |
| `a` | `Ctrl+H` | Replace all matches |
| `q` | `Escape` | Exit search |

These shortcuts follow common editor conventions (VS Code, Sublime, etc.)

### 3. Fixed Pre-fill Behavior (internal/ui/ui.go)
When entering search with text selected (Ctrl+F after selecting text), the search box is pre-filled with the selection. However, the original implementation would APPEND typed characters to the pre-filled text.

**Added:** `searchQueryPreFilled` flag
- Set to `true` when query is pre-filled from selection
- First keystroke REPLACES the pre-filled text (like VS Code)
- Subsequent keystrokes append normally

**Example:**
1. Select "Hello" and press Ctrl+F → search shows "Hello"
2. Type "w" → search changes to "w" (replaced, not "Hellow")
3. Type "o" → search becomes "wo" (appended)

### 4. Updated Documentation
Updated function comments and key binding documentation to reflect:
- Search is a TEXT INPUT field first, commands second
- Commands require modifiers (Ctrl, Alt) or function keys
- Clear explanation of all available shortcuts

## Testing

### New Tests Added (internal/ui/search_typing_test.go)
1. **TestSearchCanTypeAllLetters**: Verifies all 26 letters (a-z, A-Z) can be typed
2. **TestSearchCanTypeCommandLetters**: Specifically tests n, p, r, a, q
3. **TestSearchTypingNumbers**: Tests digit input
4. **TestSearchTypingSpecialCharacters**: Tests symbols and punctuation
5. **TestSearchTypingMixedContent**: Tests realistic search patterns like "function", "price", "return"

### Updated Existing Tests
Updated tests that expected the old behavior:
- `TestSearchNavigationWhileTyping`: Now uses F3 instead of 'n'
- `TestSearchUIIntegration`: Uses F3/Shift+F3 and Escape
- `TestReplaceCurrent`: Uses Ctrl+R instead of 'r'  
- `TestReplaceAllConfirmationDialog`: Uses Ctrl+H instead of 'a'
- `TestReplaceAllCancel`: Uses Ctrl+H
- `TestReplaceTermPersistence`: Uses Ctrl+R

## Files Modified
- `internal/ui/ui.go`: Main UX fix, added pre-fill logic, updated key handling
- `internal/ui/search_typing_test.go`: New comprehensive typing tests
- `internal/ui/search_test.go`: Updated navigation tests
- `internal/ui/ui_test.go`: Updated integration test
- `internal/ui/replace_test.go`: Updated replace tests
- `internal/ui/search_persistence_test.go`: Updated persistence tests

## Test Results
✅ All 71 tests in `internal/ui` pass
✅ Full test suite passes: `go test ./...`  
✅ Project builds successfully: `go build ./cmd/cooledit`

## User Impact
**Positive:**
- Can now search for ANY word without weird command triggers
- Follows standard editor UX conventions
- Pre-filled text behavior matches VS Code/Sublime

**Requires Adjustment:**
- Users who learned the old shortcuts need to use new ones
- F3 and Ctrl+R/Ctrl+H are more discoverable and standard
- Help screen (F1 in search) shows correct shortcuts

## Summary
This fix resolves a critical UX bug where normal text input was interrupted by command shortcuts. The search field now behaves like a proper text input, with commands accessed via standard modifier keys and function keys. This matches user expectations from other editors and makes the search feature actually usable for common searches containing letters like 'n', 'r', 'a', 'q'.
