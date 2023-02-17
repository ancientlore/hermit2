package scroller

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	header    = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	footer    = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	normal    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	highlight = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
)

// Model implements scrolling behavior over a Viewer.
type Model[T Viewer] struct {
	Header string    // Header text
	Data   T         // The view we are using
	Prev   tea.Model // The previous model (for going back)
	cursor int       // Current position of cursor
	offset int       // The offset of the view (enables scrolling)
	width  int       // The width of the current view
	height int       // The height of the current view
}

// Init initializes the model.
func (m Model[T]) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update handles messages in order to implement scrolling.
func (m Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch {

		// The "up" keys move the cursor up
		case key.Matches(msg, DefaultKeyMap.Up):
			m.cursor--
			m.fixOffset()

		// The "down" keys move the cursor down
		case key.Matches(msg, DefaultKeyMap.Down):
			m.cursor++
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.Left):
			if m.Prev != nil {
				return m.Prev, func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height + 2} }
			}

		case key.Matches(msg, DefaultKeyMap.Home):
			m.cursor = 0
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.End):
			m.cursor = m.Data.Len() - 1
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.PageUp):
			m.cursor -= m.height - 1
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.PageDown):
			m.cursor += m.height - 1
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2 // account for header and footer
		m.fixOffset()
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// View renders the contents of the scroller. It is intended to be implemented
// by the class embedding the scroller.
func (m Model[T]) View() string {
	// Header
	s := header.Width(m.width).Height(1).Render(m.Header) + "\n"

	// Viewport
	lines := 0
	for i := m.offset; i < m.Data.Len() && i < m.height+m.offset; i++ {
		style := normal.Width(m.width).Height(1).MaxWidth(m.width)
		if m.cursor == i {
			style = highlight.Width(m.width).Height(1).MaxWidth(m.width)
		}
		line := m.Data.Render(i, m.width, style)
		_, h := lipgloss.Size(line)

		/*
			if lines+h > m.height {
				if m.cursor > m.offset+lines {
					m.cursor = m.offset + lines
				}
				break
			}
		*/

		lines += h
		s += line + "\n"
	}
	//repeat := m.height - m.Data.Len()
	repeat := m.height - lines
	if repeat > 0 {
		s += strings.Repeat("\n", repeat)
	}

	// Footer
	s += footer.Width(m.width).Height(1).Render(fmt.Sprintf("%d / %d    lines=%d", m.cursor+1, m.Data.Len(), lines))
	return s
}

// Cursor returns the position of the cursor.
func (m Model[T]) Cursor() int {
	return m.cursor
}

// BumpCursor moves the cursor down, like after a select operation.
func (m *Model[T]) BumpCursor() {
	m.cursor++
	m.fixOffset()
}

// Width returns the width of the view.
func (m Model[T]) Width() int {
	return m.width
}

// Height returns the height of the view.
func (m Model[T]) Height() int {
	return m.height + 2
}

// fixOffset fixes the cursor and offset locations to be consistent
// with the requested changes. Changes could be the height, width,
// or cursor position.
func (m *Model[T]) fixOffset() {
	// Fix cursor location
	if m.cursor > m.Data.Len()-1 {
		m.cursor = m.Data.Len() - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// cursor before offset - offset needs to be decreased
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// cursor greater than offset + window size - offset needs to be increased
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}
