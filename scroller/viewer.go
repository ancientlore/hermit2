package scroller

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

// Viewer defines types that can view text data, including
// scrolling and pagination.
type Viewer interface {
	io.Closer
	Render(i, width int, baseStyle lipgloss.Style) string // Renders the line at position i
	Footer(i, width int, baseStyle lipgloss.Style) string // Renders the footer
	Len(width int) int                                    // Length of data
}
