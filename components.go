package main

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func truncate(text string, max int) string {
	if max < 4 {
		return ""
	}
	if lipgloss.Width(text) > max {
		return text[:max-3] + "..."
	}
	return text
}

func scrollBarView(vp viewport.Model, height int, interact bool) string {
	track := lipgloss.NewStyle().Foreground(ScrollTrackColor)
	thumb := lipgloss.NewStyle().Foreground(SecondaryColor)
	if !interact {
		track = track.Faint(true)
		thumb = thumb.Faint(true)
	}
	if height <= 0 {
		return ""
	}
	total := vp.TotalLineCount()
	if total <= height {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = track.Render("▏")
		}
		return strings.Join(lines, "\n")
	}
	thumbH := height * height / total
	if thumbH < 1 {
		thumbH = 1
	}
	maxY := total - height
	pos := 0
	if maxY > 0 {
		pos = vp.YOffset * (height - thumbH) / maxY
	}
	if pos < 0 {
		pos = 0
	}
	if pos+thumbH > height {
		pos = height - thumbH
	}
	lines := make([]string, height)
	for i := range lines {
		if i >= pos && i < pos+thumbH {
			lines[i] = thumb.Render("█")
		} else {
			lines[i] = track.Render("▏")
		}
	}
	return strings.Join(lines, "\n")
}

func streamingLiveDot() string {
	return lipgloss.NewStyle().
		Foreground(LiveDotColor).
		Padding(0, 1).
		Render("●")
}

func streamingStopView(focused bool) string {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(TextColor).
		Padding(0, 1)
	if focused {
		s = s.BorderForeground(TextColor)
	} else {
		s = s.BorderForeground(SecondaryColor)
	}
	return s.Render("Stop")
}

func renderTimeCell(width int, text string) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Right).
		Padding(0).
		Foreground(ResponseTimeColor).
		Render(text)
}

func (m *model) renderTabs(tabWidth int) string {
	tabs := make([]string, len(m.requestSection.tabs))
	for i, tab := range m.requestSection.tabs {
		style := m.requestStyles.InputField.Width(tabWidth - 2).BorderForeground(PrimaryColor)
		if i == m.requestSection.selectedTab {
			style = m.requestStyles.InputField.Width(tabWidth - 2)
		}
		tabs[i] = style.Render(tab)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *model) renderHeadersContent(width, height int) string {
	m.requestSection.headersView.Width = width
	m.requestSection.headersView.Height = height

	content := make([]string, len(m.requestSection.headers))
	for i, header := range m.requestSection.headers {
		if i == m.requestSection.editingHeader {
			content[i] = m.requestStyles.InputField.Width(width - 2).Render(header.Inputs.View())
		} else {
			content[i] = m.requestStyles.InputField.BorderForeground(PrimaryColor).Width(width - 2).Render(header.Key + ": " + header.Value)
		}
	}

	m.requestSection.headersView.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
	return m.requestStyles.InputField.Width(width).Height(height).Render(m.requestSection.headersView.View())
}

func (m *model) renderParamsContent(width, height int) string {
	m.requestSection.paramsView.Width = width
	m.requestSection.paramsView.Height = height

	content := make([]string, len(m.requestSection.params))
	for i, param := range m.requestSection.params {
		if i == m.requestSection.editingParam {
			content[i] = m.requestStyles.InputField.Width(width - 2).Render(param.Inputs.View())
		} else {
			content[i] = m.requestStyles.InputField.BorderForeground(PrimaryColor).Width(width - 2).Render(param.Key + "=" + param.Value)
		}
	}

	m.requestSection.paramsView.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
	return m.requestStyles.InputField.Width(width).Height(height).Render(m.requestSection.paramsView.View())
}

func (m *model) renderKeybinds() string {
	var parts []string
	switch {
	case m.showModal:
		parts = []string{"j/k navigate", "n new", "d delete", "↵ select", "m close"}
	case m.showMenuModal:
		parts = []string{"j/k navigate", "↵ select", "esc close"}
	case m.showEnvModal:
		if m.creatingEnv || m.editingEnv {
			parts = []string{"↵ confirm", "esc cancel"}
		} else {
			parts = []string{"j/k navigate", "↵ select", "n new", "e edit", "d delete", "esc close"}
		}
	case m.mode == "insert":
		if m.focus == "url" {
			parts = []string{"esc normal", "tab switch", "$ env"}
		} else {
			parts = []string{"esc normal", "$ env"}
		}
	case m.outputInteractMode:
		if m.outputSelectMode != "none" {
			parts = []string{"y yank", "esc cancel"}
		} else {
			parts = []string{"hjkl move", "w/b/e word", "W/B/E WORD", "v select", "y yank", "Y all", "esc normal"}
		}
	case m.focus == "output":
		if m.jsonPretty {
			parts = []string{"f raw", "m windows", "i interact", "p menu", "↵ send", "q quit"}
		} else {
			parts = []string{"f pretty", "m windows", "i interact", "p menu", "↵ send", "q quit"}
		}
	default:
		parts = []string{"tab switch", "m windows", "p menu", "i edit", "↵ send", "q quit"}
	}

	text := strings.Join(parts, " | ")
	style := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Width(m.width).
		Align(lipgloss.Center)

	return style.Render(truncate(text, m.width))
}

func (m *model) setOutput(raw string) {
	m.outputRaw = raw
	display := raw
	if m.jsonPretty {
		display = tryPrettyJSON(raw)
	}
	sanitized := sanitizeResponseText(display)
	m.outputLines = strings.Split(sanitized, "\n")
	m.output.SetContent(sanitized)
	if m.outputInteractMode {
		m.clampOutputCursor()
	}
}

func tryPrettyJSON(raw string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return raw
	}
	return string(out)
}

