package fileio

import (
	"os"
	"path/filepath"
	"unicode/utf8"
)

type FileData struct {
	Path     string
	BaseName string
	Lines    [][]rune
	EOL      string // "\n" or "\r\n"
	Encoding string // "UTF-8" or "ISO-8859-1"
}

func Open(path string) (*FileData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	eol := "\n"
	for i := 0; i+1 < len(data); i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			eol = "\r\n"
			break
		}
	}

	enc := "UTF-8"
	isUTF8 := utf8.Valid(data)
	if !isUTF8 {
		enc = "ISO-8859-1"
	}

	var lines [][]rune
	var current []rune

	if isUTF8 {
		for i := 0; i < len(data); {
			b := data[i]

			if b == '\n' {
				lines = append(lines, current)
				current = nil
				i++
				continue
			}
			if b == '\r' {
				i++
				continue
			}

			r, size := utf8.DecodeRune(data[i:])
			if r == utf8.RuneError && size == 1 {
				// Should not happen if utf8.Valid(data) is true, but keep safe.
				current = append(current, rune(data[i]))
				i++
				continue
			}

			current = append(current, r)
			i += size
		}
	} else {
		// ISO-8859-1: bytes map 0..255 directly to Unicode code points.
		for i := 0; i < len(data); i++ {
			b := data[i]

			if b == '\n' {
				lines = append(lines, current)
				current = nil
				continue
			}
			if b == '\r' {
				continue
			}

			current = append(current, rune(b))
		}
	}

	lines = append(lines, current)

	return &FileData{
		Path:     path,
		BaseName: filepath.Base(path),
		Lines:    lines,
		EOL:      eol,
		Encoding: enc,
	}, nil
}
