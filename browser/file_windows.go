//go:build windows

package browser

import (
	"fmt"
	"strings"
	"syscall"
	"time"
)

func owner(s any) string {
	return ""
}

var winFileAttrNames = []struct {
	bit  uint32
	name string
}{
	{0x00000001, "Read-Only"},
	{0x00000002, "Hidden"},
	{0x00000004, "System"},
	{0x00000008, "Volume Label"},
	{0x00000010, "Directory"},
	{0x00000020, "Archive"},
	{0x00000040, "Device"},
	{0x00000080, "Normal"},
	{0x00000100, "Temporary"},
	{0x00000200, "Sparse File"},
	{0x00000400, "Reparse Point"},
	{0x00000800, "Compressed"},
	{0x00001000, "Offline"},
	{0x00002000, "Not Content Indexed"},
	{0x00004000, "Encrypted"},
	{0x00008000, "Integrity Stream"},
	{0x00010000, "Virtual"},
	{0x00020000, "No Scrub Data"},
	{0x00040000, "Recall on Open"},
	{0x00080000, "Pinned"},
	{0x00100000, "Unpinned"},
	{0x00400000, "Recall on Data Access"},
}

func formatWinFileAttributes(attrs uint32) string {
	if attrs == 0 {
		return "0x0 (None)"
	}
	var names []string
	rem := attrs
	for _, a := range winFileAttrNames {
		if attrs&a.bit != 0 {
			names = append(names, a.name)
			rem &^= a.bit
		}
	}
	if rem != 0 {
		names = append(names, fmt.Sprintf("0x%x", rem))
	}
	return fmt.Sprintf("0x%x (%s)", attrs, strings.Join(names, ", "))
}

func formatWinFiletime(ft syscall.Filetime) string {
	if ft.LowDateTime == 0 && ft.HighDateTime == 0 {
		return "none"
	}
	return time.Unix(0, ft.Nanoseconds()).Local().String()
}

func sysInfo(s any) []string {
	if d, ok := s.(*syscall.Win32FileAttributeData); ok && d != nil {
		size := (uint64(d.FileSizeHigh) << 32) | uint64(d.FileSizeLow)
		return []string{
			fmt.Sprintf("File Attributes: %s", formatWinFileAttributes(d.FileAttributes)),
			fmt.Sprintf("Creation Time:   %s", formatWinFiletime(d.CreationTime)),
			fmt.Sprintf("Last Access:     %s", formatWinFiletime(d.LastAccessTime)),
			fmt.Sprintf("Last Write:      %s", formatWinFiletime(d.LastWriteTime)),
			fmt.Sprintf("File Size:       %d bytes (High: 0x%x, Low: 0x%x)", size, d.FileSizeHigh, d.FileSizeLow),
		}
	}
	return []string{fmt.Sprintf("%#v", s)}
}
