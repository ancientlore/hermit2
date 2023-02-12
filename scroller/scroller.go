package scroller

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	header = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	footer = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
)

// Model implements scrolling behavior over a Viewer.
type Model struct {
	Header   string    // Header text
	Viewport Viewer    // The view we are using
	Prev     tea.Model // The previous model (for going back)
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
			m.Viewport.Up()

		// The "down" keys move the cursor down
		case key.Matches(msg, DefaultKeyMap.Down):
			m.Viewport.Down()

		case key.Matches(msg, DefaultKeyMap.Left):
			if m.Prev != nil {
				return m.Prev, func() tea.Msg { return tea.WindowSizeMsg{Width: m.Viewport.Width(), Height: m.Viewport.Height() + 2} }
			}

		case key.Matches(msg, DefaultKeyMap.Home):
			m.Viewport.Home()

		case key.Matches(msg, DefaultKeyMap.End):
			m.Viewport.End()

		case key.Matches(msg, DefaultKeyMap.PageUp):
			m.Viewport.PageUp()

		case key.Matches(msg, DefaultKeyMap.PageDown):
			m.Viewport.PageDown()

		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.Viewport.SetWidth(msg.Width)
		m.Viewport.SetHeight(msg.Height - 2) // account for header and footer
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

// View renders the contents of the scroller. It is intended to be implemented
// by the class embedding the scroller.
func (m Model) View() string {
	s := header.Width(m.Viewport.Width()).Height(1).Render(m.Header) + "\n"
	s += m.Viewport.View()
	s += footer.Width(m.Viewport.Width()).Height(1).Render(fmt.Sprintf("%d / %d", m.Viewport.Pos()+1, m.Viewport.Len()))
	return s
}
