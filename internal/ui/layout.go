package ui

type Rect struct {
	X, Y, W, H int
}

type Layout struct {
	Width, Height int
	
	Menubar   Rect
	Viewport  Rect
	Prompt    Rect
	StatusBar Rect
}

func ComputeLayout(w, h int, mode UIMode, hasMenubar, hasStatusBar bool) Layout {
	l := Layout{Width: w, Height: h}
	
	y := 0
	remH := h
	
	// Menubar
	if hasMenubar {
		l.Menubar = Rect{0, y, w, 1}
		y++
		remH--
	}
	
	// Status Bar (always at bottom when visible)
	// Always show statusbar in prompt/message/find/replace modes
	showStatusBar := hasStatusBar || mode == ModePrompt || mode == ModeMessage || 
	                 mode == ModeFindReplace
	
	if showStatusBar {
		l.StatusBar = Rect{0, h - 1, w, 1}
		remH--
	}
	
	// Prompt/Message (above Status Bar)
	if mode == ModePrompt || mode == ModeMessage {
		l.Prompt = Rect{0, h - 2, w, 1}
		remH--
	}
	
	// Viewport takes remaining space
	if remH < 0 {
		remH = 0
	}
	l.Viewport = Rect{0, y, w, remH}
	
	return l
}