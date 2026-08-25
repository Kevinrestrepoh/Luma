package main

import (
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
			parts = []string{"↵nter confirm", "esc cancel"}
		} else {
			parts = []string{"j/k navigate", "↵ select", "n new", "e edit", "d delete", "esc close"}
		}
	case m.mode == "insert":
		parts = []string{"esc normal", "tab switch", "$ env"}
	case m.outputInteractMode:
		parts = []string{"j/k scroll", "ctrl+g follow", "i/esc normal"}
	default:
		parts = []string{"m windows", "p menu", "i edit", "↵ send", "q quit"}
	}

	text := strings.Join(parts, " | ")
	style := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Width(m.width).
		Align(lipgloss.Center)

	return style.Render(truncate(text, m.width))
}
