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

	m.output.Width = halfWidth - 2
	m.output.Height = outputHeight

	methodColor := m.methods[m.selectedMethod].Color
	methodView := m.methodStyles.InputField.Width(methodWidth).
		Foreground(methodColor).
		Align(lipgloss.Center).
		Render(m.methods[m.selectedMethod].Name)
	urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
	bodyView := m.requestStyles.InputField.Width(halfWidth - 2).Height(bodyHeight).Render(m.body.View())
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

		m.output.Width = m.width - 2
		m.output.Height = m.height/2 - 2 - maxLinesURL/2

		urlView := m.urlStyles.InputField.Width(urlWidth).Render(m.url.View())
		bodyView := m.requestStyles.InputField.Width(m.width - 2).Height(m.height/3 - maxLinesURL - 3).Render(m.body.View())
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
			lipgloss.JoinHorizontal(lipgloss.Top, methodView, urlView),
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
