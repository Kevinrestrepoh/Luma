package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Type definitions
type TextState struct {
	content []string
	cursor  struct {
		line   int
		column int
	}
}

type CustomTextarea struct {
	width   int
	height  int
	content []string
	cursor  struct {
		line   int
		column int
	}
	focused      bool
	style        lipgloss.Style
	undoStack    []TextState
	redoStack    []TextState
	lastOp       string // Track last operation to handle consecutive operations
	scrollOffset int
}

func NewCustomTextarea() CustomTextarea {
	initialContent := []string{""}
	initialState := TextState{
		content: initialContent,
		cursor: struct {
			line   int
			column int
		}{0, 0},
	}

	return CustomTextarea{
		content: initialContent,
		cursor: struct {
			line   int
			column int
		}{0, 0},
		width:        80, // Default width
		height:       20, // Default height
		style:        lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#E4E4E4f")),
		undoStack:    []TextState{initialState},
		redoStack:    []TextState{},
		lastOp:       "",
		scrollOffset: 0,
	}
}

func (t *CustomTextarea) SetWidth(w int) {
	if w <= 0 {
		w = 80 // Default width
	}
	t.width = w
}

func (t *CustomTextarea) SetHeight(h int) {
	if h <= 0 {
		h = 20 // Default height
	}
	oldHeight := t.height
	t.height = h

	// Adjust scroll position when height changes
	if h < oldHeight {
		// If height decreased, ensure cursor is still visible
		if t.cursor.line < t.scrollOffset {
			t.scrollOffset = t.cursor.line
		} else if t.cursor.line >= t.scrollOffset+h {
			t.scrollOffset = max(0, t.cursor.line-h+1)
		}
	} else {
		// If height increased, try to show more content
		if t.scrollOffset > 0 && len(t.content) <= t.scrollOffset+h {
			t.scrollOffset = max(0, len(t.content)-h)
		}
	}
}

func (t *CustomTextarea) Focus() {
	t.focused = true
}

func (t *CustomTextarea) Blur() {
	t.focused = false
}

func (t *CustomTextarea) Value() string {
	return strings.Join(t.content, "\n")
}

func (t *CustomTextarea) View() string {
	if len(t.content) == 0 {
		t.content = []string{""}
	}

	// Ensure cursor is within bounds
	if t.cursor.line >= len(t.content) {
		t.cursor.line = len(t.content) - 1
	}
	if t.cursor.column > len(t.content[t.cursor.line]) {
		t.cursor.column = len(t.content[t.cursor.line])
	}

	// Calculate visible range
	visibleStart := t.scrollOffset
	visibleEnd := min(len(t.content), visibleStart+t.height)

	// Adjust scroll position to keep cursor visible
	if t.cursor.line < visibleStart {
		t.scrollOffset = t.cursor.line
	} else if t.cursor.line >= visibleEnd {
		t.scrollOffset = max(0, t.cursor.line-t.height+1)
	}

	// Get visible lines
	visibleLines := make([]string, 0, t.height)
	for i := visibleStart; i < visibleEnd; i++ {
		line := t.content[i]
		// Ensure line doesn't exceed width
		if len(line) > t.width {
			line = line[:t.width]
		}
		if i == t.cursor.line && t.focused {
			// Insert cursor with highlighted character only when focused
			if t.cursor.column < len(line) {
				before := line[:t.cursor.column]
				char := line[t.cursor.column : t.cursor.column+1]
				after := line[t.cursor.column+1:]
				line = before + t.style.Render(char) + after
			} else {
				// If at end of line, show a space with cursor style
				line = line + t.style.Render(" ")
			}
		}
		visibleLines = append(visibleLines, line)
	}

	// If we have fewer lines than the height, pad with empty lines
	for len(visibleLines) < t.height {
		visibleLines = append(visibleLines, "")
	}

	// Ensure we don't exceed height
	if len(visibleLines) > t.height {
		visibleLines = visibleLines[:t.height]
	}

	return strings.Join(visibleLines, "\n")
}