var (
	// Raw ANSI for cursor/selection. Reset restores the TextColor foreground
	// so the rest of the line keeps its original color after the styled char.
	// Using true color (24-bit) to match lipgloss #e5e5e5 exactly.
	cursorANSI    = "\033[48;2;229;229;229m\033[38;2;26;26;46m"
	selectionANSI = "\033[48;2;74;78;105m\033[38;2;229;229;229m"
	resetForeANSI = "\033[0m\033[38;2;229;229;229m"
)

func (m *model) renderOutputWithCursor() string {
	content := m.output.View()
	if !m.outputInteractMode || content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	if totalLines == 0 {
		return content
	}

	cursorLine := m.outputCursorLine - m.output.YOffset
	cursorCol := m.outputCursorCol
	selectMode := m.outputSelectMode

	anchorLine := m.outputSelectAnchorLine - m.output.YOffset
	anchorCol := m.outputSelectAnchorCol

	// Clamp cursor to visible viewport
	if cursorLine < 0 {
		cursorLine = 0
	}
	if cursorLine >= totalLines {
		cursorLine = totalLines - 1
	}
	cursorLineLen := lipgloss.Width(lines[cursorLine])
	if cursorLineLen == 0 {
		cursorCol = 0
	} else if cursorCol < 0 {
		cursorCol = 0
	} else if cursorCol >= cursorLineLen {
		cursorCol = cursorLineLen - 1
	}

	var result []string
	for i, line := range lines {
		if line == "" {
			result = append(result, line)
			continue
		}

		runes := []rune(line)
		var styled []string

		for j := 0; j < len(runes); j++ {
			char := string(runes[j])
			isSelected := false
			isCursor := (i == cursorLine && j == cursorCol)

			switch selectMode {
			case "char":
				aLine, aCol := anchorLine, anchorCol
				cLine, cCol := cursorLine, cursorCol
				if aLine > cLine || (aLine == cLine && aCol > cCol) {
					aLine, aCol, cLine, cCol = cLine, cCol, aLine, aCol
				}
				if i == aLine && i == cLine {
					isSelected = j >= aCol && j <= cCol
				} else if i == aLine {
					isSelected = j >= aCol
				} else if i == cLine {
					isSelected = j <= cCol
				} else if i > aLine && i < cLine {
					isSelected = true
				}
			case "line":
				minLine := anchorLine
				maxLine := cursorLine
				if anchorLine > cursorLine {
					minLine = cursorLine
					maxLine = anchorLine
				}
				isSelected = i >= minLine && i <= maxLine
			}

			if isCursor {
				styled = append(styled, cursorANSI+char+resetForeANSI)
			} else if isSelected {
				styled = append(styled, selectionANSI+char+resetForeANSI)
			} else {
				styled = append(styled, char)
			}
		}

		result = append(result, strings.Join(styled, ""))
	}

	return strings.Join(result, "\n")
}
