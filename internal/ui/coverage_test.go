package ui

import (
	"os"
	"path/filepath"
	"testing"

	"cooledit/internal/core"
	"cooledit/internal/term"
)

func TestPromptCoverage(t *testing.T) {
	// 1. PromptQuitConfirm with 'y' and existing path
	t.Run("QuitConfirmSaveExisting", func(t *testing.T) {
		ui, _ := newTestUI(40, 5)
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		
		// Set content and save
		typeString(ui, "original")
		ui.editor.Apply(core.CmdSaveAs{Path: path}, 5)
		
		// Modify
		typeString(ui, "!")
		
		// Trigger Quit
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'q', Modifiers: term.ModCtrl})
		if ui.promptKind != PromptQuitConfirm {
			t.Fatalf("expected PromptQuitConfirm")
		}
		
		// Press 'y' -> Save and Quit
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'y'})
		
		if !ui.quitNow {
			t.Fatalf("expected quitNow after saving existing file")
		}
		
		content, _ := os.ReadFile(path)
		if string(content) != "original!" {
			t.Fatalf("expected saved content 'original!', got %q", string(content))
		}
	})

	// 2. PromptSaveAs with empty path
	t.Run("SaveAsEmptyPath", func(t *testing.T) {
		ui, _ := newTestUI(40, 5)
		ui.enterSaveAs(false)
		
		dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
		
		if ui.mode != ModeMessage {
			t.Fatalf("expected ModeMessage for empty path")
		}
		if ui.message != "Save As: empty path" {
			t.Fatalf("unexpected message: %q", ui.message)
		}
	})

	// 3. PromptFind with backspace and escape
	t.Run("FindBackspaceEscape", func(t *testing.T) {
		ui, _ := newTestUI(40, 5)
		ui.enterFind()
		
		typeString(ui, "abc")
		if string(ui.promptText) != "abc" {
			t.Fatalf("expected 'abc', got %q", string(ui.promptText))
		}
		
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
		if string(ui.promptText) != "ab" {
			t.Fatalf("expected 'ab', got %q", string(ui.promptText))
		}
		
		dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
		if ui.mode != ModeNormal {
			t.Fatalf("expected Normal mode after escape")
		}
	})

	// 4. PromptSaveAs with escape and backspace
	t.Run("SaveAsBackspaceEscape", func(t *testing.T) {
		ui, _ := newTestUI(40, 5)
		ui.enterSaveAs(false)
		
		typeString(ui, "file.txt")
		dispatch(ui, term.KeyEvent{Key: term.KeyBackspace})
		if string(ui.promptText) != "file.tx" {
			t.Fatalf("expected 'file.tx', got %q", string(ui.promptText))
		}
		
		dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
		if ui.mode != ModeNormal {
			t.Fatalf("expected Normal mode after escape")
		}
	})

	// 5. PromptQuitConfirm with escape
	t.Run("QuitConfirmEscape", func(t *testing.T) {
		ui, _ := newTestUI(40, 5)
		typeString(ui, "modified")
		dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'q', Modifiers: term.ModCtrl})
		
		dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
		if ui.mode != ModeNormal {
			t.Fatalf("expected Normal mode after escape")
		}
		if ui.quitNow {
			t.Fatalf("should not quit after escape")
		}
	})
	
	// 6. PromptOverwrite with escape
	t.Run("OverwriteEscape", func(t *testing.T) {
		ui, _ := newTestUI(40, 5)
		dir := t.TempDir()
		path := filepath.Join(dir, "exists.txt")
		os.WriteFile(path, []byte("hi"), 0644)
		
		ui.enterSaveAs(false)
		typeString(ui, path)
		dispatch(ui, term.KeyEvent{Key: term.KeyEnter})
		
		if ui.promptKind != PromptOverwrite {
			t.Fatalf("expected PromptOverwrite")
		}
		
		dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
		if ui.mode != ModeNormal {
			t.Fatalf("expected Normal mode after escape")
		}
	})
}

func TestTranslateKeyCoverage(t *testing.T) {
	ui, _ := newTestUI(40, 10)
	
	// F1 Help
	dispatch(ui, term.KeyEvent{Key: term.KeyF1})
	if ui.mode != ModeHelp {
		t.Fatalf("expected Help mode")
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: ' '}) // Exit help
	if ui.mode != ModeNormal {
		t.Fatalf("expected Normal mode after help")
	}
	
	// Ctrl+Shift+S
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 's', Modifiers: term.ModCtrl | term.ModShift})
	if ui.promptKind != PromptSaveAs {
		t.Fatalf("expected SaveAs prompt")
	}
	dispatch(ui, term.KeyEvent{Key: term.KeyEscape})
	
	// Ctrl+Shift+Z (Redo)
	dispatch(ui, term.KeyEvent{Key: term.KeyRune, Rune: 'z', Modifiers: term.ModCtrl | term.ModShift})
	// Should apply Redo (no crash, check message maybe)
	
	// F3 (Find Next) - no query yet
	dispatch(ui, term.KeyEvent{Key: term.KeyF3})
	if ui.mode != ModeMessage || ui.message != "No previous search" {
		t.Fatalf("expected 'No previous search' message")
	}
	
	// Shift+F3 (Find Prev)
	dispatch(ui, term.KeyEvent{Key: term.KeyF3, Modifiers: term.ModShift})
	if ui.mode != ModeMessage || ui.message != "No previous search" {
		t.Fatalf("expected 'No previous search' message")
	}
	
	// PageUp / PageDown
	dispatch(ui, term.KeyEvent{Key: term.KeyPageDown})
	dispatch(ui, term.KeyEvent{Key: term.KeyPageUp})
	
	// Ctrl+Home / Ctrl+End
	dispatch(ui, term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModCtrl})
	dispatch(ui, term.KeyEvent{Key: term.KeyEnd, Modifiers: term.ModCtrl})
	
	// Home / End
	dispatch(ui, term.KeyEvent{Key: term.KeyHome})
	dispatch(ui, term.KeyEvent{Key: term.KeyEnd})
}

