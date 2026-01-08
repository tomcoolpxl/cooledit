package fileio

import (
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

	// Try atomic save
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	
	err := os.Rename(tmp, path)
	if err != nil {
		// If rename failed (e.g. on Windows if target exists), fall back to direct write
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		os.Remove(tmp)
	}
	return nil
}