package browser

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"text/template"

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
			if err != nil && !errors.Is(err, io.EOF) {
				f.Close()
				return nil, err
			}
			if strings.HasPrefix(http.DetectContentType(b[0:n]), "text") {
				rdr = io.MultiReader(bytes.NewReader(b[0:n]), f)
				isText = true
			}
		}

		if isText {
			m, err := NewTextModel(rdr, path.Join(folder, entry.Name()), prev)
			f.Close()
			return m, err
		} else if rs, ok := rdr.(io.ReadSeekCloser); ok {
			m, err := NewBinaryModel(rs, path.Join(folder, entry.Name()), prev)
			if err != nil {
				f.Close()
				// otherwise Viewer owns the file
			}
			return m, err
		} else {
			f.Close()
		}
	}
	return nil, fmt.Errorf("not a viewable file")
}

// NewBinaryFileModel creates a new model to view a file as bytes.
func NewBinaryFileModel(fs fs.FS, folder string, entry fs.DirEntry, prev tea.Model) (tea.Model, error) {
	if entry.Type().IsRegular() {
		f, err := fs.Open(path.Join(strings.TrimPrefix(folder, "/"), entry.Name()))
		if err != nil {
			return nil, err
		}
		if rs, ok := f.(io.ReadSeekCloser); ok {
			return NewBinaryModel(rs, path.Join(folder, entry.Name()), prev)
		}
		f.Close()
	}
	return nil, fmt.Errorf("not a viewable file")
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

// NewBinaryModel creates a new model to view a binary file.
func NewBinaryModel(rdr io.ReadSeekCloser, path string, prev tea.Model) (tea.Model, error) {
	b, err := views.NewBinary(rdr)
	if err != nil {
		return nil, err
	}
	return scroller.Model[views.Binary]{
		Header: path,
		Data:   *b,
		Prev:   prev,
	}, nil
}

//go:embed *.txt
var templateFs embed.FS

var templates = template.Must(
	template.New("info").Funcs(template.FuncMap{
		"div": func(n, d int64) int64 { return n / d },
		"mode": func(m fs.FileMode) []string {
			var a []string
			if m&fs.ModeDir != 0 {
				a = append(a, "d: is a directory")
			}
			if m&fs.ModeAppend != 0 {
				a = append(a, "a: append-only")
			}
			if m&fs.ModeExclusive != 0 {
				a = append(a, "l: exclusive use")
			}
			if m&fs.ModeTemporary != 0 {
				a = append(a, "T: temporary file; Plan 9 only")
			}
			if m&fs.ModeSymlink != 0 {
				a = append(a, "L: symbolic link")
			}
			if m&fs.ModeDevice != 0 {
				a = append(a, "D: device file")
			}
			if m&fs.ModeNamedPipe != 0 {
				a = append(a, "p: named pipe (FIFO)")
			}
			if m&fs.ModeSocket != 0 {
				a = append(a, "S: Unix domain socket")
			}
			if m&fs.ModeSetuid != 0 {
				a = append(a, "u: setuid")
			}
			if m&fs.ModeSetgid != 0 {
				a = append(a, "g: setgid")
			}
			if m&fs.ModeCharDevice != 0 {
				a = append(a, "c: Unix character device, when ModeDevice is set")
			}
			if m&fs.ModeSticky != 0 {
				a = append(a, "t: sticky")
			}
			if m&fs.ModeIrregular != 0 {
				a = append(a, "?: non-regular file; nothing else is known about this file")
			}
			return a
		},
	}).ParseFS(templateFs, "*.txt"),
)

// NewFileInfoModel creates a new model to view file information.
func NewFileInfoModel(fs fs.FS, folder string, entry fs.DirEntry, prev tea.Model) (tea.Model, error) {
	var wtr bytes.Buffer
	err := templates.ExecuteTemplate(&wtr, "fileinfo.txt", entry)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	rdr := bytes.NewReader(wtr.Bytes())
	return NewTextModel(rdr, path.Join(folder, entry.Name()), prev)
}
