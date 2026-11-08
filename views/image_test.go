package views

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func createTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((x * 255) / max(1, w-1)),
				G: uint8((y * 255) / max(1, h-1)),
				B: 128,
				A: 255,
			})
		}
	}
	return img
}

func TestIsImageExt(t *testing.T) {
	supported := []string{"a.png", "b.PNG", "c.jpg", "d.JPEG", "e.gif", "f.bmp", "g.tif", "h.tiff", "i.webp"}
	for _, name := range supported {
		if !IsImageExt(name) {
			t.Errorf("expected IsImageExt(%q) = true", name)
		}
	}
	unsupported := []string{"a.txt", "b.go", "c.ico", "d.pdf", "noext"}
	for _, name := range unsupported {
		if IsImageExt(name) {
			t.Errorf("expected IsImageExt(%q) = false", name)
		}
	}
}

func TestNewImageFormats(t *testing.T) {
	src := createTestImage(64, 48)

	tests := []struct {
		name   string
		format string
		encode func(buf *bytes.Buffer, img image.Image) error
	}{
		{"png", "PNG", func(b *bytes.Buffer, m image.Image) error { return png.Encode(b, m) }},
		{"jpeg", "JPEG", func(b *bytes.Buffer, m image.Image) error { return jpeg.Encode(b, m, nil) }},
		{"gif", "GIF", func(b *bytes.Buffer, m image.Image) error { return gif.Encode(b, m, nil) }},
		{"bmp", "BMP", func(b *bytes.Buffer, m image.Image) error { return bmp.Encode(b, m) }},
		{"tiff", "TIFF", func(b *bytes.Buffer, m image.Image) error { return tiff.Encode(b, m, nil) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tc.encode(&buf, src); err != nil {
				t.Fatalf("encode %s: %v", tc.name, err)
			}
			v, err := NewImage(&buf)
			if err != nil {
				t.Fatalf("NewImage(%s): %v", tc.name, err)
			}
			if v.Format() != tc.format {
				t.Errorf("got format %q, want %q", v.Format(), tc.format)
			}
			w, h := v.Bounds()
			if w != 64 || h != 48 {
				t.Errorf("got bounds %dx%d, want 64x48", w, h)
			}
			out := v.Render(40, 20)
			lines := strings.Split(out, "\n")
			if len(lines) != 20 {
				t.Errorf("got %d rendered lines, want 20", len(lines))
			}
		})
	}
}

func TestImageZoomPanAndGrayscale(t *testing.T) {
	f, err := os.Open("../media/hermit.png")
	if err != nil {
		t.Fatalf("open media/hermit.png: %v", err)
	}
	defer f.Close()

	v, err := NewImage(f)
	if err != nil {
		t.Fatalf("NewImage: %v", err)
	}

	out1 := v.Render(80, 24)
	if lines := strings.Split(out1, "\n"); len(lines) != 24 {
		t.Fatalf("got %d lines, want 24", len(lines))
	}
	// Cached render should return identical string
	if outCached := v.Render(80, 24); outCached != out1 {
		t.Errorf("expected cached render to match")
	}

	// Zoom in
	v.ZoomIn()
	if v.ZoomPercent() <= 100 {
		t.Errorf("expected zoom > 100%%, got %d%%", v.ZoomPercent())
	}
	outZoomed := v.Render(80, 24)
	if outZoomed == out1 {
		t.Errorf("expected zoomed render to differ from initial render")
	}

	// Pan in all directions while zoomed in
	v.ZoomIn()
	v.ZoomIn()
	_ = v.Render(80, 24)
	cx, cy := v.centerX, v.centerY
	v.Right()
	if v.centerX <= cx {
		t.Errorf("expected Right() to increase centerX (%v -> %v)", cx, v.centerX)
	}
	v.Left()
	v.Down()
	if v.centerY <= cy {
		t.Errorf("expected Down() to increase centerY (%v -> %v)", cy, v.centerY)
	}
	v.Up()

	// Grayscale toggle
	v.ToggleGrayscale()
	if !v.Grayscale() {
		t.Errorf("expected grayscale to be true")
	}
	ftr := v.Footer(80, lipgloss.NewStyle())
	if !strings.Contains(ftr, "[GRAY]") {
		t.Errorf("expected footer to include [GRAY], got %q", ftr)
	}

	// Reset zoom
	v.ResetZoom()
	if v.ZoomPercent() != 100 {
		t.Errorf("expected ResetZoom to set 100%%, got %d%%", v.ZoomPercent())
	}
}

func BenchmarkImageZoomPanRender(b *testing.B) {
	// Create a 3840x2160 (4K) image and benchmark zooming/panning/rendering at 120x40 terminal size
	raw := createTestImage(3840, 2160)
	v := &Image{
		mips:    buildMipmaps(raw),
		origW:   3840,
		origH:   2160,
		format:  "PNG",
		zoom:    1.0,
		centerX: 1920,
		centerY: 1080,
		dirty:   true,
	}
	_ = v.Render(120, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch i % 6 {
		case 0:
			v.ZoomIn()
		case 1:
			v.Right()
		case 2:
			v.Down()
		case 3:
			v.Left()
		case 4:
			v.Up()
		case 5:
			v.ZoomOut()
		}
		_ = v.Render(120, 40)
	}
}

