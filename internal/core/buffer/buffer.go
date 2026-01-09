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

package buffer

type Buffer interface {
	InsertRune(r rune)
	InsertNewline()
	Backspace()

	MoveLeft()
	MoveRight()
	MoveUp()
	MoveDown()
	MoveHome()
	MoveEnd()

	Lines() [][]rune
	Cursor() (line, col int)
	SetCursor(line, col int)

	DeleteLine(line int) []rune
	InsertLine(line int, runes []rune)

	DeleteRange(startLine, startCol, endLine, endCol int)
	RangeText(startLine, startCol, endLine, endCol int) string

	// Helper for Undo/Redo to inspect state
	LineLen(line int) int
	RuneAt(line, col int) rune
}
