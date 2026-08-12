package main

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
}

var (
	PrimaryColor    = lipgloss.Color("#4a4e69")
	SecondaryColor  = lipgloss.Color("#8d99ae")
	TextColor       = lipgloss.Color("#e5e5e5")
	BackgroundColor = lipgloss.Color("#64677C")

	ResponseTimeColor = lipgloss.Color("#d7d3c1")
	LiveDotColor      = lipgloss.Color("#7fdf8a")
	ScrollTrackColor  = lipgloss.Color("#5c5f77")

	StatusColor2xx = lipgloss.Color("#aaf683")
	StatusColor3xx = lipgloss.Color("#ffd97d")
	StatusColor4xx = lipgloss.Color("#ee6055")
	StatusColor5xx = lipgloss.Color("#ff5d8f")
	StatusColorDef = lipgloss.Color("#9e9e9e")
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
	s.BorderColor = SecondaryColor
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

func statusLineColor(code int) lipgloss.Color {
	switch {
	case code >= 200 && code < 300:
		return StatusColor2xx
	case code >= 300 && code < 400:
		return StatusColor3xx
	case code >= 400 && code < 500:
		return StatusColor4xx
	case code >= 500:
		return StatusColor5xx
	default:
		return StatusColorDef
	}
}

func statusInlineStyle(code int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(statusLineColor(code)).
		Padding(0, 1)
}
