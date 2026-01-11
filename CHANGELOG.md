# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-01-11

### Added
- Current line highlight feature (toggle via View menu, off by default)
- Per-theme hardcoded CurrentLineBg colors for all 14 themes
- Autosave system with idle-based trigger (2 second default)
- Automatic recovery prompt on startup when autosave exists
- Recovery options: Recover backup, Open original, Discard backup
- Cross-platform autosave storage (Windows, Linux, macOS)
- View menu toggle for enabling/disabling autosave
- Configurable autosave timing (idle_timeout, min_interval)

### Changed
- Test count increased to 140+ with 28 new autosave tests

## [0.2.0] - 2026-01-10

### Added
- Word navigation with Ctrl+Left and Ctrl+Right arrow keys
- Bracket matching and jumping with Ctrl+B
- Whitespace visualization toggle (Ctrl+Shift+W) showing spaces (·), tabs (→), and line endings (↵/¶)
- Separate Language menu for syntax highlighting language selection
- Support for opening non-existent files (creates new file on save)
- Smart tab rendering in whitespace mode (single arrow per tab character)

### Changed
- Language selection moved from View menu to dedicated top-level Language menu
- Language menu structure: Off/Auto options at top, followed by separator, then all languages
- Configuration only stores Off/Auto state for language; specific language selections are session-only
- Improved tab visualization to show single character per tab instead of repeating
- Removed "Ins Insert/Replace" text from statusbar for cleaner UI

### Fixed
- Tab character display in whitespace mode now shows single arrow (→) at tab start position

## [0.1.0] - 2026-01-09

### Added
- Initial release of cooledit
- Cross-platform terminal UI (Windows, Linux, macOS)
- UTF-8 and ISO-8859-1 encoding support with auto-detection
- LF and CRLF line ending detection and preservation
- Undo/redo with full history
- Find and replace with non-overlapping matches
- System clipboard integration (Ctrl+C/X/V)
- Line numbers toggle (Ctrl+L)
- Word wrap toggle (Ctrl+W)
- Configurable cursor shapes (block, underline, bar)
- 14 built-in color themes including retro DOS and IBM phosphor styles
- Auto-indentation
- Zen mode (F11 to hide status bar)
- Help dialog with keyboard shortcuts (F1)
- About dialog showing license information
- GPL-3.0 license compliance with copyright headers in all source files

### Changed
- N/A (initial release)

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

---

[0.1.0]: https://github.com/tomcoolpxl/cooledit/releases/tag/v0.1.0
