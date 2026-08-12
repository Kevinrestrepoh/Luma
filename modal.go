package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Modal struct {
	windows  []*RequestWindow
	selected int
	width    int
	height   int
	methods  []*Method
}

func NewModal(windows []*RequestWindow, selected int, width int, height int, methods []*Method) *Modal {
	return &Modal{
		windows:  windows,
		selected: selected,
		width:    width,
		height:   height,
		methods:  methods,
	}
}

func (m *Modal) View() string {
	modalW := 50
	modalH := len(m.windows) + 3
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalH > m.height-4 {
		modalH = m.height - 4
	}

	var content strings.Builder

	start := 0
	end := len(m.windows)
	if end > modalH-3 {
		end = modalH - 3
	}

	for i := start; i < end; i++ {
		w := m.windows[i]
		methodName := ""
		if w.Method >= 0 && w.Method < len(m.methods) {
			methodName = m.methods[w.Method].Name
		}

		display := w.URL
		if display == "" {
			display = "(empty)"
		}
		if len(display) > modalW-14 {
			display = display[:modalW-17] + "..."
		}

		line := fmt.Sprintf("%-6s %s", methodName, display)

		var item string
		if i == m.selected {
			item = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Background(TextColor).
				Bold(true).
				Width(modalW - 4).
				Render(line)
		} else {
			methodColor := TextColor
			if w.Method >= 0 && w.Method < len(m.methods) {
				methodColor = m.methods[w.Method].Color
			}
			methodTag := lipgloss.NewStyle().
				Foreground(methodColor).
				Bold(true).
				Render(fmt.Sprintf("%-6s", methodName))
			item = lipgloss.NewStyle().
				Foreground(TextColor).
				Width(modalW - 4).
				Render(methodTag + " " + display)
		}
		content.WriteString(item)
		content.WriteString("\n")
	}

	help := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Width(modalW - 4).
		Align(lipgloss.Center).
		Render("j/k navigate | n new | d delete | m close")

	body := lipgloss.NewStyle().
		Padding(1, 1).
		Width(modalW - 2).
		Render(lipgloss.JoinVertical(lipgloss.Top,
			content.String(),
			help,
		))

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Width(modalW).
		Height(modalH).
		Render(body)

	return modal
}
