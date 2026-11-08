package views

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"path"
	"strings"

	"charm.land/lipgloss/v2"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	zoomFactor = 1.25
	panFraction = 0.10
)

type rgb struct {
	r, g, b uint8
}

type sampleSpan struct {
	i0, i1 int
	w0, w1 float32
}

// Image manages high-color half-block terminal rendering, zooming, and panning
// for decoded raster images using a precomputed mipmap pyramid.
type Image struct {
	mips    []*image.RGBA
	origW   int
	origH   int
	format  string
	zoom    float64
	centerX float64
	centerY float64
	gray    bool

	lastWidth  int
	lastHeight int
	dirty      bool
	rendered   string
	scaledBuf  []rgb
	xSpans     []sampleSpan
}

// IsImageExt reports whether the filename has a supported image file extension.
func IsImageExt(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tif", ".tiff", ".webp":
		return true
	default:
		return false
	}
}

// NewImage decodes an image from rdr and builds a mipmap pyramid for fast zooming and panning.
func NewImage(rdr io.Reader) (*Image, error) {
	src, format, err := image.Decode(rdr)
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid image dimensions %dx%d", w, h)
	}

	base := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(base, base.Bounds(), src, b.Min, draw.Src)

	mips := buildMipmaps(base)

	return &Image{
		mips:    mips,
		origW:   w,
		origH:   h,
		format:  strings.ToUpper(format),
		zoom:    1.0,
		centerX: float64(w) / 2.0,
		centerY: float64(h) / 2.0,
		dirty:   true,
	}, nil
}

func buildMipmaps(base *image.RGBA) []*image.RGBA {
	mips := []*image.RGBA{base}
	curr := base
	for {
		b := curr.Bounds()
		w, h := b.Dx(), b.Dy()
		if w <= 32 || h <= 32 {
			break
		}
		next := downsample2x(curr)
		mips = append(mips, next)
		curr = next
	}
	return mips
}

func downsample2x(src *image.RGBA) *image.RGBA {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	dw, dh := sw/2, sh/2
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	sPix := src.Pix
	dPix := dst.Pix
	sStride := src.Stride
	dStride := dst.Stride

	for y := 0; y < dh; y++ {
		sy0 := y * 2
		sy1 := sy0 + 1
		if sy1 >= sh {
			sy1 = sh - 1
		}
		row0 := sy0 * sStride
		row1 := sy1 * sStride
		dRow := y * dStride
		for x := 0; x < dw; x++ {
			sx0 := x * 2
			sx1 := sx0 + 1
			if sx1 >= sw {
				sx1 = sw - 1
			}
			i00 := row0 + sx0*4
			i10 := row0 + sx1*4
			i01 := row1 + sx0*4
			i11 := row1 + sx1*4
			di := dRow + x*4

			dPix[di+0] = uint8((uint32(sPix[i00+0]) + uint32(sPix[i10+0]) + uint32(sPix[i01+0]) + uint32(sPix[i11+0]) + 2) >> 2)
			dPix[di+1] = uint8((uint32(sPix[i00+1]) + uint32(sPix[i10+1]) + uint32(sPix[i01+1]) + uint32(sPix[i11+1]) + 2) >> 2)
			dPix[di+2] = uint8((uint32(sPix[i00+2]) + uint32(sPix[i10+2]) + uint32(sPix[i01+2]) + uint32(sPix[i11+2]) + 2) >> 2)
			dPix[di+3] = uint8((uint32(sPix[i00+3]) + uint32(sPix[i10+3]) + uint32(sPix[i01+3]) + uint32(sPix[i11+3]) + 2) >> 2)
		}
	}
	return dst
}

// Close implements io.Closer.
func (v *Image) Close() error {
	return nil
}

// Format returns the uppercase format name of the decoded image (e.g. "PNG", "JPEG").
func (v *Image) Format() string {
	return v.format
}

// Bounds returns the original width and height of the image in pixels.
func (v *Image) Bounds() (int, int) {
	return v.origW, v.origH
}

// ZoomPercent returns the current zoom level as an integer percentage (100 = fit to screen).
func (v *Image) ZoomPercent() int {
	return int(math.Round(v.zoom * 100))
}

// Grayscale reports whether grayscale mode is active.
func (v *Image) Grayscale() bool {
	return v.gray
}

// ToggleGrayscale toggles grayscale rendering mode.
func (v *Image) ToggleGrayscale() {
	v.gray = !v.gray
	v.dirty = true
}

// ResetZoom resets zoom to 100% (fit to screen) and centers the image.
func (v *Image) ResetZoom() {
	if v.zoom != 1.0 || v.centerX != float64(v.origW)/2.0 || v.centerY != float64(v.origH)/2.0 {
		v.zoom = 1.0
		v.centerX = float64(v.origW) / 2.0
		v.centerY = float64(v.origH) / 2.0
		v.dirty = true
	}
}

