package main

import "github.com/charmbracelet/lipgloss"

func (m *model) View() string {
	halfWidth := m.width / 2
	methodWidth := 8
	urlWidth := halfWidth - methodWidth - 4
	outputHeight := m.height - 2

	if m.width < 50 {
		urlWidth = m.width - methodWidth - 4
	}

	maxLinesURL := 0
	urlText := m.url.Value()
	if len(urlText) > urlWidth-2 {
		maxLinesURL = (len(urlText) / urlWidth) + 1
	}

	bodyHeight := m.height - 5 - maxLinesURL

	m.body.SetWidth(halfWidth - 2)
	m.body.SetHeight(bodyHeight)

	m.output.Width = halfWidth - 2
	m.output.Height = outputHeight

	methodColor := m.methods[m.selectedMethod].Color
	methodView := m.methodStyles.InputField.Width(methodWidth).Foreground(methodColor).Render(m.methods[m.selectedMethod].Name)
	urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
	bodyView := m.bodyStyles.InputField.Width(halfWidth - 2).Height(bodyHeight).Render(m.body.View())
	outputView := m.outputStyles.InputField.Width(halfWidth - 2).Height(outputHeight).Render(m.output.View())

	inputView := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, methodView, urlView),
		bodyView,
	)

	if m.width >= 50 {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(halfWidth).Render(inputView),
			lipgloss.NewStyle().Width(halfWidth).Render(outputView),
		)
	} else {
		m.body.SetWidth(m.width - 2)
		m.body.SetHeight(m.height/3 - maxLinesURL)

		m.output.Width = m.width - 2
		m.output.Height = m.height / 2

		urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
		bodyView := m.bodyStyles.InputField.Width(m.width - 2).Height(m.height/3 - maxLinesURL).Render(m.body.View())
		outputView := m.outputStyles.InputField.Width(m.width - 2).Height(m.height / 2).Render(m.output.View())

		inputView := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top, methodView, urlView),
			bodyView,
		)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			inputView,
			outputView,
		)
	}
}
