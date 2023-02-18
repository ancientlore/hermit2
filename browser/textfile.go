package browser

import (
	"io/ioutil"
	"path/filepath"

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

	return scroller.Model[views.Text]{
		Header: path,
		Data:   views.NewText(string(in), path),
		Prev:   prev,
	}
}