func (t *CustomTextarea) Update(msg tea.Msg) (CustomTextarea, tea.Cmd) {
	if !t.focused {
		return *t, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Handle undo
		if key == "ctrl+z" {
			if len(t.undoStack) > 1 { // Keep at least one state in undo stack
				// Save current state to redo stack
				t.pushRedo(t.getCurrentState())

				// Remove current state from undo stack
				t.undoStack = t.undoStack[:len(t.undoStack)-1]

				// Restore previous state
				prevState := t.undoStack[len(t.undoStack)-1]
				t.restoreState(prevState)
				t.lastOp = "undo"
			}
			return *t, nil
		}

		// Handle redo
		if key == "ctrl+y" {
			if len(t.redoStack) > 0 {
				// Save current state to undo stack
				t.pushUndo(t.getCurrentState())

				// Get state from redo stack
				redoState := t.redoStack[len(t.redoStack)-1]
				t.redoStack = t.redoStack[:len(t.redoStack)-1]

				// Restore redo state
				t.restoreState(redoState)
				t.lastOp = "redo"
			}
			return *t, nil
		}

		// For other operations, continue with normal processing
		switch key {
		case "tab", "ctrl+t":
			// Insert tabulation
			line := t.content[t.cursor.line]
			before := line[:t.cursor.column]
			after := line[t.cursor.column:]

			// Check if adding tab would exceed width
			if len(before)+4+len(after) > t.width {
				// If we're at the start of the line, just add spaces up to width
				if t.cursor.column == 0 {
					spaces := min(4, t.width)
					t.saveState() // Save state before modification
					t.content[t.cursor.line] = strings.Repeat(" ", spaces) + after
					t.cursor.column = spaces
				} else {
					// Otherwise, wrap the text
					remainingWidth := t.width - len(before)
					if remainingWidth > 0 {
						spaces := min(4, remainingWidth)
						t.saveState() // Save state before modification
						t.content[t.cursor.line] = before + strings.Repeat(" ", spaces)
						t.content = append(t.content[:t.cursor.line+1], append([]string{after}, t.content[t.cursor.line+1:]...)...)
						t.cursor.column += spaces
					}
				}
			} else {
				t.saveState() // Save state before modification
				t.content[t.cursor.line] = before + "    " + after
				t.cursor.column += 4
			}
			t.lastOp = "tab"

		case "shift+tab", "ctrl+d":
			// Unindent
			line := t.content[t.cursor.line]
			if len(line) > 0 {
				indent := 0
				// Count leading spaces, up to 4
				for i := 0; i < len(line) && i < 4; i++ {
					if line[i] == ' ' {
						indent++
					} else {
						break
					}
				}

				if indent > 0 {
					t.saveState() // Save state before modification
					// Remove the indentation
					t.content[t.cursor.line] = line[indent:]

					// Adjust cursor position
					if t.cursor.column > indent {
						t.cursor.column -= indent
					} else {
						t.cursor.column = 0
					}
					t.lastOp = "unindent"
				}
			}

		case "ctrl+b":
			// Move cursor back one word
			if t.cursor.line > 0 || t.cursor.column > 0 {
				line := t.content[t.cursor.line]
				if t.cursor.column > 0 {
					// Skip spaces
					for t.cursor.column > 0 && line[t.cursor.column-1] == ' ' {
						t.cursor.column--
					}
					// Skip word
					for t.cursor.column > 0 && line[t.cursor.column-1] != ' ' {
						t.cursor.column--
					}
				} else {
					t.cursor.line--
					t.cursor.column = len(t.content[t.cursor.line])
				}
			}
			// No state saved for cursor movement

		case "ctrl+f":
			// Move cursor forward one word
			line := t.content[t.cursor.line]
			if t.cursor.column < len(line) {
				// Skip spaces
				for t.cursor.column < len(line) && line[t.cursor.column] == ' ' {
					t.cursor.column++
				}
				// Skip word
				for t.cursor.column < len(line) && line[t.cursor.column] != ' ' {
					t.cursor.column++
				}
			} else if t.cursor.line < len(t.content)-1 {
				t.cursor.line++
				t.cursor.column = 0
			}
			// No state saved for cursor movement

		case "enter":
			// Split line at cursor
			line := t.content[t.cursor.line]
			before := line[:t.cursor.column]
			after := line[t.cursor.column:]

			// Calculate indentation for the new line
			indent := ""
			for _, char := range before {
				if char == ' ' {
					indent += " "
				} else {
					break
				}
			}

			t.saveState() // Save state before modification
			t.content[t.cursor.line] = before
			t.content = append(t.content[:t.cursor.line+1], append([]string{indent + after}, t.content[t.cursor.line+1:]...)...)
			t.cursor.line++
			t.cursor.column = len(indent)
			t.lastOp = "enter"

		case "backspace":
			if t.cursor.column > 0 {
				// Delete character before cursor
				t.saveState() // Save state before modification
				line := t.content[t.cursor.line]
				t.content[t.cursor.line] = line[:t.cursor.column-1] + line[t.cursor.column:]
				t.cursor.column--
				t.lastOp = "delete"
			} else if t.cursor.line > 0 {
				// Merge with previous line
				t.saveState() // Save state before modification
				prevLine := t.content[t.cursor.line-1]
				t.content[t.cursor.line-1] = prevLine + t.content[t.cursor.line]
				t.content = append(t.content[:t.cursor.line], t.content[t.cursor.line+1:]...)
				t.cursor.line--
				t.cursor.column = len(prevLine)
				t.lastOp = "delete"
			}

		case "delete":
			if t.cursor.column < len(t.content[t.cursor.line]) {
				// Delete character after cursor
				t.saveState() // Save state before modification
				line := t.content[t.cursor.line]
				t.content[t.cursor.line] = line[:t.cursor.column] + line[t.cursor.column+1:]
				t.lastOp = "delete"
			} else if t.cursor.line < len(t.content)-1 {
				// Merge with next line
				t.saveState() // Save state before modification
				t.content[t.cursor.line] += t.content[t.cursor.line+1]
				t.content = append(t.content[:t.cursor.line+1], t.content[t.cursor.line+2:]...)
				t.lastOp = "delete"
			}

		case "left", "right", "up", "down":
			// Handle cursor movement
			switch key {
			case "left":
				if t.cursor.column > 0 {
					t.cursor.column--
				} else if t.cursor.line > 0 {
					t.cursor.line--
					t.cursor.column = len(t.content[t.cursor.line])
				}
			case "right":
				if t.cursor.column < len(t.content[t.cursor.line]) {
					t.cursor.column++
				} else if t.cursor.line < len(t.content)-1 {
					t.cursor.line++
					t.cursor.column = 0
				}
			case "up":
				if t.cursor.line > 0 {
					t.cursor.line--
					if t.cursor.column > len(t.content[t.cursor.line]) {
						t.cursor.column = len(t.content[t.cursor.line])
					}
				}
			case "down":
				if t.cursor.line < len(t.content)-1 {
					t.cursor.line++
					if t.cursor.column > len(t.content[t.cursor.line]) {
						t.cursor.column = len(t.content[t.cursor.line])
					}
				}
			}
			// No state saved for cursor movement

		case "\"":
			// Auto-complete quotes
			t.saveState() // Save state before modification
			line := t.content[t.cursor.line]
			before := line[:t.cursor.column]
			after := line[t.cursor.column:]
			t.content[t.cursor.line] = before + "\"" + "\"" + after
			t.cursor.column++
			t.lastOp = "insert"

		case "'":
			// Auto-complete single quotation mark
			t.saveState() // Save state before modification
			line := t.content[t.cursor.line]
			before := line[:t.cursor.column]
			after := line[t.cursor.column:]
			t.content[t.cursor.line] = before + "'" + "'" + after
			t.cursor.column++
			t.lastOp = "insert"

		case "{":
			// Auto-complete braces with proper formatting
			t.saveState() // Save state before modification
			line := t.content[t.cursor.line]
			before := line[:t.cursor.column]
			after := line[t.cursor.column:]

			// Calculate indentation for the new lines
			indent := ""
			for _, char := range before {
				if char == ' ' {
					indent += " "
				} else {
					break
				}
			}

			// Check if we have enough space for the new lines
			if len(t.content) >= t.height-2 {
				return *t, nil // Don't add more lines if we're at the height limit
			}

			// Insert opening brace
			t.content[t.cursor.line] = before + "{" + after

			// Insert empty line with proper indentation
			newIndent := indent + "    "
			if len(newIndent) > t.width {
				newIndent = strings.Repeat(" ", t.width)
			}
			t.content = append(t.content[:t.cursor.line+1], append([]string{newIndent}, t.content[t.cursor.line+1:]...)...)

			// Insert closing brace with proper indentation
			if len(indent) > t.width {
				indent = strings.Repeat(" ", t.width)
			}
			t.content = append(t.content[:t.cursor.line+2], append([]string{indent + "}"}, t.content[t.cursor.line+2:]...)...)

			// Move cursor to the empty line between braces
			t.cursor.line++
			t.cursor.column = len(newIndent)
			t.lastOp = "insert"

		case "[":
			// Auto-complete bracket with proper formatting
			t.saveState() // Save state before modification
			line := t.content[t.cursor.line]
			before := line[:t.cursor.column]
			after := line[t.cursor.column:]

			// Calculate indentation for the new lines
			indent := ""
			for _, char := range before {
				if char == ' ' {
					indent += " "
				} else {
					break
				}
			}

			// Check if we have enough space for the new lines
			if len(t.content) >= t.height-2 {
				return *t, nil // Don't add more lines if we're at the height limit
			}

			// Insert opening bracket
			t.content[t.cursor.line] = before + "[" + after

			// Insert empty line with proper indentation
			newIndent := indent + "    "
			if len(newIndent) > t.width {
				newIndent = strings.Repeat(" ", t.width)
			}
			t.content = append(t.content[:t.cursor.line+1], append([]string{newIndent}, t.content[t.cursor.line+1:]...)...)

			// Insert closing bracket with proper indentation
			if len(indent) > t.width {
				indent = strings.Repeat(" ", t.width)
			}
			t.content = append(t.content[:t.cursor.line+2], append([]string{indent + "]"}, t.content[t.cursor.line+2:]...)...)

			// Move cursor to the empty line between brackets
			t.cursor.line++
			t.cursor.column = len(newIndent)
			t.lastOp = "insert"

		default:
			// Insert character
			if len(key) == 1 {
				t.saveState() // Save state before modification
				line := t.content[t.cursor.line]
				before := line[:t.cursor.column]
				after := line[t.cursor.column:]

				// Check if adding character would exceed width
				if len(before)+1+len(after) > t.width {
					// If we're at the start of the line, just add the character
					if t.cursor.column == 0 {
						t.content[t.cursor.line] = key + after
						t.cursor.column = 1
					} else {
						// Otherwise, wrap the text
						t.content[t.cursor.line] = before
						t.content = append(t.content[:t.cursor.line+1], append([]string{key + after}, t.content[t.cursor.line+1:]...)...)
						t.cursor.line++
						t.cursor.column = 1
					}
				} else {
					t.content[t.cursor.line] = before + key + after
					t.cursor.column++
				}
				t.lastOp = "insert"
			}
		}
	}

	return *t, nil
}

