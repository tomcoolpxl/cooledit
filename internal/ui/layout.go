// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
	// Always show statusbar in prompt/message modes
	showStatusBar := hasStatusBar || mode == ModePrompt || mode == ModeMessage

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
