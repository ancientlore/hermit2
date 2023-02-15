package scroller

import "github.com/charmbracelet/lipgloss"

// Viewer defines types that can view text data, including
// scrolling and pagination.
type Viewer interface {
	Render(i, width int, baseStyle lipgloss.Style) string // Renders the line at position i
	Len() int                                             // Length of data
}
