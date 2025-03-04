package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var lastFocus string

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
			if m.mode == "normal" && m.focus != "output" {
				m.mode = "insert"
				if m.focus == "url" {
					m.url.Focus()
					m.body.Blur()
				} else {
					m.body.Focus()
					m.url.Blur()
				}

				return m, nil
			}
		case "esc":
			m.mode = "normal"
			m.url.Blur()
			m.body.Blur()
			m.UpdateStyles()
			return m, nil
		case "v":
			if m.mode == "normal" && m.focus == "output" {
				m.mode = "visual"
				m.UpdateStyles()
				return m, nil
			}

		case "j", "down":
			if m.mode == "normal" {
				if !horizontal {
					if m.focus == "body" {
						m.focus = "output"
					} else if m.focus == "output" {
						m.focus = "url"
					} else {
						m.focus = "body"
					}
					m.UpdateStyles()
					return m, nil
				}
				m.focus = "body"
				m.UpdateStyles()
				return m, nil
			} else if m.mode == "visual" {
				m.output.LineDown(1)
				return m, nil
			}

		case "k", "up":
			if m.mode == "normal" {
				if !horizontal {
					if m.focus == "body" {
						m.focus = "url"
					} else if m.focus == "output" {
						m.focus = "body"
					} else {
						m.focus = "output"
					}
					m.UpdateStyles()
					return m, nil
				}
				m.focus = "url"
				m.UpdateStyles()
				return m, nil
			} else if m.mode == "visual" {
				m.output.LineUp(1)
				return m, nil
			}

		case "l", "right":
			if m.mode == "normal" && horizontal {
				lastFocus = m.focus
				m.focus = "output"
				m.UpdateStyles()
				return m, nil
			}

		case "h", "left":
			if m.mode == "normal" && horizontal && m.focus == "output" {
				m.focus = lastFocus
				m.UpdateStyles()
				return m, nil
			}

		case "enter":
			url := m.url.Value()
			if url != "" && m.mode == "normal" {
				method := m.methods[m.selectedMethod].Name
				body := m.body.Value()
				res, err := FetchApi(url, method, nil, body)
				if err != nil {
					return m, nil
				}
				m.output.SetContent(fmt.Sprintf("Status: %s \n\n%s", res.Status, res.Body))
				return m, nil
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
