package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/huandu/xstrings"
)

var (
	normal    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	highlight = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
)

// Text manages the base logic of the cursor position and pagination of a string of text.
type Text []string

// At returns the line of text at position i.
func (v Text) At(i int) string {
	if i >= 0 && i < len(v) {
		return v[i]
	}
	return ""
}

// Len returns the number of lines of text.
func (v Text) Len() int {
	return len(v)
}

// SplitText expands tabs and splits the string into a slice of lines.
func NewText(t string) Text {
	return strings.Split(xstrings.ExpandTabs(strings.ReplaceAll(t, "\r", ""), 8), "\n")
}
