package main

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) Init() tea.Cmd {
	m.UpdateStyles()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m.handleKeyMsg(msg)
	}

	return m.handleInsertModeMsg(msg)
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
