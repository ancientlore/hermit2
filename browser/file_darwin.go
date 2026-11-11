//go:build darwin

package browser

import (
	"fmt"
	"strings"
	"syscall"
	"time"
)

var darwinFileFlagNames = []struct {
	bit  uint32
	name string
}{
	{0x00000001, "Nodump"},
	{0x00000002, "User Immutable"},
	{0x00000004, "User Append-Only"},
	{0x00000008, "Opaque"},
	{0x00000020, "Compressed"},
	{0x00000040, "Tracked"},
	{0x00000080, "Data Vault"},
	{0x00008000, "Hidden"},
	{0x00010000, "Archived"},
	{0x00020000, "System Immutable"},
	{0x00040000, "System Append-Only"},
	{0x00080000, "Restricted"},
	{0x00100000, "No Unlink"},
	{0x00800000, "Firmlink"},
	{0x40000000, "Dataless"},
}

func formatDarwinFlags(flags uint32) string {
	if flags == 0 {
		return "0x0 (None)"
	}
	var names []string
	rem := flags
	for _, f := range darwinFileFlagNames {
		if flags&f.bit != 0 {
			names = append(names, f.name)
			rem &^= f.bit
		}
	}
	if rem != 0 {
		names = append(names, fmt.Sprintf("0x%x", rem))
	}
	return fmt.Sprintf("0x%x (%s)", flags, strings.Join(names, ", "))
}

func formatDarwinTimespec(ts syscall.Timespec) string {
	sec, nsec := ts.Unix()
	if sec == 0 && nsec == 0 {
		return "none"
	}
	return time.Unix(sec, nsec).Local().String()
}

func sysInfo(s any) []string {
	if st, ok := s.(*syscall.Stat_t); ok && st != nil {
		return []string{
			fmt.Sprintf("Device:          %d (0x%x)", uint64(uint32(st.Dev)), uint64(uint32(st.Dev))),
			fmt.Sprintf("Inode:           %d", st.Ino),
			fmt.Sprintf("Hard Links:      %d", st.Nlink),
			fmt.Sprintf("Raw Mode:        %s", formatUnixMode(uint32(st.Mode))),
			fmt.Sprintf("UID / GID:       %d / %d", st.Uid, st.Gid),
			fmt.Sprintf("Device ID:       %d (0x%x)", uint64(uint32(st.Rdev)), uint64(uint32(st.Rdev))),
			fmt.Sprintf("Size:            %d bytes", st.Size),
			fmt.Sprintf("Block Size:      %d bytes", st.Blksize),
			fmt.Sprintf("Blocks (512B):   %d", st.Blocks),
			fmt.Sprintf("Birth Time:      %s", formatDarwinTimespec(st.Birthtimespec)),
			fmt.Sprintf("Access Time:     %s", formatDarwinTimespec(st.Atimespec)),
			fmt.Sprintf("Modify Time:     %s", formatDarwinTimespec(st.Mtimespec)),
			fmt.Sprintf("Change Time:     %s", formatDarwinTimespec(st.Ctimespec)),
			fmt.Sprintf("Flags:           %s", formatDarwinFlags(st.Flags)),
			fmt.Sprintf("Generation:      %d", st.Gen),
		}
	}
	return []string{fmt.Sprintf("%#v", s)}
}

