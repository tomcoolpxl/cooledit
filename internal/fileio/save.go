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

package fileio

import (
	"fmt"
	"os"
)

func Save(path string, lines [][]rune, eol string, encoding string) error {
	var data []byte

	for i, line := range lines {
		for _, r := range line {
			if encoding == "ISO-8859-1" && r > 255 {
				r = '?'
			}
			data = append(data, byte(r))
		}
		if i < len(lines)-1 {
			if eol == "\r\n" {
				data = append(data, '\r', '\n')
			} else {
				data = append(data, '\n')
			}
		}
	}

	// Try atomic save with restricted permissions (user read/write only)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}

	err := os.Rename(tmp, path)
	if err != nil {
		// If rename failed (e.g. on Windows if target exists), fall back to direct write
		writeErr := os.WriteFile(path, data, 0600)
		if writeErr != nil {
			_ = os.Remove(tmp) // Best-effort cleanup
			return fmt.Errorf("save failed: %w", writeErr)
		}
		// Direct write succeeded, clean up temp file (ignore error - file was saved)
		_ = os.Remove(tmp)
	}
	return nil
}
