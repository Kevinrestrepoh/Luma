package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var lastFocus string

func (m *model) Init() tea.Cmd {
	m.url.Focus()
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
				return m, nil
			}
		case "esc":
			m.mode = "normal"
			return m, nil

		case "j", "down":
			if m.mode == "normal" {
				if !horizontal {
					if m.focus == "body" {
						m.focus = "output"
						m.body.Blur()
						m.url.Blur()
					} else if m.focus == "output" {
						m.focus = "url"
						m.url.Focus()
						m.body.Blur()
					} else {
						m.focus = "body"
						m.body.Focus()
						m.url.Blur()
					}
					m.UpdateStyles()
					return m, nil
				}
				m.focus = "body"
				m.url.Blur()
				m.body.Focus()
				m.UpdateStyles()
				return m, nil
			}

		case "k", "up":
			if m.mode == "normal" {
				if !horizontal {
					if m.focus == "body" {
						m.focus = "url"
						m.body.Blur()
						m.url.Focus()
					} else if m.focus == "output" {
						m.focus = "body"
						m.body.Focus()
						m.url.Blur()
					} else {
						m.focus = "output"
						m.url.Blur()
						m.body.Blur()
					}
					m.UpdateStyles()
					return m, nil
				}
				m.focus = "url"
				m.body.Blur()
				m.url.Focus()
				m.UpdateStyles()
				return m, nil
			}

		case "l", "right":
			if m.mode == "normal" && horizontal {
				lastFocus = m.focus
				m.focus = "output"
				m.body.Blur()
				m.url.Blur()
				m.UpdateStyles()
				return m, nil
			}

		case "h", "left":
			if m.mode == "normal" && horizontal && m.focus == "output" {
				m.focus = lastFocus
				if m.focus == "url" {
					m.url.Focus()
					m.body.Blur()
				} else {
					m.body.Focus()
					m.url.Blur()
				}
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
				m.output = fmt.Sprintf("Status: %s \n\n%s", res.Status, res.Body)
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
