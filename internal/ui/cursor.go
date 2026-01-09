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
