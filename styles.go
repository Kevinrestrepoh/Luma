package main

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

var (
	PrimaryColor      = lipgloss.Color("#4a4e69")
	SecundaryColor    = lipgloss.Color("#8d99ae")
	TextColor         = lipgloss.Color("#e5e5e5")
	BackgroundColor   = lipgloss.Color("#64677C")
	ResponseTimeColor = lipgloss.Color("#d7d3c1")
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
		if m.mode == "insert" {
			m.urlStyles = visual
		} else {
			m.urlStyles = focus
			m.requestStyles = base
			m.outputStyles = base
		}

	case "request":
		if m.mode == "insert" {
			m.requestStyles = visual
		} else {
			m.requestStyles = focus
			m.urlStyles = base
			m.outputStyles = base
		}
	case "output":
		if m.mode == "insert" {
			m.outputStyles = visual
		} else {
			m.outputStyles = focus
			m.urlStyles = base
			m.requestStyles = base
		}

	default:
		m.urlStyles = base
		m.requestStyles = base
		m.outputStyles = base
	}
}

func StatusStyle(code int) lipgloss.Style {
	var color lipgloss.Color

	switch {
	case code >= 200 && code < 300:
		color = lipgloss.Color("#aaf683")
	case code >= 300 && code < 400:
		color = lipgloss.Color("#ffd97d")
	case code >= 400 && code < 500:
		color = lipgloss.Color("#ee6055")
	case code >= 500:
		color = lipgloss.Color("#ff5d8f")
	default:
		color = lipgloss.Color("#9e9e9e")
	}

	return lipgloss.NewStyle().Foreground(color).Padding(1)
}