// ZoomIn increases the zoom factor around the current center.
func (v *Image) ZoomIn() {
	maxZoom := math.Max(64.0, float64(max(v.origW, v.origH)))
	newZoom := v.zoom * zoomFactor
	if newZoom > maxZoom {
		newZoom = maxZoom
	}
	if newZoom != v.zoom {
		v.zoom = newZoom
		v.clampCenter(v.viewportSize())
		v.dirty = true
	}
}

// ZoomOut decreases the zoom factor down to 100% (fit to screen).
func (v *Image) ZoomOut() {
	if v.zoom <= 1.0 {
		return
	}
	newZoom := v.zoom / zoomFactor
	if newZoom < 1.02 {
		v.ResetZoom()
		return
	}
	v.zoom = newZoom
	v.clampCenter(v.viewportSize())
	v.dirty = true
}

// Left pans the viewport left.
func (v *Image) Left() {
	w, h := v.viewportSize()
	visW, _, _, _, _ := v.visibleMetrics(w, h)
	if visW >= float64(v.origW) {
		return
	}
	prev := v.centerX
	v.centerX -= math.Max(1.0, visW*panFraction)
	v.clampCenter(w, h)
	if v.centerX != prev {
		v.dirty = true
	}
}

// Right pans the viewport right.
func (v *Image) Right() {
	w, h := v.viewportSize()
	visW, _, _, _, _ := v.visibleMetrics(w, h)
	if visW >= float64(v.origW) {
		return
	}
	prev := v.centerX
	v.centerX += math.Max(1.0, visW*panFraction)
	v.clampCenter(w, h)
	if v.centerX != prev {
		v.dirty = true
	}
}

// Up pans the viewport up.
func (v *Image) Up() {
	w, h := v.viewportSize()
	_, visH, _, _, _ := v.visibleMetrics(w, h)
	if visH >= float64(v.origH) {
		return
	}
	prev := v.centerY
	v.centerY -= math.Max(1.0, visH*panFraction)
	v.clampCenter(w, h)
	if v.centerY != prev {
		v.dirty = true
	}
}

// Down pans the viewport down.
func (v *Image) Down() {
	w, h := v.viewportSize()
	_, visH, _, _, _ := v.visibleMetrics(w, h)
	if visH >= float64(v.origH) {
		return
	}
	prev := v.centerY
	v.centerY += math.Max(1.0, visH*panFraction)
	v.clampCenter(w, h)
	if v.centerY != prev {
		v.dirty = true
	}
}

func (v *Image) viewportSize() (int, int) {
	w, h := v.lastWidth, v.lastHeight
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return w, h
}

func (v *Image) visibleMetrics(termW, termH int) (visW, visH, scale float64, fitW, fitRows int) {
	if termW <= 0 {
		termW = 1
	}
	if termH <= 0 {
		termH = 1
	}
	termH2 := float64(termH * 2)
	baseScale := math.Min(float64(termW)/float64(v.origW), termH2/float64(v.origH))
	scale = baseScale * v.zoom

	visW = math.Min(float64(v.origW), float64(termW)/scale)
	visH = math.Min(float64(v.origH), termH2/scale)

	fitW = int(math.Round(visW * scale))
	if fitW < 1 {
		fitW = 1
	} else if fitW > termW {
		fitW = termW
	}

	fitRows = int(math.Round((visH * scale) / 2.0))
	if fitRows < 1 {
		fitRows = 1
	} else if fitRows > termH {
		fitRows = termH
	}
	return visW, visH, scale, fitW, fitRows
}

func (v *Image) clampCenter(termW, termH int) {
	visW, visH, _, _, _ := v.visibleMetrics(termW, termH)
	halfW := visW / 2.0
	halfH := visH / 2.0

	if halfW >= float64(v.origW)/2.0 {
		v.centerX = float64(v.origW) / 2.0
	} else {
		v.centerX = math.Min(math.Max(v.centerX, halfW), float64(v.origW)-halfW)
	}

	if halfH >= float64(v.origH)/2.0 {
		v.centerY = float64(v.origH) / 2.0
	} else {
		v.centerY = math.Min(math.Max(v.centerY, halfH), float64(v.origH)-halfH)
	}
}

