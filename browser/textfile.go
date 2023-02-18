package browser

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/ancientlore/hermit2/scroller"
	"github.com/ancientlore/hermit2/views"
	tea "github.com/charmbracelet/bubbletea"
)

// NewTextModel creates a new model to view a text file.
func NewTextModel(path string, prev tea.Model) tea.Model {
	in, err := ioutil.ReadFile(filepath.FromSlash(path))
	if err != nil {
		in = []byte(err.Error())
	}

	var lexer string
	if l := lexers.Match(path); l != nil {
		lexer = l.Config().Name
	}

	var buf bytes.Buffer
	err = quick.Highlight(&buf, string(in), lexer, "terminal256", "native")
	if err == nil {
		in = buf.Bytes()
	}

	return scroller.Model[views.Text]{
		Header: path,
		Data:   views.NewText(string(in)),
		Prev:   prev,
	}
}
