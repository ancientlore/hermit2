package scroller

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Model implements scrolling behavior and is intended to be embedded in other models.
type Model struct {
	HeaderLines int       // Number of header lines
	FooterLines int       // Number of footer lines
	Lines       int       // The number of lines to scroll over
	Prev        tea.Model // The previous model (for going back)
	cursor      int       // Current position of cursor
	offset      int       // The offset of the view (enables scrolling)
	Width       int       // The width of the current view
	Height      int       // The height of the current view
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

// Offset returns the current offset in view.
func (m Model) Offset() int {
	return m.offset
}

// VisibleLines returns the number of lines in the scolling view,
// excluding the header and footer.
func (m Model) VisibleLines() int {
	visible := m.Height - m.HeaderLines - m.FooterLines - 1
	if visible < 0 {
		visible = 0
	}
	return visible
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update handles messages in order to implement scrolling.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch {

		// The "up" keys move the cursor up
		case key.Matches(msg, DefaultKeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" keys move the cursor down
		case key.Matches(msg, DefaultKeyMap.Down):
			if m.cursor < m.Lines-1 {
				m.cursor++
			}

		case key.Matches(msg, DefaultKeyMap.Left):
			if m.Prev != nil {
				return m.Prev, nil
			}

		case key.Matches(msg, DefaultKeyMap.Home):
			m.cursor = 0

		case key.Matches(msg, DefaultKeyMap.End):
			m.cursor = m.Lines - 1

		case key.Matches(msg, DefaultKeyMap.PageUp):
			m.cursor -= m.Height - (2 + m.HeaderLines + m.FooterLines)
			if m.cursor < 0 {
				m.cursor = 0
			}

		case key.Matches(msg, DefaultKeyMap.PageDown):
			m.cursor += m.Height - (2 + m.HeaderLines + m.FooterLines)
			if m.cursor >= m.Lines {
				m.cursor = m.Lines - 1
			}

		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	}

	m.fixOffset()

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// View renders the contents of the scroller. It is intended to be implemented
// by the class embedding the scroller.
func (m Model) View() string {
	return ""
}

func (m *Model) fixOffset() {
	// Fix cursor location
	if m.cursor > m.Lines-1 {
		m.cursor = m.Lines - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// cursor before offset - offset needs to be decreased
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// cursor greater than offset + window size - offset needs to be increased
	if m.cursor-m.offset+1 > m.Height-(m.HeaderLines+m.FooterLines+1) {
		m.offset = m.cursor + 1 - m.Height + (m.HeaderLines + m.FooterLines + 1)
	}
	if m.offset < 0 {
		m.offset = 0
	}
}
