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
type Text struct {
	cursor int      // Current position of cursor
	offset int      // The offset of the view (enables scrolling)
	width  int      // The width of the current view
	height int      // The height of the current view
	text   []string // Text of file
}

// Len returns the number of lines of text.
func (v Text) Len() int {
	return len(v.text)
}

// Pos returns the position of the cursor.
func (v Text) Pos() int {
	return v.cursor
}

// Width returns the width of the view.
func (v Text) Width() int {
	return v.width
}

// Height returns the height of the view.
func (v Text) Height() int {
	return v.height
}

// SetWidth sets the width of the view to w and makes
// any other needed adjustments.
func (v *Text) SetWidth(w int) {
	v.width = w
}

// SetHeight sets the height of the view to h and makes
// any other needed adjustments.
func (v *Text) SetHeight(h int) {
	v.height = h
	v.fixOffset()
}

// Home sets the cursor position to zero.
func (v *Text) Home() {
	v.cursor = 0
	v.fixOffset()
}

// End sets the cursor position to the last line of text.
func (v *Text) End() {
	v.cursor = v.Len() - 1
	v.fixOffset()
}

// PageUp moves the cursor up one page.
func (v *Text) PageUp() {
	v.cursor -= v.Height() - 1
	v.fixOffset()
}

// PageDown moves the cursor down one page.
func (v *Text) PageDown() {
	v.cursor += v.Height() - 1
	v.fixOffset()
}

// Up moves the cursor up one line.
func (v *Text) Up() {
	v.cursor--
	v.fixOffset()
}

// Down moves the cursor down one line.
func (v *Text) Down() {
	v.cursor++
	v.fixOffset()
}

// View returns the current page of text visible in the viewport.
func (v Text) View() string {
	var s string
	for i := v.offset; i < v.Len() && i < v.height+v.offset; i++ {
		choice := v.text[i]
		style := normal.Width(v.width).Height(1).MaxWidth(v.width)
		if v.Pos() == i {
			style = highlight.Width(v.width).Height(1).MaxWidth(v.width)
		}
		s += style.Render(style.Render(choice)) + "\n"
	}
	repeat := v.height - v.Len()
	if repeat > 0 {
		s += strings.Repeat("\n", repeat)
	}
	return s
}

// SetText assigns the text to be shown in the viewer.
func (v *Text) SetText(t string) {
	v.text = strings.Split(xstrings.ExpandTabs(strings.ReplaceAll(t, "\r", ""), 8), "\n")
	v.fixOffset()
}

// fixOffset fixes the cursor and offset locations to be consistent
// with the requested changes. Changes could be the height, width,
// or cursor position.
func (v *Text) fixOffset() {
	// Fix cursor location
	if v.cursor > v.Len()-1 {
		v.cursor = v.Len() - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}

	// cursor before offset - offset needs to be decreased
	if v.cursor < v.offset {
		v.offset = v.cursor
	}

	// cursor greater than offset + window size - offset needs to be increased
	if v.cursor >= v.offset+v.height {
		v.offset = v.cursor - v.height + 1
	}
	if v.offset < 0 {
		v.offset = 0
	}
}
