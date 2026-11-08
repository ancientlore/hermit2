package scroller

import (
	"io"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ImageViewer defines types that can render and manipulate a 2D image viewport.
type ImageViewer interface {
	io.Closer
	Render(width, height int) string
	Footer(width int, baseStyle lipgloss.Style) string
	ZoomIn()
	ZoomOut()
	ResetZoom()
	Left()
	Right()
	Up()
	Down()
	ToggleGrayscale()
}

// ImageModel implements an interactive pan/zoom image viewer with header and footer bars.
type ImageModel[T ImageViewer] struct {
	Header     string
	HeaderTail string
	Data       T
	Prev       tea.Model
	width      int
	height     int
}

// Init initializes the image model.
func (m ImageModel[T]) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation, zooming, panning, and window resizing.
func (m ImageModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, DefaultImageKeyMap.Up):
			m.Data.Up()

		case key.Matches(msg, DefaultImageKeyMap.Down):
			m.Data.Down()

		case key.Matches(msg, DefaultImageKeyMap.Left):
			m.Data.Left()

		case key.Matches(msg, DefaultImageKeyMap.Right):
			m.Data.Right()

		case key.Matches(msg, DefaultImageKeyMap.ZoomIn):
			m.Data.ZoomIn()

		case key.Matches(msg, DefaultImageKeyMap.ZoomOut):
			m.Data.ZoomOut()

		case key.Matches(msg, DefaultImageKeyMap.ResetZoom):
			m.Data.ResetZoom()

		case key.Matches(msg, DefaultImageKeyMap.Grayscale):
			m.Data.ToggleGrayscale()

		case key.Matches(msg, DefaultImageKeyMap.Back):
			if m.Prev != nil {
				m.Data.Close()
				return m.Prev, func() tea.Msg {
					return tea.WindowSizeMsg{Width: m.width, Height: m.height + 2}
				}
			}

		case key.Matches(msg, DefaultImageKeyMap.Quit):
			m.Data.Close()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2 // account for header and footer
	}

	return m, nil
}

// View renders the header, half-block image canvas, and footer.
func (m ImageModel[T]) View() tea.View {
	s := RenderHeader(m.Header, m.HeaderTail, m.width) + "\n"
	s += m.Data.Render(m.width, m.height) + "\n"
	s += m.Data.Footer(m.width, FooterStyle(m.width))

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

