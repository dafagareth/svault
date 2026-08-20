// Copyright 2026 Dafa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// clipboardBackend describes one platform clipboard backend: a command to copy
// (write stdin) and a command to paste (read stdout).
type clipboardBackend struct {
	copyCmd  []string
	pasteCmd []string
}

// clipboardBackends is the ordered list of backends we try, covering Wayland,
// X11, and macOS. The first one whose binary is found on PATH wins.
var clipboardBackends = []clipboardBackend{
	{copyCmd: []string{"wl-copy"}, pasteCmd: []string{"wl-paste", "--no-newline"}},
	{copyCmd: []string{"xclip", "-selection", "clipboard"}, pasteCmd: []string{"xclip", "-selection", "clipboard", "-o"}},
	{copyCmd: []string{"xsel", "--clipboard", "--input"}, pasteCmd: []string{"xsel", "--clipboard", "--output"}},
	{copyCmd: []string{"pbcopy"}, pasteCmd: []string{"pbpaste"}},
}

// availableClipboardBackend returns the first clipboard backend whose copy binary
// exists on the system, or false if none is available.
func availableClipboardBackend() (clipboardBackend, bool) {
	for _, t := range clipboardBackends {
		if _, err := exec.LookPath(t.copyCmd[0]); err == nil {
			return t, true
		}
	}
	return clipboardBackend{}, false
}

// clipboardAvailable reports whether any clipboard backend is installed.
func clipboardAvailable() bool {
	_, ok := availableClipboardBackend()
	return ok
}

// copyToClipboard and readClipboard are indirected through function variables
// so tests can stub clipboard access without a real backend.
var copyToClipboard = realCopyToClipboard
var readClipboard = realReadClipboard

// realCopyToClipboard writes s to the system clipboard.
func realCopyToClipboard(s string) error {
	t, ok := availableClipboardBackend()
	if !ok {
		return fmt.Errorf("no clipboard tool found (install wl-clipboard, xclip, or xsel)")
	}
	cmd := exec.Command(t.copyCmd[0], t.copyCmd[1:]...)
	cmd.Stdin = strings.NewReader(s)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard copy failed: %w", err)
	}
	return nil
}

// realReadClipboard returns the current clipboard contents, trimmed of a trailing newline.
func realReadClipboard() (string, error) {
	t, ok := availableClipboardBackend()
	if !ok {
		return "", fmt.Errorf("no clipboard tool found (install wl-clipboard, xclip, or xsel)")
	}
	out, err := exec.Command(t.pasteCmd[0], t.pasteCmd[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("clipboard read failed: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
