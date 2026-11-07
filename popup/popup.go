package popup

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	borderStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#888B7E")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	bgStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#2D2B55")).
		Foreground(lipgloss.Color("#FFFFFF"))

	fieldStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2B55")).
			Foreground(lipgloss.Color("#55FFFF"))

	fieldKeyStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2B55")).
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)

	selectedFieldStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#55FFFF"))

	selectedKeyStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFF00")).
				Bold(true)
)

// Popup represents a centered overlay dialog window modeled after the original
// C++ Popup class.
type Popup struct {
	width    int
	height   int
	title    string
	children []*lipgloss.Layer
}

// New creates a new Popup with the given dimensions and title.
func New(width, height int, title string) *Popup {
	return &Popup{
		width:  width,
		height: height,
		title:  title,
	}
}

// WriteText writes text at coordinates (x, y) relative to the popup's top-left corner.
func (p *Popup) WriteText(x, y int, text string) {
	if x < 0 || y < 0 || text == "" {
		return
	}
	z := len(p.children) + 2
	rendered := bgStyle.Render(text)
	p.children = append(p.children, lipgloss.NewLayer(rendered).X(x).Y(y).Z(z))
}

// DrawField draws a selectable option field at (x, y) relative to the popup's
// top-left corner, highlighting the first uppercase letter or digit as the hotkey.
func (p *Popup) DrawField(x, y int, name string, selected bool) {
	if x < 0 || y < 0 || name == "" {
		return
	}
	norm := fieldStyle
	high := fieldKeyStyle
	if selected {
		norm = selectedFieldStyle
		high = selectedKeyStyle
	}

	var b strings.Builder
	highlighted := false
	for _, r := range name {
		if !highlighted && (unicode.IsUpper(r) || unicode.IsDigit(r)) {
			b.WriteString(high.Render(string(r)))
			highlighted = true
		} else {
			b.WriteString(norm.Render(string(r)))
		}
	}

	z := len(p.children) + 2
	p.children = append(p.children, lipgloss.NewLayer(b.String()).X(x).Y(y).Z(z))
}

// renderBox builds the base background and border string for the popup.
func (p *Popup) renderBox() string {
	if p.width < 3 || p.height < 3 {
		return ""
	}

	var rows []string

	// Top border row with centered title
	title := p.title
	maxTitleWidth := p.width - 2
	if lipgloss.Width(title) > maxTitleWidth {
		r := []rune(title)
		for len(r) > 0 && lipgloss.Width(string(r)) > maxTitleWidth {
			r = r[:len(r)-1]
		}
		title = string(r)
	}
	titleWidth := lipgloss.Width(title)
	leftPad := p.width/2 - (titleWidth/2 + titleWidth%2)
	if leftPad < 1 {
		leftPad = 1
	}
	rightPad := p.width - leftPad - titleWidth
	if rightPad < 1 {
		rightPad = 1
	}
	topRow := borderStyle.Render(strings.Repeat(" ", leftPad) + title + strings.Repeat(" ", rightPad))
	rows = append(rows, topRow)

	// Middle rows: 1-char border on each side, background fill in between
	sideBorder := borderStyle.Render(" ")
	innerBg := bgStyle.Render(strings.Repeat(" ", p.width-2))
	middleRow := sideBorder + innerBg + sideBorder
	for y := 1; y < p.height-1; y++ {
		rows = append(rows, middleRow)
	}

	// Bottom border row
	bottomRow := borderStyle.Render(strings.Repeat(" ", p.width))
	rows = append(rows, bottomRow)

	return strings.Join(rows, "\n")
}

// Layer returns the composed lipgloss.Layer for the popup at (x, y).
func (p *Popup) Layer(x, y int) *lipgloss.Layer {
	base := lipgloss.NewLayer(p.renderBox()).X(x).Y(y).Z(1)
	if len(p.children) > 0 {
		base.AddLayers(p.children...)
	}
	return base
}

// Overlay composites the popup centered on top of the given background tea.View.
func (p *Popup) Overlay(bg tea.View) tea.View {
	scrWidth, scrHeight := lipgloss.Size(bg.Content)
	x := scrWidth/2 - (p.width/2 + p.width%2)
	y := scrHeight/2 - (p.height/2 + p.height%2)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	bgLayer := lipgloss.NewLayer(bg.Content).X(0).Y(0).Z(0)
	popupLayer := p.Layer(x, y)

	comp := lipgloss.NewCompositor(bgLayer, popupLayer)
	bg.SetContent(comp.Render())
	return bg
}

