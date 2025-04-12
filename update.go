package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

var lastFocus string = "body"

func (m *model) Init() tea.Cmd {
	m.body.ShowLineNumbers = false
	m.UpdateStyles()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	lenMethods := len(m.methods)

	horizontal := m.width >= 50

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ApiResponse:
		if msg.err != nil {
			m.status = "Error: " + msg.err.Error()
			m.statusCode = 0
			m.output.SetContent("")
		} else {
			m.statusCode = msg.statusCode
			m.status = msg.status
			m.responseTime = msg.duration
			m.output.SetContent(msg.body)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == "normal" {
				return m, tea.Quit
			}

		case "tab":
			if m.focus == "url" {
				m.selectedMethod = ((m.selectedMethod + 1) % lenMethods)
				return m, nil
			}
		case "shift+tab":
			if m.focus == "url" {
				m.selectedMethod = (m.selectedMethod - 1 + lenMethods) % lenMethods
				return m, nil
			}

		case "i":
			if m.mode == "normal" {
				m.mode = "insert"
				switch m.focus {
				case "url":
					m.url.Focus()
					m.body.Blur()
				case "body":
					m.body.Focus()
					m.url.Blur()
				case "output":
					m.body.Blur()
					m.url.Blur()
				}

				m.UpdateStyles()
				return m, nil
			}
		case "esc":
			m.mode = "normal"
			m.url.Blur()
			m.body.Blur()
			m.UpdateStyles()
			return m, nil

		case "j", "down":
			switch m.mode {
			case "normal":
				if !horizontal {
					switch m.focus {
					case "body":
						m.focus = "output"
					case "output":
						m.focus = "url"
					default:
						m.focus = "body"
					}
					m.UpdateStyles()
					return m, nil
				}
				if m.focus == "url" {
					m.focus = lastFocus
				}
				m.UpdateStyles()
				return m, nil
			case "insert":
				if m.focus == "output" {
					m.output.LineDown(1)
					return m, nil
				}
			}

		case "k", "up":
			switch m.mode {
			case "normal":
				if !horizontal {
					switch m.focus {
					case "body":
						m.focus = "url"
					case "output":
						m.focus = "body"
					default:
						m.focus = "output"
					}
					m.UpdateStyles()
					return m, nil
				}

				if m.focus != "url" {
					lastFocus = m.focus
				}
				m.focus = "url"
				m.UpdateStyles()
				return m, nil
			case "insert":
				if m.focus == "output" {
					m.output.LineUp(1)
					return m, nil
				}
			}

		case "l", "right":
			if m.mode == "normal" && horizontal {
				m.focus = "output"
				m.UpdateStyles()
				return m, nil
			}

		case "h", "left":
			if m.mode == "normal" && horizontal {
				m.focus = "body"
				m.UpdateStyles()
				return m, nil
			}

		case "enter":
			if m.mode == "normal" {
				cmd := FetchApi(m.url.Value(), m.methods[m.selectedMethod].Name, m.body.Value())
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	if m.mode == "insert" {
		if m.focus == "url" {
			m.url, cmd = m.url.Update(msg)
		} else {
			m.body, cmd = m.body.Update(msg)
		}
	}

	return m, cmd
}
