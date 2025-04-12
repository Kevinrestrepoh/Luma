package main

import "github.com/charmbracelet/lipgloss"

func (m *model) View() string {
	halfWidth := m.width / 2
	methodWidth := 8
	statusWidth := 30

	status := truncate(m.status, 30)
	if m.width < 80 {
		statusWidth = 26
		status = truncate(status, 15)
	}

	urlWidth := m.width - methodWidth - 4 - statusWidth

	if m.width < 50 {
		urlWidth = m.width - methodWidth - 4
	}

	maxLinesURL := 0
	urlText := m.url.Value()
	if len(urlText) > urlWidth-2 {
		maxLinesURL = (len(urlText) / urlWidth) + 1
	}

	bodyHeight := m.height - 5 - maxLinesURL
	outputHeight := m.height - 5 - maxLinesURL

	m.body.SetWidth(halfWidth - 2)
	m.body.SetHeight(bodyHeight)

	m.output.Width = halfWidth - 2
	m.output.Height = outputHeight

	methodColor := m.methods[m.selectedMethod].Color
	methodView := m.methodStyles.InputField.Width(methodWidth).
		Foreground(methodColor).
		Align(lipgloss.Center).
		Render(m.methods[m.selectedMethod].Name)
	urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
	bodyView := m.bodyStyles.InputField.Width(halfWidth - 2).Height(bodyHeight).Render(m.body.View())
	outputView := m.outputStyles.InputField.Width(halfWidth - 2).Height(outputHeight).Render(m.output.View())

	statusAndTime := lipgloss.NewStyle().
		Width(statusWidth).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				StatusStyle(m.statusCode).Width(statusWidth-8).PaddingBottom(0).Align(lipgloss.Center).Render(status),
				lipgloss.NewStyle().
					Width(8).
					Align(lipgloss.Center).
					Padding(1).
					PaddingBottom(0).
					Foreground(ResponseTimeColor).
					Render(m.responseTime),
			),
		)

	top := lipgloss.JoinHorizontal(
		lipgloss.Top,
		methodView,
		urlView,
		statusAndTime,
	)

	if m.width >= 50 {
		return lipgloss.JoinVertical(lipgloss.Top,
			top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				bodyView,
				outputView,
			),
		)
	} else {
		m.body.SetWidth(m.width - 2)
		m.body.SetHeight(m.height/3 - maxLinesURL)

		m.output.Width = m.width - 2
		m.output.Height = m.height/2 - 2 - maxLinesURL/2

		urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
		bodyView := m.bodyStyles.InputField.Width(m.width - 2).Height(m.height/3 - maxLinesURL).Render(m.body.View())
		outputView := m.outputStyles.InputField.Width(m.width - 2).Height(m.height/2 - 2 - maxLinesURL/2).Render(m.output.View())

		statusAndTime := lipgloss.NewStyle().
			Width(m.width - 2).
			Render(
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					StatusStyle(m.statusCode).Width(m.width-10).Padding(0).Align(lipgloss.Right).Render(m.status),
					lipgloss.NewStyle().
						Width(8).
						Align(lipgloss.Right).
						Padding(0).
						Foreground(ResponseTimeColor).
						Render(m.responseTime),
				),
			)

		inputView := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top, methodView, urlView),
			bodyView,
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
