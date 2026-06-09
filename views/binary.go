package views

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
)

// Binary manages the base logic of rendering binary data.
type Binary struct {
	rdr  io.Closer
	data []byte
}

// Render formats the line at position i using the base style and view width.
func (v Binary) Render(i, width int, baseStyle lipgloss.Style) string {
	w := dataWidth(width)
	offset := i * w
	if offset >= len(v.data) {
		return ""
	}
	end := offset + w
	if end > len(v.data) {
		end = len(v.data)
	}
	chunk := v.data[offset:end]

	s := fmt.Sprintf("% X%s  ", chunk, strings.Repeat("   ", w-len(chunk)))
	var x strings.Builder
	for i := 0; i < len(chunk); i++ {
		if unicode.IsPrint(rune(chunk[i])) {
			x.WriteRune(rune(chunk[i]))
		} else {
			x.WriteRune('.')
		}
	}
	return baseStyle.Render(s + x.String())
}

// Footer formats the footer using the base style and view width.
func (v Binary) Footer(cursor, width int, baseStyle lipgloss.Style) string {
	return baseStyle.Render(fmt.Sprintf("%d / %d bytes (%d bytes per row)", cursor*dataWidth(width), len(v.data), dataWidth(width)))
}

// Len returns the number of lines of text.
func (v Binary) Len(width int) int {
	w := dataWidth(width)
	l := len(v.data) / w
	if len(v.data)%w > 0 {
		l++
	}
	return l
}

// Close closes the underlying reader.
func (v Binary) Close() error {
	if v.rdr != nil {
		return v.rdr.Close()
	}
	return nil
}

// NewBinary prepares binary data for rendering by reading it all into memory.
func NewBinary(rdr io.ReadSeekCloser) (*Binary, error) {
	_, err := rdr.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	return &Binary{
		rdr:  rdr,
		data: data,
	}, nil
}

func dataWidth(width int) int {
	// BBBB BBBB BBBB BBBB 12345678
	w := (width-2)/4 - ((width-2)/4)%8
	if w <= 0 {
		w = 1
	}
	return w
}
