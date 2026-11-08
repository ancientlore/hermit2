package popup

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// HistSize is the maximum number of entries stored in an input History.
const HistSize = 10

var (
	editStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#008B8B")).
			Foreground(lipgloss.Color("#FFFFFF"))

	editCursorInsStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FFFF00")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true)

	editCursorOvrStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FFFFFF")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true)
)

// History maintains a MRU list of previously entered strings for Edit fields,
// modeled after the original C++ History class.
type History struct {
	items []string
}

// NewHistory creates a new empty History buffer.
func NewHistory() *History {
	return &History{}
}

// Add inserts or promotes text to the front of the history buffer.
func (h *History) Add(text string) {
	if h == nil || text == "" {
		return
	}
	for i, item := range h.items {
		if item == text {
			// Promote existing entry to front
			copy(h.items[1:i+1], h.items[0:i])
			h.items[0] = text
			return
		}
	}
	h.items = append([]string{text}, h.items...)
	if len(h.items) > HistSize {
		h.items = h.items[:HistSize]
	}
}

// Len returns the number of entries in the history.
func (h *History) Len() int {
	if h == nil {
		return 0
	}
	return len(h.items)
}

// At returns the history item at index i.
func (h *History) At(i int) string {
	if h == nil || i < 0 || i >= len(h.items) {
		return ""
	}
	return h.items[i]
}

// Items returns a copy of the history items ordered from most recent to oldest.
func (h *History) Items() []string {
	if h == nil || len(h.items) == 0 {
		return nil
	}
	out := make([]string, len(h.items))
	copy(out, h.items)
	return out
}

// SetItems replaces the history buffer with the given items (ordered most recent first),
// omitting empty or duplicate entries and capping at HistSize.
func (h *History) SetItems(items []string) {
	if h == nil {
		return
	}
	h.items = nil
	for _, item := range items {
		if item == "" {
			continue
		}
		dup := false
		for _, existing := range h.items {
			if existing == item {
				dup = true
				break
			}
		}
		if !dup {
			h.items = append(h.items, item)
			if len(h.items) >= HistSize {
				break
			}
		}
	}
}

// Edit is a single-line text editor widget for popups with history support,
// modeled after the original C++ Edit class.
type Edit struct {
	width        int
	maxLength    int
	text         []rune
	textCursor   int
	windowCursor int
	insMode      bool
	history      *History
	histIndex    int
}

// NewEdit creates a new Edit field with the given display width, maximum character
// capacity, initial text value, and optional history buffer.
func NewEdit(width, maxLength int, initialText string, history *History) Edit {
	runes := []rune(initialText)
	if maxLength > 0 && len(runes) > maxLength {
		runes = runes[:maxLength]
	}
	histIndex := -1
	if history != nil && history.Len() > 0 && history.At(0) == string(runes) {
		histIndex = 0
	}
	return Edit{
		width:        width,
		maxLength:    maxLength,
		text:         runes,
		textCursor:   0,
		windowCursor: 0,
		insMode:      true,
		history:      history,
		histIndex:    histIndex,
	}
}

// Value returns the current text in the edit buffer.
func (e Edit) Value() string {
	return string(e.text)
}

// SetValue replaces the edit buffer with s and places the cursor at the end.
func (e *Edit) SetValue(s string) {
	runes := []rune(s)
	if e.maxLength > 0 && len(runes) > e.maxLength {
		runes = runes[:e.maxLength]
	}
	e.text = runes
	e.textCursor = len(runes)
	e.windowCursor = len(runes)
	if e.width > 0 && e.windowCursor > e.width-1 {
		e.windowCursor = e.width - 1
	}
	e.histIndex = -1
}

// AtEnd reports whether the text cursor is positioned at the end of the buffer.
func (e Edit) AtEnd() bool {
	return e.textCursor >= len(e.text)
}

// Accept commits the current text to the history buffer (if non-empty) and returns it.
func (e *Edit) Accept() string {
	val := string(e.text)
	if e.history != nil {
		e.history.Add(val)
	}
	return val
}

