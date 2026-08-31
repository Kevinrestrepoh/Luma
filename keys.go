package main

import (
	"context"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var lastFocus = "request"

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+x" && m.cancelStream != nil {
		m.abortStreaming()
		m.UpdateStyles()
		return m, nil
	}

	if m.showModal {
		return m.handleModalKeys(msg)
	}

	if m.showMenuModal {
		return m.handleMenuModalKeys(msg)
	}

	if m.showEnvModal {
		return m.handleEnvModalKeys(msg)
	}

	if msg.String() == "m" && m.mode == "normal" {
		m.saveCurrentWindow()
		m.showModal = true
		m.modalSelected = m.currentWindow
		return m, nil
	}

	if msg.String() == "p" && m.mode == "normal" {
		m.showMenuModal = true
		m.menuSelected = 0
		return m, nil
	}

	if handled, cmd := m.tryOutputScrollKeys(msg); handled {
		return m, cmd
	}

	if m.outputInteractMode {
		return m.handleInteractModeKeys(msg)
	}

	horizontal := m.width >= 50
	wideUI := m.width >= 60
	lenMethods := len(m.methods)

	switch msg.String() {
	case "ctrl+c", "q":
		if m.mode == "normal" {
			if m.cancelStream != nil {
				m.cancelStream()
				m.cancelStream = nil
			}
			m.save()
			return m, tea.Quit
		}

	case "tab":
		return m.handleTabKey(msg, lenMethods)
	case "shift+tab":
		return m.handleShiftTabKey(msg, lenMethods)
	case "i":
		if m.mode == "normal" {
			return m.handleInsertKey()
		}
	case "esc":
		return m.handleEscKey()
	case "enter":
		return m.handleEnterKey()
	case "f":
		if m.mode == "normal" && m.focus == "output" && m.outputRaw != "" {
			m.jsonPretty = !m.jsonPretty
			m.setOutput(m.outputRaw)
			return m, nil
		}
	case "alt+backspace":
		return m.handleAltBackspace()
	case "j", "down":
		if m.mode == "normal" {
			return m.handleDownKey(horizontal, wideUI)
		}
	case "k", "up":
		if m.mode == "normal" {
			return m.handleUpKey(horizontal, wideUI)
		}
	case "l", "right":
		if m.mode == "normal" {
			return m.handleRightKey(horizontal, wideUI)
		}
	case "h", "left":
		if m.mode == "normal" {
			return m.handleLeftKey(horizontal, wideUI)
		}
	}

	return m.handleInsertModeMsg(msg)
}

