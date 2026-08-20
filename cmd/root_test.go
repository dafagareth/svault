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

import "testing"

func TestShouldDefaultToList(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		// No args: show help (not list).
		{nil, false},
		{[]string{}, false},

		// A list flag routes to list.
		{[]string{"-l"}, true},
		{[]string{"-m"}, true},
		{[]string{"--json"}, true},
		{[]string{"--ns", "prod"}, true},

		// Root-level help/version stay with the root command.
		{[]string{"-h"}, false},
		{[]string{"--help"}, false},
		{[]string{"-v"}, false},
		{[]string{"--version"}, false},

		// A subcommand name (no leading dash) is left untouched.
		{[]string{"list"}, false},
		{[]string{"get", "DB_URL"}, false},
		{[]string{"completion", "zsh"}, false},
		{[]string{"__complete", "get", ""}, false},
	}
	for _, c := range cases {
		if got := shouldDefaultToList(c.args); got != c.want {
			t.Errorf("shouldDefaultToList(%v) = %v, want %v", c.args, got, c.want)
		}
	}
}
