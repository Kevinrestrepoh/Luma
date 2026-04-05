package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	ProgramSend = p.Send
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
