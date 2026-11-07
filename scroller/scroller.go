package scroller

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	header    = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	footer    = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	normal    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	highlight = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
	ellipsis  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Background(lipgloss.Color("#888B7E"))
)

// Model implements scrolling behavior over a Viewer.
type Model[T Viewer] struct {
	Header         string         // Header text
	Data           T              // The view we are using
	Prev           tea.Model      // The previous model (for going back)
	cursor         int            // Current position of cursor
	offset         int            // The offset of the view (enables scrolling)
	width          int            // The width of the current view
	height         int            // The height of the current view
	normalStyle    lipgloss.Style // Cached style for normal lines
	highlightStyle lipgloss.Style // Cached style for highlighted line
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
	case tea.KeyPressMsg:

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
				m.Data.Close()
				return m.Prev, func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height + 2} }
			}

		case key.Matches(msg, DefaultKeyMap.Home):
			m.cursor = 0
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.End):
			m.cursor = m.Data.Len(m.width) - 1
			m.fixOffset()

		case key.Matches(msg, DefaultKeyMap.PageUp):
			m.pageUp()

		case key.Matches(msg, DefaultKeyMap.PageDown):
			m.pageDown()

		case key.Matches(msg, DefaultKeyMap.Quit):
			m.Data.Close()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2 // account for header and footer
		m.normalStyle = normal.Width(m.width).Height(1).MaxWidth(m.width)
		m.highlightStyle = highlight.Width(m.width).Height(1).MaxWidth(m.width)
		m.fixOffset()
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// View renders the contents of the scroller. It is intended to be implemented
// by the class embedding the scroller.
func (m Model[T]) View() tea.View {
	// Header
	hdr := m.Header
	if m.width > 3 && lipgloss.Width(hdr) > m.width {
		r := []rune(hdr)
		if len(r) > m.width-3 {
			r = r[len(r)-(m.width-3):]
		}
		for len(r) > 0 && lipgloss.Width(string(r))+3 > m.width {
			r = r[1:]
		}
		hdr = ellipsis.Render("...") + header.Render(string(r))
	}
	s := header.Width(m.width).Height(1).MaxWidth(m.width).Render(hdr) + "\n"

	// Viewport
	lines := 0
	cursorEnd := 0
	for i := m.offset; i < m.Data.Len(m.width) && lines < m.height; i++ {
		style := m.normalStyle
		if m.cursor == i {
			style = m.highlightStyle
		}
		line := m.Data.Render(i, m.width, style)
		_, h := lipgloss.Size(line)
		if m.cursor == i {
			cursorEnd = lines + h
		}
		lines += h
		s += line + "\n"
	}

	// TODO(michael.lore): Should make this more efficient at some point.
	repeat := m.height - lines
	if repeat > 0 {
		// Too few lines
		s += strings.Repeat("\n", repeat)
	} else if repeat < 0 {
		// Too many lines (due to text wrap)
		// a[0] is the header
		a := strings.Split(s, "\n")

		if cursorEnd > m.height {
			// take off from the front
			dropTop := cursorEnd - m.height
			s = strings.Join(append(a[0:1], a[1+dropTop:1+dropTop+m.height]...), "\n") + "\n"
		} else {
			// take from back
			s = strings.Join(a[0:1+m.height], "\n") + "\n"
		}
	}

	// Footer
	s += m.Data.Footer(m.cursor, m.width, footer.Width(m.width).Height(1))
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

// Cursor returns the position of the cursor.
func (m Model[T]) Cursor() int {
	return m.cursor
}

// MoveCursor moves the cursor by delta.
func (m *Model[T]) MoveCursor(delta int) {
	m.cursor += delta
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

// lineHeight returns the rendered visual height of line i.
func (m Model[T]) lineHeight(i int) int {
	if m.width <= 0 || i < 0 || i >= m.Data.Len(m.width) {
		return 1
	}
	line := m.Data.Render(i, m.width, m.normalStyle)
	_, h := lipgloss.Size(line)
	if h < 1 {
		return 1
	}
	return h
}

// pageUp moves the cursor up by one visual page.
func (m *Model[T]) pageUp() {
	lines := 0
	moved := 0
	for i := m.cursor - 1; i >= 0; i-- {
		h := m.lineHeight(i)
		if moved > 0 && lines+h > m.height-1 {
			break
		}
		lines += h
		moved++
	}
	if moved == 0 && m.cursor > 0 {
		moved = 1
	}
	m.cursor -= moved
	m.fixOffset()
}

// pageDown moves the cursor down by one visual page.
func (m *Model[T]) pageDown() {
	n := m.Data.Len(m.width)
	lines := 0
	moved := 0
	for i := m.cursor + 1; i < n; i++ {
		h := m.lineHeight(i)
		if moved > 0 && lines+h > m.height-1 {
			break
		}
		lines += h
		moved++
	}
	if moved == 0 && m.cursor < n-1 {
		moved = 1
	}
	m.cursor += moved
	m.fixOffset()
}

// fixOffset fixes the cursor and offset locations to be consistent
// with the requested changes. Changes could be the height, width,
// or cursor position.
func (m *Model[T]) fixOffset() {
	n := m.Data.Len(m.width)
	// Fix cursor location
	if m.cursor > n-1 {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// cursor before offset - offset needs to be decreased
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// Ensure visual lines from m.offset through m.cursor fit within m.height
	if n > 0 && m.height > 0 && m.width > 0 {
		lines := 0
		for i := m.cursor; i >= m.offset; i-- {
			h := m.lineHeight(i)
			if i < m.cursor && lines+h > m.height {
				m.offset = i + 1
				break
			}
			lines += h
		}
	}

	if m.offset < 0 {
		m.offset = 0
	}
}
