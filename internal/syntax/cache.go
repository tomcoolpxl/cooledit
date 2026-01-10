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

package syntax

import (
	"hash/fnv"
)

// LineCache caches syntax tokens per line for efficient viewport-based rendering.
// It invalidates cached tokens when line content changes.
type LineCache struct {
	tokens      map[int][]Token // line number -> cached tokens
	hashes      map[int]uint64  // line number -> content hash for invalidation
	language    string
	highlighter Highlighter
}

// NewLineCache creates a new line cache for the given language.
// Returns nil if the language is not supported.
func NewLineCache(languageName string) *LineCache {
	highlighter := NewChromaHighlighter(languageName)
	if highlighter == nil {
		return nil
	}

	return &LineCache{
		tokens:      make(map[int][]Token),
		hashes:      make(map[int]uint64),
		language:    languageName,
		highlighter: highlighter,
	}
}

// GetTokens returns syntax tokens for a line, using cached values when available.
// The line number is 0-indexed.
func (c *LineCache) GetTokens(lineNum int, line []rune) []Token {
	if c == nil || c.highlighter == nil {
		return nil
	}

	// Calculate hash of current line content
	hash := hashLine(line)

	// Check if cache is valid
	if cachedHash, ok := c.hashes[lineNum]; ok && cachedHash == hash {
		return c.tokens[lineNum]
	}

	// Cache miss or content changed - tokenize the line
	tokens := c.highlighter.Tokenize(line)

	// Update cache
	c.tokens[lineNum] = tokens
	c.hashes[lineNum] = hash

	return tokens
}

// InvalidateLine marks a specific line's cache as stale.
// The next GetTokens call for this line will re-tokenize.
func (c *LineCache) InvalidateLine(lineNum int) {
	if c == nil {
		return
	}
	delete(c.tokens, lineNum)
	delete(c.hashes, lineNum)
}

// InvalidateRange invalidates all lines in the given range (inclusive).
// Useful when lines are inserted or deleted.
func (c *LineCache) InvalidateRange(startLine, endLine int) {
	if c == nil {
		return
	}
	for line := startLine; line <= endLine; line++ {
		delete(c.tokens, line)
		delete(c.hashes, line)
	}
}

// InvalidateFromLine invalidates all cached lines from the given line onwards.
// Useful after line insertions/deletions that shift subsequent lines.
func (c *LineCache) InvalidateFromLine(lineNum int) {
	if c == nil {
		return
	}
	for line := range c.tokens {
		if line >= lineNum {
			delete(c.tokens, line)
			delete(c.hashes, line)
		}
	}
}

// InvalidateAll clears the entire cache.
// Useful when switching files or languages.
func (c *LineCache) InvalidateAll() {
	if c == nil {
		return
	}
	c.tokens = make(map[int][]Token)
	c.hashes = make(map[int]uint64)
}

// Language returns the language name this cache is configured for.
func (c *LineCache) Language() string {
	if c == nil {
		return ""
	}
	return c.language
}

// SetLanguage changes the language and clears the cache.
// Returns true if the language was changed successfully.
func (c *LineCache) SetLanguage(languageName string) bool {
	if c == nil {
		return false
	}

	if c.language == languageName {
		return true // Already using this language
	}

	highlighter := NewChromaHighlighter(languageName)
	if highlighter == nil {
		return false
	}

	c.highlighter = highlighter
	c.language = languageName
	c.InvalidateAll()
	return true
}

// hashLine computes a fast hash of a line's content for cache invalidation.
func hashLine(line []rune) uint64 {
	h := fnv.New64a()
	for _, r := range line {
		// Write each rune as bytes
		b := [4]byte{
			byte(r),
			byte(r >> 8),
			byte(r >> 16),
			byte(r >> 24),
		}
		h.Write(b[:])
	}
	return h.Sum64()
}
