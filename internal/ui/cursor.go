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

import "cooledit/internal/term"

// ParseCursorShape converts a string to a CursorShape
func ParseCursorShape(s string) term.CursorShape {
	switch s {
	case "block":
		return term.CursorBlock
	case "underline":
		return term.CursorUnderline
	case "bar":
		return term.CursorBar
	default:
		return term.CursorBlock
	}
}

// CursorShapeToString converts a CursorShape to a string
func CursorShapeToString(shape term.CursorShape) string {
	switch shape {
	case term.CursorBlock:
		return "block"
	case term.CursorUnderline:
		return "underline"
	case term.CursorBar:
		return "bar"
	default:
		return "block"
	}
}

// GetAlternateCursorShape returns the alternate cursor shape for replace mode
// Logic: If insert is block → replace is underline
//        If insert is underline → replace is block
//        If insert is bar → replace is block
func GetAlternateCursorShape(insertShape term.CursorShape) term.CursorShape {
	switch insertShape {
	case term.CursorBlock:
		return term.CursorUnderline
	case term.CursorUnderline:
		return term.CursorBlock
	case term.CursorBar:
		return term.CursorBlock
	default:
		return term.CursorUnderline
	}
}
