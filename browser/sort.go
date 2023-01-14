package browser

import (
	"io/fs"
	"path"
	"strings"
)

type sortByName []fs.DirEntry

func (e sortByName) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	}
	// File/dir sort next
	return e[i].Name() < e[j].Name()
}

func (e sortByName) Len() int {
	return len(e)
}

func (e sortByName) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortByNameRev []fs.DirEntry

func (e sortByNameRev) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	}
	// File/dir sort next
	return e[j].Name() < e[i].Name()
}

func (e sortByNameRev) Len() int {
	return len(e)
}

func (e sortByNameRev) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortByExt []fs.DirEntry

func (e sortByExt) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	} else if e[i].IsDir() && e[j].IsDir() {
		return e[i].Name() < e[j].Name()
	}
	// Special files next
	pi := strings.HasPrefix(e[i].Name(), ".")
	pj := strings.HasPrefix(e[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Extensions next
	ei := path.Ext(e[i].Name())
	ej := path.Ext(e[j].Name())
	if ei < ej {
		return true
	} else if ei == ej {
		// Use name when both extensions match
		return e[i].Name() < e[j].Name()
	}
	return false
}

func (e sortByExt) Len() int {
	return len(e)
}

func (e sortByExt) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortByExtRev []fs.DirEntry

func (e sortByExtRev) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	}
	// Special files next
	pi := strings.HasPrefix(e[i].Name(), ".")
	pj := strings.HasPrefix(e[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Directory name next
	if e[i].IsDir() && e[j].IsDir() {
		return e[j].Name() < e[i].Name()
	}
	// Extensions next
	ei := path.Ext(e[i].Name())
	ej := path.Ext(e[j].Name())
	if ej < ei {
		return true
	} else if ei == ej {
		// Use name when both extensions match
		return e[j].Name() < e[i].Name()
	}
	return false
}

func (e sortByExtRev) Len() int {
	return len(e)
}

func (e sortByExtRev) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortBySize []fs.DirEntry

func (e sortBySize) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	} else if e[i].IsDir() && e[j].IsDir() {
		return e[i].Name() < e[j].Name()
	}
	// Size next
	infoi, _ := e[i].Info()
	infoj, _ := e[j].Info()
	if infoi.Size() < infoj.Size() {
		return true
	} else if infoi.Size() == infoj.Size() {
		// Use name when size is equal
		return e[i].Name() < e[j].Name()
	}
	return false
}

func (e sortBySize) Len() int {
	return len(e)
}

func (e sortBySize) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortBySizeRev []fs.DirEntry

func (e sortBySizeRev) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	} else if e[i].IsDir() && e[j].IsDir() {
		return e[i].Name() < e[j].Name()
	}
	// Size next
	infoi, _ := e[i].Info()
	infoj, _ := e[j].Info()
	if infoj.Size() < infoi.Size() {
		return true
	} else if infoi.Size() == infoj.Size() {
		// Use name when size is equal
		return e[i].Name() < e[j].Name()
	}
	return false
}

func (e sortBySizeRev) Len() int {
	return len(e)
}

func (e sortBySizeRev) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortByDate []fs.DirEntry

func (e sortByDate) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	}
	// Special files next
	pi := strings.HasPrefix(e[i].Name(), ".")
	pj := strings.HasPrefix(e[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Date next
	infoi, _ := e[i].Info()
	infoj, _ := e[j].Info()
	if infoi.ModTime().Before(infoj.ModTime()) {
		return true
	} else if infoi.ModTime() == infoj.ModTime() {
		// Use name when date is equal
		return e[i].Name() < e[j].Name()
	}
	return false
}

func (e sortByDate) Len() int {
	return len(e)
}

func (e sortByDate) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type sortByDateRev []fs.DirEntry

func (e sortByDateRev) Less(i, j int) bool {
	// Directories first
	if e[i].IsDir() && !e[j].IsDir() {
		return true
	} else if !e[i].IsDir() && e[j].IsDir() {
		return false
	}
	// Special files next
	pi := strings.HasPrefix(e[i].Name(), ".")
	pj := strings.HasPrefix(e[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Date next
	infoi, _ := e[i].Info()
	infoj, _ := e[j].Info()
	if infoj.ModTime().Before(infoi.ModTime()) {
		return true
	} else if infoi.ModTime() == infoj.ModTime() {
		// Use name when date is equal
		return e[i].Name() < e[j].Name()
	}
	return false
}

func (e sortByDateRev) Len() int {
	return len(e)
}

func (e sortByDateRev) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
