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

package formatter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Default timeout for formatter execution
const DefaultTimeout = 5 * time.Second

// ErrTimeout is returned when the formatter times out
var ErrTimeout = errors.New("formatter timed out")

// ErrNotFound is returned when the formatter command is not found
var ErrNotFound = errors.New("formatter command not found")

// ExecuteResult contains the result of executing a formatter
type ExecuteResult struct {
	Stdout string
	Stderr string
}

// Execute runs an external command with the given input on stdin.
// Returns the stdout, stderr, and any error that occurred.
func Execute(command string, args []string, stdin string, timeout time.Duration) (*ExecuteResult, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	// Set up stdin
	cmd.Stdin = strings.NewReader(stdin)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	result := &ExecuteResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return result, ErrTimeout
	}

	// Check for command not found
	if err != nil {
		if execErr, ok := err.(*exec.Error); ok {
			if execErr.Err == exec.ErrNotFound {
				return result, fmt.Errorf("%w: %s", ErrNotFound, command)
			}
		}
		// Return the error with stderr info
		if result.Stderr != "" {
			// Get first line of stderr for error message
			firstLine := strings.Split(strings.TrimSpace(result.Stderr), "\n")[0]
			return result, fmt.Errorf("formatter failed: %s", firstLine)
		}
		return result, fmt.Errorf("formatter failed: %v", err)
	}

	return result, nil
}
