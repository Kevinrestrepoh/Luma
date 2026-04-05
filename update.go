package main

import (
	"context"
	"errors"
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
	wideUI := m.width >= 60

	switch msg := msg.(type) {
	case nil:
		return m, nil

	case streamResetMsg:
		if msg.id != m.streamID {
			return m, nil
		}
		m.streamBuf.Reset()
		m.streamFollow = true
		m.outputInteractMode = false
		m.showStreamControls = false
		m.output.SetContent("")
		if m.focus == "stop" {
			m.assignFocus("url")
			m.UpdateStyles()
		}
		return m, nil

	case streamHeaderMsg:
		if msg.id != m.streamID {
			return m, nil
		}
		m.statusCode = msg.statusCode
		m.status = msg.status
		m.responseTime = msg.ttfb
		m.showStreamControls = msg.showStreamControls
		if !m.showStreamControls && m.focus == "stop" {
			m.assignFocus("url")
			m.UpdateStyles()
		}
		return m, nil

	case streamDataMsg:
		if msg.id != m.streamID {
			return m, nil
		}
		_, _ = m.streamBuf.WriteString(msg.chunk)
		m.output.SetContent(sanitizeResponseText(m.streamBuf.String()))
		if m.streamFollow {
			m.output.GotoBottom()
		}
		return m, nil

	case streamDoneMsg:
		if msg.id != m.streamID {
			return m, nil
		}
		m.showStreamControls = false
		if m.cancelStream != nil {
			m.cancelStream = nil
		}
		if m.focus == "stop" {
			m.assignFocus("url")
			m.UpdateStyles()
		}
		m.responseTime = msg.duration
		if msg.err != nil && !errors.Is(msg.err, context.Canceled) {
			m.status = "Error: " + msg.err.Error()
			m.statusCode = 0
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.MouseMsg:
		if m.outputInteractMode && m.focus == "output" && m.outputScrollable() && m.mouseOverOutput(msg) {
			var cmd tea.Cmd
			m.output, cmd = m.output.Update(msg)
			m.syncStreamFollowToViewport()
			return m, cmd
		}

	case ApiResponse:
		if msg.err != nil {
			m.status = "Error: " + msg.err.Error()
			m.statusCode = 0
			m.output.SetContent("")
		} else {
			m.statusCode = msg.statusCode
			m.status = msg.status
			m.responseTime = msg.duration
			m.output.SetContent(sanitizeResponseText(msg.body))
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+x" && m.cancelStream != nil {
			m.abortStreaming()
			m.UpdateStyles()
			return m, nil
		}
		if handled, cmd := m.tryOutputScrollKeys(msg); handled {
			return m, cmd
		}
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == "normal" {
				if m.cancelStream != nil {
					m.cancelStream()
					m.cancelStream = nil
				}
				return m, tea.Quit
			}

		case "tab":
			if m.mode == "normal" && m.focus == "stop" {
				return m, nil
			}
			if m.mode == "normal" {
				if m.focus == "url" {
					m.selectedMethod = ((m.selectedMethod + 1) % lenMethods)
					return m, nil
				} else if m.focus == "request" {
					m.requestSection.selectedTab = (m.requestSection.selectedTab + 1) % len(m.requestSection.tabs)
					return m, nil
				}
			} else if m.mode == "insert" && m.focus == "request" {
				switch m.requestSection.selectedTab {
				case 0: // Body
					// Let the textarea handle the tab
					var newBody CustomTextarea
					var cmd tea.Cmd
					newBody, cmd = m.body.Update(msg)
					m.body = newBody
					return m, cmd
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
						m.requestSection.headersView.GotoTop()
						m.requestSection.headersView.LineDown(m.requestSection.editingHeader)
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
						m.requestSection.paramsView.GotoTop()
						m.requestSection.paramsView.LineDown(m.requestSection.editingParam)
					}
				}
				return m, nil
			}
		case "shift+tab":
			if m.mode == "normal" && m.focus == "stop" {
				return m, nil
			}
			if m.mode == "normal" {
				if m.focus == "url" {
					m.selectedMethod = (m.selectedMethod - 1 + lenMethods) % lenMethods
					return m, nil
				} else if m.focus == "request" {
					m.requestSection.selectedTab = (m.requestSection.selectedTab - 1 + len(m.requestSection.tabs)) % len(m.requestSection.tabs)
					return m, nil
				}
			} else if m.mode == "insert" && m.focus == "request" {
				switch m.requestSection.selectedTab {
				case 0: // Body
					// Let the textarea handle the shift+tab
					var newBody CustomTextarea
					var cmd tea.Cmd
					newBody, cmd = m.body.Update(msg)
					m.body = newBody
					return m, cmd
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
						m.requestSection.headersView.GotoTop()
						m.requestSection.headersView.LineDown(m.requestSection.editingHeader)
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
						m.requestSection.paramsView.GotoTop()
						m.requestSection.paramsView.LineDown(m.requestSection.editingParam)
					}
				}
				return m, nil
			}

		case "i":
			if m.focus == "stop" {
				return m, nil
			}
			if m.mode == "normal" && m.focus == "output" {
				m.outputInteractMode = !m.outputInteractMode
				m.UpdateStyles()
				return m, nil
			}
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
							m.requestSection.headersView.GotoTop()
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
							m.requestSection.paramsView.GotoTop()
						} else {
							// Add first param if none exist
							newParam := newRequestParam()
							m.requestSection.params = append(m.requestSection.params, newParam)
							m.requestSection.editingParam = 0
							newParam.Inputs.Focus()
						}
					}
				}

				m.UpdateStyles()
				return m, nil
			}
		case "esc":
			if m.focus == "stop" {
				m.assignFocus("url")
				m.UpdateStyles()
				return m, nil
			}
			if m.outputInteractMode {
				m.outputInteractMode = false
				m.UpdateStyles()
				return m, nil
			}
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
			if m.mode == "normal" && m.focus == "stop" && m.showStreamControls && m.cancelStream != nil {
				m.abortStreaming()
				m.UpdateStyles()
				return m, nil
			}
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

				if m.cancelStream != nil {
					m.cancelStream()
					m.cancelStream = nil
				}
				ctx, cancel := context.WithCancel(context.Background())
				m.cancelStream = cancel
				m.streamID++

				cmd := FetchApi(ctx, m.streamID, url, m.methods[m.selectedMethod].Name, m.body.Value(), headers)
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
							// Add new header if we're at the end and under limit
							if len(m.requestSection.headers) < 5 {
								newHeader := newRequestHeader()
								m.requestSection.headers = append(m.requestSection.headers, newHeader)
								m.requestSection.editingHeader = len(m.requestSection.headers) - 1
								newHeader.Inputs.Focus()
								// Scroll to the new header
								m.requestSection.headersView.GotoTop()
								m.requestSection.headersView.LineDown(m.requestSection.editingHeader)
							} else {
								m.requestSection.editingHeader = 0
								m.requestSection.headers[0].Inputs.Focus()
								// Scroll to the first header
								m.requestSection.headersView.GotoTop()
								m.requestSection.headersView.LineDown(0)
							}
						} else {
							m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
							// Scroll to the next header
							m.requestSection.headersView.GotoTop()
							m.requestSection.headersView.LineDown(m.requestSection.editingHeader)
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
							// Add new param if we're at the end and under limit
							if len(m.requestSection.params) < 5 {
								newParam := newRequestParam()
								m.requestSection.params = append(m.requestSection.params, newParam)
								m.requestSection.editingParam = len(m.requestSection.params) - 1
								newParam.Inputs.Focus()
								// Scroll to the new param
								m.requestSection.paramsView.GotoTop()
								m.requestSection.paramsView.LineDown(m.requestSection.editingParam)
							} else {
								m.requestSection.editingParam = 0
								m.requestSection.params[0].Inputs.Focus()
								// Scroll to the first param
								m.requestSection.paramsView.GotoTop()
								m.requestSection.paramsView.LineDown(0)
							}
						} else {
							m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
							// Scroll to the next param
							m.requestSection.paramsView.GotoTop()
							m.requestSection.paramsView.LineDown(m.requestSection.editingParam)
						}
					}
				}
			}

		case "alt+backspace":
			if m.mode == "insert" && m.focus == "request" {
				switch m.requestSection.selectedTab {
				case 1: // Headers
					if m.requestSection.editingHeader >= 0 {
						// Delete the current header
						m.requestSection.headers = append(
							m.requestSection.headers[:m.requestSection.editingHeader],
							m.requestSection.headers[m.requestSection.editingHeader+1:]...,
						)

						// If we still have headers, move to previous one
						if len(m.requestSection.headers) > 0 {
							m.requestSection.editingHeader--
							if m.requestSection.editingHeader < 0 {
								m.requestSection.editingHeader = 0
							}
							m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
							m.requestSection.headersView.GotoTop()
							m.requestSection.headersView.LineDown(m.requestSection.editingHeader)
						} else {
							// Add a new empty header when the list is empty
							newHeader := newRequestHeader()
							m.requestSection.headers = append(m.requestSection.headers, newHeader)
							m.requestSection.editingHeader = 0
							newHeader.Inputs.Focus()
						}

						return m, nil
					}
				case 2: // Params
					if m.requestSection.editingParam >= 0 {
						// Delete the current param
						m.requestSection.params = append(
							m.requestSection.params[:m.requestSection.editingParam],
							m.requestSection.params[m.requestSection.editingParam+1:]...,
						)

						// If we still have params, move to previous one
						if len(m.requestSection.params) > 0 {
							m.requestSection.editingParam--
							if m.requestSection.editingParam < 0 {
								m.requestSection.editingParam = 0
							}
							m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
							m.requestSection.paramsView.GotoTop()
							m.requestSection.paramsView.LineDown(m.requestSection.editingParam)
						} else {
							// Add a new empty param when the list is empty
							newParam := newRequestParam()
							m.requestSection.params = append(m.requestSection.params, newParam)
							m.requestSection.editingParam = 0
							newParam.Inputs.Focus()
						}

						return m, nil
					}
				}
			}

		case "j", "down":
			switch m.mode {
			case "normal":
				if wideUI && m.showStreamControls && m.focus == "stop" {
					m.assignFocus("output")
					m.UpdateStyles()
					return m, nil
				}
				if !horizontal {
					if m.showStreamControls {
						switch m.focus {
						case "request":
							m.assignFocus("stop")
						case "stop":
							m.assignFocus("output")
						case "output":
							m.assignFocus("url")
						default:
							m.assignFocus("request")
						}
						m.UpdateStyles()
						return m, nil
					}
					switch m.focus {
					case "request":
						m.assignFocus("output")
					case "output":
						m.assignFocus("url")
					default:
						m.assignFocus("request")
					}
					m.UpdateStyles()
					return m, nil
				}
				if m.focus == "url" {
					m.assignFocus(lastFocus)
				}
				m.UpdateStyles()
				return m, nil
			case "insert":
				if m.focus == "output" && m.outputInteractMode {
					m.output.LineDown(1)
					m.syncStreamFollowToViewport()
					return m, nil
				}
			}

		case "k", "up":
			switch m.mode {
			case "normal":
				if wideUI && m.showStreamControls {
					if m.focus == "output" {
						m.assignFocus("stop")
						m.UpdateStyles()
						return m, nil
					}
					if m.focus == "stop" {
						m.assignFocus("url")
						m.UpdateStyles()
						return m, nil
					}
				}
				if !horizontal {
					if m.showStreamControls {
						switch m.focus {
						case "request":
							m.assignFocus("url")
						case "stop":
							m.assignFocus("request")
						case "output":
							m.assignFocus("stop")
						default:
							m.assignFocus("output")
						}
						m.UpdateStyles()
						return m, nil
					}
					switch m.focus {
					case "request":
						m.assignFocus("url")
					case "output":
						m.assignFocus("request")
					default:
						m.assignFocus("output")
					}
					m.UpdateStyles()
					return m, nil
				}

				if m.focus != "url" {
					lastFocus = m.focus
				}
				m.assignFocus("url")
				m.UpdateStyles()
				return m, nil
			case "insert":
				if m.focus == "output" && m.outputInteractMode {
					m.output.LineUp(1)
					m.syncStreamFollowToViewport()
					return m, nil
				}
			}

		case "l", "right":
			if m.mode == "normal" && wideUI && m.showStreamControls && m.focus == "url" {
				m.assignFocus("stop")
				m.UpdateStyles()
				return m, nil
			}
			if m.mode == "normal" && wideUI && m.showStreamControls && m.focus == "stop" {
				m.assignFocus("output")
				m.UpdateStyles()
				return m, nil
			}
			if m.mode == "normal" && horizontal {
				m.assignFocus("output")
				m.UpdateStyles()
				return m, nil
			}

		case "h", "left":
			if m.mode == "normal" && wideUI && m.showStreamControls && m.focus == "stop" {
				m.assignFocus("url")
				m.UpdateStyles()
				return m, nil
			}
			if m.mode == "normal" && horizontal {
				m.assignFocus("request")
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
				var cmd tea.Cmd
				newBody, cmd = m.body.Update(msg)
				m.body = newBody
				return m, cmd
			case 1: // Headers
				if m.requestSection.editingHeader >= 0 {
					header := m.requestSection.headers[m.requestSection.editingHeader]
					header.Inputs, cmd = header.Inputs.Update(msg)
				} else {
					m.requestSection.headersView, cmd = m.requestSection.headersView.Update(msg)
				}
			case 2: // Params
				if m.requestSection.editingParam >= 0 {
					param := m.requestSection.params[m.requestSection.editingParam]
					param.Inputs, cmd = param.Inputs.Update(msg)
				} else {
					m.requestSection.paramsView, cmd = m.requestSection.paramsView.Update(msg)
				}
			}
		}
	}

	return m, cmd
}

