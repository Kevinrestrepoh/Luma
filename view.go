package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/Kevinrestrepoh/Luma/overlay"
)

type backgroundView struct {
	view string
}

func (b backgroundView) View() string {
	return b.view
}

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
		statusRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			statusText,
			streamingLiveDot(),
			streamingStopView(m.focus == "stop"),
			renderTimeCell(timeBlockW, rt),
		)
	} else {
		statusRow = lipgloss.JoinHorizontal(
			lipgloss.Center,
			statusText,
			renderTimeCell(timeBlockW, m.responseTime),
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

	tabWidth := (halfWidth - 2) / len(m.requestSection.tabs)
	tabsView := m.renderTabs(tabWidth)

	var contentView string
	switch m.requestSection.selectedTab {
	case 0:
		contentView = bodyView
	case 1:
		contentView = m.renderHeadersContent(halfWidth-2, bodyHeight)
	case 2:
		contentView = m.renderParamsContent(halfWidth-2, bodyHeight)
	}

	requestSection := lipgloss.JoinVertical(
		lipgloss.Top,
		tabsView,
		contentView,
	)

	if m.width >= 60 {
		baseView := lipgloss.JoinVertical(lipgloss.Top,
			top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				requestSection,
				outputView,
			),
		)
		if m.showModal {
			modal := NewModal(m.windows, m.modalSelected, m.width, m.height, m.methods)
			return overlay.Composite(modal.View(), baseView, overlay.Center, overlay.Center, 0, 0)
		}
		return baseView
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
			narrowRow = lipgloss.JoinHorizontal(
				lipgloss.Center,
				nStatusText,
				streamingLiveDot(),
				streamingStopView(m.focus == "stop"),
				renderTimeCell(nTimeW, nrt),
			)
		} else {
			narrowRow = lipgloss.JoinHorizontal(
				lipgloss.Center,
				nStatusText,
				renderTimeCell(nTimeW, m.responseTime),
			)
		}

		statusAndTime := lipgloss.NewStyle().
			Width(m.width - 2).
			Render(narrowRow)

		tabWidth := (m.width - 2) / len(m.requestSection.tabs)
		tabsView := m.renderTabs(tabWidth)

		narrowContentH := m.height/3 - maxLinesURL - 3
		var narrowContentView string
		switch m.requestSection.selectedTab {
		case 0:
			narrowContentView = bodyView
		case 1:
			narrowContentView = m.renderHeadersContent(m.width-2, narrowContentH)
		case 2:
			narrowContentView = m.renderParamsContent(m.width-2, narrowContentH)
		}

		narrowRequestSection := lipgloss.JoinVertical(
			lipgloss.Top,
			tabsView,
			narrowContentView,
		)

		inputView := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Center, methodView, urlView),
			narrowRequestSection,
		)

		baseView := lipgloss.JoinVertical(
			lipgloss.Left,
			inputView,
			statusAndTime,
			outputView,
		)

		if m.showModal {
			modal := NewModal(m.windows, m.modalSelected, m.width, m.height, m.methods)
			return overlay.Composite(modal.View(), baseView, overlay.Center, overlay.Center, 0, 0)
		}
		return baseView
	}
}
