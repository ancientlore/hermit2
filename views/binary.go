package views

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Binary manages the base logic of rendering binary data.
type Binary struct {
	io.ReadSeekCloser
	size int64
}

// Render formats the line at position i using the base style and view width.
func (v Binary) Render(i, width int, baseStyle lipgloss.Style) string {
	w := dataWidth(width)
	_, err := v.Seek(int64(i*w), io.SeekStart)
	if err != nil {
		return baseStyle.Blink(true).Render(err.Error())
	}
	b := make([]byte, w)
	n, err := v.Read(b)
	if err != nil && !errors.Is(err, io.EOF) {
		return baseStyle.Blink(true).Render(err.Error())
	}
	s := fmt.Sprintf("%X%s ", b[0:n], strings.Repeat("  ", w-n))

	return baseStyle.Render(s)
}

// Footer formats the footer using the base style and view width.
func (v Binary) Footer(cursor, width int, baseStyle lipgloss.Style) string {
	return baseStyle.Render(fmt.Sprintf("%d / %d bytes", cursor*dataWidth(width), v.size))
}

// Len returns the number of lines of text.
func (v Binary) Len(width int) int {
	w := dataWidth(width)
	l := v.size / int64(w)
	if v.size%int64(w) > 0 {
		l++
	}
	return int(l)
}

// NewBinary prepares binary data for rendering.
func NewBinary(rdr io.ReadSeekCloser) (*Binary, error) {
	n, err := rdr.Seek(0, io.SeekEnd)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return &Binary{
		ReadSeekCloser: rdr,
		size:           n,
	}, nil
}

func dataWidth(width int) int {
	// BBBB BBBB BBBB BBBB 12345678
	w := width / 3
	if w <= 0 {
		w = 1
	}
	return w
}
