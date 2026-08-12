package overlay

import tea "github.com/charmbracelet/bubbletea"

type Position int

const (
	Top Position = iota + 1
	Right
	Bottom
	Left
	Center
)

type Viewable interface {
	View() string
}

type Model struct {
	Foreground Viewable
	Background Viewable
	XPosition  Position
	YPosition  Position
	XOffset    int
	YOffset    int
}

func New(fore Viewable, back Viewable, xPos Position, yPos Position, xOff int, yOff int) *Model {
	return &Model{
		Foreground: fore,
		Background: back,
		XPosition:  xPos,
		YPosition:  yPos,
		XOffset:    xOff,
		YOffset:    yOff,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *Model) View() string {
	if m.Foreground == nil && m.Background == nil {
		return ""
	}
	if m.Foreground == nil {
		return m.Background.View()
	}
	if m.Background == nil {
		return m.Foreground.View()
	}
	return Composite(m.Foreground.View(), m.Background.View(), m.XPosition, m.YPosition, m.XOffset, m.YOffset)
}
