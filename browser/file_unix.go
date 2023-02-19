//go:build !windows

package browser

import (
	"os/user"
	"strconv"
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
