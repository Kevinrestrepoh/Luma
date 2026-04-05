package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	halfWidth := m.width / 2
	methodWidth := 8
	statusWidth := 30

	if m.width < 80 {
		statusWidth = 26
	}
	streamUI := m.showStreamControls
	stopBlockW := 0
	if streamUI {
		stopBlockW = lipgloss.Width(streamingLiveDot()) + 1 + lipgloss.Width(streamingStopView(false)) + 1
	}
	timeBlockW := 8
	statusBoxW := statusWidth - timeBlockW - stopBlockW
	if statusBoxW < 6 {
		statusBoxW = 6
	}
	status := truncate(m.status, statusBoxW)

	urlWidth := m.width - methodWidth - 4 - statusWidth

	if m.width < 60 {
		urlWidth = m.width - methodWidth - 4
	}

	maxLinesURL := 0
	urlText := m.url.Value()
	if len(urlText) > urlWidth-2 {
		maxLinesURL = (len(urlText) / urlWidth) + 1
	}

	bodyHeight := m.height - 8 - maxLinesURL
	outputHeight := m.height - 5 - maxLinesURL

	m.body.SetWidth(halfWidth - 2)
	m.body.SetHeight(bodyHeight)

	m.output.Height = outputHeight
	innerOutW := halfWidth - 2
	scrollableOut := m.output.TotalLineCount() > outputHeight && innerOutW > 4
	if scrollableOut {
		m.output.Width = innerOutW - 1
	} else {
		m.output.Width = innerOutW
	}

	methodColor := m.methods[m.selectedMethod].Color
	methodView := m.methodStyles.InputField.Width(methodWidth).
		Foreground(methodColor).
		Align(lipgloss.Center).
		Render(m.methods[m.selectedMethod].Name)
	urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
	bodyView := m.requestStyles.InputField.Width(halfWidth - 2).Height(bodyHeight).Render(m.body.View())
	outVP := m.output.View()
	if scrollableOut {
		outVP = lipgloss.JoinHorizontal(lipgloss.Top, outVP, scrollBarView(m.output, outputHeight, m.outputInteractMode))
	}
	outPaneStyle := m.outputStyles.InputField
	if m.outputInteractMode {
		outPaneStyle = outPaneStyle.BorderForeground(TextColor)
	}
	outputView := outPaneStyle.Width(innerOutW).Height(outputHeight).Render(outVP)

	statusText := statusInlineStyle(m.statusCode).Width(statusBoxW).Align(lipgloss.Left).Render(status)
	var statusRow string
	if streamUI {
		rt := m.responseTime
		if lipgloss.Width(rt) > 7 {
			rt = truncate(rt, 7)
		}
		timeCell := lipgloss.NewStyle().
			Width(timeBlockW).
			Align(lipgloss.Right).
			Padding(0).
			Foreground(ResponseTimeColor).
			Render(rt)
		statusRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			statusText,
			streamingLiveDot(),
			streamingStopView(m.focus == "stop"),
			timeCell,
		)
	} else {
		statusRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			statusText,
			lipgloss.NewStyle().Width(timeBlockW).Align(lipgloss.Center).Padding(0).Foreground(ResponseTimeColor).Render(m.responseTime),
		)
	}
	statusAndTime := lipgloss.NewStyle().
		Width(statusWidth).
		Render(statusRow)

	top := lipgloss.JoinHorizontal(
		lipgloss.Center,
		methodView,
		urlView,
		statusAndTime,
	)

	// Request section tabs
	tabWidth := (halfWidth - 2) / len(m.requestSection.tabs)
	tabs := make([]string, len(m.requestSection.tabs))
	for i, tab := range m.requestSection.tabs {
		style := m.requestStyles.InputField.Width(tabWidth - 2).BorderForeground(PrimaryColor)
		if i == m.requestSection.selectedTab {
			style = m.requestStyles.InputField.Width(tabWidth - 2)
		}
		tabs[i] = style.Render(tab)
	}
	tabsView := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Request section content
	var contentView string
	switch m.requestSection.selectedTab {
	case 0: // Body
		contentView = bodyView
	case 1: // Headers
		// Set viewport dimensions
		m.requestSection.headersView.Width = halfWidth - 2
		m.requestSection.headersView.Height = bodyHeight

		// Build headers content
		content := make([]string, len(m.requestSection.headers))
		for i, header := range m.requestSection.headers {
			if i == m.requestSection.editingHeader {
				// Show input field when editing
				content[i] = m.requestStyles.InputField.Width(halfWidth - 4).Render(header.Inputs.View())
			} else {
				// Show key-value pair when not editing
				content[i] = m.requestStyles.InputField.BorderForeground(PrimaryColor).Width(halfWidth - 4).Render(header.Key + ": " + header.Value)
			}
		}

		// Update viewport content
		m.requestSection.headersView.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
		contentView = m.requestStyles.InputField.
			Width(halfWidth - 2).
			Height(bodyHeight).
			Render(m.requestSection.headersView.View())
	case 2: // Params
		// Set viewport dimensions
		m.requestSection.paramsView.Width = halfWidth - 2
		m.requestSection.paramsView.Height = bodyHeight

		// Build params content
		content := make([]string, len(m.requestSection.params))
		for i, param := range m.requestSection.params {
			if i == m.requestSection.editingParam {
				// Show input field when editing
				content[i] = m.requestStyles.InputField.Width(halfWidth - 4).Render(param.Inputs.View())
			} else {
				// Show key-value pair when not editing
				content[i] = m.requestStyles.InputField.BorderForeground(PrimaryColor).Width(halfWidth - 4).Render(param.Key + "=" + param.Value)
			}
		}

		// Update viewport content
		m.requestSection.paramsView.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
		contentView = m.requestStyles.InputField.
			Width(halfWidth - 2).
			Height(bodyHeight).
			Render(m.requestSection.paramsView.View())
	}

	requestSection := lipgloss.JoinVertical(
		lipgloss.Top,
		tabsView,
		contentView,
	)

	if m.width >= 60 {
		return lipgloss.JoinVertical(lipgloss.Top,
			top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				requestSection,
				outputView,
			),
		)
	} else {
		m.body.SetWidth(m.width - 2)
		m.body.SetHeight(m.height/3 - maxLinesURL - 3)

		narrowOutH := m.height/2 - 2 - maxLinesURL/2
		narrowOutW := m.width - 2
		m.output.Height = narrowOutH
		scrollNarrow := m.output.TotalLineCount() > narrowOutH && narrowOutW > 4
		if scrollNarrow {
			m.output.Width = narrowOutW - 1
		} else {
			m.output.Width = narrowOutW
		}

		urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
		bodyView := m.requestStyles.InputField.Width(m.width - 2).Height(m.height/3 - maxLinesURL - 3).Render(m.body.View())
		narrowVP := m.output.View()
		if scrollNarrow {
			narrowVP = lipgloss.JoinHorizontal(lipgloss.Top, narrowVP, scrollBarView(m.output, narrowOutH, m.outputInteractMode))
		}
		narrowOutStyle := m.outputStyles.InputField
		if m.outputInteractMode {
			narrowOutStyle = narrowOutStyle.BorderForeground(TextColor)
		}
		outputView := narrowOutStyle.Width(narrowOutW).Height(narrowOutH).Render(narrowVP)

		nStream := m.showStreamControls
		nStopW := 0
		if nStream {
			nStopW = lipgloss.Width(streamingLiveDot()) + 1 + lipgloss.Width(streamingStopView(false)) + 1
		}
		nTimeW := 8
		nStatusW := m.width - 2 - nTimeW - nStopW
		if nStatusW < 8 {
			nStatusW = 8
		}
		nStat := truncate(m.status, nStatusW)
		nStatusText := statusInlineStyle(m.statusCode).Width(nStatusW).Align(lipgloss.Right).Render(nStat)
		var narrowRow string
		if nStream {
			nrt := m.responseTime
			if lipgloss.Width(nrt) > 7 {
				nrt = truncate(nrt, 7)
			}
			nTimeCell := lipgloss.NewStyle().
				Width(nTimeW).
				Align(lipgloss.Right).
				Padding(0).
				Foreground(ResponseTimeColor).
				Render(nrt)
			narrowRow = lipgloss.JoinHorizontal(
				lipgloss.Center,
				nStatusText,
				streamingLiveDot(),
				streamingStopView(m.focus == "stop"),
				nTimeCell,
			)
		} else {
			narrowRow = lipgloss.JoinHorizontal(
				lipgloss.Center,
				nStatusText,
				lipgloss.NewStyle().Width(nTimeW).Align(lipgloss.Right).Padding(0).Foreground(ResponseTimeColor).Render(m.responseTime),
			)
		}

		statusAndTime := lipgloss.NewStyle().
			Width(m.width - 2).
			Render(narrowRow)

		// Request section tabs for small width
		tabWidth := (m.width - 2) / len(m.requestSection.tabs)
		tabs := make([]string, len(m.requestSection.tabs))
		for i, tab := range m.requestSection.tabs {
			style := m.requestStyles.InputField.Width(tabWidth - 2).BorderForeground(PrimaryColor)
			if i == m.requestSection.selectedTab {
				style = m.requestStyles.InputField.Width(tabWidth - 2)
			}
			tabs[i] = style.Render(tab)
		}
		tabsView := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

		// Request section content for small width
		var contentView string
		switch m.requestSection.selectedTab {
		case 0: // Body
			contentView = bodyView
		case 1: // Headers
			// Set viewport dimensions
			m.requestSection.headersView.Width = m.width - 2
			m.requestSection.headersView.Height = m.height/3 - maxLinesURL - 3

			// Build headers content
			content := make([]string, len(m.requestSection.headers))
			for i, header := range m.requestSection.headers {
				if i == m.requestSection.editingHeader {
					// Show input field when editing
					content[i] = m.requestStyles.InputField.Width(m.width - 4).Render(header.Inputs.View())
				} else {
					// Show key-value pair when not editing
					content[i] = m.requestStyles.InputField.BorderForeground(PrimaryColor).Width(m.width - 4).Render(header.Key + ": " + header.Value)
				}
			}

			// Update viewport content
			m.requestSection.headersView.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
			contentView = m.requestStyles.InputField.
				Width(m.width - 2).
				Height(m.height/3 - maxLinesURL - 3).
				Render(m.requestSection.headersView.View())
		case 2: // Params
			// Set viewport dimensions
			m.requestSection.paramsView.Width = m.width - 2
			m.requestSection.paramsView.Height = m.height/3 - maxLinesURL - 3

			// Build params content
			content := make([]string, len(m.requestSection.params))
			for i, param := range m.requestSection.params {
				if i == m.requestSection.editingParam {
					// Show input field when editing
					content[i] = m.requestStyles.InputField.Width(m.width - 4).Render(param.Inputs.View())
				} else {
					// Show key-value pair when not editing
					content[i] = m.requestStyles.InputField.BorderForeground(PrimaryColor).Width(m.width - 4).Render(param.Key + "=" + param.Value)
				}
			}

			// Update viewport content
			m.requestSection.paramsView.SetContent(lipgloss.JoinVertical(lipgloss.Top, content...))
			contentView = m.requestStyles.InputField.
				Width(m.width - 2).
				Height(m.height/3 - maxLinesURL - 3).
				Render(m.requestSection.paramsView.View())
		}

		requestSection := lipgloss.JoinVertical(
			lipgloss.Top,
			tabsView,
			contentView,
		)

		inputView := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Center, methodView, urlView),
			requestSection,
		)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			inputView,
			statusAndTime,
			outputView,
		)
	}
}