func (m *model) outputScrollable() bool {
	return m.output.Height > 0 && m.output.TotalLineCount() > m.output.Height
}

func (m *model) urlBarWidths() (urlWidth int, statusWidth int) {
	methodWidth := 8
	statusWidth = 30
	if m.width < 80 {
		statusWidth = 26
	}
	urlWidth = m.width - methodWidth - 4 - statusWidth
	if m.width < 60 {
		urlWidth = m.width - methodWidth - 4
	}
	return urlWidth, statusWidth
}

func (m *model) maxLinesURLCalc(urlWidth int) int {
	urlText := m.url.Value()
	if len(urlText) > urlWidth-2 {
		return (len(urlText) / urlWidth) + 1
	}
	return 0
}

func (m *model) mouseOverOutput(msg tea.MouseMsg) bool {
	if msg.Action != tea.MouseActionPress {
		return false
	}
	if msg.Button != tea.MouseButtonWheelUp && msg.Button != tea.MouseButtonWheelDown {
		return false
	}
	urlW, _ := m.urlBarWidths()
	maxLines := m.maxLinesURLCalc(urlW)
	half := m.width / 2

	if m.width >= 60 {
		if msg.X <= half {
			return false
		}
		if msg.Y < 2+maxLines {
			return false
		}
		return true
	}
	outH := m.height/2 - 2 - maxLines/2
	if outH < 1 {
		outH = 1
	}
	if msg.Y < m.height-outH {
		return false
	}
	return true
}