// Render renders the image into a string of height lines (each line up to width columns)
// using Unicode lower half-block characters ('▄') and 24-bit ANSI colors.
func (v *Image) Render(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if !v.dirty && width == v.lastWidth && height == v.lastHeight {
		return v.rendered
	}

	v.lastWidth = width
	v.lastHeight = height
	v.clampCenter(width, height)

	visW, visH, _, fitW, fitRows := v.visibleMetrics(width, height)
	fitH := fitRows * 2

	winX0 := math.Max(0, v.centerX-visW/2.0)
	winY0 := math.Max(0, v.centerY-visH/2.0)

	v.resample(winX0, winY0, visW, visH, fitW, fitH)

	padTop := (height - fitRows) / 2
	padBottom := height - fitRows - padTop
	padLeft := (width - fitW) / 2
	leftSpaces := ""
	if padLeft > 0 {
		leftSpaces = strings.Repeat(" ", padLeft)
	}

	var b strings.Builder
	// Pre-allocate worst-case capacity (~42 bytes per cell for fg+bg+utf8 + newlines)
	b.Grow(fitRows*(fitW*42+padLeft+8) + (padTop+padBottom)*2)

	for i := 0; i < padTop; i++ {
		b.WriteByte('\n')
	}

	for row := 0; row < fitRows; row++ {
		if padLeft > 0 {
			b.WriteString(leftSpaces)
		}
		topRow := (row * 2) * fitW
		botRow := (row*2 + 1) * fitW

		var lastFG, lastBG rgb
		hasFG, hasBG := false, false

		for col := 0; col < fitW; col++ {
			top := v.scaledBuf[topRow+col]
			bot := v.scaledBuf[botRow+col]

			if top == bot {
				if !hasBG || top != lastBG {
					writeBG(&b, top)
					lastBG = top
					hasBG = true
				}
				b.WriteByte(' ')
			} else {
				if !hasBG || top != lastBG {
					writeBG(&b, top)
					lastBG = top
					hasBG = true
				}
				if !hasFG || bot != lastFG {
					writeFG(&b, bot)
					lastFG = bot
					hasFG = true
				}
				b.WriteString("▄")
			}
		}
		b.WriteString("\x1b[0m")
		if row < fitRows-1 || padBottom > 0 {
			b.WriteByte('\n')
		}
	}

	for i := 0; i < padBottom; i++ {
		if i < padBottom-1 {
			b.WriteByte('\n')
		}
	}

	v.rendered = b.String()
	v.dirty = false
	return v.rendered
}

// Footer renders the status and key hint footer for the image viewer.
func (v *Image) Footer(width int, baseStyle lipgloss.Style) string {
	grayTag := ""
	if v.gray {
		grayTag = " [GRAY]"
	}
	info := fmt.Sprintf("%s %dx%d  %d%%%s    +/-: zoom  0: reset  ←↑↓→: pan  g: gray  esc: back",
		v.format, v.origW, v.origH, v.ZoomPercent(), grayTag)
	if width > 0 && lipgloss.Width(info) > width {
		r := []rune(info)
		if len(r) > width {
			r = r[:width]
		}
		for len(r) > 0 && lipgloss.Width(string(r)) > width {
			r = r[:len(r)-1]
		}
		info = string(r)
	}
	return baseStyle.Render(info)
}

func computeSpan(s0, s1 float64, limit int) sampleSpan {
	i0 := int(s0)
	if i0 < 0 {
		i0 = 0
	} else if i0 >= limit {
		i0 = limit - 1
	}

	// Check if footprint straddles into pixel i0+1
	nextBoundary := float64(i0 + 1)
	if s1 <= nextBoundary || i0+1 >= limit {
		return sampleSpan{i0: i0, i1: i0, w0: 1.0, w1: 0.0}
	}

	i1 := i0 + 1
	// Area overlap with i0 vs (i1..s1)
	span := s1 - s0
	if span <= 1e-6 {
		return sampleSpan{i0: i0, i1: i0, w0: 1.0, w1: 0.0}
	}
	w0 := float32((nextBoundary - s0) / span)
	if w0 < 0 {
		w0 = 0
	} else if w0 > 1 {
		w0 = 1
	}
	return sampleSpan{i0: i0, i1: i1, w0: w0, w1: 1.0 - w0}
}