// Helper functions for undo/redo
func (t *CustomTextarea) getCurrentState() TextState {
	// Create a deep copy of the current content
	contentCopy := make([]string, len(t.content))
	copy(contentCopy, t.content)

	return TextState{
		content: contentCopy,
		cursor: struct {
			line   int
			column int
		}{
			line:   t.cursor.line,
			column: t.cursor.column,
		},
	}
}

func (t *CustomTextarea) restoreState(state TextState) {
	// Restore content
	t.content = make([]string, len(state.content))
	copy(t.content, state.content)

	// Restore cursor
	t.cursor.line = state.cursor.line
	t.cursor.column = state.cursor.column

	// Ensure cursor is within bounds
	if t.cursor.line >= len(t.content) {
		t.cursor.line = len(t.content) - 1
	}
	if t.cursor.column > len(t.content[t.cursor.line]) {
		t.cursor.column = len(t.content[t.cursor.line])
	}
}

func (t *CustomTextarea) pushUndo(state TextState) {
	t.undoStack = append(t.undoStack, state)
	// Keep undo stack at reasonable size
	if len(t.undoStack) > 100 {
		t.undoStack = t.undoStack[1:]
	}
}

func (t *CustomTextarea) pushRedo(state TextState) {
	t.redoStack = append(t.redoStack, state)
	// Keep redo stack at reasonable size
	if len(t.redoStack) > 100 {
		t.redoStack = t.redoStack[1:]
	}
}

func (t *CustomTextarea) saveState() {
	// Don't save state for consecutive operations of the same type
	if t.lastOp != "" {
		// Save the current state to the undo stack
		t.pushUndo(t.getCurrentState())
		// Clear redo stack when making a new change
		t.redoStack = nil
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to ensure content stays within bounds
func (t *CustomTextarea) ensureBounds() {
	// Ensure content doesn't exceed height
	if len(t.content) > t.height {
		t.content = t.content[:t.height]
		if t.cursor.line >= t.height {
			t.cursor.line = t.height - 1
		}
	}

	// Ensure each line doesn't exceed width
	for i, line := range t.content {
		if len(line) > t.width {
			t.content[i] = line[:t.width]
			if i == t.cursor.line && t.cursor.column > t.width {
				t.cursor.column = t.width
			}
		}
	}
}
