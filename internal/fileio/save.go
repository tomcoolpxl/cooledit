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

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
