package ui

import "github.com/atotto/clipboard"

type SystemClipboard struct{}

func (c *SystemClipboard) Get() (string, error) {
	return clipboard.ReadAll()
}

func (c *SystemClipboard) Set(text string) error {
	return clipboard.WriteAll(text)
}