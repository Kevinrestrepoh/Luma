package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var lastFocus string = "request"

func (m *model) Init() tea.Cmd {
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
			} else if m.focus == "request" && m.mode != "insert" {
				m.requestSection.selectedTab = (m.requestSection.selectedTab + 1) % len(m.requestSection.tabs)
				return m, nil
			}
		case "shift+tab":
			if m.focus == "url" {
				m.selectedMethod = (m.selectedMethod - 1 + lenMethods) % lenMethods
				return m, nil
			} else if m.focus == "request" && m.mode != "insert" {
				m.requestSection.selectedTab = (m.requestSection.selectedTab - 1 + len(m.requestSection.tabs)) % len(m.requestSection.tabs)
				return m, nil
			}

		case "i":
			if m.mode == "normal" {
				m.mode = "insert"
				switch m.focus {
				case "url":
					m.url.Focus()
					m.body.Blur()
				case "request":
					switch m.requestSection.selectedTab {
					case 0: // Body
						m.body.Focus()
						m.url.Blur()
					case 1: // Headers
						if len(m.requestSection.headers) > 0 {
							m.requestSection.editingHeader = 0
							m.requestSection.headers[0].Inputs.Focus()
						} else {
							// Add first header if none exist
							newHeader := newRequestHeader()
							m.requestSection.headers = append(m.requestSection.headers, newHeader)
							m.requestSection.editingHeader = 0
							newHeader.Inputs.Focus()
						}
					case 2: // Params
						if len(m.requestSection.params) > 0 {
							m.requestSection.editingParam = 0
							m.requestSection.params[0].Inputs.Focus()
						} else {
							// Add first param if none exist
							newParam := newRequestParam()
							m.requestSection.params = append(m.requestSection.params, newParam)
							m.requestSection.editingParam = 0
							newParam.Inputs.Focus()
						}
					}
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
			if m.requestSection.editingParam >= 0 {
				param := m.requestSection.params[m.requestSection.editingParam]
				// Parse key=value format
				input := param.Inputs.Value()
				if idx := strings.Index(input, "="); idx != -1 {
					param.Key = strings.TrimSpace(input[:idx])
					param.Value = strings.TrimSpace(input[idx+1:])
				}
				m.requestSection.editingParam = -1
			}
			if m.requestSection.editingHeader >= 0 {
				header := m.requestSection.headers[m.requestSection.editingHeader]
				// Parse key: value format
				input := header.Inputs.Value()
				if idx := strings.Index(input, ":"); idx != -1 {
					header.Key = strings.TrimSpace(input[:idx])
					header.Value = strings.TrimSpace(input[idx+1:])
				}
				m.requestSection.editingHeader = -1
			}
			m.UpdateStyles()
			return m, nil

		case "enter":
			if m.mode == "normal" {
				// Convert params and headers to API format
				headers := make([]*ApiHeaders, len(m.requestSection.headers))
				for i, h := range m.requestSection.headers {
					headers[i] = &ApiHeaders{Key: h.Key, Value: h.Value}
				}

				// Build URL with query parameters
				url := m.url.Value()
				if len(m.requestSection.params) > 0 {
					url += "?"
					for i, param := range m.requestSection.params {
						if i > 0 {
							url += "&"
						}
						url += param.Key + "=" + param.Value
					}
				}

				cmd := FetchApi(url, m.methods[m.selectedMethod].Name, m.body.Value(), m.requestSection.headers)
				return m, cmd
			} else if m.mode == "insert" && m.focus == "request" {
				switch m.requestSection.selectedTab {
				case 0: // Body
					// No special handling needed for body
				case 1: // Headers
					if m.requestSection.editingHeader >= 0 {
						// Save current header and move to next
						header := m.requestSection.headers[m.requestSection.editingHeader]
						input := header.Inputs.Value()
						if idx := strings.Index(input, ":"); idx != -1 {
							header.Key = strings.TrimSpace(input[:idx])
							header.Value = strings.TrimSpace(input[idx+1:])
						}
						m.requestSection.editingHeader++
						if m.requestSection.editingHeader >= len(m.requestSection.headers) {
							// Add new header if we're at the end
							newHeader := newRequestHeader()
							m.requestSection.headers = append(m.requestSection.headers, newHeader)
							m.requestSection.editingHeader = len(m.requestSection.headers) - 1
							newHeader.Inputs.Focus()
						} else {
							m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
						}
					}
				case 2: // Params
					if m.requestSection.editingParam >= 0 {
						// Save current param and move to next
						param := m.requestSection.params[m.requestSection.editingParam]
						input := param.Inputs.Value()
						if idx := strings.Index(input, "="); idx != -1 {
							param.Key = strings.TrimSpace(input[:idx])
							param.Value = strings.TrimSpace(input[idx+1:])
						}
						m.requestSection.editingParam++
						if m.requestSection.editingParam >= len(m.requestSection.params) {
							// Add new param if we're at the end
							newParam := newRequestParam()
							m.requestSection.params = append(m.requestSection.params, newParam)
							m.requestSection.editingParam = len(m.requestSection.params) - 1
							newParam.Inputs.Focus()
						} else {
							m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
						}
					}
				}
			}

		case "backspace":
			if m.mode == "normal" && m.focus == "request" {
				switch m.requestSection.selectedTab {
				case 1: // Headers
					if len(m.requestSection.headers) > 0 {
						// Remove last header
						m.requestSection.headers = m.requestSection.headers[:len(m.requestSection.headers)-1]
						if m.requestSection.editingHeader >= len(m.requestSection.headers) {
							m.requestSection.editingHeader = -1
						}
					}
				case 2: // Params
					if len(m.requestSection.params) > 0 {
						// Remove last param
						m.requestSection.params = m.requestSection.params[:len(m.requestSection.params)-1]
						if m.requestSection.editingParam >= len(m.requestSection.params) {
							m.requestSection.editingParam = -1
						}
					}
				}
			}

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
						m.focus = "request"
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
				} else if m.focus == "request" {
					switch m.requestSection.selectedTab {
					case 0: // Body
						// No special handling needed for body
					case 1: // Headers
						if m.requestSection.editingHeader >= 0 {
							header := m.requestSection.headers[m.requestSection.editingHeader]
							input := header.Inputs.Value()
							if idx := strings.Index(input, ":"); idx != -1 {
								header.Key = strings.TrimSpace(input[:idx])
								header.Value = strings.TrimSpace(input[idx+1:])
							}
							m.requestSection.editingHeader++
							if m.requestSection.editingHeader >= len(m.requestSection.headers) {
								m.requestSection.editingHeader = 0
							}
							m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
						}
					case 2: // Params
						if m.requestSection.editingParam >= 0 {
							param := m.requestSection.params[m.requestSection.editingParam]
							input := param.Inputs.Value()
							if idx := strings.Index(input, "="); idx != -1 {
								param.Key = strings.TrimSpace(input[:idx])
								param.Value = strings.TrimSpace(input[idx+1:])
							}
							m.requestSection.editingParam++
							if m.requestSection.editingParam >= len(m.requestSection.params) {
								m.requestSection.editingParam = 0
							}
							m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
						}
					}
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
						m.focus = "request"
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
				} else if m.focus == "request" {
					switch m.requestSection.selectedTab {
					case 0: // Body
						// No special handling needed for body
					case 1: // Headers
						if m.requestSection.editingHeader >= 0 {
							header := m.requestSection.headers[m.requestSection.editingHeader]
							input := header.Inputs.Value()
							if idx := strings.Index(input, ":"); idx != -1 {
								header.Key = strings.TrimSpace(input[:idx])
								header.Value = strings.TrimSpace(input[idx+1:])
							}
							m.requestSection.editingHeader--
							if m.requestSection.editingHeader < 0 {
								m.requestSection.editingHeader = len(m.requestSection.headers) - 1
							}
							m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
						}
					case 2: // Params
						if m.requestSection.editingParam >= 0 {
							param := m.requestSection.params[m.requestSection.editingParam]
							input := param.Inputs.Value()
							if idx := strings.Index(input, "="); idx != -1 {
								param.Key = strings.TrimSpace(input[:idx])
								param.Value = strings.TrimSpace(input[idx+1:])
							}
							m.requestSection.editingParam--
							if m.requestSection.editingParam < 0 {
								m.requestSection.editingParam = len(m.requestSection.params) - 1
							}
							m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
						}
					}
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
				m.focus = "request"
				m.UpdateStyles()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.mode == "insert" {
		if m.focus == "url" {
			m.url, cmd = m.url.Update(msg)
		} else if m.focus == "request" {
			switch m.requestSection.selectedTab {
			case 0: // Body
				var newBody CustomTextarea
				newBody, cmd = m.body.Update(msg)
				m.body = newBody
			case 1: // Headers
				if m.requestSection.editingHeader >= 0 {
					header := m.requestSection.headers[m.requestSection.editingHeader]
					header.Inputs, cmd = header.Inputs.Update(msg)
				}
			case 2: // Params
				if m.requestSection.editingParam >= 0 {
					param := m.requestSection.params[m.requestSection.editingParam]
					param.Inputs, cmd = param.Inputs.Update(msg)
				}
			}
		}
	}

	return m, cmd
}