func (m *model) handleTabKey(msg tea.KeyMsg, lenMethods int) (tea.Model, tea.Cmd) {
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
	} else if m.mode == "insert" && m.focus == "url" {
		m.selectedMethod = ((m.selectedMethod + 1) % lenMethods)
		return m, nil
	} else if m.mode == "insert" && m.focus == "request" {
		switch m.requestSection.selectedTab {
		case 0:
			var newBody Textarea
			var cmd tea.Cmd
			newBody, cmd = m.body.Update(msg)
			m.body = newBody
			return m, cmd
		case 1:
			if m.requestSection.editingHeader >= 0 {
				parseHeaderInput(m.requestSection.headers[m.requestSection.editingHeader])
				m.requestSection.editingHeader++
				if m.requestSection.editingHeader >= len(m.requestSection.headers) {
					m.requestSection.editingHeader = 0
				}
				m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
				m.scrollToEditingHeader()
			}
		case 2:
			if m.requestSection.editingParam >= 0 {
				parseParamInput(m.requestSection.params[m.requestSection.editingParam])
				m.requestSection.editingParam++
				if m.requestSection.editingParam >= len(m.requestSection.params) {
					m.requestSection.editingParam = 0
				}
				m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
				m.scrollToEditingParam()
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *model) handleShiftTabKey(msg tea.KeyMsg, lenMethods int) (tea.Model, tea.Cmd) {
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
		case 0:
			var newBody Textarea
			var cmd tea.Cmd
			newBody, cmd = m.body.Update(msg)
			m.body = newBody
			return m, cmd
		case 1:
			if m.requestSection.editingHeader >= 0 {
				parseHeaderInput(m.requestSection.headers[m.requestSection.editingHeader])
				m.requestSection.editingHeader--
				if m.requestSection.editingHeader < 0 {
					m.requestSection.editingHeader = len(m.requestSection.headers) - 1
				}
				m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
				m.scrollToEditingHeader()
			}
		case 2:
			if m.requestSection.editingParam >= 0 {
				parseParamInput(m.requestSection.params[m.requestSection.editingParam])
				m.requestSection.editingParam--
				if m.requestSection.editingParam < 0 {
					m.requestSection.editingParam = len(m.requestSection.params) - 1
				}
				m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
				m.scrollToEditingParam()
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *model) handleInsertKey() (tea.Model, tea.Cmd) {
	if m.focus == "stop" {
		return m, nil
	}
	if m.mode == "normal" && m.focus == "output" {
		m.outputInteractMode = !m.outputInteractMode
		if m.outputInteractMode {
			m.outputCursorLine = 0
			m.outputCursorCol = 0
			m.outputSelectMode = "none"
			m.output.GotoTop()
		}
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
			case 0:
				m.body.Focus()
				m.url.Blur()
			case 1:
				if len(m.requestSection.headers) > 0 {
					m.requestSection.editingHeader = 0
					m.requestSection.headers[0].Inputs.Focus()
					m.requestSection.headersView.GotoTop()
				} else {
					newHeader := newRequestHeader()
					m.requestSection.headers = append(m.requestSection.headers, newHeader)
					m.requestSection.editingHeader = 0
					newHeader.Inputs.Focus()
				}
			case 2:
				if len(m.requestSection.params) > 0 {
					m.requestSection.editingParam = 0
					m.requestSection.params[0].Inputs.Focus()
					m.requestSection.paramsView.GotoTop()
				} else {
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
	return m, nil
}

func (m *model) handleEscKey() (tea.Model, tea.Cmd) {
	if m.focus == "stop" {
		m.assignFocus("url")
		m.UpdateStyles()
		return m, nil
	}
	if m.outputInteractMode {
		m.outputInteractMode = false
		m.outputSelectMode = "none"
		m.UpdateStyles()
		return m, nil
	}
	m.mode = "normal"
	m.url.Blur()
	m.body.Blur()
	if m.requestSection.editingParam >= 0 {
		parseParamInput(m.requestSection.params[m.requestSection.editingParam])
		m.requestSection.editingParam = -1
	}
	if m.requestSection.editingHeader >= 0 {
		parseHeaderInput(m.requestSection.headers[m.requestSection.editingHeader])
		m.requestSection.editingHeader = -1
	}
	m.UpdateStyles()
	return m, nil
}

func (m *model) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.mode == "normal" && m.focus == "stop" && m.showStreamControls && m.cancelStream != nil {
		m.abortStreaming()
		m.UpdateStyles()
		return m, nil
	}
	if m.mode == "normal" {
		headers := make([]*ApiHeaders, len(m.requestSection.headers))
		for i, h := range m.requestSection.headers {
			headers[i] = &ApiHeaders{Key: h.Key, Value: h.Value}
		}

		url := m.resolveEnvVars(m.url.Value())
		if len(m.requestSection.params) > 0 {
			url += "?"
			for i, param := range m.requestSection.params {
				if i > 0 {
					url += "&"
				}
				url += m.resolveEnvVars(param.Key) + "=" + m.resolveEnvVars(param.Value)
			}
		}

		if m.cancelStream != nil {
			m.cancelStream()
			m.cancelStream = nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelStream = cancel
		m.streamID++

		cmd := FetchApi(ctx, m.streamID, url, m.methods[m.selectedMethod].Name, m.resolveEnvVars(m.body.Value()), headers)
		return m, cmd
	} else if m.mode == "insert" && m.focus == "request" {
		switch m.requestSection.selectedTab {
		case 0:
		case 1:
			if m.requestSection.editingHeader >= 0 {
				parseHeaderInput(m.requestSection.headers[m.requestSection.editingHeader])
				m.requestSection.editingHeader++
				if m.requestSection.editingHeader >= len(m.requestSection.headers) {
					if len(m.requestSection.headers) < 5 {
						newHeader := newRequestHeader()
						m.requestSection.headers = append(m.requestSection.headers, newHeader)
						m.requestSection.editingHeader = len(m.requestSection.headers) - 1
						newHeader.Inputs.Focus()
					} else {
						m.requestSection.editingHeader = 0
						m.requestSection.headers[0].Inputs.Focus()
					}
				} else {
					m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
				}
				m.scrollToEditingHeader()
			}
		case 2:
			if m.requestSection.editingParam >= 0 {
				parseParamInput(m.requestSection.params[m.requestSection.editingParam])
				m.requestSection.editingParam++
				if m.requestSection.editingParam >= len(m.requestSection.params) {
					if len(m.requestSection.params) < 5 {
						newParam := newRequestParam()
						m.requestSection.params = append(m.requestSection.params, newParam)
						m.requestSection.editingParam = len(m.requestSection.params) - 1
						newParam.Inputs.Focus()
					} else {
						m.requestSection.editingParam = 0
						m.requestSection.params[0].Inputs.Focus()
					}
				} else {
					m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
				}
				m.scrollToEditingParam()
			}
		}
	}
	return m, nil
}

func (m *model) handleAltBackspace() (tea.Model, tea.Cmd) {
	if m.mode == "insert" && m.focus == "request" {
		switch m.requestSection.selectedTab {
		case 1:
			if m.requestSection.editingHeader >= 0 {
				m.requestSection.headers = append(
					m.requestSection.headers[:m.requestSection.editingHeader],
					m.requestSection.headers[m.requestSection.editingHeader+1:]...,
				)
				if len(m.requestSection.headers) > 0 {
					m.requestSection.editingHeader--
					if m.requestSection.editingHeader < 0 {
						m.requestSection.editingHeader = 0
					}
					m.requestSection.headers[m.requestSection.editingHeader].Inputs.Focus()
					m.scrollToEditingHeader()
				} else {
					newHeader := newRequestHeader()
					m.requestSection.headers = append(m.requestSection.headers, newHeader)
					m.requestSection.editingHeader = 0
					newHeader.Inputs.Focus()
				}
				return m, nil
			}
		case 2:
			if m.requestSection.editingParam >= 0 {
				m.requestSection.params = append(
					m.requestSection.params[:m.requestSection.editingParam],
					m.requestSection.params[m.requestSection.editingParam+1:]...,
				)
				if len(m.requestSection.params) > 0 {
					m.requestSection.editingParam--
					if m.requestSection.editingParam < 0 {
						m.requestSection.editingParam = 0
					}
					m.requestSection.params[m.requestSection.editingParam].Inputs.Focus()
					m.scrollToEditingParam()
				} else {
					newParam := newRequestParam()
					m.requestSection.params = append(m.requestSection.params, newParam)
					m.requestSection.editingParam = 0
					newParam.Inputs.Focus()
				}
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *model) handleDownKey(horizontal, wideUI bool) (tea.Model, tea.Cmd) {
	if m.outputInteractMode {
		return m, nil
	}
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
	return m, nil
}

func (m *model) handleUpKey(horizontal, wideUI bool) (tea.Model, tea.Cmd) {
	if m.outputInteractMode {
		return m, nil
	}
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
	return m, nil
}

func (m *model) handleRightKey(horizontal, wideUI bool) (tea.Model, tea.Cmd) {
	if m.outputInteractMode {
		return m, nil
	}
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
	return m, nil
}

func (m *model) handleLeftKey(horizontal, wideUI bool) (tea.Model, tea.Cmd) {
	if m.outputInteractMode {
		return m, nil
	}
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
	return m, nil
}

func (m *model) handleInsertModeMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode != "insert" {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "$" && (len(m.envVars) > 0 || len(m.tuiVars) > 0) {
			m.showEnvModal = true
			m.envSelected = 0
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case "url":
		m.url, cmd = m.url.Update(msg)
	case "request":
		switch m.requestSection.selectedTab {
		case 0:
			var newBody Textarea
			newBody, cmd = m.body.Update(msg)
			m.body = newBody
			return m, cmd
		case 1:
			if m.requestSection.editingHeader >= 0 {
				header := m.requestSection.headers[m.requestSection.editingHeader]
				header.Inputs, cmd = header.Inputs.Update(msg)
			} else {
				m.requestSection.headersView, cmd = m.requestSection.headersView.Update(msg)
			}
		case 2:
			if m.requestSection.editingParam >= 0 {
				param := m.requestSection.params[m.requestSection.editingParam]
				param.Inputs, cmd = param.Inputs.Update(msg)
			} else {
				m.requestSection.paramsView, cmd = m.requestSection.paramsView.Update(msg)
			}
		}
	}
	return m, cmd
}

func (m *model) handleModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "m":
		m.showModal = false
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.modalSelected < len(m.windows)-1 {
			m.modalSelected++
		}
		return m, nil
	case "k", "up":
		if m.modalSelected > 0 {
			m.modalSelected--
		}
		return m, nil
	case "enter":
		m.save()
		m.loadWindow(m.modalSelected)
		m.showModal = false
		m.UpdateStyles()
		return m, nil
	case "n":
		m.newWindow()
		m.showModal = false
		m.UpdateStyles()
		return m, nil
	case "d":
		if len(m.windows) > 1 {
			m.deleteWindow(m.modalSelected)
			if m.modalSelected >= len(m.windows) {
				m.modalSelected = len(m.windows) - 1
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *model) handleMenuModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "p":
		m.showMenuModal = false
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.menuSelected < len([]string{"Env Variables", "Settings"})-1 {
			m.menuSelected++
		}
		return m, nil
	case "k", "up":
		if m.menuSelected > 0 {
			m.menuSelected--
		}
		return m, nil
	case "enter":
		m.showMenuModal = false
		if m.menuSelected == 0 {
			m.showEnvModal = true
			m.envSelected = 0
		}
		return m, nil
	}
	return m, nil
}

func (m *model) handleEnvModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.creatingEnv {
		return m.handleEnvCreationKeys(msg)
	}
	if m.editingEnv {
		return m.handleEnvEditKeys(msg)
	}

	totalVars := len(m.envVars) + len(m.tuiVars)
	switch msg.String() {
	case "esc", "p":
		m.showEnvModal = false
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.envSelected < totalVars-1 {
			m.envSelected++
		}
		return m, nil
	case "k", "up":
		if m.envSelected > 0 {
			m.envSelected--
		}
		return m, nil
	case "enter":
		if m.envSelected >= 0 && m.envSelected < totalVars {
			var envVar EnvVar
			if m.envSelected < len(m.envVars) {
				envVar = m.envVars[m.envSelected]
			} else {
				envVar = m.tuiVars[m.envSelected-len(m.envVars)]
			}
			insertion := "$" + envVar.Key
			m.insertIntoFocus(insertion)
		}
		m.showEnvModal = false
		return m, nil
	case "n":
		if len(m.tuiVars) < 10 {
			m.creatingEnv = true
			m.envCreatingKey = true
			m.envKeyInput.SetValue("")
			m.envKeyInput.Placeholder = "KEY"
			m.envKeyInput.Focus()
			m.envValueInput.SetValue("")
			m.envValueInput.Placeholder = "VALUE"
		}
		return m, nil
	case "e":
		if m.envSelected >= len(m.envVars) {
			tuiIdx := m.envSelected - len(m.envVars)
			m.editingEnv = true
			m.editingEnvIdx = tuiIdx
			m.editingKey = true
			m.envKeyInput.SetValue(m.tuiVars[tuiIdx].Key)
			m.envKeyInput.Focus()
			m.envValueInput.SetValue(m.tuiVars[tuiIdx].Value)
		}
		return m, nil
	case "d":
		if m.envSelected >= len(m.envVars) {
			tuiIdx := m.envSelected - len(m.envVars)
			m.tuiVars = append(m.tuiVars[:tuiIdx], m.tuiVars[tuiIdx+1:]...)
			if m.envSelected >= len(m.envVars)+len(m.tuiVars) {
				m.envSelected = len(m.envVars) + len(m.tuiVars) - 1
			}
			m.save()
		}
		return m, nil
	}
	return m, nil
}

func (m *model) handleEnvEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editingEnv = false
		m.envKeyInput.Blur()
		m.envValueInput.Blur()
		return m, nil
	case "enter":
		if m.editingKey {
			if m.envKeyInput.Value() == "" {
				m.editingEnv = false
				m.envKeyInput.Blur()
				return m, nil
			}
			m.editingKey = false
			m.envKeyInput.Blur()
			m.envValueInput.Focus()
			return m, nil
		}
		m.tuiVars[m.editingEnvIdx] = EnvVar{
			Key:   m.envKeyInput.Value(),
			Value: m.envValueInput.Value(),
		}
		m.editingEnv = false
		m.envKeyInput.Blur()
		m.envValueInput.Blur()
		m.save()
		return m, nil
	}
	var cmd tea.Cmd
	if m.editingKey {
		m.envKeyInput, cmd = m.envKeyInput.Update(msg)
	} else {
		m.envValueInput, cmd = m.envValueInput.Update(msg)
	}
	return m, cmd
}

func (m *model) handleEnvCreationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.creatingEnv = false
		m.envKeyInput.Blur()
		m.envValueInput.Blur()
		return m, nil
	case "enter":
		if m.envCreatingKey {
			if m.envKeyInput.Value() == "" {
				m.creatingEnv = false
				m.envKeyInput.Blur()
				return m, nil
			}
			m.envCreatingKey = false
			m.envKeyInput.Blur()
			m.envValueInput.Focus()
			return m, nil
		}
		m.tuiVars = append(m.tuiVars, EnvVar{
			Key:   m.envKeyInput.Value(),
			Value: m.envValueInput.Value(),
		})
		m.envSelected = len(m.envVars) + len(m.tuiVars) - 1
		m.creatingEnv = false
		m.envKeyInput.Blur()
		m.envValueInput.Blur()
		m.save()
		return m, nil
	}
	var cmd tea.Cmd
	if m.envCreatingKey {
		m.envKeyInput, cmd = m.envKeyInput.Update(msg)
	} else {
		m.envValueInput, cmd = m.envValueInput.Update(msg)
	}
	return m, cmd
}

func (m *model) insertIntoFocus(text string) {
	switch m.focus {
	case "url":
		m.url.SetValue(m.url.Value() + text)
	case "request":
		switch m.requestSection.selectedTab {
		case 0:
			m.body.SetValue(m.body.Value() + text)
		case 1:
			if m.requestSection.editingHeader >= 0 {
				h := m.requestSection.headers[m.requestSection.editingHeader]
				h.Inputs.SetValue(h.Inputs.Value() + text)
			}
		case 2:
			if m.requestSection.editingParam >= 0 {
				p := m.requestSection.params[m.requestSection.editingParam]
				p.Inputs.SetValue(p.Inputs.Value() + text)
			}
		}
	}
}

func (m *model) tryOutputScrollKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.mode == "normal" && m.focus == "output" && msg.String() == "ctrl+g" {
		m.streamFollow = true
		if m.outputInteractMode {
			m.output.GotoBottom()
			total := m.output.TotalLineCount()
			if total > 0 {
				m.outputCursorLine = total - 1
				vpLines := strings.Split(m.output.View(), "\n")
				if m.outputCursorLine < len(vpLines) {
					lineLen := lipgloss.Width(vpLines[m.outputCursorLine])
					if lineLen > 0 {
						m.outputCursorCol = lineLen - 1
					} else {
						m.outputCursorCol = 0
					}
				}
			}
		}
		return true, nil
	}
	return false, nil
}

func (m *model) clampOutputCursor() {
	lines := strings.Split(m.output.View(), "\n")
	if len(lines) == 0 {
		m.outputCursorLine = 0
		m.outputCursorCol = 0
		return
	}
	if m.outputCursorLine < 0 {
		m.outputCursorLine = 0
	}
	if m.outputCursorLine >= len(lines) {
		m.outputCursorLine = len(lines) - 1
	}
	lineLen := lipgloss.Width(lines[m.outputCursorLine])
	if lineLen == 0 {
		m.outputCursorCol = 0
	} else if m.outputCursorCol < 0 {
		m.outputCursorCol = 0
	} else if m.outputCursorCol >= lineLen {
		m.outputCursorCol = lineLen - 1
	}
}

func (m *model) scrollViewportToCursor() {
	vpH := m.output.Height
	if vpH <= 0 {
		return
	}
	if m.outputCursorLine < m.output.YOffset {
		m.output.YOffset = m.outputCursorLine
	} else if m.outputCursorLine >= m.output.YOffset+vpH {
		m.output.YOffset = m.outputCursorLine - vpH + 1
	}
}

func (m *model) handleInteractModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.outputInteractMode {
		return m, nil
	}

	key := msg.String()
	lines := strings.Split(m.output.View(), "\n")
	totalLines := len(lines)

	switch key {
	case "esc":
		if m.outputSelectMode != "none" {
			m.outputSelectMode = "none"
		} else {
			m.outputInteractMode = false
			m.UpdateStyles()
		}
		return m, nil

	case "j", "down":
		if m.outputCursorLine < totalLines-1 {
			m.outputCursorLine++
			m.clampOutputCursor()
			m.scrollViewportToCursor()
		}
		return m, nil

	case "k", "up":
		if m.outputCursorLine > 0 {
			m.outputCursorLine--
			m.clampOutputCursor()
			m.scrollViewportToCursor()
		}
		return m, nil

	case "l", "right":
		lineLen := lipgloss.Width(lines[m.outputCursorLine])
		if lineLen > 0 && m.outputCursorCol < lineLen-1 {
			m.outputCursorCol++
		} else if m.outputCursorLine < totalLines-1 {
			m.outputCursorLine++
			m.outputCursorCol = 0
			m.clampOutputCursor()
		}
		m.scrollViewportToCursor()
		return m, nil

	case "h", "left":
		if m.outputCursorCol > 0 {
			m.outputCursorCol--
		} else if m.outputCursorLine > 0 {
			m.outputCursorLine--
			prevLine := lines[m.outputCursorLine]
			m.outputCursorCol = max(0, lipgloss.Width(prevLine)-1)
		}
		m.scrollViewportToCursor()
		return m, nil

	case "w":
		col := m.outputCursorCol
		line := lines[m.outputCursorLine]
		runes := []rune(line)
		if col < len(runes) {
			if !isWordChar(runes[col]) {
				for col < len(runes) && !isWordChar(runes[col]) {
					col++
				}
			} else {
				for col < len(runes) && isWordChar(runes[col]) {
					col++
				}
				for col < len(runes) && !isWordChar(runes[col]) {
					col++
				}
			}
		}
		if col >= len(runes) {
			if m.outputCursorLine < totalLines-1 {
				m.outputCursorLine++
				m.outputCursorCol = 0
			} else {
				m.outputCursorCol = max(0, lipgloss.Width(line)-1)
			}
		} else {
			m.outputCursorCol = col
		}
		m.clampOutputCursor()
		m.scrollViewportToCursor()
		return m, nil

	case "W":
		col := m.outputCursorCol
		line := lines[m.outputCursorLine]
		runes := []rune(line)
		if col < len(runes) {
			if unicode.IsSpace(runes[col]) {
				for col < len(runes) && unicode.IsSpace(runes[col]) {
					col++
				}
			} else {
				for col < len(runes) && !unicode.IsSpace(runes[col]) {
					col++
				}
			}
		}
		if col >= len(runes) {
			if m.outputCursorLine < totalLines-1 {
				m.outputCursorLine++
				m.outputCursorCol = 0
			} else {
				m.outputCursorCol = max(0, lipgloss.Width(line)-1)
			}
		} else {
			m.outputCursorCol = col
		}
		m.clampOutputCursor()
		m.scrollViewportToCursor()
		return m, nil

	case "b":
		col := m.outputCursorCol
		if col > 0 {
			line := lines[m.outputCursorLine]
			runes := []rune(line)
			if col > len(runes) {
				col = len(runes)
			}
			col--
			for col > 0 && !isWordChar(runes[col]) {
				col--
			}
			if col > 0 {
				for col > 0 && isWordChar(runes[col-1]) {
					col--
				}
			}
			m.outputCursorCol = col
		} else if m.outputCursorLine > 0 {
			m.outputCursorLine--
			prevLine := lines[m.outputCursorLine]
			prevRunes := []rune(prevLine)
			m.outputCursorCol = max(0, len(prevRunes)-1)
			if m.outputCursorCol > 0 {
				for m.outputCursorCol > 0 && !isWordChar(prevRunes[m.outputCursorCol]) {
					m.outputCursorCol--
				}
				for m.outputCursorCol > 0 && isWordChar(prevRunes[m.outputCursorCol-1]) {
					m.outputCursorCol--
				}
			}
		}
		m.clampOutputCursor()
		m.scrollViewportToCursor()
		return m, nil

	case "B":
		col := m.outputCursorCol
		if col > 0 {
			line := lines[m.outputCursorLine]
			runes := []rune(line)
			if col > len(runes) {
				col = len(runes)
			}
			col--
			for col > 0 && unicode.IsSpace(runes[col]) {
				col--
			}
			if col > 0 {
				for col > 0 && !unicode.IsSpace(runes[col-1]) {
					col--
				}
			}
			m.outputCursorCol = col
		} else if m.outputCursorLine > 0 {
			m.outputCursorLine--
			prevLine := lines[m.outputCursorLine]
			prevRunes := []rune(prevLine)
			m.outputCursorCol = max(0, len(prevRunes)-1)
			if m.outputCursorCol > 0 {
				for m.outputCursorCol > 0 && unicode.IsSpace(prevRunes[m.outputCursorCol]) {
					m.outputCursorCol--
				}
				for m.outputCursorCol > 0 && !unicode.IsSpace(prevRunes[m.outputCursorCol-1]) {
					m.outputCursorCol--
				}
			}
		}
		m.clampOutputCursor()
		m.scrollViewportToCursor()
		return m, nil

	case "e":
		line := lines[m.outputCursorLine]
		runes := []rune(line)
		col := m.outputCursorCol
		if col < len(runes)-1 {
			col++
			for col < len(runes)-1 && !isWordChar(runes[col]) {
				col++
			}
			for col < len(runes)-1 && isWordChar(runes[col+1]) {
				col++
			}
			m.outputCursorCol = col
		} else if m.outputCursorLine < totalLines-1 {
			m.outputCursorLine++
			nextLine := lines[m.outputCursorLine]
			nextRunes := []rune(nextLine)
			col = 0
			for col < len(nextRunes)-1 && !isWordChar(nextRunes[col]) {
				col++
			}
			for col < len(nextRunes)-1 && isWordChar(nextRunes[col+1]) {
				col++
			}
			m.outputCursorCol = col
			m.clampOutputCursor()
		}
		m.scrollViewportToCursor()
		return m, nil

	case "E":
		line := lines[m.outputCursorLine]
		runes := []rune(line)
		col := m.outputCursorCol
		if col < len(runes)-1 {
			col++
			for col < len(runes)-1 && unicode.IsSpace(runes[col]) {
				col++
			}
			for col < len(runes)-1 && !unicode.IsSpace(runes[col+1]) {
				col++
			}
			m.outputCursorCol = col
		} else if m.outputCursorLine < totalLines-1 {
			m.outputCursorLine++
			nextLine := lines[m.outputCursorLine]
			nextRunes := []rune(nextLine)
			col = 0
			for col < len(nextRunes)-1 && unicode.IsSpace(nextRunes[col]) {
				col++
			}
			for col < len(nextRunes)-1 && !unicode.IsSpace(nextRunes[col+1]) {
				col++
			}
			m.outputCursorCol = col
			m.clampOutputCursor()
		}
		m.scrollViewportToCursor()
		return m, nil

	case "0":
		m.outputCursorCol = 0
		return m, nil

	case "$":
		line := lines[m.outputCursorLine]
		lineLen := lipgloss.Width(line)
		if lineLen > 0 {
			m.outputCursorCol = lineLen - 1
		}
		return m, nil

	case "v":
		if m.outputSelectMode == "char" {
			m.outputSelectMode = "none"
		} else {
			m.outputSelectMode = "char"
			m.outputSelectAnchorLine = m.outputCursorLine
			m.outputSelectAnchorCol = m.outputCursorCol
		}
		return m, nil

	case "V":
		if m.outputSelectMode == "line" {
			m.outputSelectMode = "none"
		} else {
			m.outputSelectMode = "line"
			m.outputSelectAnchorLine = m.outputCursorLine
			m.outputSelectAnchorCol = 0
		}
		return m, nil

	case "y":
		if m.outputSelectMode == "char" {
			startLine := min(m.outputSelectAnchorLine, m.outputCursorLine)
			startCol := m.outputSelectAnchorCol
			endLine := max(m.outputSelectAnchorLine, m.outputCursorLine)
			endCol := m.outputCursorCol
			if m.outputSelectAnchorLine > m.outputCursorLine {
				startCol = m.outputCursorCol
				endCol = m.outputSelectAnchorCol
			}
			var selected strings.Builder
			for i := startLine; i <= endLine && i < totalLines; i++ {
				runes := []rune(lines[i])
				fromCol := 0
				toCol := len(runes)
				if i == startLine {
					fromCol = min(startCol, len(runes))
				}
				if i == endLine {
					toCol = min(endCol+1, len(runes))
				}
				for j := fromCol; j < toCol; j++ {
					selected.WriteRune(runes[j])
				}
				if i < endLine {
					selected.WriteRune('\n')
				}
			}
			_ = clipboard.WriteAll(selected.String())
			m.outputSelectMode = "none"
		} else if m.outputSelectMode == "line" {
			startLine := min(m.outputSelectAnchorLine, m.outputCursorLine)
			endLine := max(m.outputSelectAnchorLine, m.outputCursorLine)
			var selected []string
			for i := startLine; i <= endLine && i < totalLines; i++ {
				selected = append(selected, lines[i])
			}
			_ = clipboard.WriteAll(strings.Join(selected, "\n"))
			m.outputSelectMode = "none"
		} else {
			if m.outputCursorLine < totalLines {
				_ = clipboard.WriteAll(lines[m.outputCursorLine])
			}
		}
		return m, nil

	case "Y":
		_ = clipboard.WriteAll(m.output.View())
		return m, nil

	case "f":
		m.jsonPretty = !m.jsonPretty
		m.setOutput(m.outputRaw)
		return m, nil

	case "ctrl+g":
		m.streamFollow = true
		m.output.GotoBottom()
		total := m.output.TotalLineCount()
		if total > 0 {
			m.outputCursorLine = total - 1
			lastLine := lines[m.outputCursorLine]
			lineLen := lipgloss.Width(lastLine)
			if lineLen > 0 {
				m.outputCursorCol = lineLen - 1
			} else {
				m.outputCursorCol = 0
			}
		}
		return m, nil

	case " ", "pgdown":
		m.output.ViewDown()
		m.syncStreamFollowToViewport()
		return m, nil

	case "pgup":
		m.output.ViewUp()
		m.syncStreamFollowToViewport()
		return m, nil

	case "ctrl+d":
		m.output.HalfViewDown()
		m.syncStreamFollowToViewport()
		return m, nil

	case "ctrl+u":
		m.output.HalfViewUp()
		m.syncStreamFollowToViewport()
		return m, nil
	}

	return m, nil
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

func parseHeaderInput(h *RequestHeader) {
	input := h.Inputs.Value()
	if idx := strings.Index(input, ":"); idx != -1 {
		h.Key = strings.TrimSpace(input[:idx])
		h.Value = strings.TrimSpace(input[idx+1:])
	}
}

func parseParamInput(p *RequestParam) {
	input := p.Inputs.Value()
	if idx := strings.Index(input, "="); idx != -1 {
		p.Key = strings.TrimSpace(input[:idx])
		p.Value = strings.TrimSpace(input[idx+1:])
	}
}

func (m *model) scrollToEditingHeader() {
	m.requestSection.headersView.GotoTop()
	m.requestSection.headersView.LineDown(m.requestSection.editingHeader)
}

func (m *model) scrollToEditingParam() {
	m.requestSection.paramsView.GotoTop()
	m.requestSection.paramsView.LineDown(m.requestSection.editingParam)
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
