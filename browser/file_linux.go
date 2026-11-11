//go:build linux

package browser

import (
	"fmt"
	"syscall"
	"time"
)

func formatLinuxTimespec(ts syscall.Timespec) string {
	sec, nsec := ts.Unix()
	if sec == 0 && nsec == 0 {
		return "none"
	}
	return time.Unix(sec, nsec).Local().String()
}

func sysInfo(s any) []string {
	if st, ok := s.(*syscall.Stat_t); ok && st != nil {
		return []string{
			fmt.Sprintf("Device:          %d (0x%x)", uint64(st.Dev), uint64(st.Dev)),
			fmt.Sprintf("Inode:           %d", uint64(st.Ino)),
			fmt.Sprintf("Hard Links:      %d", uint64(st.Nlink)),
			fmt.Sprintf("Raw Mode:        %s", formatUnixMode(uint32(st.Mode))),
			fmt.Sprintf("UID / GID:       %d / %d", st.Uid, st.Gid),
			fmt.Sprintf("Device ID:       %d (0x%x)", uint64(st.Rdev), uint64(st.Rdev)),
			fmt.Sprintf("Size:            %d bytes", st.Size),
			fmt.Sprintf("Block Size:      %d bytes", int64(st.Blksize)),
			fmt.Sprintf("Blocks (512B):   %d", st.Blocks),
			fmt.Sprintf("Access Time:     %s", formatLinuxTimespec(st.Atim)),
			fmt.Sprintf("Modify Time:     %s", formatLinuxTimespec(st.Mtim)),
			fmt.Sprintf("Change Time:     %s", formatLinuxTimespec(st.Ctim)),
		}
	}
	return []string{fmt.Sprintf("%#v", s)}
}