func (m *model) tryOutputScrollKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.mode == "normal" && m.focus == "output" && msg.String() == "ctrl+g" {
		m.streamFollow = true
		m.output.GotoBottom()
		return true, nil
	}
	if !m.outputScrollable() || !m.outputInteractMode {
		return false, nil
	}
	if m.mode == "insert" && m.focus == "output" {
		return false, nil
	}
	if m.mode != "normal" || m.focus != "output" {
		return false, nil
	}
	switch msg.String() {
	case "j", "down":
		m.output.LineDown(1)
		m.syncStreamFollowToViewport()
		return true, nil
	case "k", "up":
		m.output.LineUp(1)
		m.syncStreamFollowToViewport()
		return true, nil
	case "pgdown", "f":
		m.output.ViewDown()
		m.syncStreamFollowToViewport()
		return true, nil
	case "pgup", "b":
		m.output.ViewUp()
		m.syncStreamFollowToViewport()
		return true, nil
	case " ":
		m.output.ViewDown()
		m.syncStreamFollowToViewport()
		return true, nil
	case "d", "ctrl+d":
		m.output.HalfViewDown()
		m.syncStreamFollowToViewport()
		return true, nil
	case "u", "ctrl+u":
		m.output.HalfViewUp()
		m.syncStreamFollowToViewport()
		return true, nil
	default:
		return false, nil
	}
}

func (m *model) assignFocus(f string) {
	m.focus = f
	if f != "output" {
		m.outputInteractMode = false
	}
}

func (m *model) syncStreamFollowToViewport() {
	m.streamFollow = m.output.AtBottom()
}

func (m *model) abortStreaming() {
	if m.cancelStream == nil {
		return
	}
	m.cancelStream()
	m.cancelStream = nil
	m.streamID++
	m.streamFollow = false
	m.outputInteractMode = false
	m.status = "Stopped"
	m.statusCode = 0
	m.showStreamControls = false
	if m.focus == "stop" {
		m.assignFocus("url")
	}
}
