package main

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

// sanitizeResponseText strips ANSI sequences and other characters that break a TUI when
// echoed inside lipgloss/viewport (e.g. HTML pages with escape codes or control bytes).
func sanitizeResponseText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = ansi.Strip(s)
	s = strings.ToValidUTF8(s, "")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' || r == '\t':
			b.WriteRune(r)
		case unicode.IsPrint(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}
