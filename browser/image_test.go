package browser

import (
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/ancientlore/hermit2/scroller"
	"github.com/ancientlore/hermit2/views"
)

func TestBrowserImageViewer(t *testing.T) {
	bm, err := New(os.DirFS(".."), "..", "media")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	// Find hermit.png in media directory
	idx := -1
	for i := 0; i < m.Data.Len(80); i++ {
		if m.Data.At(i).Name() == "hermit.png" {
			idx = i
			break
		}
	}
	if idx == -1 {
		t.Fatalf("hermit.png not found in media folder")
	}

	m.MoveCursor(idx - m.Cursor())

	// Press Right arrow to view hermit.png
	mod, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	imgMod, ok := mod.(scroller.ImageModel[*views.Image])
	if !ok {
		t.Fatalf("expected scroller.ImageModel[*views.Image], got %T", mod)
	}

	mod, _ = imgMod.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	imgMod = mod.(scroller.ImageModel[*views.Image])
	_ = imgMod.View()

	// Zoom in with '+'
	mod, _ = imgMod.Update(tea.KeyPressMsg{Code: '+', Text: "+"})
	imgMod = mod.(scroller.ImageModel[*views.Image])
	if imgMod.Data.ZoomPercent() <= 100 {
		t.Errorf("expected zoom > 100%%, got %d%%", imgMod.Data.ZoomPercent())
	}

	// Left arrow should pan, NOT exit the image viewer
	mod, _ = imgMod.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if _, stillImg := mod.(scroller.ImageModel[*views.Image]); !stillImg {
		t.Fatalf("expected Left arrow to stay in image viewer, got %T", mod)
	}
	imgMod = mod.(scroller.ImageModel[*views.Image])

	// Esc should exit back to browser.Model
	mod, _ = imgMod.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if _, isBrowser := mod.(Model); !isBrowser {
		t.Fatalf("expected Esc to return to browser.Model, got %T", mod)
	}
}