func truncate(text string, max int) string {
	if lipgloss.Width(text) > max {
		return text[:max-3] + "..."
	}
	return text
}

func scrollBarView(vp viewport.Model, height int, interact bool) string {
	track := lipgloss.NewStyle().Foreground(lipgloss.Color("#5c5f77"))
	thumb := lipgloss.NewStyle().Foreground(SecundaryColor)
	if !interact {
		track = track.Faint(true)
		thumb = thumb.Faint(true)
	}
	if height <= 0 {
		return ""
	}
	total := vp.TotalLineCount()
	if total <= height {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = track.Render("▏")
		}
		return strings.Join(lines, "\n")
	}
	thumbH := height * height / total
	if thumbH < 1 {
		thumbH = 1
	}
	maxY := total - height
	pos := 0
	if maxY > 0 {
		pos = vp.YOffset * (height - thumbH) / maxY
	}
	if pos < 0 {
		pos = 0
	}
	if pos+thumbH > height {
		pos = height - thumbH
	}
	lines := make([]string, height)
	for i := range lines {
		if i >= pos && i < pos+thumbH {
			lines[i] = thumb.Render("█")
		} else {
			lines[i] = track.Render("▏")
		}
	}
	return strings.Join(lines, "\n")
}

func streamingLiveDot() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7fdf8a")).
		Padding(0, 1).
		Render("●")
}

func streamingStopView(focused bool) string {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(TextColor).
		Padding(0, 1)
	if focused {
		s = s.BorderForeground(TextColor)
	} else {
		s = s.BorderForeground(SecundaryColor)
	}
	return s.Render("Stop")
}
