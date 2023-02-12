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

	var v views.Text
	v.SetText(string(in))

	return scroller.Model{
		Header:   path,
		Viewport: &v,
		Prev:     prev,
	}
}
