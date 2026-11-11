//go:build !windows

package browser

import (
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func owner(s any) string {
	var r string
	if s != nil {
		if ss, ok := s.(*syscall.Stat_t); ok {
			g, err := user.LookupGroupId(strconv.Itoa(int(ss.Gid)))
			if err == nil {
				r += g.Name + ":"
			}
			u, err := user.LookupId(strconv.Itoa(int(ss.Uid)))
			if err == nil {
				r += u.Username
			}
		}
	}
	return r
}

func formatUnixMode(mode uint32) string {
	var parts []string
	switch mode & 0xf000 {
	case 0xc000:
		parts = append(parts, "Socket")
	case 0xa000:
		parts = append(parts, "Symbolic Link")
	case 0x8000:
		parts = append(parts, "Regular File")
	case 0x6000:
		parts = append(parts, "Block Device")
	case 0x4000:
		parts = append(parts, "Directory")
	case 0x2000:
		parts = append(parts, "Character Device")
	case 0x1000:
		parts = append(parts, "FIFO / Named Pipe")
	}
	if mode&04000 != 0 {
		parts = append(parts, "setuid")
	}
	if mode&02000 != 0 {
		parts = append(parts, "setgid")
	}
	if mode&01000 != 0 {
		parts = append(parts, "sticky")
	}
	parts = append(parts, fmt.Sprintf("0%03o", mode&0777))
	return fmt.Sprintf("0%o (%s)", mode, strings.Join(parts, ", "))
}