func (v *Image) resample(winX0, winY0, winW, winH float64, fitW, fitH int) {
	needed := fitW * fitH
	if cap(v.scaledBuf) < needed {
		v.scaledBuf = make([]rgb, needed)
	} else {
		v.scaledBuf = v.scaledBuf[:needed]
	}

	if cap(v.xSpans) < fitW {
		v.xSpans = make([]sampleSpan, fitW)
	} else {
		v.xSpans = v.xSpans[:fitW]
	}

	// Choose the finest mipmap level where 1 output pixel spans at most ~2 mip pixels.
	rx := winW / float64(fitW)
	ry := winH / float64(fitH)
	ratio := math.Min(rx, ry)
	level := 0
	for level+1 < len(v.mips) && ratio >= 2.0 {
		level++
		ratio *= 0.5
	}

	mip := v.mips[level]
	mb := mip.Bounds()
	mw, mh := mb.Dx(), mb.Dy()
	pix := mip.Pix
	stride := mip.Stride

	scaleX := float64(mw) / float64(v.origW)
	scaleY := float64(mh) / float64(v.origH)

	mx0 := winX0 * scaleX
	mWinW := winW * scaleX
	my0 := winY0 * scaleY
	mWinH := winH * scaleY

	stepX := mWinW / float64(fitW)
	for dx := 0; dx < fitW; dx++ {
		sx0 := mx0 + float64(dx)*stepX
		sx1 := sx0 + stepX
		v.xSpans[dx] = computeSpan(sx0, sx1, mw)
	}

	stepY := mWinH / float64(fitH)
	gray := v.gray

	for dy := 0; dy < fitH; dy++ {
		sy0 := my0 + float64(dy)*stepY
		sy1 := sy0 + stepY
		ys := computeSpan(sy0, sy1, mh)

		row0 := ys.i0 * stride
		row1 := ys.i1 * stride
		wy0 := ys.w0
		wy1 := ys.w1
		outRow := dy * fitW

		if wy1 == 0 {
			for dx := 0; dx < fitW; dx++ {
				xs := v.xSpans[dx]
				i00 := row0 + xs.i0*4
				var r, g, b uint8
				if xs.w1 == 0 {
					r = pix[i00+0]
					g = pix[i00+1]
					b = pix[i00+2]
				} else {
					i10 := row0 + xs.i1*4
					r = uint8(float32(pix[i00+0])*xs.w0 + float32(pix[i10+0])*xs.w1 + 0.5)
					g = uint8(float32(pix[i00+1])*xs.w0 + float32(pix[i10+1])*xs.w1 + 0.5)
					b = uint8(float32(pix[i00+2])*xs.w0 + float32(pix[i10+2])*xs.w1 + 0.5)
				}
				if gray {
					lum := toGray(r, g, b)
					v.scaledBuf[outRow+dx] = rgb{lum, lum, lum}
				} else {
					v.scaledBuf[outRow+dx] = rgb{r, g, b}
				}
			}
		} else {
			for dx := 0; dx < fitW; dx++ {
				xs := v.xSpans[dx]
				i00 := row0 + xs.i0*4
				i01 := row1 + xs.i0*4
				var r, g, b uint8
				if xs.w1 == 0 {
					r = uint8(float32(pix[i00+0])*wy0 + float32(pix[i01+0])*wy1 + 0.5)
					g = uint8(float32(pix[i00+1])*wy0 + float32(pix[i01+1])*wy1 + 0.5)
					b = uint8(float32(pix[i00+2])*wy0 + float32(pix[i01+2])*wy1 + 0.5)
				} else {
					i10 := row0 + xs.i1*4
					i11 := row1 + xs.i1*4
					w00 := xs.w0 * wy0
					w10 := xs.w1 * wy0
					w01 := xs.w0 * wy1
					w11 := xs.w1 * wy1
					r = uint8(float32(pix[i00+0])*w00 + float32(pix[i10+0])*w10 + float32(pix[i01+0])*w01 + float32(pix[i11+0])*w11 + 0.5)
					g = uint8(float32(pix[i00+1])*w00 + float32(pix[i10+1])*w10 + float32(pix[i01+1])*w01 + float32(pix[i11+1])*w11 + 0.5)
					b = uint8(float32(pix[i00+2])*w00 + float32(pix[i10+2])*w10 + float32(pix[i01+2])*w01 + float32(pix[i11+2])*w11 + 0.5)
				}
				if gray {
					lum := toGray(r, g, b)
					v.scaledBuf[outRow+dx] = rgb{lum, lum, lum}
				} else {
					v.scaledBuf[outRow+dx] = rgb{r, g, b}
				}
			}
		}
	}
}

func toGray(r, g, b uint8) uint8 {
	return uint8((19595*uint32(r) + 38470*uint32(g) + 7471*uint32(b) + 32768) >> 16)
}

func writeFG(b *strings.Builder, c rgb) {
	b.WriteString("\x1b[38;2;")
	writeUint8(b, c.r)
	b.WriteByte(';')
	writeUint8(b, c.g)
	b.WriteByte(';')
	writeUint8(b, c.b)
	b.WriteByte('m')
}

func writeBG(b *strings.Builder, c rgb) {
	b.WriteString("\x1b[48;2;")
	writeUint8(b, c.r)
	b.WriteByte(';')
	writeUint8(b, c.g)
	b.WriteByte(';')
	writeUint8(b, c.b)
	b.WriteByte('m')
}

func writeUint8(b *strings.Builder, v uint8) {
	if v >= 100 {
		b.WriteByte('0' + v/100)
		v %= 100
		b.WriteByte('0' + v/10)
		b.WriteByte('0' + v%10)
	} else if v >= 10 {
		b.WriteByte('0' + v/10)
		b.WriteByte('0' + v%10)
	} else {
		b.WriteByte('0' + v)
	}
}
