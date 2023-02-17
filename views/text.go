package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/huandu/xstrings"
)

// Text manages the base logic of the cursor position and pagination of a string of text.
type Text []string

// Render formats the line at position i using the base style and view width.
func (v Text) Render(i, width int, baseStyle lipgloss.Style) string {
	if i >= 0 && i < len(v) {
		return baseStyle.Render(v[i])
	}
	return ""
}

// Footer formats the footer using the base style and view width.
func (v Text) Footer(cursor, width int, baseStyle lipgloss.Style) string {
	return baseStyle.Render(fmt.Sprintf("%d / %d", cursor+1, len(v)))
}

// Len returns the number of lines of text.
func (v Text) Len() int {
	return len(v)
}

// NewText expands tabs and splits the string into a slice of lines.
func NewText(t string) Text {
	return strings.Split(xstrings.ExpandTabs(strings.ReplaceAll(t, "\r", ""), 8), "\n")
}