// HandleKey processes a key press message for the Edit field.
func (e *Edit) HandleKey(msg tea.KeyPressMsg) bool {
	switch msg.String() {
	case "up":
		if e.history != nil && e.histIndex+1 < e.history.Len() {
			e.histIndex++
			e.setTextFromHistory(e.history.At(e.histIndex))
		}
		return true

	case "down":
		if e.history != nil && e.histIndex >= 0 {
			e.histIndex--
			if e.histIndex >= 0 {
				e.setTextFromHistory(e.history.At(e.histIndex))
			} else {
				e.setTextFromHistory("")
			}
		}
		return true

	case "left":
		if e.textCursor > 0 {
			e.textCursor--
			if e.windowCursor > 0 {
				e.windowCursor--
			}
		}
		return true

	case "right":
		if e.textCursor < len(e.text) {
			e.textCursor++
			if e.windowCursor < e.width-1 {
				e.windowCursor++
			}
		}
		return true

	case "home":
		e.textCursor = 0
		e.windowCursor = 0
		return true

	case "end":
		e.textCursor = len(e.text)
		e.windowCursor = len(e.text)
		if e.windowCursor > e.width-1 {
			e.windowCursor = e.width - 1
		}
		return true

	case "insert":
		e.insMode = !e.insMode
		return true

	case "backspace":
		if e.textCursor > 0 && len(e.text) > 0 {
			e.text = append(e.text[:e.textCursor-1], e.text[e.textCursor:]...)
			e.textCursor--
			if e.windowCursor > 0 {
				e.windowCursor--
			}
			e.histIndex = -1
		}
		return true

	case "delete":
		if e.textCursor < len(e.text) && len(e.text) > 0 {
			e.text = append(e.text[:e.textCursor], e.text[e.textCursor+1:]...)
			e.histIndex = -1
		}
		return true

	case "tab":
		e.insertRune(' ')
		return true
	}

	if msg.Text != "" {
		for _, r := range msg.Text {
			e.insertRune(r)
		}
		return true
	}

	return false
}

func (e *Edit) setTextFromHistory(s string) {
	runes := []rune(s)
	if e.maxLength > 0 && len(runes) > e.maxLength {
		runes = runes[:e.maxLength]
	}
	e.text = runes
	e.textCursor = 0
	e.windowCursor = 0
}

func (e *Edit) insertRune(r rune) {
	if (e.insMode && len(e.text) < e.maxLength) || (!e.insMode && e.textCursor < e.maxLength) {
		if e.insMode || e.textCursor == len(e.text) {
			e.text = append(e.text, 0)
			copy(e.text[e.textCursor+1:], e.text[e.textCursor:])
			e.text[e.textCursor] = r
		} else {
			e.text[e.textCursor] = r
		}
		e.textCursor++
		if e.windowCursor < e.width-1 {
			e.windowCursor++
		}
		e.histIndex = -1
	}
}

// Render formats the single-line edit box with its visible text and active cursor.
func (e Edit) Render() string {
	if e.width <= 0 {
		return ""
	}
	textOffset := e.textCursor - e.windowCursor
	if textOffset < 0 {
		textOffset = 0
	}

	curStyle := editCursorInsStyle
	if !e.insMode {
		curStyle = editCursorOvrStyle
	}

	var b strings.Builder
	for i := 0; i < e.width; i++ {
		idx := textOffset + i
		ch := " "
		if idx >= 0 && idx < len(e.text) {
			ch = string(e.text[idx])
		}
		if i == e.windowCursor {
			b.WriteString(curStyle.Render(ch))
		} else {
			b.WriteString(editStyle.Render(ch))
		}
	}
	return b.String()
}

// DrawEdit draws an Edit field at (x, y) relative to the popup's top-left corner.
func (p *Popup) DrawEdit(x, y int, e *Edit) {
	if x < 0 || y < 0 || e == nil {
		return
	}
	z := len(p.children) + 2
	p.children = append(p.children, lipgloss.NewLayer(e.Render()).X(x).Y(y).Z(z))
}

