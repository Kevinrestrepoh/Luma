package main

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

var (
	PrimaryColor    = lipgloss.Color("#4a4e69")
	SecundaryColor  = lipgloss.Color("#5A6A77")
	TextColor       = lipgloss.Color("#BCC2DB")
	BackgroundColor = lipgloss.Color("#64677C")
)

func InitStyles() *Styles {
	s := &Styles{}

	s.BorderColor = PrimaryColor
	s.InputField = lipgloss.NewStyle().
		BorderForeground(s.BorderColor).
		Foreground(TextColor).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0)

	return s
}

func FocusStyles() *Styles {
	s := &Styles{}

	s.BorderColor = SecundaryColor
	s.InputField = lipgloss.NewStyle().
		BorderForeground(s.BorderColor).
		Foreground(TextColor).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0)

	return s
}

func VisualStyles() *Styles {
	s := &Styles{}

	s.BorderColor = TextColor
	s.InputField = lipgloss.NewStyle().
		BorderForeground(s.BorderColor).
		Foreground(TextColor).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0)

	return s
}

func (m *model) UpdateStyles() {
	base := InitStyles()
	focus := FocusStyles()
	visual := VisualStyles()

	switch m.focus {
	case "url":
		m.urlStyles = focus
		m.bodyStyles = base
		m.outputStyles = base

	case "body":
		m.bodyStyles = focus
		m.urlStyles = base
		m.outputStyles = base

	case "output":
		if m.mode == "visual" {
			m.outputStyles = visual
		} else {
			m.outputStyles = focus
			m.urlStyles = base
			m.bodyStyles = base
		}

	default:
		m.urlStyles = base
		m.bodyStyles = base
	}
}
