package browser

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/ancientlore/hermit2/scroller"
	"github.com/ancientlore/hermit2/views"
	tea "github.com/charmbracelet/bubbletea"
)

func NewFileModel(fs fs.FS, folder string, entry fs.DirEntry, prev tea.Model) (tea.Model, error) {
	if entry.Type().IsRegular() {
		f, err := fs.Open(path.Join(strings.TrimPrefix(folder, "/"), entry.Name()))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var rdr io.Reader
		rdr = f
		isText := false

		// Check mime type
		if strings.HasPrefix(mime.TypeByExtension(path.Ext(entry.Name())), "text") {
			isText = true
		} else {
			// Check by inspecting the file
			b := make([]byte, 512)
			n, err := f.Read(b)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(http.DetectContentType(b[0:n]), "text") {
				rdr = io.MultiReader(bytes.NewReader(b[0:n]), f)
				isText = true
			}
		}

		if isText {
			return NewTextModel(rdr, path.Join(folder, entry.Name()), prev)
		}
	}
	return nil, fmt.Errorf("not a text file")
}

// NewTextModel creates a new model to view a text file.
func NewTextModel(rdr io.Reader, path string, prev tea.Model) (tea.Model, error) {
	b, err := io.ReadAll(rdr)
	if err != nil {
		return nil, err
	}

	return scroller.Model[views.Text]{
		Header: path,
		Data:   views.NewText(string(b), path),
		Prev:   prev,
	}, nil
}
