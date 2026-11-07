package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/huandu/xstrings"
)

// Text manages the base logic of the cursor position and pagination of a string of text.
type Text struct {
	lines [][]chroma.Token
	style *chroma.Style
}

// Render formats the line at position i using the base style and view width.
func (v Text) Render(i, width int, baseStyle lipgloss.Style) string {
	if i < 0 || i >= len(v.lines) {
		return ""
	}
	bg := baseStyle.GetBackground()
	var b strings.Builder
	for _, tok := range v.lines[i] {
		val := strings.TrimSuffix(tok.Value, "\n")
		if val == "" {
			continue
		}
		ts := lipgloss.NewStyle().Background(bg)
		if v.style != nil {
			entry := v.style.Get(tok.Type)
			if entry.Colour.IsSet() {
				ts = ts.Foreground(lipgloss.Color(entry.Colour.String()))
			}
			if entry.Bold == chroma.Yes {
				ts = ts.Bold(true)
			}
			if entry.Italic == chroma.Yes {
				ts = ts.Italic(true)
			}
			if entry.Underline == chroma.Yes {
				ts = ts.Underline(true)
			}
		}
		b.WriteString(ts.Render(val))
	}
	return baseStyle.Render(b.String())
}

// Footer formats the footer using the base style and view width.
func (v Text) Footer(cursor, width int, baseStyle lipgloss.Style) string {
	return baseStyle.Render(fmt.Sprintf("%d / %d", cursor+1, len(v.lines)))
}

// Len returns the number of lines of text.
func (v Text) Len(width int) int {
	return len(v.lines)
}

// Close closes the viewer, if necessary.
func (v Text) Close() error {
	return nil
}

// NewText expands tabs and splits the string into tokenized lines for syntax-highlighted rendering.
func NewText(t string, fpath string) Text {
	s := xstrings.ExpandTabs(strings.ReplaceAll(t, "\r", ""), 8)

	l := lexers.Match(fpath)
	if l == nil {
		l = lexers.Analyse(s)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	style := styles.Get("hermit")
	if style == nil {
		style = styles.Fallback
	}

	var lines [][]chroma.Token
	if it, err := l.Tokenise(nil, s); err == nil {
		lines = chroma.SplitTokensIntoLines(it.Tokens())
	} else {
		lines = chroma.SplitTokensIntoLines([]chroma.Token{{Type: chroma.Text, Value: s}})
	}
	if len(lines) == 0 {
		lines = [][]chroma.Token{{}}
	}

	return Text{
		lines: lines,
		style: style,
	}
}

func init() {
	styles.Register(chroma.MustNewStyle("hermit", chroma.StyleEntries{
		chroma.Background:         "#d0d0d0 bg: ", //"#d0d0d0 bg:#202020",
		chroma.TextWhitespace:     "#666666",
		chroma.Comment:            "italic #999999",
		chroma.CommentPreproc:     "noitalic bold #cd2828",
		chroma.CommentSpecial:     "noitalic bold #e50808 bg: ", // "noitalic bold #e50808 bg:#520000",
		chroma.Keyword:            "bold #6ab825",
		chroma.KeywordPseudo:      "nobold",
		chroma.OperatorWord:       "bold #6ab825",
		chroma.LiteralString:      "#ed9d13",
		chroma.LiteralStringOther: "#ffa500",
		chroma.LiteralNumber:      "#3677a9",
		chroma.NameBuiltin:        "#24909d",
		chroma.NameVariable:       "#40ffff",
		chroma.NameConstant:       "#40ffff",
		chroma.NameClass:          "underline #447fcf",
		chroma.NameFunction:       "#447fcf",
		chroma.NameNamespace:      "underline #447fcf",
		chroma.NameException:      "#bbbbbb",
		chroma.NameTag:            "bold #6ab825",
		chroma.NameAttribute:      "#bbbbbb",
		chroma.NameDecorator:      "#ffa500",
		chroma.GenericHeading:     "bold #ffffff",
		chroma.GenericSubheading:  "underline #ffffff",
		chroma.GenericDeleted:     "#d22323",
		chroma.GenericInserted:    "#589819",
		chroma.GenericError:       "#d22323",
		chroma.GenericEmph:        "italic",
		chroma.GenericStrong:      "bold",
		chroma.GenericPrompt:      "#aaaaaa",
		chroma.GenericOutput:      "#cccccc",
		chroma.GenericTraceback:   "#d22323",
		chroma.GenericUnderline:   "underline",
		chroma.Error:              "#a61717 bg: ", // "bg:#e3d2d2 #a61717",
	}))

	/*
		s := styles.Get("hermit")
		for _, t := range s.Types() {
			e := s.Get(t)
			fmt.Println(t, e)
		}
	*/
}
