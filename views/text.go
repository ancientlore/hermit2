package views

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
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
func (v Text) Len(width int) int {
	return len(v)
}

// Close closes the viewer, if necessary.
func (v Text) Close() error {
	return nil
}

// NewText expands tabs and splits the string into a slice of lines.
func NewText(t string, fpath string) Text {

	var lexer string
	if l := lexers.Match(fpath); l != nil {
		lexer = l.Config().Name
	}

	s := xstrings.ExpandTabs(strings.ReplaceAll(t, "\r", ""), 8)

	var buf bytes.Buffer
	err := quick.Highlight(&buf, s, lexer, "terminal256", "hermit")
	if err == nil {
		s = buf.String()
	}

	return strings.Split(s, "\n")
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
