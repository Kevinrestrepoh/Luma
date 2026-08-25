package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type MenuModal struct {
	selected int
	width    int
	height   int
	options  []string
}

func NewMenuModal(selected int, width int, height int) *MenuModal {
	return &MenuModal{
		selected: selected,
		width:    width,
		height:   height,
		options:  []string{"Env Variables", "Settings"},
	}
}

func (m *MenuModal) View() string {
	modalW := 30
	modalH := len(m.options)
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalH > m.height-4 {
		modalH = m.height - 4
	}

	var content strings.Builder

	for i, opt := range m.options {
		var item string
		if i == m.selected {
			item = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Background(TextColor).
				Bold(true).
				Width(modalW - 4).
				Render(opt)
		} else {
			item = lipgloss.NewStyle().
				Foreground(TextColor).
				Width(modalW - 4).
				Render(opt)
		}
		content.WriteString(item)
		content.WriteString("\n")
	}

	body := lipgloss.NewStyle().
		Padding(1, 1).
		Width(modalW - 2).
		Render(content.String())

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Width(modalW).
		Height(modalH).
		Render(body)

	return modal
}

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

	body := lipgloss.NewStyle().
		Padding(1, 1).
		Width(modalW - 2).
		Render(content.String())

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Width(modalW).
		Height(modalH).
		Render(body)

	return modal
}

type EnvModal struct {
	envVars     []EnvVar
	tuiVars     []EnvVar
	selected    int
	width       int
	height      int
	creating    bool
	creatingKey bool
	keyInput    textinput.Model
	valueInput  textinput.Model
	editing     bool
	editingIdx  int
	editingKey  bool
}

func NewEnvModal(envVars []EnvVar, tuiVars []EnvVar, selected int, width int, height int, creating bool, creatingKey bool, keyInput textinput.Model, valueInput textinput.Model, editing bool, editingIdx int, editingKey bool) *EnvModal {
	return &EnvModal{
		envVars:     envVars,
		tuiVars:     tuiVars,
		selected:    selected,
		width:       width,
		height:      height,
		creating:    creating,
		creatingKey: creatingKey,
		keyInput:    keyInput,
		valueInput:  valueInput,
		editing:     editing,
		editingIdx:  editingIdx,
		editingKey:  editingKey,
	}
}

func (m *EnvModal) View() string {
	totalVars := len(m.envVars) + len(m.tuiVars)
	if m.creating {
		totalVars++
	}
	modalW := 50
	modalH := totalVars + 4
	if len(m.envVars) > 0 {
		modalH++
	}
	if len(m.tuiVars) > 0 {
		modalH++
	}
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalH > m.height-4 {
		modalH = m.height - 4
	}

	var content strings.Builder
	idx := 0

	if len(m.envVars) > 0 {
		header := lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true).
			Width(modalW - 4).
			Render(".env (read-only)")
		content.WriteString(header)
		content.WriteString("\n")

		for _, v := range m.envVars {
			line := fmt.Sprintf("  $%s = %s", v.Key, truncate(v.Value, modalW-20))
			var item string
			if idx == m.selected {
				item = lipgloss.NewStyle().
					Foreground(PrimaryColor).
					Background(TextColor).
					Bold(true).
					Width(modalW - 4).
					Render(line)
			} else {
				item = lipgloss.NewStyle().
					Foreground(TextColor).
					Width(modalW - 4).
					Render(line)
			}
			content.WriteString(item)
			content.WriteString("\n")
			idx++
		}
	}

	if len(m.tuiVars) > 0 {
		header := lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true).
			Width(modalW - 4).
			Render("TUI vars (editable)")
		content.WriteString(header)
		content.WriteString("\n")

		for i, v := range m.tuiVars {
			if m.editing && i == m.editingIdx {
				var item string
				if m.editingKey {
					label := lipgloss.NewStyle().Foreground(TextColor).Render("  KEY: ")
					item = lipgloss.NewStyle().
						Foreground(PrimaryColor).
						Width(modalW - 4).
						Render(label + m.keyInput.View())
				} else {
					label := lipgloss.NewStyle().Foreground(TextColor).Render("  KEY: ")
					valLabel := lipgloss.NewStyle().Foreground(TextColor).Render("  VALUE: ")
					item = lipgloss.NewStyle().
						Foreground(PrimaryColor).
						Width(modalW - 4).
						Render(label + m.keyInput.View() + "\n" + valLabel + m.valueInput.View())
				}
				content.WriteString(item)
				content.WriteString("\n")
				idx++
				continue
			}
			line := fmt.Sprintf("  $%s = %s", v.Key, truncate(v.Value, modalW-20))
			var item string
			if idx == m.selected {
				item = lipgloss.NewStyle().
					Foreground(PrimaryColor).
					Background(TextColor).
					Bold(true).
					Width(modalW - 4).
					Render(line)
			} else {
				item = lipgloss.NewStyle().
					Foreground(TextColor).
					Width(modalW - 4).
					Render(line)
			}
			content.WriteString(item)
			content.WriteString("\n")
			idx++
		}
	}

	if m.creating {
		header := lipgloss.NewStyle().
			Foreground(LiveDotColor).
			Bold(true).
			Width(modalW - 4).
			Render("Creating new var")
		content.WriteString(header)
		content.WriteString("\n")

		var item string
		if m.creatingKey {
			label := lipgloss.NewStyle().Foreground(TextColor).Render("  KEY: ")
			item = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Width(modalW - 4).
				Render(label + m.keyInput.View())
		} else {
			label := lipgloss.NewStyle().Foreground(TextColor).Render("  KEY: ")
			valLabel := lipgloss.NewStyle().Foreground(TextColor).Render("  VALUE: ")
			item = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Width(modalW - 4).
				Render(label + m.keyInput.View() + "\n" + valLabel + m.valueInput.View())
		}
		content.WriteString(item)
		content.WriteString("\n")
	}

	body := lipgloss.NewStyle().
		Padding(1, 1).
		Width(modalW - 2).
		Render(content.String())

	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Width(modalW).
		Height(modalH).
		Render(body)

	return modal
}
